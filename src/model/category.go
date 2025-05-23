package model

import (
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	Category struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Name                   string                  `gorm:"type:varchar(255);not null"`
		Description            string                  `gorm:"type:text"`
		Color                  string                  `gorm:"type:varchar(50)"`
		Icon                   string                  `gorm:"type:varchar(50)"`
		OrganizationCategories []*OrganizationCategory `gorm:"foreignKey:CategoryID"` // organization category
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
		ID *uuid.UUID `json:"id,omitempty"`

		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description" validate:"required,min=1,max=2048"`
		Color       string `json:"color" validate:"required,min=1,max=50"`
		Icon        string `json:"icon" validate:"required,min=1,max=50"`
	}
)

func (m *Model) Category() {
	m.Migration = append(m.Migration, &Category{})
	m.CategoryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[Category, CategoryResponse, CategoryRequest]{
		Preloads: []string{"OrganizationCategories"},
		Service:  m.provider.Service,
		Resource: func(data *Category) *CategoryResponse {
			if data == nil {
				return nil
			}
			return &CategoryResponse{
				ID:        data.ID,
				CreatedAt: data.CreatedAt.Format(time.RFC3339),
				UpdatedAt: data.UpdatedAt.Format(time.RFC3339),

				Name:        data.Name,
				Description: data.Description,
				Color:       data.Color,
				Icon:        data.Icon,

				OrganizationCategories: m.OrganizationCategoryManager.ToModels(data.OrganizationCategories),
			}
		},
		Created: func(data *Category) []string {
			return []string{
				"category.create",
				"category.create." + data.ID.String(),
			}
		},
		Updated: func(data *Category) []string {
			return []string{
				"category.update",
				"category.update." + data.ID.String(),
			}
		},
		Deleted: func(data *Category) []string {
			return []string{
				"category.delete",
				"category.delete." + data.ID.String(),
			}
		},
	})
}
