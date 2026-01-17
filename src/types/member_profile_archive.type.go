package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberProfileArchive struct {
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

		OrganizationID *uuid.UUID    `gorm:"type:uuid;index:idx_organization_branch_member_profile_media" json:"organization_id,omitempty"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       *uuid.UUID    `gorm:"type:uuid;index:idx_organization_branch_member_profile_media" json:"branch_id,omitempty"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid" json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		MediaID *uuid.UUID `gorm:"type:uuid" json:"media_id"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		Name        string `gorm:"type:varchar(255);not null" json:"name"`
		Description string `gorm:"type:text" json:"description"`
		Category    string `gorm:"type:varchar(100);index" json:"category"`
	}

	MemberProfileArchiveResponse struct {
		ID              uuid.UUID              `json:"id"`
		CreatedAt       string                 `json:"created_at"`
		CreatedByID     uuid.UUID              `json:"created_by_id"`
		CreatedBy       *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt       string                 `json:"updated_at"`
		UpdatedByID     uuid.UUID              `json:"updated_by_id"`
		UpdatedBy       *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID  *uuid.UUID             `json:"organization_id,omitempty"`
		Organization    *OrganizationResponse  `json:"organization,omitempty"`
		BranchID        *uuid.UUID             `json:"branch_id,omitempty"`
		Branch          *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID *uuid.UUID             `json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
		MediaID         *uuid.UUID             `json:"media_id,omitempty"`
		Media           *MediaResponse         `json:"media,omitempty"`
		Name            string                 `json:"name"`
		Description     string                 `json:"description"`
		Category        string                 `json:"category"`
	}

	MemberProfileArchiveRequest struct {
		Name            string     `json:"name" validate:"required,min=1,max=255"`
		Description     string     `json:"description,omitempty"`
		MemberProfileID *uuid.UUID `json:"member_profile_id,omitempty"`
		MediaID         *uuid.UUID `json:"media_id,omitempty"`
		OrganizationID  *uuid.UUID `json:"organization_id,omitempty"`
		BranchID        *uuid.UUID `json:"branch_id,omitempty"`
		Category        string     `json:"category,omitempty"`
	}

	MemberProfileArchiveCategoryResponse struct {
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}
	MemberProfileArchiveBulkRequest struct {
		IDs      uuid.UUIDs `json:"ids"`
		Category string     `json:"category" validate:"required,min=1,max=100"`
	}
)
