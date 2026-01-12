package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
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

func GeneratedReportsDownloadUsersManager(service *horizon.HorizonService) *registry.Registry[GeneratedReportsDownloadUsers, GeneratedReportsDownloadUsersResponse, GeneratedReportsDownloadUsersRequest] {
	return registry.NewRegistry(registry.RegistryParams[GeneratedReportsDownloadUsers, GeneratedReportsDownloadUsersResponse, GeneratedReportsDownloadUsersRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "User"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *GeneratedReportsDownloadUsers) *GeneratedReportsDownloadUsersResponse {
			if data == nil {
				return nil
			}
			return &GeneratedReportsDownloadUsersResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       OrganizationManager(service).ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             BranchManager(service).ToModel(data.Branch),
				UserID:             data.UserID,
				User:               UserManager(service).ToModel(data.User),
				UserOrganizationID: data.UserOrganizationID,
				UserOrganization:   UserOrganizationManager(service).ToModel(data.UserOrganization),
				GeneratedReportID:  data.GeneratedReportID,
				GeneratedReport:    GeneratedReportManager(service).ToModel(data.GeneratedReport),
			}
		},
		Created: func(data *GeneratedReportsDownloadUsers) registry.Topics {
			return []string{
				"generated_reports_download_users.create",
				fmt.Sprintf("generated_reports_download_users.create.%s", data.ID),
				fmt.Sprintf("generated_reports_download_users.create.branch.%s", data.BranchID),
				fmt.Sprintf("generated_reports_download_users.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("generated_reports_download_users.create.user_organization.%s", data.UserOrganizationID),
				fmt.Sprintf("generated_reports_download_users.create.generated_report.%s", data.GeneratedReportID),
			}
		},
		Updated: func(data *GeneratedReportsDownloadUsers) registry.Topics {
			return []string{
				"generated_reports_download_users.update",
				fmt.Sprintf("generated_reports_download_users.update.%s", data.ID),
				fmt.Sprintf("generated_reports_download_users.update.branch.%s", data.BranchID),
				fmt.Sprintf("generated_reports_download_users.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("generated_reports_download_users.update.user_organization.%s", data.UserOrganizationID),
				fmt.Sprintf("generated_reports_download_users.update.generated_report.%s", data.GeneratedReportID),
			}
		},
		Deleted: func(data *GeneratedReportsDownloadUsers) registry.Topics {
			return []string{
				"generated_reports_download_users.delete",
				fmt.Sprintf("generated_reports_download_users.delete.%s", data.ID),
				fmt.Sprintf("generated_reports_download_users.delete.branch.%s", data.BranchID),
				fmt.Sprintf("generated_reports_download_users.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("generated_reports_download_users.delete.user_organization.%s", data.UserOrganizationID),
				fmt.Sprintf("generated_reports_download_users.delete.generated_report.%s", data.GeneratedReportID),
			}
		},
	})
}
