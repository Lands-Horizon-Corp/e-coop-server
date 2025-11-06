package usecase

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/rotisserie/eris"
)

// Withdraw processes a withdrawal transaction for the specified account and amount.
// It returns the credit, debit, and resulting balance after the withdrawal operation.
func (t *TransactionService) Withdraw(ctx context.Context, account TransactionData, amount float64) (credit, debit, balance float64, err error) {
	if account.Account == nil {
		return 0, 0, 0, eris.New("account is required")
	}

	balance = 0
	if amount == 0 {
		return 0, 0, balance, eris.New("amount must be greater than zero")
	}
	if amount < 0 {
		// Use decimal arithmetic for negative amount conversion
		positiveAmount := t.provider.Service.Decimal.Abs(amount)
		return t.Deposit(ctx, account, positiveAmount)
	}
	if account.GeneralLedger != nil {
		balance = account.GeneralLedger.Balance
	}

	// Use decimal arithmetic for balance comparison
	if t.provider.Service.Decimal.IsLessThan(balance, amount) {
		return 0, 0, balance, eris.New("insufficient balance")
	}

	switch account.Account.Type {
	case core.AccountTypeDeposit, core.AccountTypeTimeDeposit, core.AccountTypeSVFLedger:
		// Money out = debit from balance using precise decimal arithmetic
		newBalance := t.provider.Service.Decimal.Subtract(balance, amount)
		return 0, amount, newBalance, nil

	case core.AccountTypeLoan, core.AccountTypeFines, core.AccountTypeInterest, core.AccountTypeAPLedger:
		// Borrowing/owing more = credit (increase liability balance) using precise decimal arithmetic
		newBalance := t.provider.Service.Decimal.Add(balance, amount)
		return amount, 0, newBalance, nil

	case core.AccountTypeARLedger, core.AccountTypeARAging:
		// Writing off receivables = debit (reduce asset) using precise decimal arithmetic
		newBalance := t.provider.Service.Decimal.Subtract(balance, amount)
		return 0, amount, newBalance, nil

	case core.AccountTypeWOff, core.AccountTypeOther:
		// Custom handling using precise decimal arithmetic
		newBalance := t.provider.Service.Decimal.Subtract(balance, amount)
		return 0, amount, newBalance, nil

	default:
		return 0, 0, balance, nil
	}
}
