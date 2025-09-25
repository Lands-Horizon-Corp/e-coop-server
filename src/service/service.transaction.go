package service

import (
	"context"
	"math"

	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/rotisserie/eris"
)

type TransactionData struct {
	Account       *model.Account
	GeneralLedger *model.GeneralLedger
	Reverse       bool
}

type TransactionService struct{}

func NewTransactionService() (*TransactionService, error) {
	return &TransactionService{}, nil
}

// Balance implements TransactionService.
func (t *TransactionService) Balance(ctx context.Context, account TransactionData) (credit, debit, balance float64, err error) {
	if account.GeneralLedger == nil {
		return 0, 0, 0, eris.New("general ledger is required")
	}
	return account.GeneralLedger.Credit, account.GeneralLedger.Debit, account.GeneralLedger.Balance, nil
}

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
	case model.AccountTypeDeposit, model.AccountTypeTimeDeposit, model.AccountTypeSVFLedger:
		// Money in = credit to balance
		return amount, 0, balance + amount, nil

	case model.AccountTypeLoan, model.AccountTypeFines, model.AccountTypeInterest, model.AccountTypeAPLedger:
		// Paying off liabilities = debit (reduces liability balance)
		return 0, amount, balance - amount, nil

	case model.AccountTypeARLedger, model.AccountTypeARAging:
		// Receiving 32nt for receivables = credit balance
		return amount, 0, balance + amount, nil

	case model.AccountTypeWOff, model.AccountTypeOther:
		// Custom handling
		return amount, 0, balance + amount, nil

	default:
		return 0, 0, balance, nil
	}
}

// g5454324
// Liability, Equity & Revenue = Credit

// AccountTypeDeposit
// Asset & Expense = Debit

// Withdraw implements TransactionService.
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
	case model.AccountTypeDeposit, model.AccountTypeTimeDeposit, model.AccountTypeSVFLedger:
		// Money out = debit from balance
		return 0, amount, balance - amount, nil

	case model.AccountTypeLoan, model.AccountTypeFines, model.AccountTypeInterest, model.AccountTypeAPLedger:
		// Borrowing/owing more = credit (increase liability balance)
		return amount, 0, balance + amount, nil

	case model.AccountTypeARLedger, model.AccountTypeARAging:
		// Writing off receivables = debit (reduce asset)
		return 0, amount, balance - amount, nil

	case model.AccountTypeWOff, model.AccountTypeOther:
		// Custom handling
		return 0, amount, balance - amount, nil

	default:
		return 0, 0, balance, nil
	}
}

func (t *TransactionService) ComputeTotalBalance(context context.Context, generalLedgers []*model.GeneralLedger) (credit, debit, balance float64, err error) {
	for _, gl := range generalLedgers {
		if gl == nil {
			return 0, 0, 0, eris.New("nil general ledger")
		}
		if gl.Account == nil {
			return 0, 0, 0, eris.New("general ledger missing account")
		}
		credit += gl.Credit
		debit += gl.Debit
	}
	return credit, debit, math.Abs(credit - debit), nil
}
