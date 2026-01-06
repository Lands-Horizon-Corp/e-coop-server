package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

func (e *Event) GenerateSavingsInterestEntries(
	context context.Context,
	userOrg *core.UserOrganization,
	generatedSavingsInterest core.GeneratedSavingsInterest,
	annualDivisor int,
) ([]*core.GeneratedSavingsInterestEntry, error) {

	result := []*core.GeneratedSavingsInterestEntry{}
	now := time.Now().UTC()

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

	memberBrowseReferences, err := e.core.MemberAccountingLedgerByBrowseReference(
		context, generatedSavingsInterest.IncludeClosedAccount, browseReferences)
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

		dailyBalances, err := e.core.GetDailyEndingBalances(
			context,
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
			case core.InterestTypeYear:
				for _, rateByYear := range memberBrowseRef.BrowseReference.InterestRatesByYear {

					memberHistory, err := e.core.GetMemberTypeHistoryLatest(
						context,
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
			case core.InterestTypeDate:
				for _, rateByDate := range memberBrowseRef.BrowseReference.InterestRatesByDate {

					memberHistory, err := e.core.GetMemberTypeHistoryLatest(
						context,
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
			case core.InterestTypeAmount:
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
			case core.SavingsComputationTypeDailyLowestBalance:
				computation := usecase.SavingsInterest{
					DailyBalance:    dailyBalances,
					InterestRate:    interestRate,
					InterestTaxRate: generatedSavingsInterest.InterestTaxRate,
					SavingsType:     usecase.SavingsTypeLowest,
					AnnualDivisor:   annualDivisor,
				}
				result := usecase.SavingsInterestComputation(computation)
				savingsComputed = &result
			case core.SavingsComputationTypeAverageDailyBalance:
				computation := usecase.SavingsInterest{
					DailyBalance:    dailyBalances,
					InterestRate:    interestRate,
					InterestTaxRate: generatedSavingsInterest.InterestTaxRate,
					SavingsType:     usecase.SavingsTypeAverage,
					AnnualDivisor:   annualDivisor,
				}
				result := usecase.SavingsInterestComputation(computation)
				savingsComputed = &result
			case core.SavingsComputationTypeMonthlyEndBalanceTotal:
				computation := usecase.SavingsInterest{
					DailyBalance:    dailyBalances,
					InterestRate:    interestRate,
					InterestTaxRate: generatedSavingsInterest.InterestTaxRate,
					SavingsType:     usecase.SavingsTypeEnd,
					AnnualDivisor:   annualDivisor,
				}
				result := usecase.SavingsInterestComputation(computation)
				savingsComputed = &result
			case core.SavingsComputationTypeADBEndBalance:
			case core.SavingsComputationTypeMonthlyEndLowestBalance:
			case core.SavingsComputationTypeMonthlyLowestBalanceAverage:
			case core.SavingsComputationTypeMonthlyEndBalanceAverage:
			}
		}

		if savingsComputed == nil {

			continue
		}
		if !account.IsTaxable {

			savingsComputed.InterestTax = 0
		}
		memberProfile = memberBrowseRef.MemberAccountingLedger.MemberProfile

		account, err = e.core.AccountManager().GetByID(context, account.ID)
		if err != nil {

			return nil, eris.Wrap(err, "failed to get account by ID")
		}

		entry := &core.GeneratedSavingsInterestEntry{
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
