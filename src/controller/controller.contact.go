package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

// ContactController manages endpoints for contact records.
func (c *Controller) ContactController() {
	req := c.provider.Service.Request

	// GET /contact: List all contact records. (NO footstep)
	req.RegisterRoute(horizon.Route{
		Route:        "/contact",
		Method:       "GET",
		Note:         "Returns all contact records in the system.",
		ResponseType: model.ContactUsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contacts, err := c.model.ContactUsManager.List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve contact records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.ContactUsManager.Filtered(context, ctx, contacts))
	})

	// GET /contact/:contact_id: Get a specific contact by ID. (NO footstep)
	req.RegisterRoute(horizon.Route{
		Route:        "/contact/:contact_id",
		Method:       "GET",
		Note:         "Returns a single contact record by its ID.",
		ResponseType: model.ContactUsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contactID, err := horizon.EngineUUIDParam(ctx, "contact_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contact ID"})
		}
		contact, err := c.model.ContactUsManager.GetByIDRaw(context, *contactID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Contact record not found"})
		}
		return ctx.JSON(http.StatusOK, contact)
	})

	// POST /contact: Create a new contact record. (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:        "/contact",
		Method:       "POST",
		ResponseType: model.ContactUsResponse{},
		RequestType:  model.ContactUsRequest{},
		Note:         "Creates a new contact record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.ContactUsManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Contact creation failed (/contact), validation error: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contact data: " + err.Error()})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Contact creation failed (/contact), db error: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create contact record: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created contact (/contact): " + contact.Email,
			Module:      "Contact",
		})

		return ctx.JSON(http.StatusCreated, c.model.ContactUsManager.ToModel(contact))
	})

	// DELETE /contact/:contact_id: Delete a contact record by ID. (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:  "/contact/:contact_id",
		Method: "DELETE",
		Note:   "Deletes the specified contact record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contactID, err := horizon.EngineUUIDParam(ctx, "contact_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Contact delete failed (/contact/:contact_id), invalid contact ID.",
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contact ID"})
		}
		contact, err := c.model.ContactUsManager.GetByID(context, *contactID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Contact delete failed (/contact/:contact_id), record not found.",
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Contact record not found"})
		}
		if err := c.model.ContactUsManager.DeleteByID(context, *contactID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Contact delete failed (/contact/:contact_id), db error: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete contact record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted contact (/contact/:contact_id): " + contact.Email,
			Module:      "Contact",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /contact/bulk-delete: Bulk delete contact records by IDs. (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:       "/contact/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple contact records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Contact bulk delete failed (/contact/bulk-delete), invalid request body.",
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Contact bulk delete failed (/contact/bulk-delete), no IDs provided.",
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Contact bulk delete failed (/contact/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		emails := ""
		for _, rawID := range reqBody.IDs {
			contactID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Contact bulk delete failed (/contact/bulk-delete), invalid UUID: " + rawID,
					Module:      "Contact",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			contact, err := c.model.ContactUsManager.GetByID(context, contactID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Contact bulk delete failed (/contact/bulk-delete), not found: " + rawID,
					Module:      "Contact",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Contact record not found with ID: %s", rawID)})
			}
			emails += contact.Email + ","
			if err := c.model.ContactUsManager.DeleteByIDWithTx(context, tx, contactID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Contact bulk delete failed (/contact/bulk-delete), db error: " + err.Error(),
					Module:      "Contact",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete contact record: " + err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Contact bulk delete failed (/contact/bulk-delete), commit error: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted contacts (/contact/bulk-delete): " + emails,
			Module:      "Contact",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
