package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (e *Event) LoanAdjustment(
	context context.Context,
	userOrg core.UserOrganization,
	loanTransactionID uuid.UUID,
	la core.LoanTransactionAdjustmentRequest,
) error {

	// ========================================
	// STEP 1: Initialize transaction and dates
	// ========================================
	tx, endTx := e.provider.Service.Database.StartTransaction(context)
	now := time.Now().UTC()

	currentDate := time.Now().UTC()
	if userOrg.TimeMachineTime != nil {
		currentDate = userOrg.UserOrgTime()
	}

	// ========================================
	// STEP 2: Validate loan transaction
	// ========================================
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, loanTransactionID)
	if err != nil {
		return endTx(eris.Wrap(err, "Loan transaction not found"))
	}

	if loanTransaction.Processing {
		return endTx(eris.New("Cannot adjust loan while transaction is being processed"))
	}
	// ========================================
	// STEP 4: Validate account information
	// ========================================
	account, err := e.core.AccountManager.GetByID(context, la.AccountID, "Currency")
	if err != nil {
		return endTx(eris.Wrap(err, "Account not found for adjustment"))
	}

	accountHistory, err := e.core.GetAccountHistoryLatestByTimeHistory(
		context,
		account.ID,
		account.OrganizationID,
		account.BranchID,
		loanTransaction.PrintedDate,
	)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to retrieve account history"))
	}
	if accountHistory != nil {
		account = e.core.AccountHistoryToModel(accountHistory)
	}

	// ========================================
	// STEP 5: Calculate adjustment amounts
	// ========================================
	memberDebit, memberCredit := 0.0, 0.0
	switch la.AdjustmentType {
	case core.LoanAdjustmentTypeAdd:
		memberCredit = la.Amount
	case core.LoanAdjustmentTypeDeduct:
		memberDebit = la.Amount
	default:
		return endTx(eris.New("Invalid adjustment type specified"))
	}

	// ========================================
	// STEP 6: Create general ledger entry
	// ========================================
	memberLedgerEntry := &core.GeneralLedger{
		CreatedAt:                  now,
		CreatedByID:                userOrg.UserID,
		UpdatedAt:                  now,
		UpdatedByID:                userOrg.UserID,
		BranchID:                   *userOrg.BranchID,
		OrganizationID:             userOrg.OrganizationID,
		ReferenceNumber:            loanTransaction.Voucher,
		EntryDate:                  &currentDate,
		AccountID:                  &account.ID,
		MemberProfileID:            loanTransaction.MemberProfileID,
		PaymentTypeID:              account.DefaultPaymentTypeID,
		TransactionReferenceNumber: loanTransaction.Voucher,
		Source:                     core.GeneralLedgerSourceCheckVoucher,
		EmployeeUserID:             &userOrg.UserID,
		Description:                account.Description,
		Credit:                     memberCredit,
		Debit:                      memberDebit,
		CurrencyID:                 account.CurrencyID,
		LoanTransactionID:          &loanTransaction.ID,
		LoanAdjustmentType:         &la.AdjustmentType,
	}

	if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, memberLedgerEntry); err != nil {
		return endTx(eris.Wrap(err, "Failed to record adjustment in general ledger"))
	}

	// ========================================
	// STEP 7: Update member accounting ledger
	// ========================================
	_, err = e.core.MemberAccountingLedgerUpdateOrCreate(
		context,
		tx,
		core.MemberAccountingLedgerUpdateOrCreateParams{
			MemberProfileID: *loanTransaction.MemberProfileID,
			AccountID:       account.ID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			UserID:          userOrg.UserID,
			DebitAmount:     memberDebit,
			CreditAmount:    memberCredit,
			LastPayTime:     now,
		},
	)
	if err != nil {
		return endTx(eris.Wrap(err, "Failed to update member account balance"))
	}

	// ========================================
	// STEP 8: Commit transaction
	// ========================================
	if err := endTx(nil); err != nil {
		return eris.Wrap(err, "Failed to save loan adjustment changes")
	}

	return nil
}
