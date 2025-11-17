package usecase

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/rotisserie/eris"
)

// Withdraw processes a withdrawal transaction for the specified account and amount.
func (t *TransactionService) Withdraw(
	ctx context.Context,
	account *core.Account,
	amount float64) (credit, debit float64, err error) {
	if account == nil {
		return 0, 0, eris.New("account is required")
	}

	if amount == 0 {
		return 0, 0, eris.New("amount must be greater than zero")
	}
	if amount < 0 {
		positiveAmount := t.provider.Service.Decimal.Abs(amount)
		return t.Deposit(ctx, account, positiveAmount)
	}

	switch account.Type {
	case core.AccountTypeDeposit, core.AccountTypeTimeDeposit, core.AccountTypeSVFLedger:
		return 0, amount, nil

	case core.AccountTypeLoan, core.AccountTypeFines, core.AccountTypeInterest, core.AccountTypeAPLedger:
		return amount, 0, nil

	case core.AccountTypeARLedger, core.AccountTypeARAging:
		return 0, amount, nil

	case core.AccountTypeWOff, core.AccountTypeOther:
		return 0, amount, nil

	default:
		return 0, 0, nil
	}
}
