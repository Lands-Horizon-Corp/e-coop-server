package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

type RecordTransactionRequest struct {
	Debit  float64
	Credit float64

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

func RecordTransaction(
	context context.Context, service *horizon.HorizonService,
	transaction RecordTransactionRequest,
	source types.GeneralLedgerSource,
	userOrg *types.UserOrganization,
) error {

	tx, endTx := service.Database.StartTransaction(context)
	if transaction.Credit == 0 && transaction.Debit == 0 {
		return endTx(eris.New("both credit and debit cannot be zero"))
	}
	if transaction.AccountID == uuid.Nil {
		return endTx(eris.New("account ID is required"))
	}
	if transaction.ReferenceNumber == "" {
		return endTx(eris.New("reference number is required"))
	}
	if userOrg == nil {
		return endTx(eris.New("user organization is nil"))
	}
	now := time.Now().UTC()
	timeMachine := userOrg.TimeMachine()

	if userOrg.BranchID == nil {
		return endTx(eris.New("user organization branch ID is nil"))
	}
	account, err := core.AccountLockForUpdate(context, service, tx, transaction.AccountID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to lock account for update"))
	}

	var paymentType *types.PaymentType
	if transaction.PaymentTypeID != nil {
		paymentType, err = core.PaymentTypeManager(service).GetByID(context, *transaction.PaymentTypeID)
		if err != nil {
			return endTx(eris.Wrap(err, "failed to resolve payment type"))
		}
	}

	var memberProfile *types.MemberProfile
	if transaction.MemberProfileID != nil {
		var err error
		memberProfile, err = core.MemberProfileManager(service).GetByID(context, *transaction.MemberProfileID)
		if err != nil {
			return endTx(eris.Wrap(err, "failed to retrieve member profile"))
		}

		if memberProfile == nil {
			return endTx(eris.New("member profile not found"))
		}
	}

	var loanTransactionID *uuid.UUID
	if transaction.LoanTransactionID != nil {
		loanTransactionID = transaction.LoanTransactionID

		loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, *transaction.LoanTransactionID)
		if err != nil {

			return endTx(eris.Wrap(err, "failed to retrieve loan transaction"))
		}

		if loanTransaction == nil {
			return endTx(eris.New("loan transaction not found"))
		}
	}

	var paymentTypeValue types.TypeOfPaymentType
	if paymentType != nil {
		paymentTypeValue = paymentType.Type
	}

	newGeneralLedger := &types.GeneralLedger{
		CreatedAt:                  now,
		CreatedByID:                userOrg.UserID,
		UpdatedAt:                  now,
		UpdatedByID:                userOrg.UserID,
		BranchID:                   *userOrg.BranchID,
		OrganizationID:             userOrg.OrganizationID,
		TransactionBatchID:         &transaction.TransactionBatchID,
		ReferenceNumber:            transaction.ReferenceNumber,
		EntryDate:                  timeMachine,
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

	if err := core.CreateGeneralLedgerEntry(context, service, tx, newGeneralLedger); err != nil {
		return endTx(eris.Wrap(err, "failed to create general ledger entry"))
	}

	if loanTransactionID != nil {
		loanAccount, err := core.GetLoanAccountByLoanTransaction(
			context, service,
			tx,
			*loanTransactionID,
			account.ID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			return endTx(eris.Wrap(err, "failed to retrieve loan account"))
		}

		if loanAccount == nil {
			return endTx(eris.New("loan account not found"))
		}

		if transaction.Credit > 0 {
			loanAccount.TotalPaymentCount += 1
			totalPaymentDec := decimal.NewFromFloat(loanAccount.TotalPayment)
			creditDec := decimal.NewFromFloat(transaction.Credit)
			loanAccount.TotalPayment = totalPaymentDec.Add(creditDec).InexactFloat64()
		}

		if transaction.Debit > 0 {
			loanAccount.TotalDeductionCount += 1
			totalDeductionDec := decimal.NewFromFloat(loanAccount.TotalDeduction)
			debitDec := decimal.NewFromFloat(transaction.Debit)
			loanAccount.TotalDeduction = totalDeductionDec.Add(debitDec).InexactFloat64()
		}

		loanAccount.UpdatedByID = userOrg.UserID
		loanAccount.UpdatedAt = now

		if err := core.LoanAccountManager(service).UpdateByIDWithTx(context, tx, loanAccount.ID, loanAccount); err != nil {
			return endTx(eris.Wrap(err, "failed to update loan account"))
		}
	}
	return endTx(nil)

}
