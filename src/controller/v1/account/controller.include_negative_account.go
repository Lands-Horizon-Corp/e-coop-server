package account

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func IncludeNegativeAccountController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/include-negative-accounts/computation-sheet/:computation_sheet_id/search",
		Method:       "GET",
		ResponseType: core.IncludeNegativeAccountResponse{},
		Note:         "Returns all include negative accounts for a computation sheet in the current user's org/branch.",
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
		records, err := core.IncludeNegativeAccountManager(service).NormalPagination(context, ctx, &core.IncludeNegativeAccount{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No include negative accounts found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, records)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/include-negative-accounts/computation-sheet/:computation_sheet_id",
		Method:       "GET",
		ResponseType: core.IncludeNegativeAccountResponse{},
		Note:         "Returns all include negative accounts for a computation sheet in the current user's org/branch.",
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
		records, err := core.IncludeNegativeAccountManager(service).Find(context, &core.IncludeNegativeAccount{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No include negative accounts found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, core.IncludeNegativeAccountManager(service).ToModels(records))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/include-negative-accounts",
		Method:       "POST",
		ResponseType: core.IncludeNegativeAccountResponse{},
		RequestType:  core.IncludeNegativeAccountRequest{},
		Note:         "Creates a new include negative account for the current user's org/branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.IncludeNegativeAccountManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Include negative account creation failed (/include-negative-accounts), validation error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Include negative account creation failed (/include-negative-accounts), user org error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Include negative account creation failed (/include-negative-accounts), user not assigned to branch.",
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		record := &core.IncludeNegativeAccount{
			ComputationSheetID: req.ComputationSheetID,
			AccountID:          req.AccountID,
			Description:        req.Description,
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			BranchID:           *userOrg.BranchID,
			OrganizationID:     userOrg.OrganizationID,
		}

		if err := core.IncludeNegativeAccountManager(service).Create(context, record); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Include negative account creation failed (/include-negative-accounts), db error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create include negative account: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created include negative account (/include-negative-accounts)",
			Module:      "IncludeNegativeAccount",
		})
		return ctx.JSON(http.StatusCreated, core.IncludeNegativeAccountManager(service).ToModel(record))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/include-negative-accounts/:include_negative_accounts_id",
		Method:       "PUT",
		ResponseType: core.IncludeNegativeAccountResponse{},
		RequestType:  core.IncludeNegativeAccountRequest{},
		Note:         "Updates an existing include negative account by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "include_negative_accounts_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Include negative account update failed (/include-negative-accounts/:include_negative_accounts_id), invalid ID.",
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid include negative account ID"})
		}

		req, err := core.IncludeNegativeAccountManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Include negative account update failed (/include-negative-accounts/:include_negative_accounts_id), validation error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Include negative account update failed (/include-negative-accounts/:include_negative_accounts_id), user org error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		record, err := core.IncludeNegativeAccountManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		record.UpdatedByID = userOrg.UserID

		if err := core.IncludeNegativeAccountManager(service).UpdateByID(context, record.ID, record); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Include negative account update failed (/include-negative-accounts/:include_negative_accounts_id), db error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update include negative account: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated include negative account (/include-negative-accounts/:include_negative_accounts_id)",
			Module:      "IncludeNegativeAccount",
		})
		return ctx.JSON(http.StatusOK, core.IncludeNegativeAccountManager(service).ToModel(record))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/include-negative-accounts/:include_negative_accounts_id",
		Method: "DELETE",
		Note:   "Deletes the specified include negative account by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "include_negative_accounts_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Include negative account delete failed (/include-negative-accounts/:include_negative_accounts_id), invalid ID.",
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid include negative account ID"})
		}
		record, err := core.IncludeNegativeAccountManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Include negative account delete failed (/include-negative-accounts/:include_negative_accounts_id), not found.",
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Include negative account not found"})
		}
		if err := core.IncludeNegativeAccountManager(service).Delete(context, record.ID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Include negative account delete failed (/include-negative-accounts/:include_negative_accounts_id), db error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete include negative account: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted include negative account (/include-negative-accounts/:include_negative_accounts_id)",
			Module:      "IncludeNegativeAccount",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/include-negative-accounts/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple include negative accounts by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/include-negative-accounts/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/include-negative-accounts/bulk-delete) | no IDs provided",
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.IncludeNegativeAccountManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/include-negative-accounts/bulk-delete) | error: " + err.Error(),
				Module:      "IncludeNegativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete include negative accounts: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted include negative accounts (/include-negative-accounts/bulk-delete)",
			Module:      "IncludeNegativeAccount",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
