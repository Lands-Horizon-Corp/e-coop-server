// Package usecase provides business logic and transaction processing services for the e-cooperative application
package usecase

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/rotisserie/eris"
)

// Deposit processes a deposit transaction for the given account and amount using precise decimal arithmetic
func (t *TransactionService) Deposit(ctx context.Context, account TransactionData, amount float64) (credit, debit, balance float64, err error) {
	if account.Account == nil {
		return 0, 0, 0, eris.New("account is required")
	}

	if amount == 0 {
		return 0, 0, balance, eris.New("amount must be greater than zero")
	}
	if amount < 0 {
		// Use decimal arithmetic for negative amount conversion
		positiveAmount := t.provider.Service.Decimal.Abs(amount)
		return t.Withdraw(ctx, account, positiveAmount)
	}
	balance = 0
	if account.GeneralLedger != nil {
		balance = account.GeneralLedger.Balance
	}

	switch account.Account.Type {
	case core.AccountTypeDeposit, core.AccountTypeTimeDeposit, core.AccountTypeSVFLedger:
		// Money in = credit to balance using precise decimal arithmetic
		newBalance := t.provider.Service.Decimal.Add(balance, amount)
		return amount, 0, newBalance, nil

	case core.AccountTypeLoan, core.AccountTypeFines, core.AccountTypeInterest, core.AccountTypeAPLedger:
		// Paying off liabilities = debit (reduces liability balance) using precise decimal arithmetic
		newBalance := t.provider.Service.Decimal.Subtract(balance, amount)
		return 0, amount, newBalance, nil

	case core.AccountTypeARLedger, core.AccountTypeARAging:
		// Receiving payment for receivables = credit balance using precise decimal arithmetic
		newBalance := t.provider.Service.Decimal.Add(balance, amount)
		return amount, 0, newBalance, nil

	case core.AccountTypeWOff, core.AccountTypeOther:
		// Custom handling using precise decimal arithmetic
		newBalance := t.provider.Service.Decimal.Add(balance, amount)
		return amount, 0, newBalance, nil

	default:
		return 0, 0, balance, nil
	}
}
