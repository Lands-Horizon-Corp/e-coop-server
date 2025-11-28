package event

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type ComputationSheetAccountValue struct {
	Account *core.AccountRequest `json:"account" validate:"required"`
	Value   float64              `json:"value" validate:"required,gte=0"`
	Total   float64              `json:"total" validate:"required,gte=0"`
}

type ComputationSheetScheduleResponse struct {
	ScheduledDate time.Time                       `json:"scheduled_date"`
	ActualDate    time.Time                       `json:"actual_date"`
	DaysSkipped   int                             `json:"days_skipped"`
	Total         float64                         `json:"total"`
	Balance       float64                         `json:"balance"`
	Accounts      []*ComputationSheetAccountValue `json:"accounts"`
}

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
	Total       float64                              `json:"total"`
	Schedule    []*ComputationSheetScheduleResponse  `json:"schedule,omitempty"`
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
			Credit:                 0,
			Debit:                  0,
			Name:                   ald.Name,
			Type:                   core.LoanTransactionDeduction,
			IsAddOn:                ald.AddOn,
			Account:                ald.Account,
			AutomaticLoanDeduction: ald,
		}
		if ald.ChargesRateSchemeID != nil { // Use ald instead of entry.AutomaticLoanDeduction
			chargesRateScheme, err := e.core.ChargesRateSchemeManager.GetByID(context, *ald.ChargesRateSchemeID)
			if err != nil {
				return nil, eris.Wrap(err, fmt.Sprintf("failed to get charges rate scheme for automatic loan deduction ID %s", ald.ID))
			}

			// Create MemberProfile with proper nil handling
			memberProfile := &core.MemberProfile{}
			if lcscr.MemberTypeID != nil {
				memberProfile.MemberTypeID = lcscr.MemberTypeID
			}

			entry.Credit = e.usecase.LoanChargesRateComputation(*chargesRateScheme, core.LoanTransaction{
				Applied1:      lcscr.Applied1,
				Terms:         lcscr.Terms,
				MemberProfile: memberProfile,
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
	if !e.provider.Service.Decimal.IsEqual(totalDebit, totalCredit) {
		return nil, eris.New("debit and credit are not equal")
	}

	// Loan Amortization Schedule ==========================================
	holidays, err := e.core.HolidayManager.Find(context, &core.Holiday{
		OrganizationID: computationSheet.OrganizationID,
		BranchID:       computationSheet.BranchID,
		CurrencyID:     *account.CurrencyID,
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
	weeklyExactDay := lcscr.ModeOfPaymentWeekly
	semiMonthlyExactDay1 := lcscr.ModeOfPaymentSemiMonthlyPay1
	semiMonthlyExactDay2 := lcscr.ModeOfPaymentSemiMonthlyPay2

	// Typically, start date comes from loanTransaction (adjust as needed)
	amortization := []*ComputationSheetScheduleResponse{}
	accountsSchedule := []*ComputationSheetAccountValue{}
	for _, acc := range lcscr.Accounts {
		accountsSchedule = append(accountsSchedule, &ComputationSheetAccountValue{
			Account: acc,
			Value:   0,
			Total:   0,
		})
	}
	accountsSchedule = append(accountsSchedule, &ComputationSheetAccountValue{
		Account: &core.AccountRequest{
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
	})

	principal := totalCredit
	balance := totalCredit
	paymentDate := time.Now().UTC()
	total := 0.0

	for i := range numberOfPayments + 1 {
		actualDate := paymentDate
		daysSkipped := 0
		rowTotal := 0.0
		daysSkipped, err := e.skippedDaysCount(paymentDate, currency, excludeSaturday, excludeSunday, excludeHolidays, holidays)
		if err != nil {
			return nil, err
		}

		scheduledDate := paymentDate.AddDate(0, 0, daysSkipped)

		// ✅ CREATE INDEPENDENT ACCOUNT SLICE FOR THIS PERIOD
		periodAccounts := make([]*ComputationSheetAccountValue, len(accountsSchedule))

		if i > 0 {
			for j := range accountsSchedule {

				// Create a new account entry for this period
				periodAccounts[j] = &ComputationSheetAccountValue{
					Account: accountsSchedule[j].Account,
					Value:   0,                         // Will be calculated below
					Total:   accountsSchedule[j].Total, // Carry over cumulative total
				}

				switch accountsSchedule[j].Account.Type {
				case core.AccountTypeLoan:
					periodAccounts[j].Value = e.provider.Service.Decimal.Clamp(
						e.provider.Service.Decimal.Divide(principal, float64(numberOfPayments)), 0, balance)

					// Update cumulative total in original slice
					accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
					periodAccounts[j].Total = accountsSchedule[j].Total

					balance = e.provider.Service.Decimal.Subtract(balance, periodAccounts[j].Value)

				case core.AccountTypeFines:
					if daysSkipped > 0 && !accountsSchedule[j].Account.NoGracePeriodDaily {
						periodAccounts[j].Value = e.usecase.ComputeFines(
							principal,
							accountsSchedule[j].Account.FinesAmort,
							accountsSchedule[j].Account.FinesMaturity,
							daysSkipped,
							lcscr.ModeOfPayment,
							accountsSchedule[j].Account.NoGracePeriodDaily,
							e.convertAccountRequestToAccount(accountsSchedule[j].Account),
						)

						// Update cumulative total in original slice
						accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
						periodAccounts[j].Total = accountsSchedule[j].Total
					}

				default:
					// Interest calculations...
					switch accountsSchedule[j].Account.ComputationType {
					case core.Straight:
						if accountsSchedule[j].Account.Type == core.AccountTypeInterest || accountsSchedule[j].Account.Type == core.AccountTypeSVFLedger {
							periodAccounts[j].Value = e.usecase.ComputeInterest(principal, accountsSchedule[j].Account.InterestStandard, lcscr.ModeOfPayment)

							accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
							periodAccounts[j].Total = accountsSchedule[j].Total
						}
					case core.Diminishing:
						if accountsSchedule[j].Account.Type == core.AccountTypeInterest || accountsSchedule[j].Account.Type == core.AccountTypeSVFLedger {
							periodAccounts[j].Value = e.usecase.ComputeInterest(balance, accountsSchedule[j].Account.InterestStandard, lcscr.ModeOfPayment)

							accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
							periodAccounts[j].Total = accountsSchedule[j].Total
						}
					case core.DiminishingStraight:
						if accountsSchedule[j].Account.Type == core.AccountTypeInterest || accountsSchedule[j].Account.Type == core.AccountTypeSVFLedger {
							periodAccounts[j].Value = e.usecase.ComputeInterest(balance, accountsSchedule[j].Account.InterestStandard, lcscr.ModeOfPayment)

							accountsSchedule[j].Total = e.provider.Service.Decimal.Add(accountsSchedule[j].Total, periodAccounts[j].Value)
							periodAccounts[j].Total = accountsSchedule[j].Total
						}
					}
				}

				total = e.provider.Service.Decimal.Add(total, periodAccounts[j].Value)
				rowTotal = e.provider.Service.Decimal.Add(rowTotal, periodAccounts[j].Value)
			}
		} else {
			// First iteration (i=0), just copy the structure
			for j := range accountsSchedule {
				periodAccounts[j] = &ComputationSheetAccountValue{
					Account: accountsSchedule[j].Account,
					Value:   0,
					Total:   0,
				}
			}
		}

		// Sort accounts in desired order: Loan → Interest → SVF → Fines
		sort.Slice(periodAccounts, func(i, j int) bool {
			return getAccountTypePriority(
				periodAccounts[i].Account.Type) <
				getAccountTypePriority(periodAccounts[j].Account.Type)
		})

		// ✅ NOW append with period-specific accounts
		switch lcscr.ModeOfPayment {
		case core.LoanModeOfPaymentDaily:
			amortization = append(amortization, &ComputationSheetScheduleResponse{
				Balance:       balance,
				ActualDate:    actualDate,
				ScheduledDate: scheduledDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      periodAccounts, // Each period has its own independent slice!
			})
			paymentDate = paymentDate.AddDate(0, 0, 1)
		case core.LoanModeOfPaymentWeekly:
			amortization = append(amortization, &ComputationSheetScheduleResponse{
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
			amortization = append(amortization, &ComputationSheetScheduleResponse{
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
			amortization = append(amortization, &ComputationSheetScheduleResponse{
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
			amortization = append(amortization, &ComputationSheetScheduleResponse{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      periodAccounts,
			})
			paymentDate = paymentDate.AddDate(0, 3, 0)
		case core.LoanModeOfPaymentSemiAnnual:
			amortization = append(amortization, &ComputationSheetScheduleResponse{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      periodAccounts,
			})
			paymentDate = paymentDate.AddDate(0, 6, 0)
		case core.LoanModeOfPaymentLumpsum:
			amortization = append(amortization, &ComputationSheetScheduleResponse{
				Balance:       balance,
				ScheduledDate: scheduledDate,
				ActualDate:    actualDate,
				DaysSkipped:   daysSkipped,
				Total:         rowTotal,
				Accounts:      periodAccounts,
			})
		case core.LoanModeOfPaymentFixedDays:
			amortization = append(amortization, &ComputationSheetScheduleResponse{
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

	return &ComputationSheetAmortizationResponse{
		Entries:     e.core.LoanTransactionEntryManager.ToModels(loanTransactionEntries),
		Currency:    *e.core.CurrencyManager.ToModel(currency),
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		Total:       total,
		Schedule:    amortization,
	}, nil
}

// Add helper function
func (e *Event) convertAccountRequestToAccount(req *core.AccountRequest) core.Account {
	return core.Account{
		GeneralLedgerDefinitionID:             req.GeneralLedgerDefinitionID,
		FinancialStatementDefinitionID:        req.FinancialStatementDefinitionID,
		AccountClassificationID:               req.AccountClassificationID,
		AccountCategoryID:                     req.AccountCategoryID,
		MemberTypeID:                          req.MemberTypeID,
		CurrencyID:                            req.CurrencyID,
		Name:                                  req.Name,
		Description:                           req.Description,
		MinAmount:                             req.MinAmount,
		MaxAmount:                             req.MaxAmount,
		Index:                                 req.Index,
		Type:                                  req.Type,
		IsInternal:                            req.IsInternal,
		CashOnHand:                            req.CashOnHand,
		PaidUpShareCapital:                    req.PaidUpShareCapital,
		ComputationType:                       req.ComputationType,
		FinesAmort:                            req.FinesAmort,
		FinesMaturity:                         req.FinesMaturity,
		InterestStandard:                      req.InterestStandard,
		InterestSecured:                       req.InterestSecured,
		ComputationSheetID:                    req.ComputationSheetID,
		CohCibFinesGracePeriodEntryCashHand:   req.CohCibFinesGracePeriodEntryCashHand,
		CohCibFinesGracePeriodEntryCashInBank: req.CohCibFinesGracePeriodEntryCashInBank,
		CohCibFinesGracePeriodEntryDailyAmortization:       req.CohCibFinesGracePeriodEntryDailyAmortization,
		CohCibFinesGracePeriodEntryDailyMaturity:           req.CohCibFinesGracePeriodEntryDailyMaturity,
		CohCibFinesGracePeriodEntryWeeklyAmortization:      req.CohCibFinesGracePeriodEntryWeeklyAmortization,
		CohCibFinesGracePeriodEntryWeeklyMaturity:          req.CohCibFinesGracePeriodEntryWeeklyMaturity,
		CohCibFinesGracePeriodEntryMonthlyAmortization:     req.CohCibFinesGracePeriodEntryMonthlyAmortization,
		CohCibFinesGracePeriodEntryMonthlyMaturity:         req.CohCibFinesGracePeriodEntryMonthlyMaturity,
		CohCibFinesGracePeriodEntrySemiMonthlyAmortization: req.CohCibFinesGracePeriodEntrySemiMonthlyAmortization,
		CohCibFinesGracePeriodEntrySemiMonthlyMaturity:     req.CohCibFinesGracePeriodEntrySemiMonthlyMaturity,
		CohCibFinesGracePeriodEntryQuarterlyAmortization:   req.CohCibFinesGracePeriodEntryQuarterlyAmortization,
		CohCibFinesGracePeriodEntryQuarterlyMaturity:       req.CohCibFinesGracePeriodEntryQuarterlyMaturity,
		CohCibFinesGracePeriodEntrySemiAnnualAmortization:  req.CohCibFinesGracePeriodEntrySemiAnnualAmortization,
		CohCibFinesGracePeriodEntrySemiAnnualMaturity:      req.CohCibFinesGracePeriodEntrySemiAnnualMaturity,
		CohCibFinesGracePeriodEntryAnnualAmortization:      req.CohCibFinesGracePeriodEntryAnnualAmortization,
		CohCibFinesGracePeriodEntryAnnualMaturity:          req.CohCibFinesGracePeriodEntryAnnualMaturity,
		CohCibFinesGracePeriodEntryLumpsumAmortization:     req.CohCibFinesGracePeriodEntryLumpsumAmortization,
		CohCibFinesGracePeriodEntryLumpsumMaturity:         req.CohCibFinesGracePeriodEntryLumpsumMaturity,
		GeneralLedgerType:                   req.GeneralLedgerType,
		LoanAccountID:                       req.LoanAccountID,
		FinesGracePeriodAmortization:        req.FinesGracePeriodAmortization,
		AdditionalGracePeriod:               req.AdditionalGracePeriod,
		NoGracePeriodDaily:                  req.NoGracePeriodDaily,
		FinesGracePeriodMaturity:            req.FinesGracePeriodMaturity,
		YearlySubscriptionFee:               req.YearlySubscriptionFee,
		CutOffDays:                          req.CutOffDays,
		CutOffMonths:                        req.CutOffMonths,
		LumpsumComputationType:              req.LumpsumComputationType,
		InterestFinesComputationDiminishing: req.InterestFinesComputationDiminishing,
		InterestFinesComputationDiminishingStraightYearly: req.InterestFinesComputationDiminishingStraightYearly,
		EarnedUnearnedInterest:                            req.EarnedUnearnedInterest,
		LoanSavingType:                                    req.LoanSavingType,
		InterestDeduction:                                 req.InterestDeduction,
		OtherDeductionEntry:                               req.OtherDeductionEntry,
		InterestSavingTypeDiminishingStraight:             req.InterestSavingTypeDiminishingStraight,
		OtherInformationOfAnAccount:                       req.OtherInformationOfAnAccount,
		HeaderRow:                                         req.HeaderRow,
		CenterRow:                                         req.CenterRow,
		TotalRow:                                          req.TotalRow,
		GeneralLedgerGroupingExcludeAccount:               req.GeneralLedgerGroupingExcludeAccount,
		Icon:                                              req.Icon,
		ShowInGeneralLedgerSourceWithdraw:                 req.ShowInGeneralLedgerSourceWithdraw,
		ShowInGeneralLedgerSourceDeposit:                  req.ShowInGeneralLedgerSourceDeposit,
		ShowInGeneralLedgerSourceJournal:                  req.ShowInGeneralLedgerSourceJournal,
		ShowInGeneralLedgerSourcePayment:                  req.ShowInGeneralLedgerSourcePayment,
		ShowInGeneralLedgerSourceAdjustment:               req.ShowInGeneralLedgerSourceAdjustment,
		ShowInGeneralLedgerSourceJournalVoucher:           req.ShowInGeneralLedgerSourceJournalVoucher,
		ShowInGeneralLedgerSourceCheckVoucher:             req.ShowInGeneralLedgerSourceCheckVoucher,
		CompassionFund:                                    req.CompassionFund,
		CompassionFundAmount:                              req.CompassionFundAmount,
		CashAndCashEquivalence:                            req.CashAndCashEquivalence,
		InterestStandardComputation:                       req.InterestStandardComputation,
	}
}

// getAccountTypePriority returns the sort priority for account types
// Lower numbers = higher priority (sorted first)
func getAccountTypePriority(accountType core.AccountType) int {
	switch accountType {
	case core.AccountTypeLoan:
		return 1 // First
	case core.AccountTypeInterest:
		return 2 // Second
	case core.AccountTypeSVFLedger:
		return 3 // Third
	case core.AccountTypeFines:
		return 4 // Fourth (Last)
	default:
		return 5 // Everything else goes last
	}
}
