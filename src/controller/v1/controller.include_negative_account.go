package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

// IncludeNegativeAccountController registers routes for managing include negative accounts.
func (c *Controller) IncludeNegativeAccountController() {
	req := c.provider.Service.Request

	// GET /include-negative-accounts/computation-sheet/:computation_sheet_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/include-negative-accounts/computation-sheet/:computation_sheet_id/search",
		Method:       "GET",
		ResponseType: model.IncludeNegativeAccountResponse{},
		Note:         "Returns all include negative accounts for a computation sheet in the current user's org/branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheetID, err := handlers.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		records, err := c.model.IncludeNegativeAccountManager.Find(context, &model.IncludeNegativeAccount{
			OrganizationID:     user.OrganizationID,
			BranchID:           *user.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No include negative accounts found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, c.model.IncludeNegativeAccountManager.Pagination(context, ctx, records))
	})

	// GET /include-negative-accounts/computation-sheet/:computation_sheet_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/include-negative-accounts/computation-sheet/:computation_sheet_id",
		Method:       "GET",
		ResponseType: model.IncludeNegativeAccountResponse{},
		Note:         "Returns all include negative accounts for a computation sheet in the current user's org/branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheetID, err := handlers.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		records, err := c.model.IncludeNegativeAccountManager.Find(context, &model.IncludeNegativeAccount{
			OrganizationID:     user.OrganizationID,
			BranchID:           *user.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No include negative accounts found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, c.model.IncludeNegativeAccountManager.Filtered(context, ctx, records))
	})

	// POST /include-negative-accounts
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/include-negative-accounts",
		Method:       "POST",
		ResponseType: model.IncludeNegativeAccountResponse{},
		RequestType:  model.IncludeNegativeAccountRequest{},
		Note:         "Creates a new include negative account for the current user's org/branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.IncludeNegativeAccountManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Include negative account creation failed (/include-negative-accounts), validation error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Include negative account creation failed (/include-negative-accounts), user org error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Include negative account creation failed (/include-negative-accounts), user not assigned to branch.",
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		record := &model.IncludeNegativeAccount{
			ComputationSheetID: req.ComputationSheetID,
			AccountID:          req.AccountID,
			Description:        req.Description,
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        user.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        user.UserID,
			BranchID:           *user.BranchID,
			OrganizationID:     user.OrganizationID,
		}

		if err := c.model.IncludeNegativeAccountManager.Create(context, record); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Include negative account creation failed (/include-negative-accounts), db error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create include negative account: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created include negative account (/include-negative-accounts)",
			Module:      "IncludeNegativeAccount",
		})
		return ctx.JSON(http.StatusCreated, c.model.IncludeNegativeAccountManager.ToModel(record))
	})

	// PUT /include-negative-accounts/:include_negative_accounts_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/include-negative-accounts/:include_negative_accounts_id",
		Method:       "PUT",
		ResponseType: model.IncludeNegativeAccountResponse{},
		RequestType:  model.IncludeNegativeAccountRequest{},
		Note:         "Updates an existing include negative account by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "include_negative_accounts_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Include negative account update failed (/include-negative-accounts/:include_negative_accounts_id), invalid ID.",
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid include negative account ID"})
		}

		req, err := c.model.IncludeNegativeAccountManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Include negative account update failed (/include-negative-accounts/:include_negative_accounts_id), validation error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Include negative account update failed (/include-negative-accounts/:include_negative_accounts_id), user org error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		record, err := c.model.IncludeNegativeAccountManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Include negative account update failed (/include-negative-accounts/:include_negative_accounts_id), not found.",
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Include negative account not found"})
		}
		record.ComputationSheetID = req.ComputationSheetID
		record.AccountID = req.AccountID
		record.Description = req.Description
		record.UpdatedAt = time.Now().UTC()
		record.UpdatedByID = user.UserID

		if err := c.model.IncludeNegativeAccountManager.UpdateFields(context, record.ID, record); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Include negative account update failed (/include-negative-accounts/:include_negative_accounts_id), db error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update include negative account: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated include negative account (/include-negative-accounts/:include_negative_accounts_id)",
			Module:      "IncludeNegativeAccount",
		})
		return ctx.JSON(http.StatusOK, c.model.IncludeNegativeAccountManager.ToModel(record))
	})

	// DELETE /include-negative-accounts/:include_negative_accounts_id
	req.RegisterRoute(handlers.Route{
		Route:  "/include-negative-accounts/:include_negative_accounts_id",
		Method: "DELETE",
		Note:   "Deletes the specified include negative account by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "include_negative_accounts_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Include negative account delete failed (/include-negative-accounts/:include_negative_accounts_id), invalid ID.",
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid include negative account ID"})
		}
		record, err := c.model.IncludeNegativeAccountManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Include negative account delete failed (/include-negative-accounts/:include_negative_accounts_id), not found.",
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Include negative account not found"})
		}
		if err := c.model.IncludeNegativeAccountManager.DeleteByID(context, record.ID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Include negative account delete failed (/include-negative-accounts/:include_negative_accounts_id), db error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete include negative account: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted include negative account (/include-negative-accounts/:include_negative_accounts_id)",
			Module:      "IncludeNegativeAccount",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /include-negative-accounts/bulk-delete
	req.RegisterRoute(handlers.Route{
		Route:       "/include-negative-accounts/bulk-delete",
		Method:      "DELETE",
		RequestType: model.IDSRequest{},
		Note:        "Deletes multiple include negative accounts by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/include-negative-accounts/bulk-delete), invalid request body.",
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/include-negative-accounts/bulk-delete), no IDs provided.",
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/include-negative-accounts/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/include-negative-accounts/bulk-delete), invalid UUID: " + rawID,
					Module:      "IncludeNegativeAccount",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			record, err := c.model.IncludeNegativeAccountManager.GetByID(context, id)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/include-negative-accounts/bulk-delete), not found: " + rawID,
					Module:      "IncludeNegativeAccount",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Include negative account not found with ID: %s", rawID)})
			}
			if err := c.model.IncludeNegativeAccountManager.DeleteByIDWithTx(context, tx, record.ID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/include-negative-accounts/bulk-delete), db error: " + err.Error(),
					Module:      "IncludeNegativeAccount",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete include negative account: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/include-negative-accounts/bulk-delete), commit error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted include negative accounts (/include-negative-accounts/bulk-delete)",
			Module:      "IncludeNegativeAccount",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
