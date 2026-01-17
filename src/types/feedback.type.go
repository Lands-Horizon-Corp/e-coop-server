package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Feedback struct {
		ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Email        string     `gorm:"type:varchar(255)"`
		Description  string     `gorm:"type:text"`
		FeedbackType string     `gorm:"type:varchar(50);not null;default:'general'"`
		MediaID      *uuid.UUID `gorm:"type:uuid"`
		Media        *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`
	}

	FeedbackResponse struct {
		ID           uuid.UUID      `json:"id"`
		Email        string         `json:"email"`
		Description  string         `json:"description"`
		FeedbackType string         `json:"feedback_type"`
		MediaID      *uuid.UUID     `json:"media_id"`
		Media        *MediaResponse `json:"media,omitempty"`
		CreatedAt    string         `json:"createdAt"`
		UpdatedAt    string         `json:"updatedAt"`
	}

	FeedbackRequest struct {
		ID           *uuid.UUID `json:"id,omitempty"`
		Email        string     `json:"email"        validate:"required,email"`
		Description  string     `json:"description"  validate:"required,min=5,max=2000"`
		FeedbackType string     `json:"feedback_type" validate:"required,oneof=general bug feature"`
		MediaID      *uuid.UUID `json:"media_id,omitempty"`
	}
)
