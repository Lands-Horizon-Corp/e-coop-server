package types

import (
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	User struct {
		ID                uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt         time.Time      `gorm:"not null;default:now()"`
		UpdatedAt         time.Time      `gorm:"not null;default:now()"`
		DeletedAt         gorm.DeletedAt `gorm:"index"`
		MediaID           *uuid.UUID     `gorm:"type:uuid"`
		Media             *Media         `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`
		SignatureMediaID  *uuid.UUID     `gorm:"type:uuid"`
		SignatureMedia    *Media         `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:SET NULL;" json:"signature,omitempty"`
		Password          string         `gorm:"type:varchar(255);not null" json:"-"`
		Birthdate         *time.Time     `gorm:"type:date" json:"birthdate"`
		Username          string         `gorm:"type:varchar(100);not null;unique" json:"user_name"`
		FirstName         *string        `gorm:"type:varchar(100)" json:"first_name,omitempty"`
		MiddleName        *string        `gorm:"type:varchar(100)" json:"middle_name,omitempty"`
		LastName          *string        `gorm:"type:varchar(100)" json:"last_name,omitempty"`
		FullName          string         `gorm:"type:varchar(255)" json:"full_name,omitempty"`
		Suffix            *string        `gorm:"type:varchar(50)" json:"suffix,omitempty"`
		Description       *string        `gorm:"type:text"`
		Email             string         `gorm:"type:varchar(255);not null;unique" json:"email"`
		IsEmailVerified   bool           `gorm:"default:false" json:"is_email_verified"`
		ContactNumber     string         `gorm:"type:varchar(20);not null;unique" json:"contact_number"`
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
		Username          string            `json:"user_name"`
		Description       *string           `json:"description"`
		FirstName         *string           `json:"first_name,omitempty"`
		MiddleName        *string           `json:"middle_name,omitempty"`
		LastName          *string           `json:"last_name,omitempty"`
		FullName          string            `json:"full_name,omitempty"`
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
		UserID                  uuid.UUID                 `json:"user_id"`
		User                    *UserResponse             `json:"user"`
		UserOrganization        *UserOrganizationResponse `json:"user_organization"`
		IsLoggedInOnOtherDevice bool                      `json:"is_logged_in_on_other_device"`
		Users                   any                       `json:"users,omitempty"` // This can be used to return multiple users if needed
	}

	UserLoginRequest struct {
		Key      string `json:"key" validate:"required"`
		Password string `json:"password" validate:"required,min=8"`
	}
	UserAdminPasswordVerificationRequest struct {
		UserOrganizationID uuid.UUID `json:"user_organization_id" validate:"required"`
		Password           string    `json:"password" validate:"required,min=8"`
	}

	UserRegisterRequest struct {
		Email         string     `json:"email" validate:"required,email"`
		Password      string     `json:"password" validate:"required,min=8"`
		Birthdate     *time.Time `json:"birthdate,omitempty"`
		Username      string     `json:"user_name" validate:"required,min=3,max=100"`
		FullName      string     `json:"full_name,omitempty"`
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
		OldPassword     string `json:"old_password" validate:"required,min=8"`
		NewPassword     string `json:"new_password" validate:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
	}

	UserSettingsChangeEmailRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	UserSettingsChangeUsernameRequest struct {
		Username string `json:"user_name" validate:"required,min=3,max=100"`
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
		Birthdate *time.Time `json:"birthdate"`

		FirstName  *string `json:"first_name,omitempty"`
		MiddleName *string `json:"middle_name,omitempty"`
		LastName   *string `json:"last_name,omitempty"`
		FullName   string  `json:"full_name,omitempty"`
		Suffix     *string `json:"suffix,omitempty"`
	}

	UserSettingsChangeGeneralRequest struct {
		ContactNumber string  `json:"contact_number" validate:"required,min=7,max=20"`
		Description   *string `json:"description,omitempty"`
		Email         string  `json:"email" validate:"required,email"`
		Username      string  `json:"user_name" validate:"required,min=3,max=100"`
	}
)
