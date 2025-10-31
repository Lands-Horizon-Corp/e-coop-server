package model_core

import (
	"context"
	"fmt"
	"strings"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	// User represents a user account in the system with personal information and authentication details
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
		Birthdate         time.Time      `gorm:"type:date" json:"birthdate"`
		UserName          string         `gorm:"type:varchar(100);not null;unique" json:"user_name"`
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

	// UserResponse represents the JSON response structure for user data
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

	// CurrentUserResponse represents the response structure for the currently authenticated user
	CurrentUserResponse struct {
		UserID                  uuid.UUID                 `json:"user_id"`
		User                    *UserResponse             `json:"user"`
		UserOrganization        *UserOrganizationResponse `json:"user_organization"`
		IsLoggedInOnOtherDevice bool                      `json:"is_logged_in_on_other_device"`
		Users                   any                       `json:"users,omitempty"` // This can be used to return multiple users if needed
	}

	// UserLoginRequest represents the request payload for user authentication
	UserLoginRequest struct {
		Key      string `json:"key" validate:"required"`
		Password string `json:"password" validate:"required,min=8"`
	}
	// UserAdminPasswordVerificationRequest represents the request payload for admin password verification
	UserAdminPasswordVerificationRequest struct {
		UserOrganizationID uuid.UUID `json:"user_organization_id" validate:"required"`
		Password           string    `json:"password" validate:"required,min=8"`
	}

	// UserRegisterRequest represents the request payload for user registration
	UserRegisterRequest struct {
		Email         string     `json:"email" validate:"required,email"`
		Password      string     `json:"password" validate:"required,min=8"`
		Birthdate     time.Time  `json:"birthdate,omitempty"`
		UserName      string     `json:"user_name" validate:"required,min=3,max=100"`
		FullName      string     `json:"full_name,omitempty"`
		FirstName     *string    `json:"first_name,omitempty"`
		MiddleName    *string    `json:"middle_name,omitempty"`
		LastName      *string    `json:"last_name,omitempty"`
		Suffix        *string    `json:"suffix,omitempty"`
		ContactNumber string     `json:"contact_number" validate:"required,min=7,max=20"`
		MediaID       *uuid.UUID `json:"media_id,omitempty"`
	}

	// UserForgotPasswordRequest represents the request payload for password reset
	UserForgotPasswordRequest struct {
		Key string `json:"key" validate:"required"`
	}

	// UserChangePasswordRequest represents the request payload for changing user password
	UserChangePasswordRequest struct {
		NewPassword     string `json:"new_password" validate:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
	}

	// UserVerifyContactNumberRequest represents the request payload for contact number verification
	UserVerifyContactNumberRequest struct {
		OTP string `json:"otp" validate:"required,min=6"`
	}

	// UserVerifyEmailRequest represents the request payload for email verification
	UserVerifyEmailRequest struct {
		OTP string `json:"otp" validate:"required,min=6"`
	}

	// UserVerifyWithEmailConfirmationRequest represents the request payload for email confirmation verification
	UserVerifyWithEmailConfirmationRequest struct {
		OTP string `json:"otp" validate:"required,min=6"`
	}

	// UserVerifyWithContactNumberRequest represents the request payload for contact number verification
	UserVerifyWithContactNumberRequest struct {
		ContactNumber string `json:"contact_number" validate:"required,min=7,max=20"`
	}

	// UserVerifyWithContactNumberConfirmationRequest represents the request payload for contact number confirmation
	UserVerifyWithContactNumberConfirmationRequest struct {
		OTP string `json:"otp" validate:"required,min=6"`
	}
	// UserVerifyWithPasswordRequest represents the request payload for password verification
	UserVerifyWithPasswordRequest struct {
		Password string `json:"password" validate:"required,min=6"`
	}

	// UserSettingsChangePasswordRequest represents the request payload for changing password in user settings
	UserSettingsChangePasswordRequest struct {
		OldPassword     string `json:"old_password" validate:"required,min=8"`
		NewPassword     string `json:"new_password" validate:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
	}

	// UserSettingsChangeEmailRequest represents the request payload for changing email in user settings
	UserSettingsChangeEmailRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	// UserSettingsChangeUsernameRequest represents the request payload for changing username in user settings
	UserSettingsChangeUsernameRequest struct {
		UserName string `json:"user_name" validate:"required,min=3,max=100"`
		Password string `json:"password" validate:"required,min=8"`
	}

	// UserSettingsChangeContactNumberRequest represents the request payload for changing contact number in user settings
	UserSettingsChangeContactNumberRequest struct {
		ContactNumber string `json:"contact_number" validate:"required,min=7,max=20"`
		Password      string `json:"password" validate:"required,min=8"`
	}

	UserSettingsChangeProfilePictureRequest struct {
		MediaID *uuid.UUID `json:"media_id" validate:"required"`
	}

	UserSettingsChangeProfileRequest struct {
		Birthdate time.Time `json:"birthdate"`

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
		UserName      string  `json:"user_name" validate:"required,min=3,max=100"`
	}
)

func (m *ModelCore) User() {
	m.Migration = append(m.Migration, &User{})
	m.UserManager = horizon_services.NewRepository(horizon_services.RepositoryParams[User, UserResponse, UserRegisterRequest]{
		Preloads: []string{
			"Media",
			"SignatureMedia",
			"Footsteps",
			"Footsteps.Media",
			"GeneratedReports",
			"GeneratedReports.Media",
		},
		Service: m.provider.Service,
		Resource: func(data *User) *UserResponse {
			context := context.Background()
			if data == nil {
				return nil
			}
			result, err := m.provider.Service.QR.EncodeQR(context, &QRUser{
				UserID:        data.ID.String(),
				Email:         data.Email,
				ContactNumber: data.ContactNumber,
				Username:      data.UserName,
				Lastname:      handlers.StringFormat(data.LastName),
				Firstname:     handlers.StringFormat(data.FirstName),
				Middlename:    handlers.StringFormat(data.MiddleName),
			}, "user-qr")
			if err != nil {
				return nil
			}
			return &UserResponse{
				ID:                data.ID,
				Birthdate:         data.Birthdate.Format("2006-01-02"),
				UserName:          data.UserName,
				Description:       data.Description,
				FirstName:         data.FirstName,
				MiddleName:        data.MiddleName,
				LastName:          data.LastName,
				Suffix:            data.Suffix,
				Email:             data.Email,
				IsEmailVerified:   data.IsEmailVerified,
				ContactNumber:     data.ContactNumber,
				IsContactVerified: data.IsContactVerified,
				QRCode:            result,
				FullName:          data.FullName,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),

				MediaID:          data.MediaID,
				Media:            m.MediaManager.ToModel(data.Media),
				SignatureMediaID: data.SignatureMediaID,
				SignatureMedia:   m.MediaManager.ToModel(data.SignatureMedia),
				Footsteps:        m.FootstepManager.ToModels(data.Footsteps),
				GeneratedReports: m.GeneratedReportManager.ToModels(data.GeneratedReports),
				Notifications:    m.NotificationManager.ToModels(data.Notification),

				UserOrganizations: m.UserOrganizationManager.ToModels(data.UserOrganizations),
			}
		},

		Created: func(data *User) []string {
			return []string{
				"user.create",
				fmt.Sprintf("user.create.%s", data.ID),
			}
		},
		Updated: func(data *User) []string {
			return []string{
				"user.update",
				fmt.Sprintf("user.update.%s", data.ID),
			}
		},
		Deleted: func(data *User) []string {
			return []string{
				"user.delete",
				fmt.Sprintf("user.delete.%s", data.ID),
			}
		},
	})
}

// user/contact-number/:contact_number_id
func (m *ModelCore) GetUserByContactNumber(context context.Context, contactNumber string) (*User, error) {
	return m.UserManager.FindOne(context, &User{ContactNumber: contactNumber})
}

// user/email/:email
func (m *ModelCore) GetUserByEmail(context context.Context, email string) (*User, error) {
	return m.UserManager.FindOne(context, &User{Email: email})
}

// user/user-name/:user-name
func (m *ModelCore) GetUserByUserName(context context.Context, userName string) (*User, error) {
	return m.UserManager.FindOne(context, &User{UserName: userName})
}

// user/identifier/:identifier
func (m *ModelCore) GetUserByIdentifier(context context.Context, identifier string) (*User, error) {
	if strings.Contains(identifier, "@") {
		if u, err := m.GetUserByEmail(context, identifier); err == nil {
			return u, nil
		}
	}
	numeric := strings.Trim(identifier, "+-0123456789")
	if numeric == "" {
		if u, err := m.GetUserByContactNumber(context, identifier); err == nil {
			return u, nil
		}
	}
	if u, err := m.GetUserByUserName(context, identifier); err == nil {
		return u, nil
	}
	return nil, eris.New("user not found by email, contact number, or username")
}
