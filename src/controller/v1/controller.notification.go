package v1

import (
	"fmt"
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func notificationController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/notification/me",
		Method:       "GET",
		ResponseType: core.NotificationResponse{},
		Note:         "Returns all notifications for the currently logged-in user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}
		notification, err := core.GetNotificationByUser(context, user.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get notifications: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.NotificationManager(service).ToModels(notification))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/notification/view",
		Method:       "PUT",
		RequestType:  core.IDSRequest{},
		ResponseType: core.NotificationResponse{},
		Note:         "Marks multiple notifications as viewed for the current user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View notifications failed: invalid request body: " + err.Error(),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View notifications failed: user error: " + err.Error(),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}

		var notifications []*core.Notification
		err = service.Database.StartTransactionWithContext(context, func(tx *gorm.DB) error {
			for _, notificationID := range reqBody.IDs {
				notification, getErr := core.NotificationManager(service).GetByID(context, notificationID)
				if getErr != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "update-error",
						Description: fmt.Sprintf("View notifications failed: notification not found: %s", notificationID.String()),
						Module:      "Notification",
					})
					return eris.Errorf("notification with ID %s not found: %v", notificationID.String(), getErr)
				}

				if notification.IsViewed {
					continue
				}

				notification.IsViewed = true
				if updateErr := core.NotificationManager(service).UpdateByID(context, notification.ID, notification); updateErr != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "update-error",
						Description: "View notifications failed: update error: " + updateErr.Error(),
						Module:      "Notification",
					})
					return eris.Errorf("failed to update notification: %v", updateErr)
				}
			}

			var getUserErr error
			notifications, getUserErr = core.GetNotificationByUser(context, user.ID)
			if getUserErr != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "update-error",
					Description: "View notifications failed: get notifications error: " + getUserErr.Error(),
					Module:      "Notification",
				})
				return eris.Errorf("failed to get notifications: %v", getUserErr)
			}

			return nil
		})

		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Marked notifications as viewed for user ID: %s", user.ID),
			Module:      "Notification",
		})

		return ctx.JSON(http.StatusOK, core.NotificationManager(service).ToModels(notifications))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/notification/view-all",
		Method:       "PUT",
		ResponseType: core.NotificationResponse{},
		Note:         "Marks all user notifications as viewed",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to view notifications: unable to get current user - " + err.Error(),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Unauthorized: Unable to get current user.",
				"error":   err.Error(),
			})
		}

		notifications, err := core.NotificationManager(service).Find(context, &core.Notification{
			UserID: user.ID,
		})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to view notifications: unable to retrieve user notifications - " + err.Error(),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Unable to retrieve notifications.",
				"error":   err.Error(),
			})
		}

		viewedCount := 0
		var newNotifications []*core.Notification
		err = service.Database.StartTransactionWithContext(context, func(tx *gorm.DB) error {
			for _, notif := range notifications {
				notification, getErr := core.NotificationManager(service).GetByID(context, notif.ID)
				if getErr != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "update-error",
						Description: fmt.Sprintf("Failed to mark notification %s as viewed: not found - %v", notif.ID, getErr),
						Module:      "Notification",
					})
					return eris.Errorf("notification with ID %s not found: %v", notif.ID, getErr)
				}

				if notification.IsViewed {
					continue
				}

				notification.IsViewed = true
				if updateErr := core.NotificationManager(service).UpdateByID(context, notification.ID, notification); updateErr != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "update-error",
						Description: fmt.Sprintf("Failed to update notification %s: %v", notif.ID, updateErr),
						Module:      "Notification",
					})
					return eris.Errorf("failed to update notification %s: %v", notif.ID, updateErr)
				}

				viewedCount++
			}

			var findErr error
			newNotifications, findErr = core.NotificationManager(service).Find(context, &core.Notification{
				UserID: user.ID,
			})
			if findErr != nil {
				return eris.Errorf("failed to get the new notification updates: %v", findErr)
			}

			return nil
		})

		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Failed to save notification updates.",
				"error":   err.Error(),
			})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("User %s marked %d notifications as viewed", user.ID, viewedCount),
			Module:      "Notification",
		})
		return ctx.JSON(http.StatusOK, core.NotificationManager(service).ToModels(newNotifications))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/notification/:notification_id",
		Method: "DELETE",
		Note:   "Deletes a specific notification record by its notificationit_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		notificationID, err := helpers.EngineUUIDParam(ctx, "notification_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete notification failed: invalid notification_id: " + err.Error(),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid notification_id: " + err.Error()})
		}
		notification, err := core.NotificationManager(service).GetByID(context, *notificationID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: fmt.Sprintf("Delete notification failed: not found (ID: %s): %v", notification.ID, err),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Notification with ID %s not found: %v", notification.ID, err)})
		}
		if err := core.NotificationManager(service).Delete(context, notification.ID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete notification failed: " + err.Error(),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete notification: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: fmt.Sprintf("Deleted notification ID: %s", notificationID),
			Module:      "Notification",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
