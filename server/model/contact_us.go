package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type (
	ContactUs struct {
		ID            uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
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
		FirstName     string    `json:"firstName"`
		LastName      string    `json:"lastName,omitempty"`
		Email         string    `json:"email,omitempty"`
		ContactNumber string    `json:"contactNumber,omitempty"`
		Description   string    `json:"description"`
		CreatedAt     string    `json:"createdAt"`
		UpdatedAt     string    `json:"updatedAt"`
	}

	ContactUsRequest struct {
		FirstName     string `json:"firstName" validate:"required,min=1,max=255"`
		LastName      string `json:"lastName,omitempty" validate:"omitempty,min=1,max=255"`
		Email         string `json:"email,omitempty" validate:"omitempty,email,max=255"`
		ContactNumber string `json:"contactNumber,omitempty" validate:"omitempty,min=1,max=20"`
		Description   string `json:"description" validate:"required,min=1"`
	}
)

func (m *Model) ContactUsValidate(ctx echo.Context) (*ContactUsRequest, error) {
	return Validate[ContactUsRequest](ctx, m.validator)
}

func (m *Model) ContactUsModel(data *ContactUs) *ContactUsResponse {
	return ToModel(data, func(data *ContactUs) *ContactUsResponse {
		return &ContactUsResponse{
			ID:            data.ID,
			FirstName:     data.FirstName,
			LastName:      data.LastName,
			Email:         data.Email,
			ContactNumber: data.ContactNumber,
			Description:   data.Description,
			CreatedAt:     data.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     data.UpdatedAt.Format(time.RFC3339),
		}
	})
}

func (m *Model) ContactUsModels(data []*ContactUs) []*ContactUsResponse {
	return ToModels(data, m.ContactUsModel)
}
