package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// GeneratedSavingsInterestEntryController registers routes for managing generated savings interest entries.
func (c *Controller) generatedSavingsInterestEntryController() {
	req := c.provider.Service.Request

	// GET /generated-savings-interest-entry: List all generated savings interest entries for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-savings-interest-entry",
		Method:       "GET",
		Note:         "Returns all generated savings interest entries for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.GeneratedSavingsInterestEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := c.core.GenerateSavingsInterestEntryCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated savings interest entries found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneratedSavingsInterestEntryManager.ToModels(entries))
	})

	// GET /generated-savings-interest-entry/search: Paginated search of generated savings interest entries for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-savings-interest-entry/search",
		Method:       "GET",
		Note:         "Returns a paginated list of generated savings interest entries for the current user's organization and branch.",
		ResponseType: core.GeneratedSavingsInterestEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := c.core.GeneratedSavingsInterestEntryManager.PaginationWithFields(context, ctx, &core.GeneratedSavingsInterestEntry{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch generated savings interest entries for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	// GET /generated-savings-interest-entry/:entry_id: Get specific generated savings interest entry by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-savings-interest-entry/:entry_id",
		Method:       "GET",
		Note:         "Returns a single generated savings interest entry by its ID.",
		ResponseType: core.GeneratedSavingsInterestEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		entryID, err := handlers.EngineUUIDParam(ctx, "entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid entry ID"})
		}
		entry, err := c.core.GeneratedSavingsInterestEntryManager.GetByIDRaw(context, *entryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated savings interest entry not found"})
		}
		return ctx.JSON(http.StatusOK, entry)
	})

	// POST /generated-savings-interest-entry: Create a new generated savings interest entry. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-savings-interest-entry/generated-savings-interest/:generated_savings_interest_id",
		Method:       "POST",
		Note:         "Creates a new generated savings interest entry for the current user's organization and branch.",
		RequestType:  core.GeneratedSavingsInterestEntryRequest{},
		ResponseType: core.GeneratedSavingsInterestEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedSavingsInterestID, err := handlers.EngineUUIDParam(ctx, "generated_savings_interest_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated savings interest entry creation failed (/generated-savings-interest-entry), invalid generated savings interest ID.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest ID"})
		}
		req, err := c.core.GeneratedSavingsInterestEntryManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated savings interest entry creation failed (/generated-savings-interest-entry), validation error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest entry data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated savings interest entry creation failed (/generated-savings-interest-entry), user org error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated savings interest entry creation failed (/generated-savings-interest-entry), user not assigned to branch.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		generatedSavingsInterest, err := c.core.GeneratedSavingsInterestManager.GetByID(
			context, *generatedSavingsInterestID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), parent generated savings interest not found.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Parent generated savings interest not found"})
		}
		if generatedSavingsInterest.PostedDate != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), parent generated savings interest is already posted.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot update entry because the parent generated savings interest is already posted"})
		}
		// Get daily ending balances for the computation period
		dailyBalances, err := c.core.GetDailyEndingBalances(
			context,
			generatedSavingsInterest.LastComputationDate,
			generatedSavingsInterest.NewComputationDate,
			req.AccountID,
			req.MemberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), failed to get daily balances: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get daily balances: " + err.Error()})
		}

		// Compute interest using the usecase
		var savingsType usecase.SavingsType
		switch generatedSavingsInterest.SavingsComputationType {
		case core.SavingsComputationTypeDailyLowestBalance:
			savingsType = usecase.SavingsTypeLowest
		case core.SavingsComputationTypeAverageDailyBalance:
			savingsType = usecase.SavingsTypeAverage
		case core.SavingsComputationTypeMonthlyEndLowestBalance:
			savingsType = usecase.SavingsTypeLowest
		case core.SavingsComputationTypeADBEndBalance:
			savingsType = usecase.SavingsTypeEnd
		case core.SavingsComputationTypeMonthlyLowestBalanceAverage:
			savingsType = usecase.SavingsTypeLowest
		case core.SavingsComputationTypeMonthlyEndBalanceAverage:
			savingsType = usecase.SavingsTypeAverage
		case core.SavingsComputationTypeMonthlyEndBalanceTotal:
			savingsType = usecase.SavingsTypeEnd
		default:
			savingsType = usecase.SavingsTypeLowest
		}

		result := c.usecase.GetSavingsEndingBalance(usecase.SavingsBalanceComputation{
			DailyBalance:   dailyBalances,
			SavingsType:    savingsType,
			InterestAmount: req.InterestAmount,
			InterestTax:    req.InterestTax,
		})

		entry := &core.GeneratedSavingsInterestEntry{
			GeneratedSavingsInterestID: *generatedSavingsInterestID,
			AccountID:                  req.AccountID,
			MemberProfileID:            req.MemberProfileID,
			EndingBalance:              result.Balance,
			InterestAmount:             result.InterestAmount,
			InterestTax:                result.InterestTax,
			CreatedAt:                  time.Now().UTC(),
			CreatedByID:                userOrg.UserID,
			UpdatedAt:                  time.Now().UTC(),
			UpdatedByID:                userOrg.UserID,
			BranchID:                   *userOrg.BranchID,
			OrganizationID:             userOrg.OrganizationID,
		}

		if err := c.core.GeneratedSavingsInterestEntryManager.Create(context, entry); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated savings interest entry creation failed (/generated-savings-interest-entry), db error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create generated savings interest entry: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created generated savings interest entry (/generated-savings-interest-entry): " + entry.ID.String(),
			Module:      "GeneratedSavingsInterestEntry",
		})
		return ctx.JSON(http.StatusCreated, c.core.GeneratedSavingsInterestEntryManager.ToModel(entry))
	})

	// PUT /generated-savings-interest-entry/:entry_id: Update generated savings interest entry by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-savings-interest-entry/:entry_id",
		Method:       "PUT",
		Note:         "Updates an existing generated savings interest entry by its ID.",
		RequestType:  core.GeneratedSavingsInterestEntryRequest{},
		ResponseType: core.GeneratedSavingsInterestEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		entryID, err := handlers.EngineUUIDParam(ctx, "entry_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), invalid entry ID.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid entry ID"})
		}

		req, err := c.core.GeneratedSavingsInterestEntryManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), validation error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest entry data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), user org error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		entry, err := c.core.GeneratedSavingsInterestEntryManager.GetByID(context, *entryID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), entry not found.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated savings interest entry not found"})
		}

		generatedSavingsInterest, err := c.core.GeneratedSavingsInterestManager.GetByID(context, entry.GeneratedSavingsInterestID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), parent generated savings interest not found.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Parent generated savings interest not found"})
		}
		if generatedSavingsInterest.PostedDate != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), parent generated savings interest is already posted.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot update entry because the parent generated savings interest is already posted"})
		}

		// Get daily ending balances for the computation period to get the final balance
		dailyBalances, err := c.core.GetDailyEndingBalances(
			context,
			generatedSavingsInterest.LastComputationDate,
			generatedSavingsInterest.NewComputationDate,
			req.AccountID,
			req.MemberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), failed to get daily balances: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get daily balances: " + err.Error()})
		}
		var savingsType usecase.SavingsType
		switch generatedSavingsInterest.SavingsComputationType {
		case core.SavingsComputationTypeDailyLowestBalance:
			savingsType = usecase.SavingsTypeLowest
		case core.SavingsComputationTypeAverageDailyBalance:
			savingsType = usecase.SavingsTypeAverage
		case core.SavingsComputationTypeMonthlyEndLowestBalance:
			savingsType = usecase.SavingsTypeLowest
		case core.SavingsComputationTypeADBEndBalance:
			savingsType = usecase.SavingsTypeEnd
		case core.SavingsComputationTypeMonthlyLowestBalanceAverage:
			savingsType = usecase.SavingsTypeLowest
		case core.SavingsComputationTypeMonthlyEndBalanceAverage:
			savingsType = usecase.SavingsTypeAverage
		case core.SavingsComputationTypeMonthlyEndBalanceTotal:
			savingsType = usecase.SavingsTypeEnd
		default:
			savingsType = usecase.SavingsTypeLowest
		}
		result := c.usecase.GetSavingsEndingBalance(usecase.SavingsBalanceComputation{
			DailyBalance:   dailyBalances,
			SavingsType:    savingsType,
			InterestAmount: req.InterestAmount,
			InterestTax:    req.InterestTax,
		})
		entry.AccountID = req.AccountID
		entry.MemberProfileID = req.MemberProfileID
		entry.EndingBalance = result.Balance
		entry.InterestAmount = result.InterestAmount
		entry.InterestTax = result.InterestTax
		entry.UpdatedAt = time.Now().UTC()
		if err := c.core.GeneratedSavingsInterestEntryManager.UpdateByID(context, entry.ID, entry); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), db error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update generated savings interest entry: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated generated savings interest entry (/generated-savings-interest-entry/:entry_id): " + entry.ID.String(),
			Module:      "GeneratedSavingsInterestEntry",
		})
		return ctx.JSON(http.StatusOK, c.core.GeneratedSavingsInterestEntryManager.ToModel(entry))
	})

	// DELETE /generated-savings-interest-entry/:entry_id: Delete a generated savings interest entry by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/generated-savings-interest-entry/:entry_id",
		Method: "DELETE",
		Note:   "Deletes the specified generated savings interest entry by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		entryID, err := handlers.EngineUUIDParam(ctx, "entry_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Generated savings interest entry delete failed (/generated-savings-interest-entry/:entry_id), invalid entry ID.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid entry ID"})
		}
		entry, err := c.core.GeneratedSavingsInterestEntryManager.GetByID(context, *entryID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Generated savings interest entry delete failed (/generated-savings-interest-entry/:entry_id), not found.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated savings interest entry not found"})
		}
		if err := c.core.GeneratedSavingsInterestEntryManager.Delete(context, *entryID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Generated savings interest entry delete failed (/generated-savings-interest-entry/:entry_id), db error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete generated savings interest entry: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted generated savings interest entry (/generated-savings-interest-entry/:entry_id): " + entry.ID.String(),
			Module:      "GeneratedSavingsInterestEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/generated-savings-interest-entry/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple generated savings interest entries by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete generated savings interest entries (/generated-savings-interest-entry/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete generated savings interest entries (/generated-savings-interest-entry/bulk-delete) | no IDs provided",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No entry IDs provided for bulk delete"})
		}

		if err := c.core.GeneratedSavingsInterestEntryManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete generated savings interest entries (/generated-savings-interest-entry/bulk-delete) | error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete generated savings interest entries: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted generated savings interest entries (/generated-savings-interest-entry/bulk-delete)",
			Module:      "GeneratedSavingsInterestEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// GET /api/v1//generated-savings-interest-entry/:generated_savings_interest_entry_id/daily-balance
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-savings-interest-entry/:generated_savings_interest_entry_id/daily-balance",
		Method:       "GET",
		Note:         "Fetches daily ending balances for all entries under a specific generated savings interest record.",
		ResponseType: core.GeneratedSavingsInterestEntryDailyBalanceResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedSavingsInterestEntryID, err := handlers.EngineUUIDParam(ctx, "generated_savings_interest_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest ID"})
		}
		dailyBalances, err := c.core.DailyBalances(context, *generatedSavingsInterestEntryID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch daily ending balances: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, dailyBalances)
	})
}
