package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberType struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Prefix                     string `gorm:"type:varchar(255)"`
		Name                       string `gorm:"type:varchar(255)"`
		Description                string `gorm:"type:text"`
		BrowseReferenceDescription string `gorm:"type:text"`

		BrowseReferences []*BrowseReference `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"browse_references,omitempty"`
	}

	MemberTypeResponse struct {
		ID                         uuid.UUID                  `json:"id"`
		CreatedAt                  string                     `json:"created_at"`
		CreatedByID                uuid.UUID                  `json:"created_by_id"`
		CreatedBy                  *UserResponse              `json:"created_by,omitempty"`
		UpdatedAt                  string                     `json:"updated_at"`
		UpdatedByID                uuid.UUID                  `json:"updated_by_id"`
		UpdatedBy                  *UserResponse              `json:"updated_by,omitempty"`
		OrganizationID             uuid.UUID                  `json:"organization_id"`
		Organization               *OrganizationResponse      `json:"organization,omitempty"`
		BranchID                   uuid.UUID                  `json:"branch_id"`
		Branch                     *BranchResponse            `json:"branch,omitempty"`
		Prefix                     string                     `json:"prefix"`
		Name                       string                     `json:"name"`
		Description                string                     `json:"description"`
		BrowseReferenceDescription string                     `json:"browse_reference_description"`
		BrowseReferences           []*BrowseReferenceResponse `json:"browse_references,omitempty"`
	}

	MemberTypeRequest struct {
		Prefix                     string `json:"prefix,omitempty"`
		Name                       string `json:"name,omitempty"`
		Description                string `json:"description,omitempty"`
		BrowseReferenceDescription string `json:"browse_reference_description,omitempty"`
	}
)
