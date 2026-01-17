package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func GeneratedSavingsInterestEntryManager(service *horizon.HorizonService) *registry.Registry[
	types.GeneratedSavingsInterestEntry, types.GeneratedSavingsInterestEntryResponse, types.GeneratedSavingsInterestEntryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.GeneratedSavingsInterestEntry, types.GeneratedSavingsInterestEntryResponse, types.GeneratedSavingsInterestEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Organization", "Branch", "GeneratedSavingsInterest", "Account", "MemberProfile",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.GeneratedSavingsInterestEntry) *types.GeneratedSavingsInterestEntryResponse {
			if data == nil {
				return nil
			}
			return &types.GeneratedSavingsInterestEntryResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),

				GeneratedSavingsInterestID: data.GeneratedSavingsInterestID,
				GeneratedSavingsInterest:   GeneratedSavingsInterestManager(service).ToModel(data.GeneratedSavingsInterest),
				AccountID:                  data.AccountID,
				Account:                    AccountManager(service).ToModel(data.Account),
				MemberProfileID:            data.MemberProfileID,
				MemberProfile:              MemberProfileManager(service).ToModel(data.MemberProfile),
				EndingBalance:              data.EndingBalance,
				InterestAmount:             data.InterestAmount,
				InterestTax:                data.InterestTax,
			}
		},

		Created: func(data *types.GeneratedSavingsInterestEntry) registry.Topics {
			return []string{
				"generated_savings_interest_entry.create",
				fmt.Sprintf("generated_savings_interest_entry.create.%s", data.ID),
				fmt.Sprintf("generated_savings_interest_entry.create.generated_savings_interest.%s", data.GeneratedSavingsInterestID),
				fmt.Sprintf("generated_savings_interest_entry.create.account.%s", data.AccountID),
				fmt.Sprintf("generated_savings_interest_entry.create.member_profile.%s", data.MemberProfileID),
				fmt.Sprintf("generated_savings_interest_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("generated_savings_interest_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.GeneratedSavingsInterestEntry) registry.Topics {
			return []string{
				"generate_savings_interest_entry.update",
				fmt.Sprintf("generate_savings_interest_entry.update.%s", data.ID),
				fmt.Sprintf("generate_savings_interest_entry.update.generated_savings_interest.%s", data.GeneratedSavingsInterestID),
				fmt.Sprintf("generate_savings_interest_entry.update.account.%s", data.AccountID),
				fmt.Sprintf("generate_savings_interest_entry.update.member_profile.%s", data.MemberProfileID),
				fmt.Sprintf("generate_savings_interest_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("generate_savings_interest_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.GeneratedSavingsInterestEntry) registry.Topics {
			return []string{
				"generated_savings_interest_entry.delete",
				fmt.Sprintf("generated_savings_interest_entry.delete.%s", data.ID),
				fmt.Sprintf("generated_savings_interest_entry.delete.generated_savings_interest.%s", data.GeneratedSavingsInterestID),
				fmt.Sprintf("generated_savings_interest_entry.delete.account.%s", data.AccountID),
				fmt.Sprintf("generated_savings_interest_entry.delete.member_profile.%s", data.MemberProfileID),
				fmt.Sprintf("generated_savings_interest_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("generated_savings_interest_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GenerateSavingsInterestEntryCurrentBranch(
	context context.Context, service *horizon.HorizonService, organizationID uuid.UUID,
	branchID uuid.UUID) ([]*types.GeneratedSavingsInterestEntry, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return GeneratedSavingsInterestEntryManager(service).ArrFind(context, filters, nil)
}

func GenerateSavingsInterestEntryByGeneratedSavingsInterest(
	context context.Context, service *horizon.HorizonService,
	generatedSavingsInterestID uuid.UUID) ([]*types.GeneratedSavingsInterestEntry, error) {
	filters := []query.ArrFilterSQL{
		{Field: "generated_savings_interest_id", Op: query.ModeEqual, Value: generatedSavingsInterestID},
	}

	return GeneratedSavingsInterestEntryManager(service).ArrFind(context, filters, nil)
}

func GenerateSavingsInterestEntryByAccount(
	context context.Context, service *horizon.HorizonService, accountID, organizationID,
	branchID uuid.UUID) ([]*types.GeneratedSavingsInterestEntry, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
	}

	return GeneratedSavingsInterestEntryManager(service).ArrFind(context, filters, nil)
}

func GenerateSavingsInterestEntryByMemberProfile(
	context context.Context, service *horizon.HorizonService, memberProfileID, organizationID, branchID uuid.UUID) (
	[]*types.GeneratedSavingsInterestEntry, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
	}

	return GeneratedSavingsInterestEntryManager(service).ArrFind(context, filters, nil)
}

func GenerateSavingsInterestEntryByEndingBalanceRange(
	context context.Context, service *horizon.HorizonService, minEndingBalance, maxEndingBalance float64,
	organizationID, branchID uuid.UUID) (
	[]*types.GeneratedSavingsInterestEntry, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "ending_balance", Op: query.ModeGTE, Value: minEndingBalance},
		{Field: "ending_balance", Op: query.ModeLTE, Value: maxEndingBalance},
	}

	return GeneratedSavingsInterestEntryManager(service).ArrFind(context, filters, nil)
}

func DailyBalances(
	context context.Context, service *horizon.HorizonService,
	generatedSavingsInterestEntryID uuid.UUID,
) (*types.GeneratedSavingsInterestEntryDailyBalanceResponse, error) {
	generatedSavingsInterestEntry, err := GeneratedSavingsInterestEntryManager(service).GetByID(context, generatedSavingsInterestEntryID)
	if err != nil {
		return nil, err
	}
	generatedSavingsInterest, err := GeneratedSavingsInterestManager(service).GetByID(context, generatedSavingsInterestEntry.GeneratedSavingsInterestID)
	if err != nil {
		return nil, err
	}

	dailyBalances, err := GetDailyEndingBalances(
		context, service,
		generatedSavingsInterest.LastComputationDate,
		generatedSavingsInterest.NewComputationDate,
		generatedSavingsInterestEntry.AccountID,
		generatedSavingsInterestEntry.MemberProfileID,
		generatedSavingsInterestEntry.OrganizationID,
		generatedSavingsInterestEntry.BranchID,
	)
	if err != nil {
		return nil, err
	}

	if len(dailyBalances) == 0 {
		return &types.GeneratedSavingsInterestEntryDailyBalanceResponse{
			BeginningBalance:    0,
			EndingBalance:       0,
			AverageDailyBalance: 0,
			LowestBalance:       0,
			HighestBalance:      0,
			DailyBalance:        []types.GeneratedSavingsInterestEntryDailyBalance{},
		}, nil
	}

	var allDailyBalances []types.GeneratedSavingsInterestEntryDailyBalance
	totalBalanceSum := decimal.NewFromFloat(0)
	totalDays := 0
	lowestBalance := decimal.NewFromFloat(-1)
	highestBalance := decimal.NewFromFloat(0)
	beginningBalance := decimal.NewFromFloat(-1)
	endingBalance := decimal.NewFromFloat(0)

	currentDate := generatedSavingsInterest.NewComputationDate
	previousBalance := decimal.NewFromFloat(-1)

	for i, balance := range dailyBalances {
		dateStr := currentDate.AddDate(0, 0, i).Format("2006-01-02")
		balanceDecimal := decimal.NewFromFloat(balance)
		var changeType string

		if previousBalance.Equal(decimal.NewFromFloat(-1)) {
			changeType = "no_change"
		} else if balanceDecimal.GreaterThan(previousBalance) {
			changeType = "increase"
		} else if balanceDecimal.LessThan(previousBalance) {
			changeType = "decrease"
		} else {
			changeType = "no_change"
		}

		allDailyBalances = append(allDailyBalances, types.GeneratedSavingsInterestEntryDailyBalance{
			Balance: balance,
			Date:    dateStr,
			Type:    changeType,
		})

		totalBalanceSum = totalBalanceSum.Add(balanceDecimal)
		totalDays++

		if beginningBalance.Equal(decimal.NewFromFloat(-1)) {
			beginningBalance = balanceDecimal
		}
		endingBalance = balanceDecimal

		if lowestBalance.Equal(decimal.NewFromFloat(-1)) || balanceDecimal.LessThan(lowestBalance) {
			lowestBalance = balanceDecimal
		}
		if balanceDecimal.GreaterThan(highestBalance) {
			highestBalance = balanceDecimal
		}

		previousBalance = balanceDecimal
	}

	averageDailyBalance := float64(0)
	if totalDays > 0 {
		daysDecimal := decimal.NewFromInt(int64(totalDays))
		averageDailyBalance, _ = totalBalanceSum.Div(daysDecimal).Float64()
	}

	if beginningBalance.Equal(decimal.NewFromFloat(-1)) {
		beginningBalance = decimal.NewFromFloat(0)
	}

	account, err := AccountManager(service).GetByID(context, generatedSavingsInterestEntry.AccountID, "Currency")
	if err != nil {
		return nil, err
	}
	memberProfile, err := MemberProfileManager(service).GetByID(context, generatedSavingsInterestEntry.MemberProfileID)
	if err != nil {
		return nil, err
	}

	return &types.GeneratedSavingsInterestEntryDailyBalanceResponse{
		BeginningBalance:    beginningBalance.InexactFloat64(),
		EndingBalance:       endingBalance.InexactFloat64(),
		AverageDailyBalance: averageDailyBalance,
		LowestBalance:       lowestBalance.InexactFloat64(),
		HighestBalance:      highestBalance.InexactFloat64(),
		DailyBalance:        allDailyBalances,
		Account:             AccountManager(service).ToModel(account),
		MemberProfile:       MemberProfileManager(service).ToModel(memberProfile),
	}, nil
}
