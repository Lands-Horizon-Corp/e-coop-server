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

func FootstepManager(service *horizon.HorizonService) *registry.Registry[types.Footstep, types.FootstepResponse, types.FootstepRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.Footstep, types.FootstepResponse, types.FootstepRequest]{
		Preloads: []string{
			"User",
			"User.Media",
			"Branch",
			"Branch.Media",
			"Organization",
			"Organization.Media",
			"Organization.CoverMedia",
			"Media",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Footstep) *types.FootstepResponse {
			if data == nil {
				return nil
			}
			return &types.FootstepResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),

				UserID:  data.UserID,
				User:    UserManager(service).ToModel(data.User),
				MediaID: data.MediaID,
				Media:   MediaManager(service).ToModel(data.Media),

				Description:    data.Description,
				Activity:       data.Activity,
				UserType:       data.UserType,
				Module:         data.Module,
				Latitude:       data.Latitude,
				Longitude:      data.Longitude,
				Timestamp:      data.Timestamp.Format(time.RFC3339),
				IsDeleted:      data.IsDeleted,
				IPAddress:      data.IPAddress,
				UserAgent:      data.UserAgent,
				Referer:        data.Referer,
				Location:       data.Location,
				AcceptLanguage: data.AcceptLanguage,
				Level:          data.Level,
			}
		},
		Created: func(data *types.Footstep) registry.Topics {
			return []string{
				"footstep.create",
				fmt.Sprintf("footstep.create.%s", data.ID),
				fmt.Sprintf("footstep.create.branch.%s", data.BranchID),
				fmt.Sprintf("footstep.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.Footstep) registry.Topics {
			return []string{
				"footstep.update",
				fmt.Sprintf("footstep.update.%s", data.ID),
				fmt.Sprintf("footstep.update.branch.%s", data.BranchID),
				fmt.Sprintf("footstep.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.Footstep) registry.Topics {
			return []string{
				"footstep.delete",
				fmt.Sprintf("footstep.delete.%s", data.ID),
				fmt.Sprintf("footstep.delete.branch.%s", data.BranchID),
				fmt.Sprintf("footstep.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GetFootstepByUser(context context.Context, service *horizon.HorizonService, userID uuid.UUID) ([]*types.Footstep, error) {
	return FootstepManager(service).Find(context, &types.Footstep{
		UserID: &userID,
	})
}

func GetFootstepByBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.Footstep, error) {
	return FootstepManager(service).Find(context, &types.Footstep{
		OrganizationID: &organizationID,
		BranchID:       &branchID,
	})
}

func GetFootstepByUserOrganization(context context.Context, service *horizon.HorizonService, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.Footstep, error) {
	return FootstepManager(service).Find(context, &types.Footstep{
		UserID:         &userID,
		OrganizationID: &organizationID,
		BranchID:       &branchID,
	})
}
