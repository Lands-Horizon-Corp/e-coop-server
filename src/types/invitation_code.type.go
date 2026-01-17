package types

import (
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type (
	InvitationCode struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt      time.Time      `gorm:"not null;default:now()"`
		CreatedByID    uuid.UUID      `gorm:"type:uuid"`
		CreatedBy      *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()"`
		UpdatedByID    uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy      *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt      gorm.DeletedAt `gorm:"index"`
		DeletedByID    *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy      *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_invitation_code"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_invitation_code"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserType       UserOrganizationType `gorm:"type:varchar(255);not null"`
		Code           string               `gorm:"type:varchar(255);not null;unique"`
		ExpirationDate time.Time            `gorm:"not null"`
		MaxUse         int                  `gorm:"not null"`
		CurrentUse     int                  `gorm:"default:0"`
		Description    string               `gorm:"type:text"`

		PermissionName        string         `gorm:"type:varchar(255);not null" json:"permission_name"`
		PermissionDescription string         `gorm:"type:varchar(255);not null" json:"permission_description"`
		Permissions           pq.StringArray `gorm:"type:varchar(255)[]" json:"permissions"`
	}

	InvitationCodeResponse struct {
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

		UserType       UserOrganizationType `json:"user_type"`
		Code           string               `json:"code"`
		ExpirationDate string               `json:"expiration_date"`
		MaxUse         int                  `json:"max_use"`
		CurrentUse     int                  `json:"current_use"`
		Description    string               `json:"description,omitempty"`
		QRCode         *horizon.QRResult    `json:"qr_code,omitempty"`

		PermissionName        string   `json:"permission_name"`
		PermissionDescription string   `json:"permission_description"`
		Permissions           []string `json:"permissions"`
	}

	InvitationCodeRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		UserType       string    `json:"user_type" validate:"required,oneof=employee owner member"`
		Code           string    `json:"code" validate:"required,max=255"`
		ExpirationDate time.Time `json:"expiration_date" validate:"required"`
		MaxUse         int       `json:"max_use" validate:"required"`
		Description    string    `json:"description,omitempty"`

		PermissionName        string   `json:"permission_name,omitempty"`
		PermissionDescription string   `json:"permission_description,omitempty"`
		Permissions           []string `json:"permissions,omitempty" validate:"dive"`
	}
)
