package account

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func BrowseExcludeIncludeAccountsController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/browse-exclude-include-accounts/computation-sheet/:computation_sheet_id/search",
		Method:       "GET",
		Note:         "Returns all browse exclude include accounts for a computation sheet in the current user's org/branch.",
		ResponseType: types.BrowseExcludeIncludeAccountsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheetID, err := helpers.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		records, err := core.BrowseExcludeIncludeAccountsManager(service).Find(context, &types.BrowseExcludeIncludeAccounts{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No browse exclude include accounts found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, core.BrowseExcludeIncludeAccountsManager(service).ToModels(records))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/browse-exclude-include-accounts/computation-sheet/:computation_sheet_id",
		Method:       "GET",
		ResponseType: types.BrowseExcludeIncludeAccountsResponse{},
		Note:         "Returns all browse exclude include accounts for a computation sheet in the current user's org/branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheetID, err := helpers.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		records, err := core.BrowseExcludeIncludeAccountsManager(service).Find(context, &types.BrowseExcludeIncludeAccounts{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No browse exclude include accounts found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, core.BrowseExcludeIncludeAccountsManager(service).ToModels(records))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/browse-exclude-include-accounts",
		Method:       "POST",
		RequestType: types.BrowseExcludeIncludeAccountsRequest{},
		ResponseType: types.BrowseExcludeIncludeAccountsResponse{},
		Note:         "Creates a new browse exclude include account for the current user's org/branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.BrowseExcludeIncludeAccountsManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Browse exclude include account creation failed (/browse-exclude-include-accounts), validation error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Browse exclude include account creation failed (/browse-exclude-include-accounts), user org error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Browse exclude include account creation failed (/browse-exclude-include-accounts), user not assigned to branch.",
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		record := &types.BrowseExcludeIncludeAccounts{
			ComputationSheetID:           req.ComputationSheetID,
			FinesAccountID:               req.FinesAccountID,
			ComakerAccountID:             req.ComakerAccountID,
			InterestAccountID:            req.InterestAccountID,
			DeliquentAccountID:           req.DeliquentAccountID,
			IncludeExistingLoanAccountID: req.IncludeExistingLoanAccountID,
			CreatedAt:                    time.Now().UTC(),
			CreatedByID:                  userOrg.UserID,
			UpdatedAt:                    time.Now().UTC(),
			UpdatedByID:                  userOrg.UserID,
			BranchID:                     *userOrg.BranchID,
			OrganizationID:               userOrg.OrganizationID,
		}

		if err := core.BrowseExcludeIncludeAccountsManager(service).Create(context, record); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Browse exclude include account creation failed (/browse-exclude-include-accounts), db error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create browse exclude include account: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created browse exclude include account (/browse-exclude-include-accounts)",
			Module:      "BrowseExcludeIncludeAccounts",
		})
		return ctx.JSON(http.StatusCreated, core.BrowseExcludeIncludeAccountsManager(service).ToModel(record))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/browse-exclude-include-accounts/:browse_exclude_include_accounts_id",
		Method:       "PUT",
		Note:         "Updates an existing browse exclude include account by its ID.",
		ResponseType: types.BrowseExcludeIncludeAccountsResponse{},
		RequestType: types.BrowseExcludeIncludeAccountsRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "browse_exclude_include_accounts_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Browse exclude include account update failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), invalid ID.",
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid browse exclude include account ID"})
		}

		req, err := core.BrowseExcludeIncludeAccountsManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Browse exclude include account update failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), validation error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Browse exclude include account update failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), user org error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		record, err := core.BrowseExcludeIncludeAccountsManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		record.UpdatedByID = userOrg.UserID

		if err := core.BrowseExcludeIncludeAccountsManager(service).UpdateByID(context, record.ID, record); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Browse exclude include account update failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), db error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update browse exclude include account: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated browse exclude include account (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id)",
			Module:      "BrowseExcludeIncludeAccounts",
		})
		return ctx.JSON(http.StatusOK, core.BrowseExcludeIncludeAccountsManager(service).ToModel(record))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/browse-exclude-include-accounts/:browse_exclude_include_accounts_id",
		Method: "DELETE",
		Note:   "Deletes the specified browse exclude include account by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "browse_exclude_include_accounts_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Browse exclude include account delete failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), invalid ID.",
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid browse exclude include account ID"})
		}
		record, err := core.BrowseExcludeIncludeAccountsManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Browse exclude include account delete failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), not found.",
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Browse exclude include account not found"})
		}
		if err := core.BrowseExcludeIncludeAccountsManager(service).Delete(context, record.ID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Browse exclude include account delete failed (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id), db error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete browse exclude include account: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted browse exclude include account (/browse-exclude-include-accounts/:browse_exclude_include_accounts_id)",
			Module:      "BrowseExcludeIncludeAccounts",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/browse-exclude-include-accounts/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple browse exclude include accounts by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete browse exclude include accounts (/browse-exclude-include-accounts/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete browse exclude include accounts (/browse-exclude-include-accounts/bulk-delete) | no IDs provided",
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.BrowseExcludeIncludeAccountsManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete browse exclude include accounts (/browse-exclude-include-accounts/bulk-delete) | error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete browse exclude include accounts: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted browse exclude include accounts (/browse-exclude-include-accounts/bulk-delete)",
			Module:      "BrowseExcludeIncludeAccounts",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
