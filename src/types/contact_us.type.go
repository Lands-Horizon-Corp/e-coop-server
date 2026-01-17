package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	ContactUs struct {
		ID            uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		FirstName     string         `gorm:"type:varchar(255);not null"`
		LastName      string         `gorm:"type:varchar(255)"`
		Email         string         `gorm:"type:varchar(255)"`
		ContactNumber string         `gorm:"type:varchar(20)"`
		Description   string         `gorm:"type:text;not null"`
		CreatedAt     time.Time      `gorm:"not null;default:now()"`
		UpdatedAt     time.Time      `gorm:"not null;default:now()"`
		DeletedAt     gorm.DeletedAt `gorm:"index"`
	}

	ContactUsResponse struct {
		ID            uuid.UUID `json:"id"`
		FirstName     string    `json:"first_name"`
		LastName      string    `json:"last_name,omitempty"`
		Email         string    `json:"email,omitempty"`
		ContactNumber string    `json:"contact_number,omitempty"`
		Description   string    `json:"description"`
		CreatedAt     string    `json:"created_at"`
		UpdatedAt     string    `json:"updated_at"`
	}

	ContactUsRequest struct {
		ID            *uuid.UUID `json:"id,omitempty"`
		FirstName     string     `json:"first_name" validate:"required,min=1,max=255"`
		LastName      string     `json:"last_name,omitempty" validate:"omitempty,min=1,max=255"`
		Email         string     `json:"email,omitempty" validate:"omitempty,email,max=255"`
		ContactNumber string     `json:"contact_number,omitempty" validate:"omitempty,min=1,max=20"`
		Description   string     `json:"description" validate:"required,min=1"`
	}
)
