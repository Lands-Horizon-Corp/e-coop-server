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

// LoanRelease processes loan release with comprehensive validations and database transactions.
// This function handles the complete loan release workflow including authentication, validation,
// ledger entries creation, interest calculations, and final transaction updates.
// Returns the updated LoanTransaction after successful release.
func (e *Event) LoanRelease(context context.Context, ctx echo.Context, loanTransactionID uuid.UUID) (*core.LoanTransaction, error) {
	// Start database transaction for atomic operations
	tx, endTx := e.provider.Service.Database.StartTransaction(context)
	// ================================================================================
	// STEP 1: USER AUTHENTICATION AND AUTHORIZATION VALIDATION
	// ================================================================================
	// Retrieve and validate current user organization with proper permissions
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(context, ctx)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "authentication-failed",
			Description: "Unable to retrieve user organization details for loan release operation: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get user organization"))
	}
	// Setup timestamp variables for consistent time tracking
	now := time.Now().UTC()
	currentTime := userOrg.UserOrgTime()

	// Validate user organization has required branch assignment
	if userOrg.BranchID == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "validation-failed",
			Description: "User organization is missing required branch assignment for loan operations",
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("invalid user organization data"))
	}

	// ================================================================================
	// STEP 2: LOAN TRANSACTION AND CURRENCY DATA RETRIEVAL
	// ================================================================================
	// Fetch loan transaction with complete account and currency relationship data
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, loanTransactionID, "Account", "Account.Currency")
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "loan-data-retrieval-failed",
			Description: "Unable to retrieve loan transaction details for release: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction"))
	}

	// Extract and validate currency information from loan account
	loanAccountCurrency := loanTransaction.Account.Currency
	if loanAccountCurrency == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "currency-validation-failed",
			Description: "Missing currency information for loan account " + loanTransaction.AccountID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("currency data is nil"))
	}

	// ================================================================================
	// STEP 3: ACTIVE TRANSACTION BATCH VALIDATION
	// ================================================================================
	// Retrieve and validate active transaction batch required for loan operations
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
	// ================================================================================
	// STEP 4: MEMBER PROFILE DATA RETRIEVAL AND VALIDATION
	// ================================================================================
	// Retrieve member profile associated with the loan transaction
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

	// ================================================================================
	// STEP 5: LOAN TRANSACTION ENTRIES PROCESSING AND LEDGER UPDATES
	// ================================================================================
	// Retrieve all loan transaction entries for processing automatic deductions
	loanTransactionEntries, err := e.core.LoanTransactionEntryManager.Find(context, &core.LoanTransactionEntry{
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
			entry.Debit += addOnEntry.Debit
		}
	}
	loanTransactionEntries = filteredEntries

	// Process each loan transaction entry for ledger updates
	for _, entry := range loanTransactionEntries {
		// Skip entries marked as deleted automatic loan deductions
		if entry.IsAutomaticLoanDeductionDeleted {
			continue
		}
		// Retrieve account history for the transaction entry at the specific time
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

		// Convert account history to account model for processing
		account := e.core.AccountHistoryToModel(accountHistory)
		if accountHistory == nil {
			return nil, endTx(eris.New("account history not found for entry"))
		}

		// Process member ledger accounts differently from subsidiary accounts

		// Load DefaultPaymentType if not already loaded
		if account.DefaultPaymentType == nil && account.DefaultPaymentTypeID != nil {
			paymentType, err := e.core.PaymentTypeManager.GetByID(context, *account.DefaultPaymentTypeID)
			if err != nil {
				return nil, endTx(eris.Wrap(err, "failed to retrieve payment type"))
			}
			account.DefaultPaymentType = paymentType
		}

		var typeOfPaymentType core.TypeOfPaymentType
		if account.DefaultPaymentType != nil {
			typeOfPaymentType = account.DefaultPaymentType.Type
		}
		// Create new general ledger entry for member account
		memberLedgerEntry := &core.GeneralLedger{
			CreatedAt:                  now,
			CreatedByID:                userOrg.UserID,
			UpdatedAt:                  now,
			UpdatedByID:                userOrg.UserID,
			BranchID:                   *userOrg.BranchID,
			OrganizationID:             userOrg.OrganizationID,
			TransactionBatchID:         &transactionBatch.ID,
			ReferenceNumber:            loanTransaction.Voucher,
			EntryDate:                  currentTime,
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

			Account: account,
		}

		// Save member ledger entry to database
		if err := e.core.CreateGeneralLedgerEntry(context, tx, memberLedgerEntry); err != nil {
			e.Footstep(ctx, FootstepEvent{
				Activity:    "member-ledger-creation-failed",
				Description: "Failed to create member ledger entry for account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
				Module:      "Loan Release",
			})
			return nil, endTx(eris.Wrap(err, "failed to create member ledger entry"))
		}

	}

	// ================================================================================
	// STEP 6: INTEREST ACCOUNT PROCESSING FOR STRAIGHT COMPUTATION
	// ================================================================================

	// Retrieve loan-related accounts for interest calculations
	loanRelatedAccounts, err := e.core.GetAccountHistoriesByFiltersAtTime(
		context,
		loanTransaction.OrganizationID,
		loanTransaction.BranchID,
		&currentTime,
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
		// Get account history at the specific transaction time
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
		if err := e.core.LoanAccountManager.CreateWithTx(context, tx, &core.LoanAccount{
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

	// ================================================================================
	// STEP 7: LOAN TRANSACTION FINALIZATION AND STATUS UPDATE
	// ================================================================================
	// Update loan transaction with release information and timestamps
	loanTransaction.ReleasedDate = &currentTime
	loanTransaction.ReleasedByID = &userOrg.UserID
	loanTransaction.UpdatedAt = now
	loanTransaction.Count++
	loanTransaction.TransactionBatchID = &transactionBatch.ID
	loanTransaction.UpdatedByID = userOrg.UserID

	// Save updated loan transaction to database
	if err := e.core.LoanTransactionManager.UpdateByIDWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
		return nil, endTx(eris.Wrap(err, "failed to update loan transaction"))
	}

	// ================================================================================
	// STEP 8: DATABASE TRANSACTION COMMIT
	// ================================================================================
	// Commit all changes to the database atomically
	if err := endTx(nil); err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "database-commit-failed",
			Description: "Unable to commit loan release transaction to database: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to commit transaction"))
	}
	// ================================================================================
	// STEP 9: FINAL LOAN TRANSACTION RETRIEVAL AND RESPONSE
	// ================================================================================
	// Retrieve and return the updated loan transaction with all relationships
	updatedloanTransaction, err := e.core.LoanTransactionManager.GetByID(context, loanTransaction.ID)
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
