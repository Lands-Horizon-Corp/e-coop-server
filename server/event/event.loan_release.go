package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func (e *Event) LoanRelease(context context.Context, ctx echo.Context, loanTransactionID uuid.UUID) (*core.LoanTransaction, error) {

	tx, endTx := e.provider.Service.Database.StartTransaction(context)

	userOrg, err := e.CurrentUserOrganization(context, ctx)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "authentication-failed",
			Description: "Unable to retrieve user organization details for loan release operation: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get user organization"))
	}

	now := time.Now().UTC()
	timeMachine := userOrg.UserOrgTime()
	if userOrg.BranchID == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "validation-failed",
			Description: "User organization is missing required branch assignment for loan operations",
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("invalid user organization data"))
	}

	loanTransaction, err := e.core.LoanTransactionManager().GetByID(context, loanTransactionID, "Account", "Account.Currency")
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "loan-data-retrieval-failed",
			Description: "Unable to retrieve loan transaction details for release: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction"))
	}

	loanAccountCurrency := loanTransaction.Account.Currency

	if loanAccountCurrency == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "currency-validation-failed",
			Description: "Missing currency information for loan account " + loanTransaction.AccountID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("currency data is nil"))
	}

	transactionBatch, err := e.core.TransactionBatchCurrent(context, *loanTransaction.EmployeeUserID, userOrg.OrganizationID, *userOrg.BranchID)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "batch-retrieval-failed",
			Description: "Unable to retrieve active transaction batch for user " + userOrg.UserID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve transaction batch - The one who created the loan must have created the transaction batch"))
	}

	if transactionBatch == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "batch-validation-failed",
			Description: "No active transaction batch found - batch is required for loan release operations",
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("transaction batch is nil"))
	}

	memberProfile, err := e.core.MemberProfileManager().GetByID(context, *loanTransaction.MemberProfileID)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "member-profile-retrieval-failed",
			Description: "Unable to retrieve member profile " + loanTransaction.MemberProfileID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve member profile"))
	}

	if memberProfile == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "member-profile-not-found",
			Description: "Member profile does not exist for ID: " + loanTransaction.MemberProfileID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("member profile not found"))
	}

	loanTransactionEntries, err := e.core.LoanTransactionEntryManager().Find(context, &core.LoanTransactionEntry{
		LoanTransactionID: loanTransaction.ID,
		OrganizationID:    loanTransaction.OrganizationID,
		BranchID:          loanTransaction.BranchID,
	})
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to retrieve loan transaction entries"))
	}

	var addOnEntry *core.LoanTransactionEntry
	var filteredEntries []*core.LoanTransactionEntry

	for _, entry := range loanTransactionEntries {
		if entry.Type == core.LoanTransactionAddOn {
			addOnEntry = entry
		} else {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	for _, entry := range filteredEntries {
		if entry.Type == core.LoanTransactionStatic && handlers.UUIDPtrEqual(entry.AccountID, loanTransaction.AccountID) {
			if addOnEntry != nil {
				entry.Debit += addOnEntry.Debit
			}
		}
	}

	loanTransactionEntries = filteredEntries

	for _, entry := range loanTransactionEntries {

		if entry.IsAutomaticLoanDeductionDeleted {
			continue
		}

		if entry.AccountID == nil {

			return nil, endTx(eris.New("entry.AccountID is nil"))
		}

		accountHistory, err := e.core.GetAccountHistoryLatestByTimeHistory(
			context,
			*entry.AccountID,
			entry.OrganizationID,
			entry.BranchID,
			loanTransaction.ReleasedDate,
		)
		if err != nil {
			return nil, endTx(eris.Wrap(err, "failed to retrieve account history"))
		}

		if accountHistory == nil {
			return nil, endTx(eris.New("account history not found for entry"))
		}

		account := e.core.AccountHistoryToModel(accountHistory)

		if account.DefaultPaymentType == nil && account.DefaultPaymentTypeID != nil {
			paymentType, err := e.core.PaymentTypeManager().GetByID(context, *account.DefaultPaymentTypeID)
			if err != nil {
				return nil, endTx(eris.Wrap(err, "failed to retrieve payment type"))
			}
			account.DefaultPaymentType = paymentType

		}

		var typeOfPaymentType core.TypeOfPaymentType
		if account.DefaultPaymentType != nil {
			typeOfPaymentType = account.DefaultPaymentType.Type
		}

		memberLedgerEntry := &core.GeneralLedger{
			CreatedAt:                  now,
			CreatedByID:                userOrg.UserID,
			UpdatedAt:                  now,
			UpdatedByID:                userOrg.UserID,
			BranchID:                   *userOrg.BranchID,
			OrganizationID:             userOrg.OrganizationID,
			TransactionBatchID:         &transactionBatch.ID,
			ReferenceNumber:            loanTransaction.Voucher,
			EntryDate:                  timeMachine,
			AccountID:                  &account.ID,
			MemberProfileID:            &memberProfile.ID,
			PaymentTypeID:              account.DefaultPaymentTypeID,
			TransactionReferenceNumber: loanTransaction.Voucher,
			Source:                     core.GeneralLedgerSourceLoan,
			EmployeeUserID:             &userOrg.UserID,
			Description:                loanTransaction.Account.Description,
			TypeOfPaymentType:          typeOfPaymentType,
			Credit:                     entry.Credit,
			Debit:                      entry.Debit,
			CurrencyID:                 &loanAccountCurrency.ID,
			LoanTransactionID:          &loanTransaction.ID,
			Account:                    account,
		}

		if err := e.core.CreateGeneralLedgerEntry(context, tx, memberLedgerEntry); err != nil {
			e.Footstep(ctx, FootstepEvent{
				Activity:    "member-ledger-creation-failed",
				Description: "Failed to create member ledger entry for account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
				Module:      "Loan Release",
			})
			return nil, endTx(eris.Wrap(err, "failed to create member ledger entry"))
		}

	}

	loanRelatedAccounts, err := e.core.GetAccountHistoriesByFiltersAtTime(
		context,
		loanTransaction.OrganizationID,
		loanTransaction.BranchID,
		&timeMachine,
		loanTransaction.AccountID,
		&loanAccountCurrency.ID,
	)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "account-retrieval-failed",
			Description: "Failed to retrieve loan-related accounts for interest processing: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrapf(err, "failed to retrieve accounts for loan transaction ID: %s", loanTransaction.ID.String()))
	}

	loanRelatedAccounts = append(loanRelatedAccounts, loanTransaction.Account)

	for _, interestAccount := range loanRelatedAccounts {

		interestAccountHistory, err := e.core.GetAccountHistoryLatestByTimeHistory(
			context,
			interestAccount.ID,
			interestAccount.OrganizationID,
			interestAccount.BranchID,
			loanTransaction.ReleasedDate,
		)
		if err != nil {
			return nil, endTx(eris.Wrap(err, "failed to retrieve interest account history"))
		}

		if interestAccountHistory == nil {
			return nil, endTx(eris.New("interest account history is nil"))
		}

		if err := e.core.LoanAccountManager().CreateWithTx(context, tx, &core.LoanAccount{
			CreatedAt:         now,
			CreatedByID:       userOrg.UserID,
			UpdatedAt:         now,
			UpdatedByID:       userOrg.UserID,
			OrganizationID:    interestAccount.OrganizationID,
			BranchID:          interestAccount.BranchID,
			LoanTransactionID: loanTransaction.ID,
			AccountID:         &interestAccount.ID,
			AccountHistoryID:  &interestAccountHistory.ID,
			Amount:            0.0,
		}); err != nil {
			return nil, endTx(eris.Wrap(err, "failed to create loan account"))
		}

	}

	loanTransaction.ReleasedDate = &timeMachine
	loanTransaction.ReleasedByID = &userOrg.UserID
	loanTransaction.UpdatedAt = now
	loanTransaction.Count++
	loanTransaction.TransactionBatchID = &transactionBatch.ID
	loanTransaction.UpdatedByID = userOrg.UserID

	if err := e.core.LoanTransactionManager().UpdateByIDWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
		return nil, endTx(eris.Wrap(err, "failed to update loan transaction"))
	}

	if err := endTx(nil); err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "database-commit-failed",
			Description: "Unable to commit loan release transaction to database: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to commit transaction"))
	}

	updatedloanTransaction, err := e.core.LoanTransactionManager().GetByID(context, loanTransaction.ID)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "final-retrieval-failed",
			Description: "Unable to retrieve updated loan transaction after successful release: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, eris.Wrap(err, "failed to get updated loan transaction")
	}

	e.Footstep(ctx, FootstepEvent{
		Activity:    "loan-release-completed",
		Description: "Successfully completed loan release for transaction " + updatedloanTransaction.ID.String(),
		Module:      "Loan Release",
	})

	return updatedloanTransaction, nil
}
