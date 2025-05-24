package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
		user, err := c.userToken.CurrentUser(ctx)
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
		Route:    "/authentication/login",
		Method:   "POST",
		Request:  "ISignInRequest",
		Response: "TUser",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/logout",
		Method: "POST",
	}, func(ctx echo.Context) error {
		return nil
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

		claim := cooperative_tokens.UserCSRF{
			UserID:        user.ID.String(),
			Email:         user.Email,
			ContactNumber: user.ContactNumber,
			Password:      user.Password,
			Username:      user.UserName,
		}
		if err := c.userToken.CSRF.SetCSRF(context, ctx, claim, 8*time.Hour); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
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
		return nil
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/profile/general",
		Method:   "PUT",
		Request:  "IUserSettingsGeneralRequest",
		Response: "TUser",
		Note:     "Change General Settings: this is used to change the user's general settings.",
	}, func(ctx echo.Context) error {
		return nil
	})
}
