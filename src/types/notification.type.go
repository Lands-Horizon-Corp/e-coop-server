package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
	NotificationType string
	Notification     struct {
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
