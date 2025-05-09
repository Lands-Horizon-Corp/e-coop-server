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

	ContactUsCollection struct {
		validator *validator.Validate
	}
)

func NewContactUsCollection() *ContactUsCollection {
	return &ContactUsCollection{}
}

func (c *ContactUsCollection) ValidateCreate(ctx echo.Context) (*ContactUsRequest, error) {
	req := new(ContactUsRequest)
	if err := ctx.Bind(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return req, nil
}

func (c *ContactUsCollection) ToModel(m *ContactUs) *ContactUsResponse {
	if m == nil {
		return nil
	}
	return &ContactUsResponse{
		ID:            m.ID,
		FirstName:     m.FirstName,
		LastName:      m.LastName,
		Email:         m.Email,
		ContactNumber: m.ContactNumber,
		Description:   m.Description,
		CreatedAt:     m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     m.UpdatedAt.Format(time.RFC3339),
	}
}

func (c *ContactUsCollection) ToModels(data []*ContactUs) []*ContactUsResponse {
	if data == nil {
		return []*ContactUsResponse{}
	}
	var out []*ContactUsResponse
	for _, o := range data {
		if m := c.ToModel(o); m != nil {
			out = append(out, m)
		}
	}
	return out
}
