package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

func (h *Handler) FeedbackList(c echo.Context) error {
	feedback, err := h.repository.FeedbackList()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.FeedbackModels(feedback))
}

func (h *Handler) FeedbackGet(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid feedback ID"})
	}
	feedback, err := h.repository.FeedbackGetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.FeedbackModel(feedback))
}

func (h *Handler) FeedbackCreate(c echo.Context) error {
	req, err := h.model.FeedbackValidate(c)
	if err != nil {
		return err
	}
	model := &model.Feedback{
		Email:        req.Email,
		Description:  req.Description,
		FeedbackType: req.FeedbackType,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if err := h.repository.FeedbackCreate(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, h.model.FeedbackModel(model))
}

func (h *Handler) FeedbackUpdate(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid feedback ID"})
	}
	req, err := h.model.FeedbackValidate(c)
	if err != nil {
		return err
	}
	model := &model.Feedback{
		ID:           id,
		Email:        req.Email,
		Description:  req.Description,
		FeedbackType: req.FeedbackType,
		UpdatedAt:    time.Now().UTC(),
	}
	if err := h.repository.FeedbackUpdate(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.FeedbackModel(model))

}

func (h *Handler) FeedbackDelete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid feedback ID"})
	}
	model := &model.Feedback{ID: id}
	if err := h.repository.FeedbackDelete(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
