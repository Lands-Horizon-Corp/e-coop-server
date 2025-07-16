package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/cooperative_tokens"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) UserController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/user/:user_id",
		Method:   "GET",
		Response: "TUserRating[]",
		Note:     "Returns all user ratings given by the specified user (rater).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userId, err := horizon.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid user ID")
		}
		userRating, err := c.model.UserManager.GetByIDRaw(context, *userId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, userRating)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/authentication/current",
		Method:   "GET",
		Response: "TUser",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.userOrganizationToken.Token.CleanToken(context, ctx)
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

	req.RegisterRoute(horizon.Route{
		Route:    "/authentication/current-logged-in-accounts",
		Note:     "Current Logged In User: this is used to get the current logged in user in other apps/browsers. this is used for the frontend to get the current user.",
		Method:   "GET",
		Response: "ILoggedInUser",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		_, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		loggedIn, err := c.userToken.CSRF.GetLoggedInUsers(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		var resp cooperative_tokens.UserCSRFResponse
		loggedInPtrs := make([]*cooperative_tokens.UserCSRF, len(loggedIn))
		for i := range loggedIn {
			loggedInPtrs[i] = &loggedIn[i]
		}
		responses := resp.UserCSRFModels(loggedInPtrs)
		return ctx.JSON(http.StatusOK, responses)
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/current-logged-in-accounts/logout",
		Method: "POST",
		Note:   "Logout all users including itself",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		_, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		if err := c.userToken.CSRF.LogoutOtherDevices(context, ctx); err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/authentication/login",
		Method:   "POST",
		Request:  "ISignInRequest",
		Response: "TUser",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserLoginRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Validate the request using the validator service
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Find user by email
		user, err := c.model.GetUserByIdentifier(context, req.Key)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		}

		// Verify the password
		valid, err := c.provider.Service.Security.VerifyPassword(context, user.Password, req.Password)
		if err != nil || !valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		}

		// Set user token after successful login
		if err := c.userToken.SetUser(context, ctx, user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to set user token: "+err.Error())
		}

		return ctx.JSON(http.StatusOK, model.CurrentUserResponse{
			UserID: user.ID,
			User:   c.model.UserManager.ToModel(user),
		})
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/logout",
		Method: "POST",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		c.userToken.CSRF.ClearCSRF(context, ctx)
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/authentication/register",
		Method:   "POST",
		Request:  "ISignUpRequest",
		Response: "TUser",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.UserManager.Validate(ctx)
		if err != nil {
			return err
		}
		hashedPwd, err := c.provider.Service.Security.HashPassword(context, req.Password)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
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
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could not register user: %v", err))
		}
		if err := c.userToken.SetUser(context, ctx, user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to set user token: "+err.Error())
		}
		return ctx.JSON(http.StatusOK, model.CurrentUserResponse{
			UserID: user.ID,
			User:   c.model.UserManager.ToModel(user),
		})
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/authentication/forgot-password",
		Method:   "POST",
		Request:  "IForgotPasswordRequest",
		Response: "TUser",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserForgotPasswordRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		// Validate the request using the validator service
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		user, err := c.model.GetUserByIdentifier(context, req.Key)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "no account found with those details")
		}

		token, err := c.provider.Service.Security.GenerateUUIDv5(context, user.Password)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "invalid generating token")
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
			return echo.NewHTTPError(http.StatusNotFound, "failed sending email. please try again later.")
		}
		if err := c.provider.Service.Cache.Set(context, token, user.ID, 10*time.Minute); err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "invalid storing token")
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/verify-reset-link/:reset_id",
		Method: "GET",
		Note:   "Verify Reset Link: this is the link that is sent to the user to reset their password. this will verify if the link is valid.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		resetID := ctx.Param("reset_id")
		if resetID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "reset ID is required")
		}
		userID, err := c.provider.Service.Cache.Get(context, resetID)
		if err != nil || userID == nil {
			return echo.NewHTTPError(http.StatusNotFound, "Reset link is invalid or expired")
		}
		userId, err := uuid.Parse(string(userID))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid ID")
		}
		_, err = c.model.UserManager.GetByID(context, userId)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "User not found for reset token")
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/change-password/:reset_id",
		Method:  "POST",
		Request: "IChangePasswordRequest",
		Note:    "Change Password: this is the link that is sent to the user to reset their password. this will change the user's password.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserChangePasswordRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		resetID := ctx.Param("reset_id")
		if resetID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "reset ID is required")
		}
		userID, err := c.provider.Service.Cache.Get(context, resetID)
		if err != nil || userID == nil {
			return echo.NewHTTPError(http.StatusNotFound, "Reset link is invalid or expired")
		}
		userId, err := uuid.Parse(string(userID))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid ID")
		}
		user, err := c.model.UserManager.GetByID(context, userId)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "User not found for reset token")
		}
		hashedPwd, err := c.provider.Service.Security.HashPassword(context, req.NewPassword)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
		}
		user.Password = hashedPwd
		if err := c.model.UserManager.UpdateFields(context, user.ID, user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		if err := c.provider.Service.Cache.Delete(context, resetID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete token from cache: "+err.Error())
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/apply-contact-number",
		Method: "POST",
		Note:   "Apply Contact Number: this is used to send OTP for contact number verification.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s-%s", user.Password, user.ContactNumber)
		otp, err := c.provider.Service.OTP.Generate(context, key)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Fail sending OTP from user contact number: "+err.Error())
		}
		if err := c.provider.Service.SMS.Send(context, horizon.SMSRequest{
			To:   user.ContactNumber,
			Body: "Lands Horizon: Hello {{.name}} Please dont share this to someone else to protect your account and privacy. This is your OTP:{{.otp}}",
			Vars: map[string]string{
				"otp":  otp,
				"name": *user.FirstName,
			},
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to send otp cache: "+err.Error())
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-contact-number",
		Method:  "POST",
		Request: "IVerifyContactNumberRequest",
		Note:    "Verify Contact Number: this is used to verify the OTP sent to the user's new contact number.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserVerifyContactNumberRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s-%s", user.Password, user.ContactNumber)
		ok, err := c.provider.Service.OTP.Verify(context, key, req.OTP)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify token: "+err.Error())
		}
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "Invalid OTP ")
		}
		if err := c.provider.Service.OTP.Revoke(context, key); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify token: "+err.Error())
		}
		user.IsContactVerified = true
		if err := c.model.UserManager.UpdateFields(context, user.ID, user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		updatedUser, err := c.model.UserManager.GetByID(context, user.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		if err := c.userToken.SetUser(context, ctx, updatedUser); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to set user token: "+err.Error())
		}
		return ctx.JSON(http.StatusOK, c.model.UserManager.ToModel(updatedUser))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/apply-email",
		Method: "POST",
		Note:   "Apply Email: this is used to send OTP for email verification.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s-%s", user.Password, user.Email)
		otp, err := c.provider.Service.OTP.Generate(context, key)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Fail sending OTP from user contact number: "+err.Error())
		}
		if err := c.provider.Service.SMTP.Send(context, horizon.SMTPRequest{
			To:      user.Email,
			Body:    "templates/email-otp.html",
			Subject: "Email Verification: Lands Horizon",
			Vars: map[string]string{
				"otp": otp,
			},
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to send otp cache: "+err.Error())
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-email",
		Method:  "POST",
		Request: "IVerifyEmailRequest",
		Note:    "Verify Email: this is used to verify the OTP sent to the user's new email.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserVerifyEmailRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s-%s", user.Password, user.Email)
		ok, err := c.provider.Service.OTP.Verify(context, key, req.OTP)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify token: "+err.Error())
		}
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "Invalid OTP ")
		}
		if err := c.provider.Service.OTP.Revoke(context, key); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify token: "+err.Error())
		}
		user.IsEmailVerified = true
		if err := c.model.UserManager.UpdateFields(context, user.ID, user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		updatedUser, err := c.model.UserManager.GetByID(context, user.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		if err := c.userToken.SetUser(context, ctx, updatedUser); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to set user token: "+err.Error())
		}
		return ctx.JSON(http.StatusOK, c.model.UserManager.ToModel(updatedUser))
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-with-password",
		Method:  "POST",
		Request: "password & password confirmation",
		Note:    "Verify with Password: this is used to verify the user's password. [for preceeding protected self actions]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req model.UserVerifyWithPasswordRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Validate the request using the validator service
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Get the current user from the context
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get current user: "+err.Error())
		}
		// Verify the password
		valid, err := c.provider.Service.Security.VerifyPassword(context, user.Password, req.Password)
		if err != nil || !valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		}

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/profile/password",
		Method:  "PUT",
		Request: "IChangePasswordRequest",
		Note:    "Change Password: this is used to change the user's password.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserSettingsChangePasswordRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get current user: "+err.Error())
		}
		valid, err := c.provider.Service.Security.VerifyPassword(context, user.Password, req.OldPassword)
		if err != nil || !valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		}
		hashedPwd, err := c.provider.Service.Security.HashPassword(context, req.NewPassword)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
		}
		user.Password = hashedPwd
		if err := c.model.UserManager.UpdateFields(context, user.ID, user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		updatedUser, err := c.model.UserManager.GetByID(context, user.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		if err := c.userToken.SetUser(context, ctx, user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to set user token: "+err.Error())
		}
		return ctx.JSON(http.StatusOK, c.model.UserManager.ToModel(updatedUser))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/profile/profile-picture",
		Method:   "PUT",
		Request:  "IUserSettingsPhotoUpdateRequest",
		Response: "TUser",
		Note:     "Change Profile Picture: this is used to change the user's profile picture.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		// Bind the request body
		var req model.UserSettingsChangeProfilePictureRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Validate the request
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		// Get current user
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}

		// If the same MediaID is submitted, reject
		if user.MediaID == req.MediaID {
			return echo.NewHTTPError(http.StatusBadRequest, "media ID is the same as the current one")
		}

		user.MediaID = req.MediaID
		if err := c.model.UserManager.UpdateFields(context, user.ID, user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		updatedUser, err := c.model.UserManager.GetByID(context, user.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		if err := c.userToken.SetUser(context, ctx, updatedUser); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to set user token: "+err.Error())
		}
		return ctx.JSON(http.StatusOK, c.model.UserManager.ToModel(updatedUser))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/profile/general",
		Method:   "PUT",
		Request:  "IUserSettingsGeneralRequest",
		Response: "TUser",
		Note:     "Change General Settings: this is used to change the user's general settings.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.UserSettingsChangeGeneralRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
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
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}

		updatedUser, err := c.model.UserManager.GetByID(context, user.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		if err := c.userToken.SetUser(context, ctx, updatedUser); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to set user token: "+err.Error())
		}
		return ctx.JSON(http.StatusOK, c.model.UserManager.ToModel(updatedUser))
	})
}

func (c *Controller) UserRatingController() {
	req := c.provider.Service.Request

	// Get all user ratings made by the specified user (rater)
	req.RegisterRoute(horizon.Route{
		Route:    "/user-rating/user-rater/:user_id",
		Method:   "GET",
		Response: "TUserRating[]",
		Note:     "Returns all user ratings given by the specified user (rater).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userId, err := horizon.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid user ID")
		}
		userRating, err := c.model.GetUserRater(context, *userId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserRatingManager.ToModels(userRating))
	})

	// Get all user ratings received by the specified user (ratee)
	req.RegisterRoute(horizon.Route{
		Route:    "/user-rating/user-ratee/:user_id",
		Method:   "GET",
		Response: "TUserRating[]",
		Note:     "Returns all user ratings received by the specified user (ratee).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userId, err := horizon.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid user ID")
		}
		userRating, err := c.model.GetUserRatee(context, *userId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserRatingManager.ToModels(userRating))
	})

	// Get a specific user rating by its ID
	req.RegisterRoute(horizon.Route{
		Route:    "/user-rating/:user_rating_id",
		Method:   "GET",
		Response: "TUserRating",
		Note:     "Returns a specific user rating by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userRatingId, err := horizon.EngineUUIDParam(ctx, "user_rating_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid rating ID")
		}
		userRating, err := c.model.UserRatingManager.GetByID(context, *userRatingId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserRatingManager.ToModel(userRating))
	})

	// Get all user ratings for the current user's branch
	req.RegisterRoute(horizon.Route{
		Route:    "/user-rating/branch",
		Method:   "GET",
		Response: "TUserRating[]",
		Note:     "Returns all user ratings in the current user's active branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		userRatig, err := c.model.UserRatingCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserRatingManager.ToModels(userRatig))
	})

	// Create a new user rating in the current branch
	req.RegisterRoute(horizon.Route{
		Route:    "/user-rating",
		Method:   "POST",
		Response: "TUserRating",
		Request:  "TUserRating",
		Note:     "Creates a new user rating in the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.UserRatingManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}

		userRating := &model.UserRating{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			RateeUserID:    req.RateeUserID,
			RaterUserID:    req.RaterUserID,
			Rate:           req.Rate,
			Remark:         req.Remark,
		}

		if err := c.model.UserRatingManager.Create(context, userRating); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.UserRatingManager.ToModel(userRating))
	})

	// Delete a user rating by its ID
	req.RegisterRoute(horizon.Route{
		Route:  "/user-rating/:user_rating_id",
		Method: "DELETE",
		Note:   "Deletes a user rating by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userRatingId, err := horizon.EngineUUIDParam(ctx, "user_rating_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid rating ID")
		}
		if err := c.model.UserRatingManager.DeleteByID(context, *userRatingId); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
