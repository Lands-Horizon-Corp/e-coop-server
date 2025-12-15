package usecase

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/rotisserie/eris"
)

func (t *UsecaseService) Deposit(
	ctx context.Context,
	account *core.Account,
	amount float64,
) (credit, debit float64, err error) {

	if account == nil {
		return 0, 0, eris.New("account is required")
	}

	if amount == 0 {
		return 0, 0, eris.New("amount must be greater than zero")
	}
	if amount < 0 {
		positiveAmount := t.provider.Service.Decimal.Abs(amount)
		return t.Withdraw(ctx, account, positiveAmount)
	}
	switch account.Type {
	case core.AccountTypeDeposit, core.AccountTypeTimeDeposit, core.AccountTypeSVFLedger:
		return amount, 0, nil

	case core.AccountTypeLoan, core.AccountTypeFines, core.AccountTypeInterest, core.AccountTypeAPLedger:
		return amount, 0, nil

	case core.AccountTypeARLedger, core.AccountTypeARAging:
		return amount, 0, nil

	case core.AccountTypeWOff, core.AccountTypeOther:
		return amount, 0, nil
	default:
		return 0, 0, nil
	}
}
