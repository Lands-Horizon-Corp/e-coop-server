package user

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func AuthenticationController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/authentication/current",
		Method:       "GET",
		ResponseType: types.CurrentUserResponse{},
		Note:         "Returns the current authenticated user and their user organization, if any.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			event.ClearCurrentToken(context, service, ctx)
			return ctx.NoContent(http.StatusUnauthorized)
		}
		userOrganization, _ := event.CurrentUserOrganization(context, service, ctx)
		var memberProfile *types.MemberProfileResponse
		var userOrg *types.UserOrganizationResponse
		if userOrganization != nil {
			userOrg = core.UserOrganizationManager(service).ToModel(userOrganization)
			if userOrganization.UserType == types.UserOrganizationTypeMember || userOrganization.UserType == types.UserOrganizationTypeOwner {
				memberProfile, _ = core.MemberProfileManager(service).FindOneRaw(context, &types.MemberProfile{
					UserID:         &userOrg.UserID,
					BranchID:       *userOrg.BranchID,
					OrganizationID: userOrg.OrganizationID,
				})
			}

		}

		return ctx.JSON(http.StatusOK, types.CurrentUserResponse{
			UserID:           user.ID,
			User:             core.UserManager(service).ToModel(user),
			MemberProfile:    memberProfile,
			UserOrganization: userOrg,
		})
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/authentication/login",
		Method:       "POST",
		RequestType:  types.UserLoginRequest{},
		ResponseType: types.CurrentUserResponse{},
		Note:         "Authenticates a user and returns user details.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.UserLoginRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid login payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := core.GetUserByIdentifier(context, service, req.Key)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials: " + err.Error()})
		}
		valid, err := service.Security.VerifyPassword(user.Password, req.Password)
		if err != nil || !valid {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		if err := event.SetUser(context, service, ctx, user); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user token: " + err.Error()})
		}
		var memberProfile *types.MemberProfileResponse
		var userOrg *types.UserOrganizationResponse
		org, ok := event.GetOrganization(service, ctx)
		if ok {
			userOrganization, err := core.UserOrganizationManager(service).FindOne(context, &types.UserOrganization{
				UserID:         user.ID,
				OrganizationID: org.ID,
				UserType:       types.UserOrganizationTypeMember,
			})
			if err == nil && userOrganization.ApplicationStatus == "accepted" {
				if err := event.SetUserOrganization(context, service, ctx, userOrganization); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user organization: " + err.Error()})
				}
				userOrg = core.UserOrganizationManager(service).ToModel(userOrganization)
				if userOrganization.UserType == types.UserOrganizationTypeMember {
					memberProfile, _ = core.MemberProfileManager(service).FindOneRaw(context, &types.MemberProfile{
						UserID:         &userOrg.UserID,
						BranchID:       *userOrg.BranchID,
						OrganizationID: userOrg.OrganizationID,
					})
				}
			}
		}

		return ctx.JSON(http.StatusOK, types.CurrentUserResponse{
			UserID:           user.ID,
			User:             core.UserManager(service).ToModel(user),
			MemberProfile:    memberProfile,
			UserOrganization: userOrg,
		})
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
		Route:        "/api/v1/authentication/current-logged-in-accounts",
		Note:         "Returns all currently logged-in users for the session.",
		Method:       "GET",
		ResponseType: event.UserCSRFResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		_, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		loggedIn, err := event.LoggedInUsers(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get logged in users: " + err.Error()})
		}
		var resp event.UserCSRFResponse
		loggedInPtrs := make([]*event.UserCSRF, len(loggedIn))
		for i := range loggedIn {
			loggedInPtrs[i] = &loggedIn[i]
		}
		responses := resp.UserCSRFModels(loggedInPtrs)
		return ctx.JSON(http.StatusOK, responses)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/authentication/logout",
		Method: "POST",
		Note:   "Logs out the current user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "User logged out successfully",
			Module:      "User",
		})
		event.ClearCurrentCSRF(context, service, ctx)
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/authentication/register",
		Method:       "POST",
		ResponseType: types.CurrentUserResponse{},
		RequestType:  types.UserRegisterRequest{},
		Note:         "Registers a new user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.UserManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Register failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		hashedPwd, err := service.Security.HashPassword(req.Password)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Register failed: hash password error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password: " + err.Error()})
		}
		user := &types.User{
			Email:             req.Email,
			Password:          hashedPwd,
			Birthdate:         req.Birthdate,
			Username:          req.Username,
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
		if err := core.UserManager(service).Create(context, user); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Register failed: create user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not register user: " + err.Error()})
		}
		if err := event.SetUser(context, service, ctx, user); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Register failed: failed to set user token: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user token: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "User registered successfully: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, types.CurrentUserResponse{
			UserID: user.ID,
			User:   core.UserManager(service).ToModel(user),
		})
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/authentication/forgot-password",
		Method:      "POST",
		RequestType: types.UserForgotPasswordRequest{},
		Note:        "Initiates forgot password flow and sends a reset link.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.UserForgotPasswordRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Forgot password failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid forgot password payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Forgot password failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := core.GetUserByIdentifier(context, service, req.Key)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Forgot password failed: user not found: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No account found with those details: " + err.Error()})
		}
		token, err := service.Security.GenerateUUIDv5(user.Password)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Forgot password failed: generate token error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Error generating reset token: " + err.Error()})
		}
		fallbackStr := service.Config.AppClientURL
		if err := service.SMTP.Send(context, horizon.SMTPRequest{
			To:      req.Key,
			Subject: "Forgot Password: Lands Horizon",
			Body:    "templates/email-change-password.html",
			Vars: map[string]string{
				"name":      user.FullName,
				"eventLink": fallbackStr + "/auth/password-reset/" + token,
			},
		}); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Forgot password failed: send email error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed sending email: " + err.Error()})
		}
		if err := service.Cache.Set(context, token, user.ID, 10*time.Minute); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Forgot password failed: cache set error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed storing token: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Forgot password initiated for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/authentication/verify-reset-link/:reset_id",
		Method: "GET",
		Note:   "Verifies if the reset password link is valid.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		resetID := ctx.Param("reset_id")
		if resetID == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Reset ID is required"})
		}
		userID, err := service.Cache.Get(context, resetID)
		if err != nil || userID == nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Reset link is invalid or expired"})
		}
		parsedUserID, err := uuid.Parse(string(userID))
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID: " + err.Error()})
		}
		_, err = core.UserManager(service).GetByID(context, parsedUserID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found for reset token: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/authentication/change-password/:reset_id",
		Method:      "POST",
		RequestType: types.UserChangePasswordRequest{},
		Note:        "Changes the user's password using the reset link.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.UserChangePasswordRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid change password payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		userID, err := service.Cache.Get(context, resetID)
		if err != nil || userID == nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Reset link is invalid or expired"})
		}
		parsedUserID, err := uuid.Parse(string(userID))
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID: " + err.Error()})
		}
		user, err := core.UserManager(service).GetByID(context, parsedUserID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found for reset token: " + err.Error()})
		}
		hashedPwd, err := service.Security.HashPassword(req.NewPassword)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password failed: hash password error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password: " + err.Error()})
		}
		user.Password = hashedPwd
		if err := core.UserManager(service).UpdateByID(context, user.ID, user); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user password: " + err.Error()})
		}
		if err := service.Cache.Delete(context, resetID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change password failed: delete cache error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete token from cache: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Password changed successfully for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/authentication/apply-contact-number",
		Method: "POST",
		Note:   "Sends OTP for contact number verification.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		key := fmt.Sprintf("%s-%s", user.Password, user.ContactNumber)
		otp, err := service.OTP.Generate(context, key)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Apply contact number failed: generate OTP error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate OTP: " + err.Error()})
		}
		if err := service.SMS.Send(context, horizon.SMSRequest{
			To:   user.ContactNumber,
			Body: "Lands Horizon: Hello {{.name}} Please dont share this to someone else to protect your account and privacy. This is your OTP:{{.otp}}",
			Vars: map[string]string{
				"otp":  otp,
				"name": *user.FirstName,
			},
		}); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Apply contact number failed: send SMS error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send OTP SMS: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "OTP sent for contact number verification: " + user.ContactNumber,
			Module:      "User",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/authentication/verify-contact-number",
		Method:       "POST",
		RequestType:  types.UserVerifyContactNumberRequest{},
		ResponseType: types.UserResponse{},
		Note:         "Verifies OTP for contact number verification.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.UserVerifyContactNumberRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid verify contact number payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		key := fmt.Sprintf("%s-%s", user.Password, user.ContactNumber)
		ok, err := service.OTP.Verify(context, key, req.OTP)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: verify OTP error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to verify OTP: " + err.Error()})
		}
		if !ok {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: invalid OTP",
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid OTP"})
		}
		if err := service.OTP.Revoke(context, key); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: revoke OTP error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to revoke OTP: " + err.Error()})
		}
		user.IsContactVerified = true
		if err := core.UserManager(service).UpdateByID(context, user.ID, user); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := core.UserManager(service).GetByID(context, user.ID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: get updated user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated user: " + err.Error()})
		}
		if err := event.SetUser(context, service, ctx, updatedUser); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify contact number failed: set user token error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user token: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Contact number verified for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, core.UserManager(service).ToModel(updatedUser))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/authentication/apply-email",
		Method: "POST",
		Note:   "Sends OTP for email verification.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		key := fmt.Sprintf("%s-%s", user.Password, user.Email)
		otp, err := service.OTP.Generate(context, key)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Apply email failed: generate OTP error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate OTP: " + err.Error()})
		}
		if err := service.SMTP.Send(context, horizon.SMTPRequest{
			To:      user.Email,
			Body:    "templates/email-otp.html",
			Subject: "Email Verification: Lands Horizon",
			Vars: map[string]string{
				"otp": otp,
			},
		}); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Apply email failed: send email error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send OTP email: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "OTP sent for email verification: " + user.Email,
			Module:      "User",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/authentication/verify-with-password",
		Method:       "POST",
		Note:         "Verifies the user's password for protected self actions.",
		ResponseType: types.UserResponse{},
		RequestType:  types.UserVerifyWithPasswordRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.UserVerifyWithPasswordRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid verify with password payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}
		valid, err := service.Security.VerifyPassword(user.Password, req.Password)
		if err != nil || !valid {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/authentication/verify-with-password/owner",
		Method: "POST",
		Note:   "Verifies the user's password for protected owner actions. (must be owner and inside a branch)",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.UserAdminPasswordVerificationRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid login payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, req.UserOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		if userOrganization.UserType != types.UserOrganizationTypeOwner {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden: User is not an owner"})
		}
		valid, err := service.Security.VerifyPassword(userOrganization.User.Password, req.Password)
		if err != nil || !valid {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Owner password verification successful for user organization: " + userOrg.ID.String(),
			Module:      "User",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/authentication/verify-email",
		Method:       "POST",
		Note:         "Verifies OTP for email verification.",
		ResponseType: types.UserResponse{},
		RequestType:  types.UserVerifyEmailRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.UserVerifyEmailRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: invalid payload: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid verify email payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		key := fmt.Sprintf("%s-%s", user.Password, user.Email)
		ok, err := service.OTP.Verify(context, key, req.OTP)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: verify OTP error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to verify OTP: " + err.Error()})
		}
		if !ok {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: invalid OTP",
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid OTP"})
		}
		if err := service.OTP.Revoke(context, key); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: revoke OTP error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to revoke OTP: " + err.Error()})
		}
		user.IsEmailVerified = true
		if err := core.UserManager(service).UpdateByID(context, user.ID, user); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: update user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		}
		updatedUser, err := core.UserManager(service).GetByID(context, user.ID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: get updated user error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated user: " + err.Error()})
		}
		if err := event.SetUser(context, service, ctx, updatedUser); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Verify email failed: set user token error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user token: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Email verified for user: " + user.ID.String(),
			Module:      "User",
		})
		return ctx.JSON(http.StatusOK, core.UserManager(service).ToModel(updatedUser))
	})

}
