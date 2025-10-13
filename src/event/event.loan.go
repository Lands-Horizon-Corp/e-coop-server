package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
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
	// ================================================================================
	// STEP 1: AUTHENTICATION & USER ORGANIZATION RETRIEVAL
	// ================================================================================
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

	// ================================================================================
	// STEP 2: LOAN TRANSACTION & RELATED DATA RETRIEVAL
	// ================================================================================
	// Get the main loan transaction
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

	// Get the account associated with the loan transaction
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

	// Get existing loan transaction entries
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

	// Get automatic loan deductions for the computation sheet
	automaticLoanDeductions, err := e.model.AutomaticLoanDeductionManager.Find(ctx, &model.AutomaticLoanDeduction{
		OrganizationID:     userOrg.OrganizationID,
		BranchID:           *userOrg.BranchID,
		ComputationSheetID: account.ComputationSheetID,
	})
	disableLoanDeduction := loanTransaction.LoanType == model.LoanTypeRenewalWithoutDeduct || loanTransaction.LoanType == model.LoanTypeRestructured || loanTransaction.LoanType == model.LoanTypeStandardPrevious
	if err != nil || disableLoanDeduction {
		automaticLoanDeductions = []*model.AutomaticLoanDeduction{}
	}

	// ================================================================================
	// STEP 3: CATEGORIZE EXISTING LOAN TRANSACTION ENTRIES BY TYPE
	// ================================================================================
	result := []*model.LoanTransactionEntry{}
	static, addOn, deduction, postComputed := []*model.LoanTransactionEntry{}, []*model.LoanTransactionEntry{}, []*model.LoanTransactionEntry{}, []*model.LoanTransactionEntry{}

	// Categorize existing entries by their transaction type
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
		if entry.Type == model.LoanTransactionAutomaticDeduction && !disableLoanDeduction {
			postComputed = append(postComputed, entry)
		}
	}

	// ================================================================================
	// STEP 4: CREATE DEFAULT STATIC ENTRIES IF NOT EXISTS
	// ================================================================================
	// If we don't have the required 2 static entries, create them
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

		// Create the two required static entries: credit to cash equivalent and debit to loan account
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

	// ================================================================================
	// STEP 5: ARRANGE STATIC ENTRIES & CLEANUP EXISTING ADD-ON ENTRIES
	// ================================================================================
	// Order static entries: cash equivalent first, then loan account
	if static[0].Account.CashAndCashEquivalence {
		result = append(result, static[0])
		result = append(result, static[1])
	} else {
		result = append(result, static[1])
		result = append(result, static[0])
	}

	// Delete existing add-on entries (they will be recalculated)
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

	// ================================================================================
	// STEP 6: PREPARE ADD-ON INTEREST ENTRY TEMPLATE
	// ================================================================================
	addOnEntry := &model.LoanTransactionEntry{
		Account:           nil,
		Credit:            0,
		Debit:             0,
		Name:              "ADD ON INTEREST",
		Type:              model.LoanTransactionAddOn,
		LoanTransactionID: loanTransaction.ID,
		IsAddOn:           true,
	}

	// ================================================================================
	// STEP 7: PROCESS EXISTING DEDUCTIONS & CALCULATE TOTALS
	// ================================================================================
	total_non_add_ons, total_add_ons := 0.0, 0.0

	// Add existing deduction entries and calculate running totals
	for _, entry := range deduction {
		if !entry.IsAddOn {
			total_non_add_ons += entry.Credit
		} else {
			total_add_ons += entry.Credit
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
			entry.Credit = e.service.LoanComputation(ctx, *entry.AutomaticLoanDeduction, *loanTransaction)
		}

		if !entry.IsAddOn {
			total_non_add_ons += entry.Credit
		} else {
			total_add_ons += entry.Credit
		}
		result = append(result, entry)
	}

	// ================================================================================
	// STEP 8: ADD MISSING AUTOMATIC DEDUCTIONS
	// ================================================================================
	// Process automatic loan deductions that weren't already included in post-computed entries
	for _, ald := range automaticLoanDeductions {
		exist := false
		for _, computed := range postComputed {
			if handlers.UuidPtrEqual(&ald.ID, computed.AutomaticLoanDeductionID) {
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
				Amount:                   0,
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

	if (loanTransaction.LoanType == model.LoanTypeRestructured ||
		loanTransaction.LoanType == model.LoanTypeRenewalWithoutDeduct ||
		loanTransaction.LoanType == model.LoanTypeRenewal) && loanTransaction.PreviousLoanID != nil {
		previous := loanTransaction.PreviousLoan
		result = append(result, &model.LoanTransactionEntry{
			Account:           previous.Account,
			AccountID:         previous.AccountID,
			Credit:            previous.Balance,
			Debit:             0,
			Name:              previous.Account.Name,
			Description:       previous.Account.Description,
			Type:              model.LoanTransactionPrevious,
			LoanTransactionID: loanTransaction.ID,
		})
		total_non_add_ons += previous.Balance
	}

	// ================================================================================
	// STEP 9: CALCULATE FINAL CREDIT AMOUNTS & ADD-ON INTEREST
	// ================================================================================
	// Adjust the first entry (cash equivalent) credit based on loan type and deductions
	if loanTransaction.IsAddOn {
		result[0].Credit = loanTransaction.Applied1 - total_non_add_ons
	} else {
		result[0].Credit = loanTransaction.Applied1 - (total_non_add_ons + total_add_ons)
	}

	// Add the add-on interest entry if applicable
	if loanTransaction.IsAddOn && total_add_ons > 0 {
		addOnEntry.Debit = total_add_ons
		result = append(result, addOnEntry)
	}

	// ================================================================================
	// STEP 10: CLEANUP OLD ENTRIES & CREATE NEW BALANCED ENTRIES
	// ================================================================================
	// Delete all existing loan transaction entries before creating new ones
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

	// Set the debit amount for the loan account entry
	result[1].Debit = loanTransaction.Applied1
	switch loanTransaction.LoanType {

	case model.LoanTypeStandard:
		result[1].Name = loanTransaction.Account.Name
	case model.LoanTypeStandardPrevious:
		result[1].Name = loanTransaction.Account.Name
	case model.LoanTypeRestructured:
		result[1].Name = loanTransaction.Account.Name + " - RESTRUCTURED"
	case model.LoanTypeRenewal:
		result[1].Name = loanTransaction.Account.Name + " - CURRENT"
	case model.LoanTypeRenewalWithoutDeduct:
		result[1].Name = loanTransaction.Account.Name + " - CURRENT"

	}

	// Create new loan transaction entries and calculate totals
	totalDebit, totalCredit := 0.0, 0.0
	for index, entry := range result {

		value := &model.LoanTransactionEntry{
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
			totalDebit += entry.Debit
			totalCredit += entry.Credit
		}

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
	// Amortization
	amort, err := e.service.LoanModeOfPayment(ctx, loanTransaction)
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to calculate loan amortization (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to calculate loan amortization + "+err.Error())
	}

	// ================================================================================
	// STEP 11: UPDATE LOAN TRANSACTION TOTALS & COMMIT CHANGES
	// ================================================================================
	// Update the loan transaction with calculated totals
	loanTransaction.Amortization = amort
	loanTransaction.Balance = totalCredit
	loanTransaction.TotalCredit = totalCredit
	loanTransaction.TotalDebit = totalDebit
	loanTransaction.UpdatedAt = time.Now().UTC()
	loanTransaction.UpdatedByID = userOrg.UserID
	if err := e.model.LoanTransactionManager.UpdateFieldsWithTx(ctx, tx, loanTransaction.ID, loanTransaction); err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to update loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to update loan transaction + "+err.Error())
	}

	// Commit all database changes
	if err := tx.Commit().Error; err != nil {
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "db-commit-error",
			Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to commit transaction")
	}

	// ================================================================================
	// STEP 12: RETRIEVE & RETURN UPDATED LOAN TRANSACTION
	// ================================================================================
	// Get the updated loan transaction with all related data
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

func (e *Event) LoanPayment(ctx context.Context, echoCtx echo.Context, tx *gorm.DB, data LoanBalanceEvent) (*model.LoanTransaction, error) {
	// eneralLedger, err := c.event.TransactionPayment(context, ctx, tx, event.TransactionEvent{
	// 		// Will be filled by transaction
	// 		TransactionID:        nil,
	// 		MemberProfileID:      loanTransaction.MemberProfileID,
	// 		MemberJointAccountID: loanTransaction.MemberJointAccountID,
	// 		ReferenceNumber:      loanTransaction.Re,

	// 		// On Request
	// 		Source:                model.GeneralLedgerSourcePayment,
	// 		Amount:                req.Amount,
	// 		AccountID:             req.AccountID,
	// 		PaymentTypeID:         req.PaymentTypeID,
	// 		SignatureMediaID:      req.SignatureMediaID,
	// 		EntryDate:             req.EntryDate,
	// 		BankID:                req.BankID,
	// 		ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
	// 		Description:           req.Description,
	// 		BankReferenceNumber:   req.BankReferenceNumber,
	// 		ORAutoGenerated:       req.ORAutoGenerated,
	// 	})
	// Commit all database changes
	if err := tx.Commit().Error; err != nil {
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "db-commit-error",
			Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to commit transaction")
	}
	return nil, nil
}
