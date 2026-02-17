package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func TransactionBatchHistoryTotalSummary(
	context context.Context,
	service *horizon.HorizonService,
	userOrg *types.UserOrganization,
	transactionBatchID *uuid.UUID,
) (*types.TransactionBatchHistoryTotal, error) {
	transactionBatch, err := core.TransactionBatchManager(service).GetByID(context, *transactionBatchID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve current transaction batch for summary")
	}
	funding, _ := TBBatchFunding(context, service, transactionBatch.ID, userOrg.OrganizationID, *userOrg.BranchID)
	disbursement, _ := TBDisbursementTransaction(context, service, transactionBatch.ID, userOrg.OrganizationID, *userOrg.BranchID)
	glEntries, err := core.GeneralLedgerManager(service).Find(context, &types.GeneralLedger{
		TransactionBatchID: &transactionBatch.ID,
		OrganizationID:     userOrg.OrganizationID,
		BranchID:           *userOrg.BranchID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to fetch ledger entries for batch history")
	}
	summary := &types.TransactionBatchHistoryTotal{
		BatchFundingTotal:            funding,
		DisbursementTransactionTotal: disbursement,
	}
	for _, gl := range glEntries {
		summary.GeneralLedgerDebitTotal += gl.Debit
		summary.GeneralLedgerCreditTotal += gl.Credit
		switch gl.TypeOfPaymentType {
		case types.PaymentTypeCash:
			summary.CashEntryDebitTotal += gl.Debit
			summary.CashEntryCreditTotal += gl.Credit
		case types.PaymentTypeCheck:
			summary.CheckEntryDebitTotal += gl.Debit
			summary.CheckEntryCreditTotal += gl.Credit
		case types.PaymentTypeOnline:
			summary.OnlineEntryDebitTotal += gl.Debit
			summary.OnlineEntryCreditTotal += gl.Credit
		}
		switch gl.Source {
		case types.GeneralLedgerSourcePayment:
			summary.PaymentEntryDebitTotal += gl.Debit
			summary.PaymentEntryCreditTotal += gl.Credit
		case types.GeneralLedgerSourceWithdraw:
			summary.WithdrawEntryDebitTotal += gl.Debit
			summary.WithdrawEntryCreditTotal += gl.Credit
		case types.GeneralLedgerSourceDeposit:
			summary.DepositEntryDebitTotal += gl.Debit
			summary.DepositEntryCreditTotal += gl.Credit
		}
	}
	return summary, nil
}
