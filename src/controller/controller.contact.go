package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) ContactController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/contact",
		Method:   "GET",
		Response: "TContact[]",
	}, func(ctx echo.Context) error {
		contact, err := c.model.ContactUsManager.ListRaw(context.Background())
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
		contactID, err := horizon.EngineUUIDParam(ctx, "contact_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid contact ID")
		}

		contact, err := c.model.ContactUsManager.GetByIDRaw(context.Background(), *contactID)
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

		if err := c.model.ContactUsManager.Create(context.Background(), contact); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusCreated, c.model.ContactUsManager.ToModel(contact))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/contact/:contact_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		contactID, err := horizon.EngineUUIDParam(ctx, "contact_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid contact ID")
		}

		if err := c.model.ContactUsManager.DeleteByID(context.Background(), *contactID); err != nil {
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
		context := context.Background()
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
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

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
