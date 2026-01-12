package usecase

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

func Deposit(
	ctx context.Context,
	account *core.Account,
	amount float64,
) (credit float64, debit float64, err error) {

	if account == nil {
		return 0, 0, eris.New("account is required")
	}

	amt := decimal.NewFromFloat(amount)

	if amt.Equal(decimal.Zero) {
		return 0, 0, eris.New("amount must be greater than zero")
	}

	// Negative deposit → withdraw
	if amt.LessThan(decimal.Zero) {
		return Withdraw(ctx, account, amt.Abs().InexactFloat64())
	}

	switch account.Type {

	case core.AccountTypeDeposit,
		core.AccountTypeTimeDeposit,
		core.AccountTypeSVFLedger,
		core.AccountTypeLoan,
		core.AccountTypeFines,
		core.AccountTypeInterest,
		core.AccountTypeAPLedger,
		core.AccountTypeARLedger,
		core.AccountTypeARAging,
		core.AccountTypeWOff,
		core.AccountTypeOther:

		return amt.InexactFloat64(), 0, nil

	default:
		return 0, 0, nil
	}
}
