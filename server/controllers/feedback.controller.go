package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /feedback
func (c *Controller) FeedbackList(ctx echo.Context) error {
	feedback, err := c.feedback.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.FeedbackModels(feedback))
}

// GET /feedback/:feedback_id
func (c *Controller) FeedbackGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "feedback_id")
	if err != nil {
		return err
	}
	feedback, err := c.feedback.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.FeedbackModel(feedback))
}

// POST /feedback
func (c *Controller) FeedbackCreate(ctx echo.Context) error {
	req, err := c.model.FeedbackValidate(ctx)
	if err != nil {
		return err
	}
	model := &model.Feedback{
		Email:        req.Email,
		Description:  req.Description,
		FeedbackType: req.FeedbackType,
		MediaID:      req.MediaID,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if err := c.feedback.Manager.Create(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.FeedbackModel(model))
}

// PUT /feedback/feedback_id
func (c *Controller) FeedbackUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "feedback_id")
	if err != nil {
		return err
	}
	req, err := c.model.FeedbackValidate(ctx)
	if err != nil {
		return err
	}
	model := &model.Feedback{
		Email:        req.Email,
		Description:  req.Description,
		FeedbackType: req.FeedbackType,
		MediaID:      req.MediaID,
		UpdatedAt:    time.Now().UTC(),
	}
	if err := c.feedback.Manager.UpdateByID(*id, model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.FeedbackModel(model))
}

// DELETE /feedback/feedback_id
func (c *Controller) FeedbackDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "feedback_id")
	if err != nil {
		return err
	}
	if err := c.feedback.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}
