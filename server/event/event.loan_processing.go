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
	user, err := e.userOrganizationToken.CurrentUserOrganization(context, ctx)
	if err != nil {
		return nil, eris.Wrap(err, "loan processing: failed to get current user organization")
	}
	if user.BranchID == nil {
		return nil, eris.New("loan processing: user organization has no branch assigned")
	}

	// ===============================
	// STEP 5: FETCH RELATED ACCOUNTS & CURRENCY
	// ===============================
	// currency := loanTransaction.Account.Currency
	// accounts, err := e.core.AccountManager.Find(context, &core.Account{
	// 	OrganizationID: loanTransaction.OrganizationID,
	// 	BranchID:       loanTransaction.BranchID,
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
	// 	OrganizationID: loanTransaction.OrganizationID,
	// 	BranchID:       loanTransaction.BranchID,
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

	return loanTransaction, nil
}
