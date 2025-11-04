package v1

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DisbursementController registers routes for managing disbursements.
func (c *Controller) disbursementController() {
	req := c.provider.Service.Request

	// GET /disbursement: List all disbursements for the current user's branch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement",
		Method:       "GET",
		Note:         "Returns all disbursements for the current user's organization and branch.",
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		transactionBatch, err := c.core.TransactionBatchCurrent(context, user.UserID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch current transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for the current branch"})
		}
		disbursements, err := c.core.DisbursementManager.Find(context, &core.Disbursement{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
			CurrencyID:     transactionBatch.CurrencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No disbursements found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.DisbursementManager.ToModels(disbursements))
	})

	// GET /disbursement/search: Paginated search of disbursements for the current branch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement/search",
		Method:       "GET",
		Note:         "Returns a paginated list of disbursements for the current user's organization and branch.",
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		disbursements, err := c.core.DisbursementCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch disbursements for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.DisbursementManager.Pagination(context, ctx, disbursements))
	})

	// GET /disbursement/:disbursement_id: Get specific disbursement by ID.
	req.RegisterRoute(handlers.Route{
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

	// POST /disbursement: Create a new disbursement.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement",
		Method:       "POST",
		Note:         "Creates a new disbursement for the current user's organization and branch.",
		RequestType:  core.DisbursementRequest{},
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.DisbursementManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Disbursement creation failed (/disbursement), validation error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid disbursement data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Disbursement creation failed (/disbursement), user org error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
			CurrencyID:     req.CurrencyID,
		}

		if err := c.core.DisbursementManager.Create(context, disbursement); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Disbursement creation failed (/disbursement), db error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create disbursement: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created disbursement (/disbursement): " + disbursement.Name,
			Module:      "Disbursement",
		})
		return ctx.JSON(http.StatusCreated, c.core.DisbursementManager.ToModel(disbursement))
	})

	// PUT /disbursement/:disbursement_id: Update disbursement by ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement/:disbursement_id",
		Method:       "PUT",
		Note:         "Updates an existing disbursement by its ID.",
		RequestType:  core.DisbursementRequest{},
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		disbursementID, err := handlers.EngineUUIDParam(ctx, "disbursement_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Disbursement update failed (/disbursement/:disbursement_id), invalid ID.",
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid disbursement ID"})
		}

		req, err := c.core.DisbursementManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Disbursement update failed (/disbursement/:disbursement_id), validation error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid disbursement data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Disbursement update failed (/disbursement/:disbursement_id), user org error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		disbursement, err := c.core.DisbursementManager.GetByID(context, *disbursementID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		disbursement.UpdatedByID = user.UserID
		disbursement.CurrencyID = req.CurrencyID
		if err := c.core.DisbursementManager.UpdateByID(context, disbursement.ID, disbursement); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Disbursement update failed (/disbursement/:disbursement_id), db error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update disbursement: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated disbursement (/disbursement/:disbursement_id): " + disbursement.Name,
			Module:      "Disbursement",
		})
		return ctx.JSON(http.StatusOK, c.core.DisbursementManager.ToModel(disbursement))
	})

	// DELETE /disbursement/:disbursement_id: Delete a disbursement by ID.
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/disbursement/:disbursement_id",
		Method: "DELETE",
		Note:   "Deletes the specified disbursement by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		disbursementID, err := handlers.EngineUUIDParam(ctx, "disbursement_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Disbursement delete failed (/disbursement/:disbursement_id), invalid ID.",
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid disbursement ID"})
		}
		disbursement, err := c.core.DisbursementManager.GetByID(context, *disbursementID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Disbursement delete failed (/disbursement/:disbursement_id), not found.",
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Disbursement not found"})
		}
		if err := c.core.DisbursementManager.Delete(context, *disbursementID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Disbursement delete failed (/disbursement/:disbursement_id), db error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete disbursement: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted disbursement (/disbursement/:disbursement_id): " + disbursement.Name,
			Module:      "Disbursement",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /disbursement/bulk-delete: Bulk delete disbursements by IDs.
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/disbursement/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple disbursements by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/disbursement/bulk-delete), invalid request body.",
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/disbursement/bulk-delete), no IDs provided.",
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No disbursement IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/disbursement/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		var namesSlice []string
		for _, rawID := range reqBody.IDs {
			disbursementID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/disbursement/bulk-delete), invalid UUID: " + rawID,
					Module:      "Disbursement",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			disbursement, err := c.core.DisbursementManager.GetByID(context, disbursementID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/disbursement/bulk-delete), not found: " + rawID,
					Module:      "Disbursement",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Disbursement not found with ID: %s", rawID)})
			}
			namesSlice = append(namesSlice, disbursement.Name)
			if err := c.core.DisbursementManager.DeleteWithTx(context, tx, disbursementID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/disbursement/bulk-delete), db error: " + err.Error(),
					Module:      "Disbursement",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete disbursement: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/disbursement/bulk-delete), commit error: " + err.Error(),
				Module:      "Disbursement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		names := strings.Join(namesSlice, ",")
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted disbursements (/disbursement/bulk-delete): " + names,
			Module:      "Disbursement",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
