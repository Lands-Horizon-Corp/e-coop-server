// Package service provides business logic and transaction processing services for the e-cooperative application
package service

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelcore"
	"github.com/rotisserie/eris"
)

// Deposit processes a deposit transaction for the given account and amount
func (t *TransactionService) Deposit(ctx context.Context, account TransactionData, amount float64) (credit, debit, balance float64, err error) {
	if account.Account == nil {
		return 0, 0, 0, eris.New("account is required")
	}

	if amount == 0 {
		return 0, 0, balance, eris.New("amount must be greater than zero")
	}
	if amount < 0 {
		return t.Withdraw(ctx, account, -amount)
	}
	balance = 0
	if account.GeneralLedger != nil {
		balance = account.GeneralLedger.Balance
	}
	switch account.Account.Type {
	case modelcore.AccountTypeDeposit, modelcore.AccountTypeTimeDeposit, modelcore.AccountTypeSVFLedger:
		// Money in = credit to balance
		return amount, 0, balance + amount, nil

	case modelcore.AccountTypeLoan, modelcore.AccountTypeFines, modelcore.AccountTypeInterest, modelcore.AccountTypeAPLedger:
		// Paying off liabilities = debit (reduces liability balance)
		return 0, amount, balance - amount, nil

	case modelcore.AccountTypeARLedger, modelcore.AccountTypeARAging:
		// Receiving 32nt for receivables = credit balance
		return amount, 0, balance + amount, nil

	case modelcore.AccountTypeWOff, modelcore.AccountTypeOther:
		// Custom handling
		return amount, 0, balance + amount, nil

	default:
		return 0, 0, balance, nil
	}
}
