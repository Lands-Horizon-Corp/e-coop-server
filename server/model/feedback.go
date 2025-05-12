package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	horizon_manager "horizon.com/server/horizon/manager"
)

type (
	Feedback struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
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
		ID           *string    `json:"id,omitempty"`
		Email        string     `json:"email"        validate:"required,email"`
		Description  string     `json:"description"  validate:"required,min=5,max=2000"`
		FeedbackType string     `json:"feedback_type" validate:"required,oneof=general bug feature"`
		MediaID      *uuid.UUID `json:"media_id,omitempty"`
	}
	FeedbackCollection struct {
		Manager horizon_manager.CollectionManager[Feedback]
	}
)

func (m *Model) FeedbackValidate(ctx echo.Context) (*FeedbackRequest, error) {
	return horizon_manager.Validate[FeedbackRequest](ctx, m.validator)
}

func (m *Model) FeedbackModel(data *Feedback) *FeedbackResponse {
	return horizon_manager.ToModel(data, func(data *Feedback) *FeedbackResponse {
		return &FeedbackResponse{
			ID:           data.ID,
			CreatedAt:    data.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    data.UpdatedAt.Format(time.RFC3339),
			MediaID:      data.MediaID,
			Media:        m.MediaModel(data.Media),
			Email:        data.Email,
			Description:  data.Description,
			FeedbackType: data.FeedbackType,
		}
	})
}

func (m *Model) FeedbackModels(data []*Feedback) []*FeedbackResponse {
	return horizon_manager.ToModels(data, m.FeedbackModel)
}

func NewFeedbackCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*FeedbackCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *Feedback) ([]string, any) {
			return []string{
				"feedback.create",
				fmt.Sprintf("feedback.create.%s", data.ID),
			}, model.FeedbackModel(data)
		},
		func(data *Feedback) ([]string, any) {
			return []string{
				"feedback.update",
				fmt.Sprintf("feedback.update.%s", data.ID),
			}, model.FeedbackModel(data)
		},
		func(data *Feedback) ([]string, any) {
			return []string{
				"feedback.delete",
				fmt.Sprintf("feedback.delete.%s", data.ID),
			}, model.FeedbackModel(data)
		},
		[]string{"Media"},
	)
	return &FeedbackCollection{
		Manager: manager,
	}, nil
}
