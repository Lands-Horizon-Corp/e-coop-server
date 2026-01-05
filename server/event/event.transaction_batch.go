package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (m *Event) TransactionBatchBalancing(context context.Context, transactionBatchID *uuid.UUID) error {

	if transactionBatchID == nil {
		return eris.New("transactionBatchID is nil")
	}
	tx, endTx := m.provider.Service.Database.StartTransaction(context)
	transactionBatch, err := m.core.TransactionBatchManager().GetByIDLock(context, tx, *transactionBatchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to get transaction batch by ID"))
	}
	/*
		deposit_in_bank decimal // too lazy to cash count, just know the total
		cash_count_total decimal // cash count total
		grand_total decimal // cash_count + deposit in bank

		// LESS
		petty_cash decimal // disbursement, commercial check, transfer to rf
		loan_releases decimal
		time_deposit_withdrawal decimal // WTF unknown
		savings_withdrawal decimal
		// end LESS

		total_online_remittance decimal // input sa online
		total_deposit_in_bank decimal
		total_actual_remittance decimal // total_check_remitance + total_online_remittance + total_cash_on_hand + total_deposit_in_bank
		total_actual_supposed_comparison decimal // abs(total supposed remittance) + abs(total actual remittance)
	*/

	payments, err := m.TBPayment(context, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total payments"))
	}
	transactionBatch.TotalCashCollection = payments.Balance

	totalDeposit, err := m.TBDeposit(context, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total deposits"))
	}
	transactionBatch.TotalDepositEntry = totalDeposit.Balance

	batchFunding, err := m.TBBatchFunding(context, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total batch funding"))
	}
	transactionBatch.BeginningBalance = batchFunding

	// total_cash_handled  = beg_bal + deposit + collection (REAL CASH ONLY)
	transactionBatch.TotalCashHandled = m.provider.Service.Decimal.Add(
		m.provider.Service.Decimal.Add(
			transactionBatch.BeginningBalance,
			transactionBatch.DepositInBank,
		),
		transactionBatch.TotalCashCollection,
	)
	totalWithdraw, err := m.TBWithdraw(context, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total withdrawals"))
	}

	less := 0.0
	transactionBatch.SavingsWithdrawal = m.provider.Service.Decimal.Multiply(totalWithdraw.Balance, -1)

	less = m.provider.Service.Decimal.Add(less, transactionBatch.SavingsWithdrawal)
	disbursementTransaction, err := m.TBDisbursementTransaction(context, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total disbursement transactions"))
	}
	transactionBatch.PettyCash = disbursementTransaction
	less = m.provider.Service.Decimal.Add(less, transactionBatch.PettyCash)

	transactionBatch.TotalSupposedRemmitance = m.provider.Service.Decimal.Subtract(
		transactionBatch.TotalCashHandled,
		less,
	)
	tbCashCount, err := m.TBCashCount(context, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total cash count"))
	}
	// total_cash_on_hand decimal
	transactionBatch.TotalCashOnHand = tbCashCount
	transactionBatch.CashCountTotal = tbCashCount

	// total_check_remittance decimal // input sa check remitance module
	tbCheckRemittance, err := m.TBCheckRemittance(context, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total check remittance"))
	}
	transactionBatch.TotalCheckRemittance = tbCheckRemittance

	tbOnlineRemittance, err := m.TBOnlineRemittance(context, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total online remittance"))
	}

	transactionBatch.TotalOnlineRemittance = tbOnlineRemittance
	transactionBatch.TotalDepositInBank = transactionBatch.DepositInBank

	// GrandTotal
	transactionBatch.GrandTotal = m.provider.Service.Decimal.Add(tbCashCount, transactionBatch.DepositInBank)
	transactionBatch.GrandTotal = m.provider.Service.Decimal.Add(transactionBatch.GrandTotal, batchFunding)

	transactionBatch.TotalActualRemittance = m.provider.Service.Decimal.Add(
		m.provider.Service.Decimal.Add(
			transactionBatch.TotalCheckRemittance,
			transactionBatch.TotalOnlineRemittance,
		),
		m.provider.Service.Decimal.Add(
			transactionBatch.TotalCashOnHand,
			transactionBatch.TotalDepositInBank,
		),
	)

	// LoanReleases
	// TimeDepositWithdrawal
	transactionBatch.TotalActualSupposedComparison = m.provider.Service.Decimal.Subtract(
		transactionBatch.TotalActualRemittance,
		transactionBatch.TotalSupposedRemmitance,
	)
	if err := m.core.TransactionBatchManager().UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
		return endTx(eris.Wrap(err, "failed to update transaction batch"))
	}
	if err := endTx(nil); err != nil {
		return eris.Wrap(err, "failed to end transaction")
	}
	return nil
}

func (m *Event) TBBatchFunding(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (float64, error) {
	batchFunding, err := m.core.BatchFundingManager().Find(context, &core.BatchFunding{
		TransactionBatchID: transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
	})
	if err != nil {
		return 0, eris.Wrap(err, "failed to find batch fundings")
	}
	var totalBatchFunding float64
	for _, funding := range batchFunding {
		totalBatchFunding = m.provider.Service.Decimal.Add(totalBatchFunding, funding.Amount)
	}
	return totalBatchFunding, nil
}

func (m *Event) TBCashCount(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (float64, error) {
	cashCounts, err := m.core.CashCountManager().Find(context, &core.CashCount{
		TransactionBatchID: transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
	})
	if err != nil {
		return 0, eris.Wrap(err, "failed to find cash counts")
	}
	var totalCashCount float64
	for _, cashCount := range cashCounts {
		totalCashCount = m.provider.Service.Decimal.Add(totalCashCount, m.provider.Service.Decimal.Multiply(cashCount.Amount, float64(cashCount.Quantity)))
	}
	return totalCashCount, nil
}

func (m *Event) TBCheckRemittance(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (float64, error) {
	checkRemittances, err := m.core.CheckRemittanceManager().Find(context, &core.CheckRemittance{
		TransactionBatchID: &transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
	})
	if err != nil {
		return 0, eris.Wrap(err, "failed to find check remittances")
	}
	var totalCheckRemittance float64
	for _, remittance := range checkRemittances {
		totalCheckRemittance = m.provider.Service.Decimal.Add(totalCheckRemittance, remittance.Amount)
	}
	return totalCheckRemittance, nil
}

func (m *Event) TBOnlineRemittance(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (float64, error) {
	onlineRemittances, err := m.core.OnlineRemittanceManager().Find(context, &core.OnlineRemittance{
		TransactionBatchID: &transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
	})
	if err != nil {
		return 0, eris.Wrap(err, "failed to find online remittances")
	}
	var totalOnlineRemittance float64
	for _, remittance := range onlineRemittances {
		totalOnlineRemittance = m.provider.Service.Decimal.Add(totalOnlineRemittance, remittance.Amount)
	}
	return totalOnlineRemittance, nil
}

func (m *Event) TBDisbursementTransaction(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (float64, error) {
	disbursementTransactions, err := m.core.DisbursementTransactionManager().Find(context, &core.DisbursementTransaction{
		TransactionBatchID: transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
	})
	if err != nil {
		return 0, eris.Wrap(err, "failed to find disbursement transactions")
	}
	var totalDisbursementTransaction float64
	for _, transaction := range disbursementTransactions {
		totalDisbursementTransaction = m.provider.Service.Decimal.Add(totalDisbursementTransaction, transaction.Amount)
	}
	return totalDisbursementTransaction, nil
}

func (m *Event) TBWithdraw(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	withdrawals, err := m.core.GeneralLedgerManager().Find(context, &core.GeneralLedger{
		TransactionBatchID: &transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
		Source:             core.GeneralLedgerSourceWithdraw,
	})
	if err != nil {
		return usecase.BalanceResponse{}, eris.Wrap(err, "failed to find withdrawals")
	}
	return m.usecase.Balance(usecase.Balance{
		GeneralLedgers: withdrawals,
	})
}

func (m *Event) TBDeposit(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	deposits, err := m.core.GeneralLedgerManager().Find(context, &core.GeneralLedger{
		TransactionBatchID: &transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
		Source:             core.GeneralLedgerSourceDeposit,
	})
	if err != nil {
		return usecase.BalanceResponse{}, eris.Wrap(err, "failed to find deposits")
	}
	return m.usecase.Balance(usecase.Balance{
		GeneralLedgers: deposits,
	})
}

func (m *Event) TBJournal(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	journals, err := m.core.GeneralLedgerManager().Find(context, &core.GeneralLedger{
		TransactionBatchID: &transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
		Source:             core.GeneralLedgerSourceJournal,
	})
	if err != nil {
		return usecase.BalanceResponse{}, eris.Wrap(err, "failed to find journals")
	}
	return m.usecase.Balance(usecase.Balance{
		GeneralLedgers: journals,
	})
}

func (m *Event) TBPayment(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	payments, err := m.core.GeneralLedgerManager().Find(context, &core.GeneralLedger{
		TransactionBatchID: &transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
		Source:             core.GeneralLedgerSourcePayment,
	})
	if err != nil {
		return usecase.BalanceResponse{}, eris.Wrap(err, "failed to find payments")
	}
	return m.usecase.Balance(usecase.Balance{
		GeneralLedgers: payments,
	})
}

func (m *Event) TBAdjustment(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	adjustments, err := m.core.GeneralLedgerManager().Find(context, &core.GeneralLedger{
		TransactionBatchID: &transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
		Source:             core.GeneralLedgerSourceAdjustment,
	})
	if err != nil {
		return usecase.BalanceResponse{}, eris.Wrap(err, "failed to find adjustments")
	}
	return m.usecase.Balance(usecase.Balance{
		GeneralLedgers: adjustments,
	})
}

func (m *Event) TBJournalVoucher(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	journalVouchers, err := m.core.GeneralLedgerManager().Find(context, &core.GeneralLedger{
		TransactionBatchID: &transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
		Source:             core.GeneralLedgerSourceJournalVoucher,
	})
	if err != nil {
		return usecase.BalanceResponse{}, eris.Wrap(err, "failed to find journal vouchers")
	}
	return m.usecase.Balance(usecase.Balance{
		GeneralLedgers: journalVouchers,
	})
}

func (m *Event) TBCheckVoucher(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	checkVouchers, err := m.core.GeneralLedgerManager().Find(context, &core.GeneralLedger{
		TransactionBatchID: &transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
		Source:             core.GeneralLedgerSourceCheckVoucher,
	})
	if err != nil {
		return usecase.BalanceResponse{}, eris.Wrap(err, "failed to find check vouchers")
	}
	return m.usecase.Balance(usecase.Balance{
		GeneralLedgers: checkVouchers,
	})
}

func (m *Event) TBLoan(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	loanVouchers, err := m.core.GeneralLedgerManager().Find(context, &core.GeneralLedger{
		TransactionBatchID: &transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
		Source:             core.GeneralLedgerSourceLoan,
	})
	if err != nil {
		return usecase.BalanceResponse{}, eris.Wrap(err, "failed to find loan vouchers")
	}
	return m.usecase.Balance(usecase.Balance{
		GeneralLedgers: loanVouchers,
	})
}
