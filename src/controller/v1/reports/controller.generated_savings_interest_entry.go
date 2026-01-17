package reports

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/usecase"
	"github.com/labstack/echo/v4"
)

func GeneratedSavingsInterestEntryController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-savings-interest-entry",
		Method:       "GET",
		Note:         "Returns all generated savings interest entries for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: types.GeneratedSavingsInterestEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := core.GenerateSavingsInterestEntryCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated savings interest entries found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.GeneratedSavingsInterestEntryManager(service).ToModels(entries))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-savings-interest-entry/search",
		Method:       "GET",
		Note:         "Returns a paginated list of generated savings interest entries for the current user's organization and branch.",
		ResponseType: types.GeneratedSavingsInterestEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := core.GeneratedSavingsInterestEntryManager(service).NormalPagination(context, ctx, &types.GeneratedSavingsInterestEntry{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch generated savings interest entries for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-savings-interest-entry/:entry_id",
		Method:       "GET",
		Note:         "Returns a single generated savings interest entry by its ID.",
		ResponseType: types.GeneratedSavingsInterestEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		entryID, err := helpers.EngineUUIDParam(ctx, "entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid entry ID"})
		}
		entry, err := core.GeneratedSavingsInterestEntryManager(service).GetByIDRaw(context, *entryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated savings interest entry not found"})
		}
		return ctx.JSON(http.StatusOK, entry)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-savings-interest-entry/generated-savings-interest/:generated_savings_interest_id",
		Method:       "POST",
		Note:         "Creates a new generated savings interest entry for the current user's organization and branch.",
		RequestType:  types.GeneratedSavingsInterestEntryRequest{},
		ResponseType: types.GeneratedSavingsInterestEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedSavingsInterestID, err := helpers.EngineUUIDParam(ctx, "generated_savings_interest_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated savings interest entry creation failed (/generated-savings-interest-entry), invalid generated savings interest ID.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest ID"})
		}
		req, err := core.GeneratedSavingsInterestEntryManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated savings interest entry creation failed (/generated-savings-interest-entry), validation error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest entry data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated savings interest entry creation failed (/generated-savings-interest-entry), user org error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated savings interest entry creation failed (/generated-savings-interest-entry), user not assigned to branch.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		generatedSavingsInterest, err := core.GeneratedSavingsInterestManager(service).GetByID(
			context, *generatedSavingsInterestID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), parent generated savings interest not found.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Parent generated savings interest not found"})
		}
		if generatedSavingsInterest.PostedDate != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), parent generated savings interest is already posted.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot update entry because the parent generated savings interest is already posted"})
		}
		dailyBalances, err := core.GetDailyEndingBalances(
			context, service,
			generatedSavingsInterest.LastComputationDate,
			generatedSavingsInterest.NewComputationDate,
			req.AccountID,
			req.MemberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), failed to get daily balances: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get daily balances: " + err.Error()})
		}

		var savingsType usecase.SavingsType
		switch generatedSavingsInterest.SavingsComputationType {
		case types.SavingsComputationTypeDailyLowestBalance:
			savingsType = usecase.SavingsTypeLowest
		case types.SavingsComputationTypeAverageDailyBalance:
			savingsType = usecase.SavingsTypeAverage
		case types.SavingsComputationTypeMonthlyEndLowestBalance:
			savingsType = usecase.SavingsTypeLowest
		case types.SavingsComputationTypeADBEndBalance:
			savingsType = usecase.SavingsTypeEnd
		case types.SavingsComputationTypeMonthlyLowestBalanceAverage:
			savingsType = usecase.SavingsTypeLowest
		case types.SavingsComputationTypeMonthlyEndBalanceAverage:
			savingsType = usecase.SavingsTypeAverage
		case types.SavingsComputationTypeMonthlyEndBalanceTotal:
			savingsType = usecase.SavingsTypeEnd
		default:
			savingsType = usecase.SavingsTypeLowest
		}

		result := usecase.GetSavingsEndingBalance(usecase.SavingsBalanceComputation{
			DailyBalance:   dailyBalances,
			SavingsType:    savingsType,
			InterestAmount: req.InterestAmount,
			InterestTax:    req.InterestTax,
		})

		entry := &types.GeneratedSavingsInterestEntry{
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

		if err := core.GeneratedSavingsInterestEntryManager(service).Create(context, entry); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated savings interest entry creation failed (/generated-savings-interest-entry), db error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create generated savings interest entry: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created generated savings interest entry (/generated-savings-interest-entry): " + entry.ID.String(),
			Module:      "GeneratedSavingsInterestEntry",
		})
		return ctx.JSON(http.StatusCreated, core.GeneratedSavingsInterestEntryManager(service).ToModel(entry))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-savings-interest-entry/:entry_id",
		Method:       "PUT",
		Note:         "Updates an existing generated savings interest entry by its ID.",
		RequestType:  types.GeneratedSavingsInterestEntryRequest{},
		ResponseType: types.GeneratedSavingsInterestEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		entryID, err := helpers.EngineUUIDParam(ctx, "entry_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), invalid entry ID.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid entry ID"})
		}

		req, err := core.GeneratedSavingsInterestEntryManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), validation error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest entry data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), user org error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		entry, err := core.GeneratedSavingsInterestEntryManager(service).GetByID(context, *entryID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), entry not found.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated savings interest entry not found"})
		}

		generatedSavingsInterest, err := core.GeneratedSavingsInterestManager(service).GetByID(context, entry.GeneratedSavingsInterestID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), parent generated savings interest not found.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Parent generated savings interest not found"})
		}
		if generatedSavingsInterest.PostedDate != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), parent generated savings interest is already posted.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot update entry because the parent generated savings interest is already posted"})
		}

		dailyBalances, err := core.GetDailyEndingBalances(
			context, service,
			generatedSavingsInterest.LastComputationDate,
			generatedSavingsInterest.NewComputationDate,
			req.AccountID,
			req.MemberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), failed to get daily balances: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get daily balances: " + err.Error()})
		}
		var savingsType usecase.SavingsType
		switch generatedSavingsInterest.SavingsComputationType {
		case types.SavingsComputationTypeDailyLowestBalance:
			savingsType = usecase.SavingsTypeLowest
		case types.SavingsComputationTypeAverageDailyBalance:
			savingsType = usecase.SavingsTypeAverage
		case types.SavingsComputationTypeMonthlyEndLowestBalance:
			savingsType = usecase.SavingsTypeLowest
		case types.SavingsComputationTypeADBEndBalance:
			savingsType = usecase.SavingsTypeEnd
		case types.SavingsComputationTypeMonthlyLowestBalanceAverage:
			savingsType = usecase.SavingsTypeLowest
		case types.SavingsComputationTypeMonthlyEndBalanceAverage:
			savingsType = usecase.SavingsTypeAverage
		case types.SavingsComputationTypeMonthlyEndBalanceTotal:
			savingsType = usecase.SavingsTypeEnd
		default:
			savingsType = usecase.SavingsTypeLowest
		}
		result := usecase.GetSavingsEndingBalance(usecase.SavingsBalanceComputation{
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
		if err := core.GeneratedSavingsInterestEntryManager(service).UpdateByID(context, entry.ID, entry); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated savings interest entry update failed (/generated-savings-interest-entry/:entry_id), db error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update generated savings interest entry: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated generated savings interest entry (/generated-savings-interest-entry/:entry_id): " + entry.ID.String(),
			Module:      "GeneratedSavingsInterestEntry",
		})
		return ctx.JSON(http.StatusOK, core.GeneratedSavingsInterestEntryManager(service).ToModel(entry))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/generated-savings-interest-entry/:entry_id",
		Method: "DELETE",
		Note:   "Deletes the specified generated savings interest entry by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		entryID, err := helpers.EngineUUIDParam(ctx, "entry_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Generated savings interest entry delete failed (/generated-savings-interest-entry/:entry_id), invalid entry ID.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid entry ID"})
		}
		entry, err := core.GeneratedSavingsInterestEntryManager(service).GetByID(context, *entryID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Generated savings interest entry delete failed (/generated-savings-interest-entry/:entry_id), not found.",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated savings interest entry not found"})
		}
		if err := core.GeneratedSavingsInterestEntryManager(service).Delete(context, *entryID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Generated savings interest entry delete failed (/generated-savings-interest-entry/:entry_id), db error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete generated savings interest entry: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted generated savings interest entry (/generated-savings-interest-entry/:entry_id): " + entry.ID.String(),
			Module:      "GeneratedSavingsInterestEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/generated-savings-interest-entry/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple generated savings interest entries by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete generated savings interest entries (/generated-savings-interest-entry/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete generated savings interest entries (/generated-savings-interest-entry/bulk-delete) | no IDs provided",
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No entry IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.GeneratedSavingsInterestEntryManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete generated savings interest entries (/generated-savings-interest-entry/bulk-delete) | error: " + err.Error(),
				Module:      "GeneratedSavingsInterestEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete generated savings interest entries: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted generated savings interest entries (/generated-savings-interest-entry/bulk-delete)",
			Module:      "GeneratedSavingsInterestEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-savings-interest-entry/:generated_savings_interest_entry_id/daily-balance",
		Method:       "GET",
		Note:         "Fetches daily ending balances for all entries under a specific generated savings interest record.",
		ResponseType: types.GeneratedSavingsInterestEntryDailyBalanceResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedSavingsInterestEntryID, err := helpers.EngineUUIDParam(ctx, "generated_savings_interest_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest ID"})
		}
		dailyBalances, err := core.DailyBalances(context, service, *generatedSavingsInterestEntryID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch daily ending balances: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, dailyBalances)
	})
}
