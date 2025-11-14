package event

import (
	"context"
	"fmt"
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
	fmt.Printf("[DEBUG] Starting LoanBalancing for transaction ID: %s\n", data.LoanTransactionID)
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(ctx, echoCtx)
	if err != nil {
		fmt.Printf("[ERROR] Failed to get user organization: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, endTx(eris.Wrap(err, "failed to get user organization"))
	}
	fmt.Printf("[DEBUG] User organization retrieved: %s (Branch: %s)\n", userOrg.OrganizationID, *userOrg.BranchID)
	// ================================================================================
	// STEP 2: LOAN TRANSACTION & RELATED DATA RETRIEVAL
	// ================================================================================
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, data.LoanTransactionID)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction"))
	}
	fmt.Printf("[DEBUG] Retrieved loan transaction: ID=%s, Applied1=%.2f, LoanType=%s\n", loanTransaction.ID, loanTransaction.Applied1, loanTransaction.LoanType)

	// Get the account associated with the loan transaction
	fmt.Printf("[DEBUG] Getting account with ID: %s\n", *loanTransaction.AccountID)
	account, err := e.core.AccountManager.GetByID(ctx, *loanTransaction.AccountID)
	if err != nil {
		fmt.Printf("[ERROR] Failed to get account: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get cash on cash equivalence parent account (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, endTx(eris.Wrap(err, "failed to get cash on cash equivalence parent account"))
	}
	fmt.Printf("[DEBUG] Account retrieved: %s (ComputationSheetID: %s)\n", account.Name, account.ComputationSheetID)

	// Get existing loan transaction entries
	fmt.Printf("[DEBUG] Getting loan transaction entries for transaction: %s\n", loanTransaction.ID)
	loanTransactionEntries, err := e.core.LoanTransactionEntryManager.Find(ctx, &core.LoanTransactionEntry{
		LoanTransactionID: loanTransaction.ID,
		OrganizationID:    userOrg.OrganizationID,
		BranchID:          *userOrg.BranchID,
	})
	if err != nil {
		fmt.Printf("[ERROR] Failed to get loan transaction entries: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction entries (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction entries"))
	}
	fmt.Printf("[DEBUG] Found %d existing loan transaction entries\n", len(loanTransactionEntries))

	// Get automatic loan deductions for the computation sheet
	fmt.Printf("[DEBUG] Getting automatic loan deductions for computation sheet: %s\n", account.ComputationSheetID)
	automaticLoanDeductions, err := e.core.AutomaticLoanDeductionManager.Find(ctx, &core.AutomaticLoanDeduction{
		OrganizationID:     userOrg.OrganizationID,
		BranchID:           *userOrg.BranchID,
		ComputationSheetID: account.ComputationSheetID,
	})
	disableLoanDeduction := loanTransaction.LoanType == core.LoanTypeRenewalWithoutDeduct || loanTransaction.LoanType == core.LoanTypeRestructured || loanTransaction.LoanType == core.LoanTypeStandardPrevious
	if err != nil || disableLoanDeduction {
		if err != nil {
			fmt.Printf("[ERROR] Failed to get automatic loan deductions: %v\n", err)
		}
		fmt.Printf("[DEBUG] Disabling automatic loan deductions (error: %v, disableFlag: %t)\n", err != nil, disableLoanDeduction)
		automaticLoanDeductions = []*core.AutomaticLoanDeduction{}
	}
	fmt.Printf("[DEBUG] Found %d automatic loan deductions\n", len(automaticLoanDeductions))

	// ================================================================================
	// STEP 3: CATEGORIZE EXISTING LOAN TRANSACTION ENTRIES BY TYPE
	// ================================================================================
	result := []*core.LoanTransactionEntry{}
	static, addOn, deduction, postComputed := []*core.LoanTransactionEntry{}, []*core.LoanTransactionEntry{}, []*core.LoanTransactionEntry{}, []*core.LoanTransactionEntry{}

	// Categorize existing entries by their transaction type
	for _, entry := range loanTransactionEntries {
		if entry.Type == core.LoanTransactionStatic {
			static = append(static, entry)
		}
		if entry.Type == core.LoanTransactionAddOn {
			addOn = append(addOn, entry)
		}
		if entry.Type == core.LoanTransactionDeduction {
			deduction = append(deduction, entry)
		}
		if entry.Type == core.LoanTransactionAutomaticDeduction && !disableLoanDeduction {
			postComputed = append(postComputed, entry)
		}
	}
	fmt.Printf("[DEBUG] Categorized entries: static=%d, addOn=%d, deduction=%d, postComputed=%d, disableLoanDeduction=%t\n", len(static), len(addOn), len(deduction), len(postComputed), disableLoanDeduction)
	// ================================================================================
	// STEP 4: CREATE DEFAULT STATIC ENTRIES IF NOT EXISTS
	// ================================================================================
	// If we don't have the required 2 static entries, create them
	if len(static) < 2 {
		fmt.Printf("[DEBUG] Creating static entries (current count: %d)\n", len(static))
		cashOnCashEquivalenceAccount, err := e.core.AccountManager.GetByID(ctx, data.CashOnCashEquivalenceAccountID)
		if err != nil {
			fmt.Printf("[ERROR] Failed to get cash on cash equivalence account: %v\n", err)
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to get cash on cash equivalence account (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return nil, endTx(eris.Wrap(err, "failed to get cash on cash equivalence account"))
		}
		fmt.Printf("[DEBUG] Cash equivalence account: %s\n", cashOnCashEquivalenceAccount.Name)

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
	// STEP 5: ARRANGE STATIC ENTRIES & CLEANUP EXISTING ADD-ON ENTRIES
	// ================================================================================
	// Order static entries: cash equivalent first, then loan account
	fmt.Printf("[DEBUG] Arranging static entries (first entry is cash equivalent: %t)\n", static[0].Account.CashAndCashEquivalence)
	if static[0].Account.CashAndCashEquivalence {
		result = append(result, static[0])
		result = append(result, static[1])
	} else {
		result = append(result, static[1])
		result = append(result, static[0])
	}
	fmt.Printf("[DEBUG] Static entries arranged - Cash: %.2f, Loan: %.2f\n", result[0].Credit, result[1].Debit)

	// Delete existing add-on entries (they will be recalculated)
	for _, entry := range addOn {
		if err := e.core.LoanTransactionEntryManager.DeleteWithTx(ctx, tx, entry.ID); err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to delete existing add on interest entries (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return nil, endTx(eris.Wrap(err, "failed to delete existing add on interest entries + "+err.Error()))
		}
	}

	// ================================================================================
	// STEP 6: PREPARE ADD-ON INTEREST ENTRY TEMPLATE
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
	fmt.Printf("[DEBUG] Starting deduction processing with %d deduction entries\n", len(deduction))

	// Add existing deduction entries and calculate running totals using precise decimal arithmetic
	for _, entry := range deduction {
		if !entry.IsAddOn {
			totalNonAddOns = e.provider.Service.Decimal.Add(totalNonAddOns, entry.Credit)
		} else {
			totalAddOns = e.provider.Service.Decimal.Add(totalAddOns, entry.Credit)
		}
		result = append(result, entry)
	}
	fmt.Printf("[DEBUG] After deduction processing: totalNonAddOns=%.2f, totalAddOns=%.2f\n", totalNonAddOns, totalAddOns)

	// Process post-computed (automatic deduction) entries
	fmt.Printf("[DEBUG] Processing %d post-computed entries\n", len(postComputed))

	for i, entry := range postComputed {
		fmt.Printf("[DEBUG] Processing post-computed entry %d: %s (deleted: %t, amount: %.2f)\n", i, entry.Name, entry.IsAutomaticLoanDeductionDeleted, entry.Amount)

		if entry.IsAutomaticLoanDeductionDeleted {
			result = append(result, entry)
			continue
		}

		if entry.Amount != 0 {
			entry.Credit = entry.Amount
			fmt.Printf("[DEBUG] Using entry amount as credit: %.2f\n", entry.Credit)
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
	fmt.Printf("[DEBUG] After post-computed processing: totalNonAddOns=%.2f, totalAddOns=%.2f\n", totalNonAddOns, totalAddOns)

	// ================================================================================
	// STEP 8: ADD MISSING AUTOMATIC DEDUCTIONS
	// ================================================================================

	fmt.Printf("[DEBUG] Checking for missing automatic deductions (%d total)\n", len(automaticLoanDeductions))
	for i, ald := range automaticLoanDeductions {
		fmt.Printf("[DEBUG] Checking automatic deduction %d: %s\n", i, ald.Name)
		exist := false
		for _, computed := range postComputed {
			if handlers.UUIDPtrEqual(&ald.ID, computed.AutomaticLoanDeductionID) {
				exist = true

				break
			}
		}

		if !exist {
			fmt.Printf("[DEBUG] Adding missing automatic deduction: %s\n", ald.Name)
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

	if (loanTransaction.LoanType == core.LoanTypeRestructured ||
		loanTransaction.LoanType == core.LoanTypeRenewalWithoutDeduct ||
		loanTransaction.LoanType == core.LoanTypeRenewal) && loanTransaction.PreviousLoanID != nil {
		previous := loanTransaction.PreviousLoan
		fmt.Printf("[DEBUG] Adding previous loan entry: %s (Balance: %.2f)\n", previous.Account.Name, previous.Balance)
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
	// STEP 9: CALCULATE FINAL CREDIT AMOUNTS & ADD-ON INTEREST
	// ================================================================================
	fmt.Printf("[DEBUG] Final calculation: Applied1=%.2f, totalNonAddOns=%.2f, totalAddOns=%.2f, IsAddOn=%t\n", loanTransaction.Applied1, totalNonAddOns, totalAddOns, loanTransaction.IsAddOn)

	// Adjust the first entry (cash equivalent) credit based on loan type and deductions using precise decimal arithmetic
	if loanTransaction.IsAddOn {
		result[0].Credit = e.provider.Service.Decimal.Subtract(loanTransaction.Applied1, totalNonAddOns)
	} else {
		totalDeductions := e.provider.Service.Decimal.Add(totalNonAddOns, totalAddOns)
		result[0].Credit = e.provider.Service.Decimal.Subtract(loanTransaction.Applied1, totalDeductions)
	}

	// Add the add-on interest entry if applicable
	if loanTransaction.IsAddOn && totalAddOns > 0 {
		addOnEntry.Debit = totalAddOns
		result = append(result, addOnEntry)
	}
	fmt.Printf("[DEBUG] Final result[0].Credit=%.2f, result entries count=%d\n", result[0].Credit, len(result))

	// ================================================================================
	// STEP 10: CLEANUP OLD ENTRIES & CREATE NEW BALANCED ENTRIES
	// ================================================================================

	// Delete all existing loan transaction entries before creating new ones
	fmt.Printf("[DEBUG] Deleting %d existing loan transaction entries\n", len(loanTransactionEntries))
	for i, entry := range loanTransactionEntries {
		fmt.Printf("[DEBUG] Deleting entry %d: %s\n", i, entry.Name)
		if err := e.core.LoanTransactionEntryManager.DeleteWithTx(ctx, tx, entry.ID); err != nil {
			fmt.Printf("[ERROR] Failed to delete entry %d: %v\n", i, err)
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to delete existing automatic loan deduction entries (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return nil, endTx(eris.Wrap(err, "failed to delete existing automatic loan deduction entries + "+err.Error()))
		}
	}

	// Set the debit amount for the loan account entry
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

	// Create new loan transaction entries and calculate totals

	totalDebit, totalCredit := 0.0, 0.0
	fmt.Printf("[DEBUG] Creating %d new loan transaction entries\n", len(result))
	for index, entry := range result {
		fmt.Printf("[DEBUG] Creating entry %d: %s (Credit: %.2f, Debit: %.2f)\n", index, entry.Name, entry.Credit, entry.Debit)

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
		if !entry.IsAutomaticLoanDeductionDeleted {
			totalDebit = e.provider.Service.Decimal.Add(totalDebit, entry.Debit)
			totalCredit = e.provider.Service.Decimal.Add(totalCredit, entry.Credit)
		}

		if err := e.core.LoanTransactionEntryManager.CreateWithTx(ctx, tx, value); err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to create loan transaction entry (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return nil, endTx(eris.Wrap(err, "failed to create loan transaction entry + "+err.Error()))
		}
	}
	fmt.Printf("[DEBUG] Created %d entries, totalCredit=%.2f, totalDebit=%.2f\n", len(result), totalCredit, totalDebit)
	// Amortization
	fmt.Printf("[DEBUG] Calculating amortization for loan transaction\n")
	amort, err := e.usecase.LoanModeOfPayment(loanTransaction)
	if err != nil {
		fmt.Printf("[ERROR] Failed to calculate amortization: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to calculate loan amortization (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, endTx(eris.Wrap(err, "failed to calculate loan amortization + "+err.Error()))
	}
	fmt.Printf("[DEBUG] Calculated amortization: %.2f\n", amort)

	// ================================================================================
	// STEP 11: UPDATE LOAN TRANSACTION TOTALS & COMMIT CHANGES
	// ================================================================================

	// Update the loan transaction with calculated totals
	fmt.Printf("[DEBUG] Updating loan transaction with totals: Amortization=%.2f, Balance=%.2f, TotalCredit=%.2f, TotalDebit=%.2f\n", amort, totalCredit, totalCredit, totalDebit)
	loanTransaction.Amortization = amort
	loanTransaction.TotalPrincipal = totalCredit
	loanTransaction.Balance = totalCredit
	loanTransaction.TotalCredit = totalCredit
	loanTransaction.TotalDebit = totalDebit
	loanTransaction.UpdatedAt = time.Now().UTC()
	loanTransaction.UpdatedByID = userOrg.UserID
	if err := e.core.LoanTransactionManager.UpdateByIDWithTx(ctx, tx, loanTransaction.ID, loanTransaction); err != nil {
		fmt.Printf("[ERROR] Failed to update loan transaction: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to update loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, endTx(eris.Wrap(err, "failed to update loan transaction + "+err.Error()))
	}

	// Commit all database changes
	fmt.Printf("[DEBUG] Committing database transaction\n")
	if err := endTx(nil); err != nil {
		fmt.Printf("[ERROR] Failed to commit transaction: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "db-commit-error",
			Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, endTx(eris.Wrap(err, "failed to commit transaction"))
	}
	fmt.Printf("[DEBUG] Transaction committed successfully\n")

	// ================================================================================
	// STEP 12: RETRIEVE & RETURN UPDATED LOAN TRANSACTION
	// ================================================================================

	// Get the updated loan transaction with all related data
	fmt.Printf("[DEBUG] Retrieving updated loan transaction\n")
	newLoanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, loanTransaction.ID)
	if err != nil {
		fmt.Printf("[ERROR] Failed to get updated loan transaction: %v\n", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get updated loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, endTx(eris.Wrap(err, "failed to get updated loan transaction"))
	}

	fmt.Printf("[DEBUG] LoanBalancing completed successfully for transaction: %s\n", newLoanTransaction.ID)
	return newLoanTransaction, nil
}
