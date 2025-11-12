package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func (e *Event) LoanProcessing(context context.Context, ctx echo.Context, loanTransactionID *uuid.UUID) (*core.LoanTransaction, error) {
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
	if err != nil {
		return nil, eris.Wrap(err, "loan processing: failed to get loan transaction by id")
	}

	// Get current user organization
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(context, ctx)
	if err != nil {
		return nil, eris.Wrap(err, "loan processing: failed to get current user organization")
	}
	if userOrg.BranchID == nil {
		return nil, eris.New("loan processing: user organization has no branch assigned")
	}

	// // ===============================
	// // STEP 5: FETCH RELATED ACCOUNTS & CURRENCY
	// // ===============================
	// currency := loanTransaction.Account.Currency
	// accounts, err := e.core.AccountManager.Find(context, &core.Account{
	// 	OrganizationID: userOrg.OrganizationID,
	// 	BranchID:       *userOrg.BranchID,
	// 	LoanAccountID:  loanTransaction.AccountID,
	// 	CurrencyID:     &currency.ID,
	// }, "Currency")
	// if err != nil {
	// 	e.Footstep(ctx, FootstepEvent{
	// 		Activity:    "data-retrieval-failed",
	// 		Description: "Failed to retrieve loan-related accounts for amortization schedule: " + err.Error(),
	// 		Module:      "Loan Amortization",
	// 	})
	// 	return nil, eris.Wrapf(err, "failed to retrieve accounts for loan transaction ID: %s", loanTransactionID.String())
	// }

	// // ===============================
	// // STEP 6: FETCH HOLIDAY CALENDAR
	// // ===============================
	// holidays, err := e.core.HolidayManager.Find(context, &core.Holiday{
	// 	OrganizationID: userOrg.OrganizationID,
	// 	BranchID:       *userOrg.BranchID,
	// 	CurrencyID:     currency.ID,
	// })
	// if err != nil {
	// 	e.Footstep(ctx, FootstepEvent{
	// 		Activity:    "data-retrieval-failed",
	// 		Description: "Failed to retrieve holiday calendar for payment schedule calculations: " + err.Error(),
	// 		Module:      "Loan Amortization",
	// 	})
	// 	return nil, eris.Wrapf(err, "failed to retrieve holidays for loan amortization schedule")
	// }

	// // ===============================
	// // STEP 7: CALCULATE NUMBER OF PAYMENTS
	// // ===============================
	// numberOfPayments, err := e.usecase.LoanNumberOfPayments(loanTransaction.ModeOfPayment, loanTransaction.Terms)
	// if err != nil {
	// 	e.Footstep(ctx, FootstepEvent{
	// 		Activity:    "calculation-failed",
	// 		Description: "Failed to calculate number of payments for loan amortization: " + err.Error(),
	// 		Module:      "Loan Amortization",
	// 	})
	// 	return nil, eris.Wrapf(err, "failed to calculate number of payments for loan with mode: %s and terms: %d",
	// 		loanTransaction.ModeOfPayment, loanTransaction.Terms)
	// }

	// // ===============================
	// // STEP 8: CONFIGURE PAYMENT SCHEDULE SETTINGS
	// // ===============================
	// // Weekend and holiday exclusions
	// excludeSaturday := loanTransaction.ExcludeSaturday
	// excludeSunday := loanTransaction.ExcludeSunday
	// excludeHolidays := loanTransaction.ExcludeHoliday

	// // Payment frequency settings
	// isMonthlyExactDay := loanTransaction.ModeOfPaymentMonthlyExactDay
	// weeklyExactDay := loanTransaction.ModeOfPaymentWeekly
	// semiMonthlyExactDay1 := loanTransaction.ModeOfPaymentSemiMonthlyPay1
	// semiMonthlyExactDay2 := loanTransaction.ModeOfPaymentSemiMonthlyPay2

	// if loanTransaction.PrintedDate == nil {
	// 	return nil, eris.New("loan processing: printed date is nil")
	// }
	// // Initialize payment calculation variables
	// currentDate := time.Now().UTC()
	// if userOrg.TimeMachineTime != nil {
	// 	currentDate = userOrg.UserOrgTime()
	// }
	// paymentDate := *loanTransaction.PrintedDate
	// // accounts, err := e.usecase.
	// for i := range numberOfPayments + 1 {
	// 	actualDate := paymentDate
	// 	daysSkipped := 0
	// 	rowTotal := 0.0
	// 	daysSkipped, err := e.skippedDaysCount(paymentDate, currency, excludeSaturday, excludeSunday, excludeHolidays, holidays)
	// 	if err != nil {
	// 		e.Footstep(ctx, FootstepEvent{
	// 			Activity:    "calculation-failed",
	// 			Description: "Failed to calculate skipped days for payment schedule: " + err.Error(),
	// 			Module:      "Loan Amortization",
	// 		})
	// 		return nil, eris.Wrapf(err, "failed to calculate skipped days for payment date: %s", paymentDate.Format("2006-01-02"))
	// 	}
	// 	scheduledDate := paymentDate.AddDate(0, 0, daysSkipped)

	// 	if loanTransaction.LoanCount >= i && scheduledDate.Before(currentDate) {

	// 		// loanTransaction.LoanCount = i + 1

	// 	}
	// 	// ===============================
	// 	// STEP 14: DETERMINE NEXT PAYMENT DATE
	// 	// ===============================
	// 	switch loanTransaction.ModeOfPayment {
	// 	case core.LoanModeOfPaymentDaily:
	// 		paymentDate = paymentDate.AddDate(0, 0, 1)
	// 	case core.LoanModeOfPaymentWeekly:
	// 		weekDay := e.core.LoanWeeklyIota(weeklyExactDay)
	// 		paymentDate = e.nextWeekday(paymentDate, time.Weekday(weekDay))
	// 	case core.LoanModeOfPaymentSemiMonthly:
	// 		thisDay := paymentDate.Day()
	// 		thisMonth := paymentDate.Month()
	// 		thisYear := paymentDate.Year()
	// 		loc := paymentDate.Location()
	// 		switch {
	// 		case thisDay < semiMonthlyExactDay1:
	// 			paymentDate = time.Date(thisYear, thisMonth, semiMonthlyExactDay1, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
	// 		case thisDay < semiMonthlyExactDay2:
	// 			paymentDate = time.Date(thisYear, thisMonth, semiMonthlyExactDay2, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
	// 		default:
	// 			nextMonth := paymentDate.AddDate(0, 1, 0)
	// 			paymentDate = time.Date(nextMonth.Year(), nextMonth.Month(), semiMonthlyExactDay1, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
	// 		}
	// 	case core.LoanModeOfPaymentMonthly:
	// 		loc := paymentDate.Location()
	// 		day := paymentDate.Day()
	// 		if isMonthlyExactDay {
	// 			nextMonth := paymentDate.AddDate(0, 1, 0)
	// 			paymentDate = time.Date(nextMonth.Year(), nextMonth.Month(), day, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
	// 		} else {
	// 			paymentDate = paymentDate.AddDate(0, 0, 30)
	// 		}
	// 	case core.LoanModeOfPaymentQuarterly:
	// 		paymentDate = paymentDate.AddDate(0, 3, 0)
	// 	case core.LoanModeOfPaymentSemiAnnual:
	// 		paymentDate = paymentDate.AddDate(0, 6, 0)
	// 	case core.LoanModeOfPaymentLumpsum:
	// 	case core.LoanModeOfPaymentFixedDays:
	// 		paymentDate = paymentDate.AddDate(0, 0, 1)
	// 	}
	// }
	return loanTransaction, nil
}
