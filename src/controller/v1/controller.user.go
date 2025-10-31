package controller_v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelCore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelCore"
	"github.com/labstack/echo/v4"
)

func (c *Controller) UserController() {
	req := c.provider.Service.Request

	// Returns a specific user by their ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user/:user_id",
		Method:       "GET",
		ResponseType: modelCore.UserResponse{},
		Note:         "Returns a specific user by their ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userId, err := handlers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		user, err := c.modelCore.UserManager.GetByIDRaw(context, *userId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, user)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/profile",
		Method:       "PUT",
		Note:         "Changes the profile of the current user.",
		ResponseType: modelCore.UserResponse{},
		RequestType:  modelCore.UserSettingsChangeProfileRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req modelCore.UserSettingsChangeProfileRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid change profile payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		user.Birthdate = req.Birthdate
		user.FirstName = req.FirstName
		user.MiddleName = req.MiddleName
		user.LastName = req.LastName
		user.FullName = req.FullName
		user.Suffix = req.Suffix
		if err := c.modelCore.UserManager.UpdateFields(context, user.ID, user); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.UserManager.ToModel(user))
	})

	// Change user's password from profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/profile/password",
		Method:       "PUT",
		Note:         "Changes the user's password from profile settings.",
		ResponseType: modelCore.UserResponse{},
		RequestType:  modelCore.UserSettingsChangePasswordRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req modelCore.UserSettingsChangePasswordRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password from profile failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid change password payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password from profile failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}
		valid, err := c.provider.Service.Security.VerifyPassword(context, user.Password, req.OldPassword)
		if err != nil || !valid {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		hashedPwd, err := c.provider.Service.Security.HashPassword(context, req.NewPassword)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password from profile failed: hash password error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password: " + err.Error()})
		}
		user.Password = hashedPwd
		if err := c.modelCore.UserManager.UpdateFields(context, user.ID, user); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password from profile failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := c.modelCore.UserManager.GetByID(context, user.ID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password from profile failed: get updated user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated user: " + err.Error()})
		}
		if err := c.userToken.SetUser(context, ctx, updatedUser); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password from profile failed: set user token error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user token: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Password changed from profile for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, c.modelCore.UserManager.ToModel(updatedUser))
	})

	// Change user's profile picture
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/profile/profile-picture",
		Method:       "PUT",
		Note:         "Changes the user's profile picture.",
		RequestType:  modelCore.UserSettingsChangeProfilePictureRequest{},
		ResponseType: modelCore.UserResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req modelCore.UserSettingsChangeProfilePictureRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change profile picture failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid profile picture update payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change profile picture failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		if user.MediaID == req.MediaID {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Media ID is the same as the current one"})
		}
		user.MediaID = req.MediaID
		if err := c.modelCore.UserManager.UpdateFields(context, user.ID, user); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change profile picture failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := c.modelCore.UserManager.GetByID(context, user.ID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change profile picture failed: get updated user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated user: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Profile picture changed for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, c.modelCore.UserManager.ToModel(updatedUser))
	})

	// Change user's general profile settings
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/profile/general",
		Method:       "PUT",
		Note:         "Changes the user's general profile settings.",
		RequestType:  modelCore.UserSettingsChangeGeneralRequest{},
		ResponseType: modelCore.UserResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req modelCore.UserSettingsChangeGeneralRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change general profile failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general settings update payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change general profile failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		user.UserName = req.UserName
		user.Description = req.Description
		if user.Email != req.Email {
			user.Email = req.Email
			user.IsEmailVerified = false
		}
		if user.ContactNumber != req.ContactNumber {
			user.ContactNumber = req.ContactNumber
			user.IsContactVerified = false
		}
		if err := c.modelCore.UserManager.UpdateFields(context, user.ID, user); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change general profile failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := c.modelCore.UserManager.GetByID(context, user.ID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change general profile failed: get updated user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated user: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "General profile changed for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, c.modelCore.UserManager.ToModel(updatedUser))
	})

}
