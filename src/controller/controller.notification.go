package controller

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) NotificationController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/notification/me",
		Method:   "GET",
		Response: "TNotification[]",
		Note:     "Getting your own notifications (the logged in user)",
	}, func(ctx echo.Context) error {
		context := context.Background()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		notification, err := c.model.GetNotificationByUser(context, user.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.NotificationManager.ToModels(notification))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/notification/notification_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := context.Background()
		notificationId, err := horizon.EngineUUIDParam(ctx, "notification_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid notification ID")
		}
		if err := c.model.NotificationManager.DeleteByID(context, *notificationId); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})

}
