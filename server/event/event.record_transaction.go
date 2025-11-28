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
	LoanTransactionID     *uuid.UUID `json:"loan_transaction_id"`
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
	// STEP 7: TRANSACTION PROCESSING
	// ================================================================================

	// Validate member profile if provided
	var memberProfile *core.MemberProfile
	if transaction.MemberProfileID != nil {
		var err error
		memberProfile, err = e.core.MemberProfileManager.GetByID(context, *transaction.MemberProfileID)
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
	}

	// Validate loan transaction if provided
	var loanTransactionID *uuid.UUID
	if transaction.LoanTransactionID != nil {
		loanTransactionID = transaction.LoanTransactionID

		loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, *transaction.LoanTransactionID)
		if err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "loan-transaction-retrieval-failed",
				Description: "Failed to retrieve loan transaction " + transaction.LoanTransactionID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to retrieve loan transaction"))
		}

		if loanTransaction == nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "loan-transaction-not-found",
				Description: "Loan transaction not found for ID: " + transaction.LoanTransactionID.String(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.New("loan transaction not found"))
		}
	}

	// Prepare payment type
	var paymentTypeValue core.TypeOfPaymentType
	if paymentType != nil {
		paymentTypeValue = paymentType.Type
	}

	// Prepare entry date
	userOrgTime := userOrg.UserOrgTime()
	if transaction.EntryDate != nil {
		userOrgTime = *transaction.EntryDate
	}

	// Create general ledger entry (handles both member and subsidiary accounts)
	newGeneralLedger := &core.GeneralLedger{
		CreatedAt:                  now,
		CreatedByID:                userOrg.UserID,
		UpdatedAt:                  now,
		UpdatedByID:                userOrg.UserID,
		BranchID:                   *userOrg.BranchID,
		OrganizationID:             userOrg.OrganizationID,
		TransactionBatchID:         &transaction.TransactionBatchID,
		ReferenceNumber:            transaction.ReferenceNumber,
		EntryDate:                  userOrgTime,
		SignatureMediaID:           transaction.SignatureMediaID,
		ProofOfPaymentMediaID:      transaction.ProofOfPaymentMediaID,
		BankID:                     transaction.BankID,
		AccountID:                  &transaction.AccountID,
		MemberProfileID:            transaction.MemberProfileID,
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
		Account:                    account,
	}

	if err := e.core.CreateGeneralLedgerEntry(context, tx, newGeneralLedger); err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "ledger-creation-failed",
			Description: "Failed to create general ledger entry for account " + account.ID.String() + ": " + err.Error(),
			Module:      "Transaction Recording",
		})
		return endTx(eris.Wrap(err, "failed to create general ledger entry"))
	}

	// Handle loan account updates if loan transaction exists
	if loanTransactionID != nil {
		loanAccount, err := e.core.GetLoanAccountByLoanTransaction(
			context,
			tx,
			*loanTransactionID,
			account.ID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "loan-account-retrieval-failed",
				Description: "Failed to retrieve loan account for loan transaction " + loanTransactionID.String() + ": " + err.Error(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to retrieve loan account"))
		}

		if loanAccount == nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "loan-account-not-found",
				Description: "Loan account not found for loan transaction ID: " + loanTransactionID.String(),
				Module:      "Transaction Recording",
			})
			return endTx(eris.New("loan account not found"))
		}

		// Handle Credit (Payment)
		if transaction.Credit > 0 {
			loanAccount.TotalPaymentCount += 1
			loanAccount.TotalPayment = e.provider.Service.Decimal.Add(
				loanAccount.TotalPayment, transaction.Credit)
		}

		// Handle Debit (Deduction/Add-on)
		if transaction.Debit > 0 {
			loanAccount.TotalDeductionCount += 1
			loanAccount.TotalDeduction = e.provider.Service.Decimal.Add(
				loanAccount.TotalDeduction, transaction.Debit)
		}
		loanAccount.UpdatedByID = userOrg.UserID
		loanAccount.UpdatedAt = now

		if err := e.core.LoanAccountManager.UpdateByIDWithTx(context, tx, loanAccount.ID, loanAccount); err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity: "loan-account-update-failed",
				Description: "Failed to update loan account " +
					loanAccount.ID.String() + ": " + err.Error(),
				Module: "Transaction Recording",
			})
			return endTx(eris.Wrap(err, "failed to update loan account"))
		}
	}

	// Log successful transaction completion
	transactionType := "subsidiary"
	if memberProfile != nil {
		transactionType = "member"
	}
	e.Footstep(echoCtx, FootstepEvent{
		Activity: "transaction-completed",
		Description: "Successfully recorded " + transactionType + " transaction for account " +
			account.ID.String(),
		Module: "Transaction Recording",
	})

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
