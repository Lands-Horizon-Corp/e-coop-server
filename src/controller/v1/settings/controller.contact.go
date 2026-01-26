package settings

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func ContactController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/contact",
		Method:       "GET",
		Note:         "Returns all contact records in the system.",
		ResponseType: types.ContactUsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contacts, err := core.ContactUsManager(service).List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve contact records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.ContactUsManager(service).ToModels(contacts))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/contact/:contact_id",
		Method:       "GET",
		Note:         "Returns a single contact record by its ID.",
		ResponseType: types.ContactUsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contactID, err := helpers.EngineUUIDParam(ctx, "contact_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contact ID"})
		}
		contact, err := core.ContactUsManager(service).GetByIDRaw(context, *contactID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Contact record not found"})
		}
		return ctx.JSON(http.StatusOK, contact)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/contact",
		Method:       "POST",
		ResponseType: types.ContactUsResponse{},
		RequestType:  types.ContactUsRequest{},
		Note:         "Creates a new contact record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.ContactUsManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Contact creation failed (/contact), validation error: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contact data: " + err.Error()})
		}

		contact := &types.ContactUs{
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			Email:         req.Email,
			ContactNumber: req.ContactNumber,
			Description:   req.Description,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}

		if err := core.ContactUsManager(service).Create(context, contact); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Contact creation failed (/contact), db error: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create contact record: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created contact (/contact): " + contact.Email,
			Module:      "Contact",
		})

		return ctx.JSON(http.StatusCreated, core.ContactUsManager(service).ToModel(contact))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/contact/:contact_id",
		Method: "DELETE",
		Note:   "Deletes the specified contact record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		contactID, err := helpers.EngineUUIDParam(ctx, "contact_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Contact delete failed (/contact/:contact_id), invalid contact ID.",
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contact ID"})
		}
		contact, err := core.ContactUsManager(service).GetByID(context, *contactID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Contact delete failed (/contact/:contact_id), record not found.",
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Contact record not found"})
		}
		if err := core.ContactUsManager(service).Delete(context, *contactID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Contact delete failed (/contact/:contact_id), db error: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete contact record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted contact (/contact/:contact_id): " + contact.Email,
			Module:      "Contact",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/contact/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple contact records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Contact bulk delete failed (/contact/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.ContactUsManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Contact bulk delete failed (/contact/bulk-delete) | error: " + err.Error(),
				Module:      "Contact",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete contact records: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted contacts (/contact/bulk-delete)",
			Module:      "Contact",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
