package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GeneratedReportType string

const (
	GeneratedReportTypePDF   GeneratedReportType = "pdf"
	GeneratedReportTypeExcel GeneratedReportType = "excel"
)

type GeneratedReportStatus string

const (
	GeneratedReportStatusPending    GeneratedReportStatus = "pending"
	GeneratedReportStatusInProgress GeneratedReportStatus = "in_progress"
	GeneratedReportStatusCompleted  GeneratedReportStatus = "completed"
	GeneratedReportStatusFailed     GeneratedReportStatus = "failed"
)

type (
	// GeneratedReport represents the GeneratedReport model.
	GeneratedReport struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_generated_report"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_generated_report"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserID        *uuid.UUID            `gorm:"type:uuid"`
		User          *User                 `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;" json:"user,omitempty"`
		MediaID       *uuid.UUID            `gorm:"type:uuid;not null"`
		Media         *Media                `gorm:"foreignKey:MediaID"`
		Name          string                `gorm:"type:varchar(255);not null"`
		Description   string                `gorm:"type:text;not null"`
		Status        GeneratedReportStatus `gorm:"type:varchar(50);not null"`
		SystemMessage string                `gorm:"type:text" json:"system_message,omitempty"`

		FilterSearch string `gorm:"type:text" json:"filter_search,omitempty"`
		IsFavorite   bool   `gorm:"type:boolean;default:false" json:"is_favorite"`
		Model        string `gorm:"type:varchar(255)" json:"model,omitempty"`
		URL          string `gorm:"type:text" json:"url,omitempty"`

		GeneratedReportType GeneratedReportType `gorm:"type:varchar(50);not null;default:'report'" json:"generated_report_type"`

		// Progress tracking fields
		TotalItems          int `gorm:"type:integer;default:0" json:"total_items,omitempty"`
		TotalItemsProcessed int `gorm:"type:integer;default:0" json:"total_items_processed,omitempty"`

		// One-to-many relationship with GeneratedReportsDownloadUsers
		DownloadUsers []*GeneratedReportsDownloadUsers `gorm:"foreignKey:GeneratedReportID" json:"download_users,omitempty"`
	}

	// GeneratedReportResponse represents the response structure for generatedreport data
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
		SystemMessage string                `json:"system_message,omitempty"`

		// Progress tracking fields
		TotalItems          int `json:"total_items,omitempty"`
		TotalItemsProcessed int `json:"total_items_processed,omitempty"`

		// One-to-many relationship with GeneratedReportsDownloadUsers
		DownloadUsers []*GeneratedReportsDownloadUsersResponse `json:"download_users,omitempty"`
	}

	// GeneratedReportRequest represents the request structure for GeneratedReport.
	GeneratedReportRequest struct {
		Name         string `json:"name" validate:"required,min=1,max=255"`
		Description  string `json:"description" validate:"required,min=1"`
		FilterSearch string `json:"filter_search,omitempty"`
		URL          string `json:"url,omitempty"`
		Model        string `json:"model,omitempty"`
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

func (m *Core) generatedReport() {
	m.Migration = append(m.Migration, &GeneratedReport{})
	m.GeneratedReportManager = *registry.NewRegistry(registry.RegistryParams[GeneratedReport, GeneratedReportResponse, GeneratedReportRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
			"User",
			"Media",
			"DownloadUsers",
			"DownloadUsers.UserOrganization",
			"DownloadUsers.UserOrganization.User.Media",
		},
		Service: m.provider.Service,
		Resource: func(data *GeneratedReport) *GeneratedReportResponse {
			if data == nil {
				return nil
			}
			return &GeneratedReportResponse{
				ID:                  data.ID,
				GeneratedReportType: data.GeneratedReportType,
				CreatedAt:           data.CreatedAt.Format(time.RFC3339),
				CreatedByID:         data.CreatedByID,
				CreatedBy:           m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:           data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:         data.UpdatedByID,
				UpdatedBy:           m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:      data.OrganizationID,
				Organization:        m.OrganizationManager.ToModel(data.Organization),
				BranchID:            data.BranchID,
				Branch:              m.BranchManager.ToModel(data.Branch),
				SystemMessage:       data.SystemMessage,
				UserID:              data.UserID,
				User:                m.UserManager.ToModel(data.User),
				MediaID:             data.MediaID,
				Media:               m.MediaManager.ToModel(data.Media),
				Name:                data.Name,
				Description:         data.Description,
				Status:              data.Status,
				IsFavorite:          data.IsFavorite,
				Model:               data.Model,
				URL:                 data.URL,
				TotalItems:          data.TotalItems,
				TotalItemsProcessed: data.TotalItemsProcessed,
				DownloadUsers:       m.GeneratedReportsDownloadUsersManager.ToModels(data.DownloadUsers),
			}
		},
		Created: func(data *GeneratedReport) []string {
			return []string{
				"generated_report.create",
				fmt.Sprintf("generated_report.create.%s", data.ID),
				fmt.Sprintf("generated_report.create.branch.%s", data.BranchID),
				fmt.Sprintf("generated_report.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("generated_report.create.user.%s", data.UserID),
			}
		},
		Updated: func(data *GeneratedReport) []string {
			return []string{
				"generated_report.update",
				fmt.Sprintf("generated_report.update.%s", data.ID),
				fmt.Sprintf("generated_report.update.branch.%s", data.BranchID),
				fmt.Sprintf("generated_report.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("generated_report.update.user.%s", data.UserID),
			}
		},
		Deleted: func(data *GeneratedReport) []string {
			return []string{
				"generated_report.delete",
				fmt.Sprintf("generated_report.delete.%s", data.ID),
				fmt.Sprintf("generated_report.delete.branch.%s", data.BranchID),
				fmt.Sprintf("generated_report.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("generated_report.delete.user.%s", data.UserID),
			}
		},
	})
}

func (e *Core) GeneratedReportAvailableModels(context context.Context, organizationID, branchID uuid.UUID) ([]GeneratedReportAvailableModelsResponse, error) {
	var results []GeneratedReportAvailableModelsResponse
	err := e.GeneratedReportManager.Client(context).
		Select("model, COUNT(*) as count").
		Where("organization_id = ? AND branch_id = ?", organizationID, branchID).
		Group("model").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
