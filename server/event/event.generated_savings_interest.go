package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (e *Event) GenerateSavingsInterestEntries(
	context context.Context,
	userOrg *core.UserOrganization,
	generatedSavingsInterest core.GeneratedSavingsInterest,
	annualDivisor int,
) ([]*core.GeneratedSavingsInterestEntry, error) {

	result := []*core.GeneratedSavingsInterestEntry{}
	// Step 1: Get browse references based on the generated savings interest criteria
	browseReferences, err := e.core.BrowseReferenceByField(
		context,
		userOrg.OrganizationID,
		*userOrg.BranchID,
		generatedSavingsInterest.AccountID,
		generatedSavingsInterest.MemberTypeID,
	)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get browse references")
	}
	// Step 2: Get member accounting ledgers using the browse references
	memberBrowseReferences, err := e.core.MemberAccountingLedgerByBrowseReference(context, browseReferences)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get member accounting ledgers by browse reference")
	}
	for _, memberBrowseRef := range memberBrowseReferences {
		currency := memberBrowseRef.MemberAccountingLedger.Account.Currency
		loc, err := time.LoadLocation(currency.Timezone)
		if err != nil {
			return nil, eris.Wrap(err, "failed to load location")
		}
		from := generatedSavingsInterest.LastComputationDate.In(loc)
		to := generatedSavingsInterest.NewComputationDate.In(loc)

		// Step 3: Get daily ending balances for the member within the computation period
		dailyBalances, err := e.core.GetDailyEndingBalances(
			context,
			from, to,
			*memberBrowseRef.BrowseReference.AccountID,
			memberBrowseRef.MemberAccountingLedger.MemberProfileID,
			userOrg.OrganizationID, *userOrg.BranchID,
		)
		if err != nil {
			return nil, eris.Wrap(err, "failed to get daily ending balances")
		}
		if len(dailyBalances) == 0 {
			continue
		}
		// switch memberBrowseRef.BrowseReference.InterestType {
		// case core.InterestTypeYear:
		// case core.InterestTypeDate:
		// case core.InterestTypeAmount:
		// case core.InterestTypeNone:
		// }
		var savingsComputed *usecase.SavingsInterestComputationResult
		switch generatedSavingsInterest.SavingsComputationType {
		case core.SavingsComputationTypeDailyLowestBalance:
			computation := usecase.SavingsInterestComputation{
				DailyBalance:    dailyBalances,
				InterestRate:    memberBrowseRef.BrowseReference.InterestRate,
				InterestTaxRate: generatedSavingsInterest.InterestTaxRate,
				SavingsType:     usecase.SavingsTypeLowest,
				AnnualDivisor:   annualDivisor,
			}

			result := e.usecase.SavingsInterestComputation(computation)
			savingsComputed = &result

		case core.SavingsComputationTypeAverageDailyBalance:
		case core.SavingsComputationTypeMonthlyEndLowestBalance:
		case core.SavingsComputationTypeADBEndBalance:
		case core.SavingsComputationTypeMonthlyLowestBalanceAverage:
		case core.SavingsComputationTypeMonthlyEndBalanceAverage:
		case core.SavingsComputationTypeMonthlyEndBalanceTotal:
		}
		if savingsComputed == nil {
			continue
		}
		entry := &core.GeneratedSavingsInterestEntry{
			CreatedAt:                  time.Now().UTC(),
			CreatedByID:                userOrg.UserID,
			UpdatedAt:                  time.Now().UTC(),
			UpdatedByID:                userOrg.UserID,
			OrganizationID:             userOrg.OrganizationID,
			BranchID:                   *userOrg.BranchID,
			GeneratedSavingsInterestID: generatedSavingsInterest.ID,
			MemberProfileID:            memberBrowseRef.MemberAccountingLedger.MemberProfileID,
			AccountID:                  *memberBrowseRef.BrowseReference.AccountID,
			InterestAmount:             savingsComputed.Interest,
			InterestTax:                savingsComputed.InterestTax,
			EndingBalance:              savingsComputed.EndingBalance,
		}

		if err := e.core.GeneratedSavingsInterestEntryManager.Create(context, entry); err != nil {
			return nil, eris.Wrap(err, "failed to create generated savings interest entry")
		}
		result = append(result, entry)
	}
	return result, nil
}

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
