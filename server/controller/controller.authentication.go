package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) authenticationController() {
	req := c.provider.Service.Request
	// Returns the current authenticated user and their user organization, if any.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/authentication/current",
		Method:       "GET",
		ResponseType: core.CurrentUserResponse{},
		Note:         "Returns the current authenticated user and their user organization, if any.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.userOrganizationToken.ClearCurrentToken(context, ctx)
			return ctx.NoContent(http.StatusUnauthorized)
		}
		userOrganization, _ := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		var userOrg *core.UserOrganizationResponse
		if userOrganization != nil {
			userOrg = c.core.UserOrganizationManager.ToModel(userOrganization)
		}
		return ctx.JSON(http.StatusOK, core.CurrentUserResponse{
			UserID:           user.ID,
			User:             c.core.UserManager.ToModel(user),
			UserOrganization: userOrg,
		})
	})
	// Logout all users including itself for the session
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/authentication/current-logged-in-accounts/logout",
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

	// Returns all currently logged-in users for the session
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/authentication/current-logged-in-accounts",
		Note:         "Returns all currently logged-in users for the session.",
		Method:       "GET",
		ResponseType: tokens.UserCSRFResponse{},
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
		var resp tokens.UserCSRFResponse
		loggedInPtrs := make([]*tokens.UserCSRF, len(loggedIn))
		for i := range loggedIn {
			loggedInPtrs[i] = &loggedIn[i]
		}
		responses := resp.UserCSRFModels(loggedInPtrs)
		return ctx.JSON(http.StatusOK, responses)
	})
	// Authenticate user login
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/authentication/login",
		Method:       "POST",
		RequestType:  core.UserLoginRequest{},
		ResponseType: core.CurrentUserResponse{},
		Note:         "Authenticates a user and returns user details.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.UserLoginRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid login payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.core.GetUserByIdentifier(context, req.Key)
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
		return ctx.JSON(http.StatusOK, core.CurrentUserResponse{
			UserID: user.ID,
			User:   c.core.UserManager.ToModel(user),
		})
	})

	// Logout the current user
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/authentication/logout",
		Method: "POST",
		Note:   "Logs out the current user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "User logged out successfully",
			Module:      "User",
		})
		c.userToken.ClearCurrentCSRF(context, ctx)
		return ctx.NoContent(http.StatusNoContent)
	})

	// Register a new user
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/authentication/register",
		Method:       "POST",
		ResponseType: core.CurrentUserResponse{},
		RequestType:  core.UserRegisterRequest{},
		Note:         "Registers a new user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.UserManager.Validate(ctx)
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
		user := &core.User{
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
		if err := c.core.UserManager.Create(context, user); err != nil {
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
		return ctx.JSON(http.StatusOK, core.CurrentUserResponse{
			UserID: user.ID,
			User:   c.core.UserManager.ToModel(user),
		})
	})

	// Forgot password flow
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/authentication/forgot-password",
		Method:      "POST",
		RequestType: core.UserForgotPasswordRequest{},
		Note:        "Initiates forgot password flow and sends a reset link.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.UserForgotPasswordRequest
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
		user, err := c.core.GetUserByIdentifier(context, req.Key)
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
		Route:  "/api/v1/authentication/verify-reset-link/:reset_id",
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
		parsedUserID, err := uuid.Parse(string(userID))
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID: " + err.Error()})
		}
		_, err = c.core.UserManager.GetByID(context, parsedUserID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found for reset token: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
	// Change password using the reset link
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/authentication/change-password/:reset_id",
		Method:      "POST",
		RequestType: core.UserChangePasswordRequest{},
		Note:        "Changes the user's password using the reset link.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.UserChangePasswordRequest
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
		parsedUserID, err := uuid.Parse(string(userID))
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID: " + err.Error()})
		}
		user, err := c.core.UserManager.GetByID(context, parsedUserID)
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
		if err := c.core.UserManager.UpdateByID(context, user.ID, user); err != nil {
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
		Route:  "/api/v1/authentication/apply-contact-number",
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
		RequestType:  core.UserVerifyContactNumberRequest{},
		ResponseType: core.UserResponse{},
		Note:         "Verifies OTP for contact number verification.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.UserVerifyContactNumberRequest
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
		if err := c.core.UserManager.UpdateByID(context, user.ID, user); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := c.core.UserManager.GetByID(context, user.ID)
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
		return ctx.JSON(http.StatusOK, c.core.UserManager.ToModel(updatedUser))
	})

	// Send OTP for email verification
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/authentication/apply-email",
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

	// Verify user with password for self-protected actions
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/authentication/verify-with-password",
		Method:       "POST",
		Note:         "Verifies the user's password for protected self actions.",
		ResponseType: core.UserResponse{},
		RequestType:  core.UserVerifyWithPasswordRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.UserVerifyWithPasswordRequest
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

	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/authentication/verify-with-password/owner",
		Method: "POST",
		Note:   "Verifies the user's password for protected owner actions. (must be owner and inside a branch)",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.UserAdminPasswordVerificationRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid login payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		userOrganization, err := c.core.UserOrganizationManager.GetByID(context, req.UserOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		if userOrganization.UserType != core.UserOrganizationTypeOwner {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden: User is not an owner"})
		}
		valid, err := c.provider.Service.Security.VerifyPassword(context, userOrganization.User.Password, req.Password)
		if err != nil || !valid {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Owner password verification successful for user organization: " + userOrg.ID.String(),
			Module:      "User",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Verify OTP for email
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/authentication/verify-email",
		Method:       "POST",
		Note:         "Verifies OTP for email verification.",
		ResponseType: core.UserResponse{},
		RequestType:  core.UserVerifyEmailRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.UserVerifyEmailRequest
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
		if err := c.core.UserManager.UpdateByID(context, user.ID, user); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := c.core.UserManager.GetByID(context, user.ID)
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
		return ctx.JSON(http.StatusOK, c.core.UserManager.ToModel(updatedUser))
	})

}
