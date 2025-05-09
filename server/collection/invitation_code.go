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
	InvitationCode struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt      time.Time      `gorm:"not null;default:now()"`
		CreatedByID    uuid.UUID      `gorm:"type:uuid"`
		CreatedBy      *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()"`
		UpdatedByID    uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy      *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt      gorm.DeletedAt `gorm:"index"`
		DeletedByID    *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy      *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		UserType       string         `gorm:"type:varchar(255);not null"`
		Code           string         `gorm:"type:varchar(255);not null;unique"`
		ExpirationDate time.Time      `gorm:"not null"`
		MaxUse         int            `gorm:"not null"`
		CurrentUse     int            `gorm:"default:0"`
		Description    string         `gorm:"type:text"`
	}

	InvitationCodeRequest struct {
		UserType       string    `json:"user_type" validate:"required,oneof=employee owner member"`
		Code           string    `json:"code" validate:"required,max=255"`
		ExpirationDate time.Time `json:"expiration_date" validate:"required"`
		MaxUse         int       `json:"max_use" validate:"required"`
		Description    string    `json:"description,omitempty"`
	}

	InvitationCodeResponse struct {
		ID             uuid.UUID     `json:"id"`
		CreatedAt      string        `json:"created_at"`
		CreatedByID    uuid.UUID     `json:"created_by_id"`
		CreatedBy      *UserResponse `json:"created_by,omitempty"`
		UpdatedAt      string        `json:"updated_at"`
		UpdatedByID    uuid.UUID     `json:"updated_by_id"`
		UpdatedBy      *UserResponse `json:"updated_by,omitempty"`
		UserType       string        `json:"user_type"`
		Code           string        `json:"code"`
		ExpirationDate string        `json:"expiration_date"`
		MaxUse         int           `json:"max_use"`
		CurrentUse     int           `json:"current_use"`
		Description    string        `json:"description,omitempty"`
	}

	InvitationCodeCollection struct {
		validator *validator.Validate
		userCol   *UserCollection
	}
)

func NewInvitationCodeCollection(userCol *UserCollection) *InvitationCodeCollection {
	return &InvitationCodeCollection{
		validator: validator.New(),
		userCol:   userCol,
	}
}

func (icc *InvitationCodeCollection) ValidateCreate(c echo.Context) (*InvitationCodeRequest, error) {
	req := new(InvitationCodeRequest)
	if err := c.Bind(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := icc.validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return req, nil
}

func (icc *InvitationCodeCollection) ToModel(ic *InvitationCode) *InvitationCodeResponse {
	if ic == nil {
		return nil
	}
	return &InvitationCodeResponse{
		ID:             ic.ID,
		CreatedAt:      ic.CreatedAt.Format(time.RFC3339),
		CreatedByID:    ic.CreatedByID,
		CreatedBy:      icc.userCol.ToModel(ic.CreatedBy),
		UpdatedAt:      ic.UpdatedAt.Format(time.RFC3339),
		UpdatedByID:    ic.UpdatedByID,
		UpdatedBy:      icc.userCol.ToModel(ic.UpdatedBy),
		UserType:       ic.UserType,
		Code:           ic.Code,
		ExpirationDate: ic.ExpirationDate.Format(time.RFC3339),
		MaxUse:         ic.MaxUse,
		CurrentUse:     ic.CurrentUse,
		Description:    ic.Description,
	}
}

func (icc *InvitationCodeCollection) ToModels(data []*InvitationCode) []*InvitationCodeResponse {
	if len(data) == 0 {
		return []*InvitationCodeResponse{}
	}
	out := make([]*InvitationCodeResponse, 0, len(data))
	for _, ic := range data {
		if m := icc.ToModel(ic); m != nil {
			out = append(out, m)
		}
	}
	return out
}
