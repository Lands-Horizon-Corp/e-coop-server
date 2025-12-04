package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ComputationType represents the mutual fund computation type
type MutualFundComputationType string

// Computation type constants
const (
	ComputationTypeTotalAmount         MutualFundComputationType = "total_amount"
	ComputationTypeByAmount            MutualFundComputationType = "by_amount"
	ComputationTypeByMemberAmount      MutualFundComputationType = "by_member_amount"
	ComputationTypeByMemberClassAmount MutualFundComputationType = "by_member_class_amount"
	ComputationTypeByMembershipYear    MutualFundComputationType = "by_membership_year"
)

type (
	// MutualFund represents the MutualFund model.
	MutualFund struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_mutual_fund" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_mutual_fund" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null" json:"member_profile_id"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		MutualAidContributionID *uuid.UUID             `gorm:"type:uuid;index:idx_mutual_fund_aid_contribution" json:"mutual_aid_contribution_id"`
		MutualAidContribution   *MutualAidContribution `gorm:"foreignKey:MutualAidContributionID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"mutual_aid_contribution,omitempty"`

		// One-to-many relationship: one mutual fund can have many additional members
		AdditionalMembers []*MutualFundAdditionalMembers `gorm:"foreignKey:MutualFundID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"additional_members,omitempty"`

		Name            string                    `gorm:"type:varchar(255);not null" json:"name"`
		Description     string                    `gorm:"type:text" json:"description"`
		DateOfDeath     time.Time                 `gorm:"not null" json:"date_of_death"`
		ExtensionOnly   bool                      `gorm:"not null;default:false" json:"extension_only"`
		Amount          float64                   `gorm:"type:decimal;not null" json:"amount"`
		ComputationType MutualFundComputationType `gorm:"type:varchar(50);not null" json:"computation_type"`
	}

	// MutualFundResponse represents the response structure for mutual fund data
	MutualFundResponse struct {
		ID                      uuid.UUID                              `json:"id"`
		CreatedAt               string                                 `json:"created_at"`
		CreatedByID             uuid.UUID                              `json:"created_by_id"`
		CreatedBy               *UserResponse                          `json:"created_by,omitempty"`
		UpdatedAt               string                                 `json:"updated_at"`
		UpdatedByID             uuid.UUID                              `json:"updated_by_id"`
		UpdatedBy               *UserResponse                          `json:"updated_by,omitempty"`
		OrganizationID          uuid.UUID                              `json:"organization_id"`
		Organization            *OrganizationResponse                  `json:"organization,omitempty"`
		BranchID                uuid.UUID                              `json:"branch_id"`
		Branch                  *BranchResponse                        `json:"branch,omitempty"`
		MemberProfileID         uuid.UUID                              `json:"member_profile_id"`
		MemberProfile           *MemberProfileResponse                 `json:"member_profile,omitempty"`
		MutualAidContributionID *uuid.UUID                             `json:"mutual_aid_contribution_id,omitempty"`
		MutualAidContribution   *MutualAidContributionResponse         `json:"mutual_aid_contribution,omitempty"`
		AdditionalMembers       []*MutualFundAdditionalMembersResponse `json:"additional_members,omitempty"`
		Name                    string                                 `json:"name"`
		Description             string                                 `json:"description"`
		DateOfDeath             string                                 `json:"date_of_death"`
		ExtensionOnly           bool                                   `json:"extension_only"`
		Amount                  float64                                `json:"amount"`
		ComputationType         MutualFundComputationType              `json:"computation_type"`
	}

	// MutualFundRequest represents the request structure for creating/updating mutual fund
	MutualFundRequest struct {
		MemberProfileID         uuid.UUID                 `json:"member_profile_id" validate:"required"`
		MutualAidContributionID *uuid.UUID                `json:"mutual_aid_contribution_id,omitempty"`
		Name                    string                    `json:"name" validate:"required,min=1,max=255"`
		Description             string                    `json:"description,omitempty"`
		DateOfDeath             time.Time                 `json:"date_of_death" validate:"required"`
		ExtensionOnly           bool                      `json:"extension_only"`
		Amount                  float64                   `json:"amount" validate:"required,gte=0"`
		ComputationType         MutualFundComputationType `json:"computation_type" validate:"required"`
	}
)

func (m *Core) mutualFund() {
	m.Migration = append(m.Migration, &MutualFund{})
	m.MutualFundManager = *registry.NewRegistry(registry.RegistryParams[MutualFund, MutualFundResponse, MutualFundRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Organization", "Branch", "MemberProfile", "MutualAidContribution", "AdditionalMembers", "AdditionalMembers.MemberType"},
		Service:  m.provider.Service,
		Resource: func(data *MutualFund) *MutualFundResponse {
			if data == nil {
				return nil
			}
			return &MutualFundResponse{
				ID:                      data.ID,
				CreatedAt:               data.CreatedAt.Format(time.RFC3339),
				CreatedByID:             data.CreatedByID,
				CreatedBy:               m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:               data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:             data.UpdatedByID,
				UpdatedBy:               m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:          data.OrganizationID,
				Organization:            m.OrganizationManager.ToModel(data.Organization),
				BranchID:                data.BranchID,
				Branch:                  m.BranchManager.ToModel(data.Branch),
				MemberProfileID:         data.MemberProfileID,
				MemberProfile:           m.MemberProfileManager.ToModel(data.MemberProfile),
				MutualAidContributionID: data.MutualAidContributionID,
				MutualAidContribution:   m.MutualAidContributionManager.ToModel(data.MutualAidContribution),
				AdditionalMembers:       m.MutualFundAdditionalMembersManager.ToModels(data.AdditionalMembers),
				Name:                    data.Name,
				Description:             data.Description,
				DateOfDeath:             data.DateOfDeath.Format(time.RFC3339),
				ExtensionOnly:           data.ExtensionOnly,
				Amount:                  data.Amount,
				ComputationType:         data.ComputationType,
			}
		},
		Created: func(data *MutualFund) []string {
			return []string{
				"mutual_fund.create",
				fmt.Sprintf("mutual_fund.create.%s", data.ID),
				fmt.Sprintf("mutual_fund.create.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund.create.member.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *MutualFund) []string {
			return []string{
				"mutual_fund.update",
				fmt.Sprintf("mutual_fund.update.%s", data.ID),
				fmt.Sprintf("mutual_fund.update.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund.update.member.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *MutualFund) []string {
			return []string{
				"mutual_fund.delete",
				fmt.Sprintf("mutual_fund.delete.%s", data.ID),
				fmt.Sprintf("mutual_fund.delete.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund.delete.member.%s", data.MemberProfileID),
			}
		},
	})
}

// MutualFundCurrentBranch retrieves all mutual funds associated with the specified organization and branch.
func (m *Core) MutualFundCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MutualFund, error) {
	return m.MutualFundManager.Find(context, &MutualFund{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

// MutualFundByMember retrieves all mutual funds for a specific member profile.
func (m *Core) MutualFundByMember(context context.Context, memberProfileID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*MutualFund, error) {
	return m.MutualFundManager.Find(context, &MutualFund{
		MemberProfileID: memberProfileID,
		OrganizationID:  organizationID,
		BranchID:        branchID,
	})
}
