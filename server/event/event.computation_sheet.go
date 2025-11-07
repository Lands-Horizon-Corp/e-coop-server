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
	ModeOfPaymentMonthlyExactDay bool          `json:"mode_of_payment_monthly_exact_day,omitempty"`
	ModeOfPaymentWeekly          core.Weekdays `json:"mode_of_payment_weekly,omitempty"`
	ModeOfPaymentSemiMonthlyPay1 int           `json:"mode_of_payment_semi_monthly_pay_1,omitempty"`
	ModeOfPaymentSemiMonthlyPay2 int           `json:"mode_of_payment_semi_monthly_pay_2,omitempty"`

	ModeOfPayment core.LoanModeOfPayment `json:"mode_of_payment"`
	Accounts      []*core.AccountRequest `json:"accounts,omitempty"`

	CashOnHandAccountID *uuid.UUID `json:"cash_on_hand_account_id,omitempty"`
	ComputationSheetID  *uuid.UUID `json:"computation_sheet_id,omitempty"`
}

type ComputationSheetAmortizationResponse struct {
	Entries     []*core.LoanTransactionEntryResponse `json:"entries"`
	TotalDebit  float64                              `json:"total_debit"`
	TotalCredit float64                              `json:"total_credit"`
	Currency    core.CurrencyResponse                `json:"currency"`

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
			entry.Credit = e.usecase.LoanChargesRateComputation(*chargesRateScheme, core.LoanTransaction{
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
		return nil, eris.Wrap(err, "failed to calculate number of payments")
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
		{
			Account: core.AccountRequest{
				GeneralLedgerDefinitionID:             account.GeneralLedgerDefinitionID,
				FinancialStatementDefinitionID:        account.FinancialStatementDefinitionID,
				AccountClassificationID:               account.AccountClassificationID,
				AccountCategoryID:                     account.AccountCategoryID,
				MemberTypeID:                          account.MemberTypeID,
				CurrencyID:                            account.CurrencyID,
				Name:                                  account.Name,
				Description:                           account.Description,
				MinAmount:                             account.MinAmount,
				MaxAmount:                             account.MaxAmount,
				Index:                                 account.Index,
				Type:                                  account.Type,
				IsInternal:                            account.IsInternal,
				CashOnHand:                            account.CashOnHand,
				PaidUpShareCapital:                    account.PaidUpShareCapital,
				ComputationType:                       account.ComputationType,
				FinesAmort:                            account.FinesAmort,
				FinesMaturity:                         account.FinesMaturity,
				InterestStandard:                      account.InterestStandard,
				InterestSecured:                       account.InterestSecured,
				ComputationSheetID:                    account.ComputationSheetID,
				CohCibFinesGracePeriodEntryCashHand:   account.CohCibFinesGracePeriodEntryCashHand,
				CohCibFinesGracePeriodEntryCashInBank: account.CohCibFinesGracePeriodEntryCashInBank,
				CohCibFinesGracePeriodEntryDailyAmortization:       account.CohCibFinesGracePeriodEntryDailyAmortization,
				CohCibFinesGracePeriodEntryDailyMaturity:           account.CohCibFinesGracePeriodEntryDailyMaturity,
				CohCibFinesGracePeriodEntryWeeklyAmortization:      account.CohCibFinesGracePeriodEntryWeeklyAmortization,
				CohCibFinesGracePeriodEntryWeeklyMaturity:          account.CohCibFinesGracePeriodEntryWeeklyMaturity,
				CohCibFinesGracePeriodEntryMonthlyAmortization:     account.CohCibFinesGracePeriodEntryMonthlyAmortization,
				CohCibFinesGracePeriodEntryMonthlyMaturity:         account.CohCibFinesGracePeriodEntryMonthlyMaturity,
				CohCibFinesGracePeriodEntrySemiMonthlyAmortization: account.CohCibFinesGracePeriodEntrySemiMonthlyAmortization,
				CohCibFinesGracePeriodEntrySemiMonthlyMaturity:     account.CohCibFinesGracePeriodEntrySemiMonthlyMaturity,
				CohCibFinesGracePeriodEntryQuarterlyAmortization:   account.CohCibFinesGracePeriodEntryQuarterlyAmortization,
				CohCibFinesGracePeriodEntryQuarterlyMaturity:       account.CohCibFinesGracePeriodEntryQuarterlyMaturity,
				CohCibFinesGracePeriodEntrySemiAnnualAmortization:  account.CohCibFinesGracePeriodEntrySemiAnnualAmortization,
				CohCibFinesGracePeriodEntrySemiAnnualMaturity:      account.CohCibFinesGracePeriodEntrySemiAnnualMaturity,
				CohCibFinesGracePeriodEntryAnnualAmortization:      account.CohCibFinesGracePeriodEntryAnnualAmortization,
				CohCibFinesGracePeriodEntryAnnualMaturity:          account.CohCibFinesGracePeriodEntryAnnualMaturity,
				CohCibFinesGracePeriodEntryLumpsumAmortization:     account.CohCibFinesGracePeriodEntryLumpsumAmortization,
				CohCibFinesGracePeriodEntryLumpsumMaturity:         account.CohCibFinesGracePeriodEntryLumpsumMaturity,
				GeneralLedgerType:                   account.GeneralLedgerType,
				LoanAccountID:                       account.LoanAccountID,
				FinesGracePeriodAmortization:        account.FinesGracePeriodAmortization,
				AdditionalGracePeriod:               account.AdditionalGracePeriod,
				NoGracePeriodDaily:                  account.NoGracePeriodDaily,
				FinesGracePeriodMaturity:            account.FinesGracePeriodMaturity,
				YearlySubscriptionFee:               account.YearlySubscriptionFee,
				CutOffDays:                          account.CutOffDays,
				CutOffMonths:                        account.CutOffMonths,
				LumpsumComputationType:              account.LumpsumComputationType,
				InterestFinesComputationDiminishing: account.InterestFinesComputationDiminishing,
				InterestFinesComputationDiminishingStraightYearly: account.InterestFinesComputationDiminishingStraightYearly,
				EarnedUnearnedInterest:                            account.EarnedUnearnedInterest,
				LoanSavingType:                                    account.LoanSavingType,
				InterestDeduction:                                 account.InterestDeduction,
				OtherDeductionEntry:                               account.OtherDeductionEntry,
				InterestSavingTypeDiminishingStraight:             account.InterestSavingTypeDiminishingStraight,
				OtherInformationOfAnAccount:                       account.OtherInformationOfAnAccount,
				HeaderRow:                                         account.HeaderRow,
				CenterRow:                                         account.CenterRow,
				TotalRow:                                          account.TotalRow,
				GeneralLedgerGroupingExcludeAccount:               account.GeneralLedgerGroupingExcludeAccount,
				Icon:                                              account.Icon,
				ShowInGeneralLedgerSourceWithdraw:                 account.ShowInGeneralLedgerSourceWithdraw,
				ShowInGeneralLedgerSourceDeposit:                  account.ShowInGeneralLedgerSourceDeposit,
				ShowInGeneralLedgerSourceJournal:                  account.ShowInGeneralLedgerSourceJournal,
				ShowInGeneralLedgerSourcePayment:                  account.ShowInGeneralLedgerSourcePayment,
				ShowInGeneralLedgerSourceAdjustment:               account.ShowInGeneralLedgerSourceAdjustment,
				ShowInGeneralLedgerSourceJournalVoucher:           account.ShowInGeneralLedgerSourceJournalVoucher,
				ShowInGeneralLedgerSourceCheckVoucher:             account.ShowInGeneralLedgerSourceCheckVoucher,
				CompassionFund:                                    account.CompassionFund,
				CompassionFundAmount:                              account.CompassionFundAmount,
				CashAndCashEquivalence:                            account.CashAndCashEquivalence,
				InterestStandardComputation:                       account.InterestStandardComputation,
			},
			Value: 0,
			Total: 0,
		},
	}
	for _, acc := range lcscr.Accounts {
		accounts = append(accounts, &AccountValue{
			Account: *acc,
			Value:   0,
			Total:   0,
		})
	}

	principal := totalCredit
	balance := totalCredit
	paymentDate := time.Now().UTC()

	for i := range numberOfPayments + 1 {
		// Find next valid payment date (skip excluded days)
		actualDate := paymentDate
		daysSkipped := 0
		checkDate := paymentDate
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
				paymentDate = checkDate
				break
			}
			paymentDate = paymentDate.AddDate(0, 0, 1)
			daysSkipped++
		}

		// Skip calculations for the first entry (loan disbursement date)
		if i > 0 {
			// Calculate account values for payments (not for disbursement)
			for _, acc := range accounts {
				switch acc.Account.ComputationType {
				case core.Straight:
					switch acc.Account.Type {
					case core.AccountTypeInterest:
						acc.Value = e.usecase.ComputeInterest(principal, acc.Account.InterestStandard, lcscr.ModeOfPayment)
						acc.Total = e.provider.Service.Decimal.Add(acc.Total, acc.Value)
					case core.AccountTypeSVFLedger:
						acc.Value = e.usecase.ComputeInterest(principal, acc.Account.InterestStandard, lcscr.ModeOfPayment)
						acc.Total = e.provider.Service.Decimal.Add(acc.Total, acc.Value)
					}
				case core.Diminishing:
					switch acc.Account.Type {
					case core.AccountTypeInterest:
						acc.Value = e.usecase.ComputeInterest(balance, acc.Account.InterestStandard, lcscr.ModeOfPayment)
						acc.Total = e.provider.Service.Decimal.Add(acc.Total, acc.Value)
					case core.AccountTypeSVFLedger:
						acc.Value = e.usecase.ComputeInterest(balance, acc.Account.InterestStandard, lcscr.ModeOfPayment)
						acc.Total = e.provider.Service.Decimal.Add(acc.Total, acc.Value)
					}
				case core.DiminishingStraight:
					switch acc.Account.Type {
					case core.AccountTypeInterest:
						acc.Value = e.usecase.ComputeInterest(principal, acc.Account.InterestStandard, lcscr.ModeOfPayment)
						acc.Total = e.provider.Service.Decimal.Add(acc.Total, acc.Value)
					case core.AccountTypeSVFLedger:
						acc.Value = e.usecase.ComputeInterest(principal, acc.Account.InterestStandard, lcscr.ModeOfPayment)
						acc.Total = e.provider.Service.Decimal.Add(acc.Total, acc.Value)
					}
				}

				if acc.Account.Type == core.AccountTypeFines {
					if daysSkipped > 0 && !acc.Account.NoGracePeriodDaily {
						acc.Value = e.usecase.ComputeFines(
							principal,
							acc.Account.FinesAmort,
							acc.Account.FinesMaturity,
							daysSkipped,
							lcscr.ModeOfPayment,
							acc.Account.NoGracePeriodDaily,
							acc.Account,
						)
						acc.Total = e.provider.Service.Decimal.Add(acc.Total, acc.Value)
					} else {
						acc.Value = 0
					}
				}
			}

			// Calculate loan principal payment
			for _, acc := range accounts {
				if acc.Account.Type == core.AccountTypeLoan {
					// Calculate principal payment for this period
					principalPayment := e.provider.Service.Decimal.Divide(principal, float64(numberOfPayments))

					// Ensure we don't pay more than the remaining balance
					acc.Value = e.provider.Service.Decimal.Clamp(principalPayment, 0, balance)
					acc.Total = e.provider.Service.Decimal.Add(acc.Total, acc.Value)

					// Update the remaining balance
					balance = e.provider.Service.Decimal.Subtract(balance, acc.Value)

					fmt.Printf("Payment %d - Principal: %f, Payments: %d, Payment Amount: %f, Balance: %f\n",
						i, principal, numberOfPayments, acc.Value, balance)
				}
			}
		} else {
			// First entry (disbursement date) - reset all account values to 0
			for _, acc := range accounts {
				acc.Value = 0
			}
			fmt.Printf("Disbursement Date - No payment calculations\n")
		}

		// Add to amortization schedule AFTER calculations
		switch lcscr.ModeOfPayment {
		case core.LoanModeOfPaymentDaily:
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				Balance:       balance,
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
			paymentDate = paymentDate.AddDate(0, 0, 1)
		case core.LoanModeOfPaymentWeekly:
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				Balance:       balance,
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
			weekDay := e.core.LoanWeeklyIota(weeklyExactDay)
			paymentDate = e.nextWeekday(paymentDate, time.Weekday(weekDay))
		case core.LoanModeOfPaymentSemiMonthly:
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				Balance:       balance,
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
			// Semi-monthly logic...
			thisDay := paymentDate.Day()
			thisMonth := paymentDate.Month()
			thisYear := paymentDate.Year()
			loc := paymentDate.Location()

			if thisDay < semiMonthlyExactDay1 {
				paymentDate = time.Date(thisYear, thisMonth, semiMonthlyExactDay1, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			} else if thisDay < semiMonthlyExactDay2 {
				paymentDate = time.Date(thisYear, thisMonth, semiMonthlyExactDay2, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			} else {
				nextMonth := paymentDate.AddDate(0, 1, 0)
				paymentDate = time.Date(nextMonth.Year(), nextMonth.Month(), semiMonthlyExactDay1, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			}
		case core.LoanModeOfPaymentMonthly:
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				Balance:       balance,
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
			loc := paymentDate.Location()
			day := paymentDate.Day()
			if isMonthlyExactDay {
				nextMonth := paymentDate.AddDate(0, 1, 0)
				paymentDate = time.Date(nextMonth.Year(), nextMonth.Month(), day, paymentDate.Hour(), paymentDate.Minute(), paymentDate.Second(), paymentDate.Nanosecond(), loc)
			} else {
				paymentDate = paymentDate.AddDate(0, 1, 0)
			}
		case core.LoanModeOfPaymentQuarterly:
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				Balance:       balance,
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
			paymentDate = paymentDate.AddDate(0, 3, 0)
		case core.LoanModeOfPaymentSemiAnnual:
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				Balance:       balance,
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
			paymentDate = paymentDate.AddDate(0, 6, 0)
		case core.LoanModeOfPaymentLumpsum:
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				Balance:       balance,
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
		case core.LoanModeOfPaymentFixedDays:
			amortization = append(amortization, &LoanAmortizationScheduleResponse{
				Balance:       balance,
				ScheduledDate: paymentDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         e.sumAccountValues(accounts),
				Accounts:      accounts,
			})
			paymentDate = paymentDate.AddDate(0, 0, 1)
		}
	}

	return &ComputationSheetAmortizationResponse{
		Entries:     e.core.LoanTransactionEntryManager.ToModels(loanTransactionEntries),
		Currency:    *e.core.CurrencyManager.ToModel(currency),
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
