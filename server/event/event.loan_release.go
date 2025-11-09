package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// LoanRelease processes loan release with necessary validations and commits the transaction.
// Returns the updated LoanTransaction after successful release.
func (e *Event) LoanRelease(ctx context.Context, echoCtx echo.Context, tx *gorm.DB, endTx func(error) error, data LoanBalanceEvent) (*core.LoanTransaction, error) {
	fmt.Printf("🚀 [LOAN-RELEASE] Starting loan release process for transaction ID: %s\n", data.LoanTransactionID.String())
	now := time.Now().UTC()

	// ================================================================================
	// STEP 1: AUTHENTICATION AND AUTHORIZATION
	// ================================================================================
	fmt.Printf("🔐 [STEP-1] Starting authentication and authorization\n")
	// Retrieve current user organization and validate permissions
	currentUserOrg, err := e.userOrganizationToken.CurrentUserOrganization(ctx, echoCtx)
	if err != nil {
		fmt.Printf("❌ [STEP-1] Failed to get user organization: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "authentication-failed",
			Description: "Unable to retrieve user organization details for loan release operation: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get user organization"))
	}

	fmt.Printf("✅ [STEP-1] User organization retrieved successfully. UserID: %s, BranchID: %v\n", currentUserOrg.UserID.String(), currentUserOrg.BranchID)

	// Validate branch assignment
	if currentUserOrg.BranchID == nil {
		fmt.Printf("❌ [STEP-1] Missing branch ID in user organization\n")
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "validation-failed",
			Description: "User organization is missing required branch assignment for loan operations",
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("invalid user organization data"))
	}

	// Validate user permissions for loan release
	if currentUserOrg.UserType != core.UserOrganizationTypeOwner && currentUserOrg.UserType != core.UserOrganizationTypeEmployee {
		fmt.Printf("❌ [STEP-1] Unauthorized user type: %s\n", currentUserOrg.UserType)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "authorization-failed",
			Description: "User does not have sufficient permissions to perform loan release operations",
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("unauthorized user role"))
	}

	fmt.Printf("✅ [STEP-1] Authentication and authorization completed successfully\n")

	// ================================================================================
	// STEP 2: TRANSACTION BATCH VALIDATION
	// ================================================================================
	fmt.Printf("📦 [STEP-2] Starting transaction batch validation\n")
	// Ensure there's an active transaction batch for recording the loan release
	activeBatch, err := e.core.TransactionBatchCurrent(ctx, currentUserOrg.UserID, currentUserOrg.OrganizationID, *currentUserOrg.BranchID)
	if err != nil {
		fmt.Printf("❌ [STEP-2] Failed to retrieve transaction batch: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "batch-retrieval-failed",
			Description: "Unable to retrieve active transaction batch for user " + currentUserOrg.UserID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve transaction batch"))
	}

	if activeBatch == nil {
		fmt.Printf("❌ [STEP-2] No active transaction batch found\n")
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "batch-validation-failed",
			Description: "No active transaction batch found - batch is required for loan release operations",
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("transaction batch is nil"))
	}

	fmt.Printf("✅ [STEP-2] Active transaction batch found: %s\n", activeBatch.ID.String())

	// ================================================================================
	// STEP 3: LOAN TRANSACTION DATA RETRIEVAL AND VALIDATION
	// ================================================================================
	fmt.Printf("💰 [STEP-3] Starting loan transaction data retrieval\n")
	// Fetch the loan transaction with account and currency details
	targetLoanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, data.LoanTransactionID,
		"Account", "Account.Currency")
	if err != nil {
		fmt.Printf("❌ [STEP-3] Failed to retrieve loan transaction: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "loan-data-retrieval-failed",
			Description: "Unable to retrieve loan transaction details for release: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction"))
	}

	fmt.Printf("✅ [STEP-3] Loan transaction retrieved. ID: %s, Balance: %.2f\n", targetLoanTransaction.ID.String(), targetLoanTransaction.Balance)

	// Validate currency information
	loanCurrency := targetLoanTransaction.Account.Currency
	if loanCurrency == nil {
		fmt.Printf("❌ [STEP-3] Currency is nil for account: %s\n", targetLoanTransaction.AccountID.String())
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "currency-validation-failed",
			Description: "Missing currency information for loan account " + targetLoanTransaction.AccountID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("currency data is nil"))
	}

	fmt.Printf("✅ [STEP-3] Currency validated: %s\n", loanCurrency.Name)

	// ================================================================================
	// STEP 4: CASH ON HAND ACCOUNT PROCESSING
	// ================================================================================
	fmt.Printf("💵 [STEP-4] Starting cash on hand account processing\n")
	// Retrieve the cash on hand account for the loan release
	cashAccount, err := e.core.GetCashOnCashEquivalence(
		ctx, targetLoanTransaction.ID, currentUserOrg.OrganizationID, *currentUserOrg.BranchID)
	if err != nil {
		fmt.Printf("❌ [STEP-4] Failed to get cash on hand account: %v\n", err)
		return nil, endTx(eris.Wrap(err, "failed to get cash on hand account"))
	}

	fmt.Printf("✅ [STEP-4] Cash account retrieved: %s\n", cashAccount.ID.String())

	// Lock the subsidiary ledger for the cash account
	cashAccountLedger, err := e.core.GeneralLedgerCurrentSubsidiaryAccountForUpdate(
		ctx, tx, cashAccount.ID, currentUserOrg.OrganizationID, *currentUserOrg.BranchID)
	if err != nil {
		fmt.Printf("❌ [STEP-4] Failed to lock cash account ledger: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "cash-ledger-lock-failed",
			Description: "Unable to acquire lock on cash account subsidiary ledger " + cashAccount.ID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve subsidiary general ledger"))
	}

	// Determine current cash balance
	var currentCashBalance float64 = 0
	if cashAccountLedger == nil {
		fmt.Printf("💡 [STEP-4] No previous cash ledger found, initializing with zero balance\n")
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "cash-ledger-initialization",
			Description: "Initializing new cash account ledger for " + cashAccount.ID.String() + " with zero balance",
			Module:      "Loan Release",
		})
	} else {
		currentCashBalance = cashAccountLedger.Balance
		fmt.Printf("✅ [STEP-4] Current cash balance: %.2f\n", currentCashBalance)
	}

	// ================================================================================
	// STEP 5: CASH ACCOUNT BALANCE CALCULATION AND LEDGER ENTRY
	// ================================================================================
	fmt.Printf("🧮 [STEP-5] Starting cash account balance calculation\n")
	// Calculate the adjusted debit, credit, and resulting balance for cash account
	cashDebit, cashCredit, newCashBalance := e.usecase.Adjustment(*cashAccount.Account, targetLoanTransaction.Balance, 0.0, currentCashBalance)

	fmt.Printf("🧮 [STEP-5] Cash calculation - Debit: %.2f, Credit: %.2f, New Balance: %.2f\n", cashDebit, cashCredit, newCashBalance)

	// Create general ledger entry for cash on hand
	fmt.Printf("🔍 [STEP-5] Starting to create cash ledger entry...\n")

	// Check each pointer before using it
	fmt.Printf("🔍 [STEP-5] Checking currentUserOrg.UserID: %v\n", currentUserOrg.UserID)
	fmt.Printf("🔍 [STEP-5] Checking currentUserOrg.BranchID: %v\n", currentUserOrg.BranchID)
	if currentUserOrg.BranchID != nil {
		fmt.Printf("🔍 [STEP-5] BranchID value: %s\n", currentUserOrg.BranchID.String())
	}
	fmt.Printf("🔍 [STEP-5] Checking activeBatch.ID: %v\n", activeBatch.ID)
	fmt.Printf("🔍 [STEP-5] Checking targetLoanTransaction.CheckNumber: %v\n", targetLoanTransaction.CheckNumber)
	fmt.Printf("🔍 [STEP-5] Checking cashAccount.ID: %v\n", cashAccount.ID)
	fmt.Printf("🔍 [STEP-5] Checking cashAccount.Account: %v\n", cashAccount.Account)
	if cashAccount.Account != nil {
		fmt.Printf("🔍 [STEP-5] Checking cashAccount.Account.DefaultPaymentTypeID: %v\n", cashAccount.Account.DefaultPaymentTypeID)
		fmt.Printf("🔍 [STEP-5] Checking cashAccount.Account.DefaultPaymentType: %v\n", cashAccount.Account.DefaultPaymentType)
		if cashAccount.Account.DefaultPaymentType != nil {
			fmt.Printf("🔍 [STEP-5] DefaultPaymentType.Type: %v\n", cashAccount.Account.DefaultPaymentType.Type)
		}
	}
	fmt.Printf("🔍 [STEP-5] Checking cashAccount.Description: %v\n", cashAccount.Description)
	fmt.Printf("🔍 [STEP-5] Checking loanCurrency.ID: %v\n", loanCurrency.ID)
	fmt.Printf("🔍 [STEP-5] Checking targetLoanTransaction.ID: %v\n", targetLoanTransaction.ID)

	cashLedgerEntry := &core.GeneralLedger{
		CreatedAt:                  now,
		CreatedByID:                currentUserOrg.UserID,
		UpdatedAt:                  now,
		UpdatedByID:                currentUserOrg.UserID,
		BranchID:                   *currentUserOrg.BranchID,
		OrganizationID:             currentUserOrg.OrganizationID,
		TransactionBatchID:         &activeBatch.ID,
		ReferenceNumber:            targetLoanTransaction.CheckNumber,
		EntryDate:                  &now,
		AccountID:                  &cashAccount.ID,
		PaymentTypeID:              cashAccount.Account.DefaultPaymentTypeID,
		TransactionReferenceNumber: targetLoanTransaction.CheckNumber,
		Source:                     core.GeneralLedgerSourceCheckVoucher,
		BankReferenceNumber:        "",
		EmployeeUserID:             &currentUserOrg.UserID,
		Description:                cashAccount.Description,
		TypeOfPaymentType:          cashAccount.Account.DefaultPaymentType.Type,
		Credit:                     cashCredit,
		Debit:                      cashDebit,
		Balance:                    newCashBalance,
		CurrencyID:                 &loanCurrency.ID,
		LoanTransactionID:          &targetLoanTransaction.ID,
	}

	fmt.Printf("✅ [STEP-5] Cash ledger entry struct created successfully\n")

	if err := e.core.GeneralLedgerManager.CreateWithTx(ctx, tx, cashLedgerEntry); err != nil {
		fmt.Printf("❌ [STEP-5] Failed to create cash ledger entry: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "cash-ledger-creation-failed",
			Description: "Unable to create cash account ledger entry for " + cashAccount.ID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to create general ledger entry"))
	}

	fmt.Printf("✅ [STEP-5] Cash ledger entry created successfully\n")

	e.Footstep(echoCtx, FootstepEvent{
		Activity:    "cash-transaction-completed",
		Description: "Successfully updated cash account " + cashAccount.ID.String() + " with new balance: " + fmt.Sprintf("%.2f", newCashBalance),
		Module:      "Loan Release",
	})

	// ================================================================================
	// STEP 6: MEMBER ACCOUNT LEDGER PROCESSING
	// ================================================================================
	fmt.Printf("👤 [STEP-6] Starting member account ledger processing\n")
	// Retrieve member profile associated with the loan
	memberProfile, err := e.core.MemberProfileManager.GetByID(ctx, *targetLoanTransaction.MemberProfileID)
	if err != nil {
		fmt.Printf("❌ [STEP-6] Failed to retrieve member profile: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "member-profile-retrieval-failed",
			Description: "Unable to retrieve member profile " + targetLoanTransaction.MemberProfileID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve member profile"))
	}
	if memberProfile == nil {
		fmt.Printf("❌ [STEP-6] Member profile is nil\n")
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "member-profile-not-found",
			Description: "Member profile does not exist for ID: " + targetLoanTransaction.MemberProfileID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("member profile not found"))
	}

	fmt.Printf("✅ [STEP-6] Member profile retrieved: %s\n", memberProfile.ID.String())

	// Lock member's general ledger for the loan account
	memberAccountLedger, err := e.core.GeneralLedgerCurrentMemberAccountForUpdate(
		ctx, tx,
		memberProfile.ID,
		*targetLoanTransaction.AccountID,
		memberProfile.OrganizationID,
		memberProfile.BranchID,
	)
	if err != nil {
		fmt.Printf("❌ [STEP-6] Failed to lock member account ledger: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "member-ledger-lock-failed",
			Description: "Unable to acquire lock on member ledger for account " + targetLoanTransaction.AccountID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve member ledger for update"))
	}

	// Determine current member account balance
	var currentMemberBalance float64 = 0
	if memberAccountLedger == nil {
		fmt.Printf("💡 [STEP-6] No previous member ledger found, initializing with zero balance\n")
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "member-ledger-initialization",
			Description: "Initializing new member ledger for account " + targetLoanTransaction.AccountID.String() + " and member " + memberProfile.ID.String(),
			Module:      "Loan Release",
		})
	} else {
		currentMemberBalance = memberAccountLedger.Balance
		fmt.Printf("✅ [STEP-6] Current member balance: %.2f\n", currentMemberBalance)
	}

	// ================================================================================
	// STEP 7: MEMBER ACCOUNT BALANCE CALCULATION AND LEDGER ENTRY
	// ================================================================================
	fmt.Printf("🧮 [STEP-7] Starting member account balance calculation\n")
	// Calculate adjusted debit, credit, and resulting balance for member's loan account
	memberDebit, memberCredit, newMemberBalance := e.usecase.Adjustment(
		*targetLoanTransaction.Account, 0.0, targetLoanTransaction.Balance, currentMemberBalance)

	fmt.Printf("🧮 [STEP-7] Member calculation - Debit: %.2f, Credit: %.2f, New Balance: %.2f\n", memberDebit, memberCredit, newMemberBalance)

	fmt.Printf("🔍 [STEP-7] Starting to create member ledger entry...\n")

	// Check each pointer before using it for member ledger
	fmt.Printf("🔍 [STEP-7] Checking currentUserOrg.UserID: %v\n", currentUserOrg.UserID)
	fmt.Printf("🔍 [STEP-7] Checking currentUserOrg.BranchID: %v\n", currentUserOrg.BranchID)
	fmt.Printf("🔍 [STEP-7] Checking activeBatch.ID: %v\n", activeBatch.ID)
	fmt.Printf("🔍 [STEP-7] Checking targetLoanTransaction.CheckNumber: %v\n", targetLoanTransaction.CheckNumber)
	fmt.Printf("🔍 [STEP-7] Checking targetLoanTransaction.AccountID: %v\n", targetLoanTransaction.AccountID)
	fmt.Printf("🔍 [STEP-7] Checking memberProfile.ID: %v\n", memberProfile.ID)
	fmt.Printf("🔍 [STEP-7] Checking cashAccount.Account: %v\n", cashAccount.Account)
	if cashAccount.Account != nil {
		fmt.Printf("🔍 [STEP-7] Checking cashAccount.Account.DefaultPaymentTypeID: %v\n", cashAccount.Account.DefaultPaymentTypeID)
	}
	fmt.Printf("🔍 [STEP-7] Checking targetLoanTransaction.Account: %v\n", targetLoanTransaction.Account)
	if targetLoanTransaction.Account != nil {
		fmt.Printf("🔍 [STEP-7] Checking targetLoanTransaction.Account.Description: %v\n", targetLoanTransaction.Account.Description)
	}
	fmt.Printf("🔍 [STEP-7] Checking loanCurrency.ID: %v\n", loanCurrency.ID)

	var paymentTypeValue core.TypeOfPaymentType
	memberLedgerEntry := &core.GeneralLedger{
		CreatedAt:                  now,
		CreatedByID:                currentUserOrg.UserID,
		UpdatedAt:                  now,
		UpdatedByID:                currentUserOrg.UserID,
		BranchID:                   *currentUserOrg.BranchID,
		OrganizationID:             currentUserOrg.OrganizationID,
		TransactionBatchID:         &activeBatch.ID,
		ReferenceNumber:            targetLoanTransaction.CheckNumber,
		EntryDate:                  &now,
		AccountID:                  targetLoanTransaction.AccountID,
		MemberProfileID:            &memberProfile.ID,
		PaymentTypeID:              cashAccount.Account.DefaultPaymentTypeID,
		TransactionReferenceNumber: targetLoanTransaction.CheckNumber,
		Source:                     core.GeneralLedgerSourceCheckVoucher,
		EmployeeUserID:             &currentUserOrg.UserID,
		Description:                targetLoanTransaction.Account.Description,
		TypeOfPaymentType:          paymentTypeValue,
		Credit:                     memberCredit,
		Debit:                      memberDebit,
		Balance:                    newMemberBalance,
		CurrencyID:                 &loanCurrency.ID,
	}

	fmt.Printf("✅ [STEP-7] Member ledger entry struct created successfully\n")

	// Create the member's general ledger entry in the database
	if err := e.core.GeneralLedgerManager.CreateWithTx(ctx, tx, memberLedgerEntry); err != nil {
		fmt.Printf("❌ [STEP-7] Failed to create member ledger entry: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "member-ledger-creation-failed",
			Description: "Unable to create member ledger entry for account " + memberLedgerEntry.AccountID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to create general ledger entry"))
	}

	fmt.Printf("✅ [STEP-7] Member ledger entry created successfully\n")

	// ================================================================================
	// STEP 8: MEMBER ACCOUNTING LEDGER UPDATE
	// ================================================================================
	fmt.Printf("📊 [STEP-8] Starting member accounting ledger update\n")
	// Update or create member accounting ledger with new balance
	_, err = e.core.MemberAccountingLedgerUpdateOrCreate(
		ctx,
		tx,
		*targetLoanTransaction.MemberProfileID,
		*targetLoanTransaction.AccountID,
		currentUserOrg.OrganizationID,
		*currentUserOrg.BranchID,
		currentUserOrg.UserID,
		newMemberBalance,
		now,
	)
	if err != nil {
		fmt.Printf("❌ [STEP-8] Failed to update member accounting ledger: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "member-accounting-ledger-update-failed",
			Description: "Unable to update member accounting ledger for member " + targetLoanTransaction.MemberProfileID.String() + " on account " + targetLoanTransaction.AccountID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to update member accounting ledger"))
	}

	fmt.Printf("✅ [STEP-8] Member accounting ledger updated successfully\n")

	// Log successful member transaction completion
	e.Footstep(echoCtx, FootstepEvent{
		Activity:    "member-transaction-completed",
		Description: "Successfully processed member loan account for " + memberProfile.ID.String() + " with new balance: " + fmt.Sprintf("%.2f", newMemberBalance),
		Module:      "Loan Release",
	})

	// ================================================================================
	// STEP 9: LOAN TRANSACTION FINALIZATION
	// ================================================================================
	fmt.Printf("📝 [STEP-9] Starting loan transaction finalization\n")
	// Update loan transaction with release information
	targetLoanTransaction.ReleasedDate = &now
	targetLoanTransaction.ReleasedByID = &currentUserOrg.UserID
	targetLoanTransaction.UpdatedAt = now
	targetLoanTransaction.UpdatedByID = currentUserOrg.UserID

	if err := e.core.LoanTransactionManager.UpdateByIDWithTx(ctx, tx, targetLoanTransaction.ID, targetLoanTransaction); err != nil {
		fmt.Printf("❌ [STEP-9] Failed to update loan transaction: %v\n", err)
		return nil, endTx(eris.Wrap(err, "failed to update loan transaction"))
	}

	fmt.Printf("✅ [STEP-9] Loan transaction finalized successfully\n")

	// ================================================================================
	// STEP 10: DATABASE TRANSACTION COMMIT
	// ================================================================================
	fmt.Printf("💾 [STEP-10] Starting database transaction commit\n")
	// Commit all changes to the database
	if err := endTx(nil); err != nil {
		fmt.Printf("❌ [STEP-10] Failed to commit database transaction: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "database-commit-failed",
			Description: "Unable to commit loan release transaction to database: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to commit transaction"))
	}

	fmt.Printf("✅ [STEP-10] Database transaction committed successfully\n")

	// ================================================================================
	// STEP 11: FINAL TRANSACTION RETRIEVAL AND RETURN
	// ================================================================================
	fmt.Printf("🔍 [STEP-11] Starting final transaction retrieval\n")
	// Retrieve and return the updated loan transaction
	updatedLoanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, targetLoanTransaction.ID)
	if err != nil {
		fmt.Printf("❌ [STEP-11] Failed to retrieve updated loan transaction: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "final-retrieval-failed",
			Description: "Unable to retrieve updated loan transaction after successful release: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get updated loan transaction"))
	}

	fmt.Printf("✅ [STEP-11] Updated loan transaction retrieved successfully\n")
	fmt.Printf("🎉 [LOAN-RELEASE] Loan release process completed successfully for transaction: %s\n", updatedLoanTransaction.ID.String())

	return updatedLoanTransaction, nil
}
