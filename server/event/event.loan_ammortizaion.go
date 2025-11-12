package event

import (
	"context"
	"sort"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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

func (e Event) LoanAmortizationSchedule(context context.Context, ctx echo.Context, loanTransactionID uuid.UUID) (*LoanTransactionAmortizationResponse, error) {

	// ===============================
	// STEP 1: AUTHENTICATION & AUTHORIZATION
	// ===============================
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(context, ctx)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "authentication-failed",
			Description: "Failed to authenticate user organization for loan amortization schedule generation: " + err.Error(),
			Module:      "Loan Amortization",
		})
		return nil, eris.Wrap(err, "failed to authenticate user organization for loan amortization schedule")
	}

	// Validate user organization context
	if userOrg == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "authentication-failed",
			Description: "User organization context is missing - cannot generate loan amortization schedule without proper authentication",
			Module:      "Loan Amortization",
		})
		return nil, eris.New("user organization context is required for loan amortization schedule generation")
	}

	// Validate branch assignment
	if userOrg.BranchID == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "branch-validation-failed",
			Description: "User is not assigned to any branch - branch context is required for loan amortization calculations",
			Module:      "Loan Amortization",
		})
		return nil, eris.New("branch assignment is required for loan amortization schedule generation")
	}

	// ===============================
	// STEP 2: FETCH LOAN TRANSACTION DATA
	// ===============================
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, loanTransactionID, "Branch", "Account.Currency")
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "data-retrieval-failed",
			Description: "Failed to retrieve loan transaction data for amortization schedule: " + err.Error(),
			Module:      "Loan Amortization",
		})
		return nil, eris.Wrapf(err, "failed to retrieve loan transaction with ID: %s", loanTransactionID.String())
	}

	// ===============================
	// STEP 3: FETCH LOAN TRANSACTION ENTRIES
	// ===============================
	loanTransactionEntries, err := e.core.LoanTransactionEntryManager.Find(context, &core.LoanTransactionEntry{
		OrganizationID:    loanTransaction.OrganizationID,
		BranchID:          loanTransaction.BranchID,
		LoanTransactionID: loanTransaction.ID,
	})
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "data-retrieval-failed",
			Description: "Failed to retrieve loan transaction entries for amortization calculations: " + err.Error(),
			Module:      "Loan Amortization",
		})
		return nil, eris.Wrapf(err, "failed to retrieve loan transaction entries for transaction ID: %s", loanTransactionID.String())
	}

	// ===============================
	// STEP 4: CALCULATE DEBIT/CREDIT TOTALS
	// ===============================
	totalDebit, totalCredit := 0.0, 0.0
	for _, entry := range loanTransactionEntries {
		totalDebit = e.provider.Service.Decimal.Add(totalDebit, entry.Debit)
		totalCredit = e.provider.Service.Decimal.Add(totalCredit, entry.Credit)
	}

	// ===============================
	// STEP 5: FETCH RELATED ACCOUNTS & CURRENCY
	// ===============================
	currency := loanTransaction.Account.Currency
	accounts, err := e.core.AccountManager.Find(context, &core.Account{
		OrganizationID: loanTransaction.OrganizationID,
		BranchID:       loanTransaction.BranchID,
		LoanAccountID:  loanTransaction.AccountID,
		CurrencyID:     &currency.ID,
	}, "Currency")
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "data-retrieval-failed",
			Description: "Failed to retrieve loan-related accounts for amortization schedule: " + err.Error(),
			Module:      "Loan Amortization",
		})
		return nil, eris.Wrapf(err, "failed to retrieve accounts for loan transaction ID: %s", loanTransactionID.String())
	}

	// ===============================
	// STEP 6: FETCH HOLIDAY CALENDAR
	// ===============================
	holidays, err := e.core.HolidayManager.Find(context, &core.Holiday{
		OrganizationID: loanTransaction.OrganizationID,
		BranchID:       loanTransaction.BranchID,
		CurrencyID:     currency.ID,
	})
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "data-retrieval-failed",
			Description: "Failed to retrieve holiday calendar for payment schedule calculations: " + err.Error(),
			Module:      "Loan Amortization",
		})
		return nil, eris.Wrapf(err, "failed to retrieve holidays for loan amortization schedule")
	}

	// ===============================
	// STEP 7: CALCULATE NUMBER OF PAYMENTS
	// ===============================
	numberOfPayments, err := e.usecase.LoanNumberOfPayments(loanTransaction.ModeOfPayment, loanTransaction.Terms)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "calculation-failed",
			Description: "Failed to calculate number of payments for loan amortization: " + err.Error(),
			Module:      "Loan Amortization",
		})
		return nil, eris.Wrapf(err, "failed to calculate number of payments for loan with mode: %s and terms: %d",
			loanTransaction.ModeOfPayment, loanTransaction.Terms)
	}

	// ===============================
	// STEP 8: CONFIGURE PAYMENT SCHEDULE SETTINGS
	// ===============================
	// Weekend and holiday exclusions
	excludeSaturday := loanTransaction.ExcludeSaturday
	excludeSunday := loanTransaction.ExcludeSunday
	excludeHolidays := loanTransaction.ExcludeHoliday

	// Payment frequency settings
	isMonthlyExactDay := loanTransaction.ModeOfPaymentMonthlyExactDay
	weeklyExactDay := loanTransaction.ModeOfPaymentWeekly
	semiMonthlyExactDay1 := loanTransaction.ModeOfPaymentSemiMonthlyPay1
	semiMonthlyExactDay2 := loanTransaction.ModeOfPaymentSemiMonthlyPay2

	// ===============================
	// STEP 9: INITIALIZE AMORTIZATION VARIABLES
	// ===============================
	accountsSchedule := []*AccountValue{}
	amortization := []*LoanAmortizationSchedule{}

	// Setup account schedule with related accounts
	for _, acc := range accounts {
		accountsSchedule = append(accountsSchedule, &AccountValue{
			Account: acc,
			Value:   0,
			Total:   0,
		})
	}

	// Add the main loan account
	accountsSchedule = append(accountsSchedule, &AccountValue{
		Account: loanTransaction.Account,
		Value:   0,
		Total:   0,
	})

	// Initialize payment calculation variables
	paymentDate := userOrg.UserOrgTime()
	if loanTransaction.PrintedDate != nil {
		paymentDate = loanTransaction.PrintedDate.UTC()
	}
	principal := totalCredit
	balance := totalCredit
	total := 0.0

	// ===============================
	// STEP 10: GENERATE AMORTIZATION SCHEDULE
	// ===============================
	for i := range numberOfPayments {
		actualDate := paymentDate
		daysSkipped := 0
		rowTotal := 0.0

		// Calculate skipped days due to weekends/holidays
		daysSkipped, err := e.skippedDaysCount(paymentDate, currency, excludeSaturday, excludeSunday, excludeHolidays, holidays)
		if err != nil {
			e.Footstep(ctx, FootstepEvent{
				Activity:    "calculation-failed",
				Description: "Failed to calculate skipped days for payment schedule: " + err.Error(),
				Module:      "Loan Amortization",
			})
			return nil, eris.Wrapf(err, "failed to calculate skipped days for payment date: %s", paymentDate.Format("2006-01-02"))
		}

		scheduledDate := paymentDate.AddDate(0, 0, daysSkipped)

		// ===============================
		// STEP 11: CREATE PERIOD-SPECIFIC ACCOUNT CALCULATIONS
		// ===============================
		periodAccounts := make([]*AccountValue, len(accountsSchedule))

		if i > 0 {
			// For subsequent payments, calculate based on account history
			for j := range accountsSchedule {
				accountHistory, err := e.core.GetAccountHistoryLatestByTime(
					context, accountsSchedule[j].Account.ID, loanTransaction.OrganizationID,
					loanTransaction.BranchID, loanTransaction.PrintedDate,
				)
				if err != nil {
					e.Footstep(ctx, FootstepEvent{
						Activity:    "calculation-failed",
						Description: "Failed to retrieve account history for amortization calculations: " + err.Error(),
						Module:      "Loan Amortization",
					})
					return nil, eris.Wrapf(err, "failed to get account history for account ID: %s", accountsSchedule[j].Account.ID.String())
				}

				// Create new account entry for this period
				periodAccounts[j] = &AccountValue{
					Account: accountHistory,
					Value:   0,
					Total:   accountsSchedule[j].Total,
				}

				// ===============================
				// STEP 12: CALCULATE ACCOUNT-SPECIFIC VALUES
				// ===============================
				switch accountHistory.Type {
				case core.AccountTypeLoan:
					// LOAN PRINCIPAL PAYMENT: Principal ÷ Number of Payments
					periodAccounts[j].Value = e.provider.Service.Decimal.Clamp(
						e.provider.Service.Decimal.Divide(principal, float64(numberOfPayments)), 0, balance)

					// Update cumulative totals
					accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
					periodAccounts[j].Total = accountsSchedule[j].Total

					// Update remaining balance
					balance = e.provider.Service.Decimal.Subtract(balance, periodAccounts[j].Value)

				case core.AccountTypeFines:
					// FINES CALCULATION: Based on days skipped and penalty rates
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

						accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
						periodAccounts[j].Total = accountsSchedule[j].Total
					}

				default:
					// ===============================
					// STEP 13: INTEREST CALCULATIONS
					// ===============================
					switch accountHistory.ComputationType {
					case core.Straight:
						if accountHistory.Type == core.AccountTypeInterest || accountHistory.Type == core.AccountTypeSVFLedger {
							// STRAIGHT INTEREST: Fixed percentage of original principal
							periodAccounts[j].Value = e.usecase.ComputeInterest(principal, accountHistory.InterestStandard, loanTransaction.ModeOfPayment)

							accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
							periodAccounts[j].Total = accountsSchedule[j].Total
						}
					case core.Diminishing:
						if accountHistory.Type == core.AccountTypeInterest || accountHistory.Type == core.AccountTypeSVFLedger {
							// DIMINISHING INTEREST: Percentage of remaining balance
							periodAccounts[j].Value = e.usecase.ComputeInterest(balance, accountHistory.InterestStandard, loanTransaction.ModeOfPayment)

							accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
							periodAccounts[j].Total = accountsSchedule[j].Total
						}
					case core.DiminishingStraight:
						if accountHistory.Type == core.AccountTypeInterest || accountHistory.Type == core.AccountTypeSVFLedger {
							// DIMINISHING STRAIGHT: Hybrid calculation method
							periodAccounts[j].Value = e.usecase.ComputeInterest(balance, accountHistory.InterestStandard, loanTransaction.ModeOfPayment)

							accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
							periodAccounts[j].Total = accountsSchedule[j].Total
						}
					}
				}

				// Update running totals
				total = e.provider.Service.Decimal.Add(total, periodAccounts[j].Value)
				rowTotal = e.provider.Service.Decimal.Add(rowTotal, periodAccounts[j].Value)
			}
		} else {
			// For first payment, initialize account entries
			for j := range accountsSchedule {
				accountHistory, err := e.core.GetAccountHistoryLatestByTime(
					context, accountsSchedule[j].Account.ID, loanTransaction.OrganizationID,
					loanTransaction.BranchID, loanTransaction.PrintedDate,
				)
				if err != nil {
					e.Footstep(ctx, FootstepEvent{
						Activity:    "initialization-failed",
						Description: "Failed to initialize account history for first payment: " + err.Error(),
						Module:      "Loan Amortization",
					})
					return nil, eris.Wrapf(err, "failed to initialize account history for first payment")
				}

				periodAccounts[j] = &AccountValue{
					Account: accountHistory,
					Value:   0,
					Total:   0,
				}
				if periodAccounts[j].Account.Type == core.AccountTypeLoan {
					// LOAN PRINCIPAL PAYMENT: Principal ÷ Number of Payments
					periodAccounts[j].Value = e.provider.Service.Decimal.Clamp(
						e.provider.Service.Decimal.Divide(principal, float64(numberOfPayments)), 0, balance)

					// Update cumulative totals
					accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
					periodAccounts[j].Total = accountsSchedule[j].Total

					// Update remaining balance
					balance = e.provider.Service.Decimal.Subtract(balance, periodAccounts[j].Value)
				}
			}
		}

		// Sort accounts by type priority for consistent ordering
		sort.Slice(periodAccounts, func(i, j int) bool {
			return getAccountTypePriority(periodAccounts[i].Account.Type) <
				getAccountTypePriority(periodAccounts[j].Account.Type)
		})

		// ===============================
		// STEP 14: DETERMINE NEXT PAYMENT DATE
		// ===============================
		switch loanTransaction.ModeOfPayment {
		case core.LoanModeOfPaymentDaily:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ActualDate:    actualDate,
				ScheduledDate: scheduledDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      periodAccounts,
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

	// ===============================
	// STEP 15: RETURN COMPLETE AMORTIZATION SCHEDULE
	// ===============================
	return &LoanTransactionAmortizationResponse{
		Entries:     e.core.LoanTransactionEntryManager.ToModels(loanTransactionEntries),
		Currency:    *e.core.CurrencyManager.ToModel(currency),
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		Total:       total,
		Schedule:    amortization,
	}, nil
}
