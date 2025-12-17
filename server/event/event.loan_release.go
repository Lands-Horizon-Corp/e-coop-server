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
)

func (e *Event) LoanRelease(context context.Context, ctx echo.Context, loanTransactionID uuid.UUID) (*core.LoanTransaction, error) {
	fmt.Println("DEBUG 0: Entering LoanRelease - loanTransactionID:", loanTransactionID)

	tx, endTx := e.provider.Service.Database.StartTransaction(context)
	fmt.Println("DEBUG 1: Transaction started")

	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(context, ctx)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "authentication-failed",
			Description: "Unable to retrieve user organization details for loan release operation: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get user organization"))
	}
	fmt.Println("DEBUG 2: Got userOrg - UserID:", userOrg.UserID, "BranchID nil?", userOrg.BranchID == nil)

	now := time.Now().UTC()
	currentTime := userOrg.UserOrgTime()
	fmt.Println("DEBUG 3: Calculated currentTime")

	if userOrg.BranchID == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "validation-failed",
			Description: "User organization is missing required branch assignment for loan operations",
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("invalid user organization data"))
	}
	fmt.Println("DEBUG 4: BranchID validated")

	loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, loanTransactionID, "Account", "Account.Currency")
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "loan-data-retrieval-failed",
			Description: "Unable to retrieve loan transaction details for release: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to get loan transaction"))
	}
	fmt.Println("DEBUG 5: Got loanTransaction - ID:", loanTransaction.ID, "Account nil?", loanTransaction.Account == nil)

	loanAccountCurrency := loanTransaction.Account.Currency
	fmt.Println("DEBUG 6: Accessed Account.Currency - nil?", loanAccountCurrency == nil)

	if loanAccountCurrency == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "currency-validation-failed",
			Description: "Missing currency information for loan account " + loanTransaction.AccountID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("currency data is nil"))
	}
	fmt.Println("DEBUG 7: Currency validated - CurrencyID:", loanAccountCurrency.ID)

	transactionBatch, err := e.core.TransactionBatchCurrent(context, *loanTransaction.EmployeeUserID, userOrg.OrganizationID, *userOrg.BranchID)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "batch-retrieval-failed",
			Description: "Unable to retrieve active transaction batch for user " + userOrg.UserID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve transaction batch - The one who created the loan must have created the transaction batch"))
	}
	fmt.Println("DEBUG 8: Got transactionBatch - nil?", transactionBatch == nil)

	if transactionBatch == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "batch-validation-failed",
			Description: "No active transaction batch found - batch is required for loan release operations",
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("transaction batch is nil"))
	}
	fmt.Println("DEBUG 9: Transaction batch validated")

	memberProfile, err := e.core.MemberProfileManager.GetByID(context, *loanTransaction.MemberProfileID)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "member-profile-retrieval-failed",
			Description: "Unable to retrieve member profile " + loanTransaction.MemberProfileID.String() + ": " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to retrieve member profile"))
	}
	fmt.Println("DEBUG 10: Got memberProfile - nil?", memberProfile == nil)

	if memberProfile == nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "member-profile-not-found",
			Description: "Member profile does not exist for ID: " + loanTransaction.MemberProfileID.String(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.New("member profile not found"))
	}

	loanTransactionEntries, err := e.core.LoanTransactionEntryManager.Find(context, &core.LoanTransactionEntry{
		LoanTransactionID: loanTransaction.ID,
		OrganizationID:    loanTransaction.OrganizationID,
		BranchID:          loanTransaction.BranchID,
	})
	if err != nil {
		return nil, endTx(eris.Wrap(err, "failed to retrieve loan transaction entries"))
	}
	fmt.Println("DEBUG 11: Retrieved loanTransactionEntries - count:", len(loanTransactionEntries))

	var addOnEntry *core.LoanTransactionEntry
	var filteredEntries []*core.LoanTransactionEntry

	fmt.Println("DEBUG 12.0: Starting to classify entries - addOnEntry is currently nil")

	for i, entry := range loanTransactionEntries {
		fmt.Printf("DEBUG 12.%d: Processing entry - Type: %v, AccountID: %v\n", i, entry.Type, entry.AccountID)

		if entry.Type == core.LoanTransactionAddOn {
			addOnEntry = entry
			fmt.Printf("DEBUG 12.ADDON: Found add-on entry! Debit = %.2f\n", entry.Debit)
		} else {
			filteredEntries = append(filteredEntries, entry)
			fmt.Printf("DEBUG 12.REGULAR: Added regular entry (Type: %v)\n", entry.Type)
		}
	}

	fmt.Printf("DEBUG 12.1: Classification done. addOnEntry is nil? %v\n", addOnEntry == nil)
	if addOnEntry == nil {
		fmt.Println("DEBUG 12.2: WARNING - No add-on entry found! Will use 0 for add-on debit")
	}

	for i, entry := range filteredEntries {
		fmt.Printf("DEBUG 13.%d: Checking entry for add-on adjustment - Type: %v, AccountID: %v\n",
			i, entry.Type, entry.AccountID)

		if entry.Type == core.LoanTransactionStatic && handlers.UUIDPtrEqual(entry.AccountID, loanTransaction.AccountID) {
			fmt.Printf("DEBUG 13.%d.MATCH: Found matching static entry! Current Debit = %.2f\n", i, entry.Debit)

			if addOnEntry == nil {
				fmt.Println("DEBUG 13.ERROR: addOnEntry is nil - cannot add debit!")
				// You can decide what to do: either skip, set to 0, or panic explicitly
				// For now, let's just log and continue safely
				// entry.Debit += 0
			} else {
				entry.Debit += addOnEntry.Debit
				fmt.Printf("DEBUG 13.%d.MATCH: Added add-on debit (%.2f) → New Debit = %.2f\n",
					i, addOnEntry.Debit, entry.Debit)
			}
		}
	}

	loanTransactionEntries = filteredEntries
	fmt.Println("DEBUG 13.FINAL: Processed add-on entries - filteredEntries count:", len(filteredEntries))

	for i, entry := range loanTransactionEntries {
		fmt.Printf("DEBUG 13.%d: Processing entry - IsAutomaticLoanDeductionDeleted: %v, AccountID nil? %v\n", i, entry.IsAutomaticLoanDeductionDeleted, entry.AccountID == nil)

		if entry.IsAutomaticLoanDeductionDeleted {
			continue
		}

		if entry.AccountID == nil {
			fmt.Println("DEBUG 13.ERROR: entry.AccountID is nil - skipping or failing?")
			return nil, endTx(eris.New("entry.AccountID is nil"))
		}

		accountHistory, err := e.core.GetAccountHistoryLatestByTimeHistory(
			context,
			*entry.AccountID,
			entry.OrganizationID,
			entry.BranchID,
			loanTransaction.ReleasedDate,
		)
		if err != nil {
			return nil, endTx(eris.Wrap(err, "failed to retrieve account history"))
		}
		fmt.Printf("DEBUG 14.%d: Got accountHistory - nil? %v\n", i, accountHistory == nil)

		if accountHistory == nil {
			return nil, endTx(eris.New("account history not found for entry"))
		}

		account := e.core.AccountHistoryToModel(accountHistory)
		fmt.Printf("DEBUG 15.%d: Converted account - DefaultPaymentTypeID nil? %v\n", i, account.DefaultPaymentTypeID == nil)

		if account.DefaultPaymentType == nil && account.DefaultPaymentTypeID != nil {
			paymentType, err := e.core.PaymentTypeManager.GetByID(context, *account.DefaultPaymentTypeID)
			if err != nil {
				return nil, endTx(eris.Wrap(err, "failed to retrieve payment type"))
			}
			account.DefaultPaymentType = paymentType
			fmt.Printf("DEBUG 16.%d: Loaded DefaultPaymentType\n", i)
		}

		var typeOfPaymentType core.TypeOfPaymentType
		if account.DefaultPaymentType != nil {
			typeOfPaymentType = account.DefaultPaymentType.Type
		}
		fmt.Printf("DEBUG 17.%d: Determined typeOfPaymentType\n", i)

		memberLedgerEntry := &core.GeneralLedger{
			CreatedAt:                  now,
			CreatedByID:                userOrg.UserID,
			UpdatedAt:                  now,
			UpdatedByID:                userOrg.UserID,
			BranchID:                   *userOrg.BranchID,
			OrganizationID:             userOrg.OrganizationID,
			TransactionBatchID:         &transactionBatch.ID,
			ReferenceNumber:            loanTransaction.Voucher,
			EntryDate:                  currentTime,
			AccountID:                  &account.ID,
			MemberProfileID:            &memberProfile.ID,
			PaymentTypeID:              account.DefaultPaymentTypeID,
			TransactionReferenceNumber: loanTransaction.Voucher,
			Source:                     core.GeneralLedgerSourceLoan,
			EmployeeUserID:             &userOrg.UserID,
			Description:                loanTransaction.Account.Description,
			TypeOfPaymentType:          typeOfPaymentType,
			Credit:                     entry.Credit,
			Debit:                      entry.Debit,
			CurrencyID:                 &loanAccountCurrency.ID,
			LoanTransactionID:          &loanTransaction.ID,
			Account:                    account,
		}

		fmt.Printf("DEBUG 18.%d: About to create GeneralLedger entry for AccountID: %s\n", i, account.ID)

		if err := e.core.CreateGeneralLedgerEntry(context, tx, memberLedgerEntry); err != nil {
			e.Footstep(ctx, FootstepEvent{
				Activity:    "member-ledger-creation-failed",
				Description: "Failed to create member ledger entry for account " + account.ID.String() + " and member " + memberProfile.ID.String() + ": " + err.Error(),
				Module:      "Loan Release",
			})
			return nil, endTx(eris.Wrap(err, "failed to create member ledger entry"))
		}
		fmt.Printf("DEBUG 19.%d: Successfully created ledger entry\n", i)
	}

	fmt.Println("DEBUG 20: Starting loan-related accounts processing")

	loanRelatedAccounts, err := e.core.GetAccountHistoriesByFiltersAtTime(
		context,
		loanTransaction.OrganizationID,
		loanTransaction.BranchID,
		&currentTime,
		loanTransaction.AccountID,
		&loanAccountCurrency.ID,
	)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "account-retrieval-failed",
			Description: "Failed to retrieve loan-related accounts for interest processing: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrapf(err, "failed to retrieve accounts for loan transaction ID: %s", loanTransaction.ID.String()))
	}
	fmt.Println("DEBUG 21: Got loanRelatedAccounts - count:", len(loanRelatedAccounts))

	loanRelatedAccounts = append(loanRelatedAccounts, loanTransaction.Account)
	fmt.Println("DEBUG 22: Appended main loan account")

	for i, interestAccount := range loanRelatedAccounts {
		fmt.Printf("DEBUG 23.%d: Processing interestAccount ID: %s\n", i, interestAccount.ID)

		interestAccountHistory, err := e.core.GetAccountHistoryLatestByTimeHistory(
			context,
			interestAccount.ID,
			interestAccount.OrganizationID,
			interestAccount.BranchID,
			loanTransaction.ReleasedDate,
		)
		if err != nil {
			return nil, endTx(eris.Wrap(err, "failed to retrieve interest account history"))
		}
		fmt.Printf("DEBUG 24.%d: Got interestAccountHistory - nil? %v\n", i, interestAccountHistory == nil)

		if interestAccountHistory == nil {
			return nil, endTx(eris.New("interest account history is nil"))
		}

		if err := e.core.LoanAccountManager.CreateWithTx(context, tx, &core.LoanAccount{
			CreatedAt:         now,
			CreatedByID:       userOrg.UserID,
			UpdatedAt:         now,
			UpdatedByID:       userOrg.UserID,
			OrganizationID:    interestAccount.OrganizationID,
			BranchID:          interestAccount.BranchID,
			LoanTransactionID: loanTransaction.ID,
			AccountID:         &interestAccount.ID,
			AccountHistoryID:  &interestAccountHistory.ID,
			Amount:            0.0,
		}); err != nil {
			return nil, endTx(eris.Wrap(err, "failed to create loan account"))
		}
		fmt.Printf("DEBUG 25.%d: Created LoanAccount entry\n", i)
	}

	fmt.Println("DEBUG 26: Updating loanTransaction fields")

	loanTransaction.ReleasedDate = &currentTime
	loanTransaction.ReleasedByID = &userOrg.UserID
	loanTransaction.UpdatedAt = now
	loanTransaction.Count++
	loanTransaction.TransactionBatchID = &transactionBatch.ID
	loanTransaction.UpdatedByID = userOrg.UserID

	if err := e.core.LoanTransactionManager.UpdateByIDWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
		return nil, endTx(eris.Wrap(err, "failed to update loan transaction"))
	}
	fmt.Println("DEBUG 27: Updated loanTransaction")

	if err := endTx(nil); err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "database-commit-failed",
			Description: "Unable to commit loan release transaction to database: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, endTx(eris.Wrap(err, "failed to commit transaction"))
	}
	fmt.Println("DEBUG 28: Transaction committed successfully")

	updatedloanTransaction, err := e.core.LoanTransactionManager.GetByID(context, loanTransaction.ID)
	if err != nil {
		e.Footstep(ctx, FootstepEvent{
			Activity:    "final-retrieval-failed",
			Description: "Unable to retrieve updated loan transaction after successful release: " + err.Error(),
			Module:      "Loan Release",
		})
		return nil, eris.Wrap(err, "failed to get updated loan transaction")
	}
	fmt.Println("DEBUG 29: Retrieved final updated loan transaction")

	e.Footstep(ctx, FootstepEvent{
		Activity:    "loan-release-completed",
		Description: "Successfully completed loan release for transaction " + updatedloanTransaction.ID.String(),
		Module:      "Loan Release",
	})

	fmt.Println("DEBUG 30: Loan release completed successfully - returning")

	return updatedloanTransaction, nil
}
