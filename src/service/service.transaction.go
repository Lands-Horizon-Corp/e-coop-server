package service

import (
	"context"
	"math"

	modelCore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelCore"

	"github.com/rotisserie/eris"
)

// TransactionData holds the necessary data for processing financial transactions
// including the account, general ledger entry, and reverse transaction flag.
type TransactionData struct {
	Account       *modelCore.Account
	GeneralLedger *modelCore.GeneralLedger
	Reverse       bool
}

// TransactionService provides methods for handling financial transactions
// and balance calculations in the cooperative system.
type TransactionService struct {
	model *modelCore.ModelCore
}

// NewTransactionService creates a new instance of TransactionService
// with the provided model core for database operations.
func NewTransactionService(model *modelCore.ModelCore) (*TransactionService, error) {
	return &TransactionService{
		model: model,
	}, nil
}

// ComputeTotalBalance calculates the total credit, debit, and balance
// from a slice of general ledger entries.
func (t *TransactionService) ComputeTotalBalance(_ context.Context, generalLedgers []*modelCore.GeneralLedger) (credit, debit, balance float64, err error) {
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
