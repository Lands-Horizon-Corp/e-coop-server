package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
)

type QRUser struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	ContactNumber string `json:"contact_number"`
	Username      string `json:"user_name"`
	Name          string `json:"name"`
	Lastname      string `json:"lastname"`
	Firstname     string `json:"firstname"`
	Middlename    string `json:"middlename"`
}

type (
	User struct {
		ID                uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt         time.Time      `gorm:"not null;default:now()"`
		UpdatedAt         time.Time      `gorm:"not null;default:now()"`
		DeletedAt         gorm.DeletedAt `gorm:"index"`
		MediaID           *uuid.UUID     `gorm:"type:uuid"`
		Media             *Media         `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`
		Password          string         `gorm:"type:varchar(255);not null" json:"-"`
		Birthdate         time.Time      `gorm:"type:date" json:"birthdate,omitempty"`
		UserName          string         `gorm:"type:varchar(100);not null;unique" json:"user_name"`
		FirstName         *string        `gorm:"type:varchar(100)" json:"first_name,omitempty"`
		MiddleName        *string        `gorm:"type:varchar(100)" json:"middle_name,omitempty"`
		LastName          *string        `gorm:"type:varchar(100)" json:"last_name,omitempty"`
		FullName          *string        `gorm:"type:varchar(255)" json:"full_name,omitempty"`
		Suffix            *string        `gorm:"type:varchar(50)" json:"suffix,omitempty"`
		Description       *string        `gorm:"type:text"`
		Email             string         `gorm:"type:varchar(255);not null;unique" json:"email"`
		IsEmailVerified   bool           `gorm:"default:false" json:"is_email_verified"`
		ContactNumber     string         `gorm:"type:varchar(20);not null" json:"contact_number"`
		IsContactVerified bool           `gorm:"default:false" json:"is_contact_verified"`

		Footsteps        []*Footstep        `gorm:"foreignKey:UserID" json:"footsteps,omitempty"`
		GeneratedReports []*GeneratedReport `gorm:"foreignKey:UserID" json:"generated_reports,omitempty"`
		Notification     []*Notification    `gorm:"foreignKey:UserID" json:"notications,omitempty"`
	}
	UserResponse struct {
		ID                uuid.UUID         `json:"id"`
		MediaID           *uuid.UUID        `json:"media_id,omitempty"`
		Media             *MediaResponse    `json:"media,omitempty"`
		Birthdate         string            `json:"birthdate,omitempty"`
		UserName          string            `json:"user_name"`
		Description       *string           `gorm:"type:text"`
		FirstName         *string           `json:"first_name,omitempty"`
		MiddleName        *string           `json:"middle_name,omitempty"`
		LastName          *string           `json:"last_name,omitempty"`
		FullName          *string           `json:"full_name,omitempty"`
		Suffix            *string           `json:"suffix,omitempty"`
		Email             string            `json:"email"`
		IsEmailVerified   bool              `json:"is_email_verified"`
		ContactNumber     string            `json:"contact_number"`
		IsContactVerified bool              `json:"is_contact_verified"`
		CreatedAt         string            `json:"created_at"`
		UpdatedAt         string            `json:"updated_at"`
		QRCode            *horizon.QRResult `json:"qr_code,omitempty"`

		Footsteps        []*FootstepResponse        `json:"footstep"`
		GeneratedReports []*GeneratedReportResponse `json:"generated_reports"`
		Notifications    []*NotificationResponse    `json:"notications"`
	}
	CurrentUserResponse struct {
		UserID uuid.UUID     `json:"user_id"`
		User   *UserResponse `json:"user"`
	}

	UserLoginRequest struct {
		Key      string `json:"key" validate:"required"`
		Password string `json:"password" validate:"required,min=8"`
	}

	UserRegisterRequest struct {
		Email         string     `json:"email" validate:"required,email"`
		Password      string     `json:"password" validate:"required,min=8"`
		Birthdate     time.Time  `json:"birthdate,omitempty"`
		UserName      string     `json:"user_name" validate:"required,min=3,max=100"`
		FullName      *string    `json:"full_name,omitempty"`
		FirstName     *string    `json:"first_name,omitempty"`
		MiddleName    *string    `json:"middle_name,omitempty"`
		LastName      *string    `json:"last_name,omitempty"`
		Suffix        *string    `json:"suffix,omitempty"`
		ContactNumber string     `json:"contact_number" validate:"required,min=7,max=20"`
		MediaID       *uuid.UUID `json:"media_id,omitempty"`
	}

	UserForgotPasswordRequest struct {
		Key string `json:"key" validate:"required"`
	}

	UserChangePasswordRequest struct {
		NewPassword     string `json:"new_password" validate:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
	}

	UserVerifyContactNumberRequest struct {
		OTP string `json:"otp" validate:"required,min=6"`
	}

	UserVerifyEmailRequest struct {
		OTP string `json:"otp" validate:"required,min=6"`
	}

	UserVerifyWithEmailConfirmationRequest struct {
		OTP string `json:"otp" validate:"required,min=6"`
	}

	UserVerifyWithContactNumberRequest struct {
		ContactNumber string `json:"contact_number" validate:"required,min=7,max=20"`
	}

	UserVerifyWithContactNumberConfirmationRequest struct {
		OTP string `json:"otp" validate:"required,min=6"`
	}

	UserSettingsChangePasswordRequest struct {
		Password        string `json:"password" validate:"required,min=8"`
		NewPassword     string `json:"new_password" validate:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
	}

	UserSettingsChangeEmailRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	UserSettingsChangeUsernameRequest struct {
		UserName string `json:"user_name" validate:"required,min=3,max=100"`
		Password string `json:"password" validate:"required,min=8"`
	}

	UserSettingsChangeContactNumberRequest struct {
		ContactNumber string `json:"contact_number" validate:"required,min=7,max=20"`
		Password      string `json:"password" validate:"required,min=8"`
	}

	UserSettingsChangeProfilePictureRequest struct {
		MediaID *uuid.UUID `json:"media_id,omitempty"`
	}

	UserSettingsChangeProfileRequest struct {
		Birthdate   time.Time `json:"birthdate,omitempty"`
		Description *string   `json:"description,omitempty"`
		FirstName   *string   `json:"first_name,omitempty"`
		MiddleName  *string   `json:"middle_name,omitempty"`
		LastName    *string   `json:"last_name,omitempty"`
		FullName    *string   `json:"full_name,omitempty"`
		Suffix      *string   `json:"suffix,omitempty"`
	}

	UserSettingsChangeGeneralRequest struct {
		Email         string    `json:"email" validate:"required,email"`
		UserName      string    `json:"user_name" validate:"required,min=3,max=100"`
		ContactNumber string    `json:"contact_number" validate:"required,min=7,max=20"`
		Birthdate     time.Time `json:"birthdate,omitempty"`
		Description   *string   `json:"description,omitempty"`
		FirstName     *string   `json:"first_name,omitempty"`
		MiddleName    *string   `json:"middle_name,omitempty"`
		LastName      *string   `json:"last_name,omitempty"`
		FullName      *string   `json:"full_name,omitempty"`
		Suffix        *string   `json:"suffix,omitempty"`
		Password      string    `json:"password" validate:"required,min=8"`
	}
)

func (m *Model) UserModel(data *User) *UserResponse {
	return ToModel(data, func(data *User) *UserResponse {
		encoded, err := m.qr.Encode(&QRUser{
			UserID:        data.ID.String(),
			Email:         data.Email,
			ContactNumber: data.ContactNumber,
			Username:      data.UserName,
			Lastname:      horizon.StringFormat(data.LastName),
			Firstname:     horizon.StringFormat(data.FirstName),
			Middlename:    horizon.StringFormat(data.MiddleName),
		})
		if err != nil {
			return nil
		}
		return &UserResponse{
			ID:                data.ID,
			MediaID:           data.MediaID,
			Media:             m.MediaModel(data.Media),
			Birthdate:         data.Birthdate.Format(time.RFC3339),
			UserName:          data.UserName,
			FirstName:         data.FirstName,
			Description:       data.Description,
			MiddleName:        data.MiddleName,
			LastName:          data.LastName,
			FullName:          data.FullName,
			Suffix:            data.Suffix,
			Email:             data.Email,
			IsEmailVerified:   data.IsEmailVerified,
			ContactNumber:     data.ContactNumber,
			IsContactVerified: data.IsContactVerified,
			CreatedAt:         data.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
			QRCode:            encoded,

			Footsteps:        m.FootstepModels(data.Footsteps),
			GeneratedReports: m.GeneratedReportModels(data.GeneratedReports),
			Notifications:    m.NotificationModels(data.Notification),
		}
	})
}

func (m *Model) UserModels(data []*User) []*UserResponse {
	return ToModels(data, m.UserModel)
}
func (m *Model) UserLoginValidate(ctx echo.Context) (*UserLoginRequest, error) {
	return Validate[UserLoginRequest](ctx, m.validator)
}
func (m *Model) UserRegisterValidate(ctx echo.Context) (*UserRegisterRequest, error) {
	return Validate[UserRegisterRequest](ctx, m.validator)
}
func (m *Model) UserForgotPasswordValidate(ctx echo.Context) (*UserForgotPasswordRequest, error) {
	return Validate[UserForgotPasswordRequest](ctx, m.validator)
}
func (m *Model) UserChangePasswordValidate(ctx echo.Context) (*UserChangePasswordRequest, error) {
	return Validate[UserChangePasswordRequest](ctx, m.validator)
}
func (m *Model) UserVerifyContactNumberValidate(ctx echo.Context) (*UserVerifyContactNumberRequest, error) {
	return Validate[UserVerifyContactNumberRequest](ctx, m.validator)
}
func (m *Model) UserVerifyEmailValidate(ctx echo.Context) (*UserVerifyEmailRequest, error) {
	return Validate[UserVerifyEmailRequest](ctx, m.validator)
}
func (m *Model) UserVerifyWithEmailConfirmationValidate(ctx echo.Context) (*UserVerifyWithEmailConfirmationRequest, error) {
	return Validate[UserVerifyWithEmailConfirmationRequest](ctx, m.validator)
}
func (m *Model) UserVerifyWithContactNumberValidate(ctx echo.Context) (*UserVerifyWithContactNumberRequest, error) {
	return Validate[UserVerifyWithContactNumberRequest](ctx, m.validator)
}
func (m *Model) UserVerifyWithContactNumberConfirmationValidate(ctx echo.Context) (*UserVerifyWithContactNumberConfirmationRequest, error) {
	return Validate[UserVerifyWithContactNumberConfirmationRequest](ctx, m.validator)
}
func (m *Model) UserSettingsChangePasswordValidate(ctx echo.Context) (*UserSettingsChangePasswordRequest, error) {
	return Validate[UserSettingsChangePasswordRequest](ctx, m.validator)
}
func (m *Model) UserSettingsChangeEmailValidate(ctx echo.Context) (*UserSettingsChangeEmailRequest, error) {
	return Validate[UserSettingsChangeEmailRequest](ctx, m.validator)
}
func (m *Model) UserSettingsChangeUsernameValidate(ctx echo.Context) (*UserSettingsChangeUsernameRequest, error) {
	return Validate[UserSettingsChangeUsernameRequest](ctx, m.validator)
}
func (m *Model) UserSettingsChangeContactNumberValidate(ctx echo.Context) (*UserSettingsChangeContactNumberRequest, error) {
	return Validate[UserSettingsChangeContactNumberRequest](ctx, m.validator)
}
func (m *Model) UserSettingsChangeProfilePictureValidate(ctx echo.Context) (*UserSettingsChangeProfilePictureRequest, error) {
	return Validate[UserSettingsChangeProfilePictureRequest](ctx, m.validator)
}
func (m *Model) UserSettingsChangeProfileValidate(ctx echo.Context) (*UserSettingsChangeProfileRequest, error) {
	return Validate[UserSettingsChangeProfileRequest](ctx, m.validator)
}
func (m *Model) UserSettingsChangeGeneralValidate(ctx echo.Context) (*UserSettingsChangeGeneralRequest, error) {
	return Validate[UserSettingsChangeGeneralRequest](ctx, m.validator)
}

// func NewUserCollection(
// 	media *MediaCollection,
// 	qr *horizon.HorizonQR,
// 	footstep *FootstepCollection,
// 	generatedReport *GeneratedReportCollection,
// 	notification *NotificationCollection,
// ) (*UserCollection, error) {
// 	return &UserCollection{
// 		media:           media,
// 		qr:              qr,
// 		validator:       validator.New(),
// 		footstep:        footstep,
// 		generatedReport: generatedReport,
// 		notification:    notification,
// 	}, nil
// }

// func (uc *UserCollection) ToModel(data *User) *UserResponse {
// 	if data == nil {
// 		return nil
// 	}
// 	encoded, err := uc.qr.Encode(&QRUser{
// 		UserID:        data.ID.String(),
// 		Email:         data.Email,
// 		ContactNumber: data.ContactNumber,
// 		Username:      data.UserName,
// 		Lastname:      horizon.StringFormat(data.LastName),
// 		Firstname:     horizon.StringFormat(data.FirstName),
// 		Middlename:    horizon.StringFormat(data.MiddleName),
// 	})
// 	if err != nil {
// 		return nil
// 	}
// 	return &UserResponse{
// 		ID:                data.ID,
// 		MediaID:           data.MediaID,
// 		Media:             uc.media.ToModel(data.Media),
// 		Birthdate:         data.Birthdate.Format(time.RFC3339),
// 		UserName:          data.UserName,
// 		FirstName:         data.FirstName,
// 		Description:       data.Description,
// 		MiddleName:        data.MiddleName,
// 		LastName:          data.LastName,
// 		FullName:          data.FullName,
// 		Suffix:            data.Suffix,
// 		Email:             data.Email,
// 		IsEmailVerified:   data.IsEmailVerified,
// 		ContactNumber:     data.ContactNumber,
// 		IsContactVerified: data.IsContactVerified,
// 		CreatedAt:         data.CreatedAt.Format(time.RFC3339),
// 		UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
// 		QRCode:            encoded,

// 		Footsteps:        uc.footstep.ToModels(data.Footsteps),
// 		GeneratedReports: uc.generatedReport.ToModels(data.GeneratedReports),
// 		Notifications:    uc.notification.ToModels(data.Notification),
// 	}
// }

// func (uc *UserCollection) ToModels(data []*User) []*UserResponse {
// 	if data == nil {
// 		return make([]*UserResponse, 0)
// 	}
// 	var response []*UserResponse
// 	for _, value := range data {
// 		model := uc.ToModel(value)
// 		if model != nil {
// 			response = append(response, model)
// 		}
// 	}
// 	if len(response) <= 0 {
// 		return make([]*UserResponse, 0)
// 	}
// 	return response
// }

// func (uc *UserCollection) UserLoginValidation(c echo.Context) (*UserLoginRequest, error) {
// 	u := new(UserLoginRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }
// func (uc *UserCollection) UserRegisterValidation(c echo.Context) (*UserRegisterRequest, error) {
// 	u := new(UserRegisterRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }
// func (uc *UserCollection) UserForgotPasswordValidation(c echo.Context) (*UserForgotPasswordRequest, error) {
// 	u := new(UserForgotPasswordRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }
// func (uc *UserCollection) UserChangePasswordValidation(c echo.Context) (*UserChangePasswordRequest, error) {
// 	u := new(UserChangePasswordRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }

// func (uc *UserCollection) UserVerifyContactNumberValidation(c echo.Context) (*UserVerifyContactNumberRequest, error) {
// 	u := new(UserVerifyContactNumberRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }

// func (uc *UserCollection) UserVerifyEmailValidation(c echo.Context) (*UserVerifyEmailRequest, error) {
// 	u := new(UserVerifyEmailRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }

// func (uc *UserCollection) UserVerifyWithEmailConfirmationValidation(c echo.Context) (*UserVerifyWithEmailConfirmationRequest, error) {
// 	u := new(UserVerifyWithEmailConfirmationRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }
// func (uc *UserCollection) UserVerifyWithContactNumberValidation(c echo.Context) (*UserVerifyWithContactNumberRequest, error) {
// 	u := new(UserVerifyWithContactNumberRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }
// func (uc *UserCollection) UserVerifyWithContactNumberConfirmationValidation(c echo.Context) (*UserVerifyWithContactNumberConfirmationRequest, error) {
// 	u := new(UserVerifyWithContactNumberConfirmationRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }
// func (uc *UserCollection) UserSettingsChangePasswordValidation(c echo.Context) (*UserSettingsChangePasswordRequest, error) {
// 	u := new(UserSettingsChangePasswordRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }
// func (uc *UserCollection) UserSettingsChangeEmailValidation(c echo.Context) (*UserSettingsChangeEmailRequest, error) {
// 	u := new(UserSettingsChangeEmailRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }
// func (uc *UserCollection) UserSettingsChangeUsernameValidation(c echo.Context) (*UserSettingsChangeUsernameRequest, error) {
// 	u := new(UserSettingsChangeUsernameRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }
// func (uc *UserCollection) UserSettingsChangeContactNumberValidation(c echo.Context) (*UserSettingsChangeContactNumberRequest, error) {
// 	u := new(UserSettingsChangeContactNumberRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }
// func (uc *UserCollection) UserSettingsChangeProfilePictureValidation(c echo.Context) (*UserSettingsChangeProfilePictureRequest, error) {
// 	u := new(UserSettingsChangeProfilePictureRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }

// func (uc *UserCollection) UserSettingsChangeProfileValidation(c echo.Context) (*UserSettingsChangeProfileRequest, error) {
// 	u := new(UserSettingsChangeProfileRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }

// func (uc *UserCollection) UserSettingsChangeGeneralValidation(c echo.Context) (*UserSettingsChangeGeneralRequest, error) {
// 	u := new(UserSettingsChangeGeneralRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := uc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }
