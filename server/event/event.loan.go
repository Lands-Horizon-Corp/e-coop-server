package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type LoanBalanceEvent struct {
	CashOnCashEquivalenceAccountID uuid.UUID
	LoanTransactionID              uuid.UUID
}

func (e *Event) LoanBalancing(ctx context.Context, echoCtx echo.Context, tx *gorm.DB, endTx func(error) error, data LoanBalanceEvent) (*core.LoanTransaction, error) {

	userOrg, err := e.CurrentUserOrganization(ctx, echoCtx)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to get user organization"))
	}

	loanTransaction, err := e.core.LoanTransactionManager().GetByID(ctx, data.LoanTransactionID)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction"))
	}

	if loanTransaction.AccountID == nil {
	}
	account, err := e.core.AccountManager().GetByID(ctx, *loanTransaction.AccountID)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan account during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan account"))
	}

	loanTransactionEntries, err := e.core.LoanTransactionEntryManager().Find(ctx, &core.LoanTransactionEntry{
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

	automaticLoanDeductions, err := e.core.AutomaticLoanDeductionManager().Find(ctx, &core.AutomaticLoanDeduction{
		OrganizationID:     userOrg.OrganizationID,
		BranchID:           *userOrg.BranchID,
		ComputationSheetID: account.ComputationSheetID,
	})
	disableLoanDeduction := loanTransaction.LoanType == core.LoanTypeRenewalWithoutDeduct || loanTransaction.LoanType == core.LoanTypeRestructured || loanTransaction.LoanType == core.LoanTypeStandardPrevious
	if err != nil || disableLoanDeduction {
		automaticLoanDeductions = []*core.AutomaticLoanDeduction{}
	} else {
	}

	result := []*core.LoanTransactionEntry{}
	static, deduction, postComputed := []*core.LoanTransactionEntry{}, []*core.LoanTransactionEntry{}, []*core.LoanTransactionEntry{}

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

	if len(static) < 2 {
		cashOnCashEquivalenceAccount, err := e.core.AccountManager().GetByID(ctx, data.CashOnCashEquivalenceAccountID)
		if err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to get cash on cash equivalence account during loan balancing: " + err.Error(),
				Module:      "LoanBalancing",
			})
			return nil, endTx(eris.Wrap(err, "failed to get cash on cash equivalence account"))
		}

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

	if len(static) >= 2 {
		if static[0].Account != nil && static[0].Account.CashAndCashEquivalence {
			result = append(result, static[0])
			result = append(result, static[1])
		} else {
			result = append(result, static[1])
			result = append(result, static[0])
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

	totalNonAddOnsDec := decimal.Zero
	totalAddOnsDec := decimal.Zero

	for _, entry := range deduction {
		creditDec := decimal.NewFromFloat(entry.Credit)

		if !entry.IsAddOn {
			totalNonAddOnsDec = totalNonAddOnsDec.Add(creditDec)
		} else {
			totalAddOnsDec = totalAddOnsDec.Add(creditDec)
		}

		result = append(result, entry)
	}
	totalNonAddOns := totalNonAddOnsDec.InexactFloat64()
	totalAddOns := totalAddOnsDec.InexactFloat64()

	for _, entry := range postComputed {
		if entry.IsAutomaticLoanDeductionDeleted {
			result = append(result, entry)
			continue
		}

		if entry.Amount != 0 {
			entry.Credit = entry.Amount
		} else {
			if entry.AutomaticLoanDeduction.ChargesRateSchemeID != nil {
				chargesRateScheme, err := e.core.ChargesRateSchemeManager().GetByID(ctx, *entry.AutomaticLoanDeduction.ChargesRateSchemeID)
				if err != nil {
					return nil, endTx(err)
				}
				entry.Credit = usecase.LoanChargesRateComputation(*chargesRateScheme, *loanTransaction)
			}

			if entry.Credit <= 0 {
				entry.Credit = usecase.LoanComputation(*entry.AutomaticLoanDeduction, *loanTransaction)
			}
		}
		creditDec := decimal.NewFromFloat(entry.Credit)

		if !entry.IsAddOn {
			totalNonAddOnsDec = totalNonAddOnsDec.Add(creditDec)
		} else {
			totalAddOnsDec = totalAddOnsDec.Add(creditDec)
		}

		if creditDec.GreaterThan(decimal.Zero) {
			result = append(result, entry)
		}
	}

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

			// Compute credit using rate scheme if available
			if ald.ChargesRateSchemeID != nil {
				chargesRateScheme, err := e.core.ChargesRateSchemeManager().GetByID(ctx, *ald.ChargesRateSchemeID)
				if err != nil {
					return nil, endTx(err)
				}
				entry.Credit = usecase.LoanChargesRateComputation(*chargesRateScheme, *loanTransaction)
			}

			// Fallback computation
			if entry.Credit <= 0 {
				entry.Credit = usecase.LoanComputation(*ald, *loanTransaction)
			}

			// Update totals using shopspring/decimal
			creditDec := decimal.NewFromFloat(entry.Credit)
			if !entry.IsAddOn {
				totalNonAddOnsDec = totalNonAddOnsDec.Add(creditDec)
			} else {
				totalAddOnsDec = totalAddOnsDec.Add(creditDec)
			}
			totalNonAddOns = totalNonAddOnsDec.InexactFloat64()
			totalAddOns = totalAddOnsDec.InexactFloat64()

			if entry.Credit > 0 {
				result = append(result, entry)
			}
		}
	}

	if (loanTransaction.LoanType == core.LoanTypeRestructured ||
		loanTransaction.LoanType == core.LoanTypeRenewalWithoutDeduct ||
		loanTransaction.LoanType == core.LoanTypeRenewal) && loanTransaction.PreviousLoanID != nil {
		previous := loanTransaction.PreviousLoan
		if previous == nil {
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
			totalNonAddOnsDec := decimal.NewFromFloat(totalNonAddOns)
			totalNonAddOnsDec = totalNonAddOnsDec.Add(decimal.NewFromFloat(previous.Balance))
			totalNonAddOns = totalNonAddOnsDec.InexactFloat64()
		}
	}

	if loanTransaction.IsAddOn {
		result[0].Credit = decimal.NewFromFloat(loanTransaction.Applied1).Sub(decimal.NewFromFloat(totalNonAddOns)).InexactFloat64()
	} else {
		totalDeductionsDec := decimal.NewFromFloat(totalNonAddOns).Add(decimal.NewFromFloat(totalAddOns))
		result[0].Credit = decimal.NewFromFloat(loanTransaction.Applied1).Sub(totalDeductionsDec).InexactFloat64()
	}
	for _, entry := range loanTransactionEntries {
		if entry.ID == uuid.Nil {
			continue
		}
		if err := e.core.LoanTransactionEntryManager().DeleteWithTx(ctx, tx, entry.ID); err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to delete existing loan transaction entries during loan balancing: " + err.Error(),
				Module:      "LoanBalancing",
			})
			return nil, endTx(eris.Wrap(err, "failed to delete existing loan transaction entries: "+err.Error()))
		}
	}
	if len(result) >= 2 {
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
	}

	if loanTransaction.IsAddOn && totalAddOns > 0 {
		addOnEntry.Debit = totalAddOns
		result = append(result, addOnEntry)
	}

	totalDebitDec := decimal.Zero
	totalCreditDec := decimal.Zero

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
			totalDebitDec = totalDebitDec.Add(decimal.NewFromFloat(entry.Debit))
			totalCreditDec = totalCreditDec.Add(decimal.NewFromFloat(entry.Credit))
		}
		if err := e.core.LoanTransactionEntryManager().CreateWithTx(ctx, tx, value); err != nil {
			e.Footstep(echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to create loan transaction entry during loan balancing: " + err.Error(),
				Module:      "LoanBalancing",
			})
			return nil, endTx(eris.Wrap(err, "failed to create loan transaction entry: "+err.Error()))
		}
	}

	amountGrantedDec := totalCreditDec
	if loanTransaction.IsAddOn {
		amountGrantedDec = amountGrantedDec.Add(decimal.NewFromFloat(totalAddOns))
	}
	amountGranted := amountGrantedDec.InexactFloat64()

	amort, err := usecase.LoanModeOfPayment(amountGranted, loanTransaction)
	if err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to calculate loan amortization during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to calculate loan amortization: "+err.Error()))
	}
	loanTransaction.Amortization = amort
	loanTransaction.AmountGranted = amountGranted
	loanTransaction.TotalAddOn = totalAddOns
	loanTransaction.TotalPrincipal = totalCreditDec.InexactFloat64()
	loanTransaction.Balance = totalCreditDec.InexactFloat64()
	loanTransaction.TotalCredit = totalCreditDec.InexactFloat64()
	loanTransaction.TotalDebit = totalDebitDec.InexactFloat64()
	loanTransaction.UpdatedAt = time.Now().UTC()
	loanTransaction.UpdatedByID = userOrg.UserID

	if err := e.core.LoanTransactionManager().UpdateByIDWithTx(ctx, tx, loanTransaction.ID, loanTransaction); err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to update loan transaction during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
		return nil, endTx(eris.Wrap(err, "failed to update loan transaction: "+err.Error()))
	}

	if err := endTx(nil); err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "db-commit-error",
			Description: "Failed to commit transaction during loan balancing: " + err.Error(),
			Module:      "LoanBalancing",
		})
	}

	newLoanTransaction, err := e.core.LoanTransactionManager().GetByID(ctx, loanTransaction.ID)
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
