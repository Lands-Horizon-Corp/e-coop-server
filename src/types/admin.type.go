package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// =========================
	// Admin DB Model
	// =========================
	Admin struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Password        string `gorm:"type:varchar(255);not null" json:"-"`
		Username        string `gorm:"type:varchar(100);not null;unique"`
		Email           string `gorm:"type:varchar(255);not null;unique"`
		IsEmailVerified bool   `gorm:"default:false"`

		FirstName   *string `gorm:"type:varchar(100)"`
		MiddleName  *string `gorm:"type:varchar(100)"`
		LastName    *string `gorm:"type:varchar(100)"`
		FullName    string  `gorm:"type:varchar(255)"`
		Suffix      *string `gorm:"type:varchar(50)"`
		Description *string `gorm:"type:text"`

		IsActive    bool `gorm:"default:true"`
		LastLoginAt *time.Time
	}

	// =========================
	// Admin API Response
	// =========================
	AdminResponse struct {
		ID              uuid.UUID `json:"id"`
		Username        string    `json:"user_name"`
		Email           string    `json:"email"`
		IsEmailVerified bool      `json:"is_email_verified"`

		FirstName   *string `json:"first_name,omitempty"`
		MiddleName  *string `json:"middle_name,omitempty"`
		LastName    *string `json:"last_name,omitempty"`
		FullName    string  `json:"full_name,omitempty"`
		Suffix      *string `json:"suffix,omitempty"`
		Description *string `json:"description,omitempty"`

		IsActive    bool       `json:"is_active"`
		LastLoginAt *time.Time `json:"last_login_at,omitempty"`

		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	// =========================
	// Admin Auth
	// =========================
	AdminLoginRequest struct {
		Key      string `json:"key" validate:"required"` // username or email
		Password string `json:"password" validate:"required,min=8"`
	}

	AdminRegisterRequest struct {
		Username string `json:"user_name" validate:"required,min=3,max=100"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`

		FirstName  *string `json:"first_name,omitempty"`
		MiddleName *string `json:"middle_name,omitempty"`
		LastName   *string `json:"last_name,omitempty"`
		FullName   string  `json:"full_name,omitempty"`
		Suffix     *string `json:"suffix,omitempty"`
	}

	AdminForgotPasswordRequest struct {
		Key string `json:"key" validate:"required"`
	}

	AdminChangePasswordRequest struct {
		NewPassword     string `json:"new_password" validate:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
	}

	// =========================
	// Admin Verification
	// =========================
	AdminVerifyEmailRequest struct {
		OTP string `json:"otp" validate:"required,min=6"`
	}

	// =========================
	// Admin Settings
	// =========================
	AdminSettingsChangePasswordRequest struct {
		OldPassword     string `json:"old_password" validate:"required,min=8"`
		NewPassword     string `json:"new_password" validate:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
	}

	AdminSettingsChangeEmailRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	AdminSettingsChangeUsernameRequest struct {
		Username string `json:"user_name" validate:"required,min=3,max=100"`
		Password string `json:"password" validate:"required,min=8"`
	}

	AdminSettingsChangeProfileRequest struct {
		FirstName   *string `json:"first_name,omitempty"`
		MiddleName  *string `json:"middle_name,omitempty"`
		LastName    *string `json:"last_name,omitempty"`
		FullName    string  `json:"full_name,omitempty"`
		Suffix      *string `json:"suffix,omitempty"`
		Description *string `json:"description,omitempty"`
	}

	AdminSettingsChangeGeneralRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Username string `json:"user_name" validate:"required,min=3,max=100"`
		IsActive bool   `json:"is_active"`
	}
)
