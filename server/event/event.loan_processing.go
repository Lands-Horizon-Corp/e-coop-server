package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func (e *Event) LoanProcessing(context context.Context, ctx echo.Context, loanTransactionID *uuid.UUID) (*core.LoanTransaction, error) {
	tx, endTx := e.provider.Service.Database.StartTransaction(context)
	now := time.Now().UTC()
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "loan processing: failed to get loan transaction by id"))
	}

	// Get current user organization
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(context, ctx)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "loan processing: failed to get current user organization"))
	}
	if userOrg.BranchID == nil {
		return nil, endTx(eris.New("loan processing: user organization has no branch assigned"))
	}

	memberProfile, err := e.core.MemberProfileManager.GetByID(context, *loanTransaction.MemberProfileID)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "member-profile-retrieval-failed",
			Description: "Unable to retrieve member profile " + loanTransaction.MemberProfileID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve member profile"))
	}
	if memberProfile == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "member-profile-not-found",
			Description: "Member profile does not exist for ID: " + loanTransaction.MemberProfileID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("member profile not found"))
	}

	// ===============================
	// STEP 5: FETCH RELATED ACCOUNTS & CURRENCY
	// ===============================
	currency := loanTransaction.Account.Currency
	accounts, err := e.core.AccountManager.Find(context, &core.Account{
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
		LoanAccountID:  loanTransaction.AccountID,
		CurrencyID:     &currency.ID,
	}, "Currency")
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "data-retrieval-failed",
			Description: "Failed to retrieve loan-related accounts for amortization schedule: " + err.Error(),
			Module:      "Loan Amortization",
		})
		return nil, endTx(eris.Wrapf(err, "failed to retrieve accounts for loan transaction ID: %s", loanTransactionID.String()))
	}

	// ===============================
	// STEP 6: FETCH HOLIDAY CALENDAR
	// ===============================
	holidays, err := e.core.HolidayManager.Find(context, &core.Holiday{
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
		CurrencyID:     currency.ID,
	})
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "data-retrieval-failed",
			Description: "Failed to retrieve holiday calendar for payment schedule calculations: " + err.Error(),
			Module:      "Loan Amortization",
		})
		return nil, endTx(eris.Wrapf(err, "failed to retrieve holidays for loan amortization schedule"))
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
		return nil, endTx(eris.Wrapf(err, "failed to calculate number of payments for loan with mode: %s and terms: %d",
			loanTransaction.ModeOfPayment, loanTransaction.Terms))
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

	if loanTransaction.PrintedDate == nil {
		return nil, endTx(eris.New("loan processing: printed date is nil"))
	}
	// Initialize payment calculation variables
	currentDate := time.Now().UTC()
	if userOrg.TimeMachineTime != nil {
		currentDate = userOrg.UserOrgTime()
	}
	paymentDate := *loanTransaction.PrintedDate
	balance := loanTransaction.TotalPrincipal
	principal := loanTransaction.TotalPrincipal
	for i := range numberOfPayments + 1 {
		daysSkipped := 0
		daysSkipped, err := e.skippedDaysCount(paymentDate, currency, excludeSaturday, excludeSunday, excludeHolidays, holidays)
		if err != nil {
			e.Footstep(ctx, FootstepEvent{
				Activity:    "calculation-failed",
				Description: "Failed to calculate skipped days for payment schedule: " + err.Error(),
				Module:      "Loan Amortization",
			})
			return nil, endTx(eris.Wrapf(err, "failed to calculate skipped days for payment date: %s", paymentDate.Format("2006-01-02")))
		}

		scheduledDate := paymentDate.AddDate(0, 0, daysSkipped)

		currentBalance := e.provider.Service.Decimal.Clamp(
			e.provider.Service.Decimal.Divide(principal, float64(numberOfPayments)), 0, balance)

		balance = e.provider.Service.Decimal.Subtract(balance, currentBalance)

		// ===============================
		// STEP 11: CREATE PERIOD-SPECIFIC ACCOUNT CALCULATIONS
		// ===============================
		if loanTransaction.LoanCount >= i && scheduledDate.Before(currentDate) {
			for _, account := range accounts {
				if account.LoanAccountID == nil || account.ComputationType == core.Straight || account.Type == core.AccountTypeFines {
					continue
				}
				var price float64
				switch account.ComputationType {
				case core.Diminishing:
					price = e.usecase.ComputeInterest(balance, account.InterestStandard, loanTransaction.ModeOfPayment)
				case core.DiminishingStraight:
					price = e.usecase.ComputeInterest(principal, account.InterestStandard, loanTransaction.ModeOfPayment)
				}
				if price <= 0 {
					continue
				}
				memberAccountLedger, err := e.core.GeneralLedgerCurrentMemberAccountForUpdate(
					context, tx,
					memberProfile.ID,
					account.ID,
					memberProfile.OrganizationID,
					memberProfile.BranchID,
				)
				if err != nil {
					return nil, endTx(eris.New("failed to fetch current member general member thats ready for update"))
				}
				var currentMemberBalance float64 = 0
				if memberAccountLedger == nil {
					e.Footstep(ctx, FootstepEvent{
						Activity:    "member-ledger-initialization",
						Description: "Initializing new member ledger for account " + loanTransaction.AccountID.String() + " and member " + memberProfile.ID.String(),
						Module:      "Loan Release",
					})
				} else {
					currentMemberBalance = memberAccountLedger.Balance
				}
				memberDebit, memberCredit, newMemberBalance := e.usecase.Adjustment(
					*loanTransaction.Account, 0.0, price, currentMemberBalance)
				memberLedgerEntry := &core.GeneralLedger{
					CreatedAt:                  now,
					CreatedByID:                userOrg.UserID,
					UpdatedAt:                  now,
					UpdatedByID:                userOrg.UserID,
					BranchID:                   *userOrg.BranchID,
					OrganizationID:             userOrg.OrganizationID,
					ReferenceNumber:            loanTransaction.CheckNumber,
					EntryDate:                  &currentDate,
					AccountID:                  loanTransaction.AccountID,
					MemberProfileID:            &memberProfile.ID,
					PaymentTypeID:              account.DefaultPaymentTypeID,
					TransactionReferenceNumber: loanTransaction.CheckNumber,
					Source:                     core.GeneralLedgerSourceCheckVoucher,
					EmployeeUserID:             &userOrg.UserID,
					Description:                loanTransaction.Account.Description,
					Credit:                     memberCredit,
					Debit:                      memberDebit,
					Balance:                    newMemberBalance,
					CurrencyID:                 &currency.ID,
					LoanTransactionID:          &loanTransaction.ID,
				}
				if err := e.core.GeneralLedgerManager.CreateWithTx(context, tx, memberLedgerEntry); err != nil {
					e.Footstep(ctx, FootstepEvent{
						Activity:    "cash-ledger-creation-failed",
						Description: "Unable to create cash account ledger entry for :" + account.ID.String(),
						Module:      "Loan Release",
					})
					return nil, endTx(eris.Wrap(err, "failed to create general ledger entry"))
				}
				_, err = e.core.MemberAccountingLedgerUpdateOrCreate(
					context,
					tx,
					*loanTransaction.MemberProfileID,
					account.ID,
					userOrg.OrganizationID,
					*userOrg.BranchID,
					userOrg.UserID,
					newMemberBalance,
					now,
				)
				if err != nil {
					e.Footstep(ctx, FootstepEvent{
						Activity:    "interest-accounting-ledger-failed",
						Description: "Failed to update member accounting ledger for interest account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
						Module:      "Loan Release",
					})
					return nil, endTx(eris.Wrap(err, "failed to update interest accounting ledger"))
				}
			}
			loanTransaction.LoanCount = i + 1
			if err := e.core.LoanTransactionManager.UpdateByIDWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
				e.Footstep(ctx, FootstepEvent{
					Activity:    "update-failed",
					Description: "Failed to update loan count during processing: " + err.Error(),
					Module:      "Loan Processing",
				})
				return nil, endTx(eris.Wrapf(err, "failed to update loan count for loan transaction ID: %s", loanTransaction.ID.String()))
			}
		}

		// ===============================
		// STEP 14: DETERMINE NEXT PAYMENT DATE
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
	// ================================================================================
	// STEP 10: DATABASE TRANSACTION COMMIT
	// ================================================================================
	// Commit all changes to the database
	if err := endTx(nil); err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "database-commit-failed",
			Description: "Unable to commit loan release transaction to database: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to commit transaction"))
	}

	// ================================================================================
	// STEP 11: FINAL TRANSACTION RETRIEVAL AND RETURN
	// ================================================================================
	// Retrieve and return the updated loan transaction
	updatedLoanTransaction, err := e.core.LoanTransactionManager.GetByID(context, loanTransaction.ID)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "final-retrieval-failed",
			Description: "Unable to retrieve updated loan transaction after successful release: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get updated loan transaction"))
	}
	return updatedLoanTransaction, nil
}
