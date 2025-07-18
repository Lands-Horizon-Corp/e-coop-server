package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

// FeedbackController manages endpoints for feedback records.
func (c *Controller) FeedbackController() {
	req := c.provider.Service.Request

	// GET /feedback: List all feedback records.
	req.RegisterRoute(horizon.Route{
		Route:    "/feedback",
		Method:   "GET",
		Response: "TFeedback[]",
		Note:     "Returns all feedback records in the system.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedback, err := c.model.FeedbackManager.ListRaw(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve feedback records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, feedback)
	})

	// GET /feedback/:feedback_id: Get a specific feedback by ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/feedback/:feedback_id",
		Method:   "GET",
		Response: "TFeedback",
		Note:     "Returns a single feedback record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedbackID, err := horizon.EngineUUIDParam(ctx, "feedback_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid feedback ID"})
		}

		feedback, err := c.model.FeedbackManager.GetByIDRaw(context, *feedbackID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Feedback record not found"})
		}

		return ctx.JSON(http.StatusOK, feedback)
	})

	// POST /feedback: Create a new feedback record.
	req.RegisterRoute(horizon.Route{
		Route:    "/feedback",
		Method:   "POST",
		Request:  "TFeedback",
		Response: "TFeedback",
		Note:     "Creates a new feedback record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.FeedbackManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid feedback data: " + err.Error()})
		}

		feedback := &model.Feedback{
			Email:        req.Email,
			Description:  req.Description,
			FeedbackType: req.FeedbackType,
			MediaID:      req.MediaID,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		if err := c.model.FeedbackManager.Create(context, feedback); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create feedback record: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, c.model.FeedbackManager.ToModel(feedback))
	})

	// DELETE /feedback/:feedback_id: Delete a feedback record by ID.
	req.RegisterRoute(horizon.Route{
		Route:  "/feedback/:feedback_id",
		Method: "DELETE",
		Note:   "Deletes the specified feedback record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedbackID, err := horizon.EngineUUIDParam(ctx, "feedback_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid feedback ID"})
		}

		if err := c.model.FeedbackManager.DeleteByID(context, *feedbackID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete feedback record: " + err.Error()})
		}

		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /feedback/bulk-delete: Bulk delete feedback records by IDs.
	req.RegisterRoute(horizon.Route{
		Route:   "/feedback/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple feedback records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}

		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		if len(reqBody.IDs) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			feedbackID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}

			if _, err := c.model.FeedbackManager.GetByID(context, feedbackID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Feedback record not found with ID: %s", rawID)})
			}

			if err := c.model.FeedbackManager.DeleteByIDWithTx(context, tx, feedbackID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete feedback record: " + err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
