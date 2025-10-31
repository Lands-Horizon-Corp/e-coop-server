package modelCore

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// Notification represents a system notification sent to users
	Notification struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
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

	// NotificationResponse represents the JSON response structure for notification data
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
)

// Notification initializes the Notification model and its repository manager
func (m *ModelCore) Notification() {
	m.Migration = append(m.Migration, &Notification{})
	m.NotificationManager = horizon_services.NewRepository(horizon_services.RepositoryParams[Notification, NotificationResponse, any]{
		Preloads: nil,
		Service:  m.provider.Service,
		Resource: func(data *Notification) *NotificationResponse {
			if data == nil {
				return nil
			}
			return &NotificationResponse{
				ID:               data.ID,
				UserID:           data.UserID,
				User:             m.UserManager.ToModel(data.User),
				Title:            data.Title,
				Description:      data.Description,
				IsViewed:         data.IsViewed,
				NotificationType: data.NotificationType,
				CreatedAt:        data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:        data.UpdatedAt.Format(time.RFC3339),
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
func (m *ModelCore) GetNotificationByUser(context context.Context, userID uuid.UUID) ([]*Notification, error) {
	return m.NotificationManager.Find(context, &Notification{
		UserID: userID,
	})
}
