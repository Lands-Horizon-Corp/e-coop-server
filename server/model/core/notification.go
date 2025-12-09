package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationType string

const (
	NotificationSuccess NotificationType = "success"
	NotificationError   NotificationType = "error"
	NotificationWarning NotificationType = "warning"
	NotificationInfo    NotificationType = "info"
	NotificationDebug   NotificationType = "debug"
	NotificationAlert   NotificationType = "alert"
	NotificationMessage NotificationType = "message"
	NotificationSystem  NotificationType = "system"
)

type (
	// Notification represents a system notification sent to users
	Notification struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		UserID uuid.UUID `gorm:"type:uuid;not null"`
		User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`

		RecipientID *uuid.UUID `gorm:"type:uuid"`
		Recipient   *User      `gorm:"foreignKey:RecipientID;constraint:OnDelete:CASCADE;" json:"recipient,omitempty"`

		Title            string           `gorm:"type:varchar(255);not null"`
		Description      string           `gorm:"type:text;not null"`
		IsViewed         bool             `gorm:"default:false" json:"is_viewed"`
		NotificationType NotificationType `gorm:"type:varchar(50);not null"`

		UserType UserOrganizationType `gorm:"type:varchar(50);not null"`
	}

	// NotificationResponse represents the JSON response structure for notification data
	NotificationResponse struct {
		ID     uuid.UUID     `json:"id"`
		UserID uuid.UUID     `json:"user_id"`
		User   *UserResponse `json:"user,omitempty"`

		RecipientID *uuid.UUID    `json:"recipient_id"`
		Recipient   *UserResponse `json:"recipient,omitempty"`

		Title            string           `json:"title"`
		Description      string           `json:"description"`
		IsViewed         bool             `json:"is_viewed"`
		NotificationType NotificationType `json:"notification_type"`
		CreatedAt        string           `json:"created_at"`
		UpdatedAt        string           `json:"updated_at"`

		UserType UserOrganizationType `json:"user_type"`
	}
)

// Notification initializes the Notification model and its repository manager
func (m *Core) notification() {
	m.Migration = append(m.Migration, &Notification{})
	m.NotificationManager = *registry.NewRegistry(registry.RegistryParams[Notification, NotificationResponse, any]{
		Preloads: []string{"Recipient", "Recipient.Media"},
		Database: m.provider.Service.Database.Client(),
Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		}
		Resource: func(data *Notification) *NotificationResponse {
			if data == nil {
				return nil
			}
			return &NotificationResponse{
				ID:               data.ID,
				UserID:           data.UserID,
				User:             m.UserManager.ToModel(data.User),
				RecipientID:      data.RecipientID,
				Recipient:        m.UserManager.ToModel(data.Recipient),
				Title:            data.Title,
				Description:      data.Description,
				IsViewed:         data.IsViewed,
				NotificationType: data.NotificationType,
				CreatedAt:        data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:        data.UpdatedAt.Format(time.RFC3339),
				UserType:         data.UserType,
			}
		},
		Created: func(data *Notification) []string {
			return []string{
				"notification.create",
				fmt.Sprintf("notification.create.user.%s", data.UserID),
				fmt.Sprintf("notification.create.%s", data.ID),
			}
		},
		Updated: func(data *Notification) []string {
			return []string{
				"notification.update",
				fmt.Sprintf("notification.update.user.%s", data.UserID),
				fmt.Sprintf("notification.update.%s", data.ID),
			}
		},
		Deleted: func(data *Notification) []string {
			return []string{
				"notification.delete",
				fmt.Sprintf("notification.delete.user.%s", data.UserID),
				fmt.Sprintf("notification.delete.%s", data.ID),
			}
		},
	})
}

// GetNotificationByUser retrieves all notifications for a specific user
func (m *Core) GetNotificationByUser(context context.Context, userID uuid.UUID) ([]*Notification, error) {
	return m.NotificationManager.Find(context, &Notification{
		UserID: userID,
	})
}
