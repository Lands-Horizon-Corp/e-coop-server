package event

import (
	"context"
	"fmt"

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

	ExcludeSaturday              bool                   `json:"exclude_saturday"`
	ExcludeSunday                bool                   `json:"exclude_sunday"`
	ExcludeHoliday               bool                   `json:"exclude_holiday"`
	ModeOfPaymentMonthlyExactDay bool                   `json:"mode_of_payment_monthly_exact_day"`
	ModeOfPaymentWeekly          core.Weekdays          `json:"mode_of_payment_weekly"`
	ModeOfPaymentSemiMonthlyPay1 int                    `json:"mode_of_payment_semi_monthly_pay_1"`
	ModeOfPaymentSemiMonthlyPay2 int                    `json:"mode_of_payment_semi_monthly_pay_2"`
	ModeOfPayment                core.LoanModeOfPayment `json:"mode_of_payment"`
	AccountsIDs                  uuid.UUIDs             `json:"account_ids,omitempty"`

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
	return &ComputationSheetAmortizationResponse{
		Entries:     e.core.LoanTransactionEntryManager.ToModels(loanTransactionEntries),
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		Schedule:    []*LoanAmortizationScheduleResponse{},
	}, nil
}
