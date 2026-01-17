package transactions

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
	"gorm.io/gorm"
)

func AdjustmentEntryController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-entry",
		Method:       "GET",
		Note:         "Returns all adjustment entries for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: types.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		adjustmentEntries, err := core.AdjustmentEntryCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No adjustment entries found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.AdjustmentEntryManager(service).ToModels(adjustmentEntries))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-entry/search",
		Method:       "GET",
		Note:         "Returns a paginated list of adjustment entries for the current user's organization and branch.",
		ResponseType: types.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		adjustmentEntries, err := core.AdjustmentEntryManager(service).NormalPagination(context, ctx, &types.AdjustmentEntry{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entries for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, adjustmentEntries)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-entry/:adjustment_entry_id",
		Method:       "GET",
		Note:         "Returns a single adjustment entry by its ID.",
		ResponseType: types.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		adjustmentEntryID, err := helpers.EngineUUIDParam(ctx, "adjustment_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry ID"})
		}
		adjustmentEntry, err := core.AdjustmentEntryManager(service).GetByIDRaw(context, *adjustmentEntryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Adjustment entry not found"})
		}
		return ctx.JSON(http.StatusOK, adjustmentEntry)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-entry",
		Method:       "POST",
		Note:         "Creates a new adjustment entry for the current user's organization and branch.",
		RequestType: types.AdjustmentEntryRequest{},
		ResponseType: types.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.AdjustmentEntryManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), validation error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), user org error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), user not assigned to branch.",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		transactionBatch, err := core.TransactionBatchCurrent(
			context, service,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), transaction batch lookup error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}

		adjustmentEntry := &types.AdjustmentEntry{
			SignatureMediaID:   req.SignatureMediaID,
			AccountID:          req.AccountID,
			MemberProfileID:    req.MemberProfileID,
			EmployeeUserID:     &userOrg.UserID,
			PaymentTypeID:      req.PaymentTypeID,
			TypeOfPaymentType:  req.TypeOfPaymentType,
			Description:        req.Description,
			ReferenceNumber:    req.ReferenceNumber,
			EntryDate:          req.EntryDate,
			Debit:              req.Debit,
			Credit:             req.Credit,
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			BranchID:           *userOrg.BranchID,
			OrganizationID:     userOrg.OrganizationID,
			TransactionBatchID: &transactionBatch.ID,
		}
		transactionRequest := event.RecordTransactionRequest{
			Debit:  req.Debit,
			Credit: req.Credit,

			AccountID:          req.AccountID,
			MemberProfileID:    req.MemberProfileID,
			TransactionBatchID: transactionBatch.ID,

			LoanTransactionID:     req.LoanTransactionID,
			ReferenceNumber:       req.ReferenceNumber,
			Description:           req.Description,
			EntryDate:             req.EntryDate,
			SignatureMediaID:      req.SignatureMediaID,
			PaymentTypeID:         req.PaymentTypeID,
			BankReferenceNumber:   "",  // Not applicable for adjustment entries
			BankID:                nil, // Not applicable for adjustment entries
			ProofOfPaymentMediaID: nil, // Not applicable for adjustment entries
		}

		if err := event.RecordTransaction(context, service, transactionRequest, core.GeneralLedgerSourceAdjustment, userOrg); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "transaction-recording-failed",
				Description: "Failed to record adjustment entry transaction in general ledger for reference " + req.ReferenceNumber + ": " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Adjustment entry created but failed to record transaction: " + err.Error(),
			})
		}

		if err := core.AdjustmentEntryManager(service).Create(context, adjustmentEntry); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), db error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create adjustment entry: " + err.Error()})
		}

		if err := event.TransactionBatchBalancing(context, service, &transactionBatch.ID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), transaction batch balancing error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch after adjustment entry creation: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created adjustment entry (/adjustment-entry): " + adjustmentEntry.ReferenceNumber,
			Module:      "AdjustmentEntry",
		})
		return ctx.JSON(http.StatusCreated, core.AdjustmentEntryManager(service).ToModel(adjustmentEntry))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/adjustment-entry/:adjustment_entry_id",
		Method: "DELETE",
		Note:   "Deletes the specified adjustment entry by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		adjustmentEntryID, err := helpers.EngineUUIDParam(ctx, "adjustment_entry_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Adjustment entry delete failed (/adjustment-entry/:adjustment_entry_id), invalid adjustment entry ID.",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry ID"})
		}
		adjustmentEntry, err := core.AdjustmentEntryManager(service).GetByID(context, *adjustmentEntryID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Adjustment entry delete failed (/adjustment-entry/:adjustment_entry_id), not found.",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Adjustment entry not found"})
		}
		if err := core.AdjustmentEntryManager(service).Delete(context, *adjustmentEntryID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Adjustment entry delete failed (/adjustment-entry/:adjustment_entry_id), db error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete adjustment entry: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted adjustment entry (/adjustment-entry/:adjustment_entry_id): " + adjustmentEntry.ReferenceNumber,
			Module:      "AdjustmentEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/adjustment-entry/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple adjustment entries by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete adjustment entries (/adjustment-entry/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete adjustment entries (/adjustment-entry/bulk-delete) | no IDs provided",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No adjustment entry IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.AdjustmentEntryManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete adjustment entries (/adjustment-entry/bulk-delete) | error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete adjustment entries: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted adjustment entries (/adjustment-entry/bulk-delete)",
			Module:      "AdjustmentEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-entry/total",
		Method:       "GET",
		Note:         "Returns the total debit and credit of all adjustment entries for the current user's organization and branch.",
		ResponseType: types.AdjustmentEntryTotalResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := core.AdjustmentEntryCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No adjustment entries found for the current branch"})
		}
		balance, err := usecase.CalculateBalance(usecase.Balance{
			AdjustmentEntries: entries,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to compute total balance: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.AdjustmentEntryTotalResponse{
			TotalDebit:  balance.Debit,
			TotalCredit: balance.Credit,
			Balance:     balance.Balance,
			IsBalanced:  balance.IsBalanced,
		})
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-entry/currency/:currency_id/search",
		Method:       "GET",
		Note:         "Returns a paginated list of adjustment entries filtered by currency and optionally by user organization.",
		ResponseType: types.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		currencyID, err := helpers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		paginated, err := core.AdjustmentEntryManager(service).RawPagination(
			context,
			ctx,
			func(db *gorm.DB) *gorm.DB {
				query := db.Model(&types.AdjustmentEntry{}).
					Joins("JOIN accounts a ON a.id = adjustment_entries.account_id").
					Where("adjustment_entries.organization_id = ?", userOrg.OrganizationID).
					Where("adjustment_entries.branch_id = ?", *userOrg.BranchID)

				if currencyID != nil {
					query = query.Where("a.currency_id = ?", *currencyID)
				}

				return query
			},
			"Organization", "Branch", "Account", "EmployeeUser",
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to fetch adjustment entries for pagination: " + err.Error(),
			})
		}

		return ctx.JSON(http.StatusOK, paginated)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-entry/currency/:currency_id/total",
		Method:       "GET",
		Note:         "Returns the total amount of adjustment entries filtered by currency and optionally by user organization.",
		ResponseType: types.AdjustmentEntryTotalResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		currencyID, err := helpers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		entries, err := core.AdjustmentEntryManager(service).Find(context, &types.AdjustmentEntry{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entries for pagination: " + err.Error()})
		}
		balance, err := usecase.CalculateBalance(usecase.Balance{
			AdjustmentEntries: entries,
			CurrencyID:        currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to compute total balance: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.AdjustmentEntryTotalResponse{
			TotalDebit:  balance.Debit,
			TotalCredit: balance.Credit,
			Balance:     balance.Balance,
			IsBalanced:  balance.IsBalanced,
		})
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-entry/currency/:currency_id/employee/:user_organization_id/search",
		Method:       "GET",
		Note:         "Returns a paginated list of adjustment entries filtered by currency and user organization.",
		ResponseType: types.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := helpers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User organization not found"})
		}
		if userOrganization.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User organization is not assigned to a branch"})
		}
		paginated, err := core.AdjustmentEntryManager(service).RawPagination(
			context,
			ctx,
			func(db *gorm.DB) *gorm.DB {
				query := db.Model(&types.AdjustmentEntry{}).
					Joins("JOIN accounts a ON a.id = adjustment_entries.account_id").
					Where("adjustment_entries.organization_id = ?", userOrganization.OrganizationID).
					Where("adjustment_entries.branch_id = ?", *userOrganization.BranchID).
					Where("adjustment_entries.employee_user_id = ?", userOrganization.UserID)

				if currencyID != nil {
					query = query.Where("a.currency_id = ?", *currencyID)
				}

				return query
			},
			"Organization", "Branch", "EmployeeUser", "Account",
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to fetch adjustment entries for pagination: " + err.Error(),
			})
		}

		return ctx.JSON(http.StatusOK, paginated)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-entry/currency/:currency_id/employee/:user_organization_id/total",
		Method:       "GET",
		Note:         "Returns the total amount of adjustment entries filtered by currency and user organization.",
		ResponseType: types.AdjustmentEntryTotalResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := helpers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User organization not found"})
		}
		if userOrganization.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User organization is not assigned to a branch"})
		}
		entries, err := core.AdjustmentEntryManager(service).Find(context, &types.AdjustmentEntry{
			OrganizationID: userOrganization.OrganizationID,
			BranchID:       *userOrganization.BranchID,
			EmployeeUserID: &userOrganization.UserID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entries for pagination: " + err.Error()})
		}
		balance, err := usecase.CalculateBalance(usecase.Balance{
			AdjustmentEntries: entries,
			CurrencyID:        currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to compute total balance: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.AdjustmentEntryTotalResponse{
			TotalDebit:  balance.Debit,
			TotalCredit: balance.Credit,
			Balance:     balance.Balance,
			IsBalanced:  balance.IsBalanced,
		})
	})

}
