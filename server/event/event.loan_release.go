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

	// Validate user has sufficient permissions for loan release operations
	if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "authorization-failed",
			Description: "User does not have sufficient permissions to perform loan release operations",
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("unauthorized user role"))
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
			loanTransaction.PrintedDate,
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

		if account.IsMemberLedger() {
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
			}

			// Save member ledger entry to database
			if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, memberLedgerEntry); err != nil {
				e.Footstep(ctx, FootstepEvent{
					Activity:    "member-ledger-creation-failed",
					Description: "Failed to create member ledger entry for account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
					Module:      "Loan Release",
				})
				return nil, endTx(eris.Wrap(err, "failed to create member ledger entry"))
			}

			// STEP 5: Fix in member ledger update
			_, err = e.core.MemberAccountingLedgerUpdateOrCreate(
				context,
				tx,
				core.MemberAccountingLedgerUpdateOrCreateParams{
					MemberProfileID: *loanTransaction.MemberProfileID,
					AccountID:       account.ID,
					OrganizationID:  userOrg.OrganizationID,
					BranchID:        *userOrg.BranchID,
					UserID:          userOrg.UserID,
					DebitAmount:     entry.Debit,
					CreditAmount:    entry.Credit,
					LastPayTime:     now,
				},
			)
			if err != nil {
				e.Footstep(ctx, FootstepEvent{
					Activity:    "member-accounting-ledger-failed",
					Description: "Failed to update member accounting ledger for account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
					Module:      "Loan Release",
				})
				return nil, endTx(eris.Wrap(err, "failed to update member accounting ledger"))
			}
		} else {
			intMemberEntry := &core.GeneralLedger{
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
			}
			if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, intMemberEntry); err != nil {
				e.Footstep(ctx, FootstepEvent{
					Activity: "interest-member-ledger-failed",
					Description: "Failed to create interest member ledger entry for account " + account.ID.String() +
						" and member " + loanTransaction.MemberProfileID.String() + ": " + err.Error(),
					Module: "Loan Release",
				})
				return nil, endTx(eris.Wrap(err, "failed to create interest member ledger entry"))
			}
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

	// Process each interest-bearing account with straight computation
	for _, interestAccount := range loanRelatedAccounts {
		// Get account history at the specific transaction time
		interestAccountHistory, err := e.core.GetAccountHistoryLatestByTimeHistory(
			context,
			interestAccount.ID,
			interestAccount.OrganizationID,
			interestAccount.BranchID,
			loanTransaction.PrintedDate,
		)
		if err != nil {
			return nil, endTx(eris.Wrap(err, "failed to retrieve interest account history"))
		}

		// Use account history if available, otherwise use current account data
		if interestAccountHistory != nil {
			interestAccount = e.core.AccountHistoryToModel(interestAccountHistory)
		}

		// Filter accounts: only process loan-related accounts with straight computation
		// Skip loan principal accounts and fines accounts
		if interestAccount.LoanAccountID == nil ||
			interestAccount.ComputationType != core.Straight ||
			(interestAccount.Type == core.AccountTypeLoan || interestAccount.Type == core.AccountTypeFines) {
			continue
		}
		// Process member interest accounts

		// Load DefaultPaymentType for interest account if not already loaded
		if interestAccount.DefaultPaymentType == nil && interestAccount.DefaultPaymentTypeID != nil {
			paymentType, err := e.core.PaymentTypeManager.GetByID(context, *interestAccount.DefaultPaymentTypeID)
			if err != nil {
				return nil, endTx(eris.Wrap(err, "failed to retrieve interest account payment type"))
			}
			interestAccount.DefaultPaymentType = paymentType
		}

		var interestTypeOfPaymentType core.TypeOfPaymentType
		if interestAccount.DefaultPaymentType != nil {
			interestTypeOfPaymentType = interestAccount.DefaultPaymentType.Type
		}

		if interestAccount.IsMemberLedger() {
			// Lock member interest account ledger for atomic updates

			straightInterestAmount := e.usecase.ComputeInterestStraight(
				loanTransaction.TotalPrincipal, interestAccount.InterestStandard, loanTransaction.Terms)

			credit := straightInterestAmount
			debit := 0.0
			// Create member interest ledger entry
			memberInterestEntry := &core.GeneralLedger{
				CreatedAt:                  now,
				CreatedByID:                userOrg.UserID,
				UpdatedAt:                  now,
				UpdatedByID:                userOrg.UserID,
				BranchID:                   *userOrg.BranchID,
				OrganizationID:             userOrg.OrganizationID,
				TransactionBatchID:         &transactionBatch.ID,
				ReferenceNumber:            loanTransaction.Voucher,
				EntryDate:                  currentTime,
				AccountID:                  &interestAccount.ID,
				MemberProfileID:            &memberProfile.ID,
				PaymentTypeID:              interestAccount.DefaultPaymentTypeID,
				TransactionReferenceNumber: loanTransaction.Voucher,
				Source:                     core.GeneralLedgerSourceLoan,
				EmployeeUserID:             &userOrg.UserID,
				Description:                loanTransaction.Account.Description,
				TypeOfPaymentType:          interestTypeOfPaymentType,
				Credit:                     credit,
				Debit:                      debit,
				CurrencyID:                 &loanAccountCurrency.ID,
				LoanTransactionID:          &loanTransaction.ID,
			}

			// Save member interest ledger entry to database
			if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, memberInterestEntry); err != nil {
				e.Footstep(ctx, FootstepEvent{
					Activity:    "member-interest-creation-failed",
					Description: "Failed to create member interest ledger entry for account " + interestAccount.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
					Module:      "Loan Release",
				})
				return nil, endTx(eris.Wrap(err, "failed to create member interest ledger entry"))
			}
			_, err = e.core.MemberAccountingLedgerUpdateOrCreate(
				context,
				tx,
				core.MemberAccountingLedgerUpdateOrCreateParams{
					MemberProfileID: *loanTransaction.MemberProfileID, // Fixed: was loanTransaction.ID
					AccountID:       interestAccount.ID,
					OrganizationID:  userOrg.OrganizationID,
					BranchID:        *userOrg.BranchID,
					UserID:          userOrg.UserID,
					DebitAmount:     debit,
					CreditAmount:    credit,
					LastPayTime:     now,
				},
			)
			if err != nil {
				e.Footstep(ctx, FootstepEvent{
					Activity:    "member-interest-accounting-failed",
					Description: "Failed to update member accounting ledger for interest account " + interestAccount.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
					Module:      "Loan Release",
				})
				return nil, endTx(eris.Wrap(err, "failed to update member interest accounting ledger"))
			}

		} else {
			// Calculate straight interest amount for subsidiary account
			straightInterestAmount := e.usecase.ComputeInterestStraight(
				loanTransaction.TotalPrincipal, interestAccount.InterestStandard, loanTransaction.Terms)

			credit := 0.0
			debit := straightInterestAmount
			// Create subsidiary interest account ledger entry
			subsidiaryInterestEntry := &core.GeneralLedger{
				CreatedAt:                  now,
				CreatedByID:                userOrg.UserID,
				UpdatedAt:                  now,
				UpdatedByID:                userOrg.UserID,
				BranchID:                   *userOrg.BranchID,
				OrganizationID:             userOrg.OrganizationID,
				TransactionBatchID:         &transactionBatch.ID,
				ReferenceNumber:            loanTransaction.Voucher,
				EntryDate:                  currentTime,
				AccountID:                  &interestAccount.ID,
				PaymentTypeID:              interestAccount.DefaultPaymentTypeID,
				TransactionReferenceNumber: loanTransaction.Voucher,
				Source:                     core.GeneralLedgerSourceLoan,
				BankReferenceNumber:        "",
				EmployeeUserID:             &userOrg.UserID,
				Description:                interestAccount.Description,
				TypeOfPaymentType:          interestTypeOfPaymentType,
				Credit:                     credit,
				Debit:                      debit,
				CurrencyID:                 &loanAccountCurrency.ID,
				LoanTransactionID:          &loanTransaction.ID,
			}

			// Save subsidiary interest ledger entry to database
			if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, subsidiaryInterestEntry); err != nil {
				e.Footstep(ctx, FootstepEvent{
					Activity:    "subsidiary-interest-creation-failed",
					Description: "Failed to create subsidiary interest ledger entry for " + interestAccount.ID.String(),
					Module:      "Loan Release",
				})
				return nil, endTx(eris.Wrap(err, "failed to create subsidiary interest ledger entry"))
			}

			e.Footstep(ctx, FootstepEvent{
				Activity:    "interest-account-processed",
				Description: "Successfully processed interest account " + interestAccount.ID.String() + " with interest: " + fmt.Sprintf("%.2f", straightInterestAmount),
				Module:      "Loan Release",
			})
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
