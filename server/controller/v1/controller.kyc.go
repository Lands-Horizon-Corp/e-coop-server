package v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
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
		var payload core.KYCPersonalDetailsRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		// TODO: process personal details
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Personal details received successfully",
		})
	})

	// Step 2: Security / Account Credentials
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/security-details",
		Method:      "POST",
		Note:        "Create login credentials (email, phone, password)",
		RequestType: core.KYCSecurityDetailsRequest{},
	}, func(ctx echo.Context) error {
		var payload core.KYCSecurityDetailsRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		// TODO: create credentials + send OTPs
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
		var payload core.KYCVerifyEmailRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		// TODO: verify email OTP
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
		var payload core.KYCVerifyContactNumberRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		// TODO: verify phone OTP
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Phone number verified successfully",
		})
	})

	// Address Verification
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/verify-addresses",
		Method:      "POST",
		Note:        "Submit or verify address information",
		RequestType: core.KYCVerifyAddressesRequest{},
	}, func(ctx echo.Context) error {
		var payload core.KYCVerifyAddressesRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		// TODO: save / verify address
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Address information received",
		})
	})

	// Government Benefits / ID Verification
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/verify-government-benefits",
		Method:      "POST",
		Note:        "Submit government ID or benefits proof",
		RequestType: core.KYCVerifyGovernmentBenefitsRequest{},
	}, func(ctx echo.Context) error {
		var payload core.KYCVerifyGovernmentBenefitsRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		// TODO: process government ID / benefit proof
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Government document information received",
		})
	})

	// Face Recognition / Liveness Check (multipart/form-data)
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/kyc/face-recognize",
		Method: "POST",
		Note:   "Upload photo for face recognition and liveness check (multipart/form-data)",
	}, func(ctx echo.Context) error {
		// Special handling for file upload (multipart)
		_, err := ctx.FormFile("file")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Missing or invalid file field",
			})
		}

		// TODO: upload file, process face recognition/liveness
		// Usually you would call media upload service here

		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Face photo uploaded successfully",
		})
	})

	// Selfie Submission (JSON - media ID reference)
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/selfie",
		Method:      "POST",
		Note:        "Submit already uploaded selfie media ID",
		RequestType: core.KYCSelfieRequest{},
	}, func(ctx echo.Context) error {
		var payload core.KYCSelfieRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}
		if err := validator.Struct(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		// TODO: verify selfie media exists & associate with user
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Selfie reference submitted successfully",
		})
	})

	// Final KYC Registration (All-in-one)
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

		// TODO: full KYC processing (create user, save all data, start verification)

		return ctx.JSON(http.StatusCreated, map[string]string{
			"message": "KYC registration submitted successfully",
		})
	})
}
