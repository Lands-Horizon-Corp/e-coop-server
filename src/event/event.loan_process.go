package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/usecase"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type LoanProcessEventResponse struct {
	Total       int       `json:"total"`
	Processed   int       `json:"processed"`
	StartTime   time.Time `json:"start_time"`
	CurrentTime time.Time `json:"current_time"`
	AccountName string    `json:"account_name"`
	MemberName  string    `json:"member_name"`
}

func LoanProcessing(
	context context.Context, service *horizon.HorizonService,
	userOrg *core.UserOrganization,
	loanTransactionID *uuid.UUID,
) (*core.LoanTransaction, error) {
	tx, endTx := service.Database.StartTransaction(context)
	loanTransaction, err := core.LoanTransactionManager(service).GetByIDIncludingDeleted(context, *loanTransactionID)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "loan processing: failed to get loan transaction by id"))
	}

	if loanTransaction.Processing {
		return nil, endTx(eris.New("This loan transaction is still being processed"))
	}

	if userOrg.BranchID == nil {
		return nil, endTx(eris.New("loan processing: user organization has no branch assigned"))
	}

	memberProfile, err := core.MemberProfileManager(service).GetByIDIncludingDeleted(context, *loanTransaction.MemberProfileID)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to retrieve member profile"))
	}
	if memberProfile == nil {
		return nil, endTx(eris.New("member profile not found"))
	}

	currency := loanTransaction.Account.Currency
	loanAccounts, err := core.LoanAccountManager(service).Find(context, &core.LoanAccount{
		OrganizationID:    userOrg.OrganizationID,
		BranchID:          *userOrg.BranchID,
		LoanTransactionID: loanTransaction.ID,
	})
	if err != nil {
		return nil, endTx(eris.Wrapf(err, "failed to retrieve loan accounts for loan transaction id: %s", loanTransaction.ID))
	}
	if len(loanAccounts) == 0 {
		return nil, endTx(eris.New("no loan accounts found for the specified loan transaction"))
	}

	holidays, err := core.HolidayManager(service).Find(context, &core.Holiday{
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
		CurrencyID:     currency.ID,
	})
	if err != nil {
		return nil, endTx(eris.Wrapf(err, "failed to retrieve holidays for loan processing schedule"))
	}

	numberOfPayments, err := usecase.LoanNumberOfPayments(loanTransaction.ModeOfPayment, loanTransaction.Terms)
	if err != nil {
		return nil, endTx(eris.Wrapf(err, "failed to calculate number of payments for loan with mode: %s and terms: %d",
			loanTransaction.ModeOfPayment, loanTransaction.Terms))
	}

	excludeSaturday := loanTransaction.ExcludeSaturday
	excludeSunday := loanTransaction.ExcludeSunday
	excludeHolidays := loanTransaction.ExcludeHoliday

	isMonthlyExactDay := loanTransaction.ModeOfPaymentMonthlyExactDay
	weeklyExactDay := loanTransaction.ModeOfPaymentWeekly
	semiMonthlyExactDay1 := loanTransaction.ModeOfPaymentSemiMonthlyPay1
	semiMonthlyExactDay2 := loanTransaction.ModeOfPaymentSemiMonthlyPay2

	if loanTransaction.PrintedDate == nil {
		return nil, endTx(eris.New("loan processing: printed date is nil"))
	}

	currentDate := userOrg.UserOrgTime()
	paymentDate := *loanTransaction.PrintedDate
	balance := loanTransaction.TotalPrincipal
	principal := loanTransaction.TotalPrincipal

	for i := range numberOfPayments + 1 {
		daysSkipped := 0
		daysSkipped, err := skippedDaysCount(paymentDate, currency, excludeSaturday, excludeSunday, excludeHolidays, holidays)
		if err != nil {
			return nil, endTx(eris.Wrapf(err, "failed to calculate skipped days for payment date: %s", paymentDate.Format("2006-01-02")))
		}

		if i > 0 {
			scheduledDate := paymentDate.AddDate(0, 0, daysSkipped)

			principalDec := decimal.NewFromFloat(principal)
			balanceDec := decimal.NewFromFloat(balance)
			numPaymentsDec := decimal.NewFromFloat(float64(numberOfPayments))

			currentBalanceDec := principalDec.Div(numPaymentsDec)
			if currentBalanceDec.LessThan(decimal.Zero) {
				currentBalanceDec = decimal.Zero
			}
			if currentBalanceDec.GreaterThan(balanceDec) {
				currentBalanceDec = balanceDec
			}

			if i >= loanTransaction.Count && scheduledDate.Before(currentDate) {
				for _, account := range loanAccounts {
					accountHistory, err := core.AccountHistoryManager(service).GetByID(context, *account.AccountHistoryID)
					if err != nil {
						return nil, endTx(eris.Wrapf(err, "failed to get account history for loan account id: %s", account.ID))
					}
					if accountHistory == nil {
						return nil, endTx(eris.New("account history not found"))
					}
					var amountToAdd float64
					switch accountHistory.Type {
					case core.AccountTypeLoan:
						continue

					case core.AccountTypeFines:
						if daysSkipped > 0 && !accountHistory.NoGracePeriodDaily {
							account := core.AccountHistoryToModel(accountHistory)
							amountToAdd = usecase.ComputeFines(
								principal,
								accountHistory.FinesAmort,
								accountHistory.FinesMaturity,
								daysSkipped,
								loanTransaction.ModeOfPayment,
								accountHistory.NoGracePeriodDaily,
								*account,
							)
						}

					default:
						switch accountHistory.ComputationType {
						case core.Straight:
							if accountHistory.Type == core.AccountTypeInterest || accountHistory.Type == core.AccountTypeSVFLedger {
								amountToAdd = usecase.ComputeInterest(principal, accountHistory.InterestStandard, loanTransaction.ModeOfPayment)
							}
						case core.Diminishing:
							if accountHistory.Type == core.AccountTypeInterest || accountHistory.Type == core.AccountTypeSVFLedger {
								amountToAdd = usecase.ComputeInterest(balance, accountHistory.InterestStandard, loanTransaction.ModeOfPayment)
							}
						case core.DiminishingStraight:
							if accountHistory.Type == core.AccountTypeInterest || accountHistory.Type == core.AccountTypeSVFLedger {
								amountToAdd = usecase.ComputeInterest(balance, accountHistory.InterestStandard, loanTransaction.ModeOfPayment)
							}
						}
					}

					if amountToAdd > 0 {
						currentBalanceDec := principalDec.Div(numPaymentsDec)
						if currentBalanceDec.LessThan(decimal.Zero) {
							currentBalanceDec = decimal.Zero
						}
						if currentBalanceDec.GreaterThan(balanceDec) {
							currentBalanceDec = balanceDec
						}

						if err := core.LoanAccountManager(service).UpdateByIDWithTx(context, tx, account.ID, account); err != nil {
							return nil, endTx(eris.Wrapf(err, "failed to update loan account ID: %s", account.ID.String()))
						}
					}
				}
				loanTransaction.Count = i + 1
				if err := core.LoanTransactionManager(service).UpdateByIDWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
					return nil, endTx(eris.Wrapf(err, "failed to update loan count for loan transaction ID: %s", loanTransaction.ID.String()))
				}

			}
			balanceDec = balanceDec.Sub(currentBalanceDec)
			balance = balanceDec.InexactFloat64()
		}

		switch loanTransaction.ModeOfPayment {
		case core.LoanModeOfPaymentDaily:
			paymentDate = paymentDate.AddDate(0, 0, 1)
		case core.LoanModeOfPaymentWeekly:
			weekDay := core.LoanWeeklyIota(weeklyExactDay)
			paymentDate = nextWeekday(paymentDate, time.Weekday(weekDay))
		case core.LoanModeOfPaymentSemiMonthly:
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
			loc := paymentDate.Location()
			day := paymentDate.Day()
			if isMonthlyExactDay {
				nextMonth := paymentDate.AddDate(0, 1, 0)
				paymentDate = time.Date(nextMonth.Year(), nextMonth.Month(), day, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			} else {
				paymentDate = paymentDate.AddDate(0, 0, 30)
			}
		case core.LoanModeOfPaymentQuarterly:
			paymentDate = paymentDate.AddDate(0, 3, 0)
		case core.LoanModeOfPaymentSemiAnnual:
			paymentDate = paymentDate.AddDate(0, 6, 0)
		case core.LoanModeOfPaymentLumpsum:
		case core.LoanModeOfPaymentFixedDays:
			paymentDate = paymentDate.AddDate(0, 0, 1)
		}
	}

	if err := endTx(nil); err != nil {
		return nil, endTx(eris.Wrap(err, "failed to commit transaction"))
	}

	updatedLoanTransaction, err := core.LoanTransactionManager(service).GetByID(context, loanTransaction.ID)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to get updated loan transaction"))
	}
	return updatedLoanTransaction, nil

}

func ProcessAllLoans(processContext context.Context, service *horizon.HorizonService, userOrg *core.UserOrganization) error {
	if userOrg == nil {
		return eris.New("user organization is nil")
	}
	if userOrg.BranchID == nil {
		return eris.New("user organization has no branch assigned")
	}
	currentTime := time.Now().UTC()
	loanTransactions, err := core.LoanTransactionManager(service).FindIncludeDeleted(processContext, &core.LoanTransaction{
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
		Processing:     false,
	})
	if err != nil {
		return eris.Wrap(err, "failed to get loan transactions for processing")
	}

	if len(loanTransactions) == 0 {
		return eris.New("no loan transactions found to process")
	}

	for _, entry := range loanTransactions {
		entry.Processing = true
		if err := core.LoanTransactionManager(service).UpdateByID(processContext, entry.ID, entry); err != nil {
			return eris.Wrap(err, "failed to mark loan transaction as processing")
		}
	}

	go func() {
		timeoutContext, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel()

		for i, entry := range loanTransactions {
			time.Sleep(500 * time.Millisecond)

			processedLoan, err := LoanProcessing(timeoutContext, service, userOrg, &entry.ID)
			if err != nil {
				service.Logger.Error("failed to process loan transaction",
					zap.Error(err),
					zap.String("loanTransactionID", entry.ID.String()),
					zap.String("organizationID", userOrg.OrganizationID.String()),
					zap.String("branchID", userOrg.BranchID.String()),
					zap.Int("iteration", i+1),
					zap.Int("total", len(loanTransactions)))

				entry.Processing = false
				if updateErr := core.LoanTransactionManager(service).UpdateByID(timeoutContext, entry.ID, entry); updateErr != nil {
					service.Logger.Error("failed to unmark processing flag after error",
						zap.Error(updateErr),
						zap.String("loanTransactionID", entry.ID.String()))
				}
				continue
			}

			if processedLoan != nil {
				entry = processedLoan
			}

			if err := service.Broker.Dispatch([]string{
				fmt.Sprintf("loan.process.branch.%s", userOrg.BranchID),
				fmt.Sprintf("loan.process.organization.%s", userOrg.OrganizationID),
			}, LoanProcessEventResponse{
				Total:       len(loanTransactions),
				Processed:   i + 1,
				StartTime:   currentTime,
				CurrentTime: time.Now().UTC(),
				AccountName: func() string {
					if entry.Account != nil {
						return entry.Account.Name
					}
					return ""
				}(),
				MemberName: func() string {
					if entry.MemberProfile != nil {
						return entry.MemberProfile.FullName
					}
					return ""
				}(),
			}); err != nil {
				service.Logger.Error("failed to dispatch loan process event",
					zap.Error(err),
					zap.String("loanTransactionID", entry.ID.String()))
			}

			select {
			case <-timeoutContext.Done():
				service.Logger.Warn("loan processing timed out",
					zap.Int("processed", i+1),
					zap.Int("total", len(loanTransactions)))
				return
			default:
			}
		}

		for _, entry := range loanTransactions {
			entry.Processing = false
			if err := core.LoanTransactionManager(service).UpdateByID(timeoutContext, entry.ID, entry); err != nil {
				service.Logger.Error("failed to unmark loan transaction as processing",
					zap.Error(err),
					zap.String("loanTransactionID", entry.ID.String()))
			}
		}

		if err := service.Broker.Dispatch([]string{
			fmt.Sprintf("loan.process.completed.branch.%s", userOrg.BranchID),
			fmt.Sprintf("loan.process.completed.organization.%s", userOrg.OrganizationID),
		}, map[string]any{
			"total_processed": len(loanTransactions),
			"start_time":      currentTime,
			"end_time":        time.Now().UTC(),
			"organization_id": userOrg.OrganizationID,
			"branch_id":       userOrg.BranchID,
		}); err != nil {
			service.Logger.Error("failed to dispatch completion event",
				zap.Error(err))
		}

		service.Logger.Info("loan processing completed",
			zap.Int("total_processed", len(loanTransactions)),
			zap.String("organization_id", userOrg.OrganizationID.String()),
			zap.String("branch_id", userOrg.BranchID.String()))
	}()

	return nil
}
