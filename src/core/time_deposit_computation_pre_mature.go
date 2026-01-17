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

func TimeDepositComputationPreMatureManager(service *horizon.HorizonService) *registry.Registry[
	types.TimeDepositComputationPreMature, types.TimeDepositComputationPreMatureResponse, types.TimeDepositComputationPreMatureRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.TimeDepositComputationPreMature, types.TimeDepositComputationPreMatureResponse, types.TimeDepositComputationPreMatureRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "TimeDepositType",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.TimeDepositComputationPreMature) *types.TimeDepositComputationPreMatureResponse {
			if data == nil {
				return nil
			}
			return &types.TimeDepositComputationPreMatureResponse{
				ID:                data.ID,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				CreatedByID:       data.CreatedByID,
				CreatedBy:         UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:       data.UpdatedByID,
				UpdatedBy:         UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:    data.OrganizationID,
				Organization:      OrganizationManager(service).ToModel(data.Organization),
				BranchID:          data.BranchID,
				Branch:            BranchManager(service).ToModel(data.Branch),
				TimeDepositTypeID: data.TimeDepositTypeID,
				TimeDepositType:   TimeDepositTypeManager(service).ToModel(data.TimeDepositType),
				Terms:             data.Terms,
				From:              data.From,
				To:                data.To,
				Rate:              data.Rate,
			}
		},

		Created: func(data *types.TimeDepositComputationPreMature) registry.Topics {
			return []string{
				"time_deposit_computation_pre_mature.create",
				fmt.Sprintf("time_deposit_computation_pre_mature.create.%s", data.ID),
				fmt.Sprintf("time_deposit_computation_pre_mature.create.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_computation_pre_mature.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.TimeDepositComputationPreMature) registry.Topics {
			return []string{
				"time_deposit_computation_pre_mature.update",
				fmt.Sprintf("time_deposit_computation_pre_mature.update.%s", data.ID),
				fmt.Sprintf("time_deposit_computation_pre_mature.update.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_computation_pre_mature.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.TimeDepositComputationPreMature) registry.Topics {
			return []string{
				"time_deposit_computation_pre_mature.delete",
				fmt.Sprintf("time_deposit_computation_pre_mature.delete.%s", data.ID),
				fmt.Sprintf("time_deposit_computation_pre_mature.delete.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_computation_pre_mature.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func TimeDepositComputationPreMatureCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.TimeDepositComputationPreMature, error) {
	return TimeDepositComputationPreMatureManager(service).Find(context, &types.TimeDepositComputationPreMature{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
