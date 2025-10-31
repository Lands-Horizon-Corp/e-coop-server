package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelCore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelCore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// FeedbackController manages endpoints for feedback records.
func (c *Controller) FeedbackController() {
	req := c.provider.Service.Request

	// GET /feedback: List all feedback records. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/feedback",
		Method:       "GET",
		Note:         "Returns all feedback records in the system.",
		ResponseType: modelCore.FeedbackResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedback, err := c.modelCore.FeedbackManager.List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve feedback records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.FeedbackManager.Filtered(context, ctx, feedback))
	})

	// GET /feedback/:feedback_id: Get a specific feedback by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/feedback/:feedback_id",
		Method:       "GET",
		Note:         "Returns a single feedback record by its ID.",
		ResponseType: modelCore.FeedbackResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedbackID, err := handlers.EngineUUIDParam(ctx, "feedback_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid feedback ID"})
		}

		feedback, err := c.modelCore.FeedbackManager.GetByIDRaw(context, *feedbackID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Feedback record not found"})
		}

		return ctx.JSON(http.StatusOK, feedback)
	})

	// POST /feedback: Create a new feedback record. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/feedback",
		Method:       "POST",
		Note:         "Creates a new feedback record.",
		ResponseType: modelCore.FeedbackResponse{},
		RequestType:  modelCore.FeedbackRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelCore.FeedbackManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Feedback creation failed (/feedback), validation error: " + err.Error(),
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid feedback data: " + err.Error()})
		}

		feedback := &modelCore.Feedback{
			Email:        req.Email,
			Description:  req.Description,
			FeedbackType: req.FeedbackType,
			MediaID:      req.MediaID,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		if err := c.modelCore.FeedbackManager.Create(context, feedback); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Feedback creation failed (/feedback), db error: " + err.Error(),
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create feedback record: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created feedback (/feedback): " + feedback.Email,
			Module:      "Feedback",
		})

		return ctx.JSON(http.StatusCreated, c.modelCore.FeedbackManager.ToModel(feedback))
	})

	// DELETE /feedback/:feedback_id: Delete a feedback record by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/feedback/:feedback_id",
		Method: "DELETE",
		Note:   "Deletes the specified feedback record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedbackID, err := handlers.EngineUUIDParam(ctx, "feedback_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Feedback delete failed (/feedback/:feedback_id), invalid feedback ID.",
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid feedback ID"})
		}

		feedback, err := c.modelCore.FeedbackManager.GetByID(context, *feedbackID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Feedback delete failed (/feedback/:feedback_id), record not found.",
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Feedback record not found"})
		}

		if err := c.modelCore.FeedbackManager.DeleteByID(context, *feedbackID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Feedback delete failed (/feedback/:feedback_id), db error: " + err.Error(),
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete feedback record: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted feedback (/feedback/:feedback_id): " + feedback.Email,
			Module:      "Feedback",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /feedback/bulk-delete: Bulk delete feedback records by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/feedback/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple feedback records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: modelCore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelCore.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Feedback bulk delete failed (/feedback/bulk-delete), invalid request body.",
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Feedback bulk delete failed (/feedback/bulk-delete), no IDs provided.",
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Feedback bulk delete failed (/feedback/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		emails := ""
		for _, rawID := range reqBody.IDs {
			feedbackID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Feedback bulk delete failed (/feedback/bulk-delete), invalid UUID: " + rawID,
					Module:      "Feedback",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}

			feedback, err := c.modelCore.FeedbackManager.GetByID(context, feedbackID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Feedback bulk delete failed (/feedback/bulk-delete), not found: " + rawID,
					Module:      "Feedback",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Feedback record not found with ID: %s", rawID)})
			}

			emails += feedback.Email + ","

			if err := c.modelCore.FeedbackManager.DeleteByIDWithTx(context, tx, feedbackID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Feedback bulk delete failed (/feedback/bulk-delete), db error: " + err.Error(),
					Module:      "Feedback",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete feedback record: " + err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Feedback bulk delete failed (/feedback/bulk-delete), commit error: " + err.Error(),
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted feedbacks (/feedback/bulk-delete): " + emails,
			Module:      "Feedback",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
