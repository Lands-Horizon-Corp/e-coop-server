package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/repository"
)

type UserController struct {
	repo           *repository.UserRepository
	collector      *collection.UserCollection
	storage        *horizon.HorizonStorage
	broadcast      *horizon.HorizonBroadcast
	authentication *horizon.HorizonAuthentication
}

func NewUserController(
	repo *repository.UserRepository,
	collector *collection.UserCollection,
	storage *horizon.HorizonStorage,
	broadcast *horizon.HorizonBroadcast,
	authentication *horizon.HorizonAuthentication,
) (*UserController, error) {
	return &UserController{
		repo:           repo,
		collector:      collector,
		storage:        storage,
		broadcast:      broadcast,
		authentication: authentication,
	}, nil
}

func (uc *UserController) UserCurrent(c echo.Context) error {
	claim, err := uc.authentication.GetUserFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}

	user, err := uc.repo.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return c.JSON(http.StatusOK, collection.CurrentUserResponse{
		UserID: user.ID,
		User:   uc.collector.ToModel(user),
	})
}

func (uc *UserController) UserLogin(c echo.Context) error {
	req, err := uc.collector.UserLoginValidation(c)
	if err != nil {
		return err
	}
	user, err := uc.repo.FindByIdentifier(req.Key)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if ok := uc.authentication.VerifyPassword(user.Password, req.Password); !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if err := uc.authentication.SetToken(c, horizon.Claim{
		ID:            user.ID.String(),
		Email:         user.Email,
		ContactNumber: user.ContactNumber,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
	}
	return c.JSON(http.StatusOK, collection.CurrentUserResponse{
		UserID: user.ID,
		User:   uc.collector.ToModel(user),
	})
}

func (uc *UserController) UserLogout(c echo.Context) error {
	uc.authentication.CleanToken(c)
	return c.NoContent(http.StatusOK)
}

// UserRegister handles user registration
func (uc *UserController) UserRegister(c echo.Context) error {
	req, err := uc.collector.UserRegisterValidation(c)
	if err != nil {
		return err
	}

	hashedPwd, err := uc.authentication.Password(req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
	}

	user := &collection.User{
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

	if err := uc.repo.Create(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could not register user: %v", err))
	}

	if err := uc.authentication.SetToken(c, horizon.Claim{
		ID:            user.ID.String(),
		Email:         user.Email,
		ContactNumber: user.ContactNumber,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
	}

	return c.JSON(http.StatusOK, collection.CurrentUserResponse{
		UserID: user.ID,
		User:   uc.collector.ToModel(user),
	})
}

func (uc *UserController) UserForgotPassword(c echo.Context) error {
	req, err := uc.collector.UserForgotPasswordValidation(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "please provide a valid email or contact number")
	}
	user, err := uc.repo.FindByIdentifier(req.Key)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "no account found with those details")
	}
	_, err = uc.authentication.GenerateSMTPLink("/auth/password-reset/", horizon.Claim{
		ID:            user.ID.String(),
		Email:         user.Email,
		ContactNumber: user.ContactNumber,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send reset link, please try again later")
	}
	_, err = uc.authentication.GenerateSMSLink("/auth/password-reset/", horizon.Claim{
		ID:            user.ID.String(),
		Email:         user.Email,
		ContactNumber: user.ContactNumber,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send reset link, please try again later")
	}
	return c.NoContent(http.StatusOK)
}

func (uc *UserController) UserVerifyResetLink(c echo.Context) error {
	idParam := c.Param("id")
	_, err := uc.authentication.ValidateLink(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "link is not valid")
	}
	return c.NoContent(http.StatusOK)

}

func (uc *UserController) UserChangePassword(c echo.Context) error {
	idParam := c.Param("id")

	req, err := uc.collector.UserChangePasswordValidation(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	claim, err := uc.authentication.ValidateLink(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "link is not valid or has expired")
	}

	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}

	hashedPwd, err := uc.authentication.Password(req.NewPassword)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash new password")
	}

	// Use Updates() to only update specific fields
	if err := uc.repo.UpdateFields(id, map[string]any{
		"password":   hashedPwd,
		"updated_at": time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}

	// Return updated user (optional)
	updatedUser, err := uc.repo.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}

	return c.JSON(http.StatusOK, uc.collector.ToModel(updatedUser))
}

func (uc *UserController) UserApplyContactNumber(c echo.Context) error {
	claim, err := uc.authentication.GetUserFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	if err := uc.authentication.SendSMSOTP(*claim); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send OTP: "+err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (uc *UserController) UserVerifyContactNumber(c echo.Context) error {
	req, err := uc.collector.UserVerifyContactNumberValidation(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}
	claim, err := uc.authentication.GetUserFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	valid := uc.authentication.VerifySMSOTP(*claim, req.OTP)
	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired OTP")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}
	if err := uc.repo.UpdateFields(id, map[string]any{
		"is_contact_verified": true,
		"updated_at":          time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update contact verification: "+err.Error())
	}
	updatedUser, err := uc.repo.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}
	return c.JSON(http.StatusOK, uc.collector.ToModel(updatedUser))
}

func (uc *UserController) UserApplyEmail(c echo.Context) error {
	claim, err := uc.authentication.GetUserFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	if err := uc.authentication.SendSMTPOTP(*claim); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send email verification: "+err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (uc *UserController) UserVerifyEmail(c echo.Context) error {
	req, err := uc.collector.UserVerifyEmailValidation(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}
	claim, err := uc.authentication.GetUserFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	valid := uc.authentication.VerifySMTPOTP(*claim, req.OTP)
	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired email token")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}
	if err := uc.repo.UpdateFields(id, map[string]any{
		"is_email_verified": true,
		"updated_at":        time.Now().UTC(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update email verification: "+err.Error())
	}
	updatedUser, err := uc.repo.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch updated user")
	}

	return c.JSON(http.StatusOK, uc.collector.ToModel(updatedUser))
}

// UserVerifyWithEmail handles legacy verify with email

func (uc *UserController) UserVerifyWithEmail(c echo.Context) error {
	req, err := uc.collector.UserVerifyWithEmailValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
}

func (uc *UserController) UserVerifyWithEmailConfirmation(c echo.Context) error {
	req, err := uc.collector.UserVerifyWithEmailConfirmationValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
}

// UserVerifyWithContactNumber handles legacy verify with contact number

func (uc *UserController) UserVerifyWithContactNumber(c echo.Context) error {
	req, err := uc.collector.UserVerifyWithContactNumberConfirmationValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
}

// UserVerifyWithContactNumberConfirmation handles legacy contact confirmation

func (uc *UserController) UserVerifyWithContactNumberConfirmation(c echo.Context) error {
	req, err := uc.collector.UserVerifyWithContactNumberConfirmationValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
}

// UserSettingsChangePassword handles user password update via settings

func (uc *UserController) UserSettingsChangePassword(c echo.Context) error {
	req, err := uc.collector.UserSettingsChangePasswordValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
}

// UserSettingsChangeEmail handles user email update via settings

func (uc *UserController) UserSettingsChangeEmail(c echo.Context) error {
	req, err := uc.collector.UserSettingsChangeEmailValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
}

func (uc *UserController) UserSettingsChangeUsername(c echo.Context) error {
	req, err := uc.collector.UserSettingsChangeUsernameValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)
	return c.NoContent(http.StatusOK)
}

func (uc *UserController) UserSettingsChangeContactNumber(c echo.Context) error {
	req, err := uc.collector.UserSettingsChangeContactNumberValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
}

func (uc *UserController) UserSettingsChangeProfilePicture(c echo.Context) error {
	req, err := uc.collector.UserSettingsChangeProfilePictureValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
}

func (uc *UserController) APIRoutes(e *echo.Echo) {
	group := e.Group("")
	group.GET("/authentication/current", uc.UserCurrent)

	group.POST("/authentication/login", uc.UserLogin)       // Set token
	group.POST("/authentication/logout", uc.UserLogout)     // Set token
	group.POST("/authentication/register", uc.UserRegister) // Set token
	group.POST("/authentication/forgot-password", uc.UserForgotPassword)
	group.GET("/authentication/verify-reset-link/:id", uc.UserVerifyResetLink)
	group.POST("/authentication/change-password/:id", uc.UserChangePassword)

	group.POST("/authentication/apply-contact-number", uc.UserApplyContactNumber)
	group.POST("/authentication/verify-contact-number", uc.UserVerifyContactNumber)

	group.POST("/authentication/apply-email", uc.UserApplyEmail)
	group.POST("/authentication/verify-email", uc.UserVerifyEmail)
	group.POST("/authentication/verify-with-email", uc.UserVerifyWithEmail)
	group.POST("/authentication/verify-with-email-confirmation", uc.UserVerifyWithEmailConfirmation)
	group.POST("/authentication/verify-with-contact", uc.UserVerifyWithContactNumber)
	group.POST("/authentication/verify-with-contact-confirmation", uc.UserVerifyWithContactNumberConfirmation)
	group.PUT("/settings/password", uc.UserSettingsChangePassword)
	group.PUT("/settings/email", uc.UserSettingsChangeEmail)
	group.PUT("/settings/username", uc.UserSettingsChangeUsername)
	group.PUT("/settings/contact", uc.UserSettingsChangeContactNumber)
	group.PUT("/settings/profile-picture", uc.UserSettingsChangeProfilePicture)
}
