package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ComputationTypeContinuous MutualFundComputationType = "continuous"
	ComputationTypeUpToZero   MutualFundComputationType = "up_to_zero"
	ComputationTypeSufficient MutualFundComputationType = "sufficient"

	ComputationTypeByMemberClassAmount MutualFundComputationType = "by_member_class_amount"
	ComputationTypeByMembershipYear    MutualFundComputationType = "by_membership_year"
)

type (
	MutualFundComputationType string
	MutualFund                struct {
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
		PrintedDate     *time.Time `gorm:"default:NULL" json:"printed_date,omitempty"`

		PostedDate     *time.Time `json:"posted_date,omitempty"`
		PostedByUserID *uuid.UUID `gorm:"type:uuid" json:"posted_by_user_id,omitempty"`
		PostedByUser   *User      `gorm:"foreignKey:PostedByUserID;constraint:OnDelete:SET NULL;" json:"posted_by_user,omitempty"`
	}

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
		MutualFund        *MutualFundResponse        `json:"mutual_fund"`
	}
	MutualFundViewPostRequest struct {
		CheckVoucherNumber *string    `json:"check_voucher_number"`
		PostAccountID      *uuid.UUID `json:"post_account_id"`
		EntryDate          *time.Time `json:"entry_date"`
	}
)
