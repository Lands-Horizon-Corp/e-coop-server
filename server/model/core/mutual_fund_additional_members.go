package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// MutualFundAdditionalMembers represents the MutualFundAdditionalMembers model.
	MutualFundAdditionalMembers struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_mutual_fund_additional" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_mutual_fund_additional" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MutualFundID uuid.UUID   `gorm:"type:uuid;not null;index:idx_mutual_fund_additional_members" json:"mutual_fund_id"`
		MutualFund   *MutualFund `gorm:"foreignKey:MutualFundID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"mutual_fund,omitempty"`

		MemberTypeID    uuid.UUID   `gorm:"type:uuid;not null;index:idx_member_type_additional" json:"member_type_id"`
		MemberType      *MemberType `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"member_type,omitempty"`
		NumberOfMembers int         `gorm:"not null" json:"number_of_members"`
		Ratio           float64     `gorm:"type:decimal(15,4);not null" json:"ratio"`
	}

	// MutualFundAdditionalMembersResponse represents the response structure for MutualFundAdditionalMembers.
	MutualFundAdditionalMembersResponse struct {
		ID              uuid.UUID             `json:"id"`
		CreatedAt       string                `json:"created_at"`
		CreatedByID     uuid.UUID             `json:"created_by_id"`
		CreatedBy       *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt       string                `json:"updated_at"`
		UpdatedByID     uuid.UUID             `json:"updated_by_id"`
		UpdatedBy       *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID  uuid.UUID             `json:"organization_id"`
		Organization    *OrganizationResponse `json:"organization,omitempty"`
		BranchID        uuid.UUID             `json:"branch_id"`
		Branch          *BranchResponse       `json:"branch,omitempty"`
		MutualFundID    uuid.UUID             `json:"mutual_fund_id"`
		MutualFund      *MutualFundResponse   `json:"mutual_fund,omitempty"`
		MemberTypeID    uuid.UUID             `json:"member_type_id"`
		MemberType      *MemberTypeResponse   `json:"member_type,omitempty"`
		NumberOfMembers int                   `json:"number_of_members"`
		Ratio           float64               `json:"ratio"`
	}

	// MutualFundAdditionalMembersRequest represents the request structure for MutualFundAdditionalMembers.
	MutualFundAdditionalMembersRequest struct {
		ID              *uuid.UUID `json:"id,omitempty"`
		MutualFundID    uuid.UUID  `json:"mutual_fund_id" validate:"required"`
		MemberTypeID    uuid.UUID  `json:"member_type_id" validate:"required"`
		NumberOfMembers int        `json:"number_of_members" validate:"required,min=1"`
		Ratio           float64    `json:"ratio" validate:"required,min=0,max=100"`
	}
)

func (m *Core) mutualFundAdditionalMembers() {
	m.Migration = append(m.Migration, &MutualFundAdditionalMembers{})
	m.MutualFundAdditionalMembersManager = *registry.NewRegistry(registry.RegistryParams[MutualFundAdditionalMembers, MutualFundAdditionalMembersResponse, MutualFundAdditionalMembersRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Organization", "Branch", "MutualFund", "MemberType"},
		Database: m.provider.Service.Database.Client(),
Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		}
		Resource: func(data *MutualFundAdditionalMembers) *MutualFundAdditionalMembersResponse {
			if data == nil {
				return nil
			}
			return &MutualFundAdditionalMembersResponse{
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
				MutualFundID:    data.MutualFundID,
				MutualFund:      m.MutualFundManager.ToModel(data.MutualFund),
				MemberTypeID:    data.MemberTypeID,
				MemberType:      m.MemberTypeManager.ToModel(data.MemberType),
				NumberOfMembers: data.NumberOfMembers,
				Ratio:           data.Ratio,
			}
		},
		Created: func(data *MutualFundAdditionalMembers) registry.Topics {
			return []string{
				"mutual_fund_additional_members.create",
				fmt.Sprintf("mutual_fund_additional_members.create.%s", data.ID),
				fmt.Sprintf("mutual_fund_additional_members.create.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_additional_members.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_additional_members.create.mutual_fund.%s", data.MutualFundID),
				fmt.Sprintf("mutual_fund_additional_members.create.member_type.%s", data.MemberTypeID),
			}
		},
		Updated: func(data *MutualFundAdditionalMembers) registry.Topics {
			return []string{
				"mutual_fund_additional_members.update",
				fmt.Sprintf("mutual_fund_additional_members.update.%s", data.ID),
				fmt.Sprintf("mutual_fund_additional_members.update.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_additional_members.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_additional_members.update.mutual_fund.%s", data.MutualFundID),
				fmt.Sprintf("mutual_fund_additional_members.update.member_type.%s", data.MemberTypeID),
			}
		},
		Deleted: func(data *MutualFundAdditionalMembers) registry.Topics {
			return []string{
				"mutual_fund_additional_members.delete",
				fmt.Sprintf("mutual_fund_additional_members.delete.%s", data.ID),
				fmt.Sprintf("mutual_fund_additional_members.delete.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_additional_members.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_additional_members.delete.mutual_fund.%s", data.MutualFundID),
				fmt.Sprintf("mutual_fund_additional_members.delete.member_type.%s", data.MemberTypeID),
			}
		},
	})
}

// MutualFundAdditionalMembersCurrentBranch retrieves all mutual fund additional members associated with the specified organization and branch.
func (m *Core) MutualFundAdditionalMembersCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MutualFundAdditionalMembers, error) {
	return m.MutualFundAdditionalMembersManager.Find(context, &MutualFundAdditionalMembers{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

// MutualFundAdditionalMembersByMutualFund retrieves all additional members for a specific mutual fund.
func (m *Core) MutualFundAdditionalMembersByMutualFund(context context.Context, mutualFundID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*MutualFundAdditionalMembers, error) {
	return m.MutualFundAdditionalMembersManager.Find(context, &MutualFundAdditionalMembers{
		MutualFundID:   mutualFundID,
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
