package event

import (
	"context"
	"sort"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/usecase"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

type AccountValue struct {
	Account *types.Account `json:"account" validate:"required"`
	Value   float64        `json:"value" validate:"required,gte=0"`
	Total   float64        `json:"total" validate:"required,gte=0"`
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
	Entries     []*types.LoanTransactionEntryResponse `json:"entries"`
	TotalDebit  float64                               `json:"total_debit"`
	TotalCredit float64                               `json:"total_credit"`
	Currency    types.CurrencyResponse                `json:"currency"`
	Total       float64                               `json:"total"`
	Schedule    []*LoanAmortizationSchedule           `json:"schedule,omitempty"`
}

func LoanAmortization(context context.Context, service *horizon.HorizonService, loanTransactionID uuid.UUID, userOrg *types.UserOrganization) (*LoanTransactionAmortizationResponse, error) {
	if userOrg == nil {
		return nil, eris.New("user organization context is required for loan amortization schedule generation")
	}
	if userOrg.BranchID == nil {
		return nil, eris.New("branch assignment is required for loan amortization schedule generation")
	}
	loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, loanTransactionID, "Branch", "Account.Currency")
	if err != nil {
		return nil, eris.Wrapf(err, "failed to retrieve loan transaction with ID: %s", loanTransactionID.String())
	}
	loanTransactionEntries, err := core.LoanTransactionEntryManager(service).Find(context, &types.LoanTransactionEntry{
		OrganizationID:    loanTransaction.OrganizationID,
		BranchID:          loanTransaction.BranchID,
		LoanTransactionID: loanTransaction.ID,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to retrieve loan transaction entries for transaction ID: %s", loanTransactionID.String())
	}
	totalDebitDec := decimal.Zero
	totalCreditDec := decimal.Zero
	for _, entry := range loanTransactionEntries {
		totalDebitDec = totalDebitDec.Add(decimal.NewFromFloat(entry.Debit))
		totalCreditDec = totalCreditDec.Add(decimal.NewFromFloat(entry.Credit))
	}
	currency := loanTransaction.Account.Currency
	accounts, err := core.AccountManager(service).Find(context, &types.Account{
		OrganizationID: loanTransaction.OrganizationID,
		BranchID:       loanTransaction.BranchID,
		LoanAccountID:  loanTransaction.AccountID,
		CurrencyID:     &currency.ID,
	}, "Currency")
	if err != nil {
		return nil, eris.Wrapf(err, "failed to retrieve accounts for loan transaction ID: %s", loanTransactionID.String())
	}
	holidays, err := core.HolidayManager(service).Find(context, &types.Holiday{
		OrganizationID: loanTransaction.OrganizationID,
		BranchID:       loanTransaction.BranchID,
		CurrencyID:     currency.ID,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to retrieve holidays for loan amortization schedule")
	}
	numberOfPayments, err := usecase.LoanNumberOfPayments(loanTransaction.ModeOfPayment, loanTransaction.Terms)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to calculate number of payments for loan with mode: %s and terms: %d",
			loanTransaction.ModeOfPayment, loanTransaction.Terms)
	}

	excludeSaturday := loanTransaction.ExcludeSaturday
	excludeSunday := loanTransaction.ExcludeSunday
	excludeHolidays := loanTransaction.ExcludeHoliday

	isMonthlyExactDay := loanTransaction.ModeOfPaymentMonthlyExactDay
	weeklyExactDay := loanTransaction.ModeOfPaymentWeekly
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

	startDate := userOrg.TimeMachine()
	if loanTransaction.PrintedDate != nil {
		startDate = loanTransaction.PrintedDate.UTC()
	}
	if loanTransaction.PrintedDate == nil {
		loanTransaction.PrintedDate = &startDate
	}
	totalDebit := totalDebitDec.InexactFloat64()
	totalCredit := totalCreditDec.InexactFloat64()
	principal := totalCredit
	balance := totalCredit
	total := 0.0

	principalDec := decimal.NewFromFloat(principal)
	balanceDec := decimal.NewFromFloat(balance)
	totalDec := decimal.NewFromFloat(total)

	for i := range numberOfPayments + 1 {
		actualDate := startDate
		daysSkipped := 0
		rowTotalDec := decimal.Zero

		daysSkipped, err := skippedDaysCount(startDate, currency, excludeSaturday, excludeSunday, excludeHolidays, holidays)
		if err != nil {
			return nil, eris.Wrapf(err, "failed to calculate skipped days for payment date: %s", startDate.Format("2006-01-02"))
		}

		scheduledDate := startDate.AddDate(0, 0, daysSkipped)
		periodAccounts := make([]*AccountValue, len(accountsSchedule))

		if i > 0 {
			for j := range accountsSchedule {
				accountHistory, err := core.GetAccountHistoryLatestByTime(
					context, service, accountsSchedule[j].Account.ID, loanTransaction.OrganizationID,
					loanTransaction.BranchID, loanTransaction.PrintedDate,
				)
				if err != nil {
					return nil, eris.Wrapf(err, "failed to get account history for account ID: %s", accountsSchedule[j].Account.ID.String())
				}

				periodAccounts[j] = &AccountValue{
					Account: accountHistory,
					Value:   0,
					Total:   accountsSchedule[j].Total,
				}

				switch accountHistory.Type {
				case types.AccountTypeLoan:
					valueDec := principalDec.Div(decimal.NewFromFloat(float64(numberOfPayments)))
					if valueDec.GreaterThan(balanceDec) {
						valueDec = balanceDec
					}
					periodAccounts[j].Value = valueDec.InexactFloat64()

					accountsSchedule[j].Total = decimal.NewFromFloat(accountsSchedule[j].Total).Add(valueDec).InexactFloat64()
					periodAccounts[j].Total = accountsSchedule[j].Total

					balanceDec = balanceDec.Sub(valueDec)

				case types.AccountTypeFines:
					if daysSkipped > 0 && !accountHistory.NoGracePeriodDaily {
						value := usecase.ComputeFines(
							principal,
							accountHistory.FinesAmort,
							accountHistory.FinesMaturity,
							daysSkipped,
							loanTransaction.ModeOfPayment,
							accountHistory.NoGracePeriodDaily,
							*accountHistory,
						)
						valueDec := decimal.NewFromFloat(value)

						accountsSchedule[j].Total = decimal.NewFromFloat(accountsSchedule[j].Total).Add(valueDec).InexactFloat64()
						periodAccounts[j].Value = valueDec.InexactFloat64()
						periodAccounts[j].Total = accountsSchedule[j].Total
					}

				default:
					switch accountHistory.ComputationType {
					case types.Straight, types.Diminishing, types.DiminishingStraight:
						if accountHistory.Type == types.AccountTypeInterest || accountHistory.Type == types.AccountTypeSVFLedger {
							interest := usecase.ComputeInterest(balanceDec.InexactFloat64(), accountHistory.InterestStandard, loanTransaction.ModeOfPayment)
							interestDec := decimal.NewFromFloat(interest)

							accountsSchedule[j].Total = decimal.NewFromFloat(accountsSchedule[j].Total).Add(interestDec).InexactFloat64()
							periodAccounts[j].Value = interestDec.InexactFloat64()
							periodAccounts[j].Total = accountsSchedule[j].Total
						}
					}
				}

				totalDec = totalDec.Add(decimal.NewFromFloat(periodAccounts[j].Value))
				rowTotalDec = rowTotalDec.Add(decimal.NewFromFloat(periodAccounts[j].Value))
			}
		} else {
			for j := range accountsSchedule {
				accountHistory, err := core.GetAccountHistoryLatestByTime(
					context, service, accountsSchedule[j].Account.ID, loanTransaction.OrganizationID,
					loanTransaction.BranchID, loanTransaction.PrintedDate,
				)
				if err != nil {
					return nil, eris.Wrapf(err, "failed to initialize account history for first payment")
				}
				periodAccounts[j] = &AccountValue{
					Account: accountHistory,
					Value:   0,
					Total:   0,
				}
			}
		}

		sort.Slice(periodAccounts, func(i, j int) bool {
			return getAccountTypePriority(periodAccounts[i].Account.Type) <
				getAccountTypePriority(periodAccounts[j].Account.Type)
		})

		switch loanTransaction.ModeOfPayment {
		case types.LoanModeOfPaymentDaily:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ActualDate:    actualDate,
				ScheduledDate: scheduledDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotalDec.InexactFloat64(),
				Accounts:      periodAccounts,
			})
			startDate = startDate.AddDate(0, 0, 1)

		case types.LoanModeOfPaymentWeekly:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotalDec.InexactFloat64(),
				Accounts:      periodAccounts,
			})
			weekDay := types.LoanWeeklyIota(weeklyExactDay)
			startDate = nextWeekday(startDate, time.Weekday(weekDay))

		case types.LoanModeOfPaymentSemiMonthly:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotalDec.InexactFloat64(),
				Accounts:      periodAccounts,
			})
			thisDay := startDate.Day()
			thisMonth := startDate.Month()
			thisYear := startDate.Year()
			loc := startDate.Location()

			switch {
			case thisDay < semiMonthlyExactDay1:
				startDate = time.Date(thisYear, thisMonth, semiMonthlyExactDay1, startDate.Hour(), startDate.Minute(), startDate.Second(), startDate.Nanosecond(), loc)
			case thisDay < semiMonthlyExactDay2:
				startDate = time.Date(thisYear, thisMonth, semiMonthlyExactDay2, startDate.Hour(), startDate.Minute(), startDate.Second(), startDate.Nanosecond(), loc)
			default:
				nextMonth := startDate.AddDate(0, 1, 0)
				startDate = time.Date(nextMonth.Year(), nextMonth.Month(), semiMonthlyExactDay1, startDate.Hour(), startDate.Minute(), startDate.Second(), startDate.Nanosecond(), loc)
			}

		case types.LoanModeOfPaymentMonthly:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotalDec.InexactFloat64(),
				Accounts:      periodAccounts,
			})
			loc := startDate.Location()
			day := startDate.Day()

			if isMonthlyExactDay {
				nextMonth := startDate.AddDate(0, 1, 0)
				startDate = time.Date(nextMonth.Year(), nextMonth.Month(), day, startDate.Hour(), startDate.Minute(), startDate.Second(), startDate.Nanosecond(), loc)
			} else {
				startDate = startDate.AddDate(0, 0, 30)
			}

		case types.LoanModeOfPaymentQuarterly:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotalDec.InexactFloat64(),
				Accounts:      periodAccounts,
			})
			startDate = startDate.AddDate(0, 3, 0)

		case types.LoanModeOfPaymentSemiAnnual:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotalDec.InexactFloat64(),
				Accounts:      periodAccounts,
			})
			startDate = startDate.AddDate(0, 6, 0)

		case types.LoanModeOfPaymentLumpsum:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotalDec.InexactFloat64(),
				Accounts:      periodAccounts,
			})

		case types.LoanModeOfPaymentFixedDays:
			amortization = append(amortization, &LoanAmortizationSchedule{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotalDec.InexactFloat64(),
				Accounts:      periodAccounts,
			})
			startDate = startDate.AddDate(0, 0, 1)
		}
	}

	return &LoanTransactionAmortizationResponse{
		Entries:     core.LoanTransactionEntryManager(service).ToModels(loanTransactionEntries),
		Currency:    *core.CurrencyManager(service).ToModel(currency),
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		Total:       total,
		Schedule:    amortization,
	}, nil
}
