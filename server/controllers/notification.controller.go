package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

// GET /notification
func (c *Controller) NotificationList(ctx echo.Context) error {
	notification, err := c.notification.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.NotificationModels(notification))
}

// GET /notification/:notification_id
func (c *Controller) NotificationGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "notification_id")
	if err != nil {
		return err
	}
	notification, err := c.notification.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.NotificationModel(notification))
}

// DELETE /notification/:notification_id
func (c *Controller) NotificationDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "notification_id")
	if err != nil {
		return err
	}
	if err := c.notification.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET notification/user/:user_id
func (c *Controller) NotificationListByUser(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	notification, err := c.notification.ListByUser(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.NotificationModels(notification))
}

// GET notification/user/:user_id/unviewed-count
func (c *Controller) NotificationListByUserUnseenCount(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	count, err := c.notification.ListByUserUnviewedCount(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	return ctx.JSON(http.StatusOK, map[string]int64{
		"unviewed_count": count,
	})
}

// GET notification/user/:user_id/unviewed
func (c *Controller) NotificationListByUserUnviewed(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	notification, err := c.notification.ListByUserUnviewed(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.NotificationModels(notification))
}

// GET notification/user/:user_id/read-all
func (c *Controller) NotificationListByUserReadAll(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	notification, err := c.notification.ReadAll(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	return ctx.JSON(http.StatusOK, c.model.NotificationModels(notification))
}
