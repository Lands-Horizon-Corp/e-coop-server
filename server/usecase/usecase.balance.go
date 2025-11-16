package usecase

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type Balance struct {
	// Entries to balance
	GeneralLedgers          []*core.GeneralLedger
	AdjustmentEntry         []*core.AdjustmentEntry
	CashCheckVoucherRequest []*core.CashCheckVoucherEntryRequest
	JournalVoucherEntries   []*core.JournalVoucherEntryRequest

	// Strict variables
	CurrencyID *uuid.UUID
}

func (t *TransactionService) Balance(data Balance) (credit, debit, balance float64, err error) {
	credit = 0.0
	debit = 0.0
	balance = 0.0
	if data.GeneralLedgers != nil {
		for _, entry := range data.GeneralLedgers {
			if entry == nil {
				return 0, 0, 0, eris.New("nil general ledger")
			}
			if entry.Account == nil {
				return 0, 0, 0, eris.New("general ledger missing account")
			}
			if data.CurrencyID == nil || handlers.UUIDPtrEqual(entry.Account.CurrencyID, data.CurrencyID) {
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
	}
	if data.AdjustmentEntry != nil {
		for _, entry := range data.AdjustmentEntry {
			if entry == nil {
				return 0, 0, 0, eris.New("nil cash check voucher")
			}
			if data.CurrencyID == nil || handlers.UUIDPtrEqual(entry.Account.CurrencyID, data.CurrencyID) {
				credit = t.provider.Service.Decimal.Add(credit, entry.Credit)
				debit = t.provider.Service.Decimal.Add(debit, entry.Debit)
			}
			balance = t.provider.Service.Decimal.Add(balance, entry.Debit-entry.Credit)
			switch entry.Account.GeneralLedgerType {
			case core.GLTypeAssets, core.GLTypeExpenses:
				balance = t.provider.Service.Decimal.Add(balance, entry.Debit-entry.Credit)
			case core.GLTypeLiabilities, core.GLTypeEquity, core.GLTypeRevenue:
				balance = t.provider.Service.Decimal.Add(balance, entry.Credit-entry.Debit)

			default:
				balance = t.provider.Service.Decimal.Add(balance, entry.Debit-entry.Credit)
			}
		}
	}
	if data.CashCheckVoucherRequest != nil {
		for _, entry := range data.CashCheckVoucherRequest {
			if entry == nil {
				return 0, 0, 0, eris.New("nil cash check voucher")
			}
			credit = t.provider.Service.Decimal.Add(credit, entry.Credit)
			debit = t.provider.Service.Decimal.Add(debit, entry.Debit)
			balance = t.provider.Service.Decimal.Add(balance, entry.Debit-entry.Credit)
		}
	}
	if data.JournalVoucherEntries != nil {
		for _, entry := range data.JournalVoucherEntries {
			if entry == nil {
				return 0, 0, 0, eris.New("nil journal voucher")
			}
			credit = t.provider.Service.Decimal.Add(credit, entry.Credit)
			debit = t.provider.Service.Decimal.Add(debit, entry.Debit)
			balance = t.provider.Service.Decimal.Add(balance, entry.Debit-entry.Credit)

		}
	}

	return credit, debit, balance, nil
}
func (t *TransactionService) StrictBalance(data Balance) (credit, debit, balance float64, err error) {
	credit, debit, balance, err = t.Balance(data)
	if err != nil {
		return 0, 0, 0, eris.Wrap(err, "failed to calculate balance")
	}
	isBalanced := t.provider.Service.Decimal.IsEqual(balance, 0)
	if !isBalanced {
		return 0, 0, 0, eris.Errorf("entries are not balanced: balance is %.2f", balance)
	}
	if !t.provider.Service.Decimal.IsLessThan(debit, 0) {
		return 0, 0, 0, eris.New("entries cannot be empty")
	}
	return credit, debit, balance, nil
}
