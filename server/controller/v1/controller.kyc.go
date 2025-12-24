package v1

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) kycController() {
	req := c.provider.Service.Request

	// Step 1: Personal Details
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/personal-details",
		Method:      "POST",
		Note:        "Submit or update basic personal information (step 1 of KYC)",
		RequestType: core.KYCPersonalDetailsRequest{},
	}, func(ctx echo.Context) error {
		return nil
	})

	// Step 2: Security / Account Credentials
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/security-details",
		Method:      "POST",
		Note:        "Create login credentials (email, phone, password)",
		RequestType: core.KYCSecurityDetailsRequest{},
	}, func(ctx echo.Context) error {
		return nil
	})

	// Email Verification
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/verify-email",
		Method:      "POST",
		Note:        "Verify email address using OTP",
		RequestType: core.KYCVerifyEmailRequest{},
	}, func(ctx echo.Context) error {
		return nil
	})

	// Phone Number Verification
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/verify-contact-number",
		Method:      "POST",
		Note:        "Verify phone number using OTP",
		RequestType: core.KYCVerifyContactNumberRequest{},
	}, func(ctx echo.Context) error {
		return nil
	})

	// Address Verification
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/verify-addresses",
		Method:      "POST",
		Note:        "Submit or verify address information",
		RequestType: core.KYCVerifyAddressesRequest{},
	}, func(ctx echo.Context) error {
		return nil
	})

	// Government Benefits / ID Verification
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/verify-government-benefits",
		Method:      "POST",
		Note:        "Submit government ID or benefits proof",
		RequestType: core.KYCVerifyGovernmentBenefitsRequest{},
	}, func(ctx echo.Context) error {
		return nil
	})

	// Face Recognition / Liveness Check
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/kyc/face-recognize",
		Method: "POST",
		Note:   "Upload photo for face recognition and liveness check (multipart/form-data)",
	}, func(ctx echo.Context) error {
		return nil
	})

	// Selfie Submission
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/selfie",
		Method:      "POST",
		Note:        "Submit already uploaded selfie media ID",
		RequestType: core.KYCSelfieRequest{},
	}, func(ctx echo.Context) error {
		return nil
	})

	// Final KYC Registration (All-in-one or final submission)
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/kyc/register",
		Method:      "POST",
		Note:        "Complete KYC registration (all-in-one endpoint)",
		RequestType: core.KYCRegisterRequest{},
	}, func(ctx echo.Context) error {
		return nil
	})
}
