package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	OrganizationCategory struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
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
)

func OrganizationCategoryManager(service *horizon.HorizonService) *registry.Registry[OrganizationCategory, OrganizationCategoryResponse, OrganizationCategoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[OrganizationCategory, OrganizationCategoryResponse, OrganizationCategoryRequest]{
		Preloads: []string{"Organization", "Category"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *OrganizationCategory) *OrganizationCategoryResponse {
			if data == nil {
				return nil
			}
			return &OrganizationCategoryResponse{
				ID:        data.ID,
				CreatedAt: data.CreatedAt.Format(time.RFC3339),
				UpdatedAt: data.UpdatedAt.Format(time.RFC3339),

				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				CategoryID:     data.CategoryID,
				Category:       CategoryManager(service).ToModel(data.Category),
			}
		},

		Created: func(data *OrganizationCategory) registry.Topics {
			return []string{
				"organization_category.create",
				fmt.Sprintf("organization_category.create.%s", data.ID),
				fmt.Sprintf("organization_category.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *OrganizationCategory) registry.Topics {
			return []string{
				"organization_category.update",
				fmt.Sprintf("organization_category.update.%s", data.ID),
				fmt.Sprintf("organization_category.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *OrganizationCategory) registry.Topics {
			return []string{
				"organization_category.delete",
				fmt.Sprintf("organization_category.delete.%s", data.ID),
				fmt.Sprintf("organization_category.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GetOrganizationCategoryByOrganization(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID) ([]*OrganizationCategory, error) {
	return OrganizationCategoryManager(service).Find(context, &OrganizationCategory{
		OrganizationID: &organizationID,
	})
}
