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
	// MemberTypeReference represents the MemberTypeReference model.
	MemberTypeReference struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id,omitempty"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type_reference" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type_reference" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid" json:"account_id"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`

		MemberTypeID uuid.UUID   `gorm:"type:uuid" json:"member_type_id"`
		MemberType   *MemberType `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_type,omitempty"`

		MaintainingBalance                             float64 `gorm:"type:decimal;default:0" json:"maintaining_balance"`
		Description                                    string  `gorm:"type:text" json:"description"`
		InterestRate                                   float64 `gorm:"type:decimal;default:0" json:"interest_rate"`
		MinimumBalance                                 float64 `gorm:"type:decimal;default:0" json:"minimum_balance"`
		Charges                                        float64 `gorm:"type:decimal;default:0" json:"charges"`
		ActiveMemberMinimumBalance                     float64 `gorm:"type:decimal;default:0" json:"active_member_minimum_balance"`
		ActiveMemberRatio                              float64 `gorm:"type:decimal;default:0" json:"active_member_ratio"`
		OtherInterestOnSavingComputationMinimumBalance float64 `gorm:"type:decimal;default:0" json:"other_interest_on_saving_computation_minimum_balance"`
		OtherInterestOnSavingComputationInterestRate   float64 `gorm:"type:decimal;default:0" json:"other_interest_on_saving_computation_interest_rate"`
	}

	// MemberTypeReferenceResponse represents the response structure for membertypereference data

	// MemberTypeReferenceResponse represents the response structure for MemberTypeReference.
	MemberTypeReferenceResponse struct {
		ID                                             uuid.UUID             `json:"id"`
		CreatedAt                                      string                `json:"created_at"`
		CreatedByID                                    uuid.UUID             `json:"created_by_id"`
		CreatedBy                                      *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt                                      string                `json:"updated_at"`
		UpdatedByID                                    uuid.UUID             `json:"updated_by_id"`
		UpdatedBy                                      *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID                                 uuid.UUID             `json:"organization_id"`
		Organization                                   *OrganizationResponse `json:"organization,omitempty"`
		BranchID                                       uuid.UUID             `json:"branch_id"`
		Branch                                         *BranchResponse       `json:"branch,omitempty"`
		AccountID                                      uuid.UUID             `json:"account_id"`
		Account                                        *AccountResponse      `json:"account,omitempty"`
		MemberTypeID                                   uuid.UUID             `json:"member_type_id"`
		MemberType                                     *MemberTypeResponse   `json:"member_type,omitempty"`
		MaintainingBalance                             float64               `json:"maintaining_balance"`
		Description                                    string                `json:"description"`
		InterestRate                                   float64               `json:"interest_rate"`
		MinimumBalance                                 float64               `json:"minimum_balance"`
		Charges                                        float64               `json:"charges"`
		ActiveMemberMinimumBalance                     float64               `json:"active_member_minimum_balance"`
		ActiveMemberRatio                              float64               `json:"active_member_ratio"`
		OtherInterestOnSavingComputationMinimumBalance float64               `json:"other_interest_on_saving_computation_minimum_balance"`
		OtherInterestOnSavingComputationInterestRate   float64               `json:"other_interest_on_saving_computation_interest_rate"`
	}

	// MemberTypeReferenceRequest represents the request structure for creating/updating membertypereference

	// MemberTypeReferenceRequest represents the request structure for MemberTypeReference.
	MemberTypeReferenceRequest struct {
		AccountID                                      uuid.UUID `json:"account_id"`
		MemberTypeID                                   uuid.UUID `json:"member_type_id"`
		MaintainingBalance                             float64   `json:"maintaining_balance,omitempty"`
		Description                                    string    `json:"description,omitempty"`
		InterestRate                                   float64   `json:"interest_rate,omitempty"`
		MinimumBalance                                 float64   `json:"minimum_balance,omitempty"`
		Charges                                        float64   `json:"charges,omitempty"`
		ActiveMemberMinimumBalance                     float64   `json:"active_member_minimum_balance,omitempty"`
		ActiveMemberRatio                              float64   `json:"active_member_ratio,omitempty"`
		OtherInterestOnSavingComputationMinimumBalance float64   `json:"other_interest_on_saving_computation_minimum_balance,omitempty"`
		OtherInterestOnSavingComputationInterestRate   float64   `json:"other_interest_on_saving_computation_interest_rate,omitempty"`
	}
)

func (m *Core) memberTypeReference() {
	m.Migration = append(m.Migration, &MemberTypeReference{})
	m.MemberTypeReferenceManager = *registry.NewRegistry(registry.RegistryParams[
		MemberTypeReference, MemberTypeReferenceResponse, MemberTypeReferenceRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Account", "MemberType",
		},
		Service: m.provider.Service,
		Resource: func(data *MemberTypeReference) *MemberTypeReferenceResponse {
			if data == nil {
				return nil
			}
			return &MemberTypeReferenceResponse{
				ID:                         data.ID,
				CreatedAt:                  data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                data.CreatedByID,
				CreatedBy:                  m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                  data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                data.UpdatedByID,
				UpdatedBy:                  m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:             data.OrganizationID,
				Organization:               m.OrganizationManager.ToModel(data.Organization),
				BranchID:                   data.BranchID,
				Branch:                     m.BranchManager.ToModel(data.Branch),
				AccountID:                  data.AccountID,
				Account:                    m.AccountManager.ToModel(data.Account),
				MemberTypeID:               data.MemberTypeID,
				MemberType:                 m.MemberTypeManager.ToModel(data.MemberType),
				MaintainingBalance:         data.MaintainingBalance,
				Description:                data.Description,
				InterestRate:               data.InterestRate,
				MinimumBalance:             data.MinimumBalance,
				Charges:                    data.Charges,
				ActiveMemberMinimumBalance: data.ActiveMemberMinimumBalance,
				ActiveMemberRatio:          data.ActiveMemberRatio,
				OtherInterestOnSavingComputationMinimumBalance: data.OtherInterestOnSavingComputationMinimumBalance,
				OtherInterestOnSavingComputationInterestRate:   data.OtherInterestOnSavingComputationInterestRate,
			}
		},

		Created: func(data *MemberTypeReference) []string {
			return []string{
				"member_type_reference.create",
				fmt.Sprintf("member_type_reference.create.%s", data.ID),
				fmt.Sprintf("member_type_reference.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_reference.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_type_reference.create.member_type.%s", data.MemberTypeID),
			}
		},
		Updated: func(data *MemberTypeReference) []string {
			return []string{
				"member_type_reference.update",
				fmt.Sprintf("member_type_reference.update.%s", data.ID),
				fmt.Sprintf("member_type_reference.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_reference.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_type_reference.update.member_type.%s", data.MemberTypeID),
			}
		},
		Deleted: func(data *MemberTypeReference) []string {
			return []string{
				"member_type_reference.delete",
				fmt.Sprintf("member_type_reference.delete.%s", data.ID),
				fmt.Sprintf("member_type_reference.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_reference.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_type_reference.update.member_type.%s", data.MemberTypeID),
			}
		},
	})
}

// MemberTypeReferenceCurrentBranch returns MemberTypeReferenceCurrentBranch for the current branch or organization where applicable.
func (m *Core) MemberTypeReferenceCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberTypeReference, error) {
	return m.MemberTypeReferenceManager.Find(context, &MemberTypeReference{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
