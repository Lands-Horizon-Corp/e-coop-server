package usecase

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (t *TransactionService) ComputeBalance(generalLedgers []*core.GeneralLedger) (credit, debit, balance float64, err error) {
	credit = 0.0
	debit = 0.0
	balance = 0.0

	for _, gl := range generalLedgers {
		if gl == nil {
			return 0, 0, 0, eris.New("nil general ledger")
		}
		if gl.Account == nil {
			return 0, 0, 0, eris.New("general ledger missing account")
		}
		credit = t.provider.Service.Decimal.Add(credit, gl.Credit)
		debit = t.provider.Service.Decimal.Add(debit, gl.Debit)

		switch gl.Account.GeneralLedgerType {
		case core.GLTypeAssets, core.GLTypeExpenses:
			balance = t.provider.Service.Decimal.Add(balance, gl.Debit-gl.Credit)
		case core.GLTypeLiabilities, core.GLTypeEquity, core.GLTypeRevenue:
			balance = t.provider.Service.Decimal.Add(balance, gl.Credit-gl.Debit)

		default:
			balance = t.provider.Service.Decimal.Add(balance, gl.Debit-gl.Credit)
		}
	}

	return credit, debit, balance, nil
}

type ComputeAdjustment struct {
	Entries    []*core.AdjustmentEntry
	CurrencyID *uuid.UUID
}

func (t *TransactionService) ComputeAdjustment(adjustment ComputeAdjustment) (credit, debit, balance float64, err error) {
	credit = 0.0
	debit = 0.0
	balance = 0.0

	for _, entry := range adjustment.Entries {
		if entry == nil {
			return 0, 0, 0, eris.New("nil general ledger")
		}
		if entry.Account == nil {
			return 0, 0, 0, eris.New("adjustment entry missing account")
		}

		if adjustment.CurrencyID == nil || handlers.UUIDPtrEqual(entry.Account.CurrencyID, adjustment.CurrencyID) {
			credit = t.provider.Service.Decimal.Add(credit, entry.Credit)
			debit = t.provider.Service.Decimal.Add(debit, entry.Debit)
		}

		switch entry.Account.GeneralLedgerType {
		case core.GLTypeAssets, core.GLTypeExpenses:
			balance = t.provider.Service.Decimal.Add(balance, entry.Debit-entry.Credit)
		case core.GLTypeLiabilities, core.GLTypeEquity, core.GLTypeRevenue:
			balance = t.provider.Service.Decimal.Add(balance, entry.Credit-entry.Debit)

		default:
			balance = t.provider.Service.Decimal.Add(balance, entry.Debit-entry.Credit)
		}
	}

	return credit, debit, balance, nil
}
