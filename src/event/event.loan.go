package event

import (
	"context"
	"fmt"

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
	fmt.Println("Line 15: Starting LoanBalancing function")
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(ctx, echoCtx)
	fmt.Println("Line 17: Retrieved user organization, err:", err)
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get user organization")
	}

	fmt.Println("Line 26: Getting cash on cash equivalence account")
	cashOnCashEquivalenceAccount, err := e.model.AccountManager.GetByID(ctx, data.CashOnCashEquivalenceAccountID)
	fmt.Println("Line 28: Retrieved cash on cash equivalence account, err:", err)
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get cash on cash equivalence account (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get cash on cash equivalence account")
	}

	fmt.Println("Line 37: Getting loan transaction")
	loanTransaction, err := e.model.LoanTransactionManager.GetByID(ctx, data.LoanTransactionID)
	fmt.Println("Line 39: Retrieved loan transaction, err:", err)
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get loan transaction")
	}

	fmt.Println("Line 48: Getting account from loan transaction")
	account, err := e.model.AccountManager.GetByID(ctx, *loanTransaction.AccountID)
	fmt.Println("Line 50: Retrieved account, err:", err)
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get cash on cash equivalence parent account (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get cash on cash equivalence parent account")
	}

	fmt.Println("Line 59: Finding loan transaction entries")
	loanTransactionEntries, err := e.model.LoanTransactionEntryManager.Find(ctx, &model.LoanTransactionEntry{
		ID:             loanTransaction.ID,
		OrganizationID: userOrg.OrganizationID,
		BranchID:       loanTransaction.BranchID,
	})
	fmt.Println("Line 65: Found", len(loanTransactionEntries), "loan transaction entries, err:", err)
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction entries (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get loan transaction entries")
	}

	fmt.Println("Line 74: Finding automatic loan deductions")
	automaticLoanDeductions, err := e.model.AutomaticLoanDeductionManager.Find(ctx, &model.AutomaticLoanDeduction{
		OrganizationID:     userOrg.OrganizationID,
		BranchID:           *userOrg.BranchID,
		ComputationSheetID: account.ComputationSheetID,
	})
	fmt.Println("Line 80: Found", len(automaticLoanDeductions), "automatic loan deductions, err:", err)
	if err != nil {
		automaticLoanDeductions = []*model.AutomaticLoanDeduction{}
	}

	fmt.Println("Line 85: Processing loan transaction entries")
	result := []*model.LoanTransactionEntry{}
	static, addOn, deduction, postComputed := []*model.LoanTransactionEntry{}, []*model.LoanTransactionEntry{}, []*model.LoanTransactionEntry{}, []*model.LoanTransactionEntry{}
	for _, entry := range loanTransactionEntries {
		fmt.Println("Line 89: Processing entry type:", entry.Type)
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
	fmt.Printf("Line 102: Categorized entries - static:%d, addOn:%d, deduction:%d, postComputed:%d\n", len(static), len(addOn), len(deduction), len(postComputed))

	fmt.Println("Line 104: Checking static entries count:", len(static))
	if len(static) < 2 {
		fmt.Println("Line 106: Creating default static entries")
		static = []*model.LoanTransactionEntry{
			{
				Credit:      loanTransaction.Applied1,
				Debit:       0,
				Description: cashOnCashEquivalenceAccount.Description,
				Name:        cashOnCashEquivalenceAccount.Name,
				Type:        model.LoanTransactionStatic,
			},
			{
				Credit:      0,
				Debit:       loanTransaction.Applied1,
				Description: loanTransaction.Account.Description,
				Name:        loanTransaction.Account.Name,
				Type:        model.LoanTransactionStatic,
			},
		}
	}
	fmt.Println("Line 123: Ordering static entries based on CashAndCashEquivalence:", static[0].Account.CashAndCashEquivalence)
	if static[0].Account.CashAndCashEquivalence {
		result = append(result, static[0])
		result = append(result, static[1])
	} else {
		result = append(result, static[1])
		result = append(result, static[0])
	}
	fmt.Println("Line 131: Checking addOn entries count:", len(addOn))
	if len(addOn) > 1 {
		fmt.Println("Line 133: ERROR - Too many addOn entries:", len(addOn))
		return nil, eris.New("only 1 add on entry is allowed")
	}
	// Deductions
	addOnEntry := &model.LoanTransactionEntry{
		Account: nil,
		Credit:  0,
		Debit:   0,
		Name:    "ADD ON INTEREST",
		Type:    model.LoanTransactionAddOn,
		IsAddOn: true,
	}
	fmt.Println("Line 143: Processing deductions, count:", len(deduction))
	total_non_add_ons, total_add_ons := 0.0, 0.0
	for _, entry := range deduction {
		fmt.Printf("Line 146: Deduction entry - IsAddOn:%v, Credit:%f\n", entry.IsAddOn, entry.Credit)
		if !entry.IsAddOn {
			total_non_add_ons += entry.Credit
		} else {
			total_add_ons += entry.Credit
		}
	}
	fmt.Printf("Line 154: After deductions - total_non_add_ons:%f, total_add_ons:%f\n", total_non_add_ons, total_add_ons)

	// Post Computed
	fmt.Println("Line 157: Processing post computed entries, count:", len(postComputed))
	for _, entry := range postComputed {
		fmt.Printf("Line 159: Post computed entry - ID:%v\n", entry.ID)
		entry.Credit = e.service.LoanComputation(ctx, *entry.AutomaticLoanDeduction, *loanTransaction)
		fmt.Printf("Line 161: Computed credit:%f for entry, IsAddOn:%v\n", entry.Credit, entry.IsAddOn)
		if !entry.IsAddOn {
			total_non_add_ons += entry.Credit
		} else {
			total_add_ons += entry.Credit
		}
		result = append(result, entry)
	}
	fmt.Printf("Line 170: After post computed - total_non_add_ons:%f, total_add_ons:%f\n", total_non_add_ons, total_add_ons)

	// Pre computed
	fmt.Println("Line 173: Processing automatic loan deductions (pre computed), count:", len(automaticLoanDeductions))
	for _, ald := range automaticLoanDeductions {
		fmt.Printf("Line 175: Processing ALD - ID:%v, Name:%s\n", ald.ID, ald.Name)
		exist := false
		for _, computed := range postComputed {
			if ald.ID == *computed.AutomaticLoanDeductionID {
				exist = true
				break
			}
		}
		if !exist {
			fmt.Printf("Line 183: Creating new entry for ALD:%s\n", ald.Name)
			entry := &model.LoanTransactionEntry{
				Credit:  0,
				Debit:   0,
				Name:    ald.Name,
				Type:    model.LoanTransactionAutomaticDeduction,
				IsAddOn: ald.AddOn,
				Account: ald.Account,
			}
			entry.Credit = e.service.LoanComputation(ctx, *ald, *loanTransaction)
			fmt.Printf("Line 192: Pre computed credit:%f for ALD:%s, IsAddOn:%v\n", entry.Credit, ald.Name, entry.IsAddOn)
			if !entry.IsAddOn {
				total_non_add_ons += entry.Credit
			} else {
				total_add_ons += entry.Credit
			}
			result = append(result, entry)
		}
	}
	fmt.Printf("Line 202: After pre computed - total_non_add_ons:%f, total_add_ons:%f\n", total_non_add_ons, total_add_ons)

	fmt.Printf("Line 204: Calculating result[0].Credit - IsAddOn:%v, Applied1:%f\n", loanTransaction.IsAddOn, loanTransaction.Applied1)
	if loanTransaction.IsAddOn {
		result[0].Credit = loanTransaction.Applied1 - total_non_add_ons
		fmt.Printf("Line 207: Set result[0].Credit (AddOn case):%f\n", result[0].Credit)
	} else {
		result[0].Credit = loanTransaction.Applied1 - (total_non_add_ons + total_add_ons)
		fmt.Printf("Line 210: Set result[0].Credit (non-AddOn case):%f\n", result[0].Credit)
	}

	if loanTransaction.IsAddOn {
		addOnEntry.Debit = total_add_ons
		fmt.Printf("Line 214: Adding addOnEntry with Debit:%f\n", addOnEntry.Debit)
		result = append(result, addOnEntry)
	}

	if loanTransaction.IsAddOn {
		addOnEntry.Debit = total_add_ons
		fmt.Printf("Line 220: Adding duplicate addOnEntry with Debit:%f\n", addOnEntry.Debit)
		result = append(result, addOnEntry)
	}

	fmt.Println("Line 224: Deleting existing loan transaction entries, count:", len(loanTransactionEntries))
	for _, entry := range loanTransactionEntries {
		fmt.Printf("Line 226: Deleting entry ID:%v\n", entry.ID)
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
	fmt.Println("Line 235: Creating new loan transaction entries, count:", len(result))
	for _, entry := range result {
		fmt.Printf("Line 237: Creating entry - Name:%s, Credit:%f, Debit:%f\n", entry.Name, entry.Credit, entry.Debit)
		if err := e.model.LoanTransactionEntryManager.CreateWithTx(ctx, tx, entry); err != nil {
			tx.Rollback()
			e.Footstep(ctx, echoCtx, FootstepEvent{
				Activity:    "data-error",
				Description: "Failed to create loan transaction entry (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return nil, eris.Wrap(err, "failed to create loan transaction entry + "+err.Error())
		}
	}
	fmt.Println("Line 246: Committing transaction")
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
	fmt.Println("Line 259: Function completed successfully")
	return newLoanTransaction, nil
}
