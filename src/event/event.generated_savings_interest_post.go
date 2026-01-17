package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

func GenerateSavingsInterestEntriesPost(
	context context.Context, service *horizon.HorizonService,
	userOrg *types.UserOrganization,
	generateSavingsInterestID *uuid.UUID,
	request types.GenerateSavingsInterestPostRequest,
) error {
	tx, endTx := service.Database.StartTransaction(context)

	generateSavingsInterest, err := core.GeneratedSavingsInterestManager(service).GetByID(context, *generateSavingsInterestID)
	if err != nil {
		return endTx(eris.Wrap(err, "failed to get generated savings interest"))
	}

	generatedSavingEntry, err := core.GeneratedSavingsInterestEntryManager(service).Find(context, &types.GeneratedSavingsInterestEntry{
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

	totalTax := decimal.Zero
	totalInterest := decimal.Zero

	for _, entry := range generatedSavingEntry {
		var credit, debit decimal.Decimal

		interestAmount := decimal.NewFromFloat(entry.InterestAmount)
		interestTax := decimal.NewFromFloat(entry.InterestTax)

		if interestAmount.GreaterThan(decimal.Zero) {
			credit = interestAmount.Sub(interestTax)
			debit = decimal.Zero
		} else {
			credit = decimal.Zero
			debit = interestAmount.Abs()
		}

		newGeneralLedger := &types.GeneralLedger{
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
			Source:                     types.GeneralLedgerSourceSavingsInterest,
			EmployeeUserID:             &userOrg.UserID,
			Description:                entry.Account.Description + " - Generated in Savings Interest",
			TypeOfPaymentType:          types.PaymentTypeSystem,
			Credit:                     credit.InexactFloat64(),
			Debit:                      debit.InexactFloat64(),
			CurrencyID:                 entry.Account.CurrencyID,
			Account:                    entry.Account,
		}

		if err := core.CreateGeneralLedgerEntry(context, service, tx, newGeneralLedger); err != nil {
			return endTx(eris.Wrap(err, "failed to create general ledger entry"))
		}

		if generateSavingsInterest.PostAccountID != nil {
			newGeneralLedger.Credit = debit.InexactFloat64()
			newGeneralLedger.Debit = credit.InexactFloat64()
			newGeneralLedger.AccountID = generateSavingsInterest.PostAccountID
			newGeneralLedger.Account = generateSavingsInterest.PostAccount

			if err := core.CreateGeneralLedgerEntry(context, service, tx, newGeneralLedger); err != nil {
				return endTx(eris.Wrap(err, "failed to create general ledger entry"))
			}
		}

		totalTax = totalTax.Add(interestTax)
		totalInterest = totalInterest.Add(interestAmount)
	}

	nowPtr := now
	generateSavingsInterest.PostedDate = &nowPtr
	generateSavingsInterest.PostedByUserID = &userOrg.UserID
	generateSavingsInterest.PostAccountID = request.PostAccountID
	generateSavingsInterest.TotalInterest = totalInterest.InexactFloat64()
	generateSavingsInterest.TotalTax = totalTax.InexactFloat64()

	if err := core.GeneratedSavingsInterestManager(service).UpdateByIDWithTx(context, tx, generateSavingsInterest.ID, generateSavingsInterest); err != nil {
		return endTx(eris.Wrap(err, "failed to update generated savings interest"))
	}

	if err := endTx(nil); err != nil {
		return endTx(eris.Wrap(err, "failed to commit transaction"))
	}

	return nil
}
