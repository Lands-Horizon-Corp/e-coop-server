package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

func (c *Controller) UserCurrent(ctx echo.Context) error {
	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, model.CurrentUserResponse{
		UserID: user.ID,
		User:   c.model.UserModel(user),
	})
}

func (c *Controller) UserLogin(ctx echo.Context) error {
	req, err := c.model.UserLoginValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.user.ByIdentifier(req.Key)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if ok := c.authentication.VerifyPassword(user.Password, req.Password); !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if err := c.provider.SetUser(ctx, user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
	}
	return ctx.JSON(http.StatusOK, model.CurrentUserResponse{
		UserID: user.ID,
		User:   c.model.UserModel(user),
	})
}

func (c *Controller) UserLogout(ctx echo.Context) error {
	c.authentication.CleanToken(ctx)
	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) UserRegister(ctx echo.Context) error {
	req, err := c.model.UserRegisterValidate(ctx)
	if err != nil {
		return err
	}
	hashedPwd, err := c.authentication.Password(req.Password)
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
	if err := c.user.Manager.Create(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could not register user: %v", err))
	}
	if err := c.provider.SetUser(ctx, user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
	}

	return ctx.JSON(http.StatusOK, model.CurrentUserResponse{
		UserID: user.ID,
		User:   c.model.UserModel(user),
	})
}

func (c *Controller) UserForgotPassword(ctx echo.Context) error {
	req, err := c.model.UserForgotPasswordValidate(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "please provide a valid email or contact number")
	}
	user, err := c.user.ByIdentifier(req.Key)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "no account found with those details")
	}
	_, err = c.authentication.GenerateSMTPLink("/auth/password-reset/", horizon.Claim{
		ID:            user.ID.String(),
		Email:         user.Email,
		ContactNumber: user.ContactNumber,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send reset link, please try again later")
	}
	_, err = c.authentication.GenerateSMSLink("/auth/password-reset/", horizon.Claim{
		ID:            user.ID.String(),
		Email:         user.Email,
		ContactNumber: user.ContactNumber,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send reset link, please try again later")
	}
	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) UserVerifyResetLink(ctx echo.Context) error {
	idParam := ctx.Param("id")
	_, err := c.authentication.ValidateLink(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "link is not valid")
	}
	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) UserChangePassword(ctx echo.Context) error {
	idParam := ctx.Param("id")
	req, err := c.model.UserChangePasswordValidate(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}
	claim, err := c.authentication.ValidateLink(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "link is not valid or has expired")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}
	hashedPwd, err := c.authentication.Password(req.NewPassword)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash new password")
	}
	if err := c.user.Manager.UpdateFields(id, &model.User{
		Password:  hashedPwd,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}
	updatedUser, err := c.user.Manager.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}

	return ctx.JSON(http.StatusOK, c.model.UserModel(updatedUser))
}

func (c *Controller) UserApplyContactNumber(ctx echo.Context) error {
	claim, err := c.authentication.GetUserFromToken(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	if err := c.authentication.SendSMSOTP(*claim); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send OTP: "+err.Error())
	}
	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) UserVerifyContactNumber(ctx echo.Context) error {
	req, err := c.model.UserVerifyContactNumberValidate(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}
	claim, err := c.authentication.GetUserFromToken(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	valid := c.authentication.VerifySMSOTP(*claim, req.OTP)
	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired OTP")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}

	if err := c.user.Manager.UpdateFields(id, &model.User{
		IsContactVerified: true,
		UpdatedAt:         time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}

	updatedUser, err := c.user.Manager.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	return ctx.JSON(http.StatusOK, c.model.UserModel(updatedUser))
}

func (c *Controller) UserApplyEmail(ctx echo.Context) error {
	claim, err := c.authentication.GetUserFromToken(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	if err := c.authentication.SendSMTPOTP(*claim); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send email verification: "+err.Error())
	}
	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) UserVerifyEmail(ctx echo.Context) error {
	req, err := c.model.UserVerifyEmailValidate(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}
	claim, err := c.authentication.GetUserFromToken(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	valid := c.authentication.VerifySMTPOTP(*claim, req.OTP)
	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired email token")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}

	if err := c.user.Manager.UpdateFields(id, &model.User{
		IsEmailVerified: true,
		UpdatedAt:       time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}

	updatedUser, err := c.user.Manager.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}

	return ctx.JSON(http.StatusOK, c.model.UserModel(updatedUser))
}

func (c *Controller) UserVerifyWithEmail(ctx echo.Context) error {
	claim, err := c.authentication.GetUserFromToken(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	if err := c.authentication.SendSMTPOTP(*claim); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send email verification: "+err.Error())
	}
	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) UserVerifyWithEmailConfirmation(ctx echo.Context) error {
	req, err := c.model.UserVerifyWithEmailConfirmationValidate(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}
	claim, err := c.authentication.GetUserFromToken(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	valid := c.authentication.VerifySMTPOTP(*claim, req.OTP)
	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired email token")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}

	updatedUser, err := c.user.Manager.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}

	return ctx.JSON(http.StatusOK, c.model.UserModel(updatedUser))
}

func (c *Controller) UserVerifyWithContactNumber(ctx echo.Context) error {
	claim, err := c.authentication.GetUserFromToken(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	if err := c.authentication.SendSMSOTP(*claim); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send OTP: "+err.Error())
	}
	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) UserVerifyWithContactNumberConfirmation(ctx echo.Context) error {
	req, err := c.model.UserVerifyWithContactNumberConfirmationValidate(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}
	claim, err := c.authentication.GetUserFromToken(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	valid := c.authentication.VerifySMSOTP(*claim, req.OTP)
	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired OTP")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}
	updatedUser, err := c.user.Manager.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	return ctx.JSON(http.StatusOK, c.model.UserModel(updatedUser))
}

func (c *Controller) UserSettingsChangePassword(ctx echo.Context) error {
	req, err := c.model.UserSettingsChangePasswordValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}
	if ok := c.authentication.VerifyPassword(user.Password, req.Password); !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	hashedPwd, err := c.authentication.Password(req.NewPassword)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash new password")
	}

	if err := c.user.Manager.UpdateFields(user.ID, &model.User{
		Password:  hashedPwd,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}
	updatedUser, err := c.user.Manager.GetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	return ctx.JSON(http.StatusOK, c.model.UserModel(updatedUser))
}

func (c *Controller) UserSettingsChangeEmail(ctx echo.Context) error {
	req, err := c.model.UserSettingsChangeEmailValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}
	if ok := c.authentication.VerifyPassword(user.Password, req.Password); !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if err := c.user.Manager.UpdateFields(user.ID, &model.User{
		Email:           req.Email,
		IsEmailVerified: false,
		UpdatedAt:       time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}
	updatedUser, err := c.user.Manager.GetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	c.provider.SetUser(ctx, updatedUser)
	return ctx.JSON(http.StatusOK, c.model.UserModel(updatedUser))
}

func (c *Controller) UserSettingsChangeUsername(ctx echo.Context) error {
	req, err := c.model.UserSettingsChangeUsernameValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}
	if ok := c.authentication.VerifyPassword(user.Password, req.Password); !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if err := c.user.Manager.UpdateFields(user.ID, &model.User{
		UserName:  req.UserName,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}
	updatedUser, err := c.user.Manager.GetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	c.provider.SetUser(ctx, updatedUser)
	return ctx.JSON(http.StatusOK, c.model.UserModel(updatedUser))
}

func (c *Controller) UserSettingsChangeContactNumber(ctx echo.Context) error {
	req, err := c.model.UserSettingsChangeContactNumberValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}
	if ok := c.authentication.VerifyPassword(user.Password, req.Password); !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if err := c.user.Manager.UpdateFields(user.ID, &model.User{
		ContactNumber:     req.ContactNumber,
		IsContactVerified: false,
		UpdatedAt:         time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}
	updatedUser, err := c.user.Manager.GetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	c.provider.SetUser(ctx, updatedUser)
	return ctx.JSON(http.StatusOK, c.model.UserModel(updatedUser))
}

func (c *Controller) UserSettingsChangeProfilePicture(ctx echo.Context) error {
	req, err := c.model.UserSettingsChangeProfilePictureValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}
	if err := c.user.Manager.UpdateFields(user.ID, &model.User{
		MediaID:   req.MediaID,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}
	updatedUser, err := c.user.Manager.GetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	return ctx.JSON(http.StatusOK, c.model.UserModel(updatedUser))
}

func (c *Controller) UserSettingsChangeProfile(ctx echo.Context) error {
	req, err := c.model.UserSettingsChangeProfileValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}
	if err := c.user.Manager.UpdateFields(user.ID, &model.User{
		Birthdate:   req.Birthdate,
		Description: req.Description,
		FirstName:   req.FirstName,
		MiddleName:  req.MiddleName,
		LastName:    req.LastName,
		FullName:    req.FullName,
		Suffix:      req.Suffix,
		UpdatedAt:   time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}
	updatedUser, err := c.user.Manager.GetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	c.provider.SetUser(ctx, updatedUser)
	return ctx.JSON(http.StatusOK, c.model.UserModel(updatedUser))
}

func (c *Controller) UserSettingsChangeGeneral(ctx echo.Context) error {

	req, err := c.model.UserSettingsChangeGeneralValidate(ctx)
	if err != nil {
		return err
	}

	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}

	model := &model.User{}
	dirty := false

	if user.UserName != req.UserName {
		model.UserName = req.UserName
		dirty = true
	}
	if user.Description != req.Description {
		model.Description = req.Description
		dirty = true
	}
	if user.Email != req.Email {
		model.Email = req.Email
		model.IsEmailVerified = false
		dirty = true
	}
	if user.ContactNumber != req.ContactNumber {
		model.ContactNumber = req.ContactNumber
		model.IsContactVerified = false
		dirty = true
	}
	if !dirty {
		return ctx.JSON(http.StatusOK, c.model.UserModel(user))
	}
	model.UpdatedAt = time.Now().UTC()
	if err := c.user.Manager.UpdateFields(user.ID, model); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
	}
	updatedUser, err := c.user.Manager.GetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	// SetUser
	return ctx.JSON(http.StatusOK, c.model.UserModel(updatedUser))
}
