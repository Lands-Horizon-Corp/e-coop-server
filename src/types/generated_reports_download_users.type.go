package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	GeneratedReportsDownloadUsers struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_generated_reports_download_users" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_generated_reports_download_users" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_user_generated_reports_download" json:"user_id"`
		User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"user,omitempty"`

		UserOrganizationID uuid.UUID         `gorm:"type:uuid;not null;index:idx_user_organization_generated_reports_download" json:"user_organization_id"`
		UserOrganization   *UserOrganization `gorm:"foreignKey:UserOrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"user_organization,omitempty"`

		GeneratedReportID uuid.UUID        `gorm:"type:uuid;not null;index:idx_generated_report_download_users" json:"generated_report_id"`
		GeneratedReport   *GeneratedReport `gorm:"foreignKey:GeneratedReportID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"generated_report,omitempty"`
	}

	GeneratedReportsDownloadUsersResponse struct {
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

		UserID uuid.UUID     `json:"user_id"`
		User   *UserResponse `json:"user,omitempty"`

		UserOrganizationID uuid.UUID                 `json:"user_organization_id"`
		UserOrganization   *UserOrganizationResponse `json:"user_organization,omitempty"`

		GeneratedReportID uuid.UUID                `json:"generated_report_id"`
		GeneratedReport   *GeneratedReportResponse `json:"generated_report,omitempty"`
	}

	GeneratedReportsDownloadUsersRequest struct {
		UserID             uuid.UUID `json:"user_id" validate:"required"`
		UserOrganizationID uuid.UUID `json:"user_organization_id" validate:"required"`
		GeneratedReportID  uuid.UUID `json:"generated_report_id" validate:"required"`
	}
)
