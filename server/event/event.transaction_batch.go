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
	transactionBatch, err := m.core.TransactionBatchManager.GetByID(context, *transactionBatchID)
	if err != nil {
		return eris.Wrap(err, "failed to get transaction batch by ID")
	}
	if err := m.core.TransactionBatchManager.UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
		return eris.Wrap(err, "failed to update transaction batch")
	}
	return nil
}

func (m *Event) TBBatchFunding(context context.Context, transactionBatchID, orgID, branchID uuid.UUID) (float64, error) {
	batchFunding, err := m.core.BatchFundingManager.Find(context, &core.BatchFunding{
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
	cashCounts, err := m.core.CashCountManager.Find(context, &core.CashCount{
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
	checkRemittances, err := m.core.CheckRemittanceManager.Find(context, &core.CheckRemittance{
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
	onlineRemittances, err := m.core.OnlineRemittanceManager.Find(context, &core.OnlineRemittance{
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
	disbursementTransactions, err := m.core.DisbursementTransactionManager.Find(context, &core.DisbursementTransaction{
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
	withdrawals, err := m.core.GeneralLedgerManager.Find(context, &core.GeneralLedger{
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
	deposits, err := m.core.GeneralLedgerManager.Find(context, &core.GeneralLedger{
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
	journals, err := m.core.GeneralLedgerManager.Find(context, &core.GeneralLedger{
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
	payments, err := m.core.GeneralLedgerManager.Find(context, &core.GeneralLedger{
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
	adjustments, err := m.core.GeneralLedgerManager.Find(context, &core.GeneralLedger{
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
	journalVouchers, err := m.core.GeneralLedgerManager.Find(context, &core.GeneralLedger{
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
	checkVouchers, err := m.core.GeneralLedgerManager.Find(context, &core.GeneralLedger{
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
	loanVouchers, err := m.core.GeneralLedgerManager.Find(context, &core.GeneralLedger{
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

/*
--------------------------- Batch Funding
cashCounts, err := c.core.CashCountManager.Find(context, &core.CashCount{
	TransactionBatchID: transactionBatch.ID,
	OrganizationID:     userOrg.OrganizationID,
	BranchID:           *userOrg.BranchID,
})
if err != nil {
	c.event.Footstep(ctx, event.FootstepEvent{
		Activity:    "create-error",
		Description: "Batch funding creation failed (/batch-funding), cash count lookup error: " + err.Error(),
		Module:      "BatchFunding",
	})
	return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to retrieve cash counts: " + err.Error()})
}

var totalCashCount float64
for _, cashCount := range cashCounts {
	totalCashCount = c.provider.Service.Decimal.Add(totalCashCount, c.provider.Service.Decimal.Multiply(cashCount.Amount, float64(cashCount.Quantity)))
}
transactionBatch.BeginningBalance = c.provider.Service.Decimal.Add(transactionBatch.BeginningBalance, batchFundingReq.Amount)
transactionBatch.TotalCashHandled = c.provider.Service.Decimal.Add(c.provider.Service.Decimal.Add(batchFundingReq.Amount, transactionBatch.DepositInBank), totalCashCount)
transactionBatch.CashCountTotal = totalCashCount
transactionBatch.GrandTotal = c.provider.Service.Decimal.Add(totalCashCount, transactionBatch.DepositInBank)

if err := c.core.TransactionBatchManager.UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
	c.event.Footstep(ctx, event.FootstepEvent{
		Activity:    "create-error",
		Description: "Batch funding creation failed (/batch-funding), transaction batch update error: " + err.Error(),
		Module:      "BatchFunding",
	})
	return ctx.JSON(http.StatusConflict, map[string]string{"error": "Could not update transaction batch balances: " + err.Error()})
}

--------------------------- Cash Count


--------------------------- Check Remittance

allCheckRemittances, err := c.core.CheckRemittanceManager.Find(context, &core.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
})
if err != nil {
	c.event.Footstep(ctx, event.FootstepEvent{
		Activity:    "delete-error",
		Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), recalc error: " + err.Error(),
		Module:      "CheckRemittance",
	})
	return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to recalculate check remittances: " + err.Error()})
}

var totalCheckRemittance float64
for _, remittance := range allCheckRemittances {
	totalCheckRemittance = c.provider.Service.Decimal.Add(totalCheckRemittance, remittance.Amount)
}
transactionBatch.TotalCheckRemittance = totalCheckRemittance
transactionBatch.TotalActualRemittance = c.provider.Service.Decimal.Add(c.provider.Service.Decimal.Add(transactionBatch.TotalCheckRemittance, transactionBatch.TotalOnlineRemittance), transactionBatch.TotalDepositInBank)
transactionBatch.UpdatedAt = time.Now().UTC()
transactionBatch.UpdatedByID = userOrg.UserID


--------------------------- Online Remittance


allOnlineRemittances, err := c.core.OnlineRemittanceManager.Find(context, &core.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: find all error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve online remittances: " + err.Error()})
		}

		var totalOnlineRemittance float64
		for _, remittance := range allOnlineRemittances {
			totalOnlineRemittance = c.provider.Service.Decimal.Add(totalOnlineRemittance, remittance.Amount)
		}

		transactionBatch.TotalOnlineRemittance = totalOnlineRemittance
		transactionBatch.TotalActualRemittance = c.provider.Service.Decimal.Add(c.provider.Service.Decimal.Add(transactionBatch.TotalCheckRemittance, transactionBatch.TotalOnlineRemittance), transactionBatch.TotalDepositInBank)
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		if err := c.core.TransactionBatchManager.UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: update batch error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}


--------------------------- Disbursement Transaction


--------------------- Deposit in Bank
cashCounts, err := c.core.CashCountManager.Find(context, &core.CashCount{
	TransactionBatchID: transactionBatch.ID,
	OrganizationID:     userOrg.OrganizationID,
	BranchID:           *userOrg.BranchID,
})
if err != nil {
	c.event.Footstep(ctx, event.FootstepEvent{
		Activity:    "update-error",
		Description: "Update deposit in bank failed: get cash counts error: " + err.Error(),
		Module:      "TransactionBatch",
	})
	return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash counts: " + err.Error()})
}

var totalCashCount float64
for _, cashCount := range cashCounts {
	totalCashCount = c.provider.Service.Decimal.Add(totalCashCount, cashCount.Amount)
}

transactionBatch.DepositInBank = req.DepositInBank
transactionBatch.GrandTotal = c.provider.Service.Decimal.Add(totalCashCount, req.DepositInBank)
transactionBatch.TotalCashHandled = c.provider.Service.Decimal.Add(c.provider.Service.Decimal.Add(transactionBatch.BeginningBalance, req.DepositInBank), totalCashCount)
transactionBatch.TotalDepositInBank = req.DepositInBank
transactionBatch.UpdatedAt = time.Now().UTC()
transactionBatch.UpdatedByID = userOrg.UserID

*/
