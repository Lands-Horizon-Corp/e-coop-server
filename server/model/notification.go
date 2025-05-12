package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
)

type (
	Notification struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		UserID           uuid.UUID `gorm:"type:uuid;not null"`
		User             *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
		Title            string    `gorm:"type:varchar(255);not null"`
		Description      string    `gorm:"type:text;not null"`
		IsViewed         bool      `gorm:"default:false" json:"is_viewed"`
		NotificationType string    `gorm:"type:varchar(50);not null"`
	}

	NotificationResponse struct {
		ID               uuid.UUID     `json:"id"`
		UserID           uuid.UUID     `json:"user_id"`
		User             *UserResponse `json:"user,omitempty"`
		Title            string        `json:"title"`
		Description      string        `json:"description"`
		IsViewed         bool          `json:"is_viewed"`
		NotificationType string        `json:"notification_type"`
		CreatedAt        string        `json:"created_at"`
		UpdatedAt        string        `json:"updated_at"`
	}

	NotificationCollection struct {
		Manager CollectionManager[Notification]
	}
)

func (m *Model) NotificationModel(data *Notification) *NotificationResponse {
	return ToModel(data, func(data *Notification) *NotificationResponse {
		return &NotificationResponse{
			ID:               data.ID,
			UserID:           data.UserID,
			User:             m.UserModel(data.User),
			Title:            data.Title,
			Description:      data.Description,
			IsViewed:         data.IsViewed,
			NotificationType: data.NotificationType,
			CreatedAt:        data.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        data.UpdatedAt.Format(time.RFC3339),
		}
	})
}

func (m *Model) NotificationModels(data []*Notification) []*NotificationResponse {
	return ToModels(data, m.NotificationModel)
}

func NewNotificationCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*NotificationCollection, error) {
	manager := NewcollectionManager(
		database,
		broadcast,
		func(data *Notification) ([]string, any) {
			return []string{
				"notification.create",
				fmt.Sprintf("footstep.create.user.%s", data.UserID),
				fmt.Sprintf("notification.create.%s", data.ID),
			}, model.NotificationModel(data)
		},
		func(data *Notification) ([]string, any) {
			return []string{
				"notification.update",
				fmt.Sprintf("footstep.update.user.%s", data.UserID),
				fmt.Sprintf("notification.update.%s", data.ID),
			}, model.NotificationModel(data)
		},
		func(data *Notification) ([]string, any) {
			return []string{
				"notification.delete",
				fmt.Sprintf("footstep.delete.user.%s", data.UserID),
				fmt.Sprintf("notification.delete.%s", data.ID),
			}, model.NotificationModel(data)
		},
	)
	return &NotificationCollection{
		Manager: manager,
	}, nil
}
