package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
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

	MutualFundAdditionalMembersRequest struct {
		ID              *uuid.UUID `json:"id,omitempty"`
		MutualFundID    uuid.UUID  `json:"mutual_fund_id" validate:"required"`
		MemberTypeID    uuid.UUID  `json:"member_type_id" validate:"required"`
		NumberOfMembers int        `json:"number_of_members" validate:"required,min=1"`
		Ratio           float64    `json:"ratio" validate:"required,min=0,max=100"`
	}
)
