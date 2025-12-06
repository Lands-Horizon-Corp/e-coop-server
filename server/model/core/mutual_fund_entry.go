package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// MutualFundEntry represents the MutualFundEntry model.
	MutualFundEntry struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_mutual_fund_entry" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_mutual_fund_entry" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MutualFundID uuid.UUID   `gorm:"type:uuid;not null;index:idx_mutual_fund_entry_mutual_fund" json:"mutual_fund_id"`
		MutualFund   *MutualFund `gorm:"foreignKey:MutualFundID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"mutual_fund,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null;index:idx_mutual_fund_entry_member" json:"member_profile_id"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid;not null;index:idx_mutual_fund_entry_account" json:"account_id"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"account,omitempty"`

		Amount float64 `gorm:"type:decimal(15,4);not null" json:"amount"`
	}

	// MutualFundEntryResponse represents the response structure for MutualFundEntry.
	MutualFundEntryResponse struct {
		ID              uuid.UUID              `json:"id"`
		CreatedAt       string                 `json:"created_at"`
		CreatedByID     uuid.UUID              `json:"created_by_id"`
		CreatedBy       *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt       string                 `json:"updated_at"`
		UpdatedByID     uuid.UUID              `json:"updated_by_id"`
		UpdatedBy       *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID  uuid.UUID              `json:"organization_id"`
		Organization    *OrganizationResponse  `json:"organization,omitempty"`
		BranchID        uuid.UUID              `json:"branch_id"`
		Branch          *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID uuid.UUID              `json:"member_profile_id"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
		AccountID       uuid.UUID              `json:"account_id"`
		Account         *AccountResponse       `json:"account,omitempty"`
		Amount          float64                `json:"amount"`
		MutualFundID    uuid.UUID              `json:"mutual_fund_id"`
		MutualFund      *MutualFundResponse    `json:"mutual_fund,omitempty"`
	}

	// MutualFundEntryRequest represents the request structure for MutualFundEntry.
	MutualFundEntryRequest struct {
		ID              *uuid.UUID `json:"id,omitempty"`
		MemberProfileID uuid.UUID  `json:"member_profile_id" validate:"required"`
		AccountID       uuid.UUID  `json:"account_id" validate:"required"`
		Amount          float64    `json:"amount" validate:"required,gte=0"`
	}
)

func (m *Core) mutualFundEntry() {
	m.Migration = append(m.Migration, &MutualFundEntry{})
	m.MutualFundEntryManager = *registry.NewRegistry(registry.RegistryParams[MutualFundEntry, MutualFundEntryResponse, MutualFundEntryRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
			"MemberProfile", "Account", "MutualFund"},
		Service: m.provider.Service,
		Resource: func(data *MutualFundEntry) *MutualFundEntryResponse {
			if data == nil {
				return nil
			}
			return &MutualFundEntryResponse{
				ID:              data.ID,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:     data.CreatedByID,
				CreatedBy:       m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:     data.UpdatedByID,
				UpdatedBy:       m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:  data.OrganizationID,
				Organization:    m.OrganizationManager.ToModel(data.Organization),
				BranchID:        data.BranchID,
				Branch:          m.BranchManager.ToModel(data.Branch),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   m.MemberProfileManager.ToModel(data.MemberProfile),
				AccountID:       data.AccountID,
				Account:         m.AccountManager.ToModel(data.Account),
				Amount:          data.Amount,
				MutualFundID:    data.MutualFundID,
				MutualFund:      m.MutualFundManager.ToModel(data.MutualFund),
			}
		},
		Created: func(data *MutualFundEntry) []string {
			return []string{
				"mutual_fund_entry.create",
				fmt.Sprintf("mutual_fund_entry.create.%s", data.ID),
				fmt.Sprintf("mutual_fund_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_entry.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_entry.create.member.%s", data.MemberProfileID),
				fmt.Sprintf("mutual_fund_entry.create.account.%s", data.AccountID),
			}
		},
		Updated: func(data *MutualFundEntry) []string {
			return []string{
				"mutual_fund_entry.update",
				fmt.Sprintf("mutual_fund_entry.update.%s", data.ID),
				fmt.Sprintf("mutual_fund_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_entry.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_entry.update.member.%s", data.MemberProfileID),
				fmt.Sprintf("mutual_fund_entry.update.account.%s", data.AccountID),
			}
		},
		Deleted: func(data *MutualFundEntry) []string {
			return []string{
				"mutual_fund_entry.delete",
				fmt.Sprintf("mutual_fund_entry.delete.%s", data.ID),
				fmt.Sprintf("mutual_fund_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_entry.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_entry.delete.member.%s", data.MemberProfileID),
				fmt.Sprintf("mutual_fund_entry.delete.account.%s", data.AccountID),
			}
		},
	})
}

// MutualFundEntryCurrentBranch retrieves all mutual fund entries associated with the specified organization and branch.
func (m *Core) MutualFundEntryCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MutualFundEntry, error) {
	return m.MutualFundEntryManager.Find(context, &MutualFundEntry{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

// MutualFundEntryByMember retrieves all mutual fund entries for a specific member profile.
func (m *Core) MutualFundEntryByMember(context context.Context, memberProfileID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*MutualFundEntry, error) {
	return m.MutualFundEntryManager.Find(context, &MutualFundEntry{
		MemberProfileID: memberProfileID,
		OrganizationID:  organizationID,
		BranchID:        branchID,
	})
}

// MutualFundEntryByAccount retrieves all mutual fund entries for a specific account.
func (m *Core) MutualFundEntryByAccount(context context.Context, accountID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*MutualFundEntry, error) {
	return m.MutualFundEntryManager.Find(context, &MutualFundEntry{
		AccountID:      accountID,
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
