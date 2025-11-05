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

// TODO: posting loan, posting cash check voucher, posting journal voucher, adjustment entries
func (e Event) RecordTransaction(
	context context.Context,
	echoCtx echo.Context,
	transaction RecordTransactionRequest,
	source core.GeneralLedgerSource,
) error {
	now := time.Now().UTC()

	// Start transaction
	tx, endTx := e.provider.Service.Database.StartTransaction(context)
	if transaction.Credit == 0 && transaction.Debit == 0 {
		return endTx(eris.New("both credit and debit cannot be zero"))
	}

	if transaction.AccountID == uuid.Nil {
		return endTx(eris.New("account ID is required"))
	}
	if transaction.ReferenceNumber == "" {
		return endTx(eris.New("reference number is required"))
	}

	// ================================================================================
	// STEP 3: GET CURRENT USER ORGANIZATION
	// ================================================================================
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(context, echoCtx)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return endTx(eris.Wrap(err, "failed to get user organization"))
	}
	if userOrg == nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "User organization is nil (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		return endTx(eris.New("user organization is nil"))
	}
	if userOrg.BranchID == nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "branch-error",
			Description: "User organization branch ID is nil (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		return endTx(eris.New("user organization branch ID is nil"))
	}
	// ================================================================================
	// STEP 4: CURRENT TRANSACTION BATCH
	// ================================================================================
	transactionBatch, err := e.core.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "batch-error",
			Description: "Failed to retrieve transaction batch (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return endTx(eris.Wrap(err, "failed to retrieve transaction batch"))
	}
	if transactionBatch == nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "batch-error",
			Description: "Transaction batch is nil (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		return endTx(eris.New("transaction batch is nil"))
	}

	// Find account
	account, err := e.core.AccountLockForUpdate(context, tx, transaction.AccountID)
	if err != nil {
		return endTx(err)
	}

	var paymentType *core.PaymentType
	if transaction.PaymentTypeID != nil {
		paymentType, err = e.core.PaymentTypeManager.GetByID(context, *transaction.PaymentTypeID)
		if err != nil {
			return endTx(err)
		}
	}
	if transaction.MemberProfileID != nil {
		memberProfile, err := e.core.MemberProfileManager.GetByID(context, *transaction.MemberProfileID)
		if err != nil {
			return endTx(err)
		}
		if memberProfile == nil {
			return endTx(eris.New("member profile not found"))
		}

		generalLedger, err := e.core.GeneralLedgerCurrentMemberAccountForUpdate(
			context, tx, memberProfile.ID, account.ID, memberProfile.OrganizationID, memberProfile.BranchID,
		)
		debit, credit, balance := e.usecase.Adjustment(*generalLedger.Account, transaction.Debit, transaction.Credit, generalLedger.Balance)
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
			TypeOfPaymentType:          paymentType.Type,
			Credit:                     credit,
			Debit:                      debit,
			Balance:                    balance,
			CurrencyID:                 account.CurrencyID,
		}
		if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, newGeneralLedger); err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "ledger-create-error",
				Description: "Failed to create general ledger entry (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return endTx(eris.Wrap(err, "failed to create general ledger entry"))
		}
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
				Activity:    "member-accounting-ledger-error",
				Description: "Failed to update member accounting ledger (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return endTx(eris.Wrap(err, "failed to update member accounting ledger"))
		}

	} else {
		// Subsidiary ledger for organization account'
		generalLedger, err := e.core.GeneralLedgerCurrentSubsidiaryAccountForUpdate(
			context, tx, account.ID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "ledger-error",
				Description: "Failed to retrieve subsidiary general ledger (FOR UPDATE) (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return endTx(eris.Wrap(err, "failed to retrieve subsidiary general ledger"))
		}
		debit, credit, balance := e.usecase.Adjustment(*generalLedger.Account, transaction.Debit, transaction.Credit, generalLedger.Balance)
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
			TypeOfPaymentType:          paymentType.Type,
			Credit:                     credit,
			Debit:                      debit,
			Balance:                    balance,
			CurrencyID:                 account.CurrencyID,
		}
		if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, newGeneralLedger); err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "ledger-create-error",
				Description: "Failed to create general ledger entry (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return endTx(eris.Wrap(err, "failed to create general ledger entry"))
		}

	}

	return endTx(nil)

}
