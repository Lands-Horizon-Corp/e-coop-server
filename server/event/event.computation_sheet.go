package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

// LoanComputationSheetCalculatorRequest represents the request structure for creating/updating loancomputationsheetcalculator
type LoanComputationSheetCalculatorRequest struct {
	AccountID    *uuid.UUID `json:"account_id,omitempty"`
	Applied1     float64    `json:"applied_1"`
	Terms        int        `json:"terms"`
	MemberTypeID *uuid.UUID `json:"member_type_id,omitempty"`
	IsAddOn      bool       `json:"is_add_on,omitempty"`

	ExcludeSaturday              bool          `json:"exclude_saturday"`
	ExcludeSunday                bool          `json:"exclude_sunday"`
	ExcludeHoliday               bool          `json:"exclude_holiday"`
	ModeOfPaymentMonthlyExactDay bool          `json:"mode_of_payment_monthly_exact_day"`
	ModeOfPaymentWeekly          core.Weekdays `json:"mode_of_payment_weekly"`
	ModeOfPaymentSemiMonthlyPay1 int           `json:"mode_of_payment_semi_monthly_pay_1"`
	ModeOfPaymentSemiMonthlyPay2 int           `json:"mode_of_payment_semi_monthly_pay_2"`

	ModeOfPayment core.LoanModeOfPayment `json:"mode_of_payment"`
	Accounts      []*core.Account        `json:"accounts,omitempty"`

	CashOnHandAccountID *uuid.UUID `json:"cash_on_hand_account_id,omitempty"`
	ComputationSheetID  *uuid.UUID `json:"computation_sheet_id,omitempty"`
}

type ComputationSheetAmortizationResponse struct {
	Entries     []*core.LoanTransactionEntryResponse `json:"entries"`
	TotalDebit  float64                              `json:"total_debit"`
	TotalCredit float64                              `json:"total_credit"`

	Schedule []*LoanAmortizationScheduleResponse `json:"schedule,omitempty"`
}

func (e *Event) ComputationSheetCalculator(
	context context.Context,

	lcscr LoanComputationSheetCalculatorRequest,
) (*ComputationSheetAmortizationResponse, error) {
	computationSheet, err := e.core.ComputationSheetManager.GetByID(context, *lcscr.ComputationSheetID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get computation sheet")
	}

	automaticLoanDeductionEntries, err := e.core.AutomaticLoanDeductionManager.Find(context, &core.AutomaticLoanDeduction{
		ComputationSheetID: &computationSheet.ID,
		BranchID:           computationSheet.BranchID,
		OrganizationID:     computationSheet.OrganizationID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to find automatic loan deduction")
	}
	account, err := e.core.AccountManager.GetByID(context, *lcscr.AccountID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get account")
	}
	cashOnHand, err := e.core.AccountManager.GetByID(context, *lcscr.CashOnHandAccountID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get cash on hand account")
	}
	loanTransactionEntries := []*core.LoanTransactionEntry{
		{
			Account: cashOnHand,
			IsAddOn: false,
			Type:    core.LoanTransactionStatic,
			Debit:   0,
			Credit:  lcscr.Applied1,
			Name:    account.Name,
		},
		{
			Account: account,
			IsAddOn: false,
			Type:    core.LoanTransactionStatic,
			Debit:   lcscr.Applied1,
			Credit:  0,
			Name:    cashOnHand.Name,
		},
	}
	addOnEntry := &core.LoanTransactionEntry{
		Account: nil,
		Credit:  0,
		Debit:   0,
		Name:    "ADD ON INTEREST",
		Type:    core.LoanTransactionAddOn,
		IsAddOn: true,
	}
	totalNonAddOns, totalAddOns := 0.0, 0.0
	for _, ald := range automaticLoanDeductionEntries {
		if ald.AccountID == nil {
			continue
		}
		ald.Account, err = e.core.AccountManager.GetByID(context, *ald.AccountID)
		if err != nil {
			continue
		}
		entry := &core.LoanTransactionEntry{
			Credit:  0,
			Debit:   0,
			Name:    ald.Name,
			Type:    core.LoanTransactionDeduction,
			IsAddOn: ald.AddOn,
			Account: ald.Account,
		}
		if entry.AutomaticLoanDeduction.ChargesRateSchemeID != nil {
			chargesRateScheme, err := e.core.ChargesRateSchemeManager.GetByID(context, *entry.AutomaticLoanDeduction.ChargesRateSchemeID)
			if err != nil {
				return nil, eris.Wrap(err, fmt.Sprintf("failed to get charges rate scheme for automatic loan deduction ID %s", ald.ID))
			}
			entry.Credit = e.usecase.LoanChargesRateComputation(context, *chargesRateScheme, core.LoanTransaction{
				Applied1: lcscr.Applied1,
				Terms:    lcscr.Terms,
				MemberProfile: &core.MemberProfile{
					MemberTypeID: lcscr.MemberTypeID,
				},
			})

		}
		if entry.Credit <= 0 {
			entry.Credit = e.usecase.LoanComputation(*ald, core.LoanTransaction{
				Terms:    lcscr.Terms,
				Applied1: lcscr.Applied1,
			})
		}
		if !entry.IsAddOn {
			totalNonAddOns = e.provider.Service.Decimal.Add(totalNonAddOns, entry.Credit)
		} else {
			totalAddOns = e.provider.Service.Decimal.Add(totalAddOns, entry.Credit)
		}
		if entry.Credit > 0 {
			loanTransactionEntries = append(loanTransactionEntries, entry)
		}
	}
	if lcscr.IsAddOn {
		loanTransactionEntries[0].Credit = e.provider.Service.Decimal.Subtract(lcscr.Applied1, totalNonAddOns)
	} else {
		loanTransactionEntries[0].Credit = e.provider.Service.Decimal.Subtract(lcscr.Applied1, e.provider.Service.Decimal.Add(totalNonAddOns, totalAddOns))
	}
	if lcscr.IsAddOn {
		addOnEntry.Debit = totalAddOns
		loanTransactionEntries = append(loanTransactionEntries, addOnEntry)
	}

	totalDebit, totalCredit := 0.0, 0.0
	for _, entry := range loanTransactionEntries {
		totalDebit = e.provider.Service.Decimal.Add(totalDebit, entry.Debit)
		totalCredit = e.provider.Service.Decimal.Add(totalCredit, entry.Credit)
	}

	// Loan Amortization Schedule ==========================================
	holidays, err := e.core.HolidayManager.Find(context, &core.Holiday{
		OrganizationID: computationSheet.OrganizationID,
		BranchID:       computationSheet.BranchID,
	})
	if err != nil {
		return nil, err
	}

	numberOfPayments, err := e.usecase.LoanNumberOfPayments(lcscr.ModeOfPayment, lcscr.Terms)
	if err != nil {
		return nil, err
	}

	currency := account.Currency

	// Excluding
	excludeSaturday := lcscr.ExcludeSaturday
	excludeSunday := lcscr.ExcludeSunday
	excludeHolidays := lcscr.ExcludeHoliday

	// Payment custom days
	isMonthlyExactDay := lcscr.ModeOfPaymentMonthlyExactDay
	weeklyExactDay := lcscr.ModeOfPaymentWeekly // expect this to be time.Weekday (0=Sunday...)
	semiMonthlyExactDay1 := lcscr.ModeOfPaymentSemiMonthlyPay1
	semiMonthlyExactDay2 := lcscr.ModeOfPaymentSemiMonthlyPay2

	// Typically, start date comes from loanTransaction (adjust as needed)
	amortization := []*LoanAmortizationScheduleResponse{}
	accounts := []*AccountValue{
		{Account: *account, Value: totalCredit},
	}
	for _, acc := range lcscr.Accounts {
		accounts = append(accounts, &AccountValue{
			Account: *acc,
			Value:   0,
		})
	}
	paymentDate := time.Now().UTC()
	for i := range numberOfPayments {

		// Find next valid payment date (skip excluded days)
		actualDate := paymentDate
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

		for _, acc := range accounts {
			switch acc.Account.ComputationType {
			case core.Straight:
			case core.Diminishing:
			case core.DiminishingStraight:
			}
		}

		// Calculate next payment date
		switch lcscr.ModeOfPayment {
		case core.LoanModeOfPaymentDaily:
			//====================================================================
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
			paymentDate = paymentDate.AddDate(0, 0, 1)
		case core.LoanModeOfPaymentWeekly:
			//====================================================================
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
			weekDay := e.core.LoanWeeklyIota(weeklyExactDay)
			paymentDate = e.nextWeekday(paymentDate, time.Weekday(weekDay))

		case core.LoanModeOfPaymentSemiMonthly:
			//====================================================================
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
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
			//====================================================================
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
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
			//====================================================================
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
			paymentDate = paymentDate.AddDate(0, 3, 0)

		case core.LoanModeOfPaymentSemiAnnual:
			//====================================================================
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
			paymentDate = paymentDate.AddDate(0, 6, 0)

		case core.LoanModeOfPaymentLumpsum:
			//====================================================================
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
			// Usually, lumpsum means all due at once, so break after first
			if i == 0 {
				// (store/output here) and then break outside for loop if needed
			}
		case core.LoanModeOfPaymentFixedDays:
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
			// Add logic for fixed days payment mode
			paymentDate = paymentDate.AddDate(0, 0, 1) // Adjust as needed for fixed days
		}
	}

	return &ComputationSheetAmortizationResponse{
		Entries:     e.core.LoanTransactionEntryManager.ToModels(loanTransactionEntries),
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		Schedule:    amortization,
	}, nil
}

func (e *Event) sumAccountValues(accountValues []*AccountValue) float64 {
	total := 0.0
	for _, av := range accountValues {
		if av == nil {
			continue
		}
		total = e.provider.Service.Decimal.Add(total, av.Value)
	}
	return total
}
