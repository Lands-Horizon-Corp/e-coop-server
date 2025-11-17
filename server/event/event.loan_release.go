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
	userOrgTime := time.Now().UTC()
	fmt.Println("DEBUG: userOrg:", userOrg)
	if userOrg.TimeMachineTime != nil {
		userOrgTime = userOrg.UserOrgTime()
	}
	fmt.Println("DEBUG: userOrgTime:", userOrgTime)

	// Validate user organization has required branch assignment
	fmt.Println("DEBUG: userOrg.BranchID:", userOrg.BranchID)
	if userOrg.BranchID == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "validation-failed",
			Description: "User organization is missing required branch assignment for loan operations",
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("invalid user organization data"))
	}
	fmt.Println("DEBUG: Branch ID validated:", *userOrg.BranchID)

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
	fmt.Println("DEBUG: Fetching loan transaction ID:", loanTransactionID)
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, loanTransactionID, "Account", "Account.Currency")
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "loan-data-retrieval-failed",
			Description: "Unable to retrieve loan transaction details for release: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction"))
	}
	fmt.Println("DEBUG: loanTransaction:", loanTransaction)
	fmt.Println("DEBUG: loanTransaction.Account:", loanTransaction.Account)

	// Extract and validate currency information from loan account
	fmt.Println("DEBUG: loanTransaction.Account.Currency:", loanTransaction.Account.Currency)
	loanAccountCurrency := loanTransaction.Account.Currency
	if loanAccountCurrency == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "currency-validation-failed",
			Description: "Missing currency information for loan account " + loanTransaction.AccountID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("currency data is nil"))
	}
	fmt.Println("DEBUG: loanAccountCurrency:", loanAccountCurrency)

	// ================================================================================
	// STEP 3: ACTIVE TRANSACTION BATCH VALIDATION
	// ================================================================================
	// Retrieve and validate active transaction batch required for loan operations
	fmt.Println("DEBUG: loanTransaction.EmployeeUserID:", loanTransaction.EmployeeUserID)
	fmt.Println("DEBUG: userOrg.OrganizationID:", userOrg.OrganizationID)
	fmt.Println("DEBUG: userOrg.BranchID:", userOrg.BranchID)
	transactionBatch, err := e.core.TransactionBatchCurrent(context, *loanTransaction.EmployeeUserID, userOrg.OrganizationID, *userOrg.BranchID)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "batch-retrieval-failed",
			Description: "Unable to retrieve active transaction batch for user " + userOrg.UserID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve transaction batch - The one who created the loan must have created the transaction batch"))
	}
	fmt.Println("DEBUG: transactionBatch:", transactionBatch)

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
	fmt.Println("DEBUG: loanTransaction.MemberProfileID:", loanTransaction.MemberProfileID)
	memberProfile, err := e.core.MemberProfileManager.GetByID(context, *loanTransaction.MemberProfileID)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "member-profile-retrieval-failed",
			Description: "Unable to retrieve member profile " + loanTransaction.MemberProfileID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve member profile"))
	}
	fmt.Println("DEBUG: memberProfile:", memberProfile)

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
	fmt.Println("DEBUG: Fetching loan transaction entries for loan ID:", loanTransaction.ID)
	loanTransactionEntries, err := e.core.LoanTransactionEntryManager.Find(context, &core.LoanTransactionEntry{
		LoanTransactionID: loanTransaction.ID,
		OrganizationID:    loanTransaction.OrganizationID,
		BranchID:          loanTransaction.BranchID,
	})
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to retrieve loan transaction entries"))
	}
	fmt.Println("DEBUG: Found", len(loanTransactionEntries), "loan transaction entries")

	// Process each loan transaction entry for ledger updates
	for i, entry := range loanTransactionEntries {
		fmt.Println("DEBUG: Processing entry", i+1, "of", len(loanTransactionEntries))
		fmt.Println("DEBUG: entry:", entry)
		fmt.Println("DEBUG: entry.AccountID:", entry.AccountID)
		// Skip entries marked as deleted automatic loan deductions
		if entry.IsAutomaticLoanDeductionDeleted {
			fmt.Println("DEBUG: Skipping deleted entry", i+1)
			continue
		}

		// Retrieve account history for the transaction entry at the specific time
		fmt.Println("DEBUG: Fetching account history for AccountID:", *entry.AccountID)
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
		fmt.Println("DEBUG: accountHistory:", accountHistory)

		// Convert account history to account model for processing
		fmt.Println("DEBUG: Converting account history to model")
		account := e.core.AccountHistoryToModel(accountHistory)
		fmt.Println("DEBUG: account:", account)
		if accountHistory == nil {
			return nil, endTx(eris.New("account history not found for entry"))
		}

		// Process member ledger accounts differently from subsidiary accounts
		fmt.Println("DEBUG: account.IsMemberLedger():", account.IsMemberLedger())
		fmt.Println("DEBUG: account.DefaultPaymentTypeID:", account.DefaultPaymentTypeID)
		fmt.Println("DEBUG: account.DefaultPaymentType:", account.DefaultPaymentType)
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
				EntryDate:                  &userOrgTime,
				AccountID:                  &account.ID,
				MemberProfileID:            &memberProfile.ID,
				PaymentTypeID:              account.DefaultPaymentTypeID,
				TransactionReferenceNumber: loanTransaction.Voucher,
				Source:                     core.GeneralLedgerSourceLoan,
				EmployeeUserID:             &userOrg.UserID,
				Description:                loanTransaction.Account.Description,
				TypeOfPaymentType:          account.DefaultPaymentType.Type,
				Credit:                     entry.Credit,
				Debit:                      entry.Debit,
				CurrencyID:                 &loanAccountCurrency.ID,
				LoanTransactionID:          &loanTransaction.ID,
			}

			fmt.Println("DEBUG: Creating member ledger entry:", memberLedgerEntry)
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
			fmt.Println("DEBUG: Updating member accounting ledger for MemberProfileID:", *loanTransaction.MemberProfileID)
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
			fmt.Println("DEBUG: Processing non-member ledger entry")
			intMemberEntry := &core.GeneralLedger{
				CreatedAt:                  now,
				CreatedByID:                userOrg.UserID,
				UpdatedAt:                  now,
				UpdatedByID:                userOrg.UserID,
				BranchID:                   *userOrg.BranchID,
				OrganizationID:             userOrg.OrganizationID,
				TransactionBatchID:         &transactionBatch.ID,
				ReferenceNumber:            loanTransaction.Voucher,
				EntryDate:                  &userOrgTime,
				AccountID:                  &account.ID,
				MemberProfileID:            &memberProfile.ID,
				PaymentTypeID:              account.DefaultPaymentTypeID,
				TransactionReferenceNumber: loanTransaction.Voucher,
				Source:                     core.GeneralLedgerSourceLoan,
				EmployeeUserID:             &userOrg.UserID,
				Description:                loanTransaction.Account.Description,
				TypeOfPaymentType:          account.DefaultPaymentType.Type,
				Credit:                     entry.Credit,
				Debit:                      entry.Debit,
				CurrencyID:                 &loanAccountCurrency.ID,
				LoanTransactionID:          &loanTransaction.ID,
			}
			fmt.Println("DEBUG: Creating interest member ledger entry:", intMemberEntry)
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
	fmt.Println("DEBUG: Fetching loan-related accounts for interest processing")
	fmt.Println("DEBUG: loanTransaction.AccountID:", loanTransaction.AccountID)
	loanRelatedAccounts, err := e.core.GetAccountHistoriesByFiltersAtTime(
		context,
		loanTransaction.OrganizationID,
		loanTransaction.BranchID,
		&userOrgTime,
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

	fmt.Println("DEBUG: Found", len(loanRelatedAccounts), "loan-related accounts for interest")
	// Process each interest-bearing account with straight computation
	for j, interestAccount := range loanRelatedAccounts {
		fmt.Println("DEBUG: Processing interest account", j+1, "of", len(loanRelatedAccounts))
		fmt.Println("DEBUG: interestAccount:", interestAccount)
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

		fmt.Println("DEBUG: interestAccountHistory:", interestAccountHistory)
		// Use account history if available, otherwise use current account data
		if interestAccountHistory != nil {
			interestAccount = e.core.AccountHistoryToModel(interestAccountHistory)
		}
		fmt.Println("DEBUG: Final interestAccount:", interestAccount)

		// Filter accounts: only process loan-related accounts with straight computation
		// Skip loan principal accounts and fines accounts
		fmt.Println("DEBUG: interestAccount.LoanAccountID:", interestAccount.LoanAccountID)
		fmt.Println("DEBUG: interestAccount.ComputationType:", interestAccount.ComputationType)
		fmt.Println("DEBUG: interestAccount.Type:", interestAccount.Type)
		if interestAccount.LoanAccountID == nil ||
			interestAccount.ComputationType != core.Straight ||
			(interestAccount.Type == core.AccountTypeLoan || interestAccount.Type == core.AccountTypeFines) {
			fmt.Println("DEBUG: Skipping interest account due to filter conditions")
			continue
		}
		// Process member interest accounts
		fmt.Println("DEBUG: Processing interest - IsMemberLedger:", interestAccount.IsMemberLedger())
		fmt.Println("DEBUG: interestAccount.DefaultPaymentTypeID:", interestAccount.DefaultPaymentTypeID)
		fmt.Println("DEBUG: interestAccount.DefaultPaymentType:", interestAccount.DefaultPaymentType)
		if interestAccount.IsMemberLedger() {
			// Lock member interest account ledger for atomic updates

			straightInterestAmount := e.usecase.ComputeInterestStraight(
				loanTransaction.TotalPrincipal, interestAccount.InterestStandard, loanTransaction.Terms)
			fmt.Println("DEBUG: straightInterestAmount:", straightInterestAmount)

			credit := straightInterestAmount
			debit := 0.0
			fmt.Println("DEBUG: Creating member interest entry - credit:", credit, "debit:", debit)
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
				EntryDate:                  &userOrgTime,
				AccountID:                  &interestAccount.ID,
				MemberProfileID:            &memberProfile.ID,
				PaymentTypeID:              interestAccount.DefaultPaymentTypeID,
				TransactionReferenceNumber: loanTransaction.Voucher,
				Source:                     core.GeneralLedgerSourceLoan,
				EmployeeUserID:             &userOrg.UserID,
				Description:                loanTransaction.Account.Description,
				TypeOfPaymentType:          interestAccount.DefaultPaymentType.Type,
				Credit:                     credit,
				Debit:                      debit,
				CurrencyID:                 &loanAccountCurrency.ID,
				LoanTransactionID:          &loanTransaction.ID,
			}

			fmt.Println("DEBUG: Saving member interest ledger entry:", memberInterestEntry)
			// Save member interest ledger entry to database
			if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, memberInterestEntry); err != nil {
				e.Footstep(ctx, FootstepEvent{
					Activity:    "member-interest-creation-failed",
					Description: "Failed to create member interest ledger entry for account " + interestAccount.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
					Module:      "Loan Release",
				})
				return nil, endTx(eris.Wrap(err, "failed to create member interest ledger entry"))
			}
			fmt.Println("DEBUG: Updating member accounting ledger for interest - MemberProfileID:", *loanTransaction.MemberProfileID)
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
			fmt.Println("DEBUG: Processing subsidiary interest account")
			// Calculate straight interest amount for subsidiary account
			straightInterestAmount := e.usecase.ComputeInterestStraight(
				loanTransaction.TotalPrincipal, interestAccount.InterestStandard, loanTransaction.Terms)
			fmt.Println("DEBUG: Subsidiary straightInterestAmount:", straightInterestAmount)

			credit := 0.0
			debit := straightInterestAmount
			fmt.Println("DEBUG: Creating subsidiary interest entry - credit:", credit, "debit:", debit)
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
				EntryDate:                  &userOrgTime,
				AccountID:                  &interestAccount.ID,
				PaymentTypeID:              interestAccount.DefaultPaymentTypeID,
				TransactionReferenceNumber: loanTransaction.Voucher,
				Source:                     core.GeneralLedgerSourceLoan,
				BankReferenceNumber:        "",
				EmployeeUserID:             &userOrg.UserID,
				Description:                interestAccount.Description,
				TypeOfPaymentType:          interestAccount.DefaultPaymentType.Type,
				Credit:                     credit,
				Debit:                      debit,
				CurrencyID:                 &loanAccountCurrency.ID,
				LoanTransactionID:          &loanTransaction.ID,
			}

			fmt.Println("DEBUG: Saving subsidiary interest ledger entry:", subsidiaryInterestEntry)
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
	fmt.Println("DEBUG: Updating loan transaction with release information")
	// Update loan transaction with release information and timestamps
	loanTransaction.ReleasedDate = &userOrgTime
	loanTransaction.ReleasedByID = &userOrg.UserID
	loanTransaction.UpdatedAt = now
	loanTransaction.Count++
	loanTransaction.UpdatedByID = userOrg.UserID
	fmt.Println("DEBUG: Updated loanTransaction:", loanTransaction)

	// Save updated loan transaction to database
	fmt.Println("DEBUG: Saving updated loan transaction to database")
	if err := e.core.LoanTransactionManager.UpdateByIDWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
		return nil, endTx(eris.Wrap(err, "failed to update loan transaction"))
	}
	fmt.Println("DEBUG: Loan transaction saved successfully")

	// ================================================================================
	// STEP 8: DATABASE TRANSACTION COMMIT
	// ================================================================================
	// Commit all changes to the database atomically
	fmt.Println("DEBUG: Committing database transaction")
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
	fmt.Println("DEBUG: Retrieving final updated loan transaction")
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

	fmt.Println("DEBUG: Loan release completed successfully")
	return updatedloanTransaction, nil
}
