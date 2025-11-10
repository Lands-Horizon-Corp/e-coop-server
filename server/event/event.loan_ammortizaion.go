package event

import (
	"context"
	"sort"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type AccountValue struct {
	Account *core.Account `json:"account" validate:"required"`
	Value   float64       `json:"value" validate:"required,gte=0"`
	Total   float64       `json:"total" validate:"required,gte=0"`
}

type LoanAmortizationSchedule struct {
	ScheduledDate time.Time       `json:"scheduled_date"`
	ActualDate    time.Time       `json:"actual_date"`
	DaysSkipped   int             `json:"days_skipped"`
	Total         float64         `json:"total"`
	Balance       float64         `json:"balance"`
	Accounts      []*AccountValue `json:"accounts"`
}
type LoanTransactionAmortizationResponse struct {
	Entries     []*core.LoanTransactionEntryResponse `json:"entries"`
	TotalDebit  float64                              `json:"total_debit"`
	TotalCredit float64                              `json:"total_credit"`
	Currency    core.CurrencyResponse                `json:"currency"`
	Total       float64                              `json:"total"`
	Schedule    []*LoanAmortizationSchedule          `json:"schedule,omitempty"`
}

func (e Event) LoanAmortizationSchedule(ctx context.Context, loanTransactionID uuid.UUID) (*LoanTransactionAmortizationResponse, error) {
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, loanTransactionID, "Account.Currency")
	if err != nil {
		return nil, err
	}
	loanTransactionEntries, err := e.core.LoanTransactionEntryManager.Find(ctx, &core.LoanTransactionEntry{
		OrganizationID:    loanTransaction.OrganizationID,
		BranchID:          loanTransaction.BranchID,
		LoanTransactionID: loanTransaction.ID,
	})
	if err != nil {
		return nil, err
	}
	totalDebit, totalCredit := 0.0, 0.0
	for _, entry := range loanTransactionEntries {
		totalDebit = e.provider.Service.Decimal.Add(totalDebit, entry.Debit)
		totalCredit = e.provider.Service.Decimal.Add(totalCredit, entry.Credit)
	}
	currency := loanTransaction.Account.Currency
	accounts, err := e.core.AccountManager.Find(ctx, &core.Account{
		OrganizationID: loanTransaction.OrganizationID,
		BranchID:       loanTransaction.BranchID,
		LoanAccountID:  loanTransaction.AccountID,
		CurrencyID:     &currency.ID,
	}, "Currency")
	if err != nil {
		return nil, err
	}
	holidays, err := e.core.HolidayManager.Find(ctx, &core.Holiday{
		OrganizationID: loanTransaction.OrganizationID,
		BranchID:       loanTransaction.BranchID,
		CurrencyID:     currency.ID,
	})
	if err != nil {
		return nil, err
	}

	numberOfPayments, err := e.usecase.LoanNumberOfPayments(loanTransaction.ModeOfPayment, loanTransaction.Terms)
	if err != nil {
		return nil, err
	}

	// Excluding
	excludeSaturday := loanTransaction.ExcludeSaturday
	excludeSunday := loanTransaction.ExcludeSunday
	excludeHolidays := loanTransaction.ExcludeHoliday

	// Payment custom days
	isMonthlyExactDay := loanTransaction.ModeOfPaymentMonthlyExactDay
	weeklyExactDay := loanTransaction.ModeOfPaymentWeekly // expect this to be time.Weekday (0=Sunday...)
	semiMonthlyExactDay1 := loanTransaction.ModeOfPaymentSemiMonthlyPay1
	semiMonthlyExactDay2 := loanTransaction.ModeOfPaymentSemiMonthlyPay2

	accountsSchedule := []*AccountValue{}

	amortization := []*LoanAmortizationSchedule{}
	for _, acc := range accounts {
		accountsSchedule = append(accountsSchedule, &AccountValue{
			Account: acc,
			Value:   0,
			Total:   0,
		})
	}
	accountsSchedule = append(accountsSchedule, &AccountValue{
		Account: loanTransaction.Account,
		Value:   0,
		Total:   0,
	})

	// Typically, start date comes from loanTransaction (adjust as needed)
	paymentDate := time.Now().UTC()
	principal := totalCredit
	balance := totalCredit
	total := 0.0

	for i := range numberOfPayments {
		actualDate := paymentDate
		daysSkipped := 0
		rowTotal := 0.0
		daysSkipped, err := e.skippedDaysCount(paymentDate, currency, excludeSaturday, excludeSunday, excludeHolidays, holidays)
		if err != nil {
			return nil, err
		}

		scheduledDate := paymentDate.AddDate(0, 0, daysSkipped)

		// ✅ CREATE INDEPENDENT ACCOUNT SLICE FOR THIS PERIOD
		periodAccounts := make([]*AccountValue, len(accountsSchedule))

		if i > 0 {
			for j := range accountsSchedule {
				accountHistory, err := e.core.GetAccountHistoryLatestByTime(
					ctx, accountsSchedule[j].Account.ID, loanTransaction.OrganizationID,
					loanTransaction.BranchID, *loanTransaction.PrintedDate,
				)
				if err != nil {
					return nil, eris.Wrapf(err, "error getting account history")
				}

				// Create a new account entry for this period
				periodAccounts[j] = &AccountValue{
					Account: accountHistory,
					Value:   0,                         // Will be calculated below
					Total:   accountsSchedule[j].Total, // Carry over cumulative total
				}

				switch accountHistory.Type {
				case core.AccountTypeLoan:
					// LOAN PRINCIPAL PAYMENT FORMULA:
					// Payment Amount = Principal ÷ Number of Payments
					periodAccounts[j].Value = e.provider.Service.Decimal.Clamp(
						e.provider.Service.Decimal.Divide(principal, float64(numberOfPayments)), 0, balance)

					// Update cumulative total in original slice
					accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
					periodAccounts[j].Total = accountsSchedule[j].Total

					// REMAINING BALANCE FORMULA:
					balance = e.provider.Service.Decimal.Subtract(balance, periodAccounts[j].Value)

				case core.AccountTypeFines:
					// FINES CALCULATION FORMULA:
					if daysSkipped > 0 && !accountHistory.NoGracePeriodDaily {
						periodAccounts[j].Value = e.usecase.ComputeFines(
							principal,
							accountHistory.FinesAmort,
							accountHistory.FinesMaturity,
							daysSkipped,
							loanTransaction.ModeOfPayment,
							accountHistory.NoGracePeriodDaily,
							*accountHistory,
						)

						// Update cumulative total in original slice
						accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
						periodAccounts[j].Total = accountsSchedule[j].Total
					}

				default:
					// INTEREST CALCULATION based on computation type
					// Interest calculations...
					switch accountHistory.ComputationType {
					case core.Straight:
						if accountHistory.Type == core.AccountTypeInterest || accountHistory.Type == core.AccountTypeSVFLedger {
							periodAccounts[j].Value = e.usecase.ComputeInterest(principal, accountHistory.InterestStandard, loanTransaction.ModeOfPayment)

							accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
							periodAccounts[j].Total = accountsSchedule[j].Total
						}
					case core.Diminishing:
						if accountHistory.Type == core.AccountTypeInterest || accountHistory.Type == core.AccountTypeSVFLedger {
							periodAccounts[j].Value = e.usecase.ComputeInterest(balance, accountHistory.InterestStandard, loanTransaction.ModeOfPayment)

							accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
							periodAccounts[j].Total = accountsSchedule[j].Total
						}
					case core.DiminishingStraight:
						if accountHistory.Type == core.AccountTypeInterest || accountHistory.Type == core.AccountTypeSVFLedger {
							periodAccounts[j].Value = e.usecase.ComputeInterest(balance, accountHistory.InterestStandard, loanTransaction.ModeOfPayment)

							accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
							periodAccounts[j].Total = accountsSchedule[j].Total
						}
					}
				}

				// RUNNING TOTAL FORMULAS:
				total = e.provider.Service.Decimal.Add(total, periodAccounts[j].Value)
				rowTotal = e.provider.Service.Decimal.Add(rowTotal, periodAccounts[j].Value)
			}
		} else {
			for j := range accountsSchedule {
				accountHistory, err := e.core.GetAccountHistoryLatestByTime(
					ctx, accountsSchedule[j].Account.ID, loanTransaction.OrganizationID,
					loanTransaction.BranchID, *loanTransaction.PrintedDate,
				)
				if err != nil {
					return nil, eris.Wrapf(err, "error getting account history")
				}
				periodAccounts[j] = &AccountValue{
					Account: accountHistory,
					Value:   0,
					Total:   0,
				}
			}
		}

		sort.Slice(periodAccounts, func(i, j int) bool {
			return getAccountTypePriority(
				periodAccounts[i].Account.Type) <
				getAccountTypePriority(periodAccounts[j].Account.Type)
		})

		// ✅ NOW append with period-specific accounts
		switch loanTransaction.ModeOfPayment {
		case core.LoanModeOfPaymentDaily:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ActualDate:    actualDate,
				ScheduledDate: scheduledDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      periodAccounts, // Each period has its own independent slice!
			})
			paymentDate = paymentDate.AddDate(0, 0, 1)
		case core.LoanModeOfPaymentWeekly:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      periodAccounts,
			})
			weekDay := e.core.LoanWeeklyIota(weeklyExactDay)
			paymentDate = e.nextWeekday(paymentDate, time.Weekday(weekDay))
		case core.LoanModeOfPaymentSemiMonthly:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      periodAccounts,
			})
			thisDay := paymentDate.Day()
			thisMonth := paymentDate.Month()
			thisYear := paymentDate.Year()
			loc := paymentDate.Location()
			switch {
			case thisDay < semiMonthlyExactDay1:
				paymentDate = time.Date(thisYear, thisMonth, semiMonthlyExactDay1, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			case thisDay < semiMonthlyExactDay2:
				paymentDate = time.Date(thisYear, thisMonth, semiMonthlyExactDay2, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			default:
				nextMonth := paymentDate.AddDate(0, 1, 0)
				paymentDate = time.Date(nextMonth.Year(), nextMonth.Month(), semiMonthlyExactDay1, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			}
		case core.LoanModeOfPaymentMonthly:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      periodAccounts,
			})
			loc := paymentDate.Location()
			day := paymentDate.Day()
			if isMonthlyExactDay {
				nextMonth := paymentDate.AddDate(0, 1, 0)
				paymentDate = time.Date(nextMonth.Year(), nextMonth.Month(), day, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			} else {
				paymentDate = paymentDate.AddDate(0, 0, 30)
			}
		case core.LoanModeOfPaymentQuarterly:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      periodAccounts,
			})
			paymentDate = paymentDate.AddDate(0, 3, 0)
		case core.LoanModeOfPaymentSemiAnnual:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      periodAccounts,
			})
			paymentDate = paymentDate.AddDate(0, 6, 0)
		case core.LoanModeOfPaymentLumpsum:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      periodAccounts,
			})
		case core.LoanModeOfPaymentFixedDays:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      periodAccounts,
			})
			paymentDate = paymentDate.AddDate(0, 0, 1)
		}
	}

	return &LoanTransactionAmortizationResponse{
		Entries:     e.core.LoanTransactionEntryManager.ToModels(loanTransactionEntries),
		Currency:    *e.core.CurrencyManager.ToModel(currency),
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		Total:       total,
		Schedule:    amortization,
	}, nil
}
