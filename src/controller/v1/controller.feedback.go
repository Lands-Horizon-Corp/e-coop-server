package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func feedbackController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/feedback",
		Method:       "GET",
		Note:         "Returns all feedback records in the system.",
		ResponseType: core.FeedbackResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedback, err := core.FeedbackManager(service).List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve feedback records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.FeedbackManager(service).ToModels(feedback))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/feedback/:feedback_id",
		Method:       "GET",
		Note:         "Returns a single feedback record by its ID.",
		ResponseType: core.FeedbackResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedbackID, err := helpers.EngineUUIDParam(ctx, "feedback_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid feedback ID"})
		}

		feedback, err := core.FeedbackManager(service).GetByIDRaw(context, *feedbackID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Feedback record not found"})
		}

		return ctx.JSON(http.StatusOK, feedback)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/feedback",
		Method:       "POST",
		Note:         "Creates a new feedback record.",
		ResponseType: core.FeedbackResponse{},
		RequestType:  core.FeedbackRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.FeedbackManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Feedback creation failed (/feedback), validation error: " + err.Error(),
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid feedback data: " + err.Error()})
		}

		feedback := &core.Feedback{
			Email:        req.Email,
			Description:  req.Description,
			FeedbackType: req.FeedbackType,
			MediaID:      req.MediaID,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		if err := core.FeedbackManager(service).Create(context, feedback); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Feedback creation failed (/feedback), db error: " + err.Error(),
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create feedback record: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created feedback (/feedback): " + feedback.Email,
			Module:      "Feedback",
		})

		return ctx.JSON(http.StatusCreated, core.FeedbackManager(service).ToModel(feedback))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/feedback/:feedback_id",
		Method: "DELETE",
		Note:   "Deletes the specified feedback record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedbackID, err := helpers.EngineUUIDParam(ctx, "feedback_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Feedback delete failed (/feedback/:feedback_id), invalid feedback ID.",
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid feedback ID"})
		}

		feedback, err := core.FeedbackManager(service).GetByID(context, *feedbackID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Feedback delete failed (/feedback/:feedback_id), record not found.",
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Feedback record not found"})
		}

		if err := core.FeedbackManager(service).Delete(context, *feedbackID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Feedback delete failed (/feedback/:feedback_id), db error: " + err.Error(),
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete feedback record: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted feedback (/feedback/:feedback_id): " + feedback.Email,
			Module:      "Feedback",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/feedback/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple feedback records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Feedback bulk delete failed (/feedback/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Feedback bulk delete failed (/feedback/bulk-delete) | no IDs provided",
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.FeedbackManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Feedback bulk delete failed (/feedback/bulk-delete) | error: " + err.Error(),
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete feedback records: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted feedbacks (/feedback/bulk-delete)",
			Module:      "Feedback",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
