package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
)

type (
	Category struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Name                   string                  `gorm:"type:varchar(255);not null"`
		Description            string                  `gorm:"type:text"`
		Color                  string                  `gorm:"type:varchar(50)"`
		Icon                   string                  `gorm:"type:varchar(50)"`
		OrganizationCategories []*OrganizationCategory `gorm:"foreignKey:CategoryID"`
	}

	CategoryResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"createdAt"`
		UpdatedAt string    `json:"updatedAt"`

		Name                   string                          `json:"name"`
		Description            string                          `json:"description"`
		Color                  string                          `json:"color"`
		Icon                   string                          `json:"icon"`
		OrganizationCategories []*OrganizationCategoryResponse `json:"organizaton_categories"`
	}

	CategoryRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description" validate:"required,min=1,max=2048"`
		Color       string `json:"color" validate:"required,min=1,max=50"`
		Icon        string `json:"icon" validate:"required,min=1,max=50"`
	}

	CategoryCollection struct {
		Manager CollectionManager[Category]
	}
)

func (m *Model) CategoryValidate(ctx echo.Context) (*CategoryRequest, error) {
	return Validate[CategoryRequest](ctx, m.validator)
}

func (m *Model) CategoryModel(data *Category) *CategoryResponse {
	return ToModel(data, func(data *Category) *CategoryResponse {
		return &CategoryResponse{
			ID:                     data.ID,
			CreatedAt:              data.CreatedAt.Format(time.RFC3339),
			UpdatedAt:              data.UpdatedAt.Format(time.RFC3339),
			Name:                   data.Name,
			Description:            data.Description,
			Color:                  data.Color,
			Icon:                   data.Icon,
			OrganizationCategories: m.OrganizationCategoryModels(data.OrganizationCategories),
		}
	})
}

func (m *Model) CategoryModels(data []*Category) []*CategoryResponse {
	return ToModels(data, m.CategoryModel)
}

func NewCategoryCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*CategoryCollection, error) {
	manager := NewcollectionManager(
		database,
		broadcast,
		func(data *Category) ([]string, any) {
			return []string{
				"category.create",
				fmt.Sprintf("category.create.%s", data.ID),
			}, model.CategoryModel(data)
		},
		func(data *Category) ([]string, any) {
			return []string{
				"category.update",
				fmt.Sprintf("category.update.%s", data.ID),
			}, model.CategoryModel(data)
		},
		func(data *Category) ([]string, any) {
			return []string{
				"category.delete",
				fmt.Sprintf("category.delete.%s", data.ID),
			}, model.CategoryModel(data)
		},
	)
	return &CategoryCollection{
		Manager: manager,
	}, nil
}
