package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

func GenerateMutualFundEntriesPost(
	context context.Context, service *horizon.HorizonService,
	userOrg *core.UserOrganization,
	mutualFundID *uuid.UUID,
	request core.MutualFundViewPostRequest,
) error {
	tx, endTx := service.Database.StartTransaction(context)

	mutualFund, err := core.MutualFundManager(service).GetByID(context, *mutualFundID, "Account", "Account.Currency")
	if err != nil {
		return endTx(err)
	}

	mutualFundEntries, err := core.MutualFundEntryManager(service).Find(context, &core.MutualFundEntry{
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
		MutualFundID:   mutualFund.ID,
	}, "MemberProfile", "Account", "Account.Currency")
	if err != nil {
		return endTx(err)
	}

	now := time.Now().UTC()
	userOrgTime := userOrg.UserOrgTime()
	if request.EntryDate != nil {
		userOrgTime = *request.EntryDate
	}

	totalAmountDec := decimal.Zero

	for _, entry := range mutualFundEntries {
		amountDec := decimal.NewFromFloat(entry.Amount)
		var creditDec, debitDec decimal.Decimal

		if amountDec.GreaterThan(decimal.Zero) {
			creditDec = decimal.Zero
			debitDec = amountDec.Abs()
		} else {
			creditDec = amountDec.Abs()
			debitDec = decimal.Zero
		}

		// Create member ledger entry
		if err := core.CreateGeneralLedgerEntry(context, service, tx, &core.GeneralLedger{
			CreatedAt:       now,
			CreatedByID:     userOrg.UserID,
			UpdatedAt:       now,
			UpdatedByID:     userOrg.UserID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
			ReferenceNumber: *request.CheckVoucherNumber,
			EntryDate:       userOrgTime,

			MemberProfileID:   &entry.MemberProfileID,
			Source:            core.GeneralLedgerSourceMutualContribution,
			EmployeeUserID:    &userOrg.UserID,
			Description:       entry.Account.Description + " - Generated in mutual fund post",
			TypeOfPaymentType: core.PaymentTypeSystem,
			Credit:            creditDec.InexactFloat64(),
			Debit:             debitDec.InexactFloat64(),

			CurrencyID: entry.Account.CurrencyID,
			AccountID:  &entry.AccountID,
			Account:    entry.Account,
		}); err != nil {
			return endTx(eris.Wrap(err, "failed to create general ledger entry - (member ledger)"))
		}

		// Create post account entry if applicable
		if mutualFund.PostAccountID != nil {
			if err := core.CreateGeneralLedgerEntry(context, service, tx, &core.GeneralLedger{
				CreatedAt:       now,
				CreatedByID:     userOrg.UserID,
				UpdatedAt:       now,
				UpdatedByID:     userOrg.UserID,
				BranchID:        *userOrg.BranchID,
				OrganizationID:  userOrg.OrganizationID,
				ReferenceNumber: *request.CheckVoucherNumber,
				EntryDate:       userOrgTime,

				MemberProfileID:   &entry.MemberProfileID,
				Source:            core.GeneralLedgerSourceMutualContribution,
				EmployeeUserID:    &userOrg.UserID,
				Description:       mutualFund.PostAccount.Description + " - Generated in mutual fund post",
				TypeOfPaymentType: core.PaymentTypeSystem,
				Credit:            debitDec.InexactFloat64(),
				Debit:             creditDec.InexactFloat64(),

				AccountID:  mutualFund.PostAccountID,
				CurrencyID: mutualFund.PostAccount.CurrencyID,
				Account:    mutualFund.PostAccount,
			}); err != nil {
				return endTx(eris.Wrap(err, "failed to create general ledger entry - (post account)"))
			}
		}

		// Accumulate total amount
		totalAmountDec = totalAmountDec.Add(amountDec)
	}

	mutualFund.PostedDate = &now
	mutualFund.PostedByUserID = &userOrg.UserID
	mutualFund.TotalAmount = totalAmountDec.InexactFloat64()
	mutualFund.PostAccountID = request.PostAccountID

	if err := core.MutualFundManager(service).UpdateByIDWithTx(context, tx, mutualFund.ID, mutualFund); err != nil {
		return endTx(eris.Wrap(err, "failed to update generated savings interest"))
	}

	if err := endTx(nil); err != nil {
		return endTx(eris.Wrap(err, "failed to commit transaction"))
	}

	return nil
}
