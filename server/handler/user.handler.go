package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

func (h *Handler) UserCurrent(c echo.Context) error {
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.CurrentUserResponse{
		UserID: user.ID,
		User:   h.model.UserModel(user),
	})
}

func (h *Handler) UserLogin(c echo.Context) error {
	req, err := h.model.UserLoginValidate(c)

	if err != nil {
		return err
	}
	user, err := h.repository.UserFindByIdentifier(req.Key)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if ok := h.authentication.VerifyPassword(user.Password, req.Password); !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if err := h.authentication.SetToken(c, horizon.Claim{
		ID:            user.ID.String(),
		Email:         user.Email,
		ContactNumber: user.ContactNumber,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
	}
	return c.JSON(http.StatusOK, model.CurrentUserResponse{
		UserID: user.ID,
		User:   h.model.UserModel(user),
	})
}

func (h *Handler) UserLogout(c echo.Context) error {
	h.authentication.CleanToken(c)
	return c.NoContent(http.StatusOK)
}

func (h *Handler) UserRegister(c echo.Context) error {
	req, err := h.model.UserRegisterValidate(c)
	if err != nil {
		return err
	}
	hashedPwd, err := h.authentication.Password(req.Password)
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
	if err := h.repository.UserCreate(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could not register user: %v", err))
	}
	if err := h.authentication.SetToken(c, horizon.Claim{
		ID:            user.ID.String(),
		Email:         user.Email,
		ContactNumber: user.ContactNumber,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
	}

	return c.JSON(http.StatusOK, model.CurrentUserResponse{
		UserID: user.ID,
		User:   h.model.UserModel(user),
	})
}

func (h *Handler) UserForgotPassword(c echo.Context) error {
	req, err := h.model.UserForgotPasswordValidate(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "please provide a valid email or contact number")
	}
	user, err := h.repository.UserFindByIdentifier(req.Key)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "no account found with those details")
	}
	_, err = h.authentication.GenerateSMTPLink("/auth/password-reset/", horizon.Claim{
		ID:            user.ID.String(),
		Email:         user.Email,
		ContactNumber: user.ContactNumber,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send reset link, please try again later")
	}
	_, err = h.authentication.GenerateSMSLink("/auth/password-reset/", horizon.Claim{
		ID:            user.ID.String(),
		Email:         user.Email,
		ContactNumber: user.ContactNumber,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send reset link, please try again later")
	}
	return c.NoContent(http.StatusOK)
}

func (uc *Handler) UserVerifyResetLink(c echo.Context) error {
	idParam := c.Param("id")
	_, err := uc.authentication.ValidateLink(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "link is not valid")
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) UserChangePassword(c echo.Context) error {
	idParam := c.Param("id")
	req, err := h.model.UserChangePasswordValidate(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}
	claim, err := h.authentication.ValidateLink(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "link is not valid or has expired")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}
	hashedPwd, err := h.authentication.Password(req.NewPassword)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash new password")
	}
	if err := h.repository.UserUpdateFields(id, &model.User{
		Password:  hashedPwd,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}
	updatedUser, err := h.repository.UserGetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}

	return c.JSON(http.StatusOK, h.model.UserModel(updatedUser))
}

func (h *Handler) UserApplyContactNumber(c echo.Context) error {
	claim, err := h.authentication.GetUserFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	if err := h.authentication.SendSMSOTP(*claim); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send OTP: "+err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) UserVerifyContactNumber(c echo.Context) error {
	req, err := h.model.UserVerifyContactNumberValidate(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}
	claim, err := h.authentication.GetUserFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	valid := h.authentication.VerifySMSOTP(*claim, req.OTP)
	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired OTP")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}

	if err := h.repository.UserUpdateFields(id, &model.User{
		IsContactVerified: true,
		UpdatedAt:         time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}

	updatedUser, err := h.repository.UserGetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	return c.JSON(http.StatusOK, h.model.UserModel(updatedUser))
}

func (h *Handler) UserApplyEmail(c echo.Context) error {
	claim, err := h.authentication.GetUserFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	if err := h.authentication.SendSMTPOTP(*claim); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send email verification: "+err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) UserVerifyEmail(c echo.Context) error {
	req, err := h.model.UserVerifyEmailValidate(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}
	claim, err := h.authentication.GetUserFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	valid := h.authentication.VerifySMTPOTP(*claim, req.OTP)
	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired email token")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}

	if err := h.repository.UserUpdateFields(id, &model.User{
		IsEmailVerified: true,
		UpdatedAt:       time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}

	updatedUser, err := h.repository.UserGetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}

	return c.JSON(http.StatusOK, h.model.UserModel(updatedUser))
}

func (h *Handler) UserVerifyWithEmail(c echo.Context) error {
	claim, err := h.authentication.GetUserFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	if err := h.authentication.SendSMTPOTP(*claim); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send email verification: "+err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) UserVerifyWithEmailConfirmation(c echo.Context) error {
	req, err := h.model.UserVerifyWithEmailConfirmationValidate(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}
	claim, err := h.authentication.GetUserFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	valid := h.authentication.VerifySMTPOTP(*claim, req.OTP)
	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired email token")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}

	updatedUser, err := h.repository.UserGetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}

	return c.JSON(http.StatusOK, h.model.UserModel(updatedUser))
}

func (h *Handler) UserVerifyWithContactNumber(c echo.Context) error {
	claim, err := h.authentication.GetUserFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	if err := h.authentication.SendSMSOTP(*claim); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send OTP: "+err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) UserVerifyWithContactNumberConfirmation(c echo.Context) error {
	req, err := h.model.UserVerifyWithContactNumberConfirmationValidate(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}
	claim, err := h.authentication.GetUserFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	valid := h.authentication.VerifySMSOTP(*claim, req.OTP)
	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired OTP")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}
	updatedUser, err := h.repository.UserGetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	return c.JSON(http.StatusOK, h.model.UserModel(updatedUser))
}

func (h *Handler) UserSettingsChangePassword(c echo.Context) error {
	req, err := h.model.UserSettingsChangePasswordValidate(c)
	if err != nil {
		return err
	}
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	if ok := h.authentication.VerifyPassword(user.Password, req.Password); !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	hashedPwd, err := h.authentication.Password(req.NewPassword)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash new password")
	}

	if err := h.repository.UserUpdateFields(user.ID, &model.User{
		Password:  hashedPwd,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}
	updatedUser, err := h.repository.UserGetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	return c.JSON(http.StatusOK, h.model.UserModel(updatedUser))
}

func (h *Handler) UserSettingsChangeEmail(c echo.Context) error {
	req, err := h.model.UserSettingsChangeEmailValidate(c)
	if err != nil {
		return err
	}
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	if ok := h.authentication.VerifyPassword(user.Password, req.Password); !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if err := h.repository.UserUpdateFields(user.ID, &model.User{
		Email:           req.Email,
		IsEmailVerified: false,
		UpdatedAt:       time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}
	updatedUser, err := h.repository.UserGetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}

	// SetUser
	return c.JSON(http.StatusOK, h.model.UserModel(updatedUser))
}

func (h *Handler) UserSettingsChangeUsername(c echo.Context) error {
	req, err := h.model.UserSettingsChangeUsernameValidate(c)
	if err != nil {
		return err
	}
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	if ok := h.authentication.VerifyPassword(user.Password, req.Password); !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if err := h.repository.UserUpdateFields(user.ID, &model.User{
		UserName:  req.UserName,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}
	updatedUser, err := h.repository.UserGetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	return c.JSON(http.StatusOK, h.model.UserModel(updatedUser))
}

func (h *Handler) UserSettingsChangeContactNumber(c echo.Context) error {

	req, err := h.model.UserSettingsChangeContactNumberValidate(c)
	if err != nil {
		return err
	}
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	if ok := h.authentication.VerifyPassword(user.Password, req.Password); !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if err := h.repository.UserUpdateFields(user.ID, &model.User{
		ContactNumber:     req.ContactNumber,
		IsContactVerified: false,
		UpdatedAt:         time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}
	updatedUser, err := h.repository.UserGetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	// SetUser
	return c.JSON(http.StatusOK, h.model.UserModel(updatedUser))
}

func (h *Handler) UserSettingsChangeProfilePicture(c echo.Context) error {
	req, err := h.model.UserSettingsChangeProfilePictureValidate(c)
	if err != nil {
		return err
	}
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	if err := h.repository.UserUpdateFields(user.ID, &model.User{
		MediaID:   req.MediaID,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}
	updatedUser, err := h.repository.UserGetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	return c.JSON(http.StatusOK, h.model.UserModel(updatedUser))
}

func (h *Handler) UserSettingsChangeProfile(c echo.Context) error {
	req, err := h.model.UserSettingsChangeProfileValidate(c)
	if err != nil {
		return err
	}
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	if err := h.repository.UserUpdateFields(user.ID, &model.User{
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
	updatedUser, err := h.repository.UserGetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	return c.JSON(http.StatusOK, h.model.UserModel(updatedUser))
}

func (h *Handler) UserSettingsChangeGeneral(c echo.Context) error {

	req, err := h.model.UserSettingsChangeGeneralValidate(c)
	if err != nil {
		return err
	}

	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}

	if ok := h.authentication.VerifyPassword(user.Password, req.Password); !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	model := &model.User{}
	dirty := false

	if user.FirstName != req.FirstName {
		model.FirstName = req.FirstName
		dirty = true
	}
	if user.MiddleName != req.MiddleName {
		model.MiddleName = req.MiddleName
		dirty = true
	}
	if user.LastName != req.LastName {
		model.LastName = req.LastName
		dirty = true
	}
	if user.FullName != req.FullName {
		model.FullName = req.FullName
		dirty = true
	}
	if user.Suffix != req.Suffix {
		model.Suffix = req.Suffix
		dirty = true
	}
	if user.UserName != req.UserName {
		model.UserName = req.UserName
		dirty = true
	}
	if user.Description != req.Description {
		model.Description = req.Description
		dirty = true
	}
	if !user.Birthdate.Equal(user.Birthdate) {
		model.Birthdate = req.Birthdate
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
		return c.JSON(http.StatusOK, h.model.UserModel(user))
	}
	model.UpdatedAt = time.Now().UTC()
	if err := h.repository.UserUpdateFields(user.ID, model); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
	}
	updatedUser, err := h.repository.UserGetByID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	// SetUser
	return c.JSON(http.StatusOK, h.model.UserModel(updatedUser))
}
