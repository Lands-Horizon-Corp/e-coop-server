package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	horizon_manager "horizon.com/server/horizon/manager"
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
		SignatureMediaID  *uuid.UUID     `gorm:"type:uuid"`
		SignatureMedia    *Media         `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:SET NULL;" json:"signature,omitempty"`
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

		Footsteps         []*Footstep         `gorm:"foreignKey:UserID" json:"footsteps,omitempty"`          // footstep
		GeneratedReports  []*GeneratedReport  `gorm:"foreignKey:UserID" json:"generated_reports,omitempty"`  // generated report
		Notification      []*Notification     `gorm:"foreignKey:UserID" json:"notications,omitempty"`        // notification
		UserOrganizations []*UserOrganization `gorm:"foreignKey:UserID" json:"user_organizations,omitempty"` // user organization
	}
	UserResponse struct {
		ID                uuid.UUID         `json:"id"`
		MediaID           *uuid.UUID        `json:"media_id,omitempty"`
		Media             *MediaResponse    `json:"media,omitempty"`
		SignatureMediaID  *uuid.UUID        `json:"signature_media_id"`
		SignatureMedia    *MediaResponse    `json:"signature_media"`
		Birthdate         string            `json:"birthdate,omitempty"`
		UserName          string            `json:"user_name"`
		Description       *string           `json:"description"`
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

		Footsteps         []*FootstepResponse         `json:"footstep"`
		GeneratedReports  []*GeneratedReportResponse  `json:"generated_reports"`
		Notifications     []*NotificationResponse     `json:"notications"`
		UserOrganizations []*UserOrganizationResponse `json:"user_organizations,omitempty"`
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
	UserVerifyWithPasswordRequest struct {
		Password string `json:"password" validate:"required,min=6"`
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
		MediaID *uuid.UUID `json:"media_id" validate:"required"`
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
		ContactNumber string  `json:"contact_number" validate:"required,min=7,max=20"`
		Description   *string `json:"description,omitempty"`
		Email         string  `json:"email" validate:"required,email"`
		UserName      string  `json:"user_name" validate:"required,min=3,max=100"`
	}

	UserCollection struct {
		Manager horizon_manager.CollectionManager[User]
	}
)

func (m *Model) UserModel(data *User) *UserResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *User) *UserResponse {
		encoded, err := m.qr.Encode(&QRUser{
			UserID:        data.ID.String(),
			Email:         data.Email,
			ContactNumber: data.ContactNumber,
			Username:      data.UserName,
			Lastname:      horizon.StringFormat(data.LastName),
			Firstname:     horizon.StringFormat(data.FirstName),
			Middlename:    horizon.StringFormat(data.MiddleName),
		}, "user")
		if err != nil {
			return nil
		}
		return &UserResponse{
			ID:                data.ID,
			MediaID:           data.MediaID,
			Media:             m.MediaModel(data.Media),
			SignatureMediaID:  data.SignatureMediaID,
			SignatureMedia:    m.MediaModel(data.SignatureMedia),
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

			Footsteps:         m.FootstepModels(data.Footsteps),
			GeneratedReports:  m.GeneratedReportModels(data.GeneratedReports),
			Notifications:     m.NotificationModels(data.Notification),
			UserOrganizations: m.UserOrganizationModels(data.UserOrganizations),
		}
	})
}

func (m *Model) UserModels(data []*User) []*UserResponse {
	return horizon_manager.ToModels(data, m.UserModel)
}
func (m *Model) UserLoginValidate(ctx echo.Context) (*UserLoginRequest, error) {
	return horizon_manager.Validate[UserLoginRequest](ctx, m.validator)
}
func (m *Model) UserRegisterValidate(ctx echo.Context) (*UserRegisterRequest, error) {
	return horizon_manager.Validate[UserRegisterRequest](ctx, m.validator)
}
func (m *Model) UserForgotPasswordValidate(ctx echo.Context) (*UserForgotPasswordRequest, error) {
	return horizon_manager.Validate[UserForgotPasswordRequest](ctx, m.validator)
}
func (m *Model) UserChangePasswordValidate(ctx echo.Context) (*UserChangePasswordRequest, error) {
	return horizon_manager.Validate[UserChangePasswordRequest](ctx, m.validator)
}
func (m *Model) UserVerifyContactNumberValidate(ctx echo.Context) (*UserVerifyContactNumberRequest, error) {
	return horizon_manager.Validate[UserVerifyContactNumberRequest](ctx, m.validator)
}
func (m *Model) UserVerifyEmailValidate(ctx echo.Context) (*UserVerifyEmailRequest, error) {
	return horizon_manager.Validate[UserVerifyEmailRequest](ctx, m.validator)
}
func (m *Model) UserVerifyWithEmailConfirmationValidate(ctx echo.Context) (*UserVerifyWithEmailConfirmationRequest, error) {
	return horizon_manager.Validate[UserVerifyWithEmailConfirmationRequest](ctx, m.validator)
}
func (m *Model) UserVerifyWithContactNumberValidate(ctx echo.Context) (*UserVerifyWithContactNumberRequest, error) {
	return horizon_manager.Validate[UserVerifyWithContactNumberRequest](ctx, m.validator)
}
func (m *Model) UserVerifyWithContactNumberConfirmationValidate(ctx echo.Context) (*UserVerifyWithContactNumberConfirmationRequest, error) {
	return horizon_manager.Validate[UserVerifyWithContactNumberConfirmationRequest](ctx, m.validator)
}
func (m *Model) UserVerifyWithPasswordValidate(ctx echo.Context) (*UserVerifyWithPasswordRequest, error) {
	return horizon_manager.Validate[UserVerifyWithPasswordRequest](ctx, m.validator)
}

func (m *Model) UserSettingsChangePasswordValidate(ctx echo.Context) (*UserSettingsChangePasswordRequest, error) {
	return horizon_manager.Validate[UserSettingsChangePasswordRequest](ctx, m.validator)
}
func (m *Model) UserSettingsChangeEmailValidate(ctx echo.Context) (*UserSettingsChangeEmailRequest, error) {
	return horizon_manager.Validate[UserSettingsChangeEmailRequest](ctx, m.validator)
}
func (m *Model) UserSettingsChangeUsernameValidate(ctx echo.Context) (*UserSettingsChangeUsernameRequest, error) {
	return horizon_manager.Validate[UserSettingsChangeUsernameRequest](ctx, m.validator)
}
func (m *Model) UserSettingsChangeContactNumberValidate(ctx echo.Context) (*UserSettingsChangeContactNumberRequest, error) {
	return horizon_manager.Validate[UserSettingsChangeContactNumberRequest](ctx, m.validator)
}
func (m *Model) UserSettingsChangeProfilePictureValidate(ctx echo.Context) (*UserSettingsChangeProfilePictureRequest, error) {
	return horizon_manager.Validate[UserSettingsChangeProfilePictureRequest](ctx, m.validator)
}
func (m *Model) UserSettingsChangeProfileValidate(ctx echo.Context) (*UserSettingsChangeProfileRequest, error) {
	return horizon_manager.Validate[UserSettingsChangeProfileRequest](ctx, m.validator)
}
func (m *Model) UserSettingsChangeGeneralValidate(ctx echo.Context) (*UserSettingsChangeGeneralRequest, error) {
	return horizon_manager.Validate[UserSettingsChangeGeneralRequest](ctx, m.validator)
}

func NewUserCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*UserCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *User) ([]string, any) {
			return []string{
				"user.create",
				fmt.Sprintf("user.create.%s", data.ID),
			}, model.UserModel(data)
		},
		func(data *User) ([]string, any) {
			return []string{
				"user.update",
				fmt.Sprintf("user.update.%s", data.ID),
			}, model.UserModel(data)
		},
		func(data *User) ([]string, any) {
			return []string{
				"user.delete",
				fmt.Sprintf("user.delete.%s", data.ID),
			}, model.UserModel(data)
		},
		[]string{"SignatureMedia", "Media"},
	)
	return &UserCollection{
		Manager: manager,
	}, nil
}

// user/contact-number/:contact_number_id
func (fc *UserCollection) ByContactNumber(contactNumber string) (*User, error) {
	return fc.Manager.FindOne(&User{ContactNumber: contactNumber})
}

// user/email/:email
func (fc *UserCollection) ByEmail(email string) (*User, error) {
	return fc.Manager.FindOne(&User{Email: email})
}

// user/user-name/:user-name
func (fc *UserCollection) ByUserName(userName string) (*User, error) {
	return fc.Manager.FindOne(&User{UserName: userName})
}

// user/identifier/:identifier
func (fc *UserCollection) ByIdentifier(identifier string) (*User, error) {
	if strings.Contains(identifier, "@") {
		if u, err := fc.ByEmail(identifier); err == nil {
			return u, nil
		}
	}
	numeric := strings.Trim(identifier, "+-0123456789")
	if numeric == "" {
		if u, err := fc.ByContactNumber(identifier); err == nil {
			return u, nil
		}
	}
	if u, err := fc.ByUserName(identifier); err == nil {
		return u, nil
	}
	return nil, eris.New("user not found by email, contact number, or username")
}
