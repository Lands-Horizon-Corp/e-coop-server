package controller_v1

import (
	"fmt"
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) NotificationController() {
	req := c.provider.Service.Request

	// Get the current (logged in) user's notifications
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/notification/me",
		Method:       "GET",
		ResponseType: model.NotificationResponse{},
		Note:         "Returns all notifications for the currently logged-in user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}
		notification, err := c.model.GetNotificationByUser(context, user.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get notifications: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.NotificationManager.Filtered(context, ctx, notification))
	})

	// Mark multiple notifications as viewed
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/notification/view",
		Method:       "PUT",
		RequestType:  model.IDSRequest{},
		ResponseType: model.NotificationResponse{},
		Note:         "Marks multiple notifications as viewed for the current user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View notifications failed: invalid request body: " + err.Error(),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View notifications failed: user error: " + err.Error(),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View notifications failed: begin tx error: " + tx.Error.Error(),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			notificationID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "View notifications failed: invalid UUID: " + rawID,
					Module:      "Notification",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s - %v", rawID, err)})
			}

			notification, err := c.model.NotificationManager.GetByID(context, notificationID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: fmt.Sprintf("View notifications failed: notification not found: %s", rawID),
					Module:      "Notification",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Notification with ID %s not found: %v", rawID, err)})
			}

			if notification.IsViewed {
				continue
			}

			notification.IsViewed = true
			if err := c.model.NotificationManager.UpdateFields(context, notification.ID, notification); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "View notifications failed: update error: " + err.Error(),
					Module:      "Notification",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update notification: " + err.Error()})
			}
		}

		notifications, err := c.model.GetNotificationByUser(context, user.ID)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View notifications failed: get notifications error: " + err.Error(),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get notifications: " + err.Error()})
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View notifications failed: commit tx error: " + err.Error(),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Marked notifications as viewed for user ID: %s", user.ID),
			Module:      "Notification",
		})

		return ctx.JSON(http.StatusOK, c.model.NotificationManager.Filtered(context, ctx, notifications))
	})

	// Delete a notification by its ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/notification/:notification_id",
		Method: "DELETE",
		Note:   "Deletes a specific notification record by its notification_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		notificationId, err := handlers.EngineUUIDParam(ctx, "notification_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete notification failed: invalid notification_id: " + err.Error(),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid notification_id: " + err.Error()})
		}
		notification, err := c.model.NotificationManager.GetByID(context, *notificationId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: fmt.Sprintf("Delete notification failed: not found (ID: %s): %v", notification.ID, err),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Notification with ID %s not found: %v", notification.ID, err)})
		}
		if err := c.model.NotificationManager.DeleteByID(context, notification.ID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete notification failed: " + err.Error(),
				Module:      "Notification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete notification: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: fmt.Sprintf("Deleted notification ID: %s", notificationId),
			Module:      "Notification",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
