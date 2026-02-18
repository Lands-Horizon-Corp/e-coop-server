package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

func FeedManager(service *horizon.HorizonService) *registry.Registry[
	types.Feed, types.FeedResponse, types.FeedRequest] {
	return registry.GetRegistry(registry.RegistryParams[types.Feed, types.FeedResponse, types.FeedRequest]{
		Preloads: []string{
			"CreatedBy", "CreatedBy.Media",
			"FeedMedias.Media", "FeedComments.User",
			"FeedComments.User.Media", "UserLikes.User", "UserLikes.User.Media"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Feed) *types.FeedResponse {
			if data == nil {
				return nil
			}
			if data.UserLikes == nil {
				data.UserLikes = []*types.FeedLike{}
			}
			if data.FeedComments == nil {
				data.FeedComments = []*types.FeedComment{}
			}
			return &types.FeedResponse{
				ID:               data.ID,
				CreatedAt:        data.CreatedAt.Format(time.RFC3339),
				CreatedByID:      data.CreatedByID,
				CreatedBy:        UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:        data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:      data.UpdatedByID,
				UpdatedBy:        UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:   data.OrganizationID,
				Organization:     OrganizationManager(service).ToModel(data.Organization),
				BranchID:         data.BranchID,
				Branch:           BranchManager(service).ToModel(data.Branch),
				Description:      data.Description,
				FeedMedias:       FeedMediaManager(service).ToModels(data.FeedMedias),
				FeedComments:     FeedCommentManager(service).ToModels(data.FeedComments),
				UserLikes:        FeedLikeManager(service).ToModels(data.UserLikes),
				IsLiked:          false,
				NumberOfLikes:    len(data.UserLikes),
				NumberOfComments: len(data.FeedComments),
			}
		},
		Created: func(data *types.Feed) registry.Topics {
			return []string{"feed.create", fmt.Sprintf("feed.create.branch.%s", data.BranchID)}
		},
		Updated: func(data *types.Feed) registry.Topics {
			return []string{"feed.update", fmt.Sprintf("feed.update.%s", data.ID)}
		},
		Deleted: func(data *types.Feed) registry.Topics {
			return []string{"feed.delete", fmt.Sprintf("feed.delete.%s", data.ID)}
		},
	})
}
