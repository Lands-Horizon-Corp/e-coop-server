package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
)

type AccountValue struct {
	Account core.AccountRequest `json:"account" validate:"required"`
	Value   float64             `json:"value" validate:"required,gte=0"`
	Total   float64             `json:"total" validate:"required,gte=0"`
}

type LoanAmortizationScheduleResponse struct {
	ScheduledDate time.Time       `json:"scheduled_date"`
	ActualDate    time.Time       `json:"actual_date"`
	DaysSkipped   int             `json:"days_skipped"`
	Total         float64         `json:"total"`
	Balance       float64         `json:"balance"`
	Accounts      []*AccountValue `json:"accounts"`
}

type LoanAmortizationTotalResponse struct{}

func (e Event) LoanAmortizationSchedule(ctx context.Context, loanTransactionID uuid.UUID) ([]*LoanAmortizationScheduleResponse, error) {
	result := []*LoanAmortizationScheduleResponse{}
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, loanTransactionID, "Account.Currency")
	if err != nil {
		return result, err
	}
	holidays, err := e.core.HolidayManager.Find(ctx, &core.Holiday{
		OrganizationID: loanTransaction.OrganizationID,
		BranchID:       loanTransaction.BranchID,
	})
	if err != nil {
		return result, err
	}

	numberOfPayments, err := e.usecase.LoanNumberOfPayments(loanTransaction.ModeOfPayment, loanTransaction.Terms)
	if err != nil {
		return result, err
	}

	currency := loanTransaction.Account.Currency

	// Excluding
	excludeSaturday := loanTransaction.ExcludeSaturday
	excludeSunday := loanTransaction.ExcludeSunday
	excludeHolidays := loanTransaction.ExcludeHoliday

	// Payment custom days
	isMonthlyExactDay := loanTransaction.ModeOfPaymentMonthlyExactDay
	weeklyExactDay := loanTransaction.ModeOfPaymentWeekly // expect this to be time.Weekday (0=Sunday...)
	semiMonthlyExactDay1 := loanTransaction.ModeOfPaymentSemiMonthlyPay1
	semiMonthlyExactDay2 := loanTransaction.ModeOfPaymentSemiMonthlyPay2

	// Typically, start date comes from loanTransaction (adjust as needed)
	paymentDate := time.Now().UTC()

	for range numberOfPayments {
		// Find next valid payment date (skip excluded days)
		daysSkipped := 0
		for {
			var skip bool
			if excludeSaturday {
				if sat, _ := e.isSaturday(paymentDate, currency); sat {
					skip = true
				}
			}
			if excludeSunday {
				if sun, _ := e.isSunday(paymentDate, currency); sun {
					skip = true
				}
			}
			if excludeHolidays {
				if hol, _ := e.isHoliday(paymentDate, currency, holidays); hol {
					skip = true
				}
			}
			if !skip {
				break
			}
			paymentDate = paymentDate.AddDate(0, 0, 1)
			daysSkipped++
		}

		// Store or output paymentDate here as needed
		// fmt.Println("Payment", i+1, ":", paymentDate)

		// Calculate next payment date
		switch loanTransaction.ModeOfPayment {
		case core.LoanModeOfPaymentDaily:
			paymentDate = paymentDate.AddDate(0, 0, 1)
		case core.LoanModeOfPaymentWeekly:
			weekDay := e.core.LoanWeeklyIota(weeklyExactDay)
			// Use configured weekday, expects weeklyExactDay as time.Weekday
			paymentDate = e.nextWeekday(paymentDate, time.Weekday(weekDay))
		case core.LoanModeOfPaymentSemiMonthly:
			// Expect e.g. 15 and 30 as paydays. Move to next of these
			thisDay := paymentDate.Day()
			thisMonth := paymentDate.Month()
			thisYear := paymentDate.Year()
			loc := paymentDate.Location()

			// strictly next scheduled payday
			switch {
			case thisDay < semiMonthlyExactDay1:
				paymentDate = time.Date(thisYear, thisMonth, semiMonthlyExactDay1, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			case thisDay < semiMonthlyExactDay2:
				paymentDate = time.Date(thisYear, thisMonth, semiMonthlyExactDay2, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			default:
				// Go to first date next month
				nextMonth := paymentDate.AddDate(0, 1, 0)
				paymentDate = time.Date(nextMonth.Year(), nextMonth.Month(), semiMonthlyExactDay1, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			}
		case core.LoanModeOfPaymentMonthly:
			loc := paymentDate.Location()
			day := paymentDate.Day()
			if isMonthlyExactDay {
				// next month, same day-of-month as original
				nextMonth := paymentDate.AddDate(0, 1, 0)
				paymentDate = time.Date(nextMonth.Year(), nextMonth.Month(), day, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			} else {
				// Just add 1 month (will keep day if possible)
				paymentDate = paymentDate.AddDate(0, 1, 0)
			}
		case core.LoanModeOfPaymentQuarterly:
			paymentDate = paymentDate.AddDate(0, 3, 0)
		case core.LoanModeOfPaymentSemiAnnual:
			paymentDate = paymentDate.AddDate(0, 6, 0)
		case core.LoanModeOfPaymentLumpsum:

		}
	}
	return result, nil
}
