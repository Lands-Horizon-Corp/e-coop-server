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

type RecordTransactionRequest struct {

	// Amount
	Debit  float64
	Credit float64

	// AccountID
	AccountID       uuid.UUID
	MemberProfileID *uuid.UUID

	// To record current balance after transaction
	TransactionBatchID uuid.UUID

	ReferenceNumber string

	EntryDate        *time.Time `json:"entry_date"`
	SignatureMediaID *uuid.UUID `json:"signature_media_id"`

	PaymentTypeID         *uuid.UUID `json:"payment_type_id"`
	BankReferenceNumber   string     `json:"bank_reference_number"`
	Description           string     `json:"description"`
	BankID                *uuid.UUID `json:"bank_id"`
	ProofOfPaymentMediaID *uuid.UUID `json:"proof_of_payment_media_id"`
}

// RecordTransaction handles the complete transaction recording process for both member and subsidiary accounts.
// This function supports posting for: loans, cash check vouchers, journal vouchers, and adjustment entries.
//
// Process Flow:
// 1. Input validation and database transaction initialization
// 2. Authentication and organization context retrieval
// 3. Transaction batch validation
// 4. Account and payment type resolution
// 5. General ledger entry creation (member or subsidiary)
// 6. Balance updates and accounting ledger maintenance
func (e Event) RecordTransaction(
	context context.Context,
	echoCtx echo.Context,
	transaction RecordTransactionRequest,
	source core.GeneralLedgerSource,
) error {
	fmt.Printf("DEBUG LINE 42: RecordTransaction started with source: %v\n", source)
	now := time.Now().UTC()
	fmt.Printf("DEBUG LINE 44: Time initialized: %v\n", now)

	// ================================================================================
	// STEP 1: DATABASE TRANSACTION INITIALIZATION
	// ================================================================================
	// Start database transaction to ensure atomicity of all operations
	fmt.Printf("DEBUG LINE 49: Starting database transaction\n")
	tx, endTx := e.provider.Service.Database.StartTransaction(context)
	fmt.Printf("DEBUG LINE 51: Database transaction started successfully\n")

	// ================================================================================
	// STEP 2: INPUT VALIDATION
	// ================================================================================
	// Validate that at least one amount (credit or debit) is provided
	fmt.Printf("DEBUG LINE 56: Validating transaction amounts - Debit: %f, Credit: %f\n", transaction.Debit, transaction.Credit)
	if transaction.Credit == 0 && transaction.Debit == 0 {
		fmt.Printf("DEBUG LINE 58: Both amounts are zero - returning error\n")
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "validation-error",
			Description: "Transaction validation failed: Both credit and debit amounts are zero",
			Module:      "Transaction Recording",
		})
		return endTx(eris.New("both credit and debit cannot be zero"))
	}

	// Validate required account ID
	fmt.Printf("DEBUG LINE 67: Validating account ID: %v\n", transaction.AccountID)
	if transaction.AccountID == uuid.Nil {
		fmt.Printf("DEBUG LINE 69: Account ID is nil - returning error\n")
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "validation-error",
			Description: "Transaction validation failed: Account ID is missing or invalid",
			Module:      "Transaction Recording",
		})
		return endTx(eris.New("account ID is required"))
	}

	// Validate required reference number
	fmt.Printf("DEBUG LINE 78: Validating reference number: %s\n", transaction.ReferenceNumber)
	if transaction.ReferenceNumber == "" {
		fmt.Printf("DEBUG LINE 80: Reference number is empty - returning error\n")
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "validation-error",
			Description: "Transaction validation failed: Reference number is missing",
			Module:      "Transaction Recording",
		})
		return endTx(eris.New("reference number is required"))
	}
	fmt.Printf("DEBUG LINE 87: Input validation completed successfully\n")

	// ================================================================================
	// STEP 3: USER AUTHENTICATION & ORGANIZATION CONTEXT
	// ================================================================================
	// Retrieve the current user's organization context for transaction authorization
	fmt.Printf("DEBUG LINE 92: Retrieving user organization context\n")
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(context, echoCtx)
	if err != nil {
		fmt.Printf("DEBUG LINE 95: Failed to get user organization: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "authentication-failed",
			Description: "Unable to retrieve user organization context for transaction recording: " + err.Error(),
			Module:      "Transaction Recording",
		})
		return endTx(eris.Wrap(err, "failed to get user organization"))
	}
	fmt.Printf("DEBUG LINE 103: User organization retrieved successfully\n")

	// Ensure user organization context exists
	if userOrg == nil {
		fmt.Printf("DEBUG LINE 107: User organization is nil\n")
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "authentication-failed",
			Description: "User organization context is missing - cannot proceed with transaction recording",
			Module:      "Transaction Recording",
		})
		return endTx(eris.New("user organization is nil"))
	}
	fmt.Printf("DEBUG LINE 115: User organization is valid - UserID: %v, OrganizationID: %v\n", userOrg.UserID, userOrg.OrganizationID)

	// Validate branch assignment for transaction context
	if userOrg.BranchID == nil {
		fmt.Printf("DEBUG LINE 119: User branch ID is nil\n")
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "branch-context-error",
			Description: "User is not assigned to any branch - branch context required for transaction recording",
			Module:      "Transaction Recording",
		})
		return endTx(eris.New("user organization branch ID is nil"))
	}
	fmt.Printf("DEBUG LINE 127: User branch ID is valid: %v\n", *userOrg.BranchID)
	// ================================================================================
	// STEP 4: TRANSACTION BATCH VALIDATION
	// ================================================================================
	// Get the current active transaction batch for grouping related transactions
	transactionBatch, err := e.core.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "batch-retrieval-failed",
			Description: "Cannot retrieve current transaction batch for user " + userOrg.UserID.String() + ": " + err.Error(),
			Module:      "Transaction Recording",
		})
		return endTx(eris.Wrap(err, "failed to retrieve transaction batch"))
	}

	// Ensure a valid transaction batch exists
	if transactionBatch == nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "batch-missing",
			Description: "No active transaction batch found for current user session - batch is required for transaction recording",
			Module:      "Transaction Recording",
		})
		return endTx(eris.New("transaction batch is nil"))
	}

	// ================================================================================
	// STEP 5: ACCOUNT RESOLUTION AND LOCKING
	// ================================================================================
	// Lock the target account for update to prevent concurrent modifications
	account, err := e.core.AccountLockForUpdate(context, tx, transaction.AccountID)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "account-lock-failed",
			Description: "Failed to lock account " + transaction.AccountID.String() + " for transaction update: " + err.Error(),
			Module:      "Transaction Recording",
		})
		return endTx(eris.Wrap(err, "failed to lock account for update"))
	}

	// ================================================================================
	// STEP 6: PAYMENT TYPE RESOLUTION
	// ================================================================================
	// Resolve payment type details if specified
	var paymentType *core.PaymentType
	if transaction.PaymentTypeID != nil {
		paymentType, err = e.core.PaymentTypeManager.GetByID(context, *transaction.PaymentTypeID)
		if err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "payment-type-resolution-failed",
				Description: "Failed to resolve payment type " + transaction.PaymentTypeID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to resolve payment type"))
		}
	}
	// ================================================================================
	// STEP 7: TRANSACTION PROCESSING - MEMBER ACCOUNT PATH
	// ================================================================================
	if transaction.MemberProfileID != nil {
		// --- SUB-STEP 7A: MEMBER PROFILE VALIDATION ---
		// Retrieve and validate member profile for member-specific transactions
		memberProfile, err := e.core.MemberProfileManager.GetByID(context, *transaction.MemberProfileID)
		if err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "member-profile-retrieval-failed",
				Description: "Failed to retrieve member profile " + transaction.MemberProfileID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to retrieve member profile"))
		}

		if memberProfile == nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "member-profile-not-found",
				Description: "Member profile not found for ID: " + transaction.MemberProfileID.String(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.New("member profile not found"))
		}

		// --- SUB-STEP 7B: MEMBER LEDGER RETRIEVAL ---
		// Get current member account ledger with row-level locking
		generalLedger, err := e.core.GeneralLedgerCurrentMemberAccountForUpdate(
			context, tx, memberProfile.ID, account.ID, memberProfile.OrganizationID, memberProfile.BranchID,
		)
		if err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "member-ledger-lock-failed",
				Description: "Failed to lock member ledger for account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to retrieve member ledger for update"))
		}

		// --- SUB-STEP 7C: BALANCE CALCULATION ---
		// Calculate adjusted debit, credit, and resulting balance
		debit, credit, balance := e.usecase.Adjustment(*account, transaction.Debit, transaction.Credit, generalLedger.Balance)

		// --- SUB-STEP 7D: GENERAL LEDGER ENTRY PREPARATION ---
		// Prepare new general ledger entry with all transaction details
		var paymentTypeValue core.TypeOfPaymentType
		if paymentType != nil {
			paymentTypeValue = paymentType.Type
		}

		newGeneralLedger := &core.GeneralLedger{
			CreatedAt:                  now,
			CreatedByID:                userOrg.UserID,
			UpdatedAt:                  now,
			UpdatedByID:                userOrg.UserID,
			BranchID:                   *userOrg.BranchID,
			OrganizationID:             userOrg.OrganizationID,
			TransactionBatchID:         &transactionBatch.ID,
			ReferenceNumber:            transaction.ReferenceNumber,
			EntryDate:                  transaction.EntryDate,
			SignatureMediaID:           transaction.SignatureMediaID,
			ProofOfPaymentMediaID:      transaction.ProofOfPaymentMediaID,
			BankID:                     transaction.BankID,
			AccountID:                  &transaction.AccountID,
			MemberProfileID:            &memberProfile.ID,
			PaymentTypeID:              transaction.PaymentTypeID,
			TransactionReferenceNumber: transaction.ReferenceNumber,
			Source:                     source,
			BankReferenceNumber:        transaction.BankReferenceNumber,
			EmployeeUserID:             &userOrg.UserID,
			Description:                transaction.Description,
			TypeOfPaymentType:          paymentTypeValue,
			Credit:                     credit,
			Debit:                      debit,
			Balance:                    balance,
			CurrencyID:                 account.CurrencyID,
		}

		// --- SUB-STEP 7E: GENERAL LEDGER ENTRY CREATION ---
		// Create the general ledger entry in the database
		if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, newGeneralLedger); err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "member-ledger-creation-failed",
				Description: "Failed to create member general ledger entry for account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to create general ledger entry"))
		}

		// --- SUB-STEP 7F: MEMBER ACCOUNTING LEDGER UPDATE ---
		// Update or create member accounting ledger with new balance
		_, err = e.core.MemberAccountingLedgerUpdateOrCreate(
			context,
			tx,
			*transaction.MemberProfileID,
			transaction.AccountID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			userOrg.UserID,
			balance,
			now,
		)
		if err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "member-accounting-ledger-update-failed",
				Description: "Failed to update member accounting ledger for member " + transaction.MemberProfileID.String() + " on account " + transaction.AccountID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to update member accounting ledger"))
		}

		// Log successful member transaction completion
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "member-transaction-completed",
			Description: "Successfully recorded member transaction for " + memberProfile.ID.String() + " with balance: " + fmt.Sprintf("%.2f", balance),
			Module:      "Transaction Recording",
		})

	} else {
		// ================================================================================
		// STEP 8: TRANSACTION PROCESSING - SUBSIDIARY ACCOUNT PATH
		// ================================================================================
		// --- SUB-STEP 8A: SUBSIDIARY LEDGER RETRIEVAL ---
		// For organization/subsidiary accounts (non-member transactions)
		generalLedger, err := e.core.GeneralLedgerCurrentSubsidiaryAccountForUpdate(
			context, tx, account.ID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "subsidiary-ledger-lock-failed",
				Description: "Failed to lock subsidiary ledger for account " + account.ID.String() + " in organization " + userOrg.OrganizationID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to retrieve subsidiary general ledger"))
		}

		// --- SUB-STEP 8B: BALANCE CALCULATION ---
		// Calculate adjusted debit, credit, and resulting balance for subsidiary account
		debit, credit, balance := e.usecase.Adjustment(*generalLedger.Account, transaction.Debit, transaction.Credit, generalLedger.Balance)

		// --- SUB-STEP 8C: SUBSIDIARY LEDGER ENTRY PREPARATION ---
		// Prepare new subsidiary general ledger entry
		var paymentTypeValue core.TypeOfPaymentType
		if paymentType != nil {
			paymentTypeValue = paymentType.Type
		}

		newGeneralLedger := &core.GeneralLedger{
			CreatedAt:             now,
			CreatedByID:           userOrg.UserID,
			UpdatedAt:             now,
			UpdatedByID:           userOrg.UserID,
			BranchID:              *userOrg.BranchID,
			OrganizationID:        userOrg.OrganizationID,
			TransactionBatchID:    &transactionBatch.ID,
			ReferenceNumber:       transaction.ReferenceNumber,
			EntryDate:             transaction.EntryDate,
			SignatureMediaID:      transaction.SignatureMediaID,
			ProofOfPaymentMediaID: transaction.ProofOfPaymentMediaID,
			BankID:                transaction.BankID,
			AccountID:             &transaction.AccountID,

			PaymentTypeID:              transaction.PaymentTypeID,
			TransactionReferenceNumber: transaction.ReferenceNumber,
			Source:                     source,
			BankReferenceNumber:        transaction.BankReferenceNumber,
			EmployeeUserID:             &userOrg.UserID,
			Description:                transaction.Description,
			TypeOfPaymentType:          paymentTypeValue,
			Credit:                     credit,
			Debit:                      debit,
			Balance:                    balance,
			CurrencyID:                 account.CurrencyID,
		}

		// --- SUB-STEP 8D: SUBSIDIARY LEDGER ENTRY CREATION ---
		// Create the subsidiary general ledger entry in the database
		if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, newGeneralLedger); err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "subsidiary-ledger-creation-failed",
				Description: "Failed to create subsidiary general ledger entry for account " + account.ID.String() + " in organization " + userOrg.OrganizationID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to create general ledger entry"))
		}

		// Log successful subsidiary transaction completion
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "subsidiary-transaction-completed",
			Description: "Successfully recorded subsidiary transaction for account " + account.ID.String() + " with balance: " + fmt.Sprintf("%.2f", balance),
			Module:      "Transaction Recording",
		})
	}

	// ================================================================================
	// STEP 9: TRANSACTION COMPLETION
	// ================================================================================
	// Log overall transaction success
	e.Footstep(echoCtx, FootstepEvent{
		Activity:    "transaction-recording-completed",
		Description: "Transaction recording completed successfully for reference: " + transaction.ReferenceNumber + " with source: " + string(source),
		Module:      "Transaction Recording",
	})

	// Commit the database transaction
	return endTx(nil)

}
