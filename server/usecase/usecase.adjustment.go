package usecase

import "github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"

// Output should be debit, credit, balance
func (t *TransactionService) Adjustment(
	account core.Account,
	debit, credit, currentBalance float64,
) (float64, float64, float64) {
	var finalBalance float64

	switch account.GeneralLedgerType {
	case core.GLTypeAssets, core.GLTypeExpenses:
		// Normal balance is Debit → add debits, subtract credits
		finalBalance = currentBalance + debit - credit

	case core.GLTypeLiabilities, core.GLTypeEquity, core.GLTypeRevenue:
		// Normal balance is Credit → add credits, subtract debits
		finalBalance = currentBalance + credit - debit

	default:
		finalBalance = currentBalance // fallback, no change
	}

	return debit, credit, finalBalance
}
