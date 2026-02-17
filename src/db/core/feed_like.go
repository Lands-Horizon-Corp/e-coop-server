package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

func FeedLikeManager(service *horizon.HorizonService) *registry.Registry[
	types.FeedLike, types.FeedLikeResponse, types.FeedLikeRequest] {
	return registry.GetRegistry(registry.RegistryParams[types.FeedLike, types.FeedLikeResponse, types.FeedLikeRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Organization", "Branch", "User"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.FeedLike) *types.FeedLikeResponse {
			if data == nil {
				return nil
			}
			return &types.FeedLikeResponse{
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
				UserID:         data.UserID,
				User:           UserManager(service).ToModel(data.User),
			}
		},
		Created: func(data *types.FeedLike) registry.Topics {
			return []string{
				"feed.like.create",
				fmt.Sprintf("feed.like.create.%s", data.ID),
				fmt.Sprintf("feed.%s.like.added", data.FeedID),
				fmt.Sprintf("feed.like.create.branch.%s", data.BranchID),
			}
		},
		Updated: func(data *types.FeedLike) registry.Topics {
			return []string{
				"feed.like.update",
				fmt.Sprintf("feed.like.update.%s", data.ID),
			}
		},
		Deleted: func(data *types.FeedLike) registry.Topics {
			return []string{
				"feed.like.delete",
				fmt.Sprintf("feed.like.delete.%s", data.ID),
				fmt.Sprintf("feed.%s.like.removed", data.FeedID),
			}
		},
	})
}
