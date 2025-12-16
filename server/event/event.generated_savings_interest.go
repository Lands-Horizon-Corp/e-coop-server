package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/rotisserie/eris"
)

func (e *Event) GenerateSavingsInterestEntries(
	context context.Context,
	userOrg *core.UserOrganization,
	generatedSavingsInterest core.GeneratedSavingsInterest,
	annualDivisor int,
) ([]*core.GeneratedSavingsInterestEntry, error) {

	fmt.Println("DEBUG: Start GenerateSavingsInterestEntries")
	result := []*core.GeneratedSavingsInterestEntry{}
	now := time.Now().UTC()

	fmt.Printf("DEBUG: userOrg=%+v\n", userOrg)
	fmt.Printf("DEBUG: generatedSavingsInterest=%+v\n", generatedSavingsInterest)

	fmt.Println("DEBUG: Before BrowseReferenceByField")
	browseReferences, err := e.core.BrowseReferenceByField(
		context,
		userOrg.OrganizationID,
		*userOrg.BranchID,
		generatedSavingsInterest.AccountID,
		generatedSavingsInterest.MemberTypeID,
	)
	if err != nil {
		fmt.Println("DEBUG: Error after BrowseReferenceByField")
		return nil, eris.Wrap(err, "failed to get browse references")
	}
	fmt.Printf("DEBUG: browseReferences=%+v\n", browseReferences)

	fmt.Println("DEBUG: Before MemberAccountingLedgerByBrowseReference")
	memberBrowseReferences, err := e.core.MemberAccountingLedgerByBrowseReference(
		context, generatedSavingsInterest.IncludeClosedAccount, browseReferences)
	if err != nil {
		fmt.Println("DEBUG: Error after MemberAccountingLedgerByBrowseReference")
		return nil, eris.Wrap(err, "failed to get member accounting ledgers by browse reference")
	}
	fmt.Printf("DEBUG: memberBrowseReferences len=%d\n", len(memberBrowseReferences))

	for i, memberBrowseRef := range memberBrowseReferences {
		fmt.Printf("DEBUG: Loop idx=%d, memberBrowseRef=%+v\n", i, memberBrowseRef)
		if memberBrowseRef == nil {
			fmt.Printf("DEBUG: memberBrowseRef IS NIL at idx=%d\n", i)
			continue
		}

		if memberBrowseRef.MemberAccountingLedger == nil {
			fmt.Printf("DEBUG: MemberAccountingLedger IS NIL at idx=%d\n", i)
			continue
		}
		currency := memberBrowseRef.MemberAccountingLedger.Account.Currency
		memberProfile := memberBrowseRef.MemberAccountingLedger.MemberProfile
		account := memberBrowseRef.MemberAccountingLedger.Account

		fmt.Printf("DEBUG: currency=%+v, account=%+v, memberProfile=%+v\n", currency, account, memberProfile)

		fmt.Println("DEBUG: Before time.LoadLocation")
		loc, err := time.LoadLocation(currency.Timezone)
		if err != nil {
			fmt.Println("DEBUG: Error after time.LoadLocation")
			return nil, eris.Wrap(err, "failed to load location")
		}
		from := generatedSavingsInterest.LastComputationDate.In(loc)
		to := generatedSavingsInterest.NewComputationDate.In(loc)

		fmt.Println("DEBUG: Before GetDailyEndingBalances")
		dailyBalances, err := e.core.GetDailyEndingBalances(
			context,
			from, to,
			*memberBrowseRef.BrowseReference.AccountID,
			memberProfile.ID,
			userOrg.OrganizationID, *userOrg.BranchID,
		)
		if err != nil {
			fmt.Println("DEBUG: Error after GetDailyEndingBalances")
			return nil, eris.Wrap(err, "failed to get daily ending balances")
		}
		fmt.Printf("DEBUG: Got %d dailyBalances\n", len(dailyBalances))
		if len(dailyBalances) == 0 {
			fmt.Printf("DEBUG: No dailyBalances for memberProfileID=%v\n", memberProfile.ID)
			continue
		}

		interestRate := memberBrowseRef.BrowseReference.InterestRate
		var savingsComputed *usecase.SavingsInterestComputationResult

		lastBalance := dailyBalances[len(dailyBalances)-1]
		fmt.Printf("DEBUG: lastBalance=%v\n", lastBalance)

		if e.provider.Service.Decimal.IsEqual(lastBalance, 0) {
			fmt.Println("DEBUG: lastBalance is zero; skipping")
			continue
		}
		fmt.Printf("DEBUG: Checking minimum balance. last=%v, min=%v\n", lastBalance, memberBrowseRef.BrowseReference.MinimumBalance)

		if e.provider.Service.Decimal.IsLessThan(lastBalance, memberBrowseRef.BrowseReference.MinimumBalance) {
			fmt.Printf("DEBUG: lastBalance < MinimumBalance, Charges=%v\n", memberBrowseRef.BrowseReference.Charges)
			if memberBrowseRef.BrowseReference.Charges == 0 {
				continue
			}

			savingsComputed = &usecase.SavingsInterestComputationResult{
				Interest:    e.provider.Service.Decimal.Negate(memberBrowseRef.BrowseReference.Charges),
				InterestTax: 0,
				EndingBalance: e.provider.Service.Decimal.Subtract(
					lastBalance,
					memberBrowseRef.BrowseReference.Charges),
			}
		} else {
			fmt.Printf("DEBUG: Checking InterestType=%v\n", memberBrowseRef.BrowseReference.InterestType)
			switch memberBrowseRef.BrowseReference.InterestType {
			case core.InterestTypeYear:
				for _, rateByYear := range memberBrowseRef.BrowseReference.InterestRatesByYear {
					fmt.Printf("DEBUG: rateByYear=%v\n", rateByYear)
					memberHistory, err := e.core.GetMemberTypeHistoryLatest(
						context,
						memberProfile.ID, *memberBrowseRef.BrowseReference.MemberTypeID,
						userOrg.OrganizationID, *userOrg.BranchID,
					)
					if err != nil {
						fmt.Println("DEBUG: Error after GetMemberTypeHistoryLatest (Year)")
						return nil, eris.Wrap(err, "failed to get member type history latest")
					}
					if memberHistory == nil {
						fmt.Println("DEBUG: memberHistory IS NIL (Year)")
						continue
					}
					fmt.Printf("DEBUG: memberHistory=%v\n", memberHistory)
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
					fmt.Printf("DEBUG: rateByDate=%v\n", rateByDate)
					memberHistory, err := e.core.GetMemberTypeHistoryLatest(
						context,
						memberProfile.ID, *memberBrowseRef.BrowseReference.MemberTypeID,
						userOrg.OrganizationID, *userOrg.BranchID,
					)
					if err != nil {
						fmt.Println("DEBUG: Error after GetMemberTypeHistoryLatest (Date)")
						return nil, eris.Wrap(err, "failed to get member type history latest")
					}
					if memberHistory == nil {
						fmt.Println("DEBUG: memberHistory IS NIL (Date)")
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
					fmt.Printf("DEBUG: rateByAmount=%v\n", rateByAmount)
					if e.provider.Service.Decimal.IsGreaterThanOrEqual(lastBalance, rateByAmount.FromAmount) &&
						(e.provider.Service.Decimal.IsLessThanOrEqual(lastBalance, rateByAmount.ToAmount)) {
						interestRate = rateByAmount.InterestRate
						break
					}
				}
			}
			fmt.Printf("DEBUG: ComputationType=%v\n", generatedSavingsInterest.SavingsComputationType)
			switch generatedSavingsInterest.SavingsComputationType {
			case core.SavingsComputationTypeDailyLowestBalance:
				computation := usecase.SavingsInterestComputation{
					DailyBalance:    dailyBalances,
					InterestRate:    interestRate,
					InterestTaxRate: generatedSavingsInterest.InterestTaxRate,
					SavingsType:     usecase.SavingsTypeLowest,
					AnnualDivisor:   annualDivisor,
				}
				fmt.Printf("DEBUG: Run SavingsInterestComputation (Lowest Balance)\n")
				result := e.usecase.SavingsInterestComputation(computation)
				savingsComputed = &result
			case core.SavingsComputationTypeAverageDailyBalance:
				computation := usecase.SavingsInterestComputation{
					DailyBalance:    dailyBalances,
					InterestRate:    interestRate,
					InterestTaxRate: generatedSavingsInterest.InterestTaxRate,
					SavingsType:     usecase.SavingsTypeAverage,
					AnnualDivisor:   annualDivisor,
				}
				fmt.Printf("DEBUG: Run SavingsInterestComputation (Average Daily)\n")
				result := e.usecase.SavingsInterestComputation(computation)
				savingsComputed = &result
			case core.SavingsComputationTypeMonthlyEndBalanceTotal:
				computation := usecase.SavingsInterestComputation{
					DailyBalance:    dailyBalances,
					InterestRate:    interestRate,
					InterestTaxRate: generatedSavingsInterest.InterestTaxRate,
					SavingsType:     usecase.SavingsTypeEnd,
					AnnualDivisor:   annualDivisor,
				}
				fmt.Printf("DEBUG: Run SavingsInterestComputation (Monthly End Total)\n")
				result := e.usecase.SavingsInterestComputation(computation)
				savingsComputed = &result
			case core.SavingsComputationTypeADBEndBalance:
			case core.SavingsComputationTypeMonthlyEndLowestBalance:
			case core.SavingsComputationTypeMonthlyLowestBalanceAverage:
			case core.SavingsComputationTypeMonthlyEndBalanceAverage:
			}
		}

		fmt.Printf("DEBUG: savingsComputed=%+v\n", savingsComputed)
		if savingsComputed == nil {
			fmt.Println("DEBUG: savingsComputed is nil; skipping")
			continue
		}
		if !account.IsTaxable {
			fmt.Println("DEBUG: Account is not taxable; setting InterestTax=0")
			savingsComputed.InterestTax = 0
		}
		memberProfile = memberBrowseRef.MemberAccountingLedger.MemberProfile
		fmt.Println("DEBUG: Before AccountManager.GetByID")
		account, err = e.core.AccountManager.GetByID(context, account.ID)
		if err != nil {
			fmt.Println("DEBUG: Error after AccountManager.GetByID")
			return nil, eris.Wrap(err, "failed to get account by ID")
		}
		fmt.Printf("DEBUG: account=%+v\n", account)
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
		fmt.Printf("DEBUG: Generated entry=%+v\n", entry)

		result = append(result, entry)
	}

	fmt.Println("DEBUG: Finished loop")
	return result, nil
}
