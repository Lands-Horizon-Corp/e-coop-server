package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (e *Event) LoanProcessing(
	context context.Context,
	userOrg *core.UserOrganization,
	loanTransactionID *uuid.UUID,
) (*core.LoanTransaction, error) {
	// ===============================
	// STEP 1: INITIALIZE TRANSACTION AND BASIC VALIDATION
	// ===============================
	tx, endTx := e.provider.Service.Database.StartTransaction(context)
	now := time.Now().UTC()
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
	accounts, err := e.core.AccountManager.Find(context, &core.Account{
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
		LoanAccountID:  loanTransaction.AccountID,
		CurrencyID:     &currency.ID,
	}, "Currency")
	if err != nil {
		return nil, endTx(eris.Wrapf(err, "failed to retrieve accounts for loan transaction ID: %s", loanTransactionID.String()))
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
		// Calculate skipped days due to weekends/holidays
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

			// ===============================
			// STEP 10: CREATE PERIOD-SPECIFIC ACCOUNT CALCULATIONS
			if i >= loanTransaction.Count && scheduledDate.Before(currentDate) {
				for _, account := range accounts {
					if loanTransaction.AccountID == nil || account.ComputationType == core.Straight || account.Type == core.AccountTypeLoan {
						continue
					}
					accountHistory, err := e.core.GetAccountHistoryLatestByTimeHistory(
						context,
						account.ID,
						account.OrganizationID,
						account.BranchID,
						loanTransaction.PrintedDate,
					)
					if err != nil {
						return nil, endTx(eris.Wrap(err, "failed to retrieve account history"))
					}
					if accountHistory != nil {
						account = e.core.AccountHistoryToModel(accountHistory)
					}

					var price float64 = 0.0
					switch account.Type {
					case core.AccountTypeFines:
						// if !e.provider.Service.Decimal.IsLessThan(balance, currentMemberBalance) {
						// 	continue
						// }
					default:
						switch account.ComputationType {
						case core.Diminishing:
							if account.Type == core.AccountTypeInterest || account.Type == core.AccountTypeSVFLedger {
								price = e.usecase.ComputeInterest(balance, account.InterestStandard, loanTransaction.ModeOfPayment)
							}
						case core.DiminishingStraight:
							if account.Type == core.AccountTypeInterest || account.Type == core.AccountTypeSVFLedger {
								price = e.usecase.ComputeInterest(principal, account.InterestStandard, loanTransaction.ModeOfPayment)
							}
						}
					}
					if price <= 0 {
						continue
					}
					memberDebit := 0.0
					memberCredit := price
					memberLedgerEntry := &core.GeneralLedger{
						CreatedAt:                  now,
						CreatedByID:                userOrg.UserID,
						UpdatedAt:                  now,
						UpdatedByID:                userOrg.UserID,
						BranchID:                   *userOrg.BranchID,
						OrganizationID:             userOrg.OrganizationID,
						ReferenceNumber:            loanTransaction.Voucher,
						EntryDate:                  scheduledDate,
						AccountID:                  &account.ID,
						MemberProfileID:            &memberProfile.ID,
						PaymentTypeID:              account.DefaultPaymentTypeID,
						TransactionReferenceNumber: loanTransaction.Voucher,
						Source:                     core.GeneralLedgerSourceCheckVoucher,
						EmployeeUserID:             &userOrg.UserID,
						Description:                account.Description,
						Credit:                     memberCredit,
						Debit:                      memberDebit,
						CurrencyID:                 &currency.ID,
						LoanTransactionID:          &loanTransaction.ID,
					}
					if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, memberLedgerEntry); err != nil {
						return nil, endTx(eris.Wrap(err, "failed to create general ledger entry"))
					}

					_, err = e.core.MemberAccountingLedgerUpdateOrCreate(
						context,
						tx,
						core.MemberAccountingLedgerUpdateOrCreateParams{
							MemberProfileID: *loanTransaction.MemberProfileID,
							AccountID:       account.ID,
							OrganizationID:  userOrg.OrganizationID,
							BranchID:        *userOrg.BranchID,
							UserID:          userOrg.UserID,
							DebitAmount:     memberDebit,
							CreditAmount:    memberCredit,
							LastPayTime:     now,
						},
					)
					if err != nil {
						return nil, endTx(eris.Wrap(err, "failed to update accounting ledger"))
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
