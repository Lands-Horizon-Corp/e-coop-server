package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) disbursementController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/disbursement",
		Method:       "GET",
		Note:         "Returns all disbursements for the current user's organization and branch.",
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		transactionBatch, err := c.core.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch current transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for the current branch"})
		}
		disbursements, err := c.core.DisbursementManager.Find(context, &core.Disbursement{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CurrencyID:     transactionBatch.CurrencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No disbursements found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.DisbursementManager.ToModels(disbursements))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/disbursement/search",
		Method:       "GET",
		Note:         "Returns a paginated list of disbursements for the current user's organization and branch.",
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		disbursements, err := c.core.DisbursementManager.NormalPagination(context, ctx, &core.Disbursement{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch disbursements for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, disbursements)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/disbursement/:disbursement_id",
		Method:       "GET",
		Note:         "Returns a single disbursement by its ID.",
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		disbursementID, err := handlers.EngineUUIDParam(ctx, "disbursement_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid disbursement ID"})
		}
		disbursement, err := c.core.DisbursementManager.GetByIDRaw(context, *disbursementID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Disbursement not found"})
		}
		return ctx.JSON(http.StatusOK, disbursement)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/disbursement",
		Method:       "POST",
		Note:         "Creates a new disbursement for the current user's organization and branch.",
		RequestType:  core.DisbursementRequest{},
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.DisbursementManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Disbursement creation failed (/disbursement), validation error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid disbursement data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Disbursement creation failed (/disbursement), user org error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Disbursement creation failed (/disbursement), user not assigned to branch.",
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		disbursement := &core.Disbursement{
			Name:           req.Name,
			Icon:           req.Icon,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			CurrencyID:     req.CurrencyID,
		}

		if err := c.core.DisbursementManager.Create(context, disbursement); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Disbursement creation failed (/disbursement), db error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create disbursement: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created disbursement (/disbursement): " + disbursement.Name,
			Module:      "Disbursement",
		})
		return ctx.JSON(http.StatusCreated, c.core.DisbursementManager.ToModel(disbursement))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/disbursement/:disbursement_id",
		Method:       "PUT",
		Note:         "Updates an existing disbursement by its ID.",
		RequestType:  core.DisbursementRequest{},
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		disbursementID, err := handlers.EngineUUIDParam(ctx, "disbursement_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Disbursement update failed (/disbursement/:disbursement_id), invalid ID.",
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid disbursement ID"})
		}

		req, err := c.core.DisbursementManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Disbursement update failed (/disbursement/:disbursement_id), validation error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid disbursement data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Disbursement update failed (/disbursement/:disbursement_id), user org error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		disbursement, err := c.core.DisbursementManager.GetByID(context, *disbursementID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Disbursement update failed (/disbursement/:disbursement_id), not found.",
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Disbursement not found"})
		}
		disbursement.Name = req.Name
		disbursement.Icon = req.Icon
		disbursement.Description = req.Description
		disbursement.UpdatedAt = time.Now().UTC()
		disbursement.UpdatedByID = userOrg.UserID
		disbursement.CurrencyID = req.CurrencyID
		if err := c.core.DisbursementManager.UpdateByID(context, disbursement.ID, disbursement); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Disbursement update failed (/disbursement/:disbursement_id), db error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update disbursement: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated disbursement (/disbursement/:disbursement_id): " + disbursement.Name,
			Module:      "Disbursement",
		})
		return ctx.JSON(http.StatusOK, c.core.DisbursementManager.ToModel(disbursement))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/disbursement/:disbursement_id",
		Method: "DELETE",
		Note:   "Deletes the specified disbursement by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		disbursementID, err := handlers.EngineUUIDParam(ctx, "disbursement_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Disbursement delete failed (/disbursement/:disbursement_id), invalid ID.",
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid disbursement ID"})
		}
		disbursement, err := c.core.DisbursementManager.GetByID(context, *disbursementID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Disbursement delete failed (/disbursement/:disbursement_id), not found.",
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Disbursement not found"})
		}
		if err := c.core.DisbursementManager.Delete(context, *disbursementID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Disbursement delete failed (/disbursement/:disbursement_id), db error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete disbursement: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted disbursement (/disbursement/:disbursement_id): " + disbursement.Name,
			Module:      "Disbursement",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/disbursement/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple disbursements by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/disbursement/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/disbursement/bulk-delete) | no IDs provided",
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No disbursement IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := c.core.DisbursementManager.BulkDelete(context, ids); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/disbursement/bulk-delete) | error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete disbursements: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted disbursements (/disbursement/bulk-delete)",
			Module:      "Disbursement",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
