package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

func OrganizationCategoryManager(service *horizon.HorizonService) *registry.Registry[
	types.OrganizationCategory, types.OrganizationCategoryResponse, types.OrganizationCategoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.OrganizationCategory, types.OrganizationCategoryResponse, types.OrganizationCategoryRequest]{
		Preloads: []string{"Organization", "Category"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.OrganizationCategory) *types.OrganizationCategoryResponse {
			if data == nil {
				return nil
			}
			return &types.OrganizationCategoryResponse{
				ID:        data.ID,
				CreatedAt: data.CreatedAt.Format(time.RFC3339),
				UpdatedAt: data.UpdatedAt.Format(time.RFC3339),

				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				CategoryID:     data.CategoryID,
				Category:       CategoryManager(service).ToModel(data.Category),
			}
		},

		Created: func(data *types.OrganizationCategory) registry.Topics {
			return []string{
				"organization_category.create",
				fmt.Sprintf("organization_category.create.%s", data.ID),
				fmt.Sprintf("organization_category.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.OrganizationCategory) registry.Topics {
			return []string{
				"organization_category.update",
				fmt.Sprintf("organization_category.update.%s", data.ID),
				fmt.Sprintf("organization_category.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.OrganizationCategory) registry.Topics {
			return []string{
				"organization_category.delete",
				fmt.Sprintf("organization_category.delete.%s", data.ID),
				fmt.Sprintf("organization_category.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GetOrganizationCategoryByOrganization(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID) ([]*types.OrganizationCategory, error) {
	return OrganizationCategoryManager(service).Find(context, &types.OrganizationCategory{
		OrganizationID: &organizationID,
	})
}
