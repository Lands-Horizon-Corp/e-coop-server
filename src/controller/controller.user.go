package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/cooperative_tokens"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) UserController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/authentication/current",
		Method:   "GET",
		Response: "TUser",
	}, func(ctx echo.Context) error {
		context := context.Background()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, model.CurrentUserResponse{
			UserID:           user.ID,
			User:             c.model.UserManager.ToModel(user),
			UserOrganization: nil,
		})
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/authentication/current-logged-in-accounts",
		Note:     "Current Logged In User: this is used to get the current logged in user in other apps/browsers. this is used for the frontend to get the current user.",
		Method:   "GET",
		Response: "TUser",
	}, func(ctx echo.Context) error {
		context := context.Background()
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
		Route:    "/authentication/login",
		Method:   "POST",
		Request:  "ISignInRequest",
		Response: "TUser",
	}, func(ctx echo.Context) error {
		context := context.Background()
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
		context := context.Background()
		c.userToken.CSRF.ClearCSRF(context, ctx)
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/authentication/register",
		Method:   "POST",
		Request:  "ISignUpRequest",
		Response: "TUser",
	}, func(ctx echo.Context) error {
		context := context.Background()
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
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/verify-reset-link/:reset_id",
		Method: "GET",
		Note:   "Verify Reset Link: this is the link that is sent to the user to reset their password. this will verify if the link is valid.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/change-password/:reset_id",
		Method:  "POST",
		Request: "IChangePasswordRequest",
		Note:    "Change Password: this is the link that is sent to the user to reset their password. this will change the user's password.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/apply-contact-number",
		Method: "POST",
		Note:   "Apply Contact Number: this is used to send OTP for contact number verification.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-contact-number",
		Method:  "POST",
		Request: "IVerifyContactNumberRequest",
		Note:    "Verify Contact Number: this is used to verify the OTP sent to the user's new contact number.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/apply-email",
		Method: "POST",
		Note:   "Apply Email: this is used to send OTP for email verification.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-email",
		Method:  "POST",
		Request: "IVerifyEmailRequest",
		Note:    "Verify Email: this is used to verify the OTP sent to the user's new email.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/verify-with-email",
		Method: "POST",
		Note:   "Verify with Email: this is used to verify the user's email by sending OTP to email.  [for preceeding protected self actions]",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-with-email-confirmation",
		Method:  "POST",
		Request: "6 digit OTP",
		Note:    "Verify with Email Confirmation: this is used to confirm the OTP sent to the user's email.  [for preceeding protected self actions]",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/verify-with-contact",
		Method: "POST",
		Note:   "Verify with Contact: this is used to verify the user's contact number by sending OTP.  [for preceeding protected self actions]",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-with-contact-confirmation",
		Method:  "POST",
		Request: "6 digit OTP",
		Note:    "Verify with Contact Confirmation: this is used to confirm the OTP sent to the user's contact number.  [for preceeding protected self actions]",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-with-password",
		Method:  "POST",
		Request: "password & password confirmation",
		Note:    "Verify with Password: this is used to verify the user's password. [for preceeding protected self actions]",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/profile/password",
		Method:  "PUT",
		Request: "IChangePasswordRequest",
		Note:    "Change Password: this is used to change the user's password.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/profile/email",
		Method:   "PUT",
		Request:  "IChangeEmailRequest",
		Response: "TUser",
		Note:     "Change Email: this is used to change the user's email.",
	}, func(ctx echo.Context) error {
		return nil
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/profile/username",
		Method:   "PUT",
		Request:  "IChangeUsernameRequest",
		Response: "TUser",
		Note:     "Change Username: this is used to change the user's username.",
	}, func(ctx echo.Context) error {
		return nil
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/profile/contact-number",
		Method:   "PUT",
		Request:  "IChangeContactNumberRequest",
		Response: "TUser",
		Note:     "Change Contact Number: this is used to change the user's contact number.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/profile/profile-picture",
		Method:   "PUT",
		Request:  "IUserSettingsPhotoUpdateRequest",
		Response: "TUser",
		Note:     "Change Profile Picture: this is used to change the user's profile picture.",
	}, func(ctx echo.Context) error {
		context := context.Background()

		// Bind the request body to UserSettingsChangeProfilePictureRequest struct
		var req model.UserSettingsChangeProfilePictureRequest
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

		// Check if the new media ID is the same as the current one
		if user.MediaID == req.MediaID {
			return echo.NewHTTPError(http.StatusBadRequest, "media ID is the same as the current one")
		}

		// Delete the current profile picture if it exists
		if user.MediaID != nil {
			if err := c.model.MediaDelete(context, *user.MediaID); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete current media: "+err.Error())
			}
		}

		// Update the user's media ID with the new one
		user.MediaID = req.MediaID

		// Update the user in the database
		if err := c.model.UserManager.Update(context, user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}

		updatedUser, err := c.model.UserManager.GetByID(context, user.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		// Set the updated user in the token
		if err := c.userToken.SetUser(context, ctx, user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to set user token: "+err.Error())
		}

		// Return the updated user model in the response
		return ctx.JSON(http.StatusOK, c.model.UserManager.ToModel(updatedUser))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/profile/general",
		Method:   "PUT",
		Request:  "IUserSettingsGeneralRequest",
		Response: "TUser",
		Note:     "Change General Settings: this is used to change the user's general settings.",
	}, func(ctx echo.Context) error {
		context := context.Background()
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
		if err := c.model.UserManager.Update(context, user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		if err := c.userToken.SetUser(context, ctx, user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to set user token: "+err.Error())
		}
		return nil
	})
}
