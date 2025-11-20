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

// AdjustmentEntryController registers routes for managing adjustment entries.
func (c *Controller) adjustmentEntryController() {
	req := c.provider.Service.Request

	// GET /adjustment-entry: List all adjustment entries for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry",
		Method:       "GET",
		Note:         "Returns all adjustment entries for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		adjustmentEntries, err := c.core.AdjustmentEntryCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No adjustment entries found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.AdjustmentEntryManager.ToModels(adjustmentEntries))
	})

	// GET /adjustment-entry/search: Paginated search of adjustment entries for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/search",
		Method:       "GET",
		Note:         "Returns a paginated list of adjustment entries for the current user's organization and branch.",
		ResponseType: core.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		adjustmentEntries, err := c.core.AdjustmentEntryManager.PaginationWithFields(context, ctx, &core.AdjustmentEntry{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entries for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, adjustmentEntries)
	})

	// GET /adjustment-entry/:adjustment_entry_id: Get specific adjustment entry by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/:adjustment_entry_id",
		Method:       "GET",
		Note:         "Returns a single adjustment entry by its ID.",
		ResponseType: core.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		adjustmentEntryID, err := handlers.EngineUUIDParam(ctx, "adjustment_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry ID"})
		}
		adjustmentEntry, err := c.core.AdjustmentEntryManager.GetByIDRaw(context, *adjustmentEntryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Adjustment entry not found"})
		}
		return ctx.JSON(http.StatusOK, adjustmentEntry)
	})

	// POST /adjustment-entry: Create a new adjustment entry. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry",
		Method:       "POST",
		Note:         "Creates a new adjustment entry for the current user's organization and branch.",
		RequestType:  core.AdjustmentEntryRequest{},
		ResponseType: core.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.AdjustmentEntryManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), validation error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), user org error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), user not assigned to branch.",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		transactionBatch, err := c.core.TransactionBatchCurrent(
			context,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), transaction batch lookup error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}

		adjustmentEntry := &core.AdjustmentEntry{
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
		// ================================================================================
		// STEP 2: RECORD TRANSACTION IN GENERAL LEDGER
		// ================================================================================
		// Create transaction request for general ledger recording
		transactionRequest := event.RecordTransactionRequest{
			// Financial amounts
			Debit:  req.Debit,
			Credit: req.Credit,

			// Account and member information
			AccountID:          req.AccountID,
			MemberProfileID:    req.MemberProfileID,
			TransactionBatchID: transactionBatch.ID,

			LoanTransactionID: req.LoanTransactionID,
			// Transaction metadata
			ReferenceNumber:       req.ReferenceNumber,
			Description:           req.Description,
			EntryDate:             req.EntryDate,
			SignatureMediaID:      req.SignatureMediaID,
			PaymentTypeID:         req.PaymentTypeID,
			BankReferenceNumber:   "",  // Not applicable for adjustment entries
			BankID:                nil, // Not applicable for adjustment entries
			ProofOfPaymentMediaID: nil, // Not applicable for adjustment entries
		}

		// Record the transaction in general ledger with adjustment entry source
		if err := c.event.RecordTransaction(context, ctx, transactionRequest, core.GeneralLedgerSourceAdjustment); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "transaction-recording-failed",
				Description: "Failed to record adjustment entry transaction in general ledger for reference " + req.ReferenceNumber + ": " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Adjustment entry created but failed to record transaction: " + err.Error(),
			})
		}

		if err := c.core.AdjustmentEntryManager.Create(context, adjustmentEntry); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), db error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create adjustment entry: " + err.Error()})
		}

		if err := c.event.TransactionBatchBalancing(context, &transactionBatch.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), transaction batch balancing error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch after adjustment entry creation: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created adjustment entry (/adjustment-entry): " + adjustmentEntry.ReferenceNumber,
			Module:      "AdjustmentEntry",
		})
		return ctx.JSON(http.StatusCreated, c.core.AdjustmentEntryManager.ToModel(adjustmentEntry))
	})

	// DELETE /adjustment-entry/:adjustment_entry_id: Delete an adjustment entry by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/adjustment-entry/:adjustment_entry_id",
		Method: "DELETE",
		Note:   "Deletes the specified adjustment entry by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		adjustmentEntryID, err := handlers.EngineUUIDParam(ctx, "adjustment_entry_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Adjustment entry delete failed (/adjustment-entry/:adjustment_entry_id), invalid adjustment entry ID.",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry ID"})
		}
		adjustmentEntry, err := c.core.AdjustmentEntryManager.GetByID(context, *adjustmentEntryID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Adjustment entry delete failed (/adjustment-entry/:adjustment_entry_id), not found.",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Adjustment entry not found"})
		}
		if err := c.core.AdjustmentEntryManager.Delete(context, *adjustmentEntryID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Adjustment entry delete failed (/adjustment-entry/:adjustment_entry_id), db error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete adjustment entry: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted adjustment entry (/adjustment-entry/:adjustment_entry_id): " + adjustmentEntry.ReferenceNumber,
			Module:      "AdjustmentEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /adjustment-entry/bulk-delete: Bulk delete adjustment entries by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/adjustment-entry/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple adjustment entries by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete adjustment entries (/adjustment-entry/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete adjustment entries (/adjustment-entry/bulk-delete) | no IDs provided",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No adjustment entry IDs provided for bulk delete"})
		}

		if err := c.core.AdjustmentEntryManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete adjustment entries (/adjustment-entry/bulk-delete) | error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete adjustment entries: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted adjustment entries (/adjustment-entry/bulk-delete)",
			Module:      "AdjustmentEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// GET /api/v1/adjustment-entry/total
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/total",
		Method:       "GET",
		Note:         "Returns the total debit and credit of all adjustment entries for the current user's organization and branch.",
		ResponseType: core.AdjustmentEntryTotalResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		entries, err := c.core.AdjustmentEntryCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No adjustment entries found for the current branch"})
		}
		balance, err := c.usecase.Balance(usecase.Balance{
			AdjustmentEntries: entries,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to compute total balance: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.AdjustmentEntryTotalResponse{
			TotalDebit:  balance.Debit,
			TotalCredit: balance.Credit,
			Balance:     balance.Balance,
		})
	})

	// GET api/v1/adjustment-entry/currency/:currency_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/currency/:currency_id/search",
		Method:       "GET",
		Note:         "Returns a paginated list of adjustment entries filtered by currency and optionally by user organization.",
		ResponseType: core.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		adjustmentEntries, err := c.core.AdjustmentEntryManager.Find(context, &core.AdjustmentEntry{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entries for pagination: " + err.Error()})
		}
		result := []*core.AdjustmentEntry{}
		for _, entry := range adjustmentEntries {
			if handlers.UUIDPtrEqual(entry.Account.CurrencyID, currencyID) {
				result = append(result, entry)
			}
		}

		paginated, err := c.core.AdjustmentEntryManager.PaginationData(context, ctx, result)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to paginate adjustment entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, paginated)
	})

	// GET api/v1/adjustment-entry/currency/:currency_id/total
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/currency/:currency_id/total",
		Method:       "GET",
		Note:         "Returns the total amount of adjustment entries filtered by currency and optionally by user organization.",
		ResponseType: core.AdjustmentEntryTotalResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		entries, err := c.core.AdjustmentEntryManager.Find(context, &core.AdjustmentEntry{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entries for pagination: " + err.Error()})
		}
		balance, err := c.usecase.Balance(usecase.Balance{
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
		})
	})

	// GET api/v1/adjustment-entry/currency/:currency_id/employee/:user_organization_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/currency/:currency_id/employee/:user_organization_id/search",
		Method:       "GET",
		Note:         "Returns a paginated list of adjustment entries filtered by currency and user organization.",
		ResponseType: core.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrganization, err := c.core.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User organization not found"})
		}
		if userOrganization.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User organization is not assigned to a branch"})
		}
		adjustmentEntries, err := c.core.AdjustmentEntryManager.Find(context, &core.AdjustmentEntry{
			OrganizationID: userOrganization.OrganizationID,
			BranchID:       *userOrganization.BranchID,
			EmployeeUserID: &userOrganization.UserID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entries for pagination: " + err.Error()})
		}
		result := []*core.AdjustmentEntry{}
		for _, entry := range adjustmentEntries {
			if handlers.UUIDPtrEqual(entry.Account.CurrencyID, currencyID) {
				result = append(result, entry)
			}
		}
		paginated, err := c.core.AdjustmentEntryManager.PaginationData(context, ctx, result)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to paginate adjustment entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, paginated)
	})

	// GET api/v1/adjustment-entry/currency/:currency_id/employee/:user_organization_id/total
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/currency/:currency_id/employee/:user_organization_id/total",
		Method:       "GET",
		Note:         "Returns the total amount of adjustment entries filtered by currency and user organization.",
		ResponseType: core.AdjustmentEntryTotalResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrganization, err := c.core.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User organization not found"})
		}
		if userOrganization.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User organization is not assigned to a branch"})
		}
		entries, err := c.core.AdjustmentEntryManager.Find(context, &core.AdjustmentEntry{
			OrganizationID: userOrganization.OrganizationID,
			BranchID:       *userOrganization.BranchID,
			EmployeeUserID: &userOrganization.UserID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entries for pagination: " + err.Error()})
		}
		balance, err := c.usecase.Balance(usecase.Balance{
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
		})
	})

}
