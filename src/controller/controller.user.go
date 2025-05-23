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
		Note:   "Verify with Email: this is used to verify the user's email by sending OTP to email.",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-with-email-confirmation",
		Method:  "POST",
		Request: "6 digit OTP",
		Note:    "Verify with Email Confirmation: this is used to confirm the OTP sent to the user's email.",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/authentication/verify-with-contact",
		Method: "POST",
		Note:   "Verify with Contact: this is used to verify the user's contact number by sending OTP.",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-with-contact-confirmation",
		Method:  "POST",
		Request: "6 digit OTP",
		Note:    "Verify with Contact Confirmation: this is used to confirm the OTP sent to the user's contact number.",
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/authentication/verify-with-password",
		Method:  "POST",
		Request: "password & password confirmation",
		Note:    "Verify with Password: this is used to verify the user's password.",
	}, func(c echo.Context) error {
		return nil
	})

}
