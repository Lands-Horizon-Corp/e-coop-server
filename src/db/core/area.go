package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

func AreaManager(service *horizon.HorizonService) *registry.Registry[
	types.Area, types.AreaResponse, types.AreaRequest] {

	return registry.GetRegistry(registry.RegistryParams[types.Area, types.AreaResponse, types.AreaRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Area) *types.AreaResponse {
			if data == nil {
				return nil
			}
			return &types.AreaResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),
				MediaID:        data.MediaID,
				Media:          MediaManager(service).ToModel(data.Media),
				Name:           data.Name,
				Latitude:       data.Latitude,
				Longitude:      data.Longitude,
			}
		},
		Created: func(data *types.Area) registry.Topics {
			return []string{
				"area.create",
				fmt.Sprintf("area.create.%s", data.ID),
				fmt.Sprintf("area.create.branch.%s", data.BranchID),
				fmt.Sprintf("area.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.Area) registry.Topics {
			return []string{
				"area.update",
				fmt.Sprintf("area.update.%s", data.ID),
				fmt.Sprintf("area.update.branch.%s", data.BranchID),
				fmt.Sprintf("area.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.Area) registry.Topics {
			return []string{
				"area.delete",
				fmt.Sprintf("area.delete.%s", data.ID),
				fmt.Sprintf("area.delete.branch.%s", data.BranchID),
				fmt.Sprintf("area.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}
