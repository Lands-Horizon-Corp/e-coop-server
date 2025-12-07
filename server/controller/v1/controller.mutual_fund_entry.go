package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// MutualFundEntryController registers routes for managing mutual fund entries.
func (c *Controller) mutualFundEntryController() {
	req := c.provider.Service.Request

	// GET /mutual-fund-entry: List all mutual fund entries for the current user's branch. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund-entry",
		Method:       "GET",
		Note:         "Returns all mutual fund entries for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := c.core.MutualFundEntryCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No mutual fund entries found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.MutualFundEntryManager.ToModels(entries))
	})

	// GET /mutual-fund-entry/search: Paginated search of mutual fund entries for the current branch. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund-entry/search",
		Method:       "GET",
		Note:         "Returns a paginated list of mutual fund entries for the current user's organization and branch.",
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := c.core.MutualFundEntryManager.PaginationWithFields(context, ctx, &core.MutualFundEntry{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch mutual fund entries for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	// GET /mutual-fund-entry/member/:member_id: Get mutual fund entries by member profile ID. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund-entry/member/:member_id",
		Method:       "GET",
		Note:         "Returns all mutual fund entries for a specific member profile.",
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberID, err := handlers.EngineUUIDParam(ctx, "member_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := c.core.MutualFundEntryByMember(context, *memberID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No mutual fund entries found for the specified member"})
		}
		return ctx.JSON(http.StatusOK, c.core.MutualFundEntryManager.ToModels(entries))
	})

	// GET /mutual-fund-entry/account/:account_id: Get mutual fund entries by account ID. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund-entry/account/:account_id",
		Method:       "GET",
		Note:         "Returns all mutual fund entries for a specific account.",
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := c.core.MutualFundEntryByAccount(context, *accountID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No mutual fund entries found for the specified account"})
		}
		return ctx.JSON(http.StatusOK, c.core.MutualFundEntryManager.ToModels(entries))
	})

	// GET /mutual-fund-entry/:entry_id: Get specific mutual fund entry by ID. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund-entry/:entry_id",
		Method:       "GET",
		Note:         "Returns a single mutual fund entry by its ID.",
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		entryID, err := handlers.EngineUUIDParam(ctx, "entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund entry ID"})
		}
		entry, err := c.core.MutualFundEntryManager.GetByIDRaw(context, *entryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund entry not found"})
		}
		return ctx.JSON(http.StatusOK, entry)
	})

	// POST /mutual-fund-entry: Create a new mutual fund entry. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund-entry",
		Method:       "POST",
		Note:         "Creates a new mutual fund entry for the current user's organization and branch.",
		RequestType:  core.MutualFundEntryRequest{},
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.MutualFundEntryManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund entry creation failed (/mutual-fund-entry), validation error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund entry data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund entry creation failed (/mutual-fund-entry), user org error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund entry creation failed (/mutual-fund-entry), user not assigned to branch.",
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		entry := &core.MutualFundEntry{
			MemberProfileID: req.MemberProfileID,
			AccountID:       req.AccountID,
			Amount:          req.Amount,
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     userOrg.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     userOrg.UserID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
		}

		if err := c.core.MutualFundEntryManager.Create(context, entry); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund entry creation failed (/mutual-fund-entry), db error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create mutual fund entry: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: fmt.Sprintf("Created mutual fund entry (/mutual-fund-entry): Amount %.2f for member %s", entry.Amount, entry.MemberProfileID),
			Module:      "MutualFundEntry",
		})
		return ctx.JSON(http.StatusCreated, c.core.MutualFundEntryManager.ToModel(entry))
	})

	// PUT /mutual-fund-entry/:entry_id: Update mutual fund entry by ID. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund-entry/:entry_id",
		Method:       "PUT",
		Note:         "Updates an existing mutual fund entry by its ID.",
		RequestType:  core.MutualFundEntryRequest{},
		ResponseType: core.MutualFundEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		entryID, err := handlers.EngineUUIDParam(ctx, "entry_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund entry update failed (/mutual-fund-entry/:entry_id), invalid entry ID.",
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund entry ID"})
		}

		req, err := c.core.MutualFundEntryManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund entry update failed (/mutual-fund-entry/:entry_id), validation error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund entry data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund entry update failed (/mutual-fund-entry/:entry_id), user org error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		entry, err := c.core.MutualFundEntryManager.GetByID(context, *entryID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		if err := c.core.MutualFundEntryManager.UpdateByID(context, entry.ID, entry); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund entry update failed (/mutual-fund-entry/:entry_id), db error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update mutual fund entry: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Updated mutual fund entry (/mutual-fund-entry/:entry_id): Amount %.2f for member %s", entry.Amount, entry.MemberProfileID),
			Module:      "MutualFundEntry",
		})
		return ctx.JSON(http.StatusOK, c.core.MutualFundEntryManager.ToModel(entry))
	})

	// DELETE /mutual-fund-entry/:entry_id: Delete a mutual fund entry by ID. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/mutual-fund-entry/:entry_id",
		Method: "DELETE",
		Note:   "Deletes the specified mutual fund entry by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		entryID, err := handlers.EngineUUIDParam(ctx, "entry_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Mutual fund entry delete failed (/mutual-fund-entry/:entry_id), invalid entry ID.",
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund entry ID"})
		}
		entry, err := c.core.MutualFundEntryManager.GetByID(context, *entryID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Mutual fund entry delete failed (/mutual-fund-entry/:entry_id), not found.",
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund entry not found"})
		}
		if err := c.core.MutualFundEntryManager.Delete(context, *entryID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Mutual fund entry delete failed (/mutual-fund-entry/:entry_id), db error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete mutual fund entry: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: fmt.Sprintf("Deleted mutual fund entry (/mutual-fund-entry/:entry_id): Amount %.2f for member %s", entry.Amount, entry.MemberProfileID),
			Module:      "MutualFundEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /mutual-fund-entry/bulk-delete: Bulk delete multiple mutual fund entries. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/mutual-fund-entry/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple mutual fund entries by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete mutual fund entries (/mutual-fund-entry/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete mutual fund entries (/mutual-fund-entry/bulk-delete) | no IDs provided",
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No mutual fund entry IDs provided for bulk delete"})
		}

		if err := c.core.MutualFundEntryManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete mutual fund entries (/mutual-fund-entry/bulk-delete) | error: " + err.Error(),
				Module:      "MutualFundEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete mutual fund entries: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted mutual fund entries (/mutual-fund-entry/bulk-delete)",
			Module:      "MutualFundEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
