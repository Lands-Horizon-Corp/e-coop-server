package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// LoanRelease performs the necessary checks and commits to release a loan
// and returns the updated LoanTransaction.
func (e *Event) LoanRelease(ctx context.Context, echoCtx echo.Context, tx *gorm.DB, endTx func(error) error, data LoanBalanceEvent) (*core.LoanTransaction, error) {
	// ================================================================================
	// STEP 1: AUTHENTICATION & USER ORGANIZATION RETRIEVAL
	// ================================================================================
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(ctx, echoCtx)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, endTx(eris.Wrap(err, "failed to get user organization"))
	}
	if userOrg.BranchID == nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Invalid user organization data (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		return nil, endTx(eris.New("invalid user organization data"))
	}
	if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Unauthorized user role (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		return nil, endTx(eris.New("unauthorized user role"))
	}
	// ================================================================================
	// STEP 2: LOAN TRANSACTION & RELATED DATA RETRIEVAL
	// ================================================================================
	// Get the main loan transaction
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, data.LoanTransactionID)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction"))
	}

	for _, entry := range loanTransaction.LoanTransactionEntries {
		// Computation of all ammortization accounts
		if entry.Type == core.LoanTransactionPrevious {
			return nil, endTx(eris.New("cannot release a restructured or renewed loan"))
		}
		accountHistory, err := e.core.GetAccountHistoryLatestByTime(
			ctx, *entry.AccountID, userOrg.OrganizationID, *userOrg.BranchID, loanTransaction.UpdatedAt)
		if err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to get account history (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return nil, endTx(eris.Wrap(err, "failed to get account history"))
		}
		entry.Account = e.core.AccountHistoryToModel(accountHistory)
	}

	// ================================================================================
	if err := endTx(nil); err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "db-commit-error",
			Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, endTx(eris.Wrap(err, "failed to commit transaction"))
	}
	newLoanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, loanTransaction.ID)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get updated loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, endTx(eris.Wrap(err, "failed to get updated loan transaction"))
	}
	return newLoanTransaction, nil
}
