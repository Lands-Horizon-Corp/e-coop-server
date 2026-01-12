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

func chargesRateByTermController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/charges-rate-by-term/charges-rate-scheme/:charges_rate_scheme_id",
		Method:       "POST",
		Note:         "Creates a new charges rate by term for the current user's organization and branch.",
		RequestType:  core.ChargesRateByTermRequest{},
		ResponseType: core.ChargesRateByTermResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeID, err := helpers.EngineUUIDParam(ctx, "charges_rate_scheme_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate by term creation failed (/charges-rate-by-term), invalid charges rate scheme ID.",
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme ID"})
		}
		req, err := core.ChargesRateByTermManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate by term creation failed (/charges-rate-by-term), validation error: " + err.Error(),
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate by term data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate by term creation failed (/charges-rate-by-term), user org error: " + err.Error(),
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate by term creation failed (/charges-rate-by-term), user not assigned to branch.",
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		chargesRateByTerm := &core.ChargesRateByTerm{
			ChargesRateSchemeID: *chargesRateSchemeID,
			Name:                req.Name,
			Description:         req.Description,
			ModeOfPayment:       req.ModeOfPayment,
			Rate1:               req.Rate1,
			Rate2:               req.Rate2,
			Rate3:               req.Rate3,
			Rate4:               req.Rate4,
			Rate5:               req.Rate5,
			Rate6:               req.Rate6,
			Rate7:               req.Rate7,
			Rate8:               req.Rate8,
			Rate9:               req.Rate9,
			Rate10:              req.Rate10,
			Rate11:              req.Rate11,
			Rate12:              req.Rate12,
			Rate13:              req.Rate13,
			Rate14:              req.Rate14,
			Rate15:              req.Rate15,
			Rate16:              req.Rate16,
			Rate17:              req.Rate17,
			Rate18:              req.Rate18,
			Rate19:              req.Rate19,
			Rate20:              req.Rate20,
			Rate21:              req.Rate21,
			Rate22:              req.Rate22,
			CreatedAt:           time.Now().UTC(),
			CreatedByID:         userOrg.UserID,
			UpdatedAt:           time.Now().UTC(),
			UpdatedByID:         userOrg.UserID,
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
		}

		if err := core.ChargesRateByTermManager(service).Create(context, chargesRateByTerm); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate by term creation failed (/charges-rate-by-term), db error: " + err.Error(),
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate by term: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created charges rate by term (/charges-rate-by-term): " + chargesRateByTerm.ID.String(),
			Module:      "ChargesRateByTerm",
		})
		return ctx.JSON(http.StatusCreated, core.ChargesRateByTermManager(service).ToModel(chargesRateByTerm))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/charges-rate-by-term/:charges_rate_by_term_id",
		Method:       "PUT",
		Note:         "Updates an existing charges rate by term by its ID.",
		RequestType:  core.ChargesRateByTermRequest{},
		ResponseType: core.ChargesRateByTermResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateByTermID, err := helpers.EngineUUIDParam(ctx, "charges_rate_by_term_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate by term update failed (/charges-rate-by-term/:charges_rate_by_term_id), invalid charges rate by term ID.",
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate by term ID"})
		}

		req, err := core.ChargesRateByTermManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate by term update failed (/charges-rate-by-term/:charges_rate_by_term_id), validation error: " + err.Error(),
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate by term data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate by term update failed (/charges-rate-by-term/:charges_rate_by_term_id), user org error: " + err.Error(),
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		chargesRateByTerm, err := core.ChargesRateByTermManager(service).GetByID(context, *chargesRateByTermID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate by term update failed (/charges-rate-by-term/:charges_rate_by_term_id), charges rate by term not found.",
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate by term not found"})
		}
		chargesRateByTerm.Name = req.Name
		chargesRateByTerm.Description = req.Description
		chargesRateByTerm.ModeOfPayment = req.ModeOfPayment
		chargesRateByTerm.Rate1 = req.Rate1
		chargesRateByTerm.Rate2 = req.Rate2
		chargesRateByTerm.Rate3 = req.Rate3
		chargesRateByTerm.Rate4 = req.Rate4
		chargesRateByTerm.Rate5 = req.Rate5
		chargesRateByTerm.Rate6 = req.Rate6
		chargesRateByTerm.Rate7 = req.Rate7
		chargesRateByTerm.Rate8 = req.Rate8
		chargesRateByTerm.Rate9 = req.Rate9
		chargesRateByTerm.Rate10 = req.Rate10
		chargesRateByTerm.Rate11 = req.Rate11
		chargesRateByTerm.Rate12 = req.Rate12
		chargesRateByTerm.Rate13 = req.Rate13
		chargesRateByTerm.Rate14 = req.Rate14
		chargesRateByTerm.Rate15 = req.Rate15
		chargesRateByTerm.Rate16 = req.Rate16
		chargesRateByTerm.Rate17 = req.Rate17
		chargesRateByTerm.Rate18 = req.Rate18
		chargesRateByTerm.Rate19 = req.Rate19
		chargesRateByTerm.Rate20 = req.Rate20
		chargesRateByTerm.Rate21 = req.Rate21
		chargesRateByTerm.Rate22 = req.Rate22
		chargesRateByTerm.UpdatedAt = time.Now().UTC()
		chargesRateByTerm.UpdatedByID = userOrg.UserID
		if err := core.ChargesRateByTermManager(service).UpdateByID(context, chargesRateByTerm.ID, chargesRateByTerm); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate by term update failed (/charges-rate-by-term/:charges_rate_by_term_id), db error: " + err.Error(),
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate by term: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated charges rate by term (/charges-rate-by-term/:charges_rate_by_term_id): " + chargesRateByTerm.ID.String(),
			Module:      "ChargesRateByTerm",
		})
		return ctx.JSON(http.StatusOK, core.ChargesRateByTermManager(service).ToModel(chargesRateByTerm))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/charges-rate-by-term/:charges_rate_by_term_id",
		Method: "DELETE",
		Note:   "Deletes the specified charges rate by term by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateByTermID, err := helpers.EngineUUIDParam(ctx, "charges_rate_by_term_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate by term delete failed (/charges-rate-by-term/:charges_rate_by_term_id), invalid charges rate by term ID.",
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate by term ID"})
		}
		chargesRateByTerm, err := core.ChargesRateByTermManager(service).GetByID(context, *chargesRateByTermID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate by term delete failed (/charges-rate-by-term/:charges_rate_by_term_id), not found.",
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate by term not found"})
		}
		if err := core.ChargesRateByTermManager(service).Delete(context, *chargesRateByTermID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate by term delete failed (/charges-rate-by-term/:charges_rate_by_term_id), db error: " + err.Error(),
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate by term: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted charges rate by term (/charges-rate-by-term/:charges_rate_by_term_id): " + chargesRateByTerm.ID.String(),
			Module:      "ChargesRateByTerm",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/charges-rate-by-term/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple charges rate by term by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-by-term/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-by-term/bulk-delete) | no IDs provided",
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No charges rate by term IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.ChargesRateByTermManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-by-term/bulk-delete) | error: " + err.Error(),
				Module:      "ChargesRateByTerm",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete charges rate by term: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted charges rate by term (/charges-rate-by-term/bulk-delete)",
			Module:      "ChargesRateByTerm",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
