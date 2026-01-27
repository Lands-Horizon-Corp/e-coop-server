package admin

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	core_admin "github.com/Lands-Horizon-Corp/e-coop-server/src/db/admin"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

// LicenseKey handles license CRUD + activation
func LicenseKeyController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/license",
		Method:       "GET",
		Note:         "Returns all licenses for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: types.LicenseResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		licenses, err := core_admin.LicenseManager(service).List(context)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No licenses found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, licenses)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/license/:license_id",
		Method:       "GET",
		Note:         "Returns a single license by ID.",
		ResponseType: types.LicenseResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		licenseID, err := helpers.EngineUUIDParam(ctx, "license_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid license ID"})
		}
		license, err := core_admin.LicenseManager(service).GetByIDRaw(context, licenseID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "License not found"})
		}
		return ctx.JSON(http.StatusOK, license)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/license",
		Method:       "POST",
		Note:         "Creates a new license.",
		RequestType:  types.LicenseRequest{},
		ResponseType: types.LicenseResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		reqBody := &types.LicenseRequest{}
		if err := ctx.Bind(reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		licenseKey, err := helpers.GenerateLicenseKey()
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate license key: " + err.Error()})
		}
		license := &types.License{
			Name:        reqBody.Name,
			Description: reqBody.Description,
			LicenseKey:  licenseKey,
			ExpiresAt:   reqBody.ExpiresAt,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			UsedAt:      nil,
		}
		if err := core_admin.LicenseManager(service).Create(context, license); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create license: " + err.Error()})
		}
		return ctx.JSON(http.StatusCreated, license)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/license/:license_id",
		Method:       "PUT",
		Note:         "Updates a license by ID.",
		RequestType:  types.LicenseRequest{},
		ResponseType: types.LicenseResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		licenseID, err := helpers.EngineUUIDParam(ctx, "license_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid license ID"})
		}
		req, err := core_admin.LicenseManager(service).Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank data: " + err.Error()})
		}
		license, err := core_admin.LicenseManager(service).GetByID(context, licenseID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "error getting license: " + err.Error()})
		}
		license.Name = req.Name
		license.Description = req.Description
		license.ExpiresAt = req.ExpiresAt
		license.ExpiresAt = req.ExpiresAt
		license.UpdatedAt = time.Now().UTC()
		if err := core_admin.LicenseManager(service).UpdateByID(context, licenseID, license); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update license: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, license)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/license/:license_id/refresh",
		Method:       "POST",
		Note:         "Regenerates the license key for a license by ID.",
		ResponseType: types.LicenseResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		licenseID, err := helpers.EngineUUIDParam(ctx, "license_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid license ID"})
		}
		license, err := core_admin.LicenseManager(service).GetByID(context, licenseID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Error getting license: " + err.Error()})
		}
		newLicenseKey, err := helpers.GenerateLicenseKey()
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate license key: " + err.Error()})
		}
		license.LicenseKey = newLicenseKey
		license.UpdatedAt = time.Now().UTC()
		if err := core_admin.LicenseManager(service).UpdateByID(context, licenseID, license); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update license key: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, license)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/license/activate",
		Method:      "POST",
		Note:        "Activate a license key. Expects JSON { \"license_key\": \"xxx\", \"fingerprint\": \"unique_user_fp\" }. Returns a secret key for verification.",
		RequestType: types.LicenseActivateRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.LicenseActivateRequest
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := service.Validator.Struct(reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		secretKey, err := event.ActivateLicense(context, service, reqBody.LicenseKey, reqBody.Fingerprint)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, secretKey)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/license/verify",
		Method:      "POST",
		Note:        "Verify a license using secret key and fingerprint. Expects JSON { \"secret_key\": \"xxx\", \"fingerprint\": \"unique_user_fp\" }",
		RequestType: types.LicenseVerifyRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.LicenseVerifyRequest
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := service.Validator.Struct(reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		license, err := event.VerifyLicenseByFingerprint(context, service, reqBody.SecretKey, reqBody.Fingerprint)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		if license == nil || license.LicenseKey == "" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "License not found"})
		}
		return ctx.NoContent(http.StatusOK)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/license/deactivate",
		Method:      "POST",
		Note:        "Deactivate a license using secret key and fingerprint. Expects JSON { \"secret_key\": \"xxx\", \"fingerprint\": \"unique_user_fp\" }",
		RequestType: types.LicenseDeactivateRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.LicenseDeactivateRequest
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := service.Validator.Struct(reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		err := event.DeactivateLicense(context, service, reqBody.SecretKey, reqBody.Fingerprint)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, map[string]string{"message": "License deactivated successfully"})
	})
}
