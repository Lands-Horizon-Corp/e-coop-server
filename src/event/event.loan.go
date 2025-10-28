package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type LoanBalanceEvent struct {
	CashOnCashEquivalenceAccountID uuid.UUID
	LoanTransactionID              uuid.UUID
}

func (e *Event) LoanBalancing(ctx context.Context, echoCtx echo.Context, tx *gorm.DB, data LoanBalanceEvent) (*model_core.LoanTransaction, error) {
	fmt.Printf("[LOAN BALANCING] Starting loan balancing process for LoanTransactionID: %s\n", data.LoanTransactionID.String())

	// ================================================================================
	// STEP 1: AUTHENTICATION & USER ORGANIZATION RETRIEVAL
	// ================================================================================
	fmt.Printf("[LOAN BALANCING] Step 1: Authenticating user and retrieving organization\n")
	userOrg, err := e.user_organization_token.CurrentUserOrganization(ctx, echoCtx)
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get user organization")
	}
	fmt.Printf("[LOAN BALANCING] User organization retrieved: OrganizationID=%s, BranchID=%v\n", userOrg.OrganizationID.String(), userOrg.BranchID)

	// ================================================================================
	// STEP 2: LOAN TRANSACTION & RELATED DATA RETRIEVAL
	// ================================================================================
	fmt.Printf("[LOAN BALANCING] Step 2: Retrieving loan transaction and related data\n")
	// Get the main loan transaction
	loanTransaction, err := e.model_core.LoanTransactionManager.GetByID(ctx, data.LoanTransactionID)
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get loan transaction")
	}
	fmt.Printf("[LOAN BALANCING] Loan transaction retrieved: ID=%s, Applied1=%f, LoanType=%s\n",
		loanTransaction.ID.String(), loanTransaction.Applied1, loanTransaction.LoanType)

	// Get the account associated with the loan transaction
	account, err := e.model_core.AccountManager.GetByID(ctx, *loanTransaction.AccountID)
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
	loanTransactionEntries, err := e.model_core.LoanTransactionEntryManager.Find(ctx, &model_core.LoanTransactionEntry{
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
	automaticLoanDeductions, err := e.model_core.AutomaticLoanDeductionManager.Find(ctx, &model_core.AutomaticLoanDeduction{
		OrganizationID:     userOrg.OrganizationID,
		BranchID:           *userOrg.BranchID,
		ComputationSheetID: account.ComputationSheetID,
	})
	disableLoanDeduction := loanTransaction.LoanType == model_core.LoanTypeRenewalWithoutDeduct || loanTransaction.LoanType == model_core.LoanTypeRestructured || loanTransaction.LoanType == model_core.LoanTypeStandardPrevious
	if err != nil || disableLoanDeduction {
		automaticLoanDeductions = []*model_core.AutomaticLoanDeduction{}
	}

	// ================================================================================
	// STEP 3: CATEGORIZE EXISTING LOAN TRANSACTION ENTRIES BY TYPE
	// ================================================================================
	fmt.Printf("[LOAN BALANCING] Step 3: Categorizing existing loan transaction entries\n")
	fmt.Printf("[LOAN BALANCING] Found %d existing loan transaction entries\n", len(loanTransactionEntries))
	result := []*model_core.LoanTransactionEntry{}
	static, addOn, deduction, postComputed := []*model_core.LoanTransactionEntry{}, []*model_core.LoanTransactionEntry{}, []*model_core.LoanTransactionEntry{}, []*model_core.LoanTransactionEntry{}

	// Categorize existing entries by their transaction type
	for _, entry := range loanTransactionEntries {
		if entry.Type == model_core.LoanTransactionStatic {
			static = append(static, entry)
		}
		if entry.Type == model_core.LoanTransactionAddOn {
			addOn = append(addOn, entry)
		}
		if entry.Type == model_core.LoanTransactionDeduction {
			deduction = append(deduction, entry)
		}
		if entry.Type == model_core.LoanTransactionAutomaticDeduction && !disableLoanDeduction {
			postComputed = append(postComputed, entry)
		}
	}
	fmt.Printf("[LOAN BALANCING] Categorized entries - Static: %d, AddOn: %d, Deduction: %d, PostComputed: %d\n",
		len(static), len(addOn), len(deduction), len(postComputed))

	// ================================================================================
	// STEP 4: CREATE DEFAULT STATIC ENTRIES IF NOT EXISTS
	// ================================================================================
	fmt.Printf("[LOAN BALANCING] Step 4: Creating default static entries if needed\n")
	// If we don't have the required 2 static entries, create them
	if len(static) < 2 {
		cashOnCashEquivalenceAccount, err := e.model_core.AccountManager.GetByID(ctx, data.CashOnCashEquivalenceAccountID)
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
		static = []*model_core.LoanTransactionEntry{
			{
				Credit:            loanTransaction.Applied1,
				Debit:             0,
				Description:       cashOnCashEquivalenceAccount.Description,
				Account:           cashOnCashEquivalenceAccount,
				AccountID:         &cashOnCashEquivalenceAccount.ID,
				Name:              cashOnCashEquivalenceAccount.Name,
				Type:              model_core.LoanTransactionStatic,
				LoanTransactionID: loanTransaction.ID,
			},
			{
				Credit:            0,
				Debit:             loanTransaction.Applied1,
				Account:           loanTransaction.Account,
				AccountID:         loanTransaction.AccountID,
				Description:       loanTransaction.Account.Description,
				Name:              loanTransaction.Account.Name,
				Type:              model_core.LoanTransactionStatic,
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
		if err := e.model_core.LoanTransactionEntryManager.DeleteByIDWithTx(ctx, tx, entry.ID); err != nil {
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
	addOnEntry := &model_core.LoanTransactionEntry{
		Account:           nil,
		Credit:            0,
		Debit:             0,
		Name:              "ADD ON INTEREST",
		Type:              model_core.LoanTransactionAddOn,
		LoanTransactionID: loanTransaction.ID,
		IsAddOn:           true,
	}

	// ================================================================================
	// STEP 7: PROCESS EXISTING DEDUCTIONS & CALCULATE TOTALS
	// ================================================================================
	fmt.Printf("[LOAN BALANCING] Step 7: Processing existing deductions and calculating totals\n")
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
			if entry.AutomaticLoanDeduction.ChargesRateSchemeID != nil {
				chargesRateScheme, err := e.model_core.ChargesRateSchemeManager.GetByID(ctx, *entry.AutomaticLoanDeduction.ChargesRateSchemeID)
				if err != nil {
					return nil, err
				}
				entry.Credit = e.service.LoanChargesRateComputation(ctx, *chargesRateScheme, *loanTransaction)
			}
			if entry.Credit != 0 {
				entry.Credit = e.service.LoanComputation(ctx, *entry.AutomaticLoanDeduction, *loanTransaction)
			}

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
			entry := &model_core.LoanTransactionEntry{
				Credit:                   0,
				Debit:                    0,
				Name:                     ald.Name,
				Type:                     model_core.LoanTransactionAutomaticDeduction,
				IsAddOn:                  ald.AddOn,
				Account:                  ald.Account,
				AccountID:                ald.AccountID,
				Description:              ald.Account.Description,
				AutomaticLoanDeductionID: &ald.ID,
				LoanTransactionID:        loanTransaction.ID,
				Amount:                   0,
			}
			if entry.AutomaticLoanDeduction.ChargesRateSchemeID != nil {
				chargesRateScheme, err := e.model_core.ChargesRateSchemeManager.GetByID(ctx, *entry.AutomaticLoanDeduction.ChargesRateSchemeID)
				if err != nil {
					return nil, err
				}
				entry.Credit = e.service.LoanChargesRateComputation(ctx, *chargesRateScheme, *loanTransaction)
			}
			if entry.Credit != 0 {
				entry.Credit = e.service.LoanComputation(ctx, *entry.AutomaticLoanDeduction, *loanTransaction)
			}
			if !entry.IsAddOn {
				total_non_add_ons += entry.Credit
			} else {
				total_add_ons += entry.Credit
			}
			result = append(result, entry)
		}
	}

	if (loanTransaction.LoanType == model_core.LoanTypeRestructured ||
		loanTransaction.LoanType == model_core.LoanTypeRenewalWithoutDeduct ||
		loanTransaction.LoanType == model_core.LoanTypeRenewal) && loanTransaction.PreviousLoanID != nil {
		previous := loanTransaction.PreviousLoan
		result = append(result, &model_core.LoanTransactionEntry{
			Account:           previous.Account,
			AccountID:         previous.AccountID,
			Credit:            previous.Balance,
			Debit:             0,
			Name:              previous.Account.Name,
			Description:       previous.Account.Description,
			Type:              model_core.LoanTransactionPrevious,
			LoanTransactionID: loanTransaction.ID,
		})
		total_non_add_ons += previous.Balance
	}

	// ================================================================================
	// STEP 9: CALCULATE FINAL CREDIT AMOUNTS & ADD-ON INTEREST
	// ================================================================================
	fmt.Printf("[LOAN BALANCING] Step 9: Calculating final credit amounts\n")
	fmt.Printf("[LOAN BALANCING] Total non-add-ons: %f, Total add-ons: %f\n", total_non_add_ons, total_add_ons)
	// Adjust the first entry (cash equivalent) credit based on loan type and deductions
	if loanTransaction.IsAddOn {
		result[0].Credit = loanTransaction.Applied1 - total_non_add_ons
		fmt.Printf("[LOAN BALANCING] IsAddOn=true: Cash equivalent credit = %f - %f = %f\n",
			loanTransaction.Applied1, total_non_add_ons, result[0].Credit)
	} else {
		result[0].Credit = loanTransaction.Applied1 - (total_non_add_ons + total_add_ons)
		fmt.Printf("[LOAN BALANCING] IsAddOn=false: Cash equivalent credit = %f - (%f + %f) = %f\n",
			loanTransaction.Applied1, total_non_add_ons, total_add_ons, result[0].Credit)
	}

	// Add the add-on interest entry if applicable
	if loanTransaction.IsAddOn && total_add_ons > 0 {
		addOnEntry.Debit = total_add_ons
		result = append(result, addOnEntry)
	}

	// ================================================================================
	// STEP 10: CLEANUP OLD ENTRIES & CREATE NEW BALANCED ENTRIES
	// ================================================================================
	fmt.Printf("[LOAN BALANCING] Step 10: Cleaning up old entries and creating new balanced entries\n")
	fmt.Printf("[LOAN BALANCING] Deleting %d existing entries\n", len(loanTransactionEntries))
	// Delete all existing loan transaction entries before creating new ones
	for _, entry := range loanTransactionEntries {
		if err := e.model_core.LoanTransactionEntryManager.DeleteByIDWithTx(ctx, tx, entry.ID); err != nil {
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

	case model_core.LoanTypeStandard:
		result[1].Name = loanTransaction.Account.Name
	case model_core.LoanTypeStandardPrevious:
		result[1].Name = loanTransaction.Account.Name
	case model_core.LoanTypeRestructured:
		result[1].Name = loanTransaction.Account.Name + " - RESTRUCTURED"
	case model_core.LoanTypeRenewal:
		result[1].Name = loanTransaction.Account.Name + " - CURRENT"
	case model_core.LoanTypeRenewalWithoutDeduct:
		result[1].Name = loanTransaction.Account.Name + " - CURRENT"

	}

	// Create new loan transaction entries and calculate totals
	fmt.Printf("[LOAN BALANCING] Creating %d new loan transaction entries\n", len(result))
	totalDebit, totalCredit := 0.0, 0.0
	for index, entry := range result {

		value := &model_core.LoanTransactionEntry{
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

		fmt.Printf("[LOAN BALANCING] Creating entry %d: %s - Credit: %f, Debit: %f\n",
			index, entry.Name, entry.Credit, entry.Debit)

		if err := e.model_core.LoanTransactionEntryManager.CreateWithTx(ctx, tx, value); err != nil {
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
	fmt.Printf("[LOAN BALANCING] Step 11: Updating loan transaction totals\n")
	fmt.Printf("[LOAN BALANCING] Final totals - Credit: %f, Debit: %f, Amortization: %f\n",
		totalCredit, totalDebit, amort)
	// Update the loan transaction with calculated totals
	loanTransaction.Amortization = amort
	loanTransaction.Balance = totalCredit
	loanTransaction.TotalCredit = totalCredit
	loanTransaction.TotalDebit = totalDebit
	loanTransaction.UpdatedAt = time.Now().UTC()
	loanTransaction.UpdatedByID = userOrg.UserID
	if err := e.model_core.LoanTransactionManager.UpdateFieldsWithTx(ctx, tx, loanTransaction.ID, loanTransaction); err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to update loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to update loan transaction + "+err.Error())
	}

	// Commit all database changes
	fmt.Printf("[LOAN BALANCING] Committing database transaction\n")
	if err := tx.Commit().Error; err != nil {
		fmt.Printf("[LOAN BALANCING] ERROR: Failed to commit transaction: %s\n", err.Error())
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "db-commit-error",
			Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to commit transaction")
	}
	fmt.Printf("[LOAN BALANCING] Transaction committed successfully\n")

	// ================================================================================
	// STEP 12: RETRIEVE & RETURN UPDATED LOAN TRANSACTION
	// ================================================================================
	fmt.Printf("[LOAN BALANCING] Step 12: Retrieving updated loan transaction\n")
	// Get the updated loan transaction with all related data
	newLoanTransaction, err := e.model_core.LoanTransactionManager.GetByID(ctx, loanTransaction.ID)
	if err != nil {
		fmt.Printf("[LOAN BALANCING] ERROR: Failed to get updated loan transaction: %s\n", err.Error())
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get updated loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get updated loan transaction")
	}

	fmt.Printf("[LOAN BALANCING] Loan balancing completed successfully for LoanTransactionID: %s\n",
		newLoanTransaction.ID.String())
	return newLoanTransaction, nil
}

func (e *Event) LoanRelease(ctx context.Context, echoCtx echo.Context, tx *gorm.DB, data LoanBalanceEvent) (*model_core.LoanTransaction, error) {
	// eneralLedger, err := c.event.TransactionPayment(context, ctx, tx, event.TransactionEvent{
	// 		// Will be filled by transaction
	// 		TransactionID:        nil,
	// 		MemberProfileID:      loanTransaction.MemberProfileID,
	// 		MemberJointAccountID: loanTransaction.MemberJointAccountID,
	// 		ReferenceNumber:      loanTransaction.Re,

	// 		// On Request
	// 		Source:                model_core.GeneralLedgerSourcePayment,
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
	// ================================================================================
	// STEP 1: AUTHENTICATION & USER ORGANIZATION RETRIEVAL
	// ================================================================================
	userOrg, err := e.user_organization_token.CurrentUserOrganization(ctx, echoCtx)
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get user organization")
	}
	if userOrg.BranchID == nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Invalid user organization data (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		return nil, eris.New("invalid user organization data")
	}
	if userOrg.UserType != model_core.UserOrganizationTypeOwner && userOrg.UserType != model_core.UserOrganizationTypeEmployee {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Unauthorized user role (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		return nil, eris.New("unauthorized user role")
	}
	// ================================================================================
	// STEP 2: LOAN TRANSACTION & RELATED DATA RETRIEVAL
	// ================================================================================
	// Get the main loan transaction
	loanTransaction, err := e.model_core.LoanTransactionManager.GetByID(ctx, data.LoanTransactionID)
	if err != nil {
		tx.Rollback()
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "data-error",
			Description: "Failed to get loan transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to get loan transaction")
	}

	for _, entry := range loanTransaction.LoanTransactionEntries {
		// Computation of all ammortization accounts
		if entry.Type == model_core.LoanTransactionPrevious {
			tx.Rollback()
			return nil, eris.New("cannot release a restructured or renewed loan")
		}
	}

	// ================================================================================
	if err := tx.Commit().Error; err != nil {
		e.Footstep(ctx, echoCtx, FootstepEvent{
			Activity:    "db-commit-error",
			Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to commit transaction")
	}
	newLoanTransaction, err := e.model_core.LoanTransactionManager.GetByID(ctx, loanTransaction.ID)
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
