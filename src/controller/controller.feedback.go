package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) FeedbackController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/feedback",
		Method:   "GET",
		Response: "TFeedback[]",
	}, func(ctx echo.Context) error {
		feedback, err := c.model.FeedbackManager.ListRaw(context.Background())
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, feedback)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/feedback/:feedback_id",
		Method:   "GET",
		Response: "TFeedback",
	}, func(ctx echo.Context) error {
		context := context.Background()
		feedbackId, err := horizon.EngineUUIDParam(ctx, "feedback_id")
		if err != nil {
			return err
		}
		feedback, err := c.model.FeedbackManager.GetByID(context, *feedbackId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, feedback)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/feedback",
		Method:   "POST",
		Request:  "TFeedback",
		Response: "TFeedback",
	}, func(ctx echo.Context) error {
		context := context.Background()
		req, err := c.model.FeedbackManager.Validate(ctx)
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
		if err := c.model.FeedbackManager.Create(context, model); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.model.FeedbackManager.ToModel(model))
	})

}
