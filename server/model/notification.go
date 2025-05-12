package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	horizon_manager "horizon.com/server/horizon/manager"
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
		Manager horizon_manager.CollectionManager[Notification]
	}
)

func (m *Model) NotificationModel(data *Notification) *NotificationResponse {
	return horizon_manager.ToModel(data, func(data *Notification) *NotificationResponse {
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
	return horizon_manager.ToModels(data, m.NotificationModel)
}

func NewNotificationCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*NotificationCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *Notification) ([]string, any) {
			return []string{
				"notification.create",
				fmt.Sprintf("notification.create.user.%s", data.UserID),
				fmt.Sprintf("notification.create.%s", data.ID),
			}, model.NotificationModel(data)
		},
		func(data *Notification) ([]string, any) {
			return []string{
				"notification.update",
				fmt.Sprintf("notification.update.user.%s", data.UserID),
				fmt.Sprintf("notification.update.%s", data.ID),
			}, model.NotificationModel(data)
		},
		func(data *Notification) ([]string, any) {
			return []string{
				"notification.delete",
				fmt.Sprintf("notification.delete.user.%s", data.UserID),
				fmt.Sprintf("notification.delete.%s", data.ID),
			}, model.NotificationModel(data)
		},
		[]string{},
	)
	return &NotificationCollection{
		Manager: manager,
	}, nil
}

// notification/user/:user_id
func (fc *NotificationCollection) ListByUser(userID uuid.UUID) ([]*Notification, error) {
	return fc.Manager.Find(&Notification{
		UserID: userID,
	})
}

// notification/user/:user_id/unviewed-count
func (fc *NotificationCollection) ListByUserUnviewedCount(userID uuid.UUID) (int64, error) {
	return fc.Manager.Count(&Notification{
		UserID:   userID,
		IsViewed: false,
	})
}

// notification/user/:user_id/unviewed
func (fc *NotificationCollection) ListByUserUnviewed(userID uuid.UUID) ([]*Notification, error) {
	return fc.Manager.Find(&Notification{
		UserID:   userID,
		IsViewed: false,
	})
}

func (fc *NotificationCollection) ReadAll(userID uuid.UUID) ([]*Notification, error) {
	notifications, err := fc.Manager.Find(&Notification{
		UserID:   userID,
		IsViewed: false,
	})
	if err != nil {
		return nil, err
	}

	var updated []*Notification

	for _, notif := range notifications {
		notif.IsViewed = true
		if err := fc.Manager.Update(notif); err != nil {
			continue
		}
		updated = append(updated, notif)
	}

	return updated, nil
}
