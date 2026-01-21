package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func LoanRelease(context context.Context, service *horizon.HorizonService, loanTransactionID uuid.UUID, userOrg *types.UserOrganization) (*types.LoanTransaction, error) {

	tx, endTx := service.Database.StartTransaction(context)

	now := time.Now().UTC()
	timeMachine := userOrg.TimeMachine()
	if userOrg.BranchID == nil {
		return nil, endTx(eris.New("invalid user organization data"))
	}

	loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, loanTransactionID, "Account", "Account.Currency")
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction"))
	}

	loanAccountCurrency := loanTransaction.Account.Currency

	if loanAccountCurrency == nil {
		return nil, endTx(eris.New("currency data is nil"))
	}

	transactionBatch, err := core.TransactionBatchCurrent(context, service, *loanTransaction.EmployeeUserID, userOrg.OrganizationID, *userOrg.BranchID)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to retrieve transaction batch - The one who created the loan must have created the transaction batch"))
	}

	if transactionBatch == nil {
		return nil, endTx(eris.New("transaction batch is nil"))
	}

	memberProfile, err := core.MemberProfileManager(service).GetByID(context, *loanTransaction.MemberProfileID)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to retrieve member profile"))
	}

	if memberProfile == nil {
		return nil, endTx(eris.New("member profile not found"))
	}

	loanTransactionEntries, err := core.LoanTransactionEntryManager(service).Find(context, &types.LoanTransactionEntry{
		LoanTransactionID: loanTransaction.ID,
		OrganizationID:    loanTransaction.OrganizationID,
		BranchID:          loanTransaction.BranchID,
	})
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to retrieve loan transaction entries"))
	}

	var addOnEntry *types.LoanTransactionEntry
	var filteredEntries []*types.LoanTransactionEntry

	for _, entry := range loanTransactionEntries {
		if entry.Type == types.LoanTransactionAddOn {
			addOnEntry = entry
		} else {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	for _, entry := range filteredEntries {
		if entry.Type == types.LoanTransactionStatic && helpers.UUIDPtrEqual(entry.AccountID, loanTransaction.AccountID) {
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

		accountHistory, err := core.GetAccountHistoryLatestByTimeHistory(
			context, service,
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

		account := core.AccountHistoryToModel(accountHistory)

		if account.DefaultPaymentType == nil && account.DefaultPaymentTypeID != nil {
			paymentType, err := core.PaymentTypeManager(service).GetByID(context, *account.DefaultPaymentTypeID)
			if err != nil {
				return nil, endTx(eris.Wrap(err, "failed to retrieve payment type"))
			}
			account.DefaultPaymentType = paymentType

		}

		var typeOfPaymentType types.TypeOfPaymentType
		if account.DefaultPaymentType != nil {
			typeOfPaymentType = account.DefaultPaymentType.Type
		}

		memberLedgerEntry := &types.GeneralLedger{
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
			Source:                     types.GeneralLedgerSourceLoan,
			EmployeeUserID:             &userOrg.UserID,
			Description:                loanTransaction.Account.Description,
			TypeOfPaymentType:          typeOfPaymentType,
			Credit:                     entry.Credit,
			Debit:                      entry.Debit,
			CurrencyID:                 &loanAccountCurrency.ID,
			LoanTransactionID:          &loanTransaction.ID,
			Account:                    account,
		}

		if err := core.CreateGeneralLedgerEntry(context, service, tx, memberLedgerEntry); err != nil {
			return nil, endTx(eris.Wrap(err, "failed to create member ledger entry"))
		}

	}

	loanRelatedAccounts, err := core.GetAccountHistoriesByFiltersAtTime(
		context, service,
		loanTransaction.OrganizationID,
		loanTransaction.BranchID,
		&timeMachine,
		loanTransaction.AccountID,
		&loanAccountCurrency.ID,
	)
	if err != nil {
		return nil, endTx(eris.Wrapf(err, "failed to retrieve accounts for loan transaction ID: %s", loanTransaction.ID.String()))
	}

	loanRelatedAccounts = append(loanRelatedAccounts, loanTransaction.Account)

	for _, interestAccount := range loanRelatedAccounts {

		interestAccountHistory, err := core.GetAccountHistoryLatestByTimeHistory(
			context, service,
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

		if err := core.LoanAccountManager(service).CreateWithTx(context, tx, &types.LoanAccount{
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

	if err := core.LoanTransactionManager(service).UpdateByIDWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
		return nil, endTx(eris.Wrap(err, "failed to update loan transaction"))
	}

	if err := endTx(nil); err != nil {
		return nil, endTx(eris.Wrap(err, "failed to commit transaction"))
	}

	updatedloanTransaction, err := core.LoanTransactionManager(service).GetByID(context, loanTransaction.ID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get updated loan transaction")
	}

	return updatedloanTransaction, nil
}
