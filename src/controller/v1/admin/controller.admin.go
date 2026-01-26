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

func AdminController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/admin/current",
		Method:       "GET",
		ResponseType: types.AdminResponse{},
		Note:         "Returns the current authenticated admin.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		admin, err := event.CurrentAdmin(context, service, ctx)
		if err != nil {
			event.ClearCurrentAdminCSRF(context, service, ctx)
			return ctx.NoContent(http.StatusUnauthorized)
		}
		return ctx.JSON(http.StatusOK, core_admin.AdminManager(service).ToModel(admin))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/admin/login",
		Method:       "POST",
		RequestType:  types.AdminLoginRequest{},
		ResponseType: types.AdminResponse{},
		Note:         "Authenticates an admin and returns admin details.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.AdminLoginRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid login payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		if req.AdminSuperPassword != service.Config.AppToken {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Cannot create admin. You are now blocked"})
		}
		admin, err := core_admin.GetAdminByIdentifier(context, service, req.Key)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials: " + err.Error()})
		}
		valid, err := service.Security.VerifyPassword(admin.Password, req.Password)
		if err != nil || !valid {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		if err := event.SetAdmin(context, service, ctx, admin); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set admin token: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core_admin.AdminManager(service).ToModel(admin))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/admin/register",
		Method:       "POST",
		RequestType:  types.AdminRegisterRequest{},
		ResponseType: types.AdminResponse{},
		Note:         "Registers a new admin account.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		reqData, err := core_admin.AdminManager(service).Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		if reqData.AdminSuperPassword != service.Config.AppToken {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Cannot create admin. You are now blocked"})
		}
		hashedPwd, err := service.Security.HashPassword(reqData.Password)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password: " + err.Error()})
		}
		admin := &types.Admin{
			Email:     reqData.Email,
			Password:  hashedPwd,
			Username:  reqData.Username,
			FullName:  reqData.FullName,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := core_admin.AdminManager(service).Create(context, admin); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not register admin: " + err.Error()})
		}
		if err := event.SetAdmin(context, service, ctx, admin); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set admin token: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core_admin.AdminManager(service).ToModel(admin))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/authentication/current-logged-in-accounts/logout",
		Method: "POST",
		Note:   "Logs out all users including itself for the session.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		_, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		if err := event.LogoutOtherDevices(context, service, ctx); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to logout other devices: " + err.Error()})
		}
		event.ClearCurrentToken(context, service, ctx)
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/admin/:admin_id",
		Method:       "GET",
		ResponseType: types.AdminResponse{},
		Note:         "Returns a specific admin by their ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		adminID, err := helpers.EngineUUIDParam(ctx, "admin_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid admin_id: " + err.Error()})
		}
		admin, err := core_admin.AdminManager(service).GetByIDRaw(context, *adminID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve admin: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, admin)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/admin/profile",
		Method:       "PUT",
		Note:         "Changes the profile of the current admin.",
		ResponseType: types.AdminResponse{},
		RequestType:  types.AdminSettingsChangeProfileRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.AdminSettingsChangeProfileRequest
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid change profile payload: " + err.Error()})
		}
		if err := service.Validator.Struct(reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		admin, err := event.CurrentAdmin(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		admin.FirstName = reqBody.FirstName
		admin.MiddleName = reqBody.MiddleName
		admin.LastName = reqBody.LastName
		admin.FullName = reqBody.FullName
		admin.Suffix = reqBody.Suffix
		admin.Description = reqBody.Description
		if err := core_admin.AdminManager(service).UpdateByID(context, admin.ID, admin); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update admin profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core_admin.AdminManager(service).ToModel(admin))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/admin/profile/password",
		Method:       "PUT",
		Note:         "Changes the admin's password.",
		ResponseType: types.AdminResponse{},
		RequestType:  types.AdminSettingsChangePasswordRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.AdminSettingsChangePasswordRequest
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid change password payload: " + err.Error()})
		}
		if err := service.Validator.Struct(reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		admin, err := event.CurrentAdmin(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		valid, err := service.Security.VerifyPassword(admin.Password, reqBody.OldPassword)
		if err != nil || !valid {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		hashedPwd, err := service.Security.HashPassword(reqBody.NewPassword)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password: " + err.Error()})
		}
		admin.Password = hashedPwd
		if err := core_admin.AdminManager(service).UpdateByID(context, admin.ID, admin); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update admin password: " + err.Error()})
		}
		updatedAdmin, err := core_admin.AdminManager(service).GetByID(context, admin.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated admin: " + err.Error()})
		}
		if err := event.SetAdmin(context, service, ctx, updatedAdmin); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set admin token: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core_admin.AdminManager(service).ToModel(updatedAdmin))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/admin/profile/general",
		Method:       "PUT",
		Note:         "Changes the admin's general settings.",
		RequestType:  types.AdminSettingsChangeGeneralRequest{},
		ResponseType: types.AdminResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.AdminSettingsChangeGeneralRequest
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general settings update payload: " + err.Error()})
		}
		if err := service.Validator.Struct(reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		admin, err := event.CurrentAdmin(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		admin.Username = reqBody.Username
		admin.IsActive = reqBody.IsActive
		if admin.Email != reqBody.Email {
			admin.Email = reqBody.Email
			admin.IsEmailVerified = false
		}
		if err := core_admin.AdminManager(service).UpdateByID(context, admin.ID, admin); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update admin general settings: " + err.Error()})
		}
		updatedAdmin, err := core_admin.AdminManager(service).GetByID(context, admin.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated admin: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core_admin.AdminManager(service).ToModel(updatedAdmin))
	})
}
