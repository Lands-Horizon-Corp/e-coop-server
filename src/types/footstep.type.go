package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	FootstepLevelInfo    FootstepLevel = "info"
	FootstepLevelWarning FootstepLevel = "warning"
	FootstepLevelError   FootstepLevel = "error"
	FootstepLevelDebug   FootstepLevel = "debug"
)

type (
	FootstepLevel string
	Footstep      struct {
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
		OrganizationID *uuid.UUID     `gorm:"type:uuid;index:idx_branch_org_footstep"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       *uuid.UUID     `gorm:"type:uuid;index:idx_branch_org_footstep"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserID  *uuid.UUID `gorm:"type:uuid"`
		User    *User      `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;" json:"user,omitempty"`
		MediaID *uuid.UUID `gorm:"type:uuid"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		Description    string               `gorm:"type:text;not null"`
		Activity       string               `gorm:"type:text;not null"`
		UserType       UserOrganizationType `gorm:"type:varchar(11);unsigned" json:"user_type"`
		Module         string               `gorm:"type:varchar(255);unsigned" json:"module"`
		Latitude       *float64             `gorm:"type:decimal(10,7)" json:"latitude,omitempty"`
		Longitude      *float64             `gorm:"type:decimal(10,7)" json:"longitude,omitempty"`
		Timestamp      time.Time            `gorm:"type:timestamp" json:"timestamp"`
		IsDeleted      bool                 `gorm:"default:false" json:"is_deleted"`
		IPAddress      string               `gorm:"type:varchar(45)" json:"ip_address"`
		UserAgent      string               `gorm:"type:varchar(1000)" json:"user_agent"`
		Referer        string               `gorm:"type:varchar(1000)" json:"referer"`
		Location       string               `gorm:"type:varchar(255)" json:"location"`
		AcceptLanguage string               `gorm:"type:varchar(255)" json:"accept_language"`
		Level          FootstepLevel        `gorm:"type:varchar(255)" json:"level"`
	}

	FootstepResponse struct {
		ID             uuid.UUID             `json:"id"`
		CreatedAt      string                `json:"created_at"`
		CreatedByID    uuid.UUID             `json:"created_by_id"`
		CreatedBy      *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt      string                `json:"updated_at"`
		UpdatedByID    uuid.UUID             `json:"updated_by_id"`
		UpdatedBy      *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID *uuid.UUID            `json:"organization_id,omitempty"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       *uuid.UUID            `json:"branch_id,omitempty"`
		Branch         *BranchResponse       `json:"branch,omitempty"`

		UserID  *uuid.UUID     `json:"user_id,omitempty"`
		User    *UserResponse  `json:"user,omitempty"`
		MediaID *uuid.UUID     `json:"media_id,omitempty"`
		Media   *MediaResponse `json:"media,omitempty"`

		Description    string               `json:"description"`
		Activity       string               `json:"activity"`
		UserType       UserOrganizationType `json:"user_type"`
		Module         string               `json:"module"`
		Latitude       *float64             `json:"latitude,omitempty"`
		Longitude      *float64             `json:"longitude,omitempty"`
		Timestamp      string               `json:"timestamp"`
		IsDeleted      bool                 `json:"is_deleted"`
		IPAddress      string               `json:"ip_address"`
		UserAgent      string               `json:"user_agent"`
		Referer        string               `json:"referer"`
		Location       string               `json:"location"`
		AcceptLanguage string               `json:"accept_language"`
		Level          FootstepLevel        `json:"level"`
	}

	FootstepRequest struct {
		Level       FootstepLevel `json:"level" validate:"required,oneof=info warning error debug"`
		Description string        `json:"description"`
		Activity    string        `json:"activity"`
		Module      string        `json:"module"`
	}
)
