package model

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/rotisserie/eris"
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
		Birthdate         time.Time      `gorm:"type:date" json:"birthdate"`
		UserName          string         `gorm:"type:varchar(100);not null;unique" json:"user_name"`
		FirstName         *string        `gorm:"type:varchar(100)" json:"first_name,omitempty"`
		MiddleName        *string        `gorm:"type:varchar(100)" json:"middle_name,omitempty"`
		LastName          *string        `gorm:"type:varchar(100)" json:"last_name,omitempty"`
		FullName          *string        `gorm:"type:varchar(255)" json:"full_name,omitempty"`
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
		OldPassword     string `json:"old_password" validate:"required,min=8"`
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
		Birthdate   time.Time `json:"birthdate"`
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
)

func (m *Model) User() {
	m.Migration = append(m.Migration, &User{})
	m.UserManager = horizon_services.NewRepository(horizon_services.RepositoryParams[User, UserResponse, UserRegisterRequest]{
		Preloads: []string{
			"Media",
			"SignatureMedia",
			"Footsteps",
			"Footsteps.Media",
			"GeneratedReports",
			"GeneratedReports.Media",
			"UserOrganizations",
			"UserOrganizations.Organization",
			"UserOrganizations.Branch",
			"UserOrganizations.Organization.Media",
			"UserOrganizations.Branch.Media",
		},
		Service: m.provider.Service,
		Resource: func(data *User) *UserResponse {
			ctx := context.Background()
			if data == nil {
				return nil
			}
			result, err := m.provider.Service.QR.EncodeQR(ctx, &QRUser{
				UserID:        data.ID.String(),
				Email:         data.Email,
				ContactNumber: data.ContactNumber,
				Username:      data.UserName,
				Lastname:      horizon.StringFormat(data.LastName),
				Firstname:     horizon.StringFormat(data.FirstName),
				Middlename:    horizon.StringFormat(data.MiddleName),
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
				"user.create." + data.ID.String(),
			}
		},
		Updated: func(data *User) []string {
			return []string{
				"user.update",
				"user.update." + data.ID.String(),
			}
		},
		Deleted: func(data *User) []string {
			return []string{
				"user.delete",
				"user.delete." + data.ID.String(),
			}
		},
	})
}

// user/contact-number/:contact_number_id
func (m *Model) GetUserByContactNumber(context context.Context, contactNumber string) (*User, error) {
	return m.UserManager.FindOne(context, &User{ContactNumber: contactNumber})
}

// user/email/:email
func (m *Model) GetUserByEmail(context context.Context, email string) (*User, error) {
	return m.UserManager.FindOne(context, &User{Email: email})
}

// user/user-name/:user-name
func (m *Model) GetUserByUserName(context context.Context, userName string) (*User, error) {
	return m.UserManager.FindOne(context, &User{UserName: userName})
}

// user/identifier/:identifier
func (m *Model) GetUserByIdentifier(context context.Context, identifier string) (*User, error) {
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
