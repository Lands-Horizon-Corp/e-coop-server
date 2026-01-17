package usecase

import (
	"sort"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

type Balance struct {
	GeneralLedgers         []*types.GeneralLedger
	AdjustmentEntries      []*types.AdjustmentEntry
	LoanTransactionEntries []*types.LoanTransactionEntry

	CashCheckVoucherEntriesRequest []*types.CashCheckVoucherEntryRequest
	JournalVoucherEntriesRequest   []*types.JournalVoucherEntryRequest

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

func CalculateBalance(data Balance) (BalanceResponse, error) {
	credit := decimal.Zero
	debit := decimal.Zero
	balance := decimal.Zero
	addOnAmount := decimal.Zero

	countDebit := 0
	countCredit := 0

	var lastPayment *time.Time
	var lastCredit *time.Time
	var lastDebit *time.Time

	apply := func(
		accType core.GeneralLedgerType,
		dr decimal.Decimal,
		cr decimal.Decimal,
	) {
		switch accType {
		case core.GLTypeAssets, core.GLTypeExpenses:
			balance = balance.Add(dr.Sub(cr))
		case core.GLTypeLiabilities, core.GLTypeEquity, core.GLTypeRevenue:
			balance = balance.Add(cr.Sub(dr))
		default:
			balance = balance.Add(dr.Sub(cr))
		}
	}

	/* ---------------- GENERAL LEDGERS ---------------- */

	for _, entry := range data.GeneralLedgers {
		if entry == nil {
			return BalanceResponse{}, eris.New("nil general ledger")
		}
		if entry.Account == nil {
			return BalanceResponse{}, eris.New("general ledger missing account")
		}
		if data.AccountID != nil && !helpers.UUIDPtrEqual(entry.AccountID, data.AccountID) {
			continue
		}
		if data.CurrencyID != nil && !helpers.UUIDPtrEqual(entry.Account.CurrencyID, data.CurrencyID) {
			continue
		}

		dr := decimal.NewFromFloat(entry.Debit)
		cr := decimal.NewFromFloat(entry.Credit)

		debit = debit.Add(dr)
		credit = credit.Add(cr)

		if cr.GreaterThan(decimal.Zero) {
			countCredit++
			if lastCredit == nil || entry.EntryDate.After(*lastCredit) {
				lastCredit = &entry.EntryDate
			}
			if lastPayment == nil || entry.EntryDate.After(*lastPayment) {
				lastPayment = &entry.EntryDate
			}
		}
		if dr.GreaterThan(decimal.Zero) {
			countDebit++
			if lastDebit == nil || entry.EntryDate.After(*lastDebit) {
				lastDebit = &entry.EntryDate
			}
		}

		apply(entry.Account.GeneralLedgerType, dr, cr)
	}

	/* ---------------- ADJUSTMENTS ---------------- */

	for _, entry := range data.AdjustmentEntries {
		if entry == nil {
			return BalanceResponse{}, eris.New("nil adjustment entry")
		}
		if data.AccountID != nil && !helpers.UUIDPtrEqual(&entry.AccountID, data.AccountID) {
			continue
		}
		if data.CurrencyID != nil && !helpers.UUIDPtrEqual(entry.Account.CurrencyID, data.CurrencyID) {
			continue
		}

		dr := decimal.NewFromFloat(entry.Debit)
		cr := decimal.NewFromFloat(entry.Credit)

		debit = debit.Add(dr)
		credit = credit.Add(cr)

		if cr.GreaterThan(decimal.Zero) && entry.EntryDate != nil {
			countCredit++
			if lastCredit == nil || entry.EntryDate.After(*lastCredit) {
				lastCredit = entry.EntryDate
			}
		}
		if dr.GreaterThan(decimal.Zero) && entry.EntryDate != nil {
			countDebit++
			if lastDebit == nil || entry.EntryDate.After(*lastDebit) {
				lastDebit = entry.EntryDate
			}
		}

		apply(entry.Account.GeneralLedgerType, dr, cr)
	}

	/* ---------------- LOAN TRANSACTIONS ---------------- */

	for _, entry := range data.LoanTransactionEntries {
		if entry == nil {
			return BalanceResponse{}, eris.New("nil loan transaction entry")
		}

		dr := decimal.NewFromFloat(entry.Debit)
		cr := decimal.NewFromFloat(entry.Credit)

		if entry.IsAddOn && data.IsAddOn {
			addOnAmount = addOnAmount.Add(dr).Add(cr)
		}

		if data.AccountID != nil && !helpers.UUIDPtrEqual(entry.AccountID, data.AccountID) {
			continue
		}
		if data.CurrencyID != nil && !helpers.UUIDPtrEqual(entry.Account.CurrencyID, data.CurrencyID) {
			continue
		}

		debit = debit.Add(dr)
		credit = credit.Add(cr)

		if cr.GreaterThan(decimal.Zero) {
			countCredit++
		}
		if dr.GreaterThan(decimal.Zero) {
			countDebit++
		}

		apply(entry.Account.GeneralLedgerType, dr, cr)
	}

	/* ---------------- VOUCHERS ---------------- */

	for _, entry := range data.CashCheckVoucherEntriesRequest {
		if entry == nil {
			return BalanceResponse{}, eris.New("nil cash check voucher")
		}

		dr := decimal.NewFromFloat(entry.Debit)
		cr := decimal.NewFromFloat(entry.Credit)

		debit = debit.Add(dr)
		credit = credit.Add(cr)

		if cr.GreaterThan(decimal.Zero) {
			countCredit++
		}
		if dr.GreaterThan(decimal.Zero) {
			countDebit++
		}

		balance = balance.Add(dr.Sub(cr))
	}

	for _, entry := range data.JournalVoucherEntriesRequest {
		if entry == nil {
			return BalanceResponse{}, eris.New("nil journal voucher")
		}

		dr := decimal.NewFromFloat(entry.Debit)
		cr := decimal.NewFromFloat(entry.Credit)

		debit = debit.Add(dr)
		credit = credit.Add(cr)

		if cr.GreaterThan(decimal.Zero) {
			countCredit++
		}
		if dr.GreaterThan(decimal.Zero) {
			countDebit++
		}

		balance = balance.Add(dr.Sub(cr))
	}

	return BalanceResponse{
		Credit:      credit.InexactFloat64(),
		Debit:       debit.InexactFloat64(),
		Balance:     balance.InexactFloat64(),
		CountDebit:  countDebit,
		CountCredit: countCredit,
		LastPayment: lastPayment,
		LastCredit:  lastCredit,
		LastDebit:   lastDebit,
		AddOnAmount: addOnAmount.InexactFloat64(),
		IsBalanced:  credit.Equal(debit),
	}, nil
}

func CalculateStrictBalance(data Balance) (BalanceResponse, error) {
	response, err := CalculateBalance(data)
	if err != nil {
		return BalanceResponse{}, eris.Wrap(err, "failed to calculate balance")
	}

	balance := decimal.NewFromFloat(response.Balance)
	debit := decimal.NewFromFloat(response.Debit)

	if !balance.Equal(decimal.Zero) {
		return BalanceResponse{}, eris.Errorf(
			"entries are not balanced: balance is %.2f",
			response.Balance,
		)
	}

	if debit.LessThanOrEqual(decimal.Zero) {
		return BalanceResponse{}, eris.New("entries cannot be empty")
	}

	return response, nil
}

func GeneralLedgerAddBalanceByAccount(
	ledgers []*types.GeneralLedger,
) []*types.GeneralLedger {
	if len(ledgers) == 0 {
		return ledgers
	}

	// Ensure chronological order
	sort.Slice(ledgers, func(i, j int) bool {
		return ledgers[i].EntryDate.Before(ledgers[j].EntryDate)
	})

	accountBalances := make(map[uuid.UUID]decimal.Decimal)

	for _, ledger := range ledgers {
		if ledger == nil || ledger.Account == nil || ledger.AccountID == nil {
			continue
		}

		accountID := *ledger.AccountID
		current := accountBalances[accountID]

		debit := decimal.NewFromFloat(ledger.Debit)
		credit := decimal.NewFromFloat(ledger.Credit)

		switch ledger.Account.GeneralLedgerType {
		case core.GLTypeAssets, core.GLTypeExpenses:
			current = current.Add(debit.Sub(credit))

		case core.GLTypeLiabilities, core.GLTypeEquity, core.GLTypeRevenue:
			current = current.Add(credit.Sub(debit))

		default:
			current = current.Add(debit.Sub(credit))
		}

		accountBalances[accountID] = current

		// Persist running balance back to ledger
		ledger.Balance = current.InexactFloat64()
	}

	return ledgers
}

type GeneralLedgerAccountBalanceSummary struct {
	AccountID uuid.UUID
	Account   *types.Account
	Debit     float64
	Credit    float64
}

func SumGeneralLedgerByAccount(
	ledgers []*types.GeneralLedger,
) []GeneralLedgerAccountBalanceSummary {
	resultMap := make(map[uuid.UUID]*struct {
		Account *types.Account
		Debit   decimal.Decimal
		Credit  decimal.Decimal
	})
	for _, gl := range ledgers {
		if gl == nil || gl.AccountID == nil {
			continue
		}
		summary, exists := resultMap[*gl.AccountID]
		if !exists {
			summary = &struct {
				Account *types.Account
				Debit   decimal.Decimal
				Credit  decimal.Decimal
			}{
				Account: gl.Account,
				Debit:   decimal.Zero,
				Credit:  decimal.Zero,
			}
			resultMap[*gl.AccountID] = summary
		}
		summary.Debit = summary.Debit.Add(decimal.NewFromFloat(gl.Debit))
		summary.Credit = summary.Credit.Add(decimal.NewFromFloat(gl.Credit))
	}
	out := make([]GeneralLedgerAccountBalanceSummary, 0, len(resultMap))
	for id, v := range resultMap {
		if v.Debit.IsZero() && v.Credit.IsZero() {
			continue
		}
		out = append(out, GeneralLedgerAccountBalanceSummary{
			AccountID: id,
			Account:   v.Account,
			Debit:     v.Debit.InexactFloat64(),
			Credit:    v.Credit.InexactFloat64(),
		})
	}
	return out
}
