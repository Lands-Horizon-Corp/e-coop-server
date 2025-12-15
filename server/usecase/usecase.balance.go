package usecase

import (
	"sort"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type Balance struct {
	GeneralLedgers         []*core.GeneralLedger
	AdjustmentEntries      []*core.AdjustmentEntry
	LoanTransactionEntries []*core.LoanTransactionEntry

	CashCheckVoucherEntriesRequest []*core.CashCheckVoucherEntryRequest
	JournalVoucherEntriesRequest   []*core.JournalVoucherEntryRequest

	CurrencyID *uuid.UUID
	AccountID  *uuid.UUID

	IsAddOn bool
}

type BalanceResponse struct {
	Credit  float64
	Debit   float64
	Balance float64

	Deductions float64
	Added      float64

	CountDeductions int
	CountAdded      int
	CountDebit      int
	CountCredit     int

	LastPayment *time.Time
	LastCredit  *time.Time
	LastDebit   *time.Time

	AddOnAmount float64

	IsBalanced bool
}

func (t *UsecaseService) Balance(data Balance) (BalanceResponse, error) {
	credit := 0.0
	debit := 0.0
	balance := 0.0
	added := 0.0
	deductions := 0.0
	countDeductions := 0
	countAdded := 0
	countDebit := 0
	countCredit := 0
	addOnAmount := 0.0
	var lastPayment *time.Time
	var lastCredit *time.Time
	var lastDebit *time.Time
	if data.GeneralLedgers != nil {
		for _, entry := range data.GeneralLedgers {
			if entry == nil {
				return BalanceResponse{
					Credit:  credit,
					Debit:   debit,
					Balance: balance,
				}, eris.New("nil general ledger")
			}
			if entry.Account == nil {
				return BalanceResponse{
					Credit:  credit,
					Debit:   debit,
					Balance: balance,
				}, eris.New("general ledger missing account")
			}
			if data.AccountID != nil && !handlers.UUIDPtrEqual(entry.AccountID, data.AccountID) {
				continue
			}
			if data.CurrencyID != nil && !handlers.UUIDPtrEqual(entry.Account.CurrencyID, data.CurrencyID) {
				continue
			}

			credit = t.provider.Service.Decimal.Add(credit, entry.Credit)
			debit = t.provider.Service.Decimal.Add(debit, entry.Debit)

			if entry.Credit > 0 {
				countCredit++
				if lastPayment == nil || entry.EntryDate.After(*lastPayment) {
					lastPayment = &entry.EntryDate
				}
				if lastCredit == nil || entry.EntryDate.After(*lastCredit) {
					lastCredit = &entry.EntryDate
				}
			}
			if entry.Debit > 0 {
				countDebit++
				if lastDebit == nil || entry.EntryDate.After(*lastDebit) {
					lastDebit = &entry.EntryDate
				}
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

	if data.AdjustmentEntries != nil {
		for _, entry := range data.AdjustmentEntries {
			if entry == nil {
				return BalanceResponse{
					Credit:     credit,
					Debit:      debit,
					Balance:    balance,
					IsBalanced: true,
				}, eris.New("nil adjustment entry")
			}
			if data.AccountID != nil && !handlers.UUIDPtrEqual(&entry.AccountID, data.AccountID) {
				continue
			}
			if data.CurrencyID != nil && !handlers.UUIDPtrEqual(entry.Account.CurrencyID, data.CurrencyID) {
				continue
			}

			credit = t.provider.Service.Decimal.Add(credit, entry.Credit)
			debit = t.provider.Service.Decimal.Add(debit, entry.Debit)

			if entry.Credit > 0 {
				countCredit++
				if entry.EntryDate != nil {
					if lastCredit == nil || entry.EntryDate.After(*lastCredit) {
						lastCredit = entry.EntryDate
					}
				}
			}
			if entry.Debit > 0 {
				countDebit++
				if entry.EntryDate != nil {
					if lastDebit == nil || entry.EntryDate.After(*lastDebit) {
						lastDebit = entry.EntryDate
					}
				}
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

	if data.LoanTransactionEntries != nil {
		for _, entry := range data.LoanTransactionEntries {
			if entry == nil {
				return BalanceResponse{
					Credit:  credit,
					Debit:   debit,
					Balance: balance,
				}, eris.New("nil loan transaction entry")
			}
			if entry.IsAddOn && data.IsAddOn {
				addOnAmount = t.provider.Service.Decimal.Add(addOnAmount, entry.Debit+entry.Credit)
			}
			if data.AccountID != nil && !handlers.UUIDPtrEqual(entry.AccountID, data.AccountID) {
				continue
			}
			if data.CurrencyID != nil && !handlers.UUIDPtrEqual(entry.Account.CurrencyID, data.CurrencyID) {
				continue
			}

			credit = t.provider.Service.Decimal.Add(credit, entry.Credit)
			debit = t.provider.Service.Decimal.Add(debit, entry.Debit)

			if entry.Credit > 0 {
				countCredit++
			}
			if entry.Debit > 0 {
				countDebit++
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

	if data.CashCheckVoucherEntriesRequest != nil {
		for _, entry := range data.CashCheckVoucherEntriesRequest {
			if entry == nil {
				return BalanceResponse{
					Credit:  credit,
					Debit:   debit,
					Balance: balance,
				}, eris.New("nil cash check voucher")
			}
			credit = t.provider.Service.Decimal.Add(credit, entry.Credit)
			debit = t.provider.Service.Decimal.Add(debit, entry.Debit)
			if entry.Credit > 0 {
				countCredit++
			}
			if entry.Debit > 0 {
				countDebit++
			}
			balance = t.provider.Service.Decimal.Add(balance, entry.Debit-entry.Credit)
		}
	}

	if data.JournalVoucherEntriesRequest != nil {
		for _, entry := range data.JournalVoucherEntriesRequest {
			if entry == nil {
				return BalanceResponse{
					Credit:  credit,
					Debit:   debit,
					Balance: balance,
				}, eris.New("nil journal voucher")
			}
			credit = t.provider.Service.Decimal.Add(credit, entry.Credit)
			debit = t.provider.Service.Decimal.Add(debit, entry.Debit)
			if entry.Credit > 0 {
				countCredit++
			}
			if entry.Debit > 0 {
				countDebit++
			}
			balance = t.provider.Service.Decimal.Add(balance, entry.Debit-entry.Credit)
		}
	}

	return BalanceResponse{
		IsBalanced:      t.provider.Service.Decimal.IsEqual(credit, debit),
		Credit:          credit,
		Debit:           debit,
		Balance:         balance,
		Deductions:      deductions,
		Added:           added,
		CountDeductions: countDeductions,
		CountAdded:      countAdded,
		CountDebit:      countDebit,
		CountCredit:     countCredit,
		LastPayment:     lastPayment,
		LastCredit:      lastCredit,
		LastDebit:       lastDebit,
		AddOnAmount:     addOnAmount,
	}, nil
}

func (t *UsecaseService) StrictBalance(data Balance) (BalanceResponse, error) {
	response, err := t.Balance(data)
	if err != nil {
		return BalanceResponse{}, eris.Wrap(err, "failed to calculate balance")
	}
	isBalanced := t.provider.Service.Decimal.IsEqual(response.Balance, 0)
	if !isBalanced {
		return BalanceResponse{}, eris.Errorf("entries are not balanced: balance is %.2f", response.Balance)
	}
	if t.provider.Service.Decimal.IsLessThan(response.Debit, 0) {
		return BalanceResponse{}, eris.New("entries cannot be empty")
	}
	return response, nil
}

func (t *UsecaseService) GeneralLedgerAddBalanceByAccount(GeneralLedgers []*core.GeneralLedger) []*core.GeneralLedger {
	if len(GeneralLedgers) == 0 {
		return GeneralLedgers
	}

	sort.Slice(GeneralLedgers, func(i, j int) bool {
		return GeneralLedgers[i].EntryDate.Before(GeneralLedgers[j].EntryDate)
	})

	accountBalances := make(map[uuid.UUID]float64)

	for _, ledger := range GeneralLedgers {
		if ledger == nil || ledger.Account == nil || ledger.AccountID == nil {
			continue
		}

		accountID := *ledger.AccountID
		currentBalance := accountBalances[accountID]

		switch ledger.Account.GeneralLedgerType {
		case core.GLTypeAssets, core.GLTypeExpenses:
			currentBalance = t.provider.Service.Decimal.Add(currentBalance, ledger.Debit-ledger.Credit)
		case core.GLTypeLiabilities, core.GLTypeEquity, core.GLTypeRevenue:
			currentBalance = t.provider.Service.Decimal.Add(currentBalance, ledger.Credit-ledger.Debit)
		default:
			currentBalance = t.provider.Service.Decimal.Add(currentBalance, ledger.Debit-ledger.Credit)
		}

		accountBalances[accountID] = currentBalance

		ledger.Balance = currentBalance
	}

	return GeneralLedgers
}
