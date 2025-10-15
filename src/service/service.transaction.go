package service

import (
	"context"
	"math"

	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"

	"github.com/rotisserie/eris"
)

type TransactionData struct {
	Account       *model_core.Account
	GeneralLedger *model_core.GeneralLedger
	Reverse       bool
}

type TransactionService struct {
	model *model_core.ModelCore
}

func NewTransactionService(model *model_core.ModelCore) (*TransactionService, error) {
	return &TransactionService{
		model: model,
	}, nil
}

func (t *TransactionService) ComputeTotalBalance(context context.Context, generalLedgers []*model_core.GeneralLedger) (credit, debit, balance float64, err error) {
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
