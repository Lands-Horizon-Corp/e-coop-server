package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
)

// helper: returns the next date after 'from' that falls on 'weekday'
func nextWeekday(from time.Time, weekday time.Weekday) time.Time {
	// Move to the next day to avoid returning the current day if matches
	d := from.AddDate(0, 0, 1)
	for d.Weekday() != weekday {
		d = d.AddDate(0, 0, 1)
	}
	return d
}

func (e Event) LoanAmortizationSchedule(ctx context.Context, loanTransactionID uuid.UUID) error {
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, loanTransactionID, "Account.Currency")
	if err != nil {
		return err
	}
	holidays, err := e.core.HolidayManager.Find(ctx, &core.Holiday{
		OrganizationID: loanTransaction.OrganizationID,
		BranchID:       loanTransaction.BranchID,
	})

	numberOfPayments, err := e.usecase.LoanNumberOfPayments(loanTransaction)
	if err != nil {
		return err
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

	for i := 0; i < numberOfPayments; i++ {
		// Find next valid payment date (skip excluded days)
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
			paymentDate = nextWeekday(paymentDate, time.Weekday(weekDay))
		case core.LoanModeOfPaymentSemiMonthly:
			// Expect e.g. 15 and 30 as paydays. Move to next of these
			thisDay := paymentDate.Day()
			thisMonth := paymentDate.Month()
			thisYear := paymentDate.Year()
			loc := paymentDate.Location()

			// strictly next scheduled payday
			if thisDay < semiMonthlyExactDay1 {
				paymentDate = time.Date(thisYear, thisMonth, semiMonthlyExactDay1, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			} else if thisDay < semiMonthlyExactDay2 {
				paymentDate = time.Date(thisYear, thisMonth, semiMonthlyExactDay2, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			} else {
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
			// Usually, lumpsum means all due at once, so break after first
			if i == 0 {
				// (store/output here) and then break outside for loop if needed
			}
		}
	}
	return nil
}

func (e Event) isHoliday(date time.Time, currency *core.Currency, holidays []*core.Holiday) (bool, error) {
	loc, err := time.LoadLocation(currency.Timezone)
	if err != nil {
		return false, err
	}
	localDate := date.In(loc).Truncate(24 * time.Hour)
	for _, holiday := range holidays {
		holidayDate := holiday.EntryDate.In(loc).Truncate(24 * time.Hour)
		if localDate.Equal(holidayDate) {
			return true, nil
		}
	}
	return false, nil
}
func (e Event) isSunday(date time.Time, currency *core.Currency) (bool, error) {
	loc, err := time.LoadLocation(currency.Timezone)
	if err != nil {
		return false, err
	}
	localDate := date.In(loc)
	weekday := localDate.Weekday()
	if weekday == time.Sunday {
		return true, nil
	}
	return false, nil
}

func (e Event) isSaturday(date time.Time, currency *core.Currency) (bool, error) {
	loc, err := time.LoadLocation(currency.Timezone)
	if err != nil {
		return false, err
	}
	localDate := date.In(loc)
	weekday := localDate.Weekday()
	if weekday == time.Saturday {
		return true, nil
	}
	return false, nil
}
