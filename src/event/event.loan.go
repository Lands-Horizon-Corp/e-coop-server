package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/usecase"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type LoanBalanceEvent struct {
	CashOnCashEquivalenceAccountID uuid.UUID
	LoanTransactionID              uuid.UUID
}

func LoanBalancing(
	ctx context.Context,
	service *horizon.HorizonService,
	tx *gorm.DB, endTx func(error) error,
	data LoanBalanceEvent,
	userOrg *types.UserOrganization) (*types.LoanTransaction, error) {

	// =========================
	// STEP 2: LOAN TRANSACTION & ACCOUNT
	// =========================
	loanTransaction, err := core.LoanTransactionManager(service).GetByID(ctx, data.LoanTransactionID)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction"))
	}

	account, err := core.AccountManager(service).GetByID(ctx, *loanTransaction.AccountID)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to get loan account"))
	}

	loanTransactionEntries, err := core.LoanTransactionEntryManager(service).Find(ctx, &types.LoanTransactionEntry{
		LoanTransactionID: loanTransaction.ID,
		OrganizationID:    userOrg.OrganizationID,
		BranchID:          *userOrg.BranchID,
	})
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction entries"))
	}

	automaticLoanDeductions, err := core.AutomaticLoanDeductionManager(service).Find(ctx, &types.AutomaticLoanDeduction{
		OrganizationID:     userOrg.OrganizationID,
		BranchID:           *userOrg.BranchID,
		ComputationSheetID: account.ComputationSheetID,
	})

	disableLoanDeduction := loanTransaction.LoanType == types.LoanTypeRenewalWithoutDeduct ||
		loanTransaction.LoanType == types.LoanTypeRestructured ||
		loanTransaction.LoanType == types.LoanTypeStandardPrevious
	if err != nil || disableLoanDeduction {
		automaticLoanDeductions = []*types.AutomaticLoanDeduction{}
	}

	// =========================
	// STEP 3: CATEGORIZE ENTRIES
	// =========================
	result := []*types.LoanTransactionEntry{}
	static, deduction, postComputed := []*types.LoanTransactionEntry{}, []*types.LoanTransactionEntry{}, []*types.LoanTransactionEntry{}
	for _, entry := range loanTransactionEntries {
		switch entry.Type {
		case types.LoanTransactionStatic:
			static = append(static, entry)
		case types.LoanTransactionDeduction:
			deduction = append(deduction, entry)
		case types.LoanTransactionAutomaticDeduction:
			if !disableLoanDeduction {
				postComputed = append(postComputed, entry)
			}
		}
	}

	// =========================
	// STEP 4: DEFAULT STATIC ENTRIES
	// =========================
	if len(static) < 2 {
		cashOnCashEquivalenceAccount, err := core.AccountManager(service).GetByID(ctx, data.CashOnCashEquivalenceAccountID)
		if err != nil {
			return nil, endTx(eris.Wrap(err, "failed to get cash on cash equivalence account"))
		}

		static = []*types.LoanTransactionEntry{
			{
				Credit:            decimal.NewFromFloat(loanTransaction.Applied1).InexactFloat64(),
				Debit:             0,
				Description:       cashOnCashEquivalenceAccount.Description,
				Account:           cashOnCashEquivalenceAccount,
				AccountID:         &cashOnCashEquivalenceAccount.ID,
				Name:              cashOnCashEquivalenceAccount.Name,
				Type:              types.LoanTransactionStatic,
				LoanTransactionID: loanTransaction.ID,
			},
			{
				Credit:            0,
				Debit:             decimal.NewFromFloat(loanTransaction.Applied1).InexactFloat64(),
				Account:           loanTransaction.Account,
				AccountID:         loanTransaction.AccountID,
				Description:       loanTransaction.Account.Description,
				Name:              loanTransaction.Account.Name,
				Type:              types.LoanTransactionStatic,
				LoanTransactionID: loanTransaction.ID,
			},
		}
	}

	// =========================
	// STEP 5: ARRANGE STATIC
	// =========================
	if static[0].Account.CashAndCashEquivalence {
		result = append(result, static[0], static[1])
	} else {
		result = append(result, static[1], static[0])
	}

	// =========================
	// STEP 6: PROCESS DEDUCTIONS
	// =========================
	totalNonAddOns, totalAddOns := decimal.Zero, decimal.Zero

	for _, entry := range deduction {
		c := decimal.NewFromFloat(entry.Credit)
		if entry.IsAddOn {
			totalAddOns = totalAddOns.Add(c)
		} else {
			totalNonAddOns = totalNonAddOns.Add(c)
		}
		result = append(result, entry)
	}

	for _, entry := range postComputed {
		if entry.IsAutomaticLoanDeductionDeleted {
			result = append(result, entry)
			continue
		}

		if entry.Amount != 0 {
			entry.Credit = entry.Amount
		} else if entry.AutomaticLoanDeduction.ChargesRateSchemeID != nil {
			chargesRateScheme, err := core.ChargesRateSchemeManager(service).GetByID(ctx, *entry.AutomaticLoanDeduction.ChargesRateSchemeID)
			if err != nil {
				return nil, endTx(err)
			}
			entry.Credit = usecase.LoanChargesRateComputation(*chargesRateScheme, *loanTransaction)
		}

		if decimal.NewFromFloat(entry.Credit).LessThanOrEqual(decimal.Zero) {
			entry.Credit = usecase.LoanComputation(*entry.AutomaticLoanDeduction, *loanTransaction)
		}

		c := decimal.NewFromFloat(entry.Credit)
		if entry.IsAddOn {
			totalAddOns = totalAddOns.Add(c)
		} else {
			totalNonAddOns = totalNonAddOns.Add(c)
		}

		if entry.Credit > 0 {
			result = append(result, entry)
		}
	}

	// =========================
	// STEP 7: ADD MISSING AUTOMATIC DEDUCTIONS
	// =========================
	for _, ald := range automaticLoanDeductions {
		exist := false
		for _, computed := range postComputed {
			if helpers.UUIDPtrEqual(&ald.ID, computed.AutomaticLoanDeductionID) {
				exist = true
				break
			}
		}
		entry := &types.LoanTransactionEntry{
			Credit:                   0,
			Debit:                    0,
			Name:                     ald.Name,
			Type:                     types.LoanTransactionAutomaticDeduction,
			IsAddOn:                  ald.AddOn,
			Account:                  ald.Account,
			AccountID:                ald.AccountID,
			Description:              ald.Account.Description,
			AutomaticLoanDeductionID: &ald.ID,
			LoanTransactionID:        loanTransaction.ID,
		}

		if !exist {
			if ald.ChargesRateSchemeID != nil {
				chargesRateScheme, err := core.ChargesRateSchemeManager(service).GetByID(ctx, *ald.ChargesRateSchemeID)
				if err != nil {
					return nil, endTx(err)
				}
				entry.Credit = usecase.LoanChargesRateComputation(*chargesRateScheme, *loanTransaction)
			}

			if decimal.NewFromFloat(entry.Credit).LessThanOrEqual(decimal.Zero) {
				entry.Credit = usecase.LoanComputation(*ald, *loanTransaction)
			}
			c := decimal.NewFromFloat(entry.Credit)
			if entry.IsAddOn {
				totalAddOns = totalAddOns.Add(c)
			} else {
				totalNonAddOns = totalNonAddOns.Add(c)
			}

			if entry.Credit > 0 {
				result = append(result, entry)
			}
		}
	}

	// =========================
	// STEP 8: PREVIOUS LOAN BALANCES
	// =========================
	if (loanTransaction.LoanType == types.LoanTypeRestructured ||
		loanTransaction.LoanType == types.LoanTypeRenewalWithoutDeduct ||
		loanTransaction.LoanType == types.LoanTypeRenewal) && loanTransaction.PreviousLoanID != nil {

		prev := loanTransaction.PreviousLoan
		result = append(result, &types.LoanTransactionEntry{
			Account:           prev.Account,
			AccountID:         prev.AccountID,
			Credit:            prev.Balance,
			Debit:             0,
			Name:              prev.Account.Name,
			Description:       prev.Account.Description,
			Type:              types.LoanTransactionPrevious,
			LoanTransactionID: loanTransaction.ID,
		})
		totalNonAddOns = totalNonAddOns.Add(decimal.NewFromFloat(prev.Balance))
	}

	// =========================
	// STEP 9: FINAL CASH EQUIVALENT
	// =========================
	applied := decimal.NewFromFloat(loanTransaction.Applied1)
	if loanTransaction.IsAddOn {
		result[0].Credit = applied.Sub(totalNonAddOns).InexactFloat64()
	} else {
		totalDeductions := totalNonAddOns.Add(totalAddOns)
		result[0].Credit = applied.Sub(totalDeductions).InexactFloat64()
	}

	// =========================
	// STEP 10: DELETE OLD ENTRIES
	// =========================
	for _, entry := range loanTransactionEntries {
		if entry.ID == uuid.Nil {
			continue
		}
		if err := core.LoanTransactionEntryManager(service).DeleteWithTx(ctx, tx, entry.ID); err != nil {
			return nil, endTx(eris.Wrap(err, "failed to delete existing loan transaction entry"))
		}
	}

	// =========================
	// STEP 11: UPDATE LOAN ACCOUNT ENTRY
	// =========================
	result[1].Debit = applied.InexactFloat64()
	switch loanTransaction.LoanType {
	case types.LoanTypeRestructured:
		result[1].Name += " - RESTRUCTURED"
	case types.LoanTypeRenewal, types.LoanTypeRenewalWithoutDeduct:
		result[1].Name += " - CURRENT"
	}

	if loanTransaction.IsAddOn && totalAddOns.GreaterThan(decimal.Zero) {
		result[1].Name += fmt.Sprintf(" + Add On Interest (+%s)", totalAddOns.String())
		result[1].Debit = decimal.NewFromFloat(result[1].Debit).Add(totalAddOns).InexactFloat64()
	}

	// =========================
	// STEP 12: CREATE NEW ENTRIES & COMPUTE TOTALS
	// =========================
	totalCredit, totalDebit := decimal.Zero, decimal.Zero
	for index, entry := range result {
		newEntry := &types.LoanTransactionEntry{
			CreatedAt:                       time.Now().UTC(),
			CreatedByID:                     userOrg.UserID,
			UpdatedAt:                       time.Now().UTC(),
			UpdatedByID:                     userOrg.UserID,
			OrganizationID:                  userOrg.OrganizationID,
			BranchID:                        *userOrg.BranchID,
			LoanTransactionID:               loanTransaction.ID,
			Index:                           index,
			Type:                            entry.Type,
			IsAddOn:                         entry.IsAddOn,
			AccountID:                       entry.AccountID,
			AutomaticLoanDeductionID:        entry.AutomaticLoanDeductionID,
			Name:                            entry.Name,
			Description:                     entry.Description,
			Credit:                          entry.Credit,
			Debit:                           entry.Debit,
			Amount:                          entry.Amount,
			IsAutomaticLoanDeductionDeleted: entry.IsAutomaticLoanDeductionDeleted,
		}

		if !entry.IsAutomaticLoanDeductionDeleted {
			totalDebit = totalDebit.Add(decimal.NewFromFloat(entry.Debit))
			totalCredit = totalCredit.Add(decimal.NewFromFloat(entry.Credit))
		}

		if err := core.LoanTransactionEntryManager(service).CreateWithTx(ctx, tx, newEntry); err != nil {
			return nil, endTx(eris.Wrap(err, "failed to create loan transaction entry"))
		}
	}

	// =========================
	// STEP 13: CALCULATE AMORTIZATION & UPDATE LOAN
	// =========================
	amort, err := usecase.LoanModeOfPayment(loanTransaction)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to calculate loan amortization"))
	}

	loanTransaction.Amortization = amort
	loanTransaction.TotalPrincipal = totalCredit.InexactFloat64()
	loanTransaction.TotalCredit = totalCredit.InexactFloat64()
	loanTransaction.Balance = totalCredit.InexactFloat64()
	loanTransaction.TotalDebit = totalDebit.InexactFloat64()
	loanTransaction.UpdatedAt = time.Now().UTC()
	loanTransaction.UpdatedByID = userOrg.UserID

	if err := core.LoanTransactionManager(service).UpdateByIDWithTx(ctx, tx, loanTransaction.ID, loanTransaction); err != nil {
		return nil, endTx(eris.Wrap(err, "failed to update loan transaction"))
	}

	if err := endTx(nil); err != nil {
		return nil, endTx(eris.Wrap(err, "failed to do loan balancing"))
	}

	// =========================
	// STEP 14: RETURN UPDATED LOAN TRANSACTION
	// =========================
	newLoanTransaction, err := core.LoanTransactionManager(service).GetByID(ctx, loanTransaction.ID)
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to get updated loan transaction"))
	}

	return newLoanTransaction, nil
}
