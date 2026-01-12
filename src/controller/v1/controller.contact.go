package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func contactController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/contact",
		Method:       "GET",
		Note:         "Returns all contact records in the system.",
		ResponseType: core.ContactUsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contacts, err := c.core.ContactUsManager().List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve contact records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.ContactUsManager().ToModels(contacts))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/contact/:contact_id",
		Method:       "GET",
		Note:         "Returns a single contact record by its ID.",
		ResponseType: core.ContactUsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contactID, err := handlers.EngineUUIDParam(ctx, "contact_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contact ID"})
		}
		contact, err := c.core.ContactUsManager().GetByIDRaw(context, *contactID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Contact record not found"})
		}
		return ctx.JSON(http.StatusOK, contact)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/contact",
		Method:       "POST",
		ResponseType: core.ContactUsResponse{},
		RequestType:  core.ContactUsRequest{},
		Note:         "Creates a new contact record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.ContactUsManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Contact creation failed (/contact), validation error: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contact data: " + err.Error()})
		}

		contact := &core.ContactUs{
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			Email:         req.Email,
			ContactNumber: req.ContactNumber,
			Description:   req.Description,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}

		if err := c.core.ContactUsManager().Create(context, contact); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Contact creation failed (/contact), db error: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create contact record: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created contact (/contact): " + contact.Email,
			Module:      "Contact",
		})

		return ctx.JSON(http.StatusCreated, c.core.ContactUsManager().ToModel(contact))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/contact/:contact_id",
		Method: "DELETE",
		Note:   "Deletes the specified contact record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contactID, err := handlers.EngineUUIDParam(ctx, "contact_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Contact delete failed (/contact/:contact_id), invalid contact ID.",
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contact ID"})
		}
		contact, err := c.core.ContactUsManager().GetByID(context, *contactID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Contact delete failed (/contact/:contact_id), record not found.",
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Contact record not found"})
		}
		if err := c.core.ContactUsManager().Delete(context, *contactID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Contact delete failed (/contact/:contact_id), db error: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete contact record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted contact (/contact/:contact_id): " + contact.Email,
			Module:      "Contact",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/contact/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple contact records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Contact bulk delete failed (/contact/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Contact bulk delete failed (/contact/bulk-delete) | no IDs provided",
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := c.core.ContactUsManager().BulkDelete(context, ids); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Contact bulk delete failed (/contact/bulk-delete) | error: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete contact records: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted contacts (/contact/bulk-delete)",
			Module:      "Contact",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
