package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

func FeedbackManager(service *horizon.HorizonService) *registry.Registry[types.Feedback, types.FeedbackResponse, types.FeedbackRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.Feedback, types.FeedbackResponse, types.FeedbackRequest]{
		Preloads: []string{"Media"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Feedback) *types.FeedbackResponse {
			if data == nil {
				return nil
			}
			return &types.FeedbackResponse{
				ID:           data.ID,
				Email:        data.Email,
				Description:  data.Description,
				FeedbackType: data.FeedbackType,
				MediaID:      data.MediaID,
				Media:        MediaManager(service).ToModel(data.Media),
				CreatedAt:    data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:    data.UpdatedAt.Format(time.RFC3339),
			}
		},
		Created: func(data *types.Feedback) registry.Topics {
			return []string{
				"feedback.create",
				fmt.Sprintf("feedback.create.%s", data.ID),
			}
		},
		Updated: func(data *types.Feedback) registry.Topics {
			return []string{
				"feedback.update",
				fmt.Sprintf("feedback.update.%s", data.ID),
			}
		},
		Deleted: func(data *types.Feedback) registry.Topics {
			return []string{
				"feedback.delete",
				fmt.Sprintf("feedback.delete.%s", data.ID),
			}
		},
	})
}
