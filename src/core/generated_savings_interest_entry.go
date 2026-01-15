package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type (
	GeneratedSavingsInterestEntry struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt   time.Time      `gorm:"not null;default:now()"`
		CreatedByID uuid.UUID      `gorm:"type:uuid"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_generate_savings_interest_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_generate_savings_interest_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		GeneratedSavingsInterestID uuid.UUID                 `gorm:"type:uuid;not null;index:idx_generated_savings_interest_entry"`
		GeneratedSavingsInterest   *GeneratedSavingsInterest `gorm:"foreignKey:GeneratedSavingsInterestID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"generated_savings_interest,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid;not null;index:idx_account_member_profile_entry"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT;" json:"account,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null;index:idx_account_member_profile_entry"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT;" json:"member_profile,omitempty"`

		EndingBalance  float64 `gorm:"type:decimal(15,2);not null" json:"ending_balance" validate:"required"`
		InterestAmount float64 `gorm:"type:decimal(15,2);not null" json:"interest_amount" validate:"required"`
		InterestTax    float64 `gorm:"type:decimal(15,2);not null" json:"interest_tax" validate:"required"`
	}

	GeneratedSavingsInterestEntryResponse struct {
		ID             uuid.UUID             `json:"id"`
		CreatedAt      string                `json:"created_at"`
		CreatedByID    uuid.UUID             `json:"created_by_id"`
		CreatedBy      *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt      string                `json:"updated_at"`
		UpdatedByID    uuid.UUID             `json:"updated_by_id"`
		UpdatedBy      *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       uuid.UUID             `json:"branch_id"`
		Branch         *BranchResponse       `json:"branch,omitempty"`

		GeneratedSavingsInterestID uuid.UUID                         `json:"generated_savings_interest_id"`
		GeneratedSavingsInterest   *GeneratedSavingsInterestResponse `json:"generated_savings_interest,omitempty"`
		AccountID                  uuid.UUID                         `json:"account_id"`
		Account                    *AccountResponse                  `json:"account,omitempty"`
		MemberProfileID            uuid.UUID                         `json:"member_profile_id"`
		MemberProfile              *MemberProfileResponse            `json:"member_profile,omitempty"`
		EndingBalance              float64                           `json:"ending_balance"`
		InterestAmount             float64                           `json:"interest_amount"`
		InterestTax                float64                           `json:"interest_tax"`
	}

	GeneratedSavingsInterestEntryRequest struct {
		AccountID       uuid.UUID `json:"account_id" validate:"required"`
		MemberProfileID uuid.UUID `json:"member_profile_id" validate:"required"`
		InterestAmount  float64   `json:"interest_amount" validate:"required"`
		InterestTax     float64   `json:"interest_tax" validate:"required"`
	}

	GeneratedSavingsInterestEntryDailyBalance struct {
		Balance float64 `json:"balance"`
		Date    string  `json:"date"`
		Type    string  `json:"type"` // "increase", "decrease", "no_change"
	}
	GeneratedSavingsInterestEntryDailyBalanceResponse struct {
		BeginningBalance    float64                                     `json:"beginning_balance"`
		EndingBalance       float64                                     `json:"ending_balance"`
		AverageDailyBalance float64                                     `json:"average_daily_balance"`
		LowestBalance       float64                                     `json:"lowest_balance"`
		HighestBalance      float64                                     `json:"highest_balance"`
		DailyBalance        []GeneratedSavingsInterestEntryDailyBalance `json:"daily_balance"`
		Account             *AccountResponse                            `json:"account,omitempty"`
		MemberProfile       *MemberProfileResponse                      `json:"member_profile,omitempty"`
	}
)

func GeneratedSavingsInterestEntryManager(service *horizon.HorizonService) *registry.Registry[GeneratedSavingsInterestEntry, GeneratedSavingsInterestEntryResponse, GeneratedSavingsInterestEntryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		GeneratedSavingsInterestEntry, GeneratedSavingsInterestEntryResponse, GeneratedSavingsInterestEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Organization", "Branch", "GeneratedSavingsInterest", "Account", "MemberProfile",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *GeneratedSavingsInterestEntry) *GeneratedSavingsInterestEntryResponse {
			if data == nil {
				return nil
			}
			return &GeneratedSavingsInterestEntryResponse{
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

		Created: func(data *GeneratedSavingsInterestEntry) registry.Topics {
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
		Updated: func(data *GeneratedSavingsInterestEntry) registry.Topics {
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
		Deleted: func(data *GeneratedSavingsInterestEntry) registry.Topics {
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
	context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*GeneratedSavingsInterestEntry, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return GeneratedSavingsInterestEntryManager(service).ArrFind(context, filters, nil)
}

func GenerateSavingsInterestEntryByGeneratedSavingsInterest(
	context context.Context, service *horizon.HorizonService, generatedSavingsInterestID uuid.UUID) ([]*GeneratedSavingsInterestEntry, error) {
	filters := []query.ArrFilterSQL{
		{Field: "generated_savings_interest_id", Op: query.ModeEqual, Value: generatedSavingsInterestID},
	}

	return GeneratedSavingsInterestEntryManager(service).ArrFind(context, filters, nil)
}

func GenerateSavingsInterestEntryByAccount(
	context context.Context, service *horizon.HorizonService, accountID, organizationID, branchID uuid.UUID) ([]*GeneratedSavingsInterestEntry, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
	}

	return GeneratedSavingsInterestEntryManager(service).ArrFind(context, filters, nil)
}

func GenerateSavingsInterestEntryByMemberProfile(
	context context.Context, service *horizon.HorizonService, memberProfileID, organizationID, branchID uuid.UUID) (
	[]*GeneratedSavingsInterestEntry, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
	}

	return GeneratedSavingsInterestEntryManager(service).ArrFind(context, filters, nil)
}

func GenerateSavingsInterestEntryByEndingBalanceRange(
	context context.Context, service *horizon.HorizonService, minEndingBalance, maxEndingBalance float64, organizationID, branchID uuid.UUID) (
	[]*GeneratedSavingsInterestEntry, error) {
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
) (*GeneratedSavingsInterestEntryDailyBalanceResponse, error) {
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
		return &GeneratedSavingsInterestEntryDailyBalanceResponse{
			BeginningBalance:    0,
			EndingBalance:       0,
			AverageDailyBalance: 0,
			LowestBalance:       0,
			HighestBalance:      0,
			DailyBalance:        []GeneratedSavingsInterestEntryDailyBalance{},
		}, nil
	}

	var allDailyBalances []GeneratedSavingsInterestEntryDailyBalance
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

		allDailyBalances = append(allDailyBalances, GeneratedSavingsInterestEntryDailyBalance{
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

	return &GeneratedSavingsInterestEntryDailyBalanceResponse{
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
