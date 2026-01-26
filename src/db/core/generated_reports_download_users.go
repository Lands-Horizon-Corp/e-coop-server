package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

func GeneratedReportsDownloadUsersManager(service *horizon.HorizonService) *registry.Registry[
	types.GeneratedReportsDownloadUsers, types.GeneratedReportsDownloadUsersResponse, types.GeneratedReportsDownloadUsersRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.GeneratedReportsDownloadUsers, types.GeneratedReportsDownloadUsersResponse, types.GeneratedReportsDownloadUsersRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "User"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.GeneratedReportsDownloadUsers) *types.GeneratedReportsDownloadUsersResponse {
			if data == nil {
				return nil
			}
			return &types.GeneratedReportsDownloadUsersResponse{
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
		Created: func(data *types.GeneratedReportsDownloadUsers) registry.Topics {
			return []string{
				"generated_reports_download_users.create",
				fmt.Sprintf("generated_reports_download_users.create.%s", data.ID),
				fmt.Sprintf("generated_reports_download_users.create.branch.%s", data.BranchID),
				fmt.Sprintf("generated_reports_download_users.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("generated_reports_download_users.create.user_organization.%s", data.UserOrganizationID),
				fmt.Sprintf("generated_reports_download_users.create.generated_report.%s", data.GeneratedReportID),
			}
		},
		Updated: func(data *types.GeneratedReportsDownloadUsers) registry.Topics {
			return []string{
				"generated_reports_download_users.update",
				fmt.Sprintf("generated_reports_download_users.update.%s", data.ID),
				fmt.Sprintf("generated_reports_download_users.update.branch.%s", data.BranchID),
				fmt.Sprintf("generated_reports_download_users.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("generated_reports_download_users.update.user_organization.%s", data.UserOrganizationID),
				fmt.Sprintf("generated_reports_download_users.update.generated_report.%s", data.GeneratedReportID),
			}
		},
		Deleted: func(data *types.GeneratedReportsDownloadUsers) registry.Topics {
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
