package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (e *Event) GenerateSavingsInterestEntriesPost(
	context context.Context,
	userOrg *core.UserOrganization,
	generateSavingsInterestID *uuid.UUID,
	request core.GenerateSavingsInterestPostRequest,
) error {
	tx, endTx := e.provider.Service.Database.StartTransaction(context)
	generateSavingsInterest, err := e.core.GeneratedSavingsInterestManager.GetByID(context, *generateSavingsInterestID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to get generated savings interest"))
	}
	generatedSavinggEntry, err := e.core.GeneratedSavingsInterestEntryManager.Find(context, &core.GeneratedSavingsInterestEntry{
		OrganizationID:             userOrg.OrganizationID,
		BranchID:                   *userOrg.BranchID,
		GeneratedSavingsInterestID: *generateSavingsInterestID,
	}, "Account")
	if err != nil {
		return endTx(eris.Wrap(err, "failed to find generated savings interest entries"))
	}
	now := time.Now().UTC()
	userOrgTime := userOrg.UserOrgTime()
	if request.EntryDate != nil {
		userOrgTime = *request.EntryDate
	}
	for _, entry := range generatedSavinggEntry {

		var credit, debit float64
		if e.provider.Service.Decimal.IsGreaterThan(entry.InterestAmount, 0) {
			credit = e.provider.Service.Decimal.Subtract(entry.InterestAmount, entry.InterestTax)
			debit = 0
		} else {
			credit = 0
			debit = e.provider.Service.Decimal.Abs(entry.InterestAmount)
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
			Source:                     core.GeneralLedgerSourceSavingsInterest,
			EmployeeUserID:             &userOrg.UserID,
			Description:                entry.Account.Description + " - Generated in Savings Interest",
			TypeOfPaymentType:          core.PaymentTypeSystem,
			Credit:                     credit,
			Debit:                      debit,
			CurrencyID:                 entry.Account.CurrencyID,

			Account: entry.Account,
		}
		if err := e.core.CreateGeneralLedgerEntry(context, tx, newGeneralLedger); err != nil {
			return endTx(eris.Wrap(err, "failed to create general ledger entry"))
		}
	}
	generateSavingsInterest.PostedDate = &now
	generateSavingsInterest.PostedByUserID = &userOrg.UserID
	if err := e.core.GeneratedSavingsInterestManager.UpdateByIDWithTx(context, tx, generateSavingsInterest.ID, generateSavingsInterest); err != nil {
		return endTx(eris.Wrap(err, "failed to update generated savings interest"))
	}
	if err := endTx(nil); err != nil {
		return endTx(eris.Wrap(err, "failed to commit transaction"))
	}
	return nil
}
