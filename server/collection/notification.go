package collection

import (
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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

	NotificationRequest struct {
		UserID           uuid.UUID `json:"user_id" validate:"required"`
		Title            string    `json:"title" validate:"required,min=1,max=255"`
		Description      string    `json:"description" validate:"required,min=5,max=2000"`
		IsViewed         bool      `json:"is_viewed"`
		NotificationType string    `json:"notification_type" validate:"required,oneof=success warning alert info error general report message"`
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
		validator *validator.Validate
		userCol   *UserCollection
	}
)

func NewNotificationCollection(userCol *UserCollection) (*NotificationCollection, error) {
	return &NotificationCollection{
		validator: validator.New(),
		userCol:   userCol,
	}, nil
}

// ValidateCreate binds and validates a NotificationRequest
func (nc *NotificationCollection) ValidateCreate(c echo.Context) (*NotificationRequest, error) {
	req := new(NotificationRequest)
	if err := c.Bind(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := nc.validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return req, nil
}

// ToModel maps a Notification DB model to NotificationResponse
func (nc *NotificationCollection) ToModel(n *Notification) *NotificationResponse {
	if n == nil {
		return nil
	}
	resp := &NotificationResponse{
		ID:               n.ID,
		UserID:           n.UserID,
		User:             nc.userCol.ToModel(n.User),
		Title:            n.Title,
		Description:      n.Description,
		IsViewed:         n.IsViewed,
		NotificationType: n.NotificationType,
		CreatedAt:        n.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        n.UpdatedAt.Format(time.RFC3339),
	}
	return resp
}

// ToModels maps a slice of Notification DB models to NotificationResponse
func (nc *NotificationCollection) ToModels(data []*Notification) []*NotificationResponse {
	if data == nil {
		return []*NotificationResponse{}
	}
	var out []*NotificationResponse
	for _, n := range data {
		if m := nc.ToModel(n); m != nil {
			out = append(out, m)
		}
	}
	return out
}
