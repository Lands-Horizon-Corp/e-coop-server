package controller

import (
	"fmt"
	"net/http"

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
	return c.JSON(http.StatusOK, uc.collector.ToModel(user))
}

func (uc *UserController) UserLogin(c echo.Context) error {
	req, err := uc.collector.UserLoginValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)
	return c.JSON(http.StatusCreated, req)
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

	return c.JSON(http.StatusCreated, uc.collector.ToModel(user))
}

// UserForgotPassword handles forgot password

func (uc *UserController) UserForgotPassword(c echo.Context) error {
	req, err := uc.collector.UserForgotPasswordValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
}

// UserChangePassword handles change password

func (uc *UserController) UserChangePassword(c echo.Context) error {
	req, err := uc.collector.UserChangePasswordValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
}

// UserApplyContactNumber handles applying contact number

func (uc *UserController) UserApplyContactNumber(c echo.Context) error {
	req, err := uc.collector.UserApplyContactNumberValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
}

// UserVerifyContactNumber handles verification of contact number

func (uc *UserController) UserVerifyContactNumber(c echo.Context) error {
	req, err := uc.collector.UserVerifyContactNumberValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
}

func (uc *UserController) UserApplyEmail(c echo.Context) error {
	req, err := uc.collector.UserApplyEmailValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
}

// UserVerifyEmail handles verification of email

func (uc *UserController) UserVerifyEmail(c echo.Context) error {
	req, err := uc.collector.UserVerifyEmailValidation(c)
	if err != nil {
		return err
	}
	fmt.Println(req)

	return c.NoContent(http.StatusOK)
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
	group.POST("/authentication/register", uc.UserRegister) // Set token
	group.POST("/authentication/forgot-password", uc.UserForgotPassword)
	group.POST("/authentication/change-password", uc.UserChangePassword)

	group.POST("/authentication/apply-contact", uc.UserApplyContactNumber)
	group.POST("/authentication/verify-contact", uc.UserVerifyContactNumber)

	group.POST("/authentication/apply-email", uc.UserApplyEmail)
	group.POST("/authentication/verify-email", uc.UserVerifyEmail)

	// Legacy verify flows
	group.POST("/authentication/verify-with-email", uc.UserVerifyWithEmail)
	group.POST("/authentication/verify-with-email-confirmation", uc.UserVerifyWithEmailConfirmation)
	group.POST("/authentication/verify-with-contact", uc.UserVerifyWithContactNumber)
	group.POST("/authentication/verify-with-contact-confirmation", uc.UserVerifyWithContactNumberConfirmation)

	// Settings routes
	group.PUT("/settings/password", uc.UserSettingsChangePassword)
	group.PUT("/settings/email", uc.UserSettingsChangeEmail)
	group.PUT("/settings/username", uc.UserSettingsChangeUsername)
	group.PUT("/settings/contact", uc.UserSettingsChangeContactNumber)
	group.PUT("/settings/profile-picture", uc.UserSettingsChangeProfilePicture)
}
