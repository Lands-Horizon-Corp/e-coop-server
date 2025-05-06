package collection

import (
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type (
	Feedback struct {
		ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		Email        string     `gorm:"type:varchar(255)"`
		Description  string     `gorm:"type:text"`
		FeedbackType string     `gorm:"type:varchar(50);not null;default:'general'"`
		CreatedAt    time.Time  `gorm:"not null;default:now()"`
		UpdatedAt    time.Time  `gorm:"not null;default:now()"`
		DeletedAt    *time.Time `gorm:"index"`
	}
	FeedbackResponse struct {
		ID           uuid.UUID `json:"id"`
		Email        string    `json:"email"`
		Description  string    `json:"description"`
		FeedbackType string    `json:"feedbackType"`
		CreatedAt    string    `json:"createdAt"`
		UpdatedAt    string    `json:"updatedAt"`
		DeletedAt    string    `gorm:"index"`
	}

	FeedbackRequest struct {
		Email        string `json:"email"        validate:"required,email"`
		Description  string `json:"description"  validate:"required,min=5,max=2000"`
		FeedbackType string `json:"feedbackType" validate:"required,oneof=general bug feature"`
	}
)
type FeedbackCollection struct {
	validator *validator.Validate
}

func NewFeedbackCollection() (*FeedbackCollection, error) {
	return &FeedbackCollection{
		validator: validator.New(),
	}, nil
}

func (fc *FeedbackCollection) ValidateCreate(c echo.Context) (*FeedbackRequest, error) {
	u := new(FeedbackRequest)
	if err := c.Bind(u); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := fc.validator.Struct(u); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return u, nil
}

func (fc *FeedbackCollection) ToModel(data *Feedback) *FeedbackResponse {
	if data == nil {
		return nil
	}
	return &FeedbackResponse{
		ID:           data.ID,
		CreatedAt:    data.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    data.UpdatedAt.Format(time.RFC3339),
		DeletedAt:    data.DeletedAt.Format(time.RFC3339),
		Email:        data.Email,
		Description:  data.Description,
		FeedbackType: data.FeedbackType,
	}
}

func (fc *FeedbackCollection) ToModels(data []*Feedback) []*FeedbackResponse {
	if data == nil {
		return nil
	}
	var feedbackResources []*FeedbackResponse
	for _, feedback := range data {
		feedbackResources = append(feedbackResources, fc.ToModel(feedback))
	}
	return feedbackResources
}
