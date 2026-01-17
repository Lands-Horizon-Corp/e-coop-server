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

func NotificationManager(service *horizon.HorizonService) *registry.Registry[types.Notification, types.NotificationResponse, any] {
	return registry.NewRegistry(registry.RegistryParams[types.Notification, types.NotificationResponse, any]{
		Preloads: []string{"Recipient", "Recipient.Media"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Notification) *types.NotificationResponse {
			if data == nil {
				return nil
			}
			return &types.NotificationResponse{
				ID:               data.ID,
				UserID:           data.UserID,
				User:             UserManager(service).ToModel(data.User),
				RecipientID:      data.RecipientID,
				Recipient:        UserManager(service).ToModel(data.Recipient),
				Title:            data.Title,
				Description:      data.Description,
				IsViewed:         data.IsViewed,
				NotificationType: data.NotificationType,
				CreatedAt:        data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:        data.UpdatedAt.Format(time.RFC3339),
				UserType:         data.UserType,
			}
		},
		Created: func(data *types.Notification) registry.Topics {
			return []string{
				"notification.create",
				fmt.Sprintf("notification.create.user.%s", data.UserID),
				fmt.Sprintf("notification.create.%s", data.ID),
			}
		},
		Updated: func(data *types.Notification) registry.Topics {
			return []string{
				"notification.update",
				fmt.Sprintf("notification.update.user.%s", data.UserID),
				fmt.Sprintf("notification.update.%s", data.ID),
			}
		},
		Deleted: func(data *types.Notification) registry.Topics {
			return []string{
				"notification.delete",
				fmt.Sprintf("notification.delete.user.%s", data.UserID),
				fmt.Sprintf("notification.delete.%s", data.ID),
			}
		},
	})
}

func GetNotificationByUser(context context.Context, service *horizon.HorizonService, userID uuid.UUID) ([]*types.Notification, error) {
	return NotificationManager(service).Find(context, &types.Notification{
		UserID: userID,
	})
}
