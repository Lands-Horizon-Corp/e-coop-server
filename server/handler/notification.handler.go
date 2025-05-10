package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

func (h *Handler) NotificationList(c echo.Context) error {
	notification, err := h.repository.NotificationList()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.NotificationModels(notification))
}

func (h *Handler) NotificationGet(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid notification ID"})
	}
	notification, err := h.repository.NotificationGetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.NotificationModel(notification))
}

func (h *Handler) NotificationView(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid notification ID"})
	}
	model := &model.Notification{
		ID:        id,
		IsViewed:  true,
		UpdatedAt: time.Now().UTC(),
	}
	if err := h.repository.NotificationUpdate(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.NotificationModel(model))
}

func (h *Handler) NotificationUnviewedCount(c echo.Context) error {
	count, err := h.repository.CountUnviewedNotifications()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]int64{"unviewed_count": count})
}

func (h *Handler) NotificationDelete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid notification ID"})
	}
	model := &model.Notification{ID: id}
	if err := h.repository.NotificationDelete(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
