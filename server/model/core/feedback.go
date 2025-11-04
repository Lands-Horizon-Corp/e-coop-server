package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// Feedback represents the Feedback model.
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

	// FeedbackResponse represents the response structure for feedback data

	// FeedbackResponse represents the response structure for Feedback.
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

	// FeedbackRequest represents the request structure for creating/updating feedback

	// FeedbackRequest represents the request structure for Feedback.
	FeedbackRequest struct {
		ID           *uuid.UUID `json:"id,omitempty"`
		Email        string     `json:"email"        validate:"required,email"`
		Description  string     `json:"description"  validate:"required,min=5,max=2000"`
		FeedbackType string     `json:"feedback_type" validate:"required,oneof=general bug feature"`
		MediaID      *uuid.UUID `json:"media_id,omitempty"`
	}
)

func (m *Core) feedback() {
	m.Migration = append(m.Migration, &Feedback{})
	m.FeedbackManager = *registry.NewRegistry(registry.RegistryParams[Feedback, FeedbackResponse, FeedbackRequest]{
		Preloads: []string{"Media"},
		Service:  m.provider.Service,
		Resource: func(data *Feedback) *FeedbackResponse {
			if data == nil {
				return nil
			}
			return &FeedbackResponse{
				ID:           data.ID,
				Email:        data.Email,
				Description:  data.Description,
				FeedbackType: data.FeedbackType,
				MediaID:      data.MediaID,
				Media:        m.MediaManager.ToModel(data.Media),
				CreatedAt:    data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:    data.UpdatedAt.Format(time.RFC3339),
			}
		},
		Created: func(data *Feedback) []string {
			return []string{
				"feedback.create",
				fmt.Sprintf("feedback.create.%s", data.ID),
			}
		},
		Updated: func(data *Feedback) []string {
			return []string{
				"feedback.update",
				fmt.Sprintf("feedback.update.%s", data.ID),
			}
		},
		Deleted: func(data *Feedback) []string {
			return []string{
				"feedback.delete",
				fmt.Sprintf("feedback.delete.%s", data.ID),
			}
		},
	})
}
