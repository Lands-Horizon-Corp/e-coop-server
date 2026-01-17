package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	GeneratedReportStatusPending    GeneratedReportStatus = "pending"
	GeneratedReportStatusInProgress GeneratedReportStatus = "in_progress"
	GeneratedReportStatusCompleted  GeneratedReportStatus = "completed"
	GeneratedReportStatusFailed     GeneratedReportStatus = "failed"

	GeneratedReportTypePDF   GeneratedReportType = "pdf"
	GeneratedReportTypeExcel GeneratedReportType = "excel"
)

type (
	GeneratedReportType string

	GeneratedReportStatus string
	GeneratedReport       struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_branch_org_generated_report" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_branch_org_generated_report" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserID        *uuid.UUID            `gorm:"type:uuid" json:"user_id,omitempty"`
		User          *User                 `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;" json:"user,omitempty"`
		MediaID       *uuid.UUID            `gorm:"type:uuid" json:"media_id,omitempty"`
		Media         *Media                `gorm:"foreignKey:MediaID" json:"media,omitempty"`
		Name          string                `gorm:"type:varchar(255);not null" json:"name"`
		Description   string                `gorm:"type:text;not null" json:"description"`
		Status        GeneratedReportStatus `gorm:"type:varchar(50);not null" json:"status"`
		SystemMessage string                `gorm:"type:text" json:"system_message,omitempty"`

		FilterSearch string `gorm:"type:text" json:"filter_search,omitempty"`
		IsFavorite   bool   `gorm:"type:boolean;default:false" json:"is_favorite"`
		Model        string `gorm:"type:varchar(255)" json:"model,omitempty"`
		URL          string `gorm:"type:text" json:"url,omitempty"`

		PaperSize string  `gorm:"type:text;default:''" json:"paper_size,omitempty"`
		Template  string  `gorm:"type:text;default:''" json:"template,omitempty"`
		Width     float64 `gorm:"type:real" json:"width,omitempty"`
		Height    float64 `gorm:"type:real" json:"height,omitempty"`
		Unit      string  `gorm:"type:varchar(50)" json:"unit,omitempty"`
		Landscape bool    `gorm:"type:boolean;default:false" json:"landscape,omitempty"`

		GeneratedReportType GeneratedReportType `gorm:"type:varchar(50);not null;default:'report'" json:"generated_report_type"`

		DownloadUsers []*GeneratedReportsDownloadUsers `gorm:"foreignKey:GeneratedReportID" json:"download_users,omitempty"`
	}

	GeneratedReportResponse struct {
		ID                  uuid.UUID             `json:"id"`
		CreatedAt           string                `json:"created_at"`
		CreatedByID         uuid.UUID             `json:"created_by_id"`
		CreatedBy           *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt           string                `json:"updated_at"`
		UpdatedByID         uuid.UUID             `json:"updated_by_id"`
		UpdatedBy           *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID      uuid.UUID             `json:"organization_id"`
		Organization        *OrganizationResponse `json:"organization,omitempty"`
		BranchID            uuid.UUID             `json:"branch_id"`
		Branch              *BranchResponse       `json:"branch,omitempty"`
		GeneratedReportType GeneratedReportType   `json:"generated_report_type"`

		UserID        *uuid.UUID            `json:"user_id"`
		User          *UserResponse         `json:"user"`
		MediaID       *uuid.UUID            `json:"media_id"`
		Media         *MediaResponse        `json:"media,omitempty"`
		Name          string                `json:"name"`
		Description   string                `json:"description"`
		Status        GeneratedReportStatus `json:"status"`
		IsFavorite    bool                  `json:"is_favorite"`
		Model         string                `json:"model,omitempty"`
		URL           string                `json:"url,omitempty"`
		PaperSize     string                `json:"paper_size,omitempty"`
		Template      string                `json:"template,omitempty"`
		Width         float64               `json:"width,omitempty"`
		Height        float64               `json:"height,omitempty"`
		Unit          string                `json:"unit,omitempty"`
		SystemMessage string                `json:"system_message,omitempty"`

		DownloadUsers []*GeneratedReportsDownloadUsersResponse `json:"download_users,omitempty"`
		Landscape     bool                                     `json:"landscape,omitempty"`
	}

	GeneratedReportRequest struct {
		Name                string              `json:"name" validate:"required,min=1,max=255"`
		Description         string              `json:"description" validate:"required,min=1"`
		FilterSearch        string              `json:"filter_search,omitempty"`
		URL                 string              `json:"url,omitempty"`
		Model               string              `json:"model,omitempty"`
		GeneratedReportType GeneratedReportType `json:"generated_report_type" validate:"required,oneof=pdf excel"`

		PaperSize string  `json:"paper_size,omitempty"`
		Template  string  `json:"template,omitempty"`
		Width     float64 `json:"width,omitempty"`
		Height    float64 `json:"height,omitempty"`
		Unit      string  `json:"unit,omitempty"`
		Landscape bool    `json:"landscape,omitempty"`
	}

	GeneratedReportUpdateRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description" validate:"required,min=1"`
	}

	GeneratedReportAvailableModelsResponse struct {
		Model string `json:"model"`
		Count int    `json:"count"`
	}
)
