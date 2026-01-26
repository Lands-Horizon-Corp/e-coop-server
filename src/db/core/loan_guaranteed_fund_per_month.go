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

func LoanGuaranteedFundPerMonthManager(service *horizon.HorizonService) *registry.Registry[types.LoanGuaranteedFundPerMonth,
	types.LoanGuaranteedFundPerMonthResponse, types.LoanGuaranteedFundPerMonthRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.LoanGuaranteedFundPerMonth, types.LoanGuaranteedFundPerMonthResponse, types.LoanGuaranteedFundPerMonthRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.LoanGuaranteedFundPerMonth) *types.LoanGuaranteedFundPerMonthResponse {
			if data == nil {
				return nil
			}
			return &types.LoanGuaranteedFundPerMonthResponse{
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
				Month:              data.Month,
				LoanGuaranteedFund: data.LoanGuaranteedFund,
			}
		},

		Created: func(data *types.LoanGuaranteedFundPerMonth) registry.Topics {
			return []string{
				"loan_guaranteed_fund_per_month.create",
				fmt.Sprintf("loan_guaranteed_fund_per_month.create.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.LoanGuaranteedFundPerMonth) registry.Topics {
			return []string{
				"loan_guaranteed_fund_per_month.update",
				fmt.Sprintf("loan_guaranteed_fund_per_month.update.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.LoanGuaranteedFundPerMonth) registry.Topics {
			return []string{
				"loan_guaranteed_fund_per_month.delete",
				fmt.Sprintf("loan_guaranteed_fund_per_month.delete.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func LoanGuaranteedFundPerMonthCurrentBranch(ctx context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.LoanGuaranteedFundPerMonth, error) {
	return LoanGuaranteedFundPerMonthManager(service).Find(ctx, &types.LoanGuaranteedFundPerMonth{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
