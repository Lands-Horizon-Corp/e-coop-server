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
	now := time.Now().UTC()

	//===== STEP 1: GET BROWSE REFERENCES =====
	// Get browse references based on the generated savings interest criteria
	// This filters accounts by organization, branch, account type, and member type
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

	//===== STEP 2: GET MEMBER ACCOUNTING LEDGERS =====
	// Get member accounting ledgers using the browse references
	// This links browse references to actual member accounts and balances
	memberBrowseReferences, err := e.core.MemberAccountingLedgerByBrowseReference(
		context, generatedSavingsInterest.IncludeClosedAccount, browseReferences)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get member accounting ledgers by browse reference")
	}

	//===== STEP 3: PROCESS EACH MEMBER ACCOUNT =====
	for _, memberBrowseRef := range memberBrowseReferences {

		//===== STEP 3.1: SETUP TIMEZONE AND DATE RANGE =====
		currency := memberBrowseRef.MemberAccountingLedger.Account.Currency
		memberProfile := memberBrowseRef.MemberAccountingLedger.MemberProfile
		account := memberBrowseRef.MemberAccountingLedger.Account

		loc, err := time.LoadLocation(currency.Timezone)
		if err != nil {
			return nil, eris.Wrap(err, "failed to load location")
		}
		from := generatedSavingsInterest.LastComputationDate.In(loc)
		to := generatedSavingsInterest.NewComputationDate.In(loc)

		//===== STEP 3.2: GET DAILY ENDING BALANCES =====
		// Get daily ending balances for the member within the computation period
		// This returns one balance per day for ADB calculations
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

		//===== STEP 3.3: VALIDATE ACCOUNT ELIGIBILITY =====
		// Skip accounts with no balance history
		if len(dailyBalances) == 0 {
			continue
		}

		// Initialize base interest rate from browse reference
		interestRate := memberBrowseRef.BrowseReference.InterestRate
		var savingsComputed *usecase.SavingsInterestComputationResult

		// Get the last balance for minimum balance checking
		lastBalance := dailyBalances[len(dailyBalances)-1]

		// Skip accounts with zero balance
		if e.provider.Service.Decimal.IsEqual(lastBalance, 0) {
			continue
		}

		//===== STEP 3.4: CHECK MINIMUM BALANCE REQUIREMENT =====
		if e.provider.Service.Decimal.IsLessThan(lastBalance, memberBrowseRef.BrowseReference.MinimumBalance) {
			// Below minimum balance - apply charges if configured
			if memberBrowseRef.BrowseReference.Charges == 0 {
				continue // No charges configured, skip this account
			}

			// Apply penalty charges for not maintaining minimum balance
			savingsComputed = &usecase.SavingsInterestComputationResult{
				Interest:    e.provider.Service.Decimal.Negate(memberBrowseRef.BrowseReference.Charges),
				InterestTax: 0,
				EndingBalance: e.provider.Service.Decimal.Subtract(
					lastBalance,
					memberBrowseRef.BrowseReference.Charges),
			}
		} else {

			//===== STEP 3.5: DETERMINE APPLICABLE INTEREST RATE =====
			// Account meets minimum balance - determine correct interest rate
			switch memberBrowseRef.BrowseReference.InterestType {

			//===== STEP 3.5.1: YEAR-BASED INTEREST RATES =====
			case core.InterestTypeYear:
				// Check if member joined within specific year range for rate eligibility
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

					// Convert year range to full date range for comparison
					memberTypeFromUTC := time.Date(rateByYear.FromYear, 1, 1, 0, 0, 0, 0, loc).UTC()
					memberTypeToUTC := time.Date(rateByYear.ToYear, 12, 31, 23, 59, 59, 999999999, loc).UTC()
					memberHistoryDateUTC := time.Date(memberHistory.CreatedAt.Year(), memberHistory.CreatedAt.Month(), memberHistory.CreatedAt.Day(), 12, 0, 0, 0, time.UTC)

					if (memberHistoryDateUTC.Equal(memberTypeFromUTC) || memberHistoryDateUTC.After(memberTypeFromUTC)) &&
						(memberHistoryDateUTC.Equal(memberTypeToUTC) || memberHistoryDateUTC.Before(memberTypeToUTC)) {
						interestRate = rateByYear.InterestRate
						break
					}
				}

			//===== STEP 3.5.2: DATE-BASED INTEREST RATES =====
			case core.InterestTypeDate:
				// Check if member joined within specific date range for rate eligibility
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

					// Convert date range for comparison (full day coverage)
					memberTypeFromUTC := time.Date(rateByDate.FromDate.Year(), rateByDate.FromDate.Month(), rateByDate.FromDate.Day(), 0, 0, 0, 0, loc).UTC()
					memberTypeToUTC := time.Date(rateByDate.ToDate.Year(), rateByDate.ToDate.Month(), rateByDate.ToDate.Day(), 23, 59, 59, 999999999, loc).UTC()
					memberHistoryDateUTC := time.Date(memberHistory.CreatedAt.Year(), memberHistory.CreatedAt.Month(), memberHistory.CreatedAt.Day(), 12, 0, 0, 0, time.UTC)

					if (memberHistoryDateUTC.Equal(memberTypeFromUTC) || memberHistoryDateUTC.After(memberTypeFromUTC)) &&
						(memberHistoryDateUTC.Equal(memberTypeToUTC) || memberHistoryDateUTC.Before(memberTypeToUTC)) {
						interestRate = rateByDate.InterestRate
						break
					}
				}

			//===== STEP 3.5.3: AMOUNT-BASED INTEREST RATES =====
			case core.InterestTypeAmount:
				// Determine rate based on current balance amount tiers
				for _, rateByAmount := range memberBrowseRef.BrowseReference.InterestRatesByAmount {
					if e.provider.Service.Decimal.IsGreaterThanOrEqual(lastBalance, rateByAmount.FromAmount) &&
						(e.provider.Service.Decimal.IsLessThanOrEqual(lastBalance, rateByAmount.ToAmount)) {
						interestRate = rateByAmount.InterestRate
						break
					}
				}
			}

			//===== STEP 3.6: CALCULATE INTEREST BASED ON COMPUTATION TYPE =====
			switch generatedSavingsInterest.SavingsComputationType {

			//===== DAILY LOWEST BALANCE COMPUTATION =====
			case core.SavingsComputationTypeDailyLowestBalance:
				// Use the lowest balance found during the computation period
				computation := usecase.SavingsInterestComputation{
					DailyBalance:    dailyBalances,
					InterestRate:    interestRate,
					InterestTaxRate: generatedSavingsInterest.InterestTaxRate,
					SavingsType:     usecase.SavingsTypeLowest,
					AnnualDivisor:   annualDivisor,
				}
				result := e.usecase.SavingsInterestComputation(computation)
				savingsComputed = &result

			//===== AVERAGE DAILY BALANCE COMPUTATION =====
			case core.SavingsComputationTypeAverageDailyBalance:
				// Calculate average of all daily balances in the period
				computation := usecase.SavingsInterestComputation{
					DailyBalance:    dailyBalances,
					InterestRate:    interestRate,
					InterestTaxRate: generatedSavingsInterest.InterestTaxRate,
					SavingsType:     usecase.SavingsTypeAverage,
					AnnualDivisor:   annualDivisor,
				}
				result := e.usecase.SavingsInterestComputation(computation)
				savingsComputed = &result

			//===== MONTHLY END BALANCE TOTAL COMPUTATION =====
			case core.SavingsComputationTypeMonthlyEndBalanceTotal:
				// Use the final day's balance for the entire period
				computation := usecase.SavingsInterestComputation{
					DailyBalance:    dailyBalances,
					InterestRate:    interestRate,
					InterestTaxRate: generatedSavingsInterest.InterestTaxRate,
					SavingsType:     usecase.SavingsTypeEnd,
					AnnualDivisor:   annualDivisor,
				}
				result := e.usecase.SavingsInterestComputation(computation)
				savingsComputed = &result

			//===== UNIMPLEMENTED COMPUTATION TYPES =====
			// TODO: Implement these computation types that require monthly grouping
			case core.SavingsComputationTypeADBEndBalance:
			case core.SavingsComputationTypeMonthlyEndLowestBalance:
			case core.SavingsComputationTypeMonthlyLowestBalanceAverage:
			case core.SavingsComputationTypeMonthlyEndBalanceAverage:
			}
		}

		//===== STEP 3.7: VALIDATE COMPUTATION RESULT =====
		// Skip accounts where no computation was performed
		if savingsComputed == nil {
			continue
		}
		if !account.IsTaxable {
			savingsComputed.InterestTax = 0
		}

		//===== STEP 3.8: CREATE GENERATED SAVINGS INTEREST ENTRY =====
		// Create database entry with computed interest and tax amounts
		entry := &core.GeneratedSavingsInterestEntry{
			CreatedAt:                  now,
			CreatedByID:                userOrg.UserID,
			UpdatedAt:                  now,
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

		if generatedSavingsInterest.ID != uuid.Nil {
			//===== STEP 3.9: SAVE ENTRY TO DATABASE =====
			if err := e.core.GeneratedSavingsInterestEntryManager.Create(context, entry); err != nil {
				return nil, eris.Wrap(err, "failed to create generated savings interest entry")
			}
		}

		//===== STEP 3.10: ADD TO RESULT COLLECTION =====
		result = append(result, entry)
	}

	//===== STEP 4: RETURN COMPLETED ENTRIES =====
	return result, nil
}
