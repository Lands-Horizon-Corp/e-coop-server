package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

		UserID      *uuid.UUID `gorm:"type:uuid"`
		User        *User      `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;" json:"user,omitempty"`
		MediaID     *uuid.UUID `gorm:"type:uuid;not null"`
		Media       *Media     `gorm:"foreignKey:MediaID"`
		Name        string     `gorm:"type:varchar(255);not null"`
		Description string     `gorm:"type:text;not null"`
		Status      string     `gorm:"type:varchar(50);not null"`
		Progress    float64    `gorm:"not null"`
	}
	// GeneratedReportResponse represents the response structure for generatedreport data
	GeneratedReportResponse struct {
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

		UserID      *uuid.UUID     `json:"user_id"`
		User        *UserResponse  `json:"user"`
		MediaID     *uuid.UUID     `json:"media_id"`
		Media       *MediaResponse `json:"media,omitempty"`
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Status      string         `json:"status"`
		Progress    float64        `json:"progress"`
	}

	// GeneratedReportRequest represents the request structure for creating/updating generatedreport

	// GeneratedReportRequest represents the request structure for GeneratedReport.
	GeneratedReportRequest struct {
		ID          *uuid.UUID `json:"id,omitempty"`
		Name        string     `json:"firstName" validate:"required,min=1,max=255"`
		Description string     `json:"description" validate:"required,min=1"`
	}
)

func (m *Core) generatedReport() {
	m.Migration = append(m.Migration, &GeneratedReport{})
	m.GeneratedReportManager = *registry.NewRegistry(registry.RegistryParams[GeneratedReport, GeneratedReportResponse, GeneratedReportRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Organization", "Branch", "User", "Media"},
		Service:  m.provider.Service,
		Resource: func(data *GeneratedReport) *GeneratedReportResponse {
			if data == nil {
				return nil
			}
			return &GeneratedReportResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),

				UserID:      data.UserID,
				User:        m.UserManager.ToModel(data.User),
				MediaID:     data.MediaID,
				Media:       m.MediaManager.ToModel(data.Media),
				Name:        data.Name,
				Description: data.Description,
				Status:      data.Status,
				Progress:    data.Progress,
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

// GetGenerationReportByUser retrieves generated reports for a specific user
func (m *Core) GetGenerationReportByUser(context context.Context, userID uuid.UUID) ([]*GeneratedReport, error) {
	return m.GeneratedReportManager.Find(context, &GeneratedReport{
		UserID: &userID,
	})
}

// GeneratedReportCurrentBranch gets generated reports for the current branch
func (m *Core) GeneratedReportCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*GeneratedReport, error) {
	return m.GeneratedReportManager.Find(context, &GeneratedReport{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
