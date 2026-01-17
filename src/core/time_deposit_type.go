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

func TimeDepositTypeManager(service *horizon.HorizonService) *registry.Registry[
	types.TimeDepositType, types.TimeDepositTypeResponse, types.TimeDepositTypeRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.TimeDepositType, types.TimeDepositTypeResponse, types.TimeDepositTypeRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Currency", "TimeDepositComputations", "TimeDepositComputationPreMatures",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.TimeDepositType) *types.TimeDepositTypeResponse {
			if data == nil {
				return nil
			}
			return &types.TimeDepositTypeResponse{
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
				CurrencyID:     data.CurrencyID,
				Currency:       CurrencyManager(service).ToModel(data.Currency),

				Header1:  data.Header1,
				Header2:  data.Header2,
				Header3:  data.Header3,
				Header4:  data.Header4,
				Header5:  data.Header5,
				Header6:  data.Header6,
				Header7:  data.Header7,
				Header8:  data.Header8,
				Header9:  data.Header9,
				Header10: data.Header10,
				Header11: data.Header11,

				Name:          data.Name,
				Description:   data.Description,
				PreMature:     data.PreMature,
				PreMatureRate: data.PreMatureRate,
				Excess:        data.Excess,

				TimeDepositComputations:          TimeDepositComputationManager(service).ToModels(data.TimeDepositComputations),
				TimeDepositComputationPreMatures: TimeDepositComputationPreMatureManager(service).ToModels(data.TimeDepositComputationPreMatures),
			}
		},

		Created: func(data *types.TimeDepositType) registry.Topics {
			return []string{
				"time_deposit_type.create",
				fmt.Sprintf("time_deposit_type.create.%s", data.ID),
				fmt.Sprintf("time_deposit_type.create.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_type.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.TimeDepositType) registry.Topics {
			return []string{
				"time_deposit_type.update",
				fmt.Sprintf("time_deposit_type.update.%s", data.ID),
				fmt.Sprintf("time_deposit_type.update.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_type.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.TimeDepositType) registry.Topics {
			return []string{
				"time_deposit_type.delete",
				fmt.Sprintf("time_deposit_type.delete.%s", data.ID),
				fmt.Sprintf("time_deposit_type.delete.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_type.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func TimeDepositTypeCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.TimeDepositType, error) {
	return TimeDepositTypeManager(service).Find(context, &types.TimeDepositType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
