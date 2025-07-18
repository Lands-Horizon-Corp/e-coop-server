package controller

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) NotificationController() {
	req := c.provider.Service.Request

	// Get the current (logged in) user's notifications
	req.RegisterRoute(horizon.Route{
		Route:    "/notification/me",
		Method:   "GET",
		Response: "TNotification[]",
		Note:     "Returns all notifications for the currently logged-in user.",
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
		return ctx.JSON(http.StatusOK, c.model.NotificationManager.ToModels(notification))
	})

	// Mark multiple notifications as viewed
	req.RegisterRoute(horizon.Route{
		Route:   "/notification/view",
		Method:  "PUT",
		Request: "string[] ids",
		Note:    "Marks multiple notifications as viewed for the current user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			notificationID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s - %v", rawID, err)})
			}

			notification, err := c.model.NotificationManager.GetByID(context, notificationID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Notification with ID %s not found: %v", rawID, err)})
			}

			if notification.IsViewed {
				continue
			}

			notification.IsViewed = true
			if err := c.model.NotificationManager.UpdateFields(context, notification.ID, notification); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update notification: " + err.Error()})
			}
		}

		notifications, err := c.model.GetNotificationByUser(context, user.ID)
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get notifications: " + err.Error()})
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.NotificationManager.ToModels(notifications))
	})

	// Delete a notification by its ID
	req.RegisterRoute(horizon.Route{
		Route:  "/notification/:notification_id",
		Method: "DELETE",
		Note:   "Deletes a specific notification record by its notification_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		notificationId, err := horizon.EngineUUIDParam(ctx, "notification_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid notification_id: " + err.Error()})
		}
		if err := c.model.NotificationManager.DeleteByID(context, *notificationId); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete notification: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
