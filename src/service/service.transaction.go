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
