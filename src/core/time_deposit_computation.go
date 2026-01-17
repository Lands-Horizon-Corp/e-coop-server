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

func TimeDepositComputationManager(service *horizon.HorizonService) *registry.Registry[types.TimeDepositComputation, types.TimeDepositComputationResponse, types.TimeDepositComputationRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.TimeDepositComputation, types.TimeDepositComputationResponse, types.TimeDepositComputationRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "TimeDepositType",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.TimeDepositComputation) *types.TimeDepositComputationResponse {
			if data == nil {
				return nil
			}
			return &types.TimeDepositComputationResponse{
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
				MinimumAmount:     data.MinimumAmount,
				MaximumAmount:     data.MaximumAmount,
				Header1:           data.Header1,
				Header2:           data.Header2,
				Header3:           data.Header3,
				Header4:           data.Header4,
				Header5:           data.Header5,
				Header6:           data.Header6,
				Header7:           data.Header7,
				Header8:           data.Header8,
				Header9:           data.Header9,
				Header10:          data.Header10,
				Header11:          data.Header11,
			}
		},

		Created: func(data *types.TimeDepositComputation) registry.Topics {
			return []string{
				"time_deposit_computation.create",
				fmt.Sprintf("time_deposit_computation.create.%s", data.ID),
				fmt.Sprintf("time_deposit_computation.create.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_computation.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.TimeDepositComputation) registry.Topics {
			return []string{
				"time_deposit_computation.update",
				fmt.Sprintf("time_deposit_computation.update.%s", data.ID),
				fmt.Sprintf("time_deposit_computation.update.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_computation.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.TimeDepositComputation) registry.Topics {
			return []string{
				"time_deposit_computation.delete",
				fmt.Sprintf("time_deposit_computation.delete.%s", data.ID),
				fmt.Sprintf("time_deposit_computation.delete.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_computation.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func TimeDepositComputationCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.TimeDepositComputation, error) {
	return TimeDepositComputationManager(service).Find(context, &types.TimeDepositComputation{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
