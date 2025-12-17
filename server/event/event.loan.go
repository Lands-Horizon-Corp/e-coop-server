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

type LoanBalanceEvent struct {
	CashOnCashEquivalenceAccountID uuid.UUID
	LoanTransactionID              uuid.UUID
}

func (e *Event) LoanBalancing(ctx context.Context, echoCtx echo.Context, tx *gorm.DB, endTx func(error) error, data LoanBalanceEvent) (*core.LoanTransaction, error) {
	fmt.Println("[DEBUG LoanBalancing] 01 - START: Entering LoanBalancing function")
	fmt.Printf("[DEBUG LoanBalancing] 01 - Input data: LoanTransactionID=%s, CashOnCashEquivalenceAccountID=%s\n", data.LoanTransactionID, data.CashOnCashEquivalenceAccountID)

	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(ctx, echoCtx)
	if err != nil {
		fmt.Println("[DEBUG LoanBalancing] ERROR 02 - Failed to get user organization:", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to get user organization"))
	}
	fmt.Printf("[DEBUG LoanBalancing] 02 - SUCCESS: Got userOrg - OrgID=%s, BranchID=%s, UserID=%s\n", userOrg.OrganizationID, *userOrg.BranchID, userOrg.UserID)

	loanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, data.LoanTransactionID)
	if err != nil {
		fmt.Println("[DEBUG LoanBalancing] ERROR 03 - Failed to get loan transaction:", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction"))
	}
	fmt.Printf("[DEBUG LoanBalancing] 03 - SUCCESS: Got loanTransaction - ID=%s, Applied1=%.2f, LoanType=%v, AccountID=%v\n",
		loanTransaction.ID, loanTransaction.Applied1, loanTransaction.LoanType, loanTransaction.AccountID)

	if loanTransaction.AccountID == nil {
		fmt.Println("[DEBUG LoanBalancing] PANIC RISK 04 - loanTransaction.AccountID is nil!")
	}
	account, err := e.core.AccountManager.GetByID(ctx, *loanTransaction.AccountID)
	if err != nil {
		fmt.Println("[DEBUG LoanBalancing] ERROR 04 - Failed to get loan account:", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan account during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan account"))
	}
	fmt.Printf("[DEBUG LoanBalancing] 04 - SUCCESS: Got account - ID=%s, Name=%s, ComputationSheetID=%v\n", account.ID, account.Name, account.ComputationSheetID)

	loanTransactionEntries, err := e.core.LoanTransactionEntryManager.Find(ctx, &core.LoanTransactionEntry{
		LoanTransactionID: loanTransaction.ID,
		OrganizationID:    userOrg.OrganizationID,
		BranchID:          *userOrg.BranchID,
	})
	if err != nil {
		fmt.Println("[DEBUG LoanBalancing] ERROR 05 - Failed to get loan transaction entries:", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction entries during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction entries"))
	}
	fmt.Printf("[DEBUG LoanBalancing] 05 - SUCCESS: Got %d loan transaction entries\n", len(loanTransactionEntries))

	automaticLoanDeductions, err := e.core.AutomaticLoanDeductionManager.Find(ctx, &core.AutomaticLoanDeduction{
		OrganizationID:     userOrg.OrganizationID,
		BranchID:           *userOrg.BranchID,
		ComputationSheetID: account.ComputationSheetID,
	})
	disableLoanDeduction := loanTransaction.LoanType == core.LoanTypeRenewalWithoutDeduct || loanTransaction.LoanType == core.LoanTypeRestructured || loanTransaction.LoanType == core.LoanTypeStandardPrevious
	fmt.Printf("[DEBUG LoanBalancing] 06 - disableLoanDeduction=%v, err=%v\n", disableLoanDeduction, err)
	if err != nil || disableLoanDeduction {
		fmt.Println("[DEBUG LoanBalancing] 06 - Automatic deductions disabled or error - setting to empty")
		automaticLoanDeductions = []*core.AutomaticLoanDeduction{}
	} else {
		fmt.Printf("[DEBUG LoanBalancing] 06 - SUCCESS: Got %d automatic loan deductions\n", len(automaticLoanDeductions))
	}

	result := []*core.LoanTransactionEntry{}
	static, deduction, postComputed := []*core.LoanTransactionEntry{}, []*core.LoanTransactionEntry{}, []*core.LoanTransactionEntry{}

	fmt.Println("[DEBUG LoanBalancing] 07 - Categorizing existing entries")
	for i, entry := range loanTransactionEntries {
		fmt.Printf("[DEBUG LoanBalancing] 07 - Entry %d: Type=%v, IsAddOn=%v, Credit=%.2f\n", i, entry.Type, entry.IsAddOn, entry.Credit)
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
	fmt.Printf("[DEBUG LoanBalancing] 07 - Categorized: static=%d, deduction=%d, postComputed=%d\n", len(static), len(deduction), len(postComputed))

	if len(static) < 2 {
		fmt.Println("[DEBUG LoanBalancing] 08 - Less than 2 static entries - creating default cash entries")
		cashOnCashEquivalenceAccount, err := e.core.AccountManager.GetByID(ctx, data.CashOnCashEquivalenceAccountID)
		if err != nil {
			fmt.Println("[DEBUG LoanBalancing] ERROR 08 - Failed to get cash on cash equivalence account:", err)
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to get cash on cash equivalence account during loan balancing: " + err.Error(),
				Module:      "LoanBalancing",
			})
			return nil, endTx(eris.Wrap(err, "failed to get cash on cash equivalence account"))
		}
		fmt.Printf("[DEBUG LoanBalancing] 08 - SUCCESS: Got cash equivalence account - ID=%s, Name=%s, CashAndCashEquivalence=%v\n",
			cashOnCashEquivalenceAccount.ID, cashOnCashEquivalenceAccount.Name, cashOnCashEquivalenceAccount.CashAndCashEquivalence)

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
		fmt.Println("[DEBUG LoanBalancing] 09 - Created 2 default static entries")
	}

	fmt.Println("[DEBUG LoanBalancing] 10 - Ordering static entries")
	if len(static) >= 2 {
		if static[0].Account != nil && static[0].Account.CashAndCashEquivalence {
			result = append(result, static[0])
			result = append(result, static[1])
			fmt.Println("[DEBUG LoanBalancing] 10 - Order: Cash first, then Loan")
		} else {
			result = append(result, static[1])
			result = append(result, static[0])
			fmt.Println("[DEBUG LoanBalancing] 10 - Order: Loan first, then Cash")
		}
	}

	addOnEntry := &core.LoanTransactionEntry{
		Account:           nil,
		Credit:            0,
		Debit:             0,
		Name:              "ADD ON INTEREST",
		Type:              core.LoanTransactionAddOn,
		LoanTransactionID: loanTransaction.ID,
		IsAddOn:           true,
	}

	totalNonAddOns, totalAddOns := 0.0, 0.0
	fmt.Println("[DEBUG LoanBalancing] 11 - Processing deduction entries")
	for _, entry := range deduction {
		if !entry.IsAddOn {
			totalNonAddOns = e.provider.Service.Decimal.Add(totalNonAddOns, entry.Credit)
		} else {
			totalAddOns = e.provider.Service.Decimal.Add(totalAddOns, entry.Credit)
		}
		result = append(result, entry)
		fmt.Printf("[DEBUG LoanBalancing] 11 - Added deduction: Name=%s, Credit=%.2f, IsAddOn=%v\n", entry.Name, entry.Credit, entry.IsAddOn)
	}

	fmt.Println("[DEBUG LoanBalancing] 12 - Processing postComputed (automatic deductions from entries)")
	for _, entry := range postComputed {
		fmt.Printf("[DEBUG LoanBalancing] 12 - Processing postComputed entry: Name=%s, Amount=%.2f, IsAddOn=%v\n", entry.Name, entry.Amount, entry.IsAddOn)
		if entry.IsAutomaticLoanDeductionDeleted {
			result = append(result, entry)
			continue
		}
		if entry.Amount != 0 {
			entry.Credit = entry.Amount
		} else {
			if entry.AutomaticLoanDeduction.ChargesRateSchemeID != nil {
				fmt.Println("[DEBUG LoanBalancing] 12 - Has ChargesRateSchemeID - fetching scheme")
				chargesRateScheme, err := e.core.ChargesRateSchemeManager.GetByID(ctx, *entry.AutomaticLoanDeduction.ChargesRateSchemeID)
				if err != nil {
					fmt.Println("[DEBUG LoanBalancing] ERROR 12 - Failed to get charges rate scheme:", err)
					return nil, endTx(err)
				}
				entry.Credit = e.usecase.LoanChargesRateComputation(*chargesRateScheme, *loanTransaction)
				fmt.Printf("[DEBUG LoanBalancing] 12 - Computed via ChargesRate: Credit=%.2f\n", entry.Credit)
			}
			if entry.Credit <= 0 {
				entry.Credit = e.usecase.LoanComputation(*entry.AutomaticLoanDeduction, *loanTransaction)
				fmt.Printf("[DEBUG LoanBalancing] 12 - Computed via LoanComputation: Credit=%.2f\n", entry.Credit)
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

	fmt.Println("[DEBUG LoanBalancing] 13 - Adding missing automatic deductions")
	for _, ald := range automaticLoanDeductions {
		exist := false
		for _, computed := range postComputed {
			if handlers.UUIDPtrEqual(&ald.ID, computed.AutomaticLoanDeductionID) {
				exist = true
				break
			}
		}
		if !exist {
			fmt.Printf("[DEBUG LoanBalancing] 13 - Adding missing ALD: Name=%s, AddOn=%v\n", ald.Name, ald.AddOn)
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
					fmt.Println("[DEBUG LoanBalancing] ERROR 13 - Failed to get charges rate scheme:", err)
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
			fmt.Printf("[DEBUG LoanBalancing] 13 - Added missing ALD with Credit=%.2f\n", entry.Credit)
		}
	}

	fmt.Printf("[DEBUG LoanBalancing] 14 - Totals so far: totalNonAddOns=%.2f, totalAddOns=%.2f\n", totalNonAddOns, totalAddOns)

	if (loanTransaction.LoanType == core.LoanTypeRestructured ||
		loanTransaction.LoanType == core.LoanTypeRenewalWithoutDeduct ||
		loanTransaction.LoanType == core.LoanTypeRenewal) && loanTransaction.PreviousLoanID != nil {
		fmt.Println("[DEBUG LoanBalancing] 15 - Adding previous loan balance entry")
		previous := loanTransaction.PreviousLoan
		if previous == nil {
			fmt.Println("[DEBUG LoanBalancing] PANIC RISK 15 - previous loan is nil despite PreviousLoanID not nil")
		} else {
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
			fmt.Printf("[DEBUG LoanBalancing] 15 - Added previous balance: %.2f\n", previous.Balance)
		}
	}

	fmt.Println("[DEBUG LoanBalancing] 16 - Adjusting first entry (principal/credit side)")
	if loanTransaction.IsAddOn {
		result[0].Credit = e.provider.Service.Decimal.Subtract(loanTransaction.Applied1, totalNonAddOns)
		fmt.Printf("[DEBUG LoanBalancing] 16 - IsAddOn=true - New Credit=%.2f (Applied1 - totalNonAddOns)\n", result[0].Credit)
	} else {
		totalDeductions := e.provider.Service.Decimal.Add(totalNonAddOns, totalAddOns)
		result[0].Credit = e.provider.Service.Decimal.Subtract(loanTransaction.Applied1, totalDeductions)
		fmt.Printf("[DEBUG LoanBalancing] 16 - IsAddOn=false - New Credit=%.2f (Applied1 - totalDeductions)\n", result[0].Credit)
	}

	fmt.Println("[DEBUG LoanBalancing] 17 - Deleting old transaction entries")
	for _, entry := range loanTransactionEntries {
		if entry.ID == uuid.Nil {
			continue
		}
		fmt.Printf("[DEBUG LoanBalancing] 17 - Deleting entry ID=%s\n", entry.ID)
		if err := e.core.LoanTransactionEntryManager.DeleteWithTx(ctx, tx, entry.ID); err != nil {
			fmt.Println("[DEBUG LoanBalancing] ERROR 17 - Failed to delete entry:", err)
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to delete existing loan transaction entries during loan balancing: " + err.Error(),
				Module:      "LoanBalancing",
			})
			return nil, endTx(eris.Wrap(err, "failed to delete existing loan transaction entries: "+err.Error()))
		}
	}

	fmt.Println("[DEBUG LoanBalancing] 18 - Setting debit on second static entry and name")
	if len(result) >= 2 {
		result[1].Debit = loanTransaction.Applied1
		fmt.Printf("[DEBUG LoanBalancing] 18 - Set Debit=%.2f on result[1]\n", result[1].Debit)
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
		fmt.Printf("[DEBUG LoanBalancing] 18 - Set Name on result[1]: %s\n", result[1].Name)
	}

	if loanTransaction.IsAddOn && totalAddOns > 0 {
		fmt.Printf("[DEBUG LoanBalancing] 19 - Adding ADD ON INTEREST entry with Debit=%.2f\n", totalAddOns)
		addOnEntry.Debit = totalAddOns
		result = append(result, addOnEntry)
	}

	totalDebit, totalCredit := 0.0, 0.0
	fmt.Println("[DEBUG LoanBalancing] 20 - Creating new loan transaction entries")
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
		if !entry.IsAutomaticLoanDeductionDeleted {
			totalDebit = e.provider.Service.Decimal.Add(totalDebit, entry.Debit)
			totalCredit = e.provider.Service.Decimal.Add(totalCredit, entry.Credit)
		}
		fmt.Printf("[DEBUG LoanBalancing] 20 - Creating entry Index=%d, Name=%s, Credit=%.2f, Debit=%.2f\n", index, entry.Name, entry.Credit, entry.Debit)
		if err := e.core.LoanTransactionEntryManager.CreateWithTx(ctx, tx, value); err != nil {
			fmt.Println("[DEBUG LoanBalancing] ERROR 20 - Failed to create entry:", err)
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to create loan transaction entry during loan balancing: " + err.Error(),
				Module:      "LoanBalancing",
			})
			return nil, endTx(eris.Wrap(err, "failed to create loan transaction entry: "+err.Error()))
		}
	}
	fmt.Printf("[DEBUG LoanBalancing] 20 - Totals after create: totalCredit=%.2f, totalDebit=%.2f\n", totalCredit, totalDebit)

	var amountGranted float64 = 0
	if loanTransaction.IsAddOn {
		amountGranted = e.provider.Service.Decimal.Add(totalCredit, totalAddOns)
	} else {
		amountGranted = totalCredit
	}
	fmt.Printf("[DEBUG LoanBalancing] 21 - amountGranted calculated: %.2f\n", amountGranted)

	amort, err := e.usecase.LoanModeOfPayment(amountGranted, loanTransaction)
	if err != nil {
		fmt.Println("[DEBUG LoanBalancing] ERROR 21 - Failed to calculate amortization:", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to calculate loan amortization during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to calculate loan amortization: "+err.Error()))
	}
	fmt.Printf("[DEBUG LoanBalancing] 21 - Amortization calculated: %.2f\n", amort)

	loanTransaction.Amortization = amort
	loanTransaction.AmountGranted = amountGranted
	loanTransaction.TotalAddOn = totalAddOns
	loanTransaction.TotalPrincipal = totalCredit
	loanTransaction.Balance = totalCredit
	loanTransaction.TotalCredit = totalCredit
	loanTransaction.TotalDebit = totalDebit
	loanTransaction.UpdatedAt = time.Now().UTC()
	loanTransaction.UpdatedByID = userOrg.UserID

	fmt.Println("[DEBUG LoanBalancing] 22 - Updating loan transaction record")
	if err := e.core.LoanTransactionManager.UpdateByIDWithTx(ctx, tx, loanTransaction.ID, loanTransaction); err != nil {
		fmt.Println("[DEBUG LoanBalancing] ERROR 22 - Failed to update loan transaction:", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to update loan transaction during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to update loan transaction: "+err.Error()))
	}

	fmt.Println("[DEBUG LoanBalancing] 23 - Committing transaction")
	if err := endTx(nil); err != nil {
		fmt.Println("[DEBUG LoanBalancing] ERROR 23 - Failed to commit:", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "db-commit-error",
			Description: "Failed to commit transaction during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
	}

	fmt.Println("[DEBUG LoanBalancing] 24 - Fetching updated loan transaction")
	newLoanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, loanTransaction.ID)
	if err != nil {
		fmt.Println("[DEBUG LoanBalancing] ERROR 24 - Failed to get updated loan transaction:", err)
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get updated loan transaction during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to get updated loan transaction"))
	}

	fmt.Println("[DEBUG LoanBalancing] 99 - SUCCESS: LoanBalancing completed")
	return newLoanTransaction, nil
}
