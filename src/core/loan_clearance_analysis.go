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

func LoanClearanceAnalysisManager(service *horizon.HorizonService) *registry.Registry[
	types.LoanClearanceAnalysis, types.LoanClearanceAnalysisResponse, types.LoanClearanceAnalysisRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.LoanClearanceAnalysis, types.LoanClearanceAnalysisResponse, types.LoanClearanceAnalysisRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "LoanTransaction",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.LoanClearanceAnalysis) *types.LoanClearanceAnalysisResponse {
			if data == nil {
				return nil
			}
			return &types.LoanClearanceAnalysisResponse{
				ID:                          data.ID,
				CreatedAt:                   data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                 data.CreatedByID,
				CreatedBy:                   UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:                   data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                 data.UpdatedByID,
				UpdatedBy:                   UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:              data.OrganizationID,
				Organization:                OrganizationManager(service).ToModel(data.Organization),
				BranchID:                    data.BranchID,
				Branch:                      BranchManager(service).ToModel(data.Branch),
				LoanTransactionID:           data.LoanTransactionID,
				LoanTransaction:             LoanTransactionManager(service).ToModel(data.LoanTransaction),
				RegularDeductionDescription: data.RegularDeductionDescription,
				RegularDeductionAmount:      data.RegularDeductionAmount,
				BalancesDescription:         data.BalancesDescription,
				BalancesAmount:              data.BalancesAmount,
				BalancesCount:               data.BalancesCount,
			}
		},

		Created: func(data *types.LoanClearanceAnalysis) registry.Topics {
			return []string{
				"loan_clearance_analysis.create",
				fmt.Sprintf("loan_clearance_analysis.create.%s", data.ID),
				fmt.Sprintf("loan_clearance_analysis.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_clearance_analysis.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.LoanClearanceAnalysis) registry.Topics {
			return []string{
				"loan_clearance_analysis.update",
				fmt.Sprintf("loan_clearance_analysis.update.%s", data.ID),
				fmt.Sprintf("loan_clearance_analysis.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_clearance_analysis.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.LoanClearanceAnalysis) registry.Topics {
			return []string{
				"loan_clearance_analysis.delete",
				fmt.Sprintf("loan_clearance_analysis.delete.%s", data.ID),
				fmt.Sprintf("loan_clearance_analysis.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_clearance_analysis.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func LoanClearanceAnalysisCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.LoanClearanceAnalysis, error) {
	return LoanClearanceAnalysisManager(service).Find(context, &types.LoanClearanceAnalysis{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
