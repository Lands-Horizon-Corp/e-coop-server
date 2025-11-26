package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// LoanBalanceEvent contains identifiers required to balance a loan payment.
type LoanBalanceEvent struct {
	CashOnCashEquivalenceAccountID uuid.UUID
	LoanTransactionID              uuid.UUID
}

// LoanBalancing computes and persists loan transaction entries to ensure
// the loan is correctly balanced after a payment.
func (e *Event) LoanBalancing(ctx context.Context, echoCtx echo.Context, tx *gorm.DB, endTx func(error) error, data LoanBalanceEvent) (*core.LoanTransaction, error) {

	// ================================================================================
	// STEP 1: AUTHENTICATION & USER ORGANIZATION RETRIEVAL
	// ================================================================================
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(ctx, echoCtx)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to get user organization"))
	}

	// ================================================================================
	// STEP 2: LOAN TRANSACTION & RELATED DATA RETRIEVAL
	// ================================================================================
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, data.LoanTransactionID)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction"))
	}

	account, err := e.core.AccountManager.GetByID(ctx, *loanTransaction.AccountID)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan account during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan account"))
	}

	// Get existing loan transaction entries
	loanTransactionEntries, err := e.core.LoanTransactionEntryManager.Find(ctx, &core.LoanTransactionEntry{
		LoanTransactionID: loanTransaction.ID,
		OrganizationID:    userOrg.OrganizationID,
		BranchID:          *userOrg.BranchID,
	})
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction entries during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction entries"))
	}

	// Get automatic loan deductions for the computation sheet
	automaticLoanDeductions, err := e.core.AutomaticLoanDeductionManager.Find(ctx, &core.AutomaticLoanDeduction{
		OrganizationID:     userOrg.OrganizationID,
		BranchID:           *userOrg.BranchID,
		ComputationSheetID: account.ComputationSheetID,
	})
	disableLoanDeduction := loanTransaction.LoanType == core.LoanTypeRenewalWithoutDeduct || loanTransaction.LoanType == core.LoanTypeRestructured || loanTransaction.LoanType == core.LoanTypeStandardPrevious
	if err != nil || disableLoanDeduction {
		automaticLoanDeductions = []*core.AutomaticLoanDeduction{}
	}

	// ================================================================================
	// STEP 3: CATEGORIZE EXISTING LOAN TRANSACTION ENTRIES BY TYPE
	// ================================================================================
	result := []*core.LoanTransactionEntry{}
	static, deduction, postComputed := []*core.LoanTransactionEntry{}, []*core.LoanTransactionEntry{}, []*core.LoanTransactionEntry{}

	// Categorize existing entries by their transaction type
	for _, entry := range loanTransactionEntries {
		if entry.Type == core.LoanTransactionStatic {
			static = append(static, entry)
		}

		if entry.Type == core.LoanTransactionDeduction {
			deduction = append(deduction, entry)
		}
		if entry.Type == core.LoanTransactionAutomaticDeduction && !disableLoanDeduction {
			postComputed = append(postComputed, entry)
		}
	}

	// ================================================================================
	// STEP 4: CREATE DEFAULT STATIC ENTRIES IF NOT EXISTS
	// ================================================================================
	// If we don't have the required 2 static entries, create them
	if len(static) < 2 {
		cashOnCashEquivalenceAccount, err := e.core.AccountManager.GetByID(ctx, data.CashOnCashEquivalenceAccountID)
		if err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to get cash on cash equivalence account during loan balancing: " + err.Error(),
				Module:      "LoanBalancing",
			})
			return nil, endTx(eris.Wrap(err, "failed to get cash on cash equivalence account"))
		}

		// Create the two required static entries: credit to cash equivalent and debit to loan account
		static = []*core.LoanTransactionEntry{
			{
				Credit:            loanTransaction.Applied1,
				Debit:             0,
				Description:       cashOnCashEquivalenceAccount.Description,
				Account:           cashOnCashEquivalenceAccount,
				AccountID:         &cashOnCashEquivalenceAccount.ID,
				Name:              cashOnCashEquivalenceAccount.Name,
				Type:              core.LoanTransactionStatic,
				LoanTransactionID: loanTransaction.ID,
			},
			{
				Credit:            0,
				Debit:             loanTransaction.Applied1,
				Account:           loanTransaction.Account,
				AccountID:         loanTransaction.AccountID,
				Description:       loanTransaction.Account.Description,
				Name:              loanTransaction.Account.Name,
				Type:              core.LoanTransactionStatic,
				LoanTransactionID: loanTransaction.ID,
			},
		}
	}

	// ================================================================================
	// STEP 5: ARRANGE STATIC ENTRIES IN CORRECT ORDER
	// ================================================================================
	// Order static entries: cash equivalent first, then loan account
	if static[0].Account.CashAndCashEquivalence {
		result = append(result, static[0])
		result = append(result, static[1])
	} else {
		result = append(result, static[1])
		result = append(result, static[0])
	}

	// ================================================================================
	// STEP 6: PROCESS EXISTING DEDUCTIONS & CALCULATE TOTALS
	// ================================================================================

	addOnEntry := &core.LoanTransactionEntry{
		Account:           nil,
		Credit:            0,
		Debit:             0,
		Name:              "ADD ON INTEREST",
		Type:              core.LoanTransactionAddOn,
		LoanTransactionID: loanTransaction.ID,
		IsAddOn:           true,
	}

	// ================================================================================
	// STEP 7: PROCESS EXISTING DEDUCTIONS & CALCULATE TOTALS
	// ================================================================================

	totalNonAddOns, totalAddOns := 0.0, 0.0
	// Add existing deduction entries and calculate running totals using precise decimal arithmetic
	for _, entry := range deduction {
		if !entry.IsAddOn {
			totalNonAddOns = e.provider.Service.Decimal.Add(totalNonAddOns, entry.Credit)
		} else {
			totalAddOns = e.provider.Service.Decimal.Add(totalAddOns, entry.Credit)
		}
		result = append(result, entry)
	}

	// Process post-computed (automatic deduction) entries
	for _, entry := range postComputed {

		if entry.IsAutomaticLoanDeductionDeleted {
			result = append(result, entry)
			continue
		}

		if entry.Amount != 0 {
			entry.Credit = entry.Amount
		} else {
			if entry.AutomaticLoanDeduction.ChargesRateSchemeID != nil {
				chargesRateScheme, err := e.core.ChargesRateSchemeManager.GetByID(ctx, *entry.AutomaticLoanDeduction.ChargesRateSchemeID)
				if err != nil {
					return nil, endTx(err)
				}
				entry.Credit = e.usecase.LoanChargesRateComputation(*chargesRateScheme, *loanTransaction)
			}

			if entry.Credit <= 0 {
				entry.Credit = e.usecase.LoanComputation(*entry.AutomaticLoanDeduction, *loanTransaction)
			}
		}

		if !entry.IsAddOn {
			totalNonAddOns = e.provider.Service.Decimal.Add(totalNonAddOns, entry.Credit)
		} else {
			totalAddOns = e.provider.Service.Decimal.Add(totalAddOns, entry.Credit)
		}

		if entry.Credit > 0 {
			result = append(result, entry)

		}
	}
	// ================================================================================
	// STEP 7: ADD MISSING AUTOMATIC DEDUCTIONS
	// ================================================================================
	for _, ald := range automaticLoanDeductions {
		exist := false
		for _, computed := range postComputed {
			if handlers.UUIDPtrEqual(&ald.ID, computed.AutomaticLoanDeductionID) {
				exist = true
				break
			}
		}

		if !exist {
			entry := &core.LoanTransactionEntry{
				Credit:                   0,
				Debit:                    0,
				Name:                     ald.Name,
				Type:                     core.LoanTransactionAutomaticDeduction,
				IsAddOn:                  ald.AddOn,
				Account:                  ald.Account,
				AccountID:                ald.AccountID,
				Description:              ald.Account.Description,
				AutomaticLoanDeductionID: &ald.ID,
				LoanTransactionID:        loanTransaction.ID,
				Amount:                   0,
			}

			if ald.ChargesRateSchemeID != nil {
				chargesRateScheme, err := e.core.ChargesRateSchemeManager.GetByID(ctx, *ald.ChargesRateSchemeID)
				if err != nil {
					return nil, endTx(err)
				}
				entry.Credit = e.usecase.LoanChargesRateComputation(*chargesRateScheme, *loanTransaction)
			}

			if entry.Credit <= 0 {
				entry.Credit = e.usecase.LoanComputation(*ald, *loanTransaction)
			}

			if !entry.IsAddOn {
				totalNonAddOns = e.provider.Service.Decimal.Add(totalNonAddOns, entry.Credit)
			} else {
				totalAddOns = e.provider.Service.Decimal.Add(totalAddOns, entry.Credit)
			}

			if entry.Credit > 0 {
				result = append(result, entry)
			}
		}
	}

	// Add previous loan balance for renewal and restructured loans
	if (loanTransaction.LoanType == core.LoanTypeRestructured ||
		loanTransaction.LoanType == core.LoanTypeRenewalWithoutDeduct ||
		loanTransaction.LoanType == core.LoanTypeRenewal) && loanTransaction.PreviousLoanID != nil {
		previous := loanTransaction.PreviousLoan

		result = append(result, &core.LoanTransactionEntry{
			Account:           previous.Account,
			AccountID:         previous.AccountID,
			Credit:            previous.Balance,
			Debit:             0,
			Name:              previous.Account.Name,
			Description:       previous.Account.Description,
			Type:              core.LoanTransactionPrevious,
			LoanTransactionID: loanTransaction.ID,
		})
		totalNonAddOns = e.provider.Service.Decimal.Add(totalNonAddOns, previous.Balance)
	}

	// ================================================================================
	// STEP 8: CALCULATE FINAL CREDIT AMOUNTS FOR CASH EQUIVALENT ENTRY
	// ================================================================================
	// Adjust the first entry (cash equivalent) credit based on loan type and deductions using precise decimal arithmetic
	if loanTransaction.IsAddOn {
		result[0].Credit = e.provider.Service.Decimal.Subtract(loanTransaction.Applied1, totalNonAddOns)
	} else {
		totalDeductions := e.provider.Service.Decimal.Add(totalNonAddOns, totalAddOns)
		result[0].Credit = e.provider.Service.Decimal.Subtract(loanTransaction.Applied1, totalDeductions)
	}

	// ================================================================================
	// STEP 9: DELETE OLD ENTRIES & CREATE NEW BALANCED ENTRIES
	// ================================================================================
	// Delete all existing loan transaction entries before creating new ones
	for _, entry := range loanTransactionEntries {
		// Check if entry has a valid ID
		if entry.ID == uuid.Nil {
			continue
		}
		if err := e.core.LoanTransactionEntryManager.DeleteWithTx(ctx, tx, entry.ID); err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to delete existing loan transaction entries during loan balancing: " + err.Error(),
				Module:      "LoanBalancing",
			})
			return nil, endTx(eris.Wrap(err, "failed to delete existing loan transaction entries: "+err.Error()))
		}
	}

	// Set the debit amount for the loan account entry and update name based on loan type
	result[1].Debit = loanTransaction.Applied1
	switch loanTransaction.LoanType {
	case core.LoanTypeStandard:
		result[1].Name = loanTransaction.Account.Name
	case core.LoanTypeStandardPrevious:
		result[1].Name = loanTransaction.Account.Name
	case core.LoanTypeRestructured:
		result[1].Name = loanTransaction.Account.Name + " - RESTRUCTURED"
	case core.LoanTypeRenewal:
		result[1].Name = loanTransaction.Account.Name + " - CURRENT"
	case core.LoanTypeRenewalWithoutDeduct:
		result[1].Name = loanTransaction.Account.Name + " - CURRENT"
	}

	// Add the add-on interest entry if applicable
	if loanTransaction.IsAddOn && totalAddOns > 0 {
		addOnEntry.Debit = totalAddOns
		result = append(result, addOnEntry)
	}

	// Create new loan transaction entries and calculate totals
	totalDebit, totalCredit := 0.0, 0.0
	for index, entry := range result {
		value := &core.LoanTransactionEntry{
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

		// Only include non-deleted entries in total calculations
		if !entry.IsAutomaticLoanDeductionDeleted {
			totalDebit = e.provider.Service.Decimal.Add(totalDebit, entry.Debit)
			totalCredit = e.provider.Service.Decimal.Add(totalCredit, entry.Credit)
		}

		if err := e.core.LoanTransactionEntryManager.CreateWithTx(ctx, tx, value); err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to create loan transaction entry during loan balancing: " + err.Error(),
				Module:      "LoanBalancing",
			})
			return nil, endTx(eris.Wrap(err, "failed to create loan transaction entry: "+err.Error()))
		}
	}

	// Calculate loan amortization
	amort, err := e.usecase.LoanModeOfPayment(loanTransaction)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to calculate loan amortization during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to calculate loan amortization: "+err.Error()))
	}

	// ================================================================================
	// STEP 10: UPDATE LOAN TRANSACTION TOTALS & COMMIT CHANGES
	// ================================================================================
	// Update the loan transaction with calculated totals
	loanTransaction.Amortization = amort
	loanTransaction.AmountGranted = totalCredit
	loanTransaction.TotalAddOn = totalAddOns
	loanTransaction.TotalPrincipal = totalCredit
	loanTransaction.Balance = totalCredit
	loanTransaction.TotalCredit = totalCredit
	loanTransaction.TotalDebit = totalDebit
	loanTransaction.UpdatedAt = time.Now().UTC()
	loanTransaction.UpdatedByID = userOrg.UserID

	if err := e.core.LoanTransactionManager.UpdateByIDWithTx(ctx, tx, loanTransaction.ID, loanTransaction); err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to update loan transaction during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to update loan transaction: "+err.Error()))
	}

	// Commit the transaction
	if err := endTx(nil); err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "db-commit-error",
			Description: "Failed to commit transaction during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
	}

	// ================================================================================
	// STEP 11: RETRIEVE & RETURN UPDATED LOAN TRANSACTION
	// ================================================================================
	// Get the updated loan transaction with all related data
	newLoanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, loanTransaction.ID)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get updated loan transaction during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to get updated loan transaction"))
	}

	return newLoanTransaction, nil
}
