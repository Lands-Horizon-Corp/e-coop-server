package modelcore

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// OrganizationCategory represents the categorization of organizations
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

	// OrganizationCategoryResponse represents the JSON response structure for organization category data
	OrganizationCategoryResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`

		OrganizationID *uuid.UUID            `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization"`
		CategoryID     *uuid.UUID            `json:"category_id"`
		Category       *CategoryResponse     `json:"category"`
	}

	// OrganizationCategoryRequest represents the request payload for organization category operations
	OrganizationCategoryRequest struct {
		ID         *uuid.UUID `json:"id,omitempty"`
		CategoryID uuid.UUID  `json:"category_id" validate:"required"`
	}
)

// OrganizationCategory initializes the OrganizationCategory model and its repository manager
func (m *ModelCore) OrganizationCategory() {
	m.Migration = append(m.Migration, &OrganizationCategory{})
	m.OrganizationCategoryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[OrganizationCategory, OrganizationCategoryResponse, OrganizationCategoryRequest]{
		Preloads: []string{"Organization", "Category"},
		Service:  m.provider.Service,
		Resource: func(data *OrganizationCategory) *OrganizationCategoryResponse {
			if data == nil {
				return nil
			}
			return &OrganizationCategoryResponse{
				ID:        data.ID,
				CreatedAt: data.CreatedAt.Format(time.RFC3339),
				UpdatedAt: data.UpdatedAt.Format(time.RFC3339),

				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				CategoryID:     data.CategoryID,
				Category:       m.CategoryManager.ToModel(data.Category),
			}
		},

		Created: func(data *OrganizationCategory) []string {
			return []string{
				"organization_category.create",
				fmt.Sprintf("organization_category.create.%s", data.ID),
				fmt.Sprintf("organization_category.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *OrganizationCategory) []string {
			return []string{
				"organization_category.update",
				fmt.Sprintf("organization_category.update.%s", data.ID),
				fmt.Sprintf("organization_category.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *OrganizationCategory) []string {
			return []string{
				"organization_category.delete",
				fmt.Sprintf("organization_category.delete.%s", data.ID),
				fmt.Sprintf("organization_category.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// GetOrganizationCategoryByOrganization retrieves all categories assigned to a specific organization
func (m *ModelCore) GetOrganizationCategoryByOrganization(context context.Context, organizationID uuid.UUID) ([]*OrganizationCategory, error) {
	return m.OrganizationCategoryManager.Find(context, &OrganizationCategory{
		OrganizationID: &organizationID,
	})
}
