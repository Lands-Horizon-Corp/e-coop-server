// Package usecase provides business logic and transaction processing services for the e-cooperative application
package usecase

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/rotisserie/eris"
)

// Deposit processes a deposit transaction for the given account and amount using precise decimal arithmetic
func (t *TransactionService) Deposit(
	ctx context.Context,
	account TransactionData,
	amount float64,
) (credit, debit float64, err error) {

	if account.Account == nil {
		return 0, 0, eris.New("account is required")
	}

	if amount == 0 {
		return 0, 0, eris.New("amount must be greater than zero")
	}
	if amount < 0 {
		positiveAmount := t.provider.Service.Decimal.Abs(amount)
		return t.Withdraw(ctx, account, positiveAmount)
	}
	switch account.Account.Type {
	case core.AccountTypeDeposit, core.AccountTypeTimeDeposit, core.AccountTypeSVFLedger:
		return amount, 0, nil

	case core.AccountTypeLoan, core.AccountTypeFines, core.AccountTypeInterest, core.AccountTypeAPLedger:
		return 0, amount, nil

	case core.AccountTypeARLedger, core.AccountTypeARAging:
		return amount, 0, nil

	case core.AccountTypeWOff, core.AccountTypeOther:
		return amount, 0, nil
	default:
		return 0, 0, nil
	}
}
