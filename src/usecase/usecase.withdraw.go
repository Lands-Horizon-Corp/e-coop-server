package usecase

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

func Withdraw(
	account *types.Account,
	amount float64,
) (credit float64, debit float64, err error) {
	if account == nil {
		return 0, 0, eris.New("account is required")
	}
	amt := decimal.NewFromFloat(amount)
	if amt.Equal(decimal.Zero) {
		return 0, 0, eris.New("amount must be greater than zero")
	}
	if amt.LessThan(decimal.Zero) {
		return Deposit(account, amt.Abs().InexactFloat64())
	}

	switch account.Type {

	case types.AccountTypeDeposit,
		types.AccountTypeTimeDeposit,
		types.AccountTypeSVFLedger,
		types.AccountTypeLoan,
		types.AccountTypeFines,
		types.AccountTypeInterest,
		types.AccountTypeAPLedger,
		types.AccountTypeARLedger,
		types.AccountTypeARAging,
		types.AccountTypeWOff,
		types.AccountTypeOther:
		return 0, amt.InexactFloat64(), nil
	default:
		return 0, 0, nil
	}
}
