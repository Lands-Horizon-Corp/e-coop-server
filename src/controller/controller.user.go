package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) UserController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/authentication/current",
		Method:   "GET",
		Response: "TUser",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/authentication/login",
		Method:   "POST",
		Request:  "ISignInRequest",
		Response: "TUser",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/logout",
		Method: "POST",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/authentication/register",
		Method:   "POST",
		Request:  "ISignUpRequest",
		Response: "TUser",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/authentication/forgot-password",
		Method:   "POST",
		Request:  "IForgotPasswordRequest",
		Response: "TUser",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/verify-reset-link/:reset_id",
		Method: "GET",
		Note:   "Verify Reset Link: this is the link that is sent to the user to reset their password. this will verify if the link is valid.",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/change-password/:reset_id",
		Method:  "POST",
		Request: "IChangePasswordRequest",
		Note:    "Change Password: this is the link that is sent to the user to reset their password. this will change the user's password.",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/apply-contact-number",
		Method: "POST",
		Note:   "Apply Contact Number: this is used to send OTP for contact number verification.",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-contact-number",
		Method:  "POST",
		Request: "IVerifyContactNumberRequest",
		Note:    "Verify Contact Number: this is used to verify the OTP sent to the user's new contact number.",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/apply-email",
		Method: "POST",
		Note:   "Apply Email: this is used to send OTP for email verification.",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-email",
		Method:  "POST",
		Request: "IVerifyEmailRequest",
		Note:    "Verify Email: this is used to verify the OTP sent to the user's new email.",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/verify-with-email",
		Method: "POST",
		Note:   "Verify with Email: this is used to verify the user's email by sending OTP to email.  [for preceeding protected self actions]",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-with-email-confirmation",
		Method:  "POST",
		Request: "6 digit OTP",
		Note:    "Verify with Email Confirmation: this is used to confirm the OTP sent to the user's email.  [for preceeding protected self actions]",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/verify-with-contact",
		Method: "POST",
		Note:   "Verify with Contact: this is used to verify the user's contact number by sending OTP.  [for preceeding protected self actions]",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-with-contact-confirmation",
		Method:  "POST",
		Request: "6 digit OTP",
		Note:    "Verify with Contact Confirmation: this is used to confirm the OTP sent to the user's contact number.  [for preceeding protected self actions]",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-with-password",
		Method:  "POST",
		Request: "password & password confirmation",
		Note:    "Verify with Password: this is used to verify the user's password. [for preceeding protected self actions]",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/profile/password",
		Method:  "PUT",
		Request: "IChangePasswordRequest",
		Note:    "Change Password: this is used to change the user's password.",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/profile/email",
		Method:   "PUT",
		Request:  "IChangeEmailRequest",
		Response: "TUser",
		Note:     "Change Email: this is used to change the user's email.",
	}, func(c echo.Context) error {
		return nil
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/profile/username",
		Method:   "PUT",
		Request:  "IChangeUsernameRequest",
		Response: "TUser",
		Note:     "Change Username: this is used to change the user's username.",
	}, func(c echo.Context) error {
		return nil
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/profile/contact-number",
		Method:   "PUT",
		Request:  "IChangeContactNumberRequest",
		Response: "TUser",
		Note:     "Change Contact Number: this is used to change the user's contact number.",
	}, func(c echo.Context) error {
		return nil
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/profile/profile-picture",
		Method:   "PUT",
		Request:  "IUserSettingsPhotoUpdateRequest",
		Response: "TUser",
		Note:     "Change Profile Picture: this is used to change the user's profile picture.",
	}, func(c echo.Context) error {
		return nil
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/profile/general",
		Method:   "PUT",
		Request:  "IUserSettingsGeneralRequest",
		Response: "TUser",
		Note:     "Change General Settings: this is used to change the user's general settings.",
	}, func(c echo.Context) error {
		return nil
	})
}
