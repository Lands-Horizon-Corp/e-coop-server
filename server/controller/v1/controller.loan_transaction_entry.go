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

func (c *Controller) loanTransactionEntryController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction-entry/loan-transaction/:loan_transaction_id/deduction",
		Method:       "POST",
		Note:         "Adds a deduction to a loan transaction by ID.",
		RequestType:  core.LoanTransactionDeductionRequest{},
		ResponseType: core.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req core.LoanTransactionDeductionRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction deduction failed: invalid payload: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction deduction payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction deduction failed: validation error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		account, err := c.core.AccountManager.GetByID(context, req.AccountID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Account not found for loan transaction deduction: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found: " + err.Error()})
		}

		loanTransaction := &core.LoanTransactionEntry{
			CreatedByID:       userOrg.UserID,
			UpdatedByID:       userOrg.UserID,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			LoanTransactionID: *loanTransactionID,
			Type:              core.LoanTransactionDeduction,
			Debit:             0,
			Credit:            req.Amount,
			IsAddOn:           req.IsAddOn,
			AccountID:         &req.AccountID,
			Name:              account.Name,
			Description:       account.Description,
		}
		if err := c.core.LoanTransactionEntryManager.Create(context, loanTransaction); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan transaction deduction creation failed: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction deduction: " + err.Error()})
		}

		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + endTx(tx.Error).Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}
		newLoanTransaction, err := c.event.LoanBalancing(context, ctx, tx, endTx, event.LoanBalanceEvent{
			LoanTransactionID: *loanTransactionID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to balance loan transaction: %v", err)})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTransactionManager.ToModel(newLoanTransaction))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction-entry/:loan_transaction_entry_id/deduction",
		Method:       "PUT",
		Note:         "Adds a deduction to a loan transaction by ID.",
		RequestType:  core.LoanTransactionDeductionRequest{},
		ResponseType: core.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionEntryID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req core.LoanTransactionDeductionRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction deduction failed: invalid payload: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction deduction payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction deduction failed: validation error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		account, err := c.core.AccountManager.GetByID(context, req.AccountID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Account not found for loan transaction deduction: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found: " + err.Error()})
		}
		loanTransactionEntry, err := c.core.LoanTransactionEntryManager.GetByID(context, *loanTransactionEntryID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Loan transaction entry not found for deduction update: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction entry not found for deduction update: " + err.Error()})
		}
		if loanTransactionEntry.Type == core.LoanTransactionAutomaticDeduction {
			loanTransactionEntry.Credit = req.Amount
			loanTransactionEntry.IsAddOn = req.IsAddOn
		} else {
			loanTransactionEntry.Credit = req.Amount
			loanTransactionEntry.IsAddOn = req.IsAddOn
			loanTransactionEntry.AccountID = &req.AccountID
			loanTransactionEntry.Name = account.Name
		}
		loanTransactionEntry.Amount = req.Amount
		loanTransactionEntry.UpdatedAt = time.Now().UTC()
		loanTransactionEntry.UpdatedByID = userOrg.UserID
		if err := c.core.LoanTransactionEntryManager.UpdateByID(context, *loanTransactionEntryID, loanTransactionEntry); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan transaction deduction creation failed: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction deduction: " + err.Error()})
		}
		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + endTx(tx.Error).Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}
		newLoanTransaction, err := c.event.LoanBalancing(context, ctx, tx, endTx, event.LoanBalanceEvent{
			LoanTransactionID: loanTransactionEntry.LoanTransactionID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to balance loan transaction: %v", err)})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTransactionManager.ToModel(newLoanTransaction))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/loan-transaction-entry/:loan_transaction_entry_id",
		Method: "DELETE",
		Note:   "Deletes a loan transaction entry by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionEntryID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction entry ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete loan transaction entries"})
		}

		loanTransactionEntry, err := c.core.LoanTransactionEntryManager.GetByID(context, *loanTransactionEntryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction entry not found"})
		}
		if loanTransactionEntry.Type == core.LoanTransactionAutomaticDeduction {
			loanTransactionEntry.IsAutomaticLoanDeductionDeleted = true
			loanTransactionEntry.UpdatedAt = time.Now().UTC()
			loanTransactionEntry.UpdatedByID = userOrg.UserID
			if err := c.core.LoanTransactionEntryManager.UpdateByID(context, loanTransactionEntry.ID, loanTransactionEntry); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction entry: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, map[string]string{"message": "Loan transaction entry deleted successfully"})
		}

		if loanTransactionEntry.OrganizationID != userOrg.OrganizationID || loanTransactionEntry.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction entry"})
		}

		loanTransactionEntry.DeletedByID = &userOrg.UserID

		if err := c.core.LoanTransactionEntryManager.Delete(context, loanTransactionEntry.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction entry: " + err.Error()})
		}

		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + endTx(tx.Error).Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}
		_, err = c.event.LoanBalancing(context, ctx, tx, endTx, event.LoanBalanceEvent{
			LoanTransactionID: loanTransactionEntry.LoanTransactionID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to balance loan transaction: %v", err)})
		}
		return ctx.JSON(http.StatusOK, map[string]string{"message": "Loan transaction entry deleted successfully"})
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/loan-transaction-entry/:loan_transaction_entry_id/restore",
		Method: "PUT",
		Note:   "Restores a deleted automatic loan deduction entry by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionEntryID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction entry ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to restore loan transaction entries"})
		}

		loanTransactionEntry, err := c.core.LoanTransactionEntryManager.GetByID(context, *loanTransactionEntryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction entry not found"})
		}

		if loanTransactionEntry.OrganizationID != userOrg.OrganizationID || loanTransactionEntry.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction entry"})
		}

		if loanTransactionEntry.Type != core.LoanTransactionAutomaticDeduction {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Only automatic loan deduction entries can be restored"})
		}

		if !loanTransactionEntry.IsAutomaticLoanDeductionDeleted {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Loan transaction entry is not marked as deleted"})
		}

		loanTransactionEntry.IsAutomaticLoanDeductionDeleted = false
		loanTransactionEntry.UpdatedAt = time.Now().UTC()
		loanTransactionEntry.UpdatedByID = userOrg.UserID

		if err := c.core.LoanTransactionEntryManager.UpdateByID(context, loanTransactionEntry.ID, loanTransactionEntry); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to restore loan transaction entry: " + err.Error()})
		}

		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + endTx(tx.Error).Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		loanTransaction, err := c.event.LoanBalancing(context, ctx, tx, endTx, event.LoanBalanceEvent{
			LoanTransactionID: loanTransactionEntry.LoanTransactionID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to balance loan transaction: %v", err)})
		}

		return ctx.JSON(http.StatusOK, c.core.LoanTransactionManager.ToModel(loanTransaction))
	})
}
