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
	ComputationTypeContinuous MutualFundComputationType = "continuous"
	ComputationTypeUpToZero   MutualFundComputationType = "up_to_zero"
	ComputationTypeSufficient MutualFundComputationType = "sufficient"

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

		MemberTypeID *uuid.UUID  `gorm:"type:uuid" json:"member_type_id,omitempty"`
		MemberType   *MemberType `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:SET NULL;" json:"member_type,omitempty"`

		AdditionalMembers []*MutualFundAdditionalMembers `gorm:"foreignKey:MutualFundID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"additional_members,omitempty"`
		MutualFundTables  []*MutualFundTable             `gorm:"foreignKey:MutualFundID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"mutual_fund_tables,omitempty"`

		Name            string                    `gorm:"type:varchar(255);not null" json:"name"`
		Description     string                    `gorm:"type:text" json:"description"`
		DateOfDeath     time.Time                 `gorm:"not null" json:"date_of_death"`
		ExtensionOnly   bool                      `gorm:"not null;default:false" json:"extension_only"`
		Amount          float64                   `gorm:"type:decimal;not null" json:"amount"`
		ComputationType MutualFundComputationType `gorm:"type:varchar(50);not null" json:"computation_type"`

		TotalAmount float64 `gorm:"type:decimal;default:0" json:"total_amount,omitempty"`

		AccountID *uuid.UUID `gorm:"type:uuid" json:"account_id,omitempty"`
		Account   *Account   `gorm:"foreignKey:AccountID;constraint:OnDelete:SET NULL;" json:"account,omitempty"`

		PostAccountID *uuid.UUID `json:"post_account_id,omitempty"`
		PostAccount   *Account   `json:"post_account,omitempty"`

		PrintedByUserID *uuid.UUID `gorm:"type:uuid"`
		PrintedByUser   *User      `gorm:"foreignKey:PrintedByUserID;constraint:OnDelete:SET NULL;" json:"printed_by_user,omitempty"`
		PrintedDate     *time.Time `gorm:"" json:"printed_date,omitempty"`

		PostedDate     *time.Time `json:"posted_date,omitempty"`
		PostedByUserID *uuid.UUID `gorm:"type:uuid" json:"posted_by_user_id,omitempty"`
		PostedByUser   *User      `gorm:"foreignKey:PostedByUserID;constraint:OnDelete:SET NULL;" json:"posted_by_user,omitempty"`
	}

	// MutualFundResponse represents the response structure for mutual fund data
	MutualFundResponse struct {
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

		MemberTypeID *uuid.UUID          `json:"member_type_id,omitempty"`
		MemberType   *MemberTypeResponse `json:"member_type,omitempty"`

		AdditionalMembers []*MutualFundAdditionalMembersResponse `json:"additional_members,omitempty"`
		MutualFundTables  []*MutualFundTableResponse             `json:"mutual_fund_tables,omitempty"`
		Name              string                                 `json:"name"`
		Description       string                                 `json:"description"`
		DateOfDeath       string                                 `json:"date_of_death"`
		ExtensionOnly     bool                                   `json:"extension_only"`
		Amount            float64                                `json:"amount"`
		ComputationType   MutualFundComputationType              `json:"computation_type"`
		AccountID         *uuid.UUID                             `json:"account_id,omitempty"`
		Account           *Account                               `json:"account,omitempty"`

		PrintedByUserID *uuid.UUID    `json:"printed_by_user_id,omitempty"`
		PrintedByUser   *UserResponse `json:"printed_by_user,omitempty"`
		PrintedDate     *string       `json:"printed_date,omitempty"`

		PostedDate     *time.Time    `json:"posted_date,omitempty"`
		PostAccountID  *uuid.UUID    `json:"post_account_id,omitempty"`
		PostAccount    *Account      `json:"post_account,omitempty"`
		PostedByUserID *uuid.UUID    `json:"posted_by_user_id,omitempty"`
		PostedByUser   *UserResponse `json:"posted_by_user,omitempty"`
	}

	// MutualFundRequest represents the request structure for creating/updating mutual fund
	MutualFundRequest struct {
		MemberProfileID uuid.UUID                 `json:"member_profile_id" validate:"required"`
		MemberTypeID    *uuid.UUID                `json:"member_type_id,omitempty"`
		Name            string                    `json:"name" validate:"required,min=1,max=255"`
		Description     string                    `json:"description,omitempty"`
		DateOfDeath     time.Time                 `json:"date_of_death" validate:"required"`
		ExtensionOnly   bool                      `json:"extension_only"`
		Amount          float64                   `json:"amount" validate:"required,gte=0"`
		ComputationType MutualFundComputationType `json:"computation_type" validate:"required"`

		MutualFundAdditionalMembers []MutualFundAdditionalMembersRequest `json:"mutual_fund_additional_members,omitempty" validate:"dive"`
		MutualFundTables            []MutualFundTableRequest             `json:"mutual_fund_tables,omitempty" validate:"dive"`

		MutualFundAdditionalMembersDeleteIDs uuid.UUIDs `json:"mutual_fund_additional_members_delete_ids,omitempty" validate:"dive"`
		MutualFundTableDeleteIDs             uuid.UUIDs `json:"mutual_fund_table_delete_ids,omitempty" validate:"dive"`
		AccountID                            *uuid.UUID `json:"account_id,omitempty"`
	}

	MutualFundView struct {
		TotalAmount       float64                    `json:"total_amount"`
		MutualFundEntries []*MutualFundEntryResponse `json:"mutual_fund_entries"`
	}
	MutualFundViewPostRequest struct {
		CheckVoucherNumber *string    `json:"check_voucher_number"`
		PostAccountID      *uuid.UUID `json:"post_account_id"`
		EntryDate          *time.Time `json:"entry_date"`
	}
)

func (m *Core) mutualFund() {
	m.Migration = append(m.Migration, &MutualFund{})
	m.MutualFundManager = *registry.NewRegistry(registry.RegistryParams[MutualFund, MutualFundResponse, MutualFundRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"MemberProfile",
			"MemberType",
			"AdditionalMembers",
			"AdditionalMembers.MemberType",
			"MutualFundTables",
			"Account"},
		Service: m.provider.Service,
		Resource: func(data *MutualFund) *MutualFundResponse {
			if data == nil {
				return nil
			}
			var printedDate *string
			if data.PrintedDate != nil {
				formatted := data.PrintedDate.Format(time.RFC3339)
				printedDate = &formatted
			}
			return &MutualFundResponse{
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

				MemberTypeID: data.MemberTypeID,
				MemberType:   m.MemberTypeManager.ToModel(data.MemberType),

				AdditionalMembers: m.MutualFundAdditionalMembersManager.ToModels(data.AdditionalMembers),
				MutualFundTables:  m.MutualFundTableManager.ToModels(data.MutualFundTables),
				Name:              data.Name,
				Description:       data.Description,
				DateOfDeath:       data.DateOfDeath.Format(time.RFC3339),
				ExtensionOnly:     data.ExtensionOnly,
				Amount:            data.Amount,
				ComputationType:   data.ComputationType,
				AccountID:         data.AccountID,
				Account:           data.Account,

				PrintedByUserID: data.PrintedByUserID,
				PrintedByUser:   m.UserManager.ToModel(data.PrintedByUser),
				PrintedDate:     printedDate,

				PostAccountID:  data.PostAccountID,
				PostAccount:    data.PostAccount,
				PostedDate:     data.PostedDate,
				PostedByUserID: data.PostedByUserID,
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

// CreateMutualFundValue creates a mutual fund object with additional members and tables without saving to database.
func (m *Core) CreateMutualFundValue(
	context context.Context,
	req *MutualFundRequest, userOrg *UserOrganization) *MutualFund {
	now := time.Now().UTC()

	// Create additional members objects
	var additionalMembers []*MutualFundAdditionalMembers
	for _, additionalMember := range req.MutualFundAdditionalMembers {
		additionalMemberData := &MutualFundAdditionalMembers{
			ID:              uuid.New(),
			MemberTypeID:    additionalMember.MemberTypeID,
			NumberOfMembers: additionalMember.NumberOfMembers,
			Ratio:           additionalMember.Ratio,
			CreatedAt:       now,
			CreatedByID:     userOrg.UserID,
			UpdatedAt:       now,
			UpdatedByID:     userOrg.UserID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
		}
		additionalMembers = append(additionalMembers, additionalMemberData)
	}

	// Create mutual fund tables objects
	var mutualFundTables []*MutualFundTable
	for _, mutualFundTable := range req.MutualFundTables {
		mutualFundTableData := &MutualFundTable{
			ID:             uuid.New(),
			MonthFrom:      mutualFundTable.MonthFrom,
			MonthTo:        mutualFundTable.MonthTo,
			Amount:         mutualFundTable.Amount,
			CreatedAt:      now,
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      now,
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}
		mutualFundTables = append(mutualFundTables, mutualFundTableData)
	}

	// Create the mutual fund object
	mutualFund := &MutualFund{
		ID:                uuid.New(),
		MemberProfileID:   req.MemberProfileID,
		MemberTypeID:      req.MemberTypeID,
		Name:              req.Name,
		Description:       req.Description,
		DateOfDeath:       req.DateOfDeath,
		ExtensionOnly:     req.ExtensionOnly,
		Amount:            req.Amount,
		ComputationType:   req.ComputationType,
		AccountID:         req.AccountID,
		CreatedAt:         now,
		CreatedByID:       userOrg.UserID,
		UpdatedAt:         now,
		UpdatedByID:       userOrg.UserID,
		BranchID:          *userOrg.BranchID,
		OrganizationID:    userOrg.OrganizationID,
		AdditionalMembers: additionalMembers,
		MutualFundTables:  mutualFundTables,
	}

	// Set the MutualFundID for the child objects
	for _, additionalMember := range additionalMembers {
		additionalMember.MutualFundID = mutualFund.ID
	}
	for _, mutualFundTable := range mutualFundTables {
		mutualFundTable.MutualFundID = mutualFund.ID
	}

	return mutualFund
}
