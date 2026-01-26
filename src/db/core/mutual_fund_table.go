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

func MutualFundTableManager(service *horizon.HorizonService) *registry.Registry[
	types.MutualFundTable, types.MutualFundTableResponse, types.MutualFundTableRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MutualFundTable, types.MutualFundTableResponse, types.MutualFundTableRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Organization", "Branch", "MutualFund"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MutualFundTable) *types.MutualFundTableResponse {
			if data == nil {
				return nil
			}
			return &types.MutualFundTableResponse{
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
				MutualFundID:   data.MutualFundID,
				MutualFund:     MutualFundManager(service).ToModel(data.MutualFund),
				MonthFrom:      data.MonthFrom,
				MonthTo:        data.MonthTo,
				Amount:         data.Amount,
			}
		},
		Created: func(data *types.MutualFundTable) registry.Topics {
			return []string{
				"mutual_fund_table.create",
				fmt.Sprintf("mutual_fund_table.create.%s", data.ID),
				fmt.Sprintf("mutual_fund_table.create.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_table.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_table.create.mutual_fund.%s", data.MutualFundID),
			}
		},
		Updated: func(data *types.MutualFundTable) registry.Topics {
			return []string{
				"mutual_fund_table.update",
				fmt.Sprintf("mutual_fund_table.update.%s", data.ID),
				fmt.Sprintf("mutual_fund_table.update.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_table.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_table.update.mutual_fund.%s", data.MutualFundID),
			}
		},
		Deleted: func(data *types.MutualFundTable) registry.Topics {
			return []string{
				"mutual_fund_table.delete",
				fmt.Sprintf("mutual_fund_table.delete.%s", data.ID),
				fmt.Sprintf("mutual_fund_table.delete.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_table.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_table.delete.mutual_fund.%s", data.MutualFundID),
			}
		},
	})
}

func MutualFundTableCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MutualFundTable, error) {
	return MutualFundTableManager(service).Find(context, &types.MutualFundTable{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func MutualFundTableByMutualFund(context context.Context, service *horizon.HorizonService,
	mutualFundID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MutualFundTable, error) {
	return MutualFundTableManager(service).Find(context, &types.MutualFundTable{
		MutualFundID:   mutualFundID,
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
