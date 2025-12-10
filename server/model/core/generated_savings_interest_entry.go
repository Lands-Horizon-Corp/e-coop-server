package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// GeneratedSavingsInterestEntry represents individual savings interest computation entries
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

	// GeneratedSavingsInterestEntryResponse represents the response structure for generated savings interest entry data
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

	// GenerateSavingsInterestEntryRequest represents the request structure for creating/updating generate savings interest entry
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

func (m *Core) generateSavingsInterestEntry() {
	m.Migration = append(m.Migration, &GeneratedSavingsInterestEntry{})
	m.GeneratedSavingsInterestEntryManager = *registry.NewRegistry(registry.RegistryParams[
		GeneratedSavingsInterestEntry, GeneratedSavingsInterestEntryResponse, GeneratedSavingsInterestEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Organization", "Branch", "GeneratedSavingsInterest", "Account", "MemberProfile",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *GeneratedSavingsInterestEntry) *GeneratedSavingsInterestEntryResponse {
			if data == nil {
				return nil
			}
			return &GeneratedSavingsInterestEntryResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),

				GeneratedSavingsInterestID: data.GeneratedSavingsInterestID,
				GeneratedSavingsInterest:   m.GeneratedSavingsInterestManager.ToModel(data.GeneratedSavingsInterest),
				AccountID:                  data.AccountID,
				Account:                    m.AccountManager.ToModel(data.Account),
				MemberProfileID:            data.MemberProfileID,
				MemberProfile:              m.MemberProfileManager.ToModel(data.MemberProfile),
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

// GenerateSavingsInterestEntryCurrentBranch retrieves entries for the specified branch and organization
func (m *Core) GenerateSavingsInterestEntryCurrentBranch(
	context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*GeneratedSavingsInterestEntry, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return m.GeneratedSavingsInterestEntryManager.ArrFind(context, filters, nil)
}

// GenerateSavingsInterestEntryByGeneratedSavingsInterest retrieves entries for a specific generated savings interest
func (m *Core) GenerateSavingsInterestEntryByGeneratedSavingsInterest(
	context context.Context, generatedSavingsInterestID uuid.UUID) ([]*GeneratedSavingsInterestEntry, error) {
	filters := []registry.FilterSQL{
		{Field: "generated_savings_interest_id", Op: query.ModeEqual, Value: generatedSavingsInterestID},
	}

	return m.GeneratedSavingsInterestEntryManager.ArrFind(context, filters, nil)
}

// GenerateSavingsInterestEntryByAccount retrieves entries for a specific account
func (m *Core) GenerateSavingsInterestEntryByAccount(
	context context.Context, accountID, organizationID, branchID uuid.UUID) ([]*GeneratedSavingsInterestEntry, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
	}

	return m.GeneratedSavingsInterestEntryManager.ArrFind(context, filters, nil)
}

// GenerateSavingsInterestEntryByMemberProfile retrieves entries for a specific member profile
func (m *Core) GenerateSavingsInterestEntryByMemberProfile(
	context context.Context, memberProfileID, organizationID, branchID uuid.UUID) (
	[]*GeneratedSavingsInterestEntry, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
	}

	return m.GeneratedSavingsInterestEntryManager.ArrFind(context, filters, nil)
}

// GenerateSavingsInterestEntryByEndingBalanceRange retrieves entries within a specific ending balance range
func (m *Core) GenerateSavingsInterestEntryByEndingBalanceRange(
	context context.Context, minEndingBalance, maxEndingBalance float64, organizationID, branchID uuid.UUID) (
	[]*GeneratedSavingsInterestEntry, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "ending_balance", Op: query.ModeGTE, Value: minEndingBalance},
		{Field: "ending_balance", Op: query.ModeLTE, Value: maxEndingBalance},
	}

	return m.GeneratedSavingsInterestEntryManager.ArrFind(context, filters, nil)
}

func (m *Core) DailyBalances(context context.Context, generatedSavingsInterestEntryID uuid.UUID) (*GeneratedSavingsInterestEntryDailyBalanceResponse, error) {
	generatedSavingsInterestEntry, err := m.GeneratedSavingsInterestEntryManager.GetByID(context, generatedSavingsInterestEntryID)
	if err != nil {
		return nil, err
	}
	generatedSavingsInterest, err := m.GeneratedSavingsInterestManager.GetByID(context, generatedSavingsInterestEntry.GeneratedSavingsInterestID)
	if err != nil {
		return nil, err
	}

	// Get daily balances for this specific entry's account and member profile
	dailyBalances, err := m.GetDailyEndingBalances(
		context,
		generatedSavingsInterest.NewComputationDate,
		generatedSavingsInterest.LastComputationDate,
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
	var totalBalanceSum float64
	var totalDays int
	var lowestBalance float64 = -1 // Use -1 to indicate not set
	var highestBalance float64
	var beginningBalance float64 = -1 // Use -1 to indicate not set
	var endingBalance float64

	// Convert daily balances to response format
	currentDate := generatedSavingsInterest.NewComputationDate
	var previousBalance float64 = -1 // Use -1 to indicate first entry
	for i, balance := range dailyBalances {
		dateStr := currentDate.AddDate(0, 0, i).Format("2006-01-02")
		var changeType string
		if previousBalance == -1 {
			changeType = "no_change"
		} else if m.provider.Service.Decimal.IsGreaterThan(balance, previousBalance) {
			changeType = "increase"
		} else if m.provider.Service.Decimal.IsLessThan(balance, previousBalance) {
			changeType = "decrease"
		} else {
			changeType = "no_change"
		}

		allDailyBalances = append(allDailyBalances, GeneratedSavingsInterestEntryDailyBalance{
			Balance: balance,
			Date:    dateStr,
			Type:    changeType,
		})

		// Update statistics using decimal operations for precision
		totalBalanceSum = m.provider.Service.Decimal.Add(totalBalanceSum, balance)
		totalDays++

		// Track beginning balance (first balance encountered)
		if beginningBalance == -1 {
			beginningBalance = balance
		}
		// Track ending balance (last balance will be the ending balance)
		endingBalance = balance

		if lowestBalance == -1 || m.provider.Service.Decimal.IsLessThan(balance, lowestBalance) {
			lowestBalance = balance
		}
		if m.provider.Service.Decimal.IsGreaterThan(balance, highestBalance) {
			highestBalance = balance
		}

		// Update previous balance for next iteration
		previousBalance = balance
	}

	// Calculate average daily balance using decimal operations
	averageDailyBalance := float64(0)
	if totalDays > 0 {
		averageDailyBalance = m.provider.Service.Decimal.Divide(totalBalanceSum, float64(totalDays))
	}

	if beginningBalance == -1 {
		beginningBalance = 0
	}

	account, err := m.AccountManager.GetByID(context, generatedSavingsInterestEntry.AccountID, "Currency")
	if err != nil {
		return nil, err
	}
	memberProfile, err := m.MemberProfileManager.GetByID(context, generatedSavingsInterestEntry.MemberProfileID)
	if err != nil {
		return nil, err
	}

	return &GeneratedSavingsInterestEntryDailyBalanceResponse{
		BeginningBalance:    beginningBalance,
		EndingBalance:       endingBalance,
		AverageDailyBalance: averageDailyBalance,
		LowestBalance:       lowestBalance,
		HighestBalance:      highestBalance,
		DailyBalance:        allDailyBalances,
		Account:             m.AccountManager.ToModel(account),
		MemberProfile:       m.MemberProfileManager.ToModel(memberProfile),
	}, nil
}
