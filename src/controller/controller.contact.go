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

// ContactController manages endpoints for contact records.
func (c *Controller) ContactController() {
	req := c.provider.Service.Request

	// GET /contact: List all contact records.
	req.RegisterRoute(horizon.Route{
		Route:    "/contact",
		Method:   "GET",
		Response: "TContact[]",
		Note:     "Returns all contact records in the system.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contacts, err := c.model.ContactUsManager.ListRaw(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve contact records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, contacts)
	})

	// GET /contact/:contact_id: Get a specific contact by ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/contact/:contact_id",
		Method:   "GET",
		Response: "TContact",
		Note:     "Returns a single contact record by its ID.",
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

	// POST /contact: Create a new contact record.
	req.RegisterRoute(horizon.Route{
		Route:    "/contact",
		Method:   "POST",
		Request:  "TContact",
		Response: "TContact",
		Note:     "Creates a new contact record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.ContactUsManager.Validate(ctx)
		if err != nil {
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create contact record: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, c.model.ContactUsManager.ToModel(contact))
	})

	// DELETE /contact/:contact_id: Delete a contact record by ID.
	req.RegisterRoute(horizon.Route{
		Route:  "/contact/:contact_id",
		Method: "DELETE",
		Note:   "Deletes the specified contact record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contactID, err := horizon.EngineUUIDParam(ctx, "contact_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contact ID"})
		}
		if err := c.model.ContactUsManager.DeleteByID(context, *contactID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete contact record: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /contact/bulk-delete: Bulk delete contact records by IDs.
	req.RegisterRoute(horizon.Route{
		Route:   "/contact/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple contact records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
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
			contactID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			if _, err := c.model.ContactUsManager.GetByID(context, contactID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Contact record not found with ID: %s", rawID)})
			}
			if err := c.model.ContactUsManager.DeleteByIDWithTx(context, tx, contactID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete contact record: " + err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
