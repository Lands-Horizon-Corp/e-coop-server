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
	fmt.Printf("DEBUG LINE 132: Retrieving transaction batch for user: %v, org: %v, branch: %v\n", userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
	transactionBatch, err := e.core.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
	if err != nil {
		fmt.Printf("DEBUG LINE 135: Failed to retrieve transaction batch: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "batch-retrieval-failed",
			Description: "Cannot retrieve current transaction batch for user " + userOrg.UserID.String() + ": " + err.Error(),
			Module:      "Transaction Recording",
		})
		return endTx(eris.Wrap(err, "failed to retrieve transaction batch"))
	}
	fmt.Printf("DEBUG LINE 143: Transaction batch retrieved successfully\n")

	// Ensure a valid transaction batch exists
	if transactionBatch == nil {
		fmt.Printf("DEBUG LINE 147: Transaction batch is nil\n")
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "batch-missing",
			Description: "No active transaction batch found for current user session - batch is required for transaction recording",
			Module:      "Transaction Recording",
		})
		return endTx(eris.New("transaction batch is nil"))
	}
	fmt.Printf("DEBUG LINE 155: Transaction batch is valid - ID: %v\n", transactionBatch.ID)

	// ================================================================================
	// STEP 5: ACCOUNT RESOLUTION AND LOCKING
	// ================================================================================
	// Lock the target account for update to prevent concurrent modifications
	fmt.Printf("DEBUG LINE 160: Attempting to lock account: %v\n", transaction.AccountID)
	account, err := e.core.AccountLockForUpdate(context, tx, transaction.AccountID)
	if err != nil {
		fmt.Printf("DEBUG LINE 163: Failed to lock account: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "account-lock-failed",
			Description: "Failed to lock account " + transaction.AccountID.String() + " for transaction update: " + err.Error(),
			Module:      "Transaction Recording",
		})
		return endTx(eris.Wrap(err, "failed to lock account for update"))
	}
	fmt.Printf("DEBUG LINE 171: Account locked successfully - Account ID: %v\n", account.ID)

	// ================================================================================
	// STEP 6: PAYMENT TYPE RESOLUTION
	// ================================================================================
	// Resolve payment type details if specified
	fmt.Printf("DEBUG LINE 176: Starting payment type resolution\n")
	var paymentType *core.PaymentType
	if transaction.PaymentTypeID != nil {
		fmt.Printf("DEBUG LINE 179: Payment type ID provided: %v\n", *transaction.PaymentTypeID)
		paymentType, err = e.core.PaymentTypeManager.GetByID(context, *transaction.PaymentTypeID)
		if err != nil {
			fmt.Printf("DEBUG LINE 182: Failed to resolve payment type: %v\n", err)
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "payment-type-resolution-failed",
				Description: "Failed to resolve payment type " + transaction.PaymentTypeID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to resolve payment type"))
		}
		fmt.Printf("DEBUG LINE 190: Payment type resolved successfully\n")
	} else {
		fmt.Printf("DEBUG LINE 192: No payment type ID provided\n")
	}
	// ================================================================================
	// STEP 7: TRANSACTION PROCESSING - MEMBER ACCOUNT PATH
	// ================================================================================
	fmt.Printf("DEBUG LINE 197: Checking member profile ID\n")
	if transaction.MemberProfileID != nil {
		fmt.Printf("DEBUG LINE 199: Member profile ID provided: %v\n", *transaction.MemberProfileID)
		// --- SUB-STEP 7A: MEMBER PROFILE VALIDATION ---
		// Retrieve and validate member profile for member-specific transactions
		fmt.Printf("DEBUG LINE 202: Retrieving member profile\n")
		memberProfile, err := e.core.MemberProfileManager.GetByID(context, *transaction.MemberProfileID)
		if err != nil {
			fmt.Printf("DEBUG LINE 205: Failed to retrieve member profile: %v\n", err)
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "member-profile-retrieval-failed",
				Description: "Failed to retrieve member profile " + transaction.MemberProfileID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to retrieve member profile"))
		}
		fmt.Printf("DEBUG LINE 213: Member profile retrieved successfully\n")

		if memberProfile == nil {
			fmt.Printf("DEBUG LINE 216: Member profile is nil\n")
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "member-profile-not-found",
				Description: "Member profile not found for ID: " + transaction.MemberProfileID.String(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.New("member profile not found"))
		}
		fmt.Printf("DEBUG LINE 224: Member profile validation completed - ID: %v\n", memberProfile.ID)

		// --- SUB-STEP 7B: MEMBER LEDGER RETRIEVAL ---
		// Get current member account ledger with row-level locking
		fmt.Printf("DEBUG LINE 227: Retrieving member ledger for update\n")
		generalLedger, err := e.core.GeneralLedgerCurrentMemberAccountForUpdate(
			context, tx, memberProfile.ID, account.ID, memberProfile.OrganizationID, memberProfile.BranchID,
		)
		if err != nil {
			fmt.Printf("DEBUG LINE 232: Failed to retrieve member ledger: %v\n", err)
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "member-ledger-lock-failed",
				Description: "Failed to lock member ledger for account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to retrieve member ledger for update"))
		}
		fmt.Printf("DEBUG LINE 240: Member ledger retrieved successfully - Balance: %f\n", generalLedger.Balance)

		// --- SUB-STEP 7C: BALANCE CALCULATION ---
		// Calculate adjusted debit, credit, and resulting balance
		fmt.Printf("DEBUG LINE 243: Calculating member balance adjustment\n")
		debit, credit, balance := e.usecase.Adjustment(*account, transaction.Debit, transaction.Credit, generalLedger.Balance)
		fmt.Printf("DEBUG LINE 245: Balance calculated - Debit: %f, Credit: %f, Balance: %f\n", debit, credit, balance)

		// --- SUB-STEP 7D: GENERAL LEDGER ENTRY PREPARATION ---
		// Prepare new general ledger entry with all transaction details
		fmt.Printf("DEBUG LINE 249: Preparing member general ledger entry\n")
		var paymentTypeValue core.TypeOfPaymentType
		if paymentType != nil {
			fmt.Printf("DEBUG LINE 252: Using payment type: %v\n", paymentType.Type)
			paymentTypeValue = paymentType.Type
		} else {
			fmt.Printf("DEBUG LINE 255: No payment type provided\n")
		}

		fmt.Printf("DEBUG LINE 258: Creating member general ledger struct\n")
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
		fmt.Printf("DEBUG LINE 283: Member general ledger struct created successfully\n")

		// --- SUB-STEP 7E: GENERAL LEDGER ENTRY CREATION ---
		// Create the general ledger entry in the database
		fmt.Printf("DEBUG LINE 286: Creating member general ledger entry in database\n")
		if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, newGeneralLedger); err != nil {
			fmt.Printf("DEBUG LINE 288: Failed to create member general ledger entry: %v\n", err)
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "member-ledger-creation-failed",
				Description: "Failed to create member general ledger entry for account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to create general ledger entry"))
		}
		fmt.Printf("DEBUG LINE 296: Member general ledger entry created successfully\n")

		// --- SUB-STEP 7F: MEMBER ACCOUNTING LEDGER UPDATE ---
		// Update or create member accounting ledger with new balance
		fmt.Printf("DEBUG LINE 299: Updating member accounting ledger\n")
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
			fmt.Printf("DEBUG LINE 312: Failed to update member accounting ledger: %v\n", err)
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "member-accounting-ledger-update-failed",
				Description: "Failed to update member accounting ledger for member " + transaction.MemberProfileID.String() + " on account " + transaction.AccountID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to update member accounting ledger"))
		}
		fmt.Printf("DEBUG LINE 320: Member accounting ledger updated successfully\n")

		// Log successful member transaction completion
		fmt.Printf("DEBUG LINE 323: Member transaction path completed successfully\n")
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "member-transaction-completed",
			Description: "Successfully recorded member transaction for " + memberProfile.ID.String() + " with balance: " + fmt.Sprintf("%.2f", balance),
			Module:      "Transaction Recording",
		})

	} else {
		fmt.Printf("DEBUG LINE 331: No member profile ID - processing subsidiary account\n")
		// ================================================================================
		// STEP 8: TRANSACTION PROCESSING - SUBSIDIARY ACCOUNT PATH
		// ================================================================================
		// --- SUB-STEP 8A: SUBSIDIARY LEDGER RETRIEVAL ---
		// For organization/subsidiary accounts (non-member transactions)
		fmt.Printf("DEBUG LINE 337: Retrieving subsidiary ledger for update\n")
		generalLedger, err := e.core.GeneralLedgerCurrentSubsidiaryAccountForUpdate(
			context, tx, account.ID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			fmt.Printf("DEBUG LINE 341: Failed to retrieve subsidiary ledger: %v\n", err)
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "subsidiary-ledger-lock-failed",
				Description: "Failed to lock subsidiary ledger for account " + account.ID.String() + " in organization " + userOrg.OrganizationID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to retrieve subsidiary general ledger"))
		}
		fmt.Printf("DEBUG LINE 349: Subsidiary ledger retrieved successfully - Balance: %f\n", generalLedger.Balance)

		// --- SUB-STEP 8B: BALANCE CALCULATION ---
		// Calculate adjusted debit, credit, and resulting balance for subsidiary account
		fmt.Printf("DEBUG LINE 352: Calculating subsidiary balance adjustment\n")
		debit, credit, balance := e.usecase.Adjustment(*account, transaction.Debit, transaction.Credit, generalLedger.Balance)
		fmt.Printf("DEBUG LINE 354: Subsidiary balance calculated - Debit: %f, Credit: %f, Balance: %f\n", debit, credit, balance)

		// --- SUB-STEP 8C: SUBSIDIARY LEDGER ENTRY PREPARATION ---
		// Prepare new subsidiary general ledger entry
		fmt.Printf("DEBUG LINE 358: Preparing subsidiary general ledger entry\n")
		var paymentTypeValue core.TypeOfPaymentType
		if paymentType != nil {
			fmt.Printf("DEBUG LINE 361: Using payment type for subsidiary: %v\n", paymentType.Type)
			paymentTypeValue = paymentType.Type
		} else {
			fmt.Printf("DEBUG LINE 364: No payment type for subsidiary\n")
		}

		fmt.Printf("DEBUG LINE 367: Creating subsidiary general ledger struct\n")
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
		fmt.Printf("DEBUG LINE 390: Subsidiary general ledger struct created successfully\n")

		// --- SUB-STEP 8D: SUBSIDIARY LEDGER ENTRY CREATION ---
		// Create the subsidiary general ledger entry in the database
		fmt.Printf("DEBUG LINE 393: Creating subsidiary general ledger entry in database\n")
		if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, newGeneralLedger); err != nil {
			fmt.Printf("DEBUG LINE 395: Failed to create subsidiary general ledger entry: %v\n", err)
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "subsidiary-ledger-creation-failed",
				Description: "Failed to create subsidiary general ledger entry for account " + account.ID.String() + " in organization " + userOrg.OrganizationID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to create general ledger entry"))
		}
		fmt.Printf("DEBUG LINE 403: Subsidiary general ledger entry created successfully\n")

		// Log successful subsidiary transaction completion
		fmt.Printf("DEBUG LINE 406: Subsidiary transaction path completed successfully\n")
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "subsidiary-transaction-completed",
			Description: "Successfully recorded subsidiary transaction for account " + account.ID.String() + " with balance: " + fmt.Sprintf("%.2f", balance),
			Module:      "Transaction Recording",
		})
	}
	fmt.Printf("DEBUG LINE 413: Transaction processing completed, proceeding to finalization\n")

	// ================================================================================
	// STEP 9: TRANSACTION COMPLETION
	// ================================================================================
	// Log overall transaction success
	fmt.Printf("DEBUG LINE 418: Recording final footstep\n")
	e.Footstep(echoCtx, FootstepEvent{
		Activity:    "transaction-recording-completed",
		Description: "Transaction recording completed successfully for reference: " + transaction.ReferenceNumber + " with source: " + string(source),
		Module:      "Transaction Recording",
	})
	fmt.Printf("DEBUG LINE 425: Final footstep recorded\n")

	// Commit the database transaction
	fmt.Printf("DEBUG LINE 428: Committing transaction\n")
	return endTx(nil)

}
