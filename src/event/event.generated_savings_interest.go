package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/usecase"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

func GenerateSavingsInterestEntries(
	context context.Context, service *horizon.HorizonService,
	userOrg *types.UserOrganization,
	generatedSavingsInterest types.GeneratedSavingsInterest,
	annualDivisor int,
) ([]*types.GeneratedSavingsInterestEntry, error) {

	result := []*types.GeneratedSavingsInterestEntry{}
	now := time.Now().UTC()

	browseReferences, err := core.BrowseReferenceByField(
		context, service,
		userOrg.OrganizationID,
		*userOrg.BranchID,
		generatedSavingsInterest.AccountID,
		generatedSavingsInterest.MemberTypeID,
	)
	if err != nil {

		return nil, eris.Wrap(err, "failed to get browse references")
	}

	memberBrowseReferences, err := core.MemberAccountingLedgerByBrowseReference(
		context, service, generatedSavingsInterest.IncludeClosedAccount, browseReferences)
	if err != nil {

		return nil, eris.Wrap(err, "failed to get member accounting ledgers by browse reference")
	}

	for _, memberBrowseRef := range memberBrowseReferences {

		if memberBrowseRef == nil {

			continue
		}

		if memberBrowseRef.MemberAccountingLedger == nil {

			continue
		}
		currency := memberBrowseRef.MemberAccountingLedger.Account.Currency
		memberProfile := memberBrowseRef.MemberAccountingLedger.MemberProfile
		account := memberBrowseRef.MemberAccountingLedger.Account

		loc, err := time.LoadLocation(currency.Timezone)
		if err != nil {

			return nil, eris.Wrap(err, "failed to load location")
		}
		from := generatedSavingsInterest.LastComputationDate.In(loc)
		to := generatedSavingsInterest.NewComputationDate.In(loc)

		dailyBalances, err := core.GetDailyEndingBalances(
			context, service,
			from, to,
			*memberBrowseRef.BrowseReference.AccountID,
			memberProfile.ID,
			userOrg.OrganizationID, *userOrg.BranchID,
		)
		if err != nil {

			return nil, eris.Wrap(err, "failed to get daily ending balances")
		}

		if len(dailyBalances) == 0 {

			continue
		}

		interestRate := memberBrowseRef.BrowseReference.InterestRate
		var savingsComputed *usecase.SavingsInterestComputationResult

		lastBalance := dailyBalances[len(dailyBalances)-1]
		lastBalanceDec := decimal.NewFromFloat(lastBalance)
		minimumBalanceDec := decimal.NewFromFloat(memberBrowseRef.BrowseReference.MinimumBalance)
		chargesDec := decimal.NewFromFloat(memberBrowseRef.BrowseReference.Charges)

		if lastBalanceDec.Equal(decimal.Zero) {
			continue
		}

		if lastBalanceDec.LessThan(minimumBalanceDec) {

			if chargesDec.Equal(decimal.Zero) {
				continue
			}

			savingsComputed = &usecase.SavingsInterestComputationResult{
				Interest:      chargesDec.Neg().InexactFloat64(),
				InterestTax:   decimal.Zero.InexactFloat64(),
				EndingBalance: lastBalanceDec.Sub(chargesDec).InexactFloat64(),
			}
		} else {

			switch memberBrowseRef.BrowseReference.InterestType {
			case types.InterestTypeYear:
				for _, rateByYear := range memberBrowseRef.BrowseReference.InterestRatesByYear {

					memberHistory, err := core.GetMemberTypeHistoryLatest(
						context, service,
						memberProfile.ID, *memberBrowseRef.BrowseReference.MemberTypeID,
						userOrg.OrganizationID, *userOrg.BranchID,
					)
					if err != nil {
						return nil, eris.Wrap(err, "failed to get member type history latest")
					}
					if memberHistory == nil {
						continue
					}

					memberTypeFromUTC := time.Date(rateByYear.FromYear, 1, 1, 0, 0, 0, 0, loc).UTC()
					memberTypeToUTC := time.Date(rateByYear.ToYear, 12, 31, 23, 59, 59, 999999999, loc).UTC()
					memberHistoryDateUTC := time.Date(memberHistory.CreatedAt.Year(), memberHistory.CreatedAt.Month(), memberHistory.CreatedAt.Day(), 12, 0, 0, 0, time.UTC)

					if (memberHistoryDateUTC.Equal(memberTypeFromUTC) || memberHistoryDateUTC.After(memberTypeFromUTC)) &&
						(memberHistoryDateUTC.Equal(memberTypeToUTC) || memberHistoryDateUTC.Before(memberTypeToUTC)) {
						interestRate = rateByYear.InterestRate
						break
					}
				}
			case types.InterestTypeDate:
				for _, rateByDate := range memberBrowseRef.BrowseReference.InterestRatesByDate {

					memberHistory, err := core.GetMemberTypeHistoryLatest(
						context, service,
						memberProfile.ID, *memberBrowseRef.BrowseReference.MemberTypeID,
						userOrg.OrganizationID, *userOrg.BranchID,
					)
					if err != nil {
						return nil, eris.Wrap(err, "failed to get member type history latest")
					}
					if memberHistory == nil {
						continue
					}
					memberTypeFromUTC := time.Date(rateByDate.FromDate.Year(), rateByDate.FromDate.Month(), rateByDate.FromDate.Day(), 0, 0, 0, 0, loc).UTC()
					memberTypeToUTC := time.Date(rateByDate.ToDate.Year(), rateByDate.ToDate.Month(), rateByDate.ToDate.Day(), 23, 59, 59, 999999999, loc).UTC()
					memberHistoryDateUTC := time.Date(memberHistory.CreatedAt.Year(), memberHistory.CreatedAt.Month(), memberHistory.CreatedAt.Day(), 12, 0, 0, 0, time.UTC)

					if (memberHistoryDateUTC.Equal(memberTypeFromUTC) || memberHistoryDateUTC.After(memberTypeFromUTC)) &&
						(memberHistoryDateUTC.Equal(memberTypeToUTC) || memberHistoryDateUTC.Before(memberTypeToUTC)) {
						interestRate = rateByDate.InterestRate
						break
					}
				}
			case types.InterestTypeAmount:
				for _, rateByAmount := range memberBrowseRef.BrowseReference.InterestRatesByAmount {
					fromAmountDec := decimal.NewFromFloat(rateByAmount.FromAmount)
					toAmountDec := decimal.NewFromFloat(rateByAmount.ToAmount)
					if lastBalanceDec.Cmp(fromAmountDec) >= 0 && lastBalanceDec.Cmp(toAmountDec) <= 0 {
						interestRate = rateByAmount.InterestRate
						break
					}
				}
			}

			switch generatedSavingsInterest.SavingsComputationType {
			case types.SavingsComputationTypeDailyLowestBalance:
				computation := usecase.SavingsInterest{
					DailyBalance:    dailyBalances,
					InterestRate:    interestRate,
					InterestTaxRate: generatedSavingsInterest.InterestTaxRate,
					SavingsType:     usecase.SavingsTypeLowest,
					AnnualDivisor:   annualDivisor,
				}
				result := usecase.SavingsInterestComputation(computation)
				savingsComputed = &result
			case types.SavingsComputationTypeAverageDailyBalance:
				computation := usecase.SavingsInterest{
					DailyBalance:    dailyBalances,
					InterestRate:    interestRate,
					InterestTaxRate: generatedSavingsInterest.InterestTaxRate,
					SavingsType:     usecase.SavingsTypeAverage,
					AnnualDivisor:   annualDivisor,
				}
				result := usecase.SavingsInterestComputation(computation)
				savingsComputed = &result
			case types.SavingsComputationTypeMonthlyEndBalanceTotal:
				computation := usecase.SavingsInterest{
					DailyBalance:    dailyBalances,
					InterestRate:    interestRate,
					InterestTaxRate: generatedSavingsInterest.InterestTaxRate,
					SavingsType:     usecase.SavingsTypeEnd,
					AnnualDivisor:   annualDivisor,
				}
				result := usecase.SavingsInterestComputation(computation)
				savingsComputed = &result
			case types.SavingsComputationTypeADBEndBalance:
			case types.SavingsComputationTypeMonthlyEndLowestBalance:
			case types.SavingsComputationTypeMonthlyLowestBalanceAverage:
			case types.SavingsComputationTypeMonthlyEndBalanceAverage:
			}
		}

		if savingsComputed == nil {

			continue
		}
		if !account.IsTaxable {

			savingsComputed.InterestTax = 0
		}
		memberProfile = memberBrowseRef.MemberAccountingLedger.MemberProfile

		account, err = core.AccountManager(service).GetByID(context, account.ID)
		if err != nil {

			return nil, eris.Wrap(err, "failed to get account by ID")
		}

		entry := &types.GeneratedSavingsInterestEntry{
			CreatedAt:                  now,
			CreatedByID:                userOrg.UserID,
			UpdatedAt:                  now,
			UpdatedByID:                userOrg.UserID,
			OrganizationID:             userOrg.OrganizationID,
			BranchID:                   *userOrg.BranchID,
			GeneratedSavingsInterestID: generatedSavingsInterest.ID,
			MemberProfileID:            memberBrowseRef.MemberAccountingLedger.MemberProfileID,
			MemberProfile:              memberProfile,
			Account:                    account,
			AccountID:                  *memberBrowseRef.BrowseReference.AccountID,
			InterestAmount:             savingsComputed.Interest,
			InterestTax:                savingsComputed.InterestTax,
			EndingBalance:              savingsComputed.EndingBalance,
		}

		result = append(result, entry)
	}

	return result, nil
}
