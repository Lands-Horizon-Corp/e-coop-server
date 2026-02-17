package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

func FeedCommentManager(service *horizon.HorizonService) *registry.Registry[
	types.FeedComment, types.FeedCommentResponse, types.FeedCommentRequest] {
	return registry.GetRegistry(registry.RegistryParams[types.FeedComment, types.FeedCommentResponse, types.FeedCommentRequest]{
		Preloads: []string{"CreatedBy", "User", "Media", "Organization", "Branch"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.FeedComment) *types.FeedCommentResponse {
			if data == nil {
				return nil
			}
			return &types.FeedCommentResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				BranchID:       data.BranchID,
				FeedID:         data.FeedID,
				UserID:         data.UserID,
				User:           UserManager(service).ToModel(data.User),
				Comment:        data.Comment,
				MediaID:        data.MediaID,
				Media:          MediaManager(service).ToModel(data.Media),
			}
		},
		Created: func(data *types.FeedComment) registry.Topics {
			return []string{"feed.comment.create", fmt.Sprintf("feed.%s.comment.create", data.FeedID)}
		},
		Deleted: func(data *types.FeedComment) registry.Topics {
			return []string{"feed.comment.delete", fmt.Sprintf("feed.%s.comment.delete", data.FeedID)}
		},
	})
}
