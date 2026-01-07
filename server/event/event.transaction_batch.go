package event

import (
	"context"
	"fmt"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

func (m *Event) TransactionBatchBalancing(ctx context.Context, transactionBatchID *uuid.UUID) error {
	fmt.Println("Entering TransactionBatchBalancing")
	if transactionBatchID == nil {
		fmt.Println("Error: transactionBatchID is nil")
		return eris.New("transactionBatchID is nil")
	}
	fmt.Println("transactionBatchID:", *transactionBatchID)

	tx, endTx := m.provider.Service.Database.StartTransaction(ctx)
	fmt.Println("Transaction started")

	transactionBatch, err := m.core.TransactionBatchManager().GetByIDLock(ctx, tx, *transactionBatchID)
	if err != nil {
		fmt.Println("Error: failed to get transaction batch by ID", err)
		return endTx(eris.Wrap(err, "failed to get transaction batch by ID"))
	}
	fmt.Println("Successfully locked transaction batch:", transactionBatch.ID)

	payments, err := m.TBPayment(ctx, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		fmt.Println("Error: failed to calculate total payments", err)
		return endTx(eris.Wrap(err, "failed to calculate total payments"))
	}
	transactionBatch.TotalCashCollection = payments.Balance
	fmt.Println("TotalCashCollection (payments.Balance):", transactionBatch.TotalCashCollection)

	totalDeposit, err := m.TBDeposit(ctx, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		fmt.Println("Error: failed to calculate total deposits", err)
		return endTx(eris.Wrap(err, "failed to calculate total deposits"))
	}
	transactionBatch.TotalDepositEntry = totalDeposit.Balance
	fmt.Println("TotalDepositEntry (totalDeposit.Balance):", transactionBatch.TotalDepositEntry)

	batchFunding, err := m.TBBatchFunding(ctx, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		fmt.Println("Error: failed to calculate total batch funding", err)
		return endTx(eris.Wrap(err, "failed to calculate total batch funding"))
	}
	transactionBatch.BeginningBalance = batchFunding
	fmt.Println("BeginningBalance (batchFunding):", transactionBatch.BeginningBalance)

	// total_cash_handled  = beg_bal + deposit + collection (REAL CASH ONLY)
	totalCashHandled := decimal.NewFromFloat(transactionBatch.BeginningBalance).
		Add(decimal.NewFromFloat(transactionBatch.DepositInBank)).
		Add(decimal.NewFromFloat(transactionBatch.TotalCashCollection))
	transactionBatch.TotalCashHandled = totalCashHandled.InexactFloat64()
	fmt.Println("TotalCashHandled calculated:", transactionBatch.TotalCashHandled)

	totalWithdraw, err := m.TBWithdraw(ctx, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		fmt.Println("Error: failed to calculate total withdrawals", err)
		return endTx(eris.Wrap(err, "failed to calculate total withdrawals"))
	}
	fmt.Println("TotalWithdraw balance retrieved:", totalWithdraw.Balance)

	less := decimal.Zero
	savingsWithdrawal := decimal.NewFromFloat(totalWithdraw.Balance).Mul(decimal.NewFromInt(-1))
	transactionBatch.SavingsWithdrawal = savingsWithdrawal.InexactFloat64()
	less = less.Add(savingsWithdrawal)
	fmt.Println("SavingsWithdrawal (less):", transactionBatch.SavingsWithdrawal)

	disbursementTransaction, err := m.TBDisbursementTransaction(ctx, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		fmt.Println("Error: failed to calculate total disbursement transactions", err)
		return endTx(eris.Wrap(err, "failed to calculate total disbursement transactions"))
	}
	transactionBatch.PettyCash = disbursementTransaction
	less = less.Add(decimal.NewFromFloat(transactionBatch.PettyCash))
	fmt.Println("PettyCash (disbursementTransaction):", transactionBatch.PettyCash)
	fmt.Println("Total 'less' amount:", less.String())

	totalSupposed := decimal.NewFromFloat(transactionBatch.TotalCashHandled).Sub(less)
	transactionBatch.TotalSupposedRemmitance = totalSupposed.InexactFloat64()
	fmt.Println("TotalSupposedRemmitance:", transactionBatch.TotalSupposedRemmitance)

	tbCashCount, err := m.TBCashCount(ctx, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		fmt.Println("Error: failed to calculate total cash count", err)
		return endTx(eris.Wrap(err, "failed to calculate total cash count"))
	}
	transactionBatch.TotalCashOnHand = tbCashCount
	transactionBatch.CashCountTotal = tbCashCount
	fmt.Println("TotalCashOnHand / CashCountTotal:", tbCashCount)

	tbCheckRemittance, err := m.TBCheckRemittance(ctx, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		fmt.Println("Error: failed to calculate total check remittance", err)
		return endTx(eris.Wrap(err, "failed to calculate total check remittance"))
	}
	transactionBatch.TotalCheckRemittance = tbCheckRemittance
	fmt.Println("TotalCheckRemittance:", tbCheckRemittance)

	tbOnlineRemittance, err := m.TBOnlineRemittance(ctx, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		fmt.Println("Error: failed to calculate total online remittance", err)
		return endTx(eris.Wrap(err, "failed to calculate total online remittance"))
	}
	transactionBatch.TotalOnlineRemittance = tbOnlineRemittance
	fmt.Println("TotalOnlineRemittance:", tbOnlineRemittance)

	transactionBatch.TotalDepositInBank = transactionBatch.DepositInBank
	fmt.Println("TotalDepositInBank:", transactionBatch.TotalDepositInBank)

	// GrandTotal = tbCashCount + DepositInBank + batchFunding
	grandTotal := decimal.NewFromFloat(tbCashCount).
		Add(decimal.NewFromFloat(transactionBatch.DepositInBank)).
		Add(decimal.NewFromFloat(batchFunding))
	transactionBatch.GrandTotal = grandTotal.InexactFloat64()
	fmt.Println("GrandTotal calculated:", transactionBatch.GrandTotal)

	// TotalActualRemittance = check + online + cash + deposit
	totalActualRemittance := decimal.NewFromFloat(transactionBatch.TotalCheckRemittance).
		Add(decimal.NewFromFloat(transactionBatch.TotalOnlineRemittance)).
		Add(decimal.NewFromFloat(transactionBatch.TotalCashOnHand)).
		Add(decimal.NewFromFloat(transactionBatch.TotalDepositInBank))
	transactionBatch.TotalActualRemittance = totalActualRemittance.InexactFloat64()
	fmt.Println("TotalActualRemittance calculated:", transactionBatch.TotalActualRemittance)

	// TotalActualSupposedComparison = totalActualRemittance - totalSupposed
	totalActualSupposedComparison := totalActualRemittance.Sub(totalSupposed)
	transactionBatch.TotalActualSupposedComparison = totalActualSupposedComparison.InexactFloat64()
	fmt.Println("TotalActualSupposedComparison:", transactionBatch.TotalActualSupposedComparison)

	if err := m.core.TransactionBatchManager().UpdateByID(ctx, transactionBatch.ID, transactionBatch); err != nil {
		fmt.Println("Error: failed to update transaction batch in DB", err)
		return endTx(eris.Wrap(err, "failed to update transaction batch"))
	}
	fmt.Println("Transaction batch updated successfully in DB")

	if err := endTx(nil); err != nil {
		fmt.Println("Error: failed to end (commit) transaction", err)
		return eris.Wrap(err, "failed to end transaction")
	}

	fmt.Println("Function completed successfully")
	return nil
}
func (m *Event) TBBatchFunding(
	context context.Context,
	transactionBatchID, orgID, branchID uuid.UUID,
) (float64, error) {
	batchFunding, err := m.core.BatchFundingManager().Find(context, &core.BatchFunding{
		TransactionBatchID: transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
	})
	if err != nil {
		return 0, eris.Wrap(err, "failed to find batch fundings")
	}

	totalBatchDec := decimal.Zero
	for _, funding := range batchFunding {
		amountDec := decimal.NewFromFloat(funding.Amount)
		totalBatchDec = totalBatchDec.Add(amountDec)
	}

	return totalBatchDec.InexactFloat64(), nil
}

func (m *Event) TBCashCount(
	context context.Context,
	transactionBatchID, orgID, branchID uuid.UUID,
) (float64, error) {
	cashCounts, err := m.core.CashCountManager().Find(context, &core.CashCount{
		TransactionBatchID: transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
	})
	if err != nil {
		return 0, eris.Wrap(err, "failed to find cash counts")
	}

	totalCashDec := decimal.Zero
	for _, cashCount := range cashCounts {
		amountDec := decimal.NewFromFloat(cashCount.Amount)
		quantityDec := decimal.NewFromFloat(float64(cashCount.Quantity))
		totalCashDec = totalCashDec.Add(amountDec.Mul(quantityDec))
	}

	return totalCashDec.InexactFloat64(), nil
}

func (m *Event) TBCheckRemittance(
	context context.Context,
	transactionBatchID, orgID, branchID uuid.UUID,
) (float64, error) {
	checkRemittances, err := m.core.CheckRemittanceManager().Find(context, &core.CheckRemittance{
		TransactionBatchID: &transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
	})
	if err != nil {
		return 0, eris.Wrap(err, "failed to find check remittances")
	}

	totalCheckDec := decimal.Zero
	for _, remittance := range checkRemittances {
		amountDec := decimal.NewFromFloat(remittance.Amount)
		totalCheckDec = totalCheckDec.Add(amountDec)
	}

	return totalCheckDec.InexactFloat64(), nil
}

func (m *Event) TBOnlineRemittance(
	context context.Context,
	transactionBatchID, orgID, branchID uuid.UUID,
) (float64, error) {
	onlineRemittances, err := m.core.OnlineRemittanceManager().Find(context, &core.OnlineRemittance{
		TransactionBatchID: &transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
	})
	if err != nil {
		return 0, eris.Wrap(err, "failed to find online remittances")
	}

	totalOnlineDec := decimal.Zero
	for _, remittance := range onlineRemittances {
		amountDec := decimal.NewFromFloat(remittance.Amount)
		totalOnlineDec = totalOnlineDec.Add(amountDec)
	}

	return totalOnlineDec.InexactFloat64(), nil
}
func (m *Event) TBDisbursementTransaction(
	context context.Context,
	transactionBatchID, orgID, branchID uuid.UUID,
) (float64, error) {
	disbursementTransactions, err := m.core.DisbursementTransactionManager().Find(context, &core.DisbursementTransaction{
		TransactionBatchID: transactionBatchID,
		OrganizationID:     orgID,
		BranchID:           branchID,
	})
	if err != nil {
		return 0, eris.Wrap(err, "failed to find disbursement transactions")
	}

	totalDisbursementDec := decimal.Zero
	for _, transaction := range disbursementTransactions {
		amountDec := decimal.NewFromFloat(transaction.Amount)
		totalDisbursementDec = totalDisbursementDec.Add(amountDec)
	}

	return totalDisbursementDec.InexactFloat64(), nil
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
	return usecase.CalculateBalance(usecase.Balance{
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
	return usecase.CalculateBalance(usecase.Balance{
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
	return usecase.CalculateBalance(usecase.Balance{
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
	return usecase.CalculateBalance(usecase.Balance{
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
	return usecase.CalculateBalance(usecase.Balance{
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
	return usecase.CalculateBalance(usecase.Balance{
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
	return usecase.CalculateBalance(usecase.Balance{
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
	return usecase.CalculateBalance(usecase.Balance{
		GeneralLedgers: loanVouchers,
	})
}
