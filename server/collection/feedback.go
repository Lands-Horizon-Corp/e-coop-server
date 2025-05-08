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
		ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		Email        string    `gorm:"type:varchar(255)"`
		Description  string    `gorm:"type:text"`
		FeedbackType string    `gorm:"type:varchar(50);not null;default:'general'"`

		MediaID *uuid.UUID `gorm:"type:uuid"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;"`

		CreatedAt time.Time  `gorm:"not null;default:now()"`
		UpdatedAt time.Time  `gorm:"not null;default:now()"`
		DeletedAt *time.Time `json:"deletedAt,omitempty" gorm:"index"`
	}
	FeedbackResponse struct {
		ID           uuid.UUID      `json:"id"`
		Email        string         `json:"email"`
		Description  string         `json:"description"`
		FeedbackType string         `json:"feedback_type"`
		MediaID      uuid.UUID      `json:"media_id"`
		Media        *MediaResponse `json:"media,omitempty"`

		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		DeletedAt string `gorm:"index"`
	}

	FeedbackRequest struct {
		Email        string     `json:"email"        validate:"required,email"`
		Description  string     `json:"description"  validate:"required,min=5,max=2000"`
		FeedbackType string     `json:"feedback_type" validate:"required,oneof=general bug feature"`
		MediaID      *uuid.UUID `json:"media_id,omitempty"`
	}
	FeedbackCollection struct {
		validator *validator.Validate
	}
)

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
		Email:        data.Email,
		Description:  data.Description,
		FeedbackType: data.FeedbackType,
	}
}

func (fc *FeedbackCollection) ToModels(data []*Feedback) []*FeedbackResponse {
	if data == nil {
		return make([]*FeedbackResponse, 0)
	}
	var feedbackResponses []*FeedbackResponse
	for _, feedback := range data {

		model := fc.ToModel(feedback)
		if model != nil {
			feedbackResponses = append(feedbackResponses, model)
		}

	}
	if len(feedbackResponses) <= 0 {
		return make([]*FeedbackResponse, 0)
	}
	return feedbackResponses
}
