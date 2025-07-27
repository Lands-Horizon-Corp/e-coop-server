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

// BrowseExcludeIncludeAccountsController registers routes for managing browse exclude include accounts.
func (c *Controller) BrowseExcludeIncludeAccountsController() {
	req := c.provider.Service.Request

	// GET /browse-exclude-include-accounts/computation-sheet/:computation_sheet_id/search
	req.RegisterRoute(horizon.Route{
		Route:        "/browse-exclude-include-accounts/computation-sheet/:computation_sheet_id/search",
		Method:       "GET",
		Note:         "Returns all browse exclude include accounts for a computation sheet in the current user's org/branch.",
		ResponseType: model.BrowseExcludeIncludeAccountsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheetID, err := horizon.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		records, err := c.model.BrowseExcludeIncludeAccountsManager.Find(context, &model.BrowseExcludeIncludeAccounts{
			OrganizationID:     user.OrganizationID,
			BranchID:           *user.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No browse exclude include accounts found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, c.model.BrowseExcludeIncludeAccountsManager.Filtered(context, ctx, records))
	})

	// GET /browse-exclude-include-accounts/computation-sheet/:computation_sheet_id/search
	req.RegisterRoute(horizon.Route{
		Route:        "/browse-exclude-include-accounts/computation-sheet/:computation_sheet_id",
		Method:       "GET",
		ResponseType: model.BrowseExcludeIncludeAccountsResponse{},
		Note:         "Returns all browse exclude include accounts for a computation sheet in the current user's org/branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheetID, err := horizon.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		records, err := c.model.BrowseExcludeIncludeAccountsManager.Find(context, &model.BrowseExcludeIncludeAccounts{
			OrganizationID:     user.OrganizationID,
			BranchID:           *user.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No browse exclude include accounts found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, c.model.BrowseExcludeIncludeAccountsManager.Filtered(context, ctx, records))
	})

	// POST /browse-exclude-include-accounts
	req.RegisterRoute(horizon.Route{
		Route:        "/browse-exclude-include-accounts",
		Method:       "POST",
		RequestType:  model.BrowseExcludeIncludeAccountsRequest{},
		ResponseType: model.BrowseExcludeIncludeAccountsResponse{},
		Note:         "Creates a new browse exclude include account for the current user's org/branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.BrowseExcludeIncludeAccountsManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Browse exclude include account creation failed (/browse-exclude-include-accounts), validation error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Browse exclude include account creation failed (/browse-exclude-include-accounts), user org error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Browse exclude include account creation failed (/browse-exclude-include-accounts), user not assigned to branch.",
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		record := &model.BrowseExcludeIncludeAccounts{
			ComputationSheetID:           req.ComputationSheetID,
			FinesAccountID:               req.FinesAccountID,
			ComakerAccountID:             req.ComakerAccountID,
			InterestAccountID:            req.InterestAccountID,
			DeliquentAccountID:           req.DeliquentAccountID,
			IncludeExistingLoanAccountID: req.IncludeExistingLoanAccountID,
			CreatedAt:                    time.Now().UTC(),
			CreatedByID:                  user.UserID,
			UpdatedAt:                    time.Now().UTC(),
			UpdatedByID:                  user.UserID,
			BranchID:                     *user.BranchID,
			OrganizationID:               user.OrganizationID,
		}

		if err := c.model.BrowseExcludeIncludeAccountsManager.Create(context, record); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Browse exclude include account creation failed (/browse-exclude-include-accounts), db error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create browse exclude include account: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created browse exclude include account (/browse-exclude-include-accounts)",
			Module:      "BrowseExcludeIncludeAccounts",
		})
		return ctx.JSON(http.StatusCreated, c.model.BrowseExcludeIncludeAccountsManager.ToModel(record))
	})

	// PUT /browse-exclude-include-accounts/:browse_exclude_include_accounts_id
	req.RegisterRoute(horizon.Route{
		Route:        "/browse-exclude-include-accounts/:browse_exclude_include_accounts_id",
		Method:       "PUT",
		Request:      "BrowseExcludeIncludeAccounts",
		Response:     "BrowseExcludeIncludeAccounts",
		Note:         "Updates an existing browse exclude include account by its ID.",
		ResponseType: model.BrowseExcludeIncludeAccountsResponse{},
		RequestType:  model.BrowseExcludeIncludeAccountsRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "browse_exclude_include_accounts_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Browse exclude include account update failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), invalid ID.",
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid browse exclude include account ID"})
		}

		req, err := c.model.BrowseExcludeIncludeAccountsManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Browse exclude include account update failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), validation error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Browse exclude include account update failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), user org error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		record, err := c.model.BrowseExcludeIncludeAccountsManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Browse exclude include account update failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), not found.",
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Browse exclude include account not found"})
		}
		record.ComputationSheetID = req.ComputationSheetID
		record.FinesAccountID = req.FinesAccountID
		record.ComakerAccountID = req.ComakerAccountID
		record.InterestAccountID = req.InterestAccountID
		record.DeliquentAccountID = req.DeliquentAccountID
		record.IncludeExistingLoanAccountID = req.IncludeExistingLoanAccountID
		record.UpdatedAt = time.Now().UTC()
		record.UpdatedByID = user.UserID

		if err := c.model.BrowseExcludeIncludeAccountsManager.UpdateFields(context, record.ID, record); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Browse exclude include account update failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), db error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update browse exclude include account: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated browse exclude include account (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id)",
			Module:      "BrowseExcludeIncludeAccounts",
		})
		return ctx.JSON(http.StatusOK, c.model.BrowseExcludeIncludeAccountsManager.ToModel(record))
	})

	// DELETE /browse-exclude-include-accounts/:browse_exclude_include_accounts_id
	req.RegisterRoute(horizon.Route{
		Route:  "/browse-exclude-include-accounts/:browse_exclude_include_accounts_id",
		Method: "DELETE",
		Note:   "Deletes the specified browse exclude include account by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "browse_exclude_include_accounts_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Browse exclude include account delete failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), invalid ID.",
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid browse exclude include account ID"})
		}
		record, err := c.model.BrowseExcludeIncludeAccountsManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Browse exclude include account delete failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), not found.",
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Browse exclude include account not found"})
		}
		if err := c.model.BrowseExcludeIncludeAccountsManager.DeleteByID(context, record.ID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Browse exclude include account delete failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), db error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete browse exclude include account: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted browse exclude include account (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id)",
			Module:      "BrowseExcludeIncludeAccounts",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /browse-exclude-include-accounts/bulk-delete
	req.RegisterRoute(horizon.Route{
		Route:        "/browse-exclude-include-accounts/bulk-delete",
		Method:       "DELETE",
		Request:      "string[]",
		Note:         "Deletes multiple browse exclude include accounts by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		ResponseType: model.BulkDeleteRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/browse-exclude-include-accounts/bulk-delete), invalid request body.",
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/browse-exclude-include-accounts/bulk-delete), no IDs provided.",
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/browse-exclude-include-accounts/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/browse-exclude-include-accounts/bulk-delete), invalid UUID: " + rawID,
					Module:      "BrowseExcludeIncludeAccounts",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			record, err := c.model.BrowseExcludeIncludeAccountsManager.GetByID(context, id)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/browse-exclude-include-accounts/bulk-delete), not found: " + rawID,
					Module:      "BrowseExcludeIncludeAccounts",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Browse exclude include account not found with ID: %s", rawID)})
			}
			if err := c.model.BrowseExcludeIncludeAccountsManager.DeleteByIDWithTx(context, tx, record.ID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/browse-exclude-include-accounts/bulk-delete), db error: " + err.Error(),
					Module:      "BrowseExcludeIncludeAccounts",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete browse exclude include account: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/browse-exclude-include-accounts/bulk-delete), commit error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted browse exclude include accounts (/browse-exclude-include-accounts/bulk-delete)",
			Module:      "BrowseExcludeIncludeAccounts",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
