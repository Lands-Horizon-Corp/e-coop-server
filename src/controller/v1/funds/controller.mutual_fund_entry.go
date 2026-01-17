package funds

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func MutualFundEntryController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund-entry",
		Method:       "GET",
		Note:         "Returns all mutual fund entries for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := core.MutualFundEntryCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No mutual fund entries found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.MutualFundEntryManager(service).ToModels(entries))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund-entry/search",
		Method:       "GET",
		Note:         "Returns a paginated list of mutual fund entries for the current user's organization and branch.",
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := core.MutualFundEntryManager(service).NormalPagination(context, ctx, &types.MutualFundEntry{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch mutual fund entries for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund-entry/member/:member_id",
		Method:       "GET",
		Note:         "Returns all mutual fund entries for a specific member profile.",
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberID, err := helpers.EngineUUIDParam(ctx, "member_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := core.MutualFundEntryByMember(context, service, *memberID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No mutual fund entries found for the specified member"})
		}
		return ctx.JSON(http.StatusOK, core.MutualFundEntryManager(service).ToModels(entries))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund-entry/account/:account_id",
		Method:       "GET",
		Note:         "Returns all mutual fund entries for a specific account.",
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := core.MutualFundEntryByAccount(context, service, *accountID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No mutual fund entries found for the specified account"})
		}
		return ctx.JSON(http.StatusOK, core.MutualFundEntryManager(service).ToModels(entries))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund-entry/:entry_id",
		Method:       "GET",
		Note:         "Returns a single mutual fund entry by its ID.",
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		entryID, err := helpers.EngineUUIDParam(ctx, "entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund entry ID"})
		}
		entry, err := core.MutualFundEntryManager(service).GetByIDRaw(context, *entryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund entry not found"})
		}
		return ctx.JSON(http.StatusOK, entry)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund-entry/mutual-fund/:mutual_fund_id",
		Method:       "POST",
		Note:         "Creates a new mutual fund entry for the current user's organization and branch.",
		RequestType:  core.MutualFundEntryRequest{},
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		mutualFundID, err := helpers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}
		req, err := core.MutualFundEntryManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund entry creation failed, validation error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund entry data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "User organization error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		mutualFund, err := core.MutualFundManager(service).GetByID(context, mutualFundID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund not found"})
		}
		if helpers.UUIDPtrEqual(&mutualFund.MemberProfileID, &req.MemberProfileID) {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Member profile already associated with this mutual fund"})
		}

		entry := &types.MutualFundEntry{
			MemberProfileID: req.MemberProfileID,
			AccountID:       req.AccountID,
			Amount:          req.Amount,
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     userOrg.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     userOrg.UserID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
			MutualFundID:    *mutualFundID,
		}

		if err := core.MutualFundEntryManager(service).Create(context, entry); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "DB error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create mutual fund entry: " + err.Error(),
			})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: fmt.Sprintf("Created mutual fund entry: Amount %.2f for member %s", entry.Amount, entry.MemberProfileID),
			Module:      "MutualFundEntry",
		})

		return ctx.JSON(http.StatusCreated, core.MutualFundEntryManager(service).ToModel(entry))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund-entry/:entry_id",
		Method:       "PUT",
		Note:         "Updates an existing mutual fund entry by its ID.",
		RequestType:  core.MutualFundEntryRequest{},
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		entryID, err := helpers.EngineUUIDParam(ctx, "entry_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund entry update failed (/mutual-fund-entry/:entry_id), invalid entry ID.",
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund entry ID"})
		}

		req, err := core.MutualFundEntryManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund entry update failed (/mutual-fund-entry/:entry_id), validation error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund entry data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund entry update failed (/mutual-fund-entry/:entry_id), user org error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		entry, err := core.MutualFundEntryManager(service).GetByID(context, *entryID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund entry update failed (/mutual-fund-entry/:entry_id), entry not found.",
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund entry not found"})
		}
		entry.MemberProfileID = req.MemberProfileID
		entry.AccountID = req.AccountID
		entry.Amount = req.Amount
		entry.UpdatedAt = time.Now().UTC()
		entry.UpdatedByID = userOrg.UserID
		if err := core.MutualFundEntryManager(service).UpdateByID(context, entry.ID, entry); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund entry update failed (/mutual-fund-entry/:entry_id), db error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update mutual fund entry: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Updated mutual fund entry (/mutual-fund-entry/:entry_id): Amount %.2f for member %s", entry.Amount, entry.MemberProfileID),
			Module:      "MutualFundEntry",
		})
		return ctx.JSON(http.StatusOK, core.MutualFundEntryManager(service).ToModel(entry))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/mutual-fund-entry/:entry_id",
		Method: "DELETE",
		Note:   "Deletes the specified mutual fund entry by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		entryID, err := helpers.EngineUUIDParam(ctx, "entry_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Mutual fund entry delete failed (/mutual-fund-entry/:entry_id), invalid entry ID.",
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund entry ID"})
		}
		entry, err := core.MutualFundEntryManager(service).GetByID(context, *entryID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Mutual fund entry delete failed (/mutual-fund-entry/:entry_id), not found.",
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund entry not found"})
		}
		if err := core.MutualFundEntryManager(service).Delete(context, *entryID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Mutual fund entry delete failed (/mutual-fund-entry/:entry_id), db error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete mutual fund entry: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: fmt.Sprintf("Deleted mutual fund entry (/mutual-fund-entry/:entry_id): Amount %.2f for member %s", entry.Amount, entry.MemberProfileID),
			Module:      "MutualFundEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/mutual-fund-entry/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple mutual fund entries by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete mutual fund entries (/mutual-fund-entry/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete mutual fund entries (/mutual-fund-entry/bulk-delete) | no IDs provided",
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No mutual fund entry IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.MutualFundEntryManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete mutual fund entries (/mutual-fund-entry/bulk-delete) | error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete mutual fund entries: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted mutual fund entries (/mutual-fund-entry/bulk-delete)",
			Module:      "MutualFundEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
