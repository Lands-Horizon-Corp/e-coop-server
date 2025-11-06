package usecase

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"

	"github.com/rotisserie/eris"
)

// TransactionData holds the necessary data for processing financial transactions
// including the account, general ledger entry, and reverse transaction flag.
type TransactionData struct {
	Account       *core.Account
	GeneralLedger *core.GeneralLedger
	Reverse       bool
}

// TransactionService provides methods for handling financial transactions
// and balance calculations in the cooperative system.
type TransactionService struct {
	model    *core.Core
	provider *server.Provider
}

// NewTransactionService creates a new instance of TransactionService
// with the provided model core for database operations.
func NewTransactionService(
	model *core.Core,
	provider *server.Provider,
) (*TransactionService, error) {
	return &TransactionService{
		model:    model,
		provider: provider,
	}, nil
}

// ComputeTotalBalance calculates the total credit, debit, and balance
// from a slice of general ledger entries using precise decimal arithmetic.
func (t *TransactionService) ComputeTotalBalance(generalLedgers []*core.GeneralLedger) (credit, debit, balance float64, err error) {
	credit = 0.0
	debit = 0.0

	for _, gl := range generalLedgers {
		if gl == nil {
			return 0, 0, 0, eris.New("nil general ledger")
		}
		if gl.Account == nil {
			return 0, 0, 0, eris.New("general ledger missing account")
		}

		// Use precise decimal arithmetic for financial calculations
		credit = t.provider.Service.Decimal.Add(credit, gl.Credit)
		debit = t.provider.Service.Decimal.Add(debit, gl.Debit)
	}

	// Calculate balance using decimal arithmetic and get absolute value
	balanceDiff := t.provider.Service.Decimal.Subtract(credit, debit)
	balance = t.provider.Service.Decimal.Abs(balanceDiff)

	return credit, debit, balance, nil
}
