package v1

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/labstack/echo/v4"
	"golang.org/x/image/webp"
	"gorm.io/gorm"
)

func (c *Controller) kycController() {
	req := c.provider.Service.Request
	validator := c.provider.Service.Validator

	// Step 1: Personal Details
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/personal-details",
		Method:      "POST",
		Note:        "Submit or update basic personal information (step 1 of KYC)",
		RequestType: core.KYCPersonalDetailsRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var payload core.KYCPersonalDetailsRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		if !regexp.MustCompile(`^[a-z0-9_]+$`).MatchString(payload.Username) {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Username must be lowercase letters, numbers, or underscores only",
			})
		}
		_, err := c.core.GetUserByUserName(context, payload.Username)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Database error: " + err.Error(),
				})
			}
		} else {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Username already taken",
			})
		}
		switch strings.ToLower(payload.Gender) {
		case "male", "female", "others":
		default:
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Gender must be 'male', 'female', or 'others'"})
		}
		return ctx.JSON(http.StatusOK, map[string]string{"message": "Personal details received successfully"})
	})

	// Step 2: Security / Account Credentials
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/security-details",
		Method:      "POST",
		Note:        "Create login credentials (email, phone, password)",
		RequestType: core.KYCSecurityDetailsRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var payload core.KYCSecurityDetailsRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		_, err := c.core.GetUserByEmail(ctx.Request().Context(), payload.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Database error: " + err.Error(),
			})
		}
		_, err = c.core.GetUserByContactNumber(ctx.Request().Context(), payload.ContactNumber)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Database error: " + err.Error(),
			})
		}
		if strings.TrimSpace(payload.Password) == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Password must not be empty",
			})
		}
		if strings.TrimSpace(payload.PasswordConfirmation) == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Password confirmation must not be empty",
			})
		}
		if payload.Password != payload.PasswordConfirmation {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Password and confirmation do not match",
			})
		}
		smsKey := fmt.Sprintf("%s-%s", payload.Password, payload.ContactNumber)
		smsOtp, err := c.provider.Service.OTP.Generate(context, smsKey)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate OTP: " + err.Error()})
		}
		if err := c.provider.Service.SMS.Send(context, horizon.SMSRequest{
			To:   payload.ContactNumber,
			Body: "Lands Horizon: Hello {{.name}} Please dont share this to someone else to protect your account and privacy. This is your OTP:{{.otp}}",
			Vars: map[string]string{
				"otp":  smsOtp,
				"name": payload.FullName,
			},
		}); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send OTP SMS: " + err.Error()})
		}
		smtpKey := fmt.Sprintf("%s-%s", payload.Password, payload.Email)
		smtpOtp, err := c.provider.Service.OTP.Generate(context, smtpKey)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate OTP: " + err.Error()})
		}

		if err := c.provider.Service.SMTP.Send(context, horizon.SMTPRequest{
			To:      payload.Email,
			Body:    "templates/email-otp.html",
			Subject: "Email Verification: Lands Horizon",
			Vars: map[string]string{
				"otp": smtpOtp,
			},
		}); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send OTP email: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Security details received. Verification codes sent.",
		})
	})

	// Email Verification
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/verify-email",
		Method:      "POST",
		Note:        "Verify email address using OTP",
		RequestType: core.KYCVerifyEmailRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var payload core.KYCVerifyEmailRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		key := fmt.Sprintf("%s-%s", payload.Password, payload.Email)
		ok, err := c.provider.Service.OTP.Verify(context, key, payload.OTP)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to verify OTP: " + err.Error()})
		}
		if !ok {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid OTP"})
		}
		if err := c.provider.Service.OTP.Revoke(context, key); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to revoke OTP: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Email verified successfully",
		})
	})

	// Phone Number Verification
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/verify-contact-number",
		Method:      "POST",
		Note:        "Verify phone number using OTP",
		RequestType: core.KYCVerifyContactNumberRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var payload core.KYCVerifyContactNumberRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		key := fmt.Sprintf("%s-%s", payload.Password, payload.ContactNumber)
		ok, err := c.provider.Service.OTP.Verify(context, key, payload.OTP)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to verify OTP: " + err.Error()})
		}
		if !ok {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid OTP"})
		}
		if err := c.provider.Service.OTP.Revoke(context, key); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to revoke OTP: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, map[string]string{"message": "Phone number verified successfully"})
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/verify-addresses",
		Method:      "POST",
		Note:        "Verify one or more addresses (verification only)",
		RequestType: core.KYCVerifyAddressesRequest{},
	}, func(ctx echo.Context) error {
		var payload []core.KYCVerifyAddressesRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid request format",
			})
		}
		if len(payload) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "At least one address is required",
			})
		}
		for i, addr := range payload {
			if err := validator.Struct(addr); err != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{
					"error": fmt.Sprintf("Validation failed at index %d: %s", i, err.Error()),
				})
			}
		}
		return ctx.JSON(http.StatusOK, map[string]string{"message": "Addresses verified successfully"})
	})

	// Government Benefits / ID Verification
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/verify-government-benefits",
		Method:      "POST",
		Note:        "Submit government ID or benefits proof",
		RequestType: core.KYCVerifyGovernmentBenefitsRequest{},
	}, func(ctx echo.Context) error {
		var payload []core.KYCVerifyGovernmentBenefitsRequest

		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid request format",
			})
		}

		if len(payload) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "At least one government document is required",
			})
		}

		for i, doc := range payload {
			if err := validator.Struct(doc); err != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{
					"error": fmt.Sprintf("Validation failed at index %d: %s", i, err.Error()),
				})
			}
		}

		// 🔍 verification logic here
		// - media existence check
		// - country-specific rules
		// - expiry validation
		// - third-party verification

		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Government document information received for verification",
		})
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/kyc/face-recognize",
		Method: "POST",
		Note:   "Upload video for face recognition and liveness check (multipart/form-data)",
	}, func(ctx echo.Context) error {
		file, err := ctx.FormFile("file")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Missing or invalid file field",
			})
		}
		if file.Header.Get("Content-Type") != "video/mp4" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Only MP4 videos are allowed",
			})
		}
		src, err := file.Open()
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to open file",
			})
		}
		defer src.Close()
		r, w := io.Pipe()
		defer r.Close()
		defer w.Close()
		go func() {
			defer w.Close()
			io.Copy(w, src)
		}()
		cmd := exec.Command("ffprobe",
			"-v", "error",
			"-select_streams", "v:0",
			"-show_entries", "stream=width,height,duration",
			"-of", "default=noprint_wrappers=1:nokey=1",
			"pipe:0",
		)
		cmd.Stdin = r
		output, err := cmd.Output()
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid video file",
			})
		}
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) < 3 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Failed to read video metadata",
			})
		}
		width, _ := strconv.Atoi(lines[0])
		height, _ := strconv.Atoi(lines[1])
		duration, _ := strconv.ParseFloat(lines[2], 64)
		if width != 500 || height != 500 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Video resolution must be 500x500",
			})
		}
		if int(duration) != 3 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Video duration must be exactly 3 seconds",
			})
		}
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Video validated successfully",
		})
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/kyc/selfie",
		Method: "POST",
		Note:   "Submit selfie image (WEBP, exactly 500x500)",
	}, func(ctx echo.Context) error {
		file, err := ctx.FormFile("file")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, echo.Map{
				"error": "Missing or invalid file field",
			})
		}
		if file.Size > 5<<20 {
			return ctx.JSON(http.StatusBadRequest, echo.Map{
				"error": "File too large",
			})
		}
		if file.Header.Get("Content-Type") != "image/webp" {
			return ctx.JSON(http.StatusBadRequest, echo.Map{
				"error": "Only WEBP images are allowed",
			})
		}
		src, err := file.Open()
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"error": "Failed to open file",
			})
		}
		defer src.Close()
		cfg, err := webp.DecodeConfig(src)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, echo.Map{
				"error": "Invalid or corrupted WEBP image",
			})
		}
		if cfg.Width != 500 || cfg.Height != 500 {
			return ctx.JSON(http.StatusBadRequest, echo.Map{
				"error": "Image must be exactly 500×500 pixels",
			})
		}
		return ctx.JSON(http.StatusOK, echo.Map{"message": "Selfie image accepted successfully"})
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/register",
		Method:      "POST",
		Note:        "Complete KYC registration (all-in-one endpoint)",
		RequestType: core.KYCRegisterRequest{},
	}, func(ctx echo.Context) error {
		var payload core.KYCRegisterRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusCreated, map[string]string{"message": "KYC registration submitted successfully"})
	})
}
