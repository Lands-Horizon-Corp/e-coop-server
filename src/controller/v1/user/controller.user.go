package user

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func UserController(service *horizon.HorizonService) {
	

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/user/:user_id",
		Method:       "GET",
		ResponseType: types.UserResponse{},
		Note:         "Returns a specific user by their ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userID, err := helpers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		user, err := core.UserManager(service).GetByIDRaw(context, *userID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, user)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/profile",
		Method:       "PUT",
		Note:         "Changes the profile of the current user.",
		ResponseType: types.UserResponse{},
		RequestType:  types.UserSettingsChangeProfileRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.UserSettingsChangeProfileRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid change profile payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		user.Birthdate = req.Birthdate
		user.FirstName = req.FirstName
		user.MiddleName = req.MiddleName
		user.LastName = req.LastName
		user.FullName = req.FullName
		user.Suffix = req.Suffix
		if err := core.UserManager(service).UpdateByID(context, user.ID, user); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.UserManager(service).ToModel(user))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/profile/password",
		Method:       "PUT",
		Note:         "Changes the user's password from profile settings.",
		ResponseType: types.UserResponse{},
		RequestType:  types.UserSettingsChangePasswordRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.UserSettingsChangePasswordRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password from profile failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid change password payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password from profile failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}
		valid, err := service.Security.VerifyPassword(user.Password, req.OldPassword)
		if err != nil || !valid {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		hashedPwd, err := service.Security.HashPassword(req.NewPassword)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password from profile failed: hash password error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password: " + err.Error()})
		}
		user.Password = hashedPwd
		if err := core.UserManager(service).UpdateByID(context, user.ID, user); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password from profile failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := core.UserManager(service).GetByID(context, user.ID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password from profile failed: get updated user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated user: " + err.Error()})
		}
		if err := event.SetUser(context, service, ctx, updatedUser); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password from profile failed: set user token error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user token: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Password changed from profile for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, core.UserManager(service).ToModel(updatedUser))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/profile/profile-picture",
		Method:       "PUT",
		Note:         "Changes the user's profile picture.",
		RequestType:  types.UserSettingsChangeProfilePictureRequest{},
		ResponseType: types.UserResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.UserSettingsChangeProfilePictureRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change profile picture failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid profile picture update payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change profile picture failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		if user.MediaID == req.MediaID {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Media ID is the same as the current one"})
		}
		user.MediaID = req.MediaID
		if err := core.UserManager(service).UpdateByID(context, user.ID, user); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change profile picture failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := core.UserManager(service).GetByID(context, user.ID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change profile picture failed: get updated user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated user: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Profile picture changed for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, core.UserManager(service).ToModel(updatedUser))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/profile/general",
		Method:       "PUT",
		Note:         "Changes the user's general profile settings.",
		RequestType:  types.UserSettingsChangeGeneralRequest{},
		ResponseType: types.UserResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.UserSettingsChangeGeneralRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change general profile failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general settings update payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change general profile failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		user.Username = req.Username
		user.Description = req.Description
		if user.Email != req.Email {
			user.Email = req.Email
			user.IsEmailVerified = false
		}
		if user.ContactNumber != req.ContactNumber {
			user.ContactNumber = req.ContactNumber
			user.IsContactVerified = false
		}
		if err := core.UserManager(service).UpdateByID(context, user.ID, user); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change general profile failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := core.UserManager(service).GetByID(context, user.ID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change general profile failed: get updated user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated user: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "General profile changed for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, core.UserManager(service).ToModel(updatedUser))
	})

}
