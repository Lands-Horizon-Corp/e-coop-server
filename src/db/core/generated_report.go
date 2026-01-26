package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

func GeneratedReportManager(service *horizon.HorizonService) *registry.Registry[types.GeneratedReport, types.GeneratedReportResponse, types.GeneratedReportRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.GeneratedReport, types.GeneratedReportResponse, types.GeneratedReportRequest]{
		Preloads: []string{
			"CreatedBy",
			"CreatedBy.Media",
			"UpdatedBy",
			"Organization",
			"Branch",
			"User",
			"Media",
			"DownloadUsers.User.Media",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.GeneratedReport) *types.GeneratedReportResponse {
			if data == nil {
				return nil
			}
			var media *types.MediaResponse
			if data.Media != nil {
				media = MediaManager(service).ToModel(data.Media)
				media.DownloadURL = ""
			}
			return &types.GeneratedReportResponse{
				ID:                  data.ID,
				GeneratedReportType: data.GeneratedReportType,
				CreatedAt:           data.CreatedAt.Format(time.RFC3339),
				CreatedByID:         data.CreatedByID,
				CreatedBy:           UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:           data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:         data.UpdatedByID,
				UpdatedBy:           UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:      data.OrganizationID,
				Organization:        OrganizationManager(service).ToModel(data.Organization),
				BranchID:            data.BranchID,
				Branch:              BranchManager(service).ToModel(data.Branch),
				SystemMessage:       data.SystemMessage,
				UserID:              data.UserID,
				User:                UserManager(service).ToModel(data.User),
				MediaID:             data.MediaID,
				Media:               media,
				Name:                data.Name,
				Description:         data.Description,
				Status:              data.Status,
				IsFavorite:          data.IsFavorite,
				Model:               data.Model,
				URL:                 data.URL,
				PaperSize:           data.PaperSize,
				Template:            data.Template,
				Width:               data.Width,
				Height:              data.Height,
				Unit:                data.Unit,
				Landscape:           data.Landscape,

				DownloadUsers: GeneratedReportsDownloadUsersManager(service).ToModels(data.DownloadUsers),
			}
		},
		Created: func(data *types.GeneratedReport) registry.Topics {
			return []string{
				"generated_report.create",
				fmt.Sprintf("generated_report.create.%s", data.ID),
				fmt.Sprintf("generated_report.create.branch.%s", data.BranchID),
				fmt.Sprintf("generated_report.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("generated_report.create.user.%s", data.UserID),
			}
		},
		Updated: func(data *types.GeneratedReport) registry.Topics {
			return []string{
				"generated_report.update",
				fmt.Sprintf("generated_report.update.%s", data.ID),
				fmt.Sprintf("generated_report.update.branch.%s", data.BranchID),
				fmt.Sprintf("generated_report.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("generated_report.update.user.%s", data.UserID),
			}
		},
		Deleted: func(data *types.GeneratedReport) registry.Topics {
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

func GeneratedReportAvailableModels(context context.Context, service *horizon.HorizonService, organizationID, branchID uuid.UUID) ([]types.GeneratedReportAvailableModelsResponse, error) {
	var results []types.GeneratedReportAvailableModelsResponse
	err := GeneratedReportManager(service).Client(context).
		Select("model, COUNT(*) as count").
		Where("organization_id = ? AND branch_id = ?", organizationID, branchID).
		Group("model").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
