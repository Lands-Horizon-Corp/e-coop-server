package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (e *Event) GenerateMutualFundEntriesPost(
	context context.Context,
	userOrg *core.UserOrganization,
	mutualFundID *uuid.UUID,
	request core.MutualFundViewPostRequest,
) error {
	tx, endTx := e.provider.Service.Database.StartTransaction(context)
	mutualFund, err := e.core.MutualFundManager.GetByID(context, *mutualFundID)
	if err != nil {
		return endTx(err)
	}
	mutualFundEntries, err := e.core.MutualFundEntryManager.Find(context, &core.MutualFundEntry{
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
		MutualFundID:   mutualFund.ID,
	}, "MemberProfile", "Account.Currency")
	if err != nil {
		return endTx(err)
	}
	now := time.Now().UTC()
	userOrgTime := userOrg.UserOrgTime()
	if request.EntryDate != nil {
		userOrgTime = *request.EntryDate
	}
	totalAmount := 0.0
	for _, entry := range mutualFundEntries {
		var credit, debit float64
		if e.provider.Service.Decimal.IsGreaterThan(entry.Amount, 0) {
			credit = e.provider.Service.Decimal.Subtract(entry.Amount, entry.Amount)
			debit = 0
		} else {
			credit = 0
			debit = e.provider.Service.Decimal.Abs(entry.Amount)
		}
		newGeneralLedger := &core.GeneralLedger{
			CreatedAt:                  now,
			CreatedByID:                userOrg.UserID,
			UpdatedAt:                  now,
			UpdatedByID:                userOrg.UserID,
			BranchID:                   *userOrg.BranchID,
			OrganizationID:             userOrg.OrganizationID,
			ReferenceNumber:            *request.CheckVoucherNumber,
			EntryDate:                  userOrgTime,
			AccountID:                  &entry.AccountID,
			MemberProfileID:            &entry.MemberProfileID,
			TransactionReferenceNumber: *request.CheckVoucherNumber,
			Source:                     core.GeneralLedgerSourceMutualContribution,
			EmployeeUserID:             &userOrg.UserID,
			Description:                entry.Account.Description + " - Generated in mutual fund post",
			TypeOfPaymentType:          core.PaymentTypeSystem,
			Credit:                     credit,
			Debit:                      debit,
			CurrencyID:                 entry.Account.CurrencyID,

			Account: entry.Account,
		}
		if err := e.core.CreateGeneralLedgerEntry(context, tx, newGeneralLedger); err != nil {
			return endTx(eris.Wrap(err, "failed to create general ledger entry"))
		}
		if mutualFund.PostAccountID != nil {
			newGeneralLedger.Credit = debit
			newGeneralLedger.Debit = credit
			newGeneralLedger.AccountID = mutualFund.PostAccountID
			newGeneralLedger.Account = mutualFund.PostAccount
			if err := e.core.CreateGeneralLedgerEntry(context, tx, newGeneralLedger); err != nil {
				return endTx(eris.Wrap(err, "failed to create general ledger entry"))
			}
		}
		totalAmount = e.provider.Service.Decimal.Add(totalAmount, entry.Amount)

	}

	mutualFund.PostedDate = &now
	mutualFund.PostedByUserID = &userOrg.UserID
	mutualFund.TotalAmount = totalAmount
	mutualFund.PostAccountID = request.PostAccountID

	if err := e.core.MutualFundManager.UpdateByIDWithTx(context, tx, mutualFund.ID, mutualFund); err != nil {
		return endTx(eris.Wrap(err, "failed to update generated savings interest"))
	}
	if err := endTx(nil); err != nil {
		return endTx(eris.Wrap(err, "failed to commit transaction"))
	}
	return nil
}
