package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberAddress struct {
		ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`

		CreatedByID *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
		CreatedBy   *User      `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by_user,omitempty"`

		UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`

		UpdatedByID *uuid.UUID `gorm:"type:uuid" json:"updated_by,omitempty"`
		UpdatedBy   *User      `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by_user,omitempty"`

		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by,omitempty"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by_user,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_address" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`

		BranchID uuid.UUID `gorm:"type:uuid;not null;index:idx_organization_branch_member_address" json:"branch_id"`
		Branch   *Branch   `gorm:"foreignKey:BranchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid" json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		Label         string `gorm:"type:varchar(255);not null;default:home"`
		City          string `gorm:"type:varchar(255);not null"`
		CountryCode   string `gorm:"type:varchar(5);not null"`
		PostalCode    string `gorm:"type:varchar(255)"`
		ProvinceState string `gorm:"type:varchar(255)"`
		Barangay      string `gorm:"type:varchar(255)"`
		Landmark      string `gorm:"type:varchar(255)"`
		Address       string `gorm:"type:varchar(255);not null"`

		Latitude  *float64 `gorm:"type:double precision" json:"latitude,omitempty"`
		Longitude *float64 `gorm:"type:double precision" json:"longitude,omitempty"`
	}

	MemberAddressResponse struct {
		ID             uuid.UUID             `json:"id"`
		CreatedAt      string                `json:"created_at"`
		CreatedByID    uuid.UUID             `json:"created_by_id"`
		CreatedBy      *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt      string                `json:"updated_at"`
		UpdatedByID    uuid.UUID             `json:"updated_by_id"`
		UpdatedBy      *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       uuid.UUID             `json:"branch_id"`
		Branch         *BranchResponse       `json:"branch,omitempty"`

		MemberProfileID *uuid.UUID             `json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`

		Label         string   `json:"label"`
		City          string   `json:"city"`
		CountryCode   string   `json:"country_code"`
		PostalCode    string   `json:"postal_code"`
		ProvinceState string   `json:"province_state"`
		Barangay      string   `json:"barangay"`
		Landmark      string   `json:"landmark"`
		Address       string   `json:"address"`
		Longitude     *float64 `json:"longitude,omitempty"`
		Latitude      *float64 `json:"latitude,omitempty"`
	}

	MemberAddressRequest struct {
		MemberProfileID *uuid.UUID `json:"member_profile_id,omitempty"`

		Label         string   `json:"label" validate:"required,min=1,max=255"`
		City          string   `json:"city" validate:"required,min=1,max=255"`
		CountryCode   string   `json:"country_code" validate:"required,min=1,max=5"`
		PostalCode    string   `json:"postal_code,omitempty" validate:"omitempty,max=255"`
		ProvinceState string   `json:"province_state,omitempty" validate:"omitempty,max=255"`
		Barangay      string   `json:"barangay,omitempty" validate:"omitempty,max=255"`
		Landmark      string   `json:"landmark,omitempty" validate:"omitempty,max=255"`
		Address       string   `json:"address" validate:"required,min=1,max=255"`
		Longitude     *float64 `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
		Latitude      *float64 `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	}
)
