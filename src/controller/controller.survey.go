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

func (c *Controller) FeedbackController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/feedback",
		Method:   "GET",
		Response: "TFeedback[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedback, err := c.model.FeedbackManager.ListRaw(context)
		if err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, feedback)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/feedback/:feedback_id",
		Method:   "GET",
		Response: "TFeedback",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedbackID, err := horizon.EngineUUIDParam(ctx, "feedback_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid feedback ID")
		}

		feedback, err := c.model.FeedbackManager.GetByIDRaw(context, *feedbackID)
		if err != nil {
			return c.NotFound(ctx, "Feedback")
		}

		return ctx.JSON(http.StatusOK, feedback)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/feedback",
		Method:   "POST",
		Request:  "TFeedback",
		Response: "TFeedback",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.FeedbackManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
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
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.FeedbackManager.ToModel(feedback))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/feedback/:feedback_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedbackID, err := horizon.EngineUUIDParam(ctx, "feedback_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid feedback ID")
		}

		if err := c.model.FeedbackManager.DeleteByID(context, *feedbackID); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/feedback/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple feedback records",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}

		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}

		if len(reqBody.IDs) == 0 {
			return c.BadRequest(ctx, "No IDs provided")
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			feedbackID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.FeedbackManager.GetByID(context, feedbackID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("Feedback with ID %s", rawID))
			}

			if err := c.model.FeedbackManager.DeleteByIDWithTx(context, tx, feedbackID); err != nil {
				tx.Rollback()
				return c.InternalServerError(ctx, err)
			}
		}

		if err := tx.Commit().Error; err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) ContactController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/contact",
		Method:   "GET",
		Response: "TContact[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contact, err := c.model.ContactUsManager.ListRaw(context)
		if err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, contact)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/contact/:contact_id",
		Method:   "GET",
		Response: "TContact",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contactID, err := horizon.EngineUUIDParam(ctx, "contact_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid contact ID")
		}

		contact, err := c.model.ContactUsManager.GetByIDRaw(context, *contactID)
		if err != nil {
			return c.NotFound(ctx, "Contact")
		}

		return ctx.JSON(http.StatusOK, contact)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/contact",
		Method:   "POST",
		Request:  "TContact",
		Response: "TContact",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.ContactUsManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		contact := &model.ContactUs{
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			Email:         req.Email,
			ContactNumber: req.ContactNumber,
			Description:   req.Description,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}

		if err := c.model.ContactUsManager.Create(context, contact); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusCreated, c.model.ContactUsManager.ToModel(contact))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/contact/:contact_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contactID, err := horizon.EngineUUIDParam(ctx, "contact_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid contact ID")
		}

		if err := c.model.ContactUsManager.DeleteByID(context, *contactID); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/contact/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple contact records",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}

		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}

		if len(reqBody.IDs) == 0 {
			return c.BadRequest(ctx, "No IDs provided")
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			contactID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.ContactUsManager.GetByID(context, contactID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("Contact with ID %s", rawID))
			}

			if err := c.model.ContactUsManager.DeleteByIDWithTx(context, tx, contactID); err != nil {
				tx.Rollback()
				return c.InternalServerError(ctx, err)
			}
		}

		if err := tx.Commit().Error; err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
