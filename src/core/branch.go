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

func BranchManager(service *horizon.HorizonService) *registry.Registry[types.Branch, types.BranchResponse, types.BranchRequest] {
	return registry.GetRegistry(registry.RegistryParams[types.Branch, types.BranchResponse, types.BranchRequest]{
		Preloads: []string{
			"Media",
			"CreatedBy",
			"UpdatedBy",
			"Currency",
			"BranchSetting",
			"Organization.Media",
			"Organization.CreatedBy",
			"Organization.CoverMedia",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Branch) *types.BranchResponse {
			if data == nil {
				return nil
			}
			return &types.BranchResponse{
				ID:           data.ID,
				CreatedAt:    data.CreatedAt.Format(time.RFC3339),
				CreatedByID:  data.CreatedByID,
				CreatedBy:    UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:    data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:  data.UpdatedByID,
				UpdatedBy:    UserManager(service).ToModel(data.UpdatedBy),
				Organization: OrganizationManager(service).ToModel(data.Organization),

				MediaID:       data.MediaID,
				Media:         MediaManager(service).ToModel(data.Media),
				Type:          data.Type,
				Name:          data.Name,
				Email:         data.Email,
				Description:   data.Description,
				CurrencyID:    data.CurrencyID,
				Currency:      CurrencyManager(service).ToModel(data.Currency), // Use the Currency relationship
				ContactNumber: data.ContactNumber,
				Address:       data.Address,
				Province:      data.Province,
				City:          data.City,
				Region:        data.Region,
				Barangay:      data.Barangay,
				PostalCode:    data.PostalCode,
				Latitude:      data.Latitude,
				Longitude:     data.Longitude,

				IsMainBranch: data.IsMainBranch,

				BranchSetting: BranchSettingManager(service).ToModel(data.BranchSetting),

				Footsteps:           FootstepManager(service).ToModels(data.Footsteps),
				GeneratedReports:    GeneratedReportManager(service).ToModels(data.GeneratedReports),
				InvitationCodes:     InvitationCodeManager(service).ToModels(data.InvitationCodes),
				PermissionTemplates: PermissionTemplateManager(service).ToModels(data.PermissionTemplates),
				UserOrganizations:   UserOrganizationManager(service).ToModels(data.UserOrganizations),

				TaxIdentificationNumber: data.TaxIdentificationNumber,
			}
		},
		Created: func(data *types.Branch) registry.Topics {
			return []string{
				"branch.create",
				fmt.Sprintf("branch.create.%s", data.ID),
				fmt.Sprintf("branch.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.Branch) registry.Topics {
			return []string{
				"branch.update",
				fmt.Sprintf("branch.update.%s", data.ID),
				fmt.Sprintf("branch.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.Branch) registry.Topics {
			return []string{
				"branch.delete",
				fmt.Sprintf("branch.delete.%s", data.ID),
				fmt.Sprintf("branch.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GetBranchesByOrganization(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID) ([]*types.Branch, error) {
	return BranchManager(service).Find(context, &types.Branch{OrganizationID: organizationID})
}

func GetBranchesByOrganizationCount(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID) (int64, error) {
	return BranchManager(service).Count(context, &types.Branch{OrganizationID: organizationID})
}
