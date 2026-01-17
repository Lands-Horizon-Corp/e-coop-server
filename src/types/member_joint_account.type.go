package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberJointAccount struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_join_account"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_join_account"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		PictureMediaID uuid.UUID `gorm:"type:uuid;not null"`
		PictureMedia   *Media    `gorm:"foreignKey:PictureMediaID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"picture_media,omitempty"`

		SignatureMediaID uuid.UUID `gorm:"type:uuid;not null"`
		SignatureMedia   *Media    `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"signature_media,omitempty"`

		Description        string    `gorm:"type:text"`
		FirstName          string    `gorm:"type:varchar(255);not null"`
		MiddleName         string    `gorm:"type:varchar(255)"`
		LastName           string    `gorm:"type:varchar(255);not null"`
		FullName           string    `gorm:"type:varchar(255);not null"`
		Suffix             string    `gorm:"type:varchar(255)"`
		Birthday           time.Time `gorm:"not null"`
		FamilyRelationship string    `gorm:"type:varchar(255);not null"` // Enum handled on frontend/validation
	}

	MemberJointAccountResponse struct {
		ID                 uuid.UUID              `json:"id"`
		CreatedAt          string                 `json:"created_at"`
		CreatedByID        uuid.UUID              `json:"created_by_id"`
		CreatedBy          *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt          string                 `json:"updated_at"`
		UpdatedByID        uuid.UUID              `json:"updated_by_id"`
		UpdatedBy          *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID     uuid.UUID              `json:"organization_id"`
		Organization       *OrganizationResponse  `json:"organization,omitempty"`
		BranchID           uuid.UUID              `json:"branch_id"`
		Branch             *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID    uuid.UUID              `json:"member_profile_id"`
		MemberProfile      *MemberProfileResponse `json:"member_profile,omitempty"`
		PictureMediaID     uuid.UUID              `json:"picture_media_id"`
		PictureMedia       *MediaResponse         `json:"picture_media,omitempty"`
		SignatureMediaID   uuid.UUID              `json:"signature_media_id"`
		SignatureMedia     *MediaResponse         `json:"signature_media,omitempty"`
		Description        string                 `json:"description"`
		FirstName          string                 `json:"first_name"`
		MiddleName         string                 `json:"middle_name"`
		LastName           string                 `json:"last_name"`
		FullName           string                 `json:"full_name"`
		Suffix             string                 `json:"suffix"`
		Birthday           string                 `json:"birthday"`
		FamilyRelationship string                 `json:"family_relationship"`
	}

	MemberJointAccountRequest struct {
		PictureMediaID     uuid.UUID `json:"picture_media_id" validate:"required"`
		SignatureMediaID   uuid.UUID `json:"signature_media_id" validate:"required"`
		Description        string    `json:"description,omitempty"`
		FirstName          string    `json:"first_name" validate:"required,min=1,max=255"`
		MiddleName         string    `json:"middle_name,omitempty"`
		LastName           string    `json:"last_name" validate:"required,min=1,max=255"`
		FullName           string    `json:"full_name" validate:"required,min=1,max=255"`
		Suffix             string    `json:"suffix,omitempty"`
		Birthday           time.Time `json:"birthday" validate:"required"`
		FamilyRelationship string    `json:"family_relationship" validate:"required,min=1,max=255"`
	}
)
