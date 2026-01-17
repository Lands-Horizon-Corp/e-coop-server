package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberVerification struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_verification"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       *uuid.UUID    `gorm:"type:uuid;not null;index:idx_organization_branch_member_verification"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		VerifiedByUserID *uuid.UUID `gorm:"type:uuid"`
		VerifiedByUser   *User      `gorm:"foreignKey:VerifiedByUserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"verified_by_user,omitempty"`

		Status string `gorm:"type:varchar(50);not null;default:'pending'"`
	}

	MemberVerificationResponse struct {
		ID               uuid.UUID              `json:"id"`
		CreatedAt        string                 `json:"created_at"`
		CreatedByID      uuid.UUID              `json:"created_by_id"`
		CreatedBy        *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt        string                 `json:"updated_at"`
		UpdatedByID      uuid.UUID              `json:"updated_by_id"`
		UpdatedBy        *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID   uuid.UUID              `json:"organization_id"`
		Organization     *OrganizationResponse  `json:"organization,omitempty"`
		BranchID         uuid.UUID              `json:"branch_id"`
		Branch           *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID  uuid.UUID              `json:"member_profile_id"`
		MemberProfile    *MemberProfileResponse `json:"member_profile,omitempty"`
		VerifiedByUserID uuid.UUID              `json:"verified_by_user_id"`
		VerifiedByUser   *UserResponse          `json:"verified_by_user,omitempty"`
		Status           string                 `json:"status"`
	}

	MemberVerificationRequest struct {
		MemberProfileID  uuid.UUID `json:"member_profile_id" validate:"required"`
		VerifiedByUserID uuid.UUID `json:"verified_by_user_id,omitempty"`
		Status           string    `json:"status,omitempty"`
	}
)
