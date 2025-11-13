package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

// LoanRelease processes loan release with necessary validations and commits the transaction.
// Returns the updated LoanTransaction after successful release.
func (e *Event) LoanRelease(context context.Context, ctx echo.Context, loanTransactionID uuid.UUID) (*core.LoanTransaction, error) {
	tx, endTx := e.provider.Service.Database.StartTransaction(context)

	// ================================================================================
	// STEP 1: AUTHENTICATION AND AUTHORIZATION
	// ================================================================================
	// Retrieve current user organization and validate permissions
	currentUserOrg, err := e.userOrganizationToken.CurrentUserOrganization(context, ctx)
	if err != nil {

		e.Footstep(ctx, FootstepEvent{
			Activity:    "authentication-failed",
			Description: "Unable to retrieve user organization details for loan release operation: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get user organization"))
	}

	now := time.Now().UTC()
	timeNow := time.Now().UTC()
	if currentUserOrg.TimeMachineTime != nil {
		timeNow = currentUserOrg.UserOrgTime()
	}

	// Validate branch assignment
	if currentUserOrg.BranchID == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "validation-failed",
			Description: "User organization is missing required branch assignment for loan operations",
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("invalid user organization data"))
	}

	// Validate user permissions for loan release
	if currentUserOrg.UserType != core.UserOrganizationTypeOwner && currentUserOrg.UserType != core.UserOrganizationTypeEmployee {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "authorization-failed",
			Description: "User does not have sufficient permissions to perform loan release operations",
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("unauthorized user role"))
	}

	// ================================================================================
	// STEP 2: TRANSACTION BATCH VALIDATION
	// ================================================================================
	// Ensure there's an active transaction batch for recording the loan release
	activeBatch, err := e.core.TransactionBatchCurrent(context, currentUserOrg.UserID, currentUserOrg.OrganizationID, *currentUserOrg.BranchID)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "batch-retrieval-failed",
			Description: "Unable to retrieve active transaction batch for user " + currentUserOrg.UserID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve transaction batch"))
	}

	if activeBatch == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "batch-validation-failed",
			Description: "No active transaction batch found - batch is required for loan release operations",
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("transaction batch is nil"))
	}

	// ================================================================================
	// STEP 3: LOAN TRANSACTION DATA RETRIEVAL AND VALIDATION
	// ================================================================================

	// Fetch the loan transaction with account and currency details
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, loanTransactionID, "Account", "Account.Currency")
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "loan-data-retrieval-failed",
			Description: "Unable to retrieve loan transaction details for release: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction"))
	}
	// Validate currency information
	loanCurrency := loanTransaction.Account.Currency
	if loanCurrency == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "currency-validation-failed",
			Description: "Missing currency information for loan account " + loanTransaction.AccountID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("currency data is nil"))
	}

	// ================================================================================
	// STEP 4: CASH ON HAND ACCOUNT PROCESSING
	// ================================================================================
	// Retrieve the cash on hand account for the loan release
	cashAccount, err := e.core.GetCashOnCashEquivalence(
		context, loanTransaction.ID, currentUserOrg.OrganizationID, *currentUserOrg.BranchID)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to get cash on hand account"))
	}

	// Lock the subsidiary ledger for the cash account
	cashAccountLedger, err := e.core.GeneralLedgerCurrentSubsidiaryAccountForUpdate(
		context, tx, cashAccount.ID, currentUserOrg.OrganizationID, *currentUserOrg.BranchID)
	if err != nil {

		e.Footstep(ctx, FootstepEvent{
			Activity:    "cash-ledger-lock-failed",
			Description: "Unable to acquire lock on cash account subsidiary ledger " + cashAccount.ID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve subsidiary general ledger"))
	}

	// Determine current cash balance
	var currentCashBalance float64 = 0
	if cashAccountLedger == nil {

		e.Footstep(ctx, FootstepEvent{
			Activity:    "cash-ledger-initialization",
			Description: "Initializing new cash account ledger for " + cashAccount.Account.ID.String() + " with zero balance",
			Module:      "Loan Release",
		})
	} else {
		currentCashBalance = cashAccountLedger.Balance
	}

	// ================================================================================
	// STEP 5: CASH ACCOUNT BALANCE CALCULATION AND LEDGER ENTRY
	// ================================================================================
	cashDebit, cashCredit, newCashBalance := e.usecase.Adjustment(*cashAccount.Account, loanTransaction.Balance, 0.0, currentCashBalance)
	userOrgTime := currentUserOrg.UserOrgTime()
	cashLedgerEntry := &core.GeneralLedger{
		CreatedAt:                  now,
		CreatedByID:                currentUserOrg.UserID,
		UpdatedAt:                  now,
		UpdatedByID:                currentUserOrg.UserID,
		BranchID:                   *currentUserOrg.BranchID,
		OrganizationID:             currentUserOrg.OrganizationID,
		TransactionBatchID:         &activeBatch.ID,
		ReferenceNumber:            loanTransaction.Voucher,
		EntryDate:                  &userOrgTime,
		AccountID:                  &cashAccount.Account.ID,
		PaymentTypeID:              cashAccount.Account.DefaultPaymentTypeID,
		TransactionReferenceNumber: loanTransaction.Voucher,
		Source:                     core.GeneralLedgerSourceCheckVoucher,
		BankReferenceNumber:        "",
		EmployeeUserID:             &currentUserOrg.UserID,
		Description:                cashAccount.Description,
		TypeOfPaymentType:          cashAccount.Account.DefaultPaymentType.Type,
		Credit:                     cashCredit,
		Debit:                      cashDebit,
		Balance:                    newCashBalance,
		CurrencyID:                 &loanCurrency.ID,
		LoanTransactionID:          &loanTransaction.ID,
	}

	if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, cashLedgerEntry); err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "cash-ledger-creation-failed",
			Description: "Unable to create cash account ledger entry for " + cashAccount.Account.ID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to create general ledger entry"))
	}

	e.Footstep(ctx, FootstepEvent{
		Activity:    "cash-transaction-completed",
		Description: "Successfully updated cash account " + cashAccount.Account.ID.String() + " with new balance: " + fmt.Sprintf("%.2f", newCashBalance),
		Module:      "Loan Release",
	})

	// ================================================================================
	// STEP 6: MEMBER ACCOUNT LEDGER PROCESSING
	// ================================================================================
	// Retrieve member profile associated with the loan
	memberProfile, err := e.core.MemberProfileManager.GetByID(context, *loanTransaction.MemberProfileID)
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
	// Lock member's general ledger for the loan account
	memberAccountLedger, err := e.core.GeneralLedgerCurrentMemberAccountForUpdate(
		context, tx,
		memberProfile.ID,
		*loanTransaction.AccountID,
		memberProfile.OrganizationID,
		memberProfile.BranchID,
	)
	if err != nil {

		e.Footstep(ctx, FootstepEvent{
			Activity:    "member-ledger-lock-failed",
			Description: "Unable to acquire lock on member ledger for account " + loanTransaction.AccountID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve member ledger for update"))
	}

	// Determine current member account balance
	var currentMemberBalance float64 = 0
	if memberAccountLedger == nil {

		e.Footstep(ctx, FootstepEvent{
			Activity:    "member-ledger-initialization",
			Description: "Initializing new member ledger for account " + loanTransaction.AccountID.String() + " and member " + memberProfile.ID.String(),
			Module:      "Loan Release",
		})
	} else {
		currentMemberBalance = memberAccountLedger.Balance
	}

	// ================================================================================
	// STEP 7: MEMBER ACCOUNT BALANCE CALCULATION AND LEDGER ENTRY
	// ================================================================================
	memberDebit, memberCredit, newMemberBalance := e.usecase.Adjustment(
		*loanTransaction.Account, 0.0, loanTransaction.Balance, currentMemberBalance)

	memberLedgerEntry := &core.GeneralLedger{
		CreatedAt:                  now,
		CreatedByID:                currentUserOrg.UserID,
		UpdatedAt:                  now,
		UpdatedByID:                currentUserOrg.UserID,
		BranchID:                   *currentUserOrg.BranchID,
		OrganizationID:             currentUserOrg.OrganizationID,
		TransactionBatchID:         &activeBatch.ID,
		ReferenceNumber:            loanTransaction.Voucher,
		EntryDate:                  &userOrgTime,
		AccountID:                  loanTransaction.AccountID,
		MemberProfileID:            &memberProfile.ID,
		PaymentTypeID:              cashAccount.Account.DefaultPaymentTypeID,
		TransactionReferenceNumber: loanTransaction.Voucher,
		Source:                     core.GeneralLedgerSourceCheckVoucher,
		EmployeeUserID:             &currentUserOrg.UserID,
		Description:                loanTransaction.Account.Description,
		TypeOfPaymentType:          cashAccount.Account.DefaultPaymentType.Type,
		Credit:                     memberCredit,
		Debit:                      memberDebit,
		Balance:                    newMemberBalance,
		CurrencyID:                 &loanCurrency.ID,
		LoanTransactionID:          &loanTransaction.ID,
	}

	// Create the member's general ledger entry in the database
	if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, memberLedgerEntry); err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "member-ledger-creation-failed",
			Description: "Unable to create member ledger entry for account " + memberLedgerEntry.AccountID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to create general ledger entry"))
	}

	accounts, err := e.core.GetAccountHistoriesByFiltersAtTime(
		context,
		loanTransaction.OrganizationID,
		*currentUserOrg.BranchID,
		&timeNow,
		loanTransaction.AccountID,
		&loanCurrency.ID,
	)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "data-retrieval-failed",
			Description: "Failed to retrieve loan-related accounts for amortization schedule: " + err.Error(),
			Module:      "Loan Amortization",
		})
		return nil, eris.Wrapf(err, "failed to retrieve accounts for loan transaction ID: %s", loanTransaction.ID.String())
	}
	// ================================================================================
	// STEP 8: MEMBER ACCOUNTING LEDGER UPDATE
	// ================================================================================
	// Update or create member accounting ledger with new balance
	_, err = e.core.MemberAccountingLedgerUpdateOrCreate(
		context,
		tx,
		*loanTransaction.MemberProfileID,
		*loanTransaction.AccountID,
		currentUserOrg.OrganizationID,
		*currentUserOrg.BranchID,
		currentUserOrg.UserID,
		newMemberBalance,
		now,
	)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "member-accounting-ledger-update-failed",
			Description: "Unable to update member accounting ledger for member " + loanTransaction.MemberProfileID.String() + " on account " + loanTransaction.AccountID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to update member accounting ledger"))
	}

	// ================================================================================
	// STEP 8: PROCESS INTEREST ACCOUNTS (STRAIGHT COMPUTATION ONLY)
	// ================================================================================

	// Process all related accounts that have straight interest computation
	for _, account := range accounts {
		// Skip accounts that are not loan-related or don't use straight computation
		if account.LoanAccountID == nil ||
			account.ComputationType != core.Straight ||
			(account.Type == core.AccountTypeLoan || account.Type == core.AccountTypeFines) {
			continue
		}

		// Lock member's general ledger for this interest account
		intLedger, err := e.core.GeneralLedgerCurrentMemberAccountForUpdate(
			context, tx,
			memberProfile.ID,
			account.ID,
			memberProfile.OrganizationID,
			memberProfile.BranchID,
		)
		if err != nil {
			e.Footstep(ctx, FootstepEvent{
				Activity:    "interest-ledger-lock-failed",
				Description: "Failed to lock interest account ledger for account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
				Module:      "Loan Release",
			})
			return nil, endTx(eris.Wrap(err, "failed to lock interest account ledger"))
		}

		// Get current balance for this interest account
		var intBalance float64 = 0
		if intLedger == nil {
			e.Footstep(ctx, FootstepEvent{
				Activity:    "interest-ledger-initialization",
				Description: "Initializing new interest ledger for account " + account.ID.String() + " and member " + memberProfile.ID.String(),
				Module:      "Loan Release",
			})
		} else {
			intBalance = intLedger.Balance
		}

		// Calculate straight interest for this account
		straightBalance := e.usecase.ComputeInterestStraight(
			loanTransaction.TotalPrincipal, account.InterestStandard, loanTransaction.Terms)

		// Calculate adjustment amounts for member interest account
		intDebit, intCredit, newIntBalance := e.usecase.Adjustment(
			*loanTransaction.Account, 0.0, straightBalance, intBalance)

		// Create member's general ledger entry for interest account
		intMemberEntry := &core.GeneralLedger{
			CreatedAt:                  now,
			CreatedByID:                currentUserOrg.UserID,
			UpdatedAt:                  now,
			UpdatedByID:                currentUserOrg.UserID,
			BranchID:                   *currentUserOrg.BranchID,
			OrganizationID:             currentUserOrg.OrganizationID,
			TransactionBatchID:         &activeBatch.ID,
			ReferenceNumber:            loanTransaction.Voucher,
			EntryDate:                  &userOrgTime,
			AccountID:                  &account.ID,
			MemberProfileID:            &memberProfile.ID,
			PaymentTypeID:              cashAccount.Account.DefaultPaymentTypeID,
			TransactionReferenceNumber: loanTransaction.Voucher,
			Source:                     core.GeneralLedgerSourceCheckVoucher,
			EmployeeUserID:             &currentUserOrg.UserID,
			Description:                loanTransaction.Account.Description,
			TypeOfPaymentType:          cashAccount.Account.DefaultPaymentType.Type,
			Credit:                     intCredit,
			Debit:                      intDebit,
			Balance:                    newIntBalance,
			CurrencyID:                 &loanCurrency.ID,
			LoanTransactionID:          &loanTransaction.ID,
		}

		// Save member interest ledger entry to database
		if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, intMemberEntry); err != nil {
			e.Footstep(ctx, FootstepEvent{
				Activity:    "interest-member-ledger-failed",
				Description: "Failed to create interest member ledger entry for account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
				Module:      "Loan Release",
			})
			return nil, endTx(eris.Wrap(err, "failed to create interest member ledger entry"))
		}

		// Update member accounting ledger for this interest account
		_, err = e.core.MemberAccountingLedgerUpdateOrCreate(
			context,
			tx,
			*loanTransaction.MemberProfileID,
			account.ID,
			currentUserOrg.OrganizationID,
			*currentUserOrg.BranchID,
			currentUserOrg.UserID,
			newIntBalance,
			now,
		)
		if err != nil {
			e.Footstep(ctx, FootstepEvent{
				Activity:    "interest-accounting-ledger-failed",
				Description: "Failed to update member accounting ledger for interest account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
				Module:      "Loan Release",
			})
			return nil, endTx(eris.Wrap(err, "failed to update interest accounting ledger"))
		}

		// ================================================================
		// Process cash account entry for the computed interest amount
		// ================================================================

		// Lock the cash account subsidiary ledger for interest processing
		intCashLedger, err := e.core.GeneralLedgerCurrentSubsidiaryAccountForUpdate(
			context, tx, cashAccount.ID, currentUserOrg.OrganizationID, *currentUserOrg.BranchID)
		if err != nil {
			e.Footstep(ctx, FootstepEvent{
				Activity:    "interest-cash-lock-failed",
				Description: "Failed to lock cash account subsidiary ledger for interest processing " + cashAccount.ID.String(),
				Module:      "Loan Release",
			})
			return nil, endTx(eris.Wrap(err, "failed to lock cash account for interest"))
		}

		// Get current cash balance for interest processing
		var intCashBalance float64 = 0
		if intCashLedger == nil {
			e.Footstep(ctx, FootstepEvent{
				Activity:    "interest-cash-initialization",
				Description: "Initializing cash account ledger for interest processing " + cashAccount.Account.ID.String(),
				Module:      "Loan Release",
			})
		} else {
			intCashBalance = intCashLedger.Balance
		}

		// Calculate cash adjustment for interest amount
		intCashDebit, intCashCredit, newIntCashBalance := e.usecase.Adjustment(
			*cashAccount.Account, straightBalance, 0.0, intCashBalance)

		// Create cash account ledger entry for interest
		intCashEntry := &core.GeneralLedger{
			CreatedAt:                  now,
			CreatedByID:                currentUserOrg.UserID,
			UpdatedAt:                  now,
			UpdatedByID:                currentUserOrg.UserID,
			BranchID:                   *currentUserOrg.BranchID,
			OrganizationID:             currentUserOrg.OrganizationID,
			TransactionBatchID:         &activeBatch.ID,
			ReferenceNumber:            loanTransaction.Voucher,
			EntryDate:                  &userOrgTime,
			AccountID:                  &cashAccount.Account.ID,
			PaymentTypeID:              cashAccount.Account.DefaultPaymentTypeID,
			TransactionReferenceNumber: loanTransaction.Voucher,
			Source:                     core.GeneralLedgerSourceCheckVoucher,
			BankReferenceNumber:        "",
			EmployeeUserID:             &currentUserOrg.UserID,
			Description:                cashAccount.Description,
			TypeOfPaymentType:          cashAccount.Account.DefaultPaymentType.Type,
			Credit:                     intCashCredit,
			Debit:                      intCashDebit,
			Balance:                    newIntCashBalance,
			CurrencyID:                 &loanCurrency.ID,
			LoanTransactionID:          &loanTransaction.ID,
		}

		// Save cash interest ledger entry to database
		if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, intCashEntry); err != nil {
			e.Footstep(ctx, FootstepEvent{
				Activity:    "interest-cash-ledger-failed",
				Description: "Failed to create cash interest ledger entry for " + cashAccount.Account.ID.String(),
				Module:      "Loan Release",
			})
			return nil, endTx(eris.Wrap(err, "failed to create cash interest ledger entry"))
		}

		// Log successful interest processing for this account
		e.Footstep(ctx, FootstepEvent{
			Activity:    "interest-account-processed",
			Description: "Successfully processed interest account " + account.ID.String() + " with interest: " + fmt.Sprintf("%.2f", straightBalance),
			Module:      "Loan Release",
		})
	}

	// ================================================================================
	// STEP 9: LOAN TRANSACTION FINALIZATION
	// ================================================================================
	// Update loan transaction with release information
	loanTransaction.ReleasedDate = &timeNow
	loanTransaction.ReleasedByID = &currentUserOrg.UserID
	loanTransaction.UpdatedAt = now
	loanTransaction.LoanCount = loanTransaction.LoanCount + 1
	loanTransaction.UpdatedByID = currentUserOrg.UserID

	if err := e.core.LoanTransactionManager.UpdateByIDWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
		return nil, endTx(eris.Wrap(err, "failed to update loan transaction"))
	}

	// ================================================================================
	// STEP 10: DATABASE TRANSACTION COMMIT
	// ================================================================================
	// Commit all changes to the database
	if err := endTx(nil); err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "database-commit-failed",
			Description: "Unable to commit loan release transaction to database: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to commit transaction"))
	}

	// ================================================================================
	// STEP 11: FINAL TRANSACTION RETRIEVAL AND RETURN
	// ================================================================================
	// Retrieve and return the updated loan transaction
	updatedLoanTransaction, err := e.core.LoanTransactionManager.GetByID(context, loanTransaction.ID)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "final-retrieval-failed",
			Description: "Unable to retrieve updated loan transaction after successful release: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get updated loan transaction"))
	}
	return updatedLoanTransaction, nil
}
