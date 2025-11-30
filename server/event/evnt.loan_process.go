package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
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

func (e *Event) LoanProcessing(
	context context.Context,
	userOrg *core.UserOrganization,
	loanTransactionID *uuid.UUID,
) (*core.LoanTransaction, error) {
	// ===============================
	// STEP 1: INITIALIZE TRANSACTION AND BASIC VALIDATION
	// ===============================
	tx, endTx := e.provider.Service.Database.StartTransaction(context)
	loanTransaction, err := e.core.LoanTransactionManager.GetByIDIncludingDeleted(context, *loanTransactionID)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "loan processing: failed to get loan transaction by id"))
	}

	if loanTransaction.Processing {
		return nil, endTx(eris.New("This loan transaction is still being processed"))
	}

	// ===============================
	// STEP 2: VALIDATE USER ORGANIZATION AND BRANCH
	// ===============================
	if userOrg.BranchID == nil {
		return nil, endTx(eris.New("loan processing: user organization has no branch assigned"))
	}

	// ===============================
	// STEP 3: RETRIEVE AND VALIDATE MEMBER PROFILE
	// ===============================
	memberProfile, err := e.core.MemberProfileManager.GetByIDIncludingDeleted(context, *loanTransaction.MemberProfileID)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to retrieve member profile"))
	}
	if memberProfile == nil {
		return nil, endTx(eris.New("member profile not found"))
	}

	// ===============================
	// STEP 4: FETCH RELATED ACCOUNTS & CURRENCY
	// ===============================
	currency := loanTransaction.Account.Currency
	loanAccounts, err := e.core.LoanAccountManager.Find(context, &core.LoanAccount{
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

	// ===============================
	// STEP 5: FETCH HOLIDAY CALENDAR
	// ===============================
	holidays, err := e.core.HolidayManager.Find(context, &core.Holiday{
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
		CurrencyID:     currency.ID,
	})
	if err != nil {
		return nil, endTx(eris.Wrapf(err, "failed to retrieve holidays for loan processing schedule"))
	}

	// ===============================
	// STEP 6: CALCULATE NUMBER OF PAYMENTS
	// ===============================
	numberOfPayments, err := e.usecase.LoanNumberOfPayments(loanTransaction.ModeOfPayment, loanTransaction.Terms)
	if err != nil {
		return nil, endTx(eris.Wrapf(err, "failed to calculate number of payments for loan with mode: %s and terms: %d",
			loanTransaction.ModeOfPayment, loanTransaction.Terms))
	}

	// ===============================
	// STEP 7: CONFIGURE PAYMENT SCHEDULE SETTINGS
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

	if loanTransaction.PrintedDate == nil {
		return nil, endTx(eris.New("loan processing: printed date is nil"))
	}

	// ===============================
	// STEP 8: INITIALIZE PAYMENT CALCULATION VARIABLES
	// ===============================
	currentDate := userOrg.UserOrgTime()
	paymentDate := *loanTransaction.PrintedDate
	balance := loanTransaction.TotalPrincipal
	principal := loanTransaction.TotalPrincipal

	// ===============================
	// STEP 9: PROCESS PAYMENT SCHEDULE ITERATIONS
	// ===============================
	for i := range numberOfPayments + 1 {
		daysSkipped := 0
		daysSkipped, err := e.skippedDaysCount(paymentDate, currency, excludeSaturday, excludeSunday, excludeHolidays, holidays)
		if err != nil {
			return nil, endTx(eris.Wrapf(err, "failed to calculate skipped days for payment date: %s", paymentDate.Format("2006-01-02")))
		}

		if i > 0 {
			// Adjust payment date and calculate balance
			scheduledDate := paymentDate.AddDate(0, 0, daysSkipped)
			currentBalance := e.provider.Service.Decimal.Clamp(
				e.provider.Service.Decimal.Divide(principal, float64(numberOfPayments)), 0, balance)

			if i >= loanTransaction.Count && scheduledDate.Before(currentDate) {
				for _, account := range loanAccounts {
					accountHistory, err := e.core.AccountHistoryManager.GetByID(context, *account.AccountHistoryID)
					if err != nil {
						return nil, endTx(eris.Wrapf(err, "failed to get account history for loan account id: %s", account.ID))
					}
					if accountHistory == nil {
						return nil, endTx(eris.New("account history not found"))
					}
					// Calculate the amount to add based on account type
					var amountToAdd float64
					switch accountHistory.Type {
					case core.AccountTypeLoan:
						// LOAN PRINCIPAL: Skip for principal accounts in real-time processing
						continue

					case core.AccountTypeFines:
						// FINES CALCULATION: Based on days skipped and penalty rates
						if daysSkipped > 0 && !accountHistory.NoGracePeriodDaily {
							account := e.core.AccountHistoryToModel(accountHistory)
							amountToAdd = e.usecase.ComputeFines(
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
						// INTEREST CALCULATIONS
						switch accountHistory.ComputationType {
						case core.Straight:
							if accountHistory.Type == core.AccountTypeInterest || accountHistory.Type == core.AccountTypeSVFLedger {
								// STRAIGHT INTEREST: Fixed percentage of original principal
								amountToAdd = e.usecase.ComputeInterest(principal, accountHistory.InterestStandard, loanTransaction.ModeOfPayment)
							}
						case core.Diminishing:
							if accountHistory.Type == core.AccountTypeInterest || accountHistory.Type == core.AccountTypeSVFLedger {
								// DIMINISHING INTEREST: Percentage of remaining balance
								amountToAdd = e.usecase.ComputeInterest(balance, accountHistory.InterestStandard, loanTransaction.ModeOfPayment)
							}
						case core.DiminishingStraight:
							if accountHistory.Type == core.AccountTypeInterest || accountHistory.Type == core.AccountTypeSVFLedger {
								// DIMINISHING STRAIGHT: Hybrid calculation method
								amountToAdd = e.usecase.ComputeInterest(balance, accountHistory.InterestStandard, loanTransaction.ModeOfPayment)
							}
						}
					}

					// Update the loan account with the calculated amount
					if amountToAdd > 0 {
						account.TotalAddCount += 1
						account.TotalAdd = e.provider.Service.Decimal.Add(account.TotalAdd, amountToAdd)
						account.Amount = e.provider.Service.Decimal.Add(account.Amount, amountToAdd)
						account.UpdatedByID = userOrg.UserID
						account.UpdatedAt = currentDate

						// Save the updated loan account
						if err := e.core.LoanAccountManager.UpdateByIDWithTx(context, tx, account.ID, account); err != nil {
							return nil, endTx(eris.Wrapf(err, "failed to update loan account ID: %s", account.ID.String()))
						}
					}
				}
				// Update loan count AFTER successful processing
				loanTransaction.Count = i + 1
				if err := e.core.LoanTransactionManager.UpdateByIDWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
					return nil, endTx(eris.Wrapf(err, "failed to update loan count for loan transaction ID: %s", loanTransaction.ID.String()))
				}

			}
			balance = e.provider.Service.Decimal.Subtract(balance, currentBalance)
		}

		// ===============================
		// STEP 11: DETERMINE NEXT PAYMENT DATE
		// ===============================
		switch loanTransaction.ModeOfPayment {
		case core.LoanModeOfPaymentDaily:
			paymentDate = paymentDate.AddDate(0, 0, 1)
		case core.LoanModeOfPaymentWeekly:
			weekDay := e.core.LoanWeeklyIota(weeklyExactDay)
			paymentDate = e.nextWeekday(paymentDate, time.Weekday(weekDay))
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

	// ===============================
	// STEP 12: DATABASE TRANSACTION COMMIT
	// ===============================
	if err := endTx(nil); err != nil {
		return nil, endTx(eris.Wrap(err, "failed to commit transaction"))
	}

	// ===============================
	// STEP 13: FINAL TRANSACTION RETRIEVAL AND RETURN
	// ===============================
	updatedLoanTransaction, err := e.core.LoanTransactionManager.GetByID(context, loanTransaction.ID)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to get updated loan transaction"))
	}
	return updatedLoanTransaction, nil

}

func (e *Event) ProcessAllLoans(processContext context.Context, userOrg *core.UserOrganization) error {
	if userOrg == nil {
		return eris.New("user organization is nil")
	}

	if userOrg.BranchID == nil {
		return eris.New("user organization has no branch assigned")
	}

	currentTime := time.Now().UTC()

	// Get all loan transactions that are not currently being processed
	loanTransactions, err := e.core.LoanTransactionManager.FindIncludingDeleted(processContext, &core.LoanTransaction{
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

	// Mark all transactions as processing to prevent concurrent processing
	for _, entry := range loanTransactions {
		entry.Processing = true
		if err := e.core.LoanTransactionManager.UpdateByID(processContext, entry.ID, entry); err != nil {
			return eris.Wrap(err, "failed to mark loan transaction as processing")
		}
	}

	// Process asynchronously to avoid blocking the main thread
	go func() {
		timeoutContext, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel()

		// Process each loan transaction
		for i, entry := range loanTransactions {
			// Add small delay to prevent overwhelming the system
			time.Sleep(500 * time.Millisecond)

			// ============================================================
			// Process Loan Logic Here
			// ============================================================
			processedLoan, err := e.LoanProcessing(timeoutContext, userOrg, &entry.ID)
			if err != nil {
				e.provider.Service.Logger.Error("failed to process loan transaction",
					zap.Error(err),
					zap.String("loanTransactionID", entry.ID.String()),
					zap.String("organizationID", userOrg.OrganizationID.String()),
					zap.String("branchID", userOrg.BranchID.String()),
					zap.Int("iteration", i+1),
					zap.Int("total", len(loanTransactions)))

				// Mark this transaction as not processing so it can be retried
				entry.Processing = false
				if updateErr := e.core.LoanTransactionManager.UpdateByID(timeoutContext, entry.ID, entry); updateErr != nil {
					e.provider.Service.Logger.Error("failed to unmark processing flag after error",
						zap.Error(updateErr),
						zap.String("loanTransactionID", entry.ID.String()))
				}
				continue
			}

			// Update entry with processed data
			if processedLoan != nil {
				entry = processedLoan
			}

			// Dispatch progress event
			if err := e.provider.Service.Broker.Dispatch([]string{
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
				e.provider.Service.Logger.Error("failed to dispatch loan process event",
					zap.Error(err),
					zap.String("loanTransactionID", entry.ID.String()))
				// Don't return here, continue processing other loans
			}

			// Check for timeout
			select {
			case <-timeoutContext.Done():
				e.provider.Service.Logger.Warn("loan processing timed out",
					zap.Int("processed", i+1),
					zap.Int("total", len(loanTransactions)))
				return
			default:
			}
		}

		// Mark all transactions as not processing after completion
		for _, entry := range loanTransactions {
			entry.Processing = false
			if err := e.core.LoanTransactionManager.UpdateByID(timeoutContext, entry.ID, entry); err != nil {
				e.provider.Service.Logger.Error("failed to unmark loan transaction as processing",
					zap.Error(err),
					zap.String("loanTransactionID", entry.ID.String()))
			}
		}

		// Send completion event
		if err := e.provider.Service.Broker.Dispatch([]string{
			fmt.Sprintf("loan.process.completed.branch.%s", userOrg.BranchID),
			fmt.Sprintf("loan.process.completed.organization.%s", userOrg.OrganizationID),
		}, map[string]interface{}{
			"total_processed": len(loanTransactions),
			"start_time":      currentTime,
			"end_time":        time.Now().UTC(),
			"organization_id": userOrg.OrganizationID,
			"branch_id":       userOrg.BranchID,
		}); err != nil {
			e.provider.Service.Logger.Error("failed to dispatch completion event",
				zap.Error(err))
		}

		e.provider.Service.Logger.Info("loan processing completed",
			zap.Int("total_processed", len(loanTransactions)),
			zap.String("organization_id", userOrg.OrganizationID.String()),
			zap.String("branch_id", userOrg.BranchID.String()))
	}()

	return nil
}
