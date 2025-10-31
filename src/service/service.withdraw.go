package service

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
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
		return t.Deposit(ctx, account, -amount)
	}
	if account.GeneralLedger != nil {
		balance = account.GeneralLedger.Balance
	}
	if balance < amount {
		return 0, 0, balance, eris.New("insufficient balance")
	}
	switch account.Account.Type {
	case model_core.AccountTypeDeposit, model_core.AccountTypeTimeDeposit, model_core.AccountTypeSVFLedger:
		// Money out = debit from balance
		return 0, amount, balance - amount, nil

	case model_core.AccountTypeLoan, model_core.AccountTypeFines, model_core.AccountTypeInterest, model_core.AccountTypeAPLedger:
		// Borrowing/owing more = credit (increase liability balance)
		return amount, 0, balance + amount, nil

	case model_core.AccountTypeARLedger, model_core.AccountTypeARAging:
		// Writing off receivables = debit (reduce asset)
		return 0, amount, balance - amount, nil

	case model_core.AccountTypeWOff, model_core.AccountTypeOther:
		// Custom handling
		return 0, amount, balance - amount, nil

	default:
		return 0, 0, balance, nil
	}
}
