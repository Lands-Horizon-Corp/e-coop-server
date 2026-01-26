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

func OrganizationMediaManager(service *horizon.HorizonService) *registry.Registry[
	types.OrganizationMedia, types.OrganizationMediaResponse, types.OrganizationMediaRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.OrganizationMedia, types.OrganizationMediaResponse, types.OrganizationMediaRequest]{
		Preloads: []string{"Media"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.OrganizationMedia) *types.OrganizationMediaResponse {
			if data == nil {
				return nil
			}
			return &types.OrganizationMediaResponse{
				ID:        data.ID,
				CreatedAt: data.CreatedAt.Format(time.RFC3339),
				UpdatedAt: data.UpdatedAt.Format(time.RFC3339),

				Name:        data.Name,
				Description: data.Description,

				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(&data.Organization),

				MediaID: data.MediaID,
				Media:   MediaManager(service).ToModel(&data.Media),
			}
		},
		Created: func(data *types.OrganizationMedia) registry.Topics {
			return []string{
				"organization_media.create",
				fmt.Sprintf("organization_media.create.%s", data.ID),
				fmt.Sprintf("organization_media.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.OrganizationMedia) registry.Topics {
			return []string{
				"organization_media.update",
				fmt.Sprintf("organization_media.update.%s", data.ID),
				fmt.Sprintf("organization_media.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.OrganizationMedia) registry.Topics {
			return []string{
				"organization_media.delete",
				fmt.Sprintf("organization_media.delete.%s", data.ID),
				fmt.Sprintf("organization_media.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func OrganizationMediaFindByOrganization(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID) ([]*types.OrganizationMedia, error) {
	return OrganizationMediaManager(service).Find(context, &types.OrganizationMedia{
		OrganizationID: organizationID,
	})
}

func OrganizationMediaCreateForOrganization(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, mediaID uuid.UUID,
	name string, description *string) (*types.OrganizationMedia, error) {
	organizationMedia := &types.OrganizationMedia{
		Name:           name,
		Description:    description,
		OrganizationID: organizationID,
		MediaID:        mediaID,
	}

	if err := OrganizationMediaManager(service).Create(context, organizationMedia); err != nil {
		return nil, err
	}

	return organizationMedia, nil
}
