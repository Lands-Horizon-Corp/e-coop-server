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
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/google/uuid"
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
		var req core.KYCPersonalDetailsRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		if !regexp.MustCompile(`^[a-z0-9_]+$`).MatchString(req.Username) {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Username must be lowercase letters, numbers, or underscores only",
			})
		}
		_, err := c.core.GetUserByUsername(context, req.Username)
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
		switch strings.ToLower(req.Gender) {
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
		var req core.KYCSecurityDetailsRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		_, err := c.core.GetUserByEmail(ctx.Request().Context(), req.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Database error: " + err.Error(),
			})
		}
		_, err = c.core.GetUserByContactNumber(ctx.Request().Context(), req.ContactNumber)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Database error: " + err.Error(),
			})
		}
		if strings.TrimSpace(req.Password) == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Password must not be empty",
			})
		}
		if strings.TrimSpace(req.PasswordConfirmation) == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Password confirmation must not be empty",
			})
		}
		if req.Password != req.PasswordConfirmation {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Password and confirmation do not match",
			})
		}
		smsKey := fmt.Sprintf("%s-%s", req.Password, req.ContactNumber)
		smsOtp, err := c.provider.Service.OTP.Generate(context, smsKey)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate OTP: " + err.Error()})
		}
		if err := c.provider.Service.SMS.Send(context, horizon.SMSRequest{
			To:   req.ContactNumber,
			Body: "Lands Horizon: Hello {{.name}} Please dont share this to someone else to protect your account and privacy. This is your OTP:{{.otp}}",
			Vars: map[string]string{
				"otp":  smsOtp,
				"name": req.FullName,
			},
		}); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send OTP SMS: " + err.Error()})
		}
		smtpKey := fmt.Sprintf("%s-%s", req.Password, req.Email)
		smtpOtp, err := c.provider.Service.OTP.Generate(context, smtpKey)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate OTP: " + err.Error()})
		}

		if err := c.provider.Service.SMTP.Send(context, horizon.SMTPRequest{
			To:      req.Email,
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
		var req core.KYCVerifyEmailRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		key := fmt.Sprintf("%s-%s", req.Password, req.Email)
		ok, err := c.provider.Service.OTP.Verify(context, key, req.OTP)
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
		var req core.KYCVerifyContactNumberRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		key := fmt.Sprintf("%s-%s", req.Password, req.ContactNumber)
		ok, err := c.provider.Service.OTP.Verify(context, key, req.OTP)
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
		var req []core.KYCVerifyAddressesRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid request format",
			})
		}
		if len(req) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "At least one address is required",
			})
		}
		for i, addr := range req {
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
		var req []core.KYCVerifyGovernmentBenefitsRequest

		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid request format",
			})
		}

		if len(req) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "At least one government document is required",
			})
		}

		for i, doc := range req {
			if err := validator.Struct(doc); err != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{
					"error": fmt.Sprintf("Validation failed at index %d: %s", i, err.Error()),
				})
			}
		}
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
		Route:       "/api/v1/kyc/resend-email-verification",
		Method:      "POST",
		Note:        "Resend email verification OTP",
		RequestType: core.KYCResendEmailVerificationRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req core.KYCResendEmailVerificationRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, echo.Map{
				"error": "Invalid request format",
			})
		}
		if err := validator.Struct(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, echo.Map{
				"error": "Validation failed: " + err.Error(),
			})
		}
		key := fmt.Sprintf("%s-%s", req.Password, req.Email)

		otp, err := c.provider.Service.OTP.Generate(context, key)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"error": "Failed to generate OTP: " + err.Error(),
			})
		}

		if err := c.provider.Service.SMTP.Send(context, horizon.SMTPRequest{
			To:      req.Email,
			Subject: "Email Verification: Lands Horizon",
			Body:    "templates/email-otp.html",
			Vars: map[string]string{
				"otp": otp,
			},
		}); err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"error": "Failed to send verification email: " + err.Error(),
			})
		}

		return ctx.JSON(http.StatusOK, echo.Map{
			"message": "Email verification OTP resent successfully",
		})
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/resend-contact-number-verification",
		Method:      "POST",
		Note:        "Resend contact number verification OTP",
		RequestType: core.KYCResendContactNumberVerificationRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.KYCResendContactNumberVerificationRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, echo.Map{
				"error": "Invalid request format",
			})
		}
		if err := validator.Struct(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, echo.Map{
				"error": "Validation failed: " + err.Error(),
			})
		}
		key := fmt.Sprintf("%s-%s", req.Password, req.ContactNumber)
		otp, err := c.provider.Service.OTP.Generate(context, key)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"error": "Failed to generate OTP: " + err.Error(),
			})
		}
		if err := c.provider.Service.SMS.Send(context, horizon.SMSRequest{
			To:   req.ContactNumber,
			Body: "Lands Horizon: Hello {{.name}}, do not share this code. Your OTP is {{.otp}}",
			Vars: map[string]string{
				"otp":  otp,
				"name": req.FullName,
			},
		}); err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"error": "Failed to send OTP SMS: " + err.Error(),
			})
		}

		return ctx.JSON(http.StatusOK, echo.Map{
			"message": "Contact number verification OTP resent successfully",
		})
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/register",
		Method:      "POST",
		Note:        "Complete KYC registration (all-in-one endpoint)",
		RequestType: core.KYCRegisterRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.KYCRegisterRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		if req.Password != req.PasswordConfirmation {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Password and confirmation do not match"})
		}
		org, ok := c.userOrganizationToken.GetOrganization(ctx)
		if !ok {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		hashedPwd, err := c.provider.Service.Security.HashPassword(context, req.Password)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password: " + endTx(err).Error()})
		}
		userProfile := &core.User{
			Email:             req.Email,
			Username:          req.Username,
			ContactNumber:     req.Phone,
			Password:          hashedPwd,
			FullName:          req.FullName,
			FirstName:         &req.FirstName,
			MiddleName:        &req.MiddleName,
			LastName:          &req.LastName,
			Suffix:            &req.Suffix,
			IsEmailVerified:   true,
			IsContactVerified: true,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
			Birthdate:         req.BirthDate,
			MediaID:           req.SelfieMediaID,
		}
		if err := c.core.UserManager().CreateWithTx(context, tx, userProfile); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not create user profile: " + endTx(err).Error()})
		}
		passbook := req.OldPassbook
		if passbook == "" {
			passbook = handlers.GeneratePassbookNumber()
		}
		memberProfile := &core.MemberProfile{
			OrganizationID:       org.ID,
			BranchID:             *req.BranchID,
			CreatedAt:            time.Now().UTC(),
			UpdatedAt:            time.Now().UTC(),
			UserID:               &userProfile.ID,
			OldReferenceID:       req.OldPassbook,
			Passbook:             passbook,
			FirstName:            req.FirstName,
			MiddleName:           req.MiddleName,
			LastName:             req.LastName,
			FullName:             req.FullName,
			Suffix:               req.Suffix,
			MemberGenderID:       &req.MemberGenderID,
			BirthDate:            req.BirthDate,
			ContactNumber:        req.ContactNumber,
			CivilStatus:          req.CivilStatus,
			MemberOccupationID:   req.MemberOccupationID,
			IsMutualFundMember:   false,
			IsMicroFinanceMember: false,
			MediaID:              req.SelfieMediaID,
		}
		if err := c.core.MemberProfileManager().CreateWithTx(context, tx, memberProfile); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not create member profile: " + endTx(err).Error()})
		}
		for _, addrReq := range req.Addresses {
			value := &core.MemberAddress{
				MemberProfileID: &memberProfile.ID,
				Label:           addrReq.Label,
				City:            addrReq.City,
				CountryCode:     addrReq.CountryCode,
				PostalCode:      addrReq.PostalCode,
				ProvinceState:   addrReq.ProvinceState,
				Barangay:        addrReq.Barangay,
				Landmark:        addrReq.Landmark,
				Address:         addrReq.Address,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
				BranchID:        *req.BranchID,
				OrganizationID:  org.ID,
				Longitude:       addrReq.Longitude,
				Latitude:        addrReq.Latitude,
			}
			if err := c.core.MemberAddressManager().Create(context, value); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member address record: " + err.Error()})
			}
		}
		for _, govReq := range req.GovernmentBenefits {

			value := &core.MemberGovernmentBenefit{
				MemberProfileID: memberProfile.ID,
				FrontMediaID:    govReq.FrontMediaID,
				BackMediaID:     govReq.BackMediaID,
				CountryCode:     govReq.CountryCode,
				Description:     govReq.Description,
				Name:            govReq.Name,
				Value:           govReq.Value,
				ExpiryDate:      govReq.ExpiryDate,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
				BranchID:        *req.BranchID,
				OrganizationID:  org.ID,
			}
			if err := c.core.MemberGovernmentBenefitManager().Create(context, value); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create government benefit record: " + err.Error()})
			}
		}
		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, userProfile.ID.String())
		developerKey = developerKey + uuid.NewString() + "-horizon"
		newUserOrg := &core.UserOrganization{
			CreatedAt:                time.Now().UTC(),
			UpdatedAt:                time.Now().UTC(),
			OrganizationID:           org.ID,
			BranchID:                 req.BranchID,
			UserID:                   userProfile.ID,
			UserType:                 core.UserOrganizationTypeMember,
			Description:              "",
			ApplicationDescription:   "anything",
			ApplicationStatus:        "accepted",
			DeveloperSecretKey:       developerKey,
			PermissionName:           string(core.UserOrganizationTypeMember),
			PermissionDescription:    "",
			Permissions:              []string{},
			UserSettingDescription:   "user settings",
			UserSettingStartOR:       0,
			UserSettingEndOR:         1000,
			UserSettingUsedOR:        0,
			UserSettingStartVoucher:  0,
			UserSettingEndVoucher:    0,
			UserSettingUsedVoucher:   0,
			UserSettingNumberPadding: 7,
		}
		if err := c.core.UserOrganizationManager().CreateWithTx(context, tx, newUserOrg); err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to create UserOrganization: " + endTx(err).Error()})
		}
		if err := endTx(nil); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction commit failed: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, map[string]string{"message": "KYC registration submitted successfully"})
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/kyc/login",
		Method:       "POST",
		RequestType:  core.KYCLoginRequest{},
		ResponseType: core.CurrentUserResponse{},
		Note:         "Authenticates a KYC user using email, username, or phone and returns user details.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.KYCLoginRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid login payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		org, ok := c.userOrganizationToken.GetOrganization(ctx)
		if !ok {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		user, err := c.core.GetUserByIdentifier(context, req.Key)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials: " + err.Error()})
		}
		valid, err := c.provider.Service.Security.VerifyPassword(context, user.Password, req.Password)
		if err != nil || !valid {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		if !user.IsEmailVerified || !user.IsContactVerified {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User has not completed KYC verification"})
		}
		if err := c.userToken.SetUser(context, ctx, user); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user token: " + err.Error()})
		}
		userOrg, err := c.core.UserOrganizationManager().FindOne(context, &core.UserOrganization{
			UserID:         user.ID,
			OrganizationID: org.ID,
			UserType:       core.UserOrganizationTypeMember,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}
		if userOrganization.ApplicationStatus == "accepted" {
			if err := c.userOrganizationToken.SetUserOrganization(context, ctx, userOrganization); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user organization: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModel(userOrganization))
		}
		return ctx.JSON(http.StatusOK, core.CurrentUserResponse{
			UserID: user.ID,
			User:   c.core.UserManager().ToModel(user),
		})
	})
}
