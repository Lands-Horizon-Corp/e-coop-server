package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type (
	OrganizationCategory struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		OrganizationID *uuid.UUID    `gorm:"type:uuid;not null"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`

		CategoryID *uuid.UUID `gorm:"type:uuid;not null"`
		Category   *Category  `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE"`
	}

	OrganizationCategoryRequest struct {
		OrganizationID uuid.UUID `json:"organization_id" validate:"required,uuid4"`
		CategoryID     uuid.UUID `json:"category_id" validate:"required,uuid4"`
	}

	OrganizationCategoryResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`

		OrganizationID *uuid.UUID            `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization"`
		CategoryID     *uuid.UUID            `json:"category_id"`
		Category       *CategoryResponse     `json:"category"`
	}
)

func (m *Model) OrganizationCategoryValidate(ctx echo.Context) (*OrganizationCategoryRequest, error) {
	return Validate[OrganizationCategoryRequest](ctx, m.validator)
}

func (m *Model) OrganizationCategoryModel(data *OrganizationCategory) *OrganizationCategoryResponse {
	return ToModel(data, func(data *OrganizationCategory) *OrganizationCategoryResponse {
		return &OrganizationCategoryResponse{
			OrganizationID: data.OrganizationID,
			Organization:   m.OrganizationModel(data.Organization),
			CategoryID:     data.CategoryID,
			Category:       m.CategoryModel(data.Category),
		}
	})
}

func (m *Model) OrganizationCategoryModels(data []*OrganizationCategory) []*OrganizationCategoryResponse {
	return ToModels(data, m.OrganizationCategoryModel)
}
