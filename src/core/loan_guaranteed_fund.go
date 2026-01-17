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

func LoanGuaranteedFundManager(service *horizon.HorizonService) *registry.Registry[
	types.LoanGuaranteedFund, types.LoanGuaranteedFundResponse, types.LoanGuaranteedFundRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.LoanGuaranteedFund, types.LoanGuaranteedFundResponse, types.LoanGuaranteedFundRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.LoanGuaranteedFund) *types.LoanGuaranteedFundResponse {
			if data == nil {
				return nil
			}
			return &types.LoanGuaranteedFundResponse{
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
				SchemeNumber:   data.SchemeNumber,
				IncreasingRate: data.IncreasingRate,
			}
		},

		Created: func(data *types.LoanGuaranteedFund) registry.Topics {
			return []string{
				"loan_guaranteed_fund.create",
				fmt.Sprintf("loan_guaranteed_fund.create.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.LoanGuaranteedFund) registry.Topics {
			return []string{
				"loan_guaranteed_fund.update",
				fmt.Sprintf("loan_guaranteed_fund.update.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.LoanGuaranteedFund) registry.Topics {
			return []string{
				"loan_guaranteed_fund.delete",
				fmt.Sprintf("loan_guaranteed_fund.delete.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func LoanGuaranteedFundCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.LoanGuaranteedFund, error) {
	return LoanGuaranteedFundManager(service).Find(context, &types.LoanGuaranteedFund{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
