package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
)

type AccountValue struct {
	Account core.Account `json:"account" validate:"required"`
	Value   float64      `json:"value" validate:"required,gte=0"`
	Total   float64      `json:"total" validate:"required,gte=0"`
}

type LoanAmortizationSchedule struct {
	ScheduledDate time.Time      `json:"scheduled_date"`
	ActualDate    time.Time      `json:"actual_date"`
	DaysSkipped   int            `json:"days_skipped"`
	Total         float64        `json:"total"`
	Balance       float64        `json:"balance"`
	Accounts      []AccountValue `json:"accounts"`
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

	accountsSchedule := []AccountValue{}

	amortization := []*LoanAmortizationSchedule{}
	for _, acc := range accounts {
		accountsSchedule = append(accountsSchedule, AccountValue{
			Account: *acc,
			Value:   0,
			Total:   0,
		})
	}
	accountsSchedule = append(accountsSchedule, AccountValue{
		Account: *loanTransaction.Account,
		Value:   0,
		Total:   0,
	})

	// Typically, start date comes from loanTransaction (adjust as needed)
	paymentDate := time.Now().UTC()
	principal := totalCredit
	balance := totalCredit
	total := 0.0

	for range numberOfPayments {
		actualDate := paymentDate
		daysSkipped := 0
		rowTotal := 0.0
		daysSkipped, err := e.skippedDaysCount(paymentDate, currency, excludeSaturday, excludeSunday, excludeHolidays, holidays)
		if err != nil {
			return nil, err
		}
		for j := range accountsSchedule {
			switch accountsSchedule[j].Account.Type {
			case core.AccountTypeLoan:
				// LOAN PRINCIPAL PAYMENT FORMULA:
				// Payment Amount = Principal ÷ Number of Payments
				// Clamped to ensure we don't pay more than remaining balance
				// Formula: min(Principal/NumberOfPayments, RemainingBalance)
				accountsSchedule[j].Value = e.provider.Service.Decimal.Clamp(
					e.provider.Service.Decimal.Divide(principal, float64(numberOfPayments)), 0, balance)

				// CUMULATIVE TOTAL FORMULA:
				// Total = Previous Total + Current Payment
				accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, accountsSchedule[j].Value)

				// REMAINING BALANCE FORMULA:
				// New Balance = Previous Balance - Principal Payment
				balance = e.provider.Service.Decimal.Subtract(balance, accountsSchedule[j].Value)

			case core.AccountTypeFines:
				// FINES CALCULATION FORMULA:
				// Only apply fines if:
				// 1. Days skipped > 0 (payment is late)
				// 2. Account doesn't have NoGracePeriodDaily flag
				// Formula: ComputeFines(principal, fines_rate, maturity_rate, days_late, payment_mode)
				if daysSkipped > 0 && !accountsSchedule[j].Account.NoGracePeriodDaily {
					accountsSchedule[j].Value = e.usecase.ComputeFines(
						principal,
						accountsSchedule[j].Account.FinesAmort,
						accountsSchedule[j].Account.FinesMaturity,
						daysSkipped,
						loanTransaction.ModeOfPayment,
						accountsSchedule[j].Account.NoGracePeriodDaily,
						accountsSchedule[j].Account,
					)
					// CUMULATIVE FINES TOTAL FORMULA:
					// Total Fines = Previous Total Fines + Current Period Fines
					accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, accountsSchedule[j].Value)
				}

			default:
				// INTEREST CALCULATION based on computation type
				switch accountsSchedule[j].Account.ComputationType {
				case core.Straight:
					// STRAIGHT LINE INTEREST FORMULA:
					// Interest is calculated on the original principal amount
					// Formula: Interest = Principal × Interest Rate ÷ Payment Frequency
					switch accountsSchedule[j].Account.Type {
					case core.AccountTypeInterest:
						// STRAIGHT INTEREST ON PRINCIPAL:
						// Uses original principal amount throughout the loan term
						accountsSchedule[j].Value = e.usecase.ComputeInterest(principal, accountsSchedule[j].Account.InterestStandard, loanTransaction.ModeOfPayment)
						accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, accountsSchedule[j].Value)
					case core.AccountTypeSVFLedger:
						// SVF LEDGER STRAIGHT INTEREST:
						// Special Voluntary Fund interest calculated on original principal
						accountsSchedule[j].Value = e.usecase.ComputeInterest(principal, accountsSchedule[j].Account.InterestStandard, loanTransaction.ModeOfPayment)
						accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, accountsSchedule[j].Value)
					}

				case core.Diminishing:
					// DIMINISHING BALANCE INTEREST FORMULA:
					// Interest is calculated on the remaining balance
					// Formula: Interest = Remaining Balance × Interest Rate ÷ Payment Frequency
					switch accountsSchedule[j].Account.Type {
					case core.AccountTypeInterest:
						// DIMINISHING INTEREST ON BALANCE:
						// Uses current remaining balance (decreases each payment)
						accountsSchedule[j].Value = e.usecase.ComputeInterest(balance, accountsSchedule[j].Account.InterestStandard, loanTransaction.ModeOfPayment)
						accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, accountsSchedule[j].Value)
					case core.AccountTypeSVFLedger:
						// SVF LEDGER DIMINISHING - No calculation defined
						// This case is intentionally left empty
					}

				case core.DiminishingStraight:
					// DIMINISHING STRAIGHT INTEREST FORMULA:
					// Hybrid approach - uses remaining balance for calculation
					// Formula: Interest = Remaining Balance × Interest Rate ÷ Payment Frequency
					switch accountsSchedule[j].Account.Type {
					case core.AccountTypeInterest:
						// DIMINISHING STRAIGHT INTEREST ON BALANCE:
						// Uses current remaining balance like diminishing method
						accountsSchedule[j].Value = e.usecase.ComputeInterest(balance, accountsSchedule[j].Account.InterestStandard, loanTransaction.ModeOfPayment)
						accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, accountsSchedule[j].Value)
					case core.AccountTypeSVFLedger:
						// SVF LEDGER DIMINISHING STRAIGHT:
						// Uses remaining balance for SVF calculations
						accountsSchedule[j].Value = e.usecase.ComputeInterest(balance, accountsSchedule[j].Account.InterestStandard, loanTransaction.ModeOfPayment)
						accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, accountsSchedule[j].Value)
					}
				}
			}

			// RUNNING TOTAL FORMULAS:
			// Grand Total = Sum of all account values for all periods
			total = e.provider.Service.Decimal.Add(total, accountsSchedule[j].Value)
			// Row Total = Sum of all account values for current period
			rowTotal = e.provider.Service.Decimal.Add(rowTotal, accountsSchedule[j].Value)
		}
		scheduledDate := paymentDate.AddDate(0, 0, daysSkipped)
		switch loanTransaction.ModeOfPayment {
		case core.LoanModeOfPaymentDaily:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ActualDate:    actualDate,
				ScheduledDate: scheduledDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      accountsSchedule,
			})
			paymentDate = paymentDate.AddDate(0, 0, 1)
		case core.LoanModeOfPaymentWeekly:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      accountsSchedule,
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
				Accounts:      accountsSchedule,
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
				Accounts:      accountsSchedule,
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
				Accounts:      accountsSchedule,
			})
			paymentDate = paymentDate.AddDate(0, 3, 0)
		case core.LoanModeOfPaymentSemiAnnual:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      accountsSchedule,
			})
			paymentDate = paymentDate.AddDate(0, 6, 0)
		case core.LoanModeOfPaymentLumpsum:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      accountsSchedule,
			})
		case core.LoanModeOfPaymentFixedDays:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      accountsSchedule,
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
