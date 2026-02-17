package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

func FeedMediaManager(service *horizon.HorizonService) *registry.Registry[
	types.FeedMedia, types.FeedMediaResponse, types.FeedMediaRequest] {
	return registry.GetRegistry(registry.RegistryParams[types.FeedMedia, types.FeedMediaResponse, types.FeedMediaRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Organization", "Branch", "Media"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.FeedMedia) *types.FeedMediaResponse {
			if data == nil {
				return nil
			}
			return &types.FeedMediaResponse{
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
				FeedID:         data.FeedID,
				MediaID:        data.MediaID,
				Media:          MediaManager(service).ToModel(data.Media),
			}
		},
		Created: func(data *types.FeedMedia) registry.Topics {
			return []string{
				"feed.media.create",
				fmt.Sprintf("feed.media.create.%s", data.ID),
				fmt.Sprintf("feed.%s.media.added", data.FeedID),
			}
		},
		Updated: func(data *types.FeedMedia) registry.Topics {
			return []string{
				"feed.media.update",
				fmt.Sprintf("feed.media.update.%s", data.ID),
			}
		},
		Deleted: func(data *types.FeedMedia) registry.Topics {
			return []string{
				"feed.media.delete",
				fmt.Sprintf("feed.media.delete.%s", data.ID),
				fmt.Sprintf("feed.%s.media.removed", data.FeedID),
			}
		},
	})
}
