package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/usecase"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

func TransactionBatchEnd(
	context context.Context,
	service *horizon.HorizonService,
	userOrg *core.UserOrganization,
	req *core.TransactionBatchEndRequest,
) (*core.TransactionBatch, error) {
	tx, endTx := service.Database.StartTransaction(context)

	transactionBatch, err := core.TransactionBatchCurrent(
		context,
		service,
		userOrg.UserID,
		userOrg.OrganizationID,
		*userOrg.BranchID,
	)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve current transaction batch")
	}
	now := time.Now().UTC()
	transactionBatch.IsClosed = true
	transactionBatch.EmployeeUserID = &userOrg.UserID
	transactionBatch.EmployeeBySignatureMediaID = req.EmployeeBySignatureMediaID
	transactionBatch.EmployeeByName = req.EmployeeByName
	transactionBatch.EmployeeByPosition = req.EmployeeByPosition
	transactionBatch.EndedAt = &now
	if err := core.TransactionBatchManager(service).UpdateByIDWithTx(context, tx, transactionBatch.ID, transactionBatch); err != nil {
		return nil, eris.Wrap(err, "failed to update transaction batch")
	}
	if err := endTx(nil); err != nil {
		return nil, eris.Wrap(err, "failed to commit database transaction")
	}
	return transactionBatch, nil
}

func TransactionBatchBalancing(context context.Context, service *horizon.HorizonService, transactionBatchID *uuid.UUID) error {
	if transactionBatchID == nil {
		return eris.New("transactionBatchID is nil")
	}
	tx, endTx := service.Database.StartTransaction(context)
	transactionBatch, err := core.TransactionBatchManager(service).GetByIDLock(context, tx, *transactionBatchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to get transaction batch by ID"))
	}

	payments, err := TBPayment(context, service, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total payments"))
	}
	transactionBatch.TotalCashCollection = payments.Balance
	totalDeposit, err := TBDeposit(context, service, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total deposits"))
	}
	transactionBatch.TotalDepositEntry = totalDeposit.Balance
	batchFunding, err := TBBatchFunding(context, service, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total batch funding"))
	}
	transactionBatch.BeginningBalance = batchFunding

	// total_cash_handled  = beg_bal + deposit + collection (REAL CASH ONLY)
	totalCashHandled := decimal.NewFromFloat(transactionBatch.BeginningBalance).
		Add(decimal.NewFromFloat(transactionBatch.DepositInBank)).
		Add(decimal.NewFromFloat(transactionBatch.TotalCashCollection))
	transactionBatch.TotalCashHandled = totalCashHandled.InexactFloat64()

	totalWithdraw, err := TBWithdraw(context, service, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total withdrawals"))
	}

	less := decimal.Zero
	savingsWithdrawal := decimal.NewFromFloat(totalWithdraw.Balance).Mul(decimal.NewFromInt(-1))
	transactionBatch.SavingsWithdrawal = savingsWithdrawal.InexactFloat64()
	less = less.Add(savingsWithdrawal)

	disbursementTransaction, err := TBDisbursementTransaction(context, service, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total disbursement transactions"))
	}
	transactionBatch.PettyCash = disbursementTransaction
	less = less.Add(decimal.NewFromFloat(transactionBatch.PettyCash))

	totalSupposed := decimal.NewFromFloat(transactionBatch.TotalCashHandled).Sub(less)
	transactionBatch.TotalSupposedRemmitance = totalSupposed.InexactFloat64()

	tbCashCount, err := TBCashCount(context, service, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total cash count"))
	}
	transactionBatch.TotalCashOnHand = tbCashCount
	transactionBatch.CashCountTotal = tbCashCount

	tbCheckRemittance, err := TBCheckRemittance(context, service, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total check remittance"))
	}
	transactionBatch.TotalCheckRemittance = tbCheckRemittance

	tbOnlineRemittance, err := TBOnlineRemittance(context, service, transactionBatch.ID, transactionBatch.OrganizationID, transactionBatch.BranchID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to calculate total online remittance"))
	}
	transactionBatch.TotalOnlineRemittance = tbOnlineRemittance

	transactionBatch.TotalDepositInBank = transactionBatch.DepositInBank

	// GrandTotal = tbCashCount + DepositInBank + batchFunding
	grandTotal := decimal.NewFromFloat(tbCashCount).
		Add(decimal.NewFromFloat(transactionBatch.DepositInBank)).
		Add(decimal.NewFromFloat(batchFunding))
	transactionBatch.GrandTotal = grandTotal.InexactFloat64()

	// TotalActualRemittance = check + online + cash + deposit
	totalActualRemittance := decimal.NewFromFloat(transactionBatch.TotalCheckRemittance).
		Add(decimal.NewFromFloat(transactionBatch.TotalOnlineRemittance)).
		Add(decimal.NewFromFloat(transactionBatch.TotalCashOnHand)).
		Add(decimal.NewFromFloat(transactionBatch.TotalDepositInBank))
	transactionBatch.TotalActualRemittance = totalActualRemittance.InexactFloat64()

	// TotalActualSupposedComparison = totalActualRemittance - totalSupposed
	totalActualSupposedComparison := totalActualRemittance.Sub(totalSupposed)
	transactionBatch.TotalActualSupposedComparison = totalActualSupposedComparison.InexactFloat64()

	if err := core.TransactionBatchManager(service).UpdateByIDWithTx(context, tx, transactionBatch.ID, transactionBatch); err != nil {
		return endTx(eris.Wrap(err, "failed to update transaction batch"))
	}

	if err := endTx(nil); err != nil {
		return eris.Wrap(err, "failed to end transaction")
	}
	return nil
}
func TBBatchFunding(
	context context.Context, service *horizon.HorizonService,
	transactionBatchID, orgID, branchID uuid.UUID,
) (float64, error) {
	batchFunding, err := core.BatchFundingManager(service).Find(context, &core.BatchFunding{
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

func TBCashCount(
	context context.Context, service *horizon.HorizonService,
	transactionBatchID, orgID, branchID uuid.UUID,
) (float64, error) {
	cashCounts, err := core.CashCountManager(service).Find(context, &core.CashCount{
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

func TBCheckRemittance(
	context context.Context, service *horizon.HorizonService,
	transactionBatchID, orgID, branchID uuid.UUID,
) (float64, error) {
	checkRemittances, err := core.CheckRemittanceManager(service).Find(context, &core.CheckRemittance{
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

func TBOnlineRemittance(
	context context.Context, service *horizon.HorizonService,
	transactionBatchID, orgID, branchID uuid.UUID,
) (float64, error) {
	onlineRemittances, err := core.OnlineRemittanceManager(service).Find(context, &core.OnlineRemittance{
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

func TBDisbursementTransaction(
	context context.Context, service *horizon.HorizonService,
	transactionBatchID, orgID, branchID uuid.UUID,
) (float64, error) {
	disbursementTransactions, err := core.DisbursementTransactionManager(service).Find(context, &core.DisbursementTransaction{
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

func TBWithdraw(context context.Context, service *horizon.HorizonService, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	withdrawals, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
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

func TBDeposit(context context.Context, service *horizon.HorizonService, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	deposits, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
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

func TBJournal(context context.Context, service *horizon.HorizonService, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	journals, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
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

func TBPayment(context context.Context, service *horizon.HorizonService, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	payments, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
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

func TBAdjustment(context context.Context, service *horizon.HorizonService, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	adjustments, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
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

func TBJournalVoucher(context context.Context, service *horizon.HorizonService, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	journalVouchers, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
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

func TBCheckVoucher(context context.Context, service *horizon.HorizonService, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	checkVouchers, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
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

func TBLoan(context context.Context, service *horizon.HorizonService, transactionBatchID, orgID, branchID uuid.UUID) (usecase.BalanceResponse, error) {
	loanVouchers, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
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
