package event

import (
	"context"
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
func (e Event) RecordTransaction(
	context context.Context,
	echoCtx echo.Context,
	transaction RecordTransactionRequest,
	source core.GeneralLedgerSource,
) error {
	now := time.Now().UTC()
	// ================================================================================
	// STEP 1: DATABASE TRANSACTION INITIALIZATION
	// ================================================================================
	// Start database transaction to ensure atomicity of all operations
	tx, endTx := e.provider.Service.Database.StartTransaction(context)

	// ================================================================================
	// STEP 2: INPUT VALIDATION
	// ================================================================================
	// Validate that at least one amount (credit or debit) is provided
	if transaction.Credit == 0 && transaction.Debit == 0 {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "validation-error",
			Description: "Transaction validation failed: Both credit and debit amounts are zero",
			Module:      "Transaction Recording",
		})
		return endTx(eris.New("both credit and debit cannot be zero"))
	}

	// Validate required account ID
	if transaction.AccountID == uuid.Nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "validation-error",
			Description: "Transaction validation failed: Account ID is missing or invalid",
			Module:      "Transaction Recording",
		})
		return endTx(eris.New("account ID is required"))
	}

	// Validate required reference number
	if transaction.ReferenceNumber == "" {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "validation-error",
			Description: "Transaction validation failed: Reference number is missing",
			Module:      "Transaction Recording",
		})
		return endTx(eris.New("reference number is required"))
	}

	// ================================================================================
	// STEP 3: USER AUTHENTICATION & ORGANIZATION CONTEXT
	// ================================================================================
	// Retrieve the current user's organization context for transaction authorization
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(context, echoCtx)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "authentication-failed",
			Description: "Unable to retrieve user organization context for transaction recording: " + err.Error(),
			Module:      "Transaction Recording",
		})
		return endTx(eris.Wrap(err, "failed to get user organization"))
	}

	// Ensure user organization context exists
	if userOrg == nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "authentication-failed",
			Description: "User organization context is missing - cannot proceed with transaction recording",
			Module:      "Transaction Recording",
		})
		return endTx(eris.New("user organization is nil"))
	}

	// Validate branch assignment for transaction context
	if userOrg.BranchID == nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "branch-context-error",
			Description: "User is not assigned to any branch - branch context required for transaction recording",
			Module:      "Transaction Recording",
		})
		return endTx(eris.New("user organization branch ID is nil"))
	}
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
		var paymentTypeValue core.TypeOfPaymentType
		if paymentType != nil {
			paymentTypeValue = paymentType.Type
		}

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

		var loanTransactionID *uuid.UUID
		var adjustmentType *core.LoanAdjustmentType

		if generalLedger != nil && generalLedger.LoanTransactionID != nil {
			loanTransactionID = generalLedger.LoanTransactionID
			adjustmentType = generalLedger.LoanAdjustmentType
			loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
			if err != nil {
				e.Footstep(echoCtx, FootstepEvent{
					Activity:    "loan-transaction-retrieval-failed",
					Description: "Failed to retrieve loan transaction " + loanTransactionID.String() + ": " + err.Error(),
					Module:      "Transaction Recording",
				})
				return endTx(eris.Wrap(err, "failed to retrieve loan transaction"))
			}
			accountHistory, err := e.core.GetAccountHistoryLatestByTimeHistory(
				context,
				account.ID,
				account.OrganizationID,
				account.BranchID,
				loanTransaction.PrintedDate,
			)
			if err != nil {
				e.Footstep(echoCtx, FootstepEvent{
					Activity:    "account-history-retrieval-failed",
					Description: "Failed to retrieve account history for account " + account.ID.String() + " at time " + loanTransaction.PrintedDate.String() + ": " + err.Error(),
					Module:      "Transaction Recording",
				})
				return endTx(eris.Wrap(err, "failed to retrieve account history"))
			}
			if accountHistory != nil {
				account = e.core.AccountHistoryToModel(accountHistory)
			}
		}
		userOrgTime := userOrg.UserOrgTime()
		if transaction.EntryDate != nil {
			userOrgTime = *transaction.EntryDate
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
			EntryDate:                  &userOrgTime,
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
			Credit:                     transaction.Credit,
			Debit:                      transaction.Debit,
			CurrencyID:                 account.CurrencyID,
			LoanTransactionID:          loanTransactionID,
			LoanAdjustmentType:         adjustmentType,
		}

		if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, newGeneralLedger); err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "member-ledger-creation-failed",
				Description: "Failed to create member general ledger entry for account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to create general ledger entry"))
		}

		_, err = e.core.MemberAccountingLedgerUpdateOrCreate(
			context,
			tx,
			core.MemberAccountingLedgerUpdateOrCreateParams{
				MemberProfileID: memberProfile.ID,
				AccountID:       account.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				UserID:          userOrg.UserID,
				DebitAmount:     transaction.Debit,
				CreditAmount:    transaction.Credit,
				LastPayTime:     now,
			},
		)
		if err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity: "member-accounting-ledger-update-failed",
				Description: "Failed to update member accounting ledger for member " +
					transaction.MemberProfileID.String() + " on account " +
					transaction.AccountID.String() + ": " + err.Error(),
				Module: "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to update member accounting ledger"))
		}

		// Log successful member transaction completion
		e.Footstep(echoCtx, FootstepEvent{
			Activity: "member-transaction-completed",
			Description: "Successfully recorded member transaction for " +
				memberProfile.ID.String(),
			Module: "Transaction Recording",
		})

	} else {
		// ================================================================================
		// STEP 8: TRANSACTION PROCESSING - SUBSIDIARY ACCOUNT PATH
		// ================================================================================

		var paymentTypeValue core.TypeOfPaymentType
		if paymentType != nil {
			paymentTypeValue = paymentType.Type
		}

		userOrgTime := userOrg.UserOrgTime()
		if transaction.EntryDate != nil {
			userOrgTime = *transaction.EntryDate
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
			EntryDate:                  &userOrgTime,
			SignatureMediaID:           transaction.SignatureMediaID,
			ProofOfPaymentMediaID:      transaction.ProofOfPaymentMediaID,
			BankID:                     transaction.BankID,
			AccountID:                  &transaction.AccountID,
			PaymentTypeID:              transaction.PaymentTypeID,
			TransactionReferenceNumber: transaction.ReferenceNumber,
			Source:                     source,
			BankReferenceNumber:        transaction.BankReferenceNumber,
			EmployeeUserID:             &userOrg.UserID,
			Description:                transaction.Description,
			TypeOfPaymentType:          paymentTypeValue,
			Credit:                     transaction.Credit,
			Debit:                      transaction.Debit,
			CurrencyID:                 account.CurrencyID,
		}

		// --- SUB-STEP 8D: SUBSIDIARY LEDGER ENTRY CREATION ---
		// Create the subsidiary general ledger entry in the database
		if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, newGeneralLedger); err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity: "subsidiary-ledger-creation-failed",
				Description: "Failed to create subsidiary general ledger entry for account " +
					account.ID.String() + " in organization " + userOrg.OrganizationID.String() +
					": " + err.Error(),
				Module: "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to create general ledger entry"))
		}
		// Log successful subsidiary transaction completion
		e.Footstep(echoCtx, FootstepEvent{
			Activity: "subsidiary-transaction-completed",
			Description: "Successfully recorded subsidiary transaction for account " +
				account.ID.String(),
			Module: "Transaction Recording",
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
