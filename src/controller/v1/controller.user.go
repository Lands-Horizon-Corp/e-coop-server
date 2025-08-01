package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/cooperative_tokens"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) UserController() {
	req := c.provider.Service.Request

	// Returns a specific user by their ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user/:user_id",
		Method:       "GET",
		ResponseType: model.UserResponse{},
		Note:         "Returns a specific user by their ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userId, err := handlers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		user, err := c.model.UserManager.GetByIDRaw(context, *userId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, user)
	})

	// Returns the current authenticated user and their user organization, if any.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/authentication/current",
		Method:       "GET",
		ResponseType: model.CurrentUserResponse{},
		Note:         "Returns the current authenticated user and their user organization, if any.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.userOrganizationToken.ClearCurrentToken(context, ctx)
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrganization, _ := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		var userOrg *model.UserOrganizationResponse
		if userOrganization != nil {
			userOrg = c.model.UserOrganizationManager.ToModel(userOrganization)
		}
		return ctx.JSON(http.StatusOK, model.CurrentUserResponse{
			UserID:           user.ID,
			User:             c.model.UserManager.ToModel(user),
			UserOrganization: userOrg,
		})
	})

	// Returns all currently logged-in users for the session
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/authentication/current-logged-in-accounts",
		Note:         "Returns all currently logged-in users for the session.",
		Method:       "GET",
		ResponseType: cooperative_tokens.UserCSRFResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		_, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		loggedIn, err := c.userToken.CSRF.GetLoggedInUsers(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get logged in users: " + err.Error()})
		}
		var resp cooperative_tokens.UserCSRFResponse
		loggedInPtrs := make([]*cooperative_tokens.UserCSRF, len(loggedIn))
		for i := range loggedIn {
			loggedInPtrs[i] = &loggedIn[i]
		}
		responses := resp.UserCSRFModels(loggedInPtrs)
		return ctx.JSON(http.StatusOK, responses)
	})

	// Logout all users including itself for the session
	req.RegisterRoute(handlers.Route{
		Route:  "/authentication/current-logged-in-accounts/logout",
		Method: "POST",
		Note:   "Logs out all users including itself for the session.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		_, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		if err := c.userToken.CSRF.LogoutOtherDevices(context, ctx); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to logout other devices: " + err.Error()})
		}
		c.userOrganizationToken.ClearCurrentToken(context, ctx)
		return ctx.NoContent(http.StatusNoContent)
	})

	// Authenticate user login
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/authentication/login",
		Method:       "POST",
		RequestType:  model.UserLoginRequest{},
		ResponseType: model.CurrentUserResponse{},
		Note:         "Authenticates a user and returns user details.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserLoginRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid login payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.model.GetUserByIdentifier(context, req.Key)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials: " + err.Error()})
		}
		valid, err := c.provider.Service.Security.VerifyPassword(context, user.Password, req.Password)
		if err != nil || !valid {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		if err := c.userToken.SetUser(context, ctx, user); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user token: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "User logged in successfully: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, model.CurrentUserResponse{
			UserID: user.ID,
			User:   c.model.UserManager.ToModel(user),
		})
	})

	// Logout the current user
	req.RegisterRoute(handlers.Route{
		Route:  "/authentication/logout",
		Method: "POST",
		Note:   "Logs out the current user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "User logged out successfully",
			Module:      "User",
		})
		c.userToken.CSRF.ClearCSRF(context, ctx)
		return ctx.NoContent(http.StatusNoContent)
	})

	// Register a new user
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/authentication/register",
		Method:       "POST",
		ResponseType: model.CurrentUserResponse{},
		RequestType:  model.UserRegisterRequest{},
		Note:         "Registers a new user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.UserManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Register failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		hashedPwd, err := c.provider.Service.Security.HashPassword(context, req.Password)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Register failed: hash password error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password: " + err.Error()})
		}
		user := &model.User{
			Email:             req.Email,
			Password:          hashedPwd,
			Birthdate:         req.Birthdate,
			UserName:          req.UserName,
			FullName:          req.FullName,
			FirstName:         req.FirstName,
			MiddleName:        req.MiddleName,
			LastName:          req.LastName,
			Suffix:            req.Suffix,
			ContactNumber:     req.ContactNumber,
			MediaID:           req.MediaID,
			IsEmailVerified:   false,
			IsContactVerified: false,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
		}
		if err := c.model.UserManager.Create(context, user); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Register failed: create user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not register user: " + err.Error()})
		}
		if err := c.userToken.SetUser(context, ctx, user); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Register failed: failed to set user token: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user token: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "User registered successfully: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, model.CurrentUserResponse{
			UserID: user.ID,
			User:   c.model.UserManager.ToModel(user),
		})
	})

	// Forgot password flow
	req.RegisterRoute(handlers.Route{
		Route:       "/authentication/forgot-password",
		Method:      "POST",
		RequestType: model.UserForgotPasswordRequest{},
		Note:        "Initiates forgot password flow and sends a reset link.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserForgotPasswordRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Forgot password failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid forgot password payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Forgot password failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.model.GetUserByIdentifier(context, req.Key)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Forgot password failed: user not found: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No account found with those details: " + err.Error()})
		}
		token, err := c.provider.Service.Security.GenerateUUIDv5(context, user.Password)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Forgot password failed: generate token error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Error generating reset token: " + err.Error()})
		}
		fallback := c.provider.Service.Environment.Get("APP_CLIENT_URL", "")
		fallbackStr, _ := fallback.(string)
		if err := c.provider.Service.SMTP.Send(context, horizon.SMTPRequest{
			To:      req.Key,
			Subject: "Forgot Password: Lands Horizon",
			Body:    "templates/email-change-password.html",
			Vars: map[string]string{
				"name":      user.FullName,
				"eventLink": fallbackStr + "/auth/password-reset/" + token,
			},
		}); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Forgot password failed: send email error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed sending email: " + err.Error()})
		}
		if err := c.provider.Service.Cache.Set(context, token, user.ID, 10*time.Minute); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Forgot password failed: cache set error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed storing token: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Forgot password initiated for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Verify password reset link
	req.RegisterRoute(handlers.Route{
		Route:  "/authentication/verify-reset-link/:reset_id",
		Method: "GET",
		Note:   "Verifies if the reset password link is valid.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		resetID := ctx.Param("reset_id")
		if resetID == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Reset ID is required"})
		}
		userID, err := c.provider.Service.Cache.Get(context, resetID)
		if err != nil || userID == nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Reset link is invalid or expired"})
		}
		userId, err := uuid.Parse(string(userID))
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID: " + err.Error()})
		}
		_, err = c.model.UserManager.GetByID(context, userId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found for reset token: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// Change password using the reset link
	req.RegisterRoute(handlers.Route{
		Route:       "/authentication/change-password/:reset_id",
		Method:      "POST",
		RequestType: model.UserChangePasswordRequest{},
		Note:        "Changes the user's password using the reset link.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserChangePasswordRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid change password payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		resetID := ctx.Param("reset_id")
		if resetID == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Reset ID is required"})
		}
		userID, err := c.provider.Service.Cache.Get(context, resetID)
		if err != nil || userID == nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Reset link is invalid or expired"})
		}
		userId, err := uuid.Parse(string(userID))
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID: " + err.Error()})
		}
		user, err := c.model.UserManager.GetByID(context, userId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found for reset token: " + err.Error()})
		}
		hashedPwd, err := c.provider.Service.Security.HashPassword(context, req.NewPassword)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password failed: hash password error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password: " + err.Error()})
		}
		user.Password = hashedPwd
		if err := c.model.UserManager.UpdateFields(context, user.ID, user); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user password: " + err.Error()})
		}
		if err := c.provider.Service.Cache.Delete(context, resetID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password failed: delete cache error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete token from cache: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Password changed successfully for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Send OTP for contact number verification
	req.RegisterRoute(handlers.Route{
		Route:  "/authentication/apply-contact-number",
		Method: "POST",
		Note:   "Sends OTP for contact number verification.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		key := fmt.Sprintf("%s-%s", user.Password, user.ContactNumber)
		otp, err := c.provider.Service.OTP.Generate(context, key)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Apply contact number failed: generate OTP error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate OTP: " + err.Error()})
		}
		if err := c.provider.Service.SMS.Send(context, horizon.SMSRequest{
			To:   user.ContactNumber,
			Body: "Lands Horizon: Hello {{.name}} Please dont share this to someone else to protect your account and privacy. This is your OTP:{{.otp}}",
			Vars: map[string]string{
				"otp":  otp,
				"name": *user.FirstName,
			},
		}); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Apply contact number failed: send SMS error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send OTP SMS: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "OTP sent for contact number verification: " + user.ContactNumber,
			Module:      "User",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Verify OTP for contact number
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/authentication/verify-contact-number",
		Method:       "POST",
		RequestType:  model.UserVerifyContactNumberRequest{},
		ResponseType: model.UserResponse{},
		Note:         "Verifies OTP for contact number verification.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserVerifyContactNumberRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid verify contact number payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		key := fmt.Sprintf("%s-%s", user.Password, user.ContactNumber)
		ok, err := c.provider.Service.OTP.Verify(context, key, req.OTP)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: verify OTP error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to verify OTP: " + err.Error()})
		}
		if !ok {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: invalid OTP",
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid OTP"})
		}
		if err := c.provider.Service.OTP.Revoke(context, key); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: revoke OTP error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to revoke OTP: " + err.Error()})
		}
		user.IsContactVerified = true
		if err := c.model.UserManager.UpdateFields(context, user.ID, user); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := c.model.UserManager.GetByID(context, user.ID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: get updated user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated user: " + err.Error()})
		}
		if err := c.userToken.SetUser(context, ctx, updatedUser); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: set user token error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user token: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Contact number verified for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, c.model.UserManager.ToModel(updatedUser))
	})

	// Send OTP for email verification
	req.RegisterRoute(handlers.Route{
		Route:  "/authentication/apply-email",
		Method: "POST",
		Note:   "Sends OTP for email verification.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		key := fmt.Sprintf("%s-%s", user.Password, user.Email)
		otp, err := c.provider.Service.OTP.Generate(context, key)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Apply email failed: generate OTP error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate OTP: " + err.Error()})
		}
		if err := c.provider.Service.SMTP.Send(context, horizon.SMTPRequest{
			To:      user.Email,
			Body:    "templates/email-otp.html",
			Subject: "Email Verification: Lands Horizon",
			Vars: map[string]string{
				"otp": otp,
			},
		}); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Apply email failed: send email error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send OTP email: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "OTP sent for email verification: " + user.Email,
			Module:      "User",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Verify OTP for email
	req.RegisterRoute(handlers.Route{
		Route:  "/authentication/verify-email",
		Method: "POST",

		Note:         "Verifies OTP for email verification.",
		ResponseType: model.UserResponse{},
		RequestType:  model.UserVerifyEmailRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserVerifyEmailRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid verify email payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		key := fmt.Sprintf("%s-%s", user.Password, user.Email)
		ok, err := c.provider.Service.OTP.Verify(context, key, req.OTP)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: verify OTP error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to verify OTP: " + err.Error()})
		}
		if !ok {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: invalid OTP",
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid OTP"})
		}
		if err := c.provider.Service.OTP.Revoke(context, key); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: revoke OTP error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to revoke OTP: " + err.Error()})
		}
		user.IsEmailVerified = true
		if err := c.model.UserManager.UpdateFields(context, user.ID, user); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := c.model.UserManager.GetByID(context, user.ID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: get updated user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated user: " + err.Error()})
		}
		if err := c.userToken.SetUser(context, ctx, updatedUser); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: set user token error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user token: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Email verified for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, c.model.UserManager.ToModel(updatedUser))
	})

	// Verify user with password for self-protected actions
	req.RegisterRoute(handlers.Route{
		Route:  "/authentication/verify-with-password",
		Method: "POST",

		Note:         "Verifies the user's password for protected self actions.",
		ResponseType: model.UserResponse{},
		RequestType:  model.UserVerifyWithPasswordRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserVerifyWithPasswordRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid verify with password payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}
		valid, err := c.provider.Service.Security.VerifyPassword(context, user.Password, req.Password)
		if err != nil || !valid {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// Change user's password from profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/profile/password",
		Method:       "PUT",
		Note:         "Changes the user's password from profile settings.",
		ResponseType: model.UserResponse{},
		RequestType:  model.UserSettingsChangePasswordRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserSettingsChangePasswordRequest
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
		if err := c.model.UserManager.UpdateFields(context, user.ID, user); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password from profile failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := c.model.UserManager.GetByID(context, user.ID)
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
		return ctx.JSON(http.StatusOK, c.model.UserManager.ToModel(updatedUser))
	})

	// Change user's profile picture
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/profile/profile-picture",
		Method:       "PUT",
		Note:         "Changes the user's profile picture.",
		RequestType:  model.UserSettingsChangeProfilePictureRequest{},
		ResponseType: model.UserResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserSettingsChangeProfilePictureRequest
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
		if err := c.model.UserManager.UpdateFields(context, user.ID, user); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change profile picture failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := c.model.UserManager.GetByID(context, user.ID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change profile picture failed: get updated user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated user: " + err.Error()})
		}
		if err := c.userToken.SetUser(context, ctx, updatedUser); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change profile picture failed: set user token error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user token: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Profile picture changed for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, c.model.UserManager.ToModel(updatedUser))
	})

	// Change user's general profile settings
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/profile/general",
		Method:       "PUT",
		Note:         "Changes the user's general profile settings.",
		RequestType:  model.UserSettingsChangeGeneralRequest{},
		ResponseType: model.UserResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserSettingsChangeGeneralRequest
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
		if err := c.model.UserManager.UpdateFields(context, user.ID, user); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change general profile failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := c.model.UserManager.GetByID(context, user.ID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change general profile failed: get updated user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated user: " + err.Error()})
		}
		if err := c.userToken.SetUser(context, ctx, updatedUser); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change general profile failed: set user token error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user token: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "General profile changed for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, c.model.UserManager.ToModel(updatedUser))
	})
}
