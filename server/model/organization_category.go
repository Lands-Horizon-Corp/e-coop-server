package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	horizon_manager "horizon.com/server/horizon/manager"
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

	OrganizationCategoryResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`

		OrganizationID *uuid.UUID            `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization"`
		CategoryID     *uuid.UUID            `json:"category_id"`
		Category       *CategoryResponse     `json:"category"`
	}

	OrganizationCategoryRequest struct {
		ID         *uuid.UUID `json:"id,omitempty"`
		CategoryID uuid.UUID  `json:"category_id" validate:"required"`
	}

	OrganizationCategoryCollection struct {
		Manager horizon_manager.CollectionManager[OrganizationCategory]
	}
)

func (m *Model) OrganizationCategoryValidate(ctx echo.Context) (*OrganizationCategoryRequest, error) {
	return horizon_manager.Validate[OrganizationCategoryRequest](ctx, m.validator)
}

func (m *Model) OrganizationCategoryModel(data *OrganizationCategory) *OrganizationCategoryResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *OrganizationCategory) *OrganizationCategoryResponse {
		return &OrganizationCategoryResponse{
			OrganizationID: data.OrganizationID,
			Organization:   m.OrganizationModel(data.Organization),
			CategoryID:     data.CategoryID,
			Category:       m.CategoryModel(data.Category),
		}
	})
}

func (m *Model) OrganizationCategoryModels(data []*OrganizationCategory) []*OrganizationCategoryResponse {
	return horizon_manager.ToModels(data, m.OrganizationCategoryModel)
}

func NewOrganizationCategoryCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*OrganizationCategoryCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *OrganizationCategory) ([]string, any) {
			return []string{
				"organization_category.create",
				fmt.Sprintf("organization_category.create.%s", data.ID),
				fmt.Sprintf("organization_category.create.organization.%s", data.OrganizationID),
			}, model.OrganizationCategoryModel(data)
		},
		func(data *OrganizationCategory) ([]string, any) {
			return []string{
				"organization_category.update",
				fmt.Sprintf("organization_category.update.%s", data.ID),
				fmt.Sprintf("organization_category.update.organization.%s", data.OrganizationID),
			}, model.OrganizationCategoryModel(data)
		},
		func(data *OrganizationCategory) ([]string, any) {
			return []string{
				"organization_category.delete",
				fmt.Sprintf("organization_category.delete.%s", data.ID),
				fmt.Sprintf("organization_category.delete.organization.%s", data.OrganizationID),
			}, model.OrganizationCategoryModel(data)
		},
		[]string{
			"Category",
		},
	)
	return &OrganizationCategoryCollection{
		Manager: manager,
	}, nil
}

// organization-category/category/:category_id
func (fc *OrganizationCategoryCollection) ListByCategory(categoryId *uuid.UUID) ([]*OrganizationCategory, error) {
	return fc.Manager.Find(&OrganizationCategory{
		CategoryID: categoryId,
	})
}

// organization-category/organizaton/:organization_id
func (fc *OrganizationCategoryCollection) ListByOrganization(organizationId *uuid.UUID) ([]*OrganizationCategory, error) {
	return fc.Manager.Find(&OrganizationCategory{
		OrganizationID: organizationId,
	})
}
