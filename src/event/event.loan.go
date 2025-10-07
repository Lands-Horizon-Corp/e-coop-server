package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type LoanBalanceEvent struct {
	CashOnCashEquivalenceAccountID uuid.UUID
	LoanTransactionID              uuid.UUID
}

func (e *Event) LoanBalancing(ctx context.Context, echoCtx echo.Context, tx *gorm.DB, data LoanBalanceEvent) (*model.LoanTransaction, error) {
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(ctx, echoCtx)
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get user organization")
	}

	loanTransaction, err := e.model.LoanTransactionManager.GetByID(ctx, data.LoanTransactionID)
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get loan transaction")
	}
	account, err := e.model.AccountManager.GetByID(ctx, *loanTransaction.AccountID)
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get cash on cash equivalence parent account (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get cash on cash equivalence parent account")
	}
	loanTransactionEntries, err := e.model.LoanTransactionEntryManager.Find(ctx, &model.LoanTransactionEntry{
		LoanTransactionID: loanTransaction.ID,
		OrganizationID:    userOrg.OrganizationID,
		BranchID:          *userOrg.BranchID,
	})
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction entries (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get loan transaction entries")
	}
	automaticLoanDeductions, err := e.model.AutomaticLoanDeductionManager.Find(ctx, &model.AutomaticLoanDeduction{
		OrganizationID:     userOrg.OrganizationID,
		BranchID:           *userOrg.BranchID,
		ComputationSheetID: account.ComputationSheetID,
	})
	if err != nil {
		automaticLoanDeductions = []*model.AutomaticLoanDeduction{}
	}

	result := []*model.LoanTransactionEntry{}
	static, addOn, deduction, postComputed := []*model.LoanTransactionEntry{}, []*model.LoanTransactionEntry{}, []*model.LoanTransactionEntry{}, []*model.LoanTransactionEntry{}
	for _, entry := range loanTransactionEntries {
		if entry.Type == model.LoanTransactionStatic {
			static = append(static, entry)
		}
		if entry.Type == model.LoanTransactionAddOn {
			addOn = append(addOn, entry)
		}
		if entry.Type == model.LoanTransactionDeduction {
			deduction = append(deduction, entry)
		}
		if entry.Type == model.LoanTransactionAutomaticDeduction {
			postComputed = append(postComputed, entry)
		}
	}
	if len(static) < 2 {
		cashOnCashEquivalenceAccount, err := e.model.AccountManager.GetByID(ctx, data.CashOnCashEquivalenceAccountID)
		if err != nil {
			tx.Rollback()
			e.Footstep(ctx, echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to get cash on cash equivalence account (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return nil, eris.Wrap(err, "failed to get cash on cash equivalence account")
		}

		static = []*model.LoanTransactionEntry{
			{
				Credit:            loanTransaction.Applied1,
				Debit:             0,
				Description:       cashOnCashEquivalenceAccount.Description,
				Account:           cashOnCashEquivalenceAccount,
				AccountID:         &cashOnCashEquivalenceAccount.ID,
				Name:              cashOnCashEquivalenceAccount.Name,
				Type:              model.LoanTransactionStatic,
				LoanTransactionID: loanTransaction.ID,
			},
			{
				Credit:            0,
				Debit:             loanTransaction.Applied1,
				Account:           loanTransaction.Account,
				AccountID:         loanTransaction.AccountID,
				Description:       loanTransaction.Account.Description,
				Name:              loanTransaction.Account.Name,
				Type:              model.LoanTransactionStatic,
				LoanTransactionID: loanTransaction.ID,
			},
		}
	}

	if static[0].Account.CashAndCashEquivalence {
		result = append(result, static[0])
		result = append(result, static[1])
	} else {
		result = append(result, static[1])
		result = append(result, static[0])
	}
	if len(addOn) > 1 {
		for _, entry := range addOn {
			if err := e.model.LoanTransactionEntryManager.DeleteByIDWithTx(ctx, tx, entry.ID); err != nil {
				tx.Rollback()
				e.Footstep(ctx, echoCtx, FootstepEvent{
					Activity:    "data-error",
					Description: "Failed to delete existing add on interest entries (/transaction/payment/:transaction_id): " + err.Error(),
					Module:      "Transaction",
				})
				return nil, eris.Wrap(err, "failed to delete existing add on interest entries + "+err.Error())
			}
		}
	}
	// Deductions
	addOnEntry := &model.LoanTransactionEntry{
		Account:           nil,
		Credit:            0,
		Debit:             0,
		Name:              "ADD ON INTEREST",
		Type:              model.LoanTransactionAddOn,
		LoanTransactionID: loanTransaction.ID,
		IsAddOn:           true,
	}
	total_non_add_ons, total_add_ons := 0.0, 0.0
	for _, entry := range deduction {
		if !entry.IsAddOn {
			total_non_add_ons += entry.Credit
		} else {
			total_add_ons += entry.Credit
		}
		result = append(result, entry)
	}
	for _, entry := range postComputed {
		entry.Credit = e.service.LoanComputation(ctx, *entry.AutomaticLoanDeduction, *loanTransaction)
		if !entry.IsAddOn {
			total_non_add_ons += entry.Credit
		} else {
			total_add_ons += entry.Credit
		}
		result = append(result, entry)
	}
	for _, ald := range automaticLoanDeductions {

		exist := false
		for _, computed := range postComputed {
			if ald.ID == *computed.AutomaticLoanDeductionID {
				exist = true
				break
			}
		}
		if !exist {
			entry := &model.LoanTransactionEntry{
				Credit:                   0,
				Debit:                    0,
				Name:                     ald.Name,
				Type:                     model.LoanTransactionAutomaticDeduction,
				IsAddOn:                  ald.AddOn,
				Account:                  ald.Account,
				AccountID:                ald.AccountID,
				Description:              ald.Account.Description,
				AutomaticLoanDeductionID: &ald.ID,
				LoanTransactionID:        loanTransaction.ID,
			}
			entry.Credit = e.service.LoanComputation(ctx, *ald, *loanTransaction)
			if !entry.IsAddOn {
				total_non_add_ons += entry.Credit
			} else {
				total_add_ons += entry.Credit
			}
			result = append(result, entry)
		}
	}

	if loanTransaction.IsAddOn {
		result[0].Credit = loanTransaction.Applied1 - total_non_add_ons
	} else {
		result[0].Credit = loanTransaction.Applied1 - (total_non_add_ons + total_add_ons)
	}

	if loanTransaction.IsAddOn && total_add_ons > 0 {
		addOnEntry.Debit = total_add_ons
		result = append(result, addOnEntry)
	}

	for _, entry := range loanTransactionEntries {
		if err := e.model.LoanTransactionEntryManager.DeleteByIDWithTx(ctx, tx, entry.ID); err != nil {
			tx.Rollback()
			e.Footstep(ctx, echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to delete existing automatic loan deduction entries (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return nil, eris.Wrap(err, "failed to delete existing automatic loan deduction entries + "+err.Error())
		}
	}
	totalDebit, totalCredit := 0.0, 0.0
	for index, entry := range result {
		value := &model.LoanTransactionEntry{
			CreatedAt:                time.Now().UTC(),
			CreatedByID:              userOrg.UserID,
			UpdatedAt:                time.Now().UTC(),
			UpdatedByID:              userOrg.UserID,
			OrganizationID:           userOrg.OrganizationID,
			BranchID:                 *userOrg.BranchID,
			LoanTransactionID:        loanTransaction.ID,
			Index:                    index,
			Type:                     entry.Type,
			IsAddOn:                  entry.IsAddOn,
			AccountID:                entry.AccountID,
			AutomaticLoanDeductionID: entry.AutomaticLoanDeductionID,
			Name:                     entry.Name,
			Description:              entry.Description,
			Credit:                   entry.Credit,
			Debit:                    entry.Debit,
		}
		totalDebit += entry.Debit
		totalCredit += entry.Credit
		if err := e.model.LoanTransactionEntryManager.CreateWithTx(ctx, tx, value); err != nil {
			tx.Rollback()
			e.Footstep(ctx, echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to create loan transaction entry (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return nil, eris.Wrap(err, "failed to create loan transaction entry + "+err.Error())
		}
	}
	loanTransaction.TotalCredit = totalCredit
	loanTransaction.TotalDebit = totalDebit
	if err := e.model.LoanTransactionManager.UpdateFieldsWithTx(ctx, tx, loanTransaction.ID, loanTransaction); err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to update loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to update loan transaction + "+err.Error())
	}
	if err := tx.Commit().Error; err != nil {
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "db-commit-error",
			Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to commit transaction")
	}
	newLoanTransaction, err := e.model.LoanTransactionManager.GetByID(ctx, loanTransaction.ID)
	if err != nil {
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get updated loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get updated loan transaction")
	}
	return newLoanTransaction, nil
}
