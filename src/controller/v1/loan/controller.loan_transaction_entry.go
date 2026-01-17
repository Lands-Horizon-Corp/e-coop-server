package loan

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

func LoanTransactionEntryController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction-entry/loan-transaction/:loan_transaction_id/deduction",
		Method:       "POST",
		Note:         "Adds a deduction to a loan transaction by ID.",
		RequestType:  core.LoanTransactionDeductionRequest{},
		ResponseType: core.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req core.LoanTransactionDeductionRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction deduction failed: invalid payload: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction deduction payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction deduction failed: validation error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		account, err := core.AccountManager(service).GetByID(context, req.AccountID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Account not found for loan transaction deduction: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found: " + err.Error()})
		}

		loanTransaction := &types.LoanTransactionEntry{
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
		if err := core.LoanTransactionEntryManager(service).Create(context, loanTransaction); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan transaction deduction creation failed: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction deduction: " + err.Error()})
		}

		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + endTx(tx.Error).Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}
		newLoanTransaction, err := event.LoanBalancing(context, service, tx, endTx, event.LoanBalanceEvent{
			LoanTransactionID: *loanTransactionID,
		}, userOrg)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to balance loan transaction: %v", err)})
		}
		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModel(newLoanTransaction))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction-entry/:loan_transaction_entry_id/deduction",
		Method:       "PUT",
		Note:         "Adds a deduction to a loan transaction by ID.",
		RequestType:  core.LoanTransactionDeductionRequest{},
		ResponseType: core.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionEntryID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req core.LoanTransactionDeductionRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction deduction failed: invalid payload: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction deduction payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction deduction failed: validation error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		account, err := core.AccountManager(service).GetByID(context, req.AccountID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Account not found for loan transaction deduction: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found: " + err.Error()})
		}
		loanTransactionEntry, err := core.LoanTransactionEntryManager(service).GetByID(context, *loanTransactionEntryID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.LoanTransactionEntryManager(service).UpdateByID(context, *loanTransactionEntryID, loanTransactionEntry); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan transaction deduction creation failed: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction deduction: " + err.Error()})
		}
		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + endTx(tx.Error).Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}
		newLoanTransaction, err := event.LoanBalancing(context, service, tx, endTx, event.LoanBalanceEvent{
			LoanTransactionID: loanTransactionEntry.LoanTransactionID,
		}, userOrg)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to balance loan transaction: %v", err)})
		}
		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModel(newLoanTransaction))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/loan-transaction-entry/:loan_transaction_entry_id",
		Method: "DELETE",
		Note:   "Deletes a loan transaction entry by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionEntryID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction entry ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete loan transaction entries"})
		}

		loanTransactionEntry, err := core.LoanTransactionEntryManager(service).GetByID(context, *loanTransactionEntryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction entry not found"})
		}
		if loanTransactionEntry.Type == core.LoanTransactionAutomaticDeduction {
			loanTransactionEntry.IsAutomaticLoanDeductionDeleted = true
			loanTransactionEntry.UpdatedAt = time.Now().UTC()
			loanTransactionEntry.UpdatedByID = userOrg.UserID
			if err := core.LoanTransactionEntryManager(service).UpdateByID(context, loanTransactionEntry.ID, loanTransactionEntry); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction entry: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, map[string]string{"message": "Loan transaction entry deleted successfully"})
		}

		if loanTransactionEntry.OrganizationID != userOrg.OrganizationID || loanTransactionEntry.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction entry"})
		}

		loanTransactionEntry.DeletedByID = &userOrg.UserID

		if err := core.LoanTransactionEntryManager(service).Delete(context, loanTransactionEntry.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction entry: " + err.Error()})
		}

		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + endTx(tx.Error).Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}
		_, err = event.LoanBalancing(context, service, tx, endTx, event.LoanBalanceEvent{
			LoanTransactionID: loanTransactionEntry.LoanTransactionID,
		}, userOrg)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to balance loan transaction: %v", err)})
		}
		return ctx.JSON(http.StatusOK, map[string]string{"message": "Loan transaction entry deleted successfully"})
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/loan-transaction-entry/:loan_transaction_entry_id/restore",
		Method: "PUT",
		Note:   "Restores a deleted automatic loan deduction entry by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionEntryID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction entry ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to restore loan transaction entries"})
		}

		loanTransactionEntry, err := core.LoanTransactionEntryManager(service).GetByID(context, *loanTransactionEntryID)
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

		if err := core.LoanTransactionEntryManager(service).UpdateByID(context, loanTransactionEntry.ID, loanTransactionEntry); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to restore loan transaction entry: " + err.Error()})
		}

		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + endTx(tx.Error).Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		loanTransaction, err := event.LoanBalancing(context, service, tx, endTx, event.LoanBalanceEvent{
			LoanTransactionID: loanTransactionEntry.LoanTransactionID,
		}, userOrg)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to balance loan transaction: %v", err)})
		}

		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModel(loanTransaction))
	})
}
