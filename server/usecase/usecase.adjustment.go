package usecase

import "github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"

// Output should be debit, credit, balance using precise decimal arithmetic
func (t *TransactionService) Adjustment(
	account core.Account,
	debit, credit, currentBalance float64,
) (float64, float64, float64) {
	var finalBalance float64

	switch account.GeneralLedgerType {
	case core.GLTypeAssets, core.GLTypeExpenses:
		// Normal balance is Debit → add debits, subtract credits using precise decimal arithmetic
		// finalBalance = currentBalance + debit - credit
		balanceWithDebit := t.provider.Service.Decimal.Add(currentBalance, debit)
		finalBalance = t.provider.Service.Decimal.Subtract(balanceWithDebit, credit)

	case core.GLTypeLiabilities, core.GLTypeEquity, core.GLTypeRevenue:
		// Normal balance is Credit → add credits, subtract debits using precise decimal arithmetic
		// finalBalance = currentBalance + credit - debit
		balanceWithCredit := t.provider.Service.Decimal.Add(currentBalance, credit)
		finalBalance = t.provider.Service.Decimal.Subtract(balanceWithCredit, debit)

	default:
		finalBalance = currentBalance // fallback, no change
	}

	return t.provider.Service.Decimal.Abs(debit), t.provider.Service.Decimal.Abs(credit), finalBalance
}
