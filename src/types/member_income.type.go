package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberIncome struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_income"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_income"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MediaID         *uuid.UUID     `gorm:"type:uuid" json:"media_id,omitempty"`
		Media           *Media         `gorm:"foreignKey:MediaID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"media,omitempty"`
		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		Name        string     `gorm:"type:varchar(255)"`
		Source      string     `gorm:"type:varchar(255)"`
		Amount      float64    `gorm:"type:decimal(20,6)"`
		ReleaseDate *time.Time `gorm:"type:timestamp"`
	}

	MemberIncomeResponse struct {
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
		MediaID         *uuid.UUID             `json:"media_id,omitempty"`
		Media           *MediaResponse         `json:"media,omitempty"`
		MemberProfileID uuid.UUID              `json:"member_profile_id"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
		Name            string                 `json:"name"`
		Source          string                 `json:"source"`
		Amount          float64                `json:"amount"`
		ReleaseDate     *string                `json:"release_date,omitempty"`
	}

	MemberIncomeRequest struct {
		MediaID     *uuid.UUID `json:"media_id"`
		Name        string     `json:"name" validate:"required,min=1,max=255"`
		Source      string     `json:"source" validate:"required,min=1,max=255"`
		Amount      float64    `json:"amount" validate:"required"`
		ReleaseDate *time.Time `json:"release_date,omitempty"`
	}
)
