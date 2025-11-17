package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (m *Event) TransactionBatchBalancing(context context.Context, transactionBatchID uuid.UUID) error {
	transactionBatch, err := m.core.TransactionBatchManager.GetByID(context, transactionBatchID)
	if err != nil {
		return eris.Wrap(err, "failed to get transaction batch by ID")
	}

	if err := m.core.TransactionBatchManager.UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
		return eris.Wrap(err, "failed to update transaction batch")
	}
	return nil
}

func (m *Event) BatchFunding() {

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
