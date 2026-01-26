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

func ComakerCollateralManager(service *horizon.HorizonService) *registry.Registry[
	types.ComakerCollateral, types.ComakerCollateralResponse, types.ComakerCollateralRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.ComakerCollateral, types.ComakerCollateralResponse, types.ComakerCollateralRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "LoanTransaction", "Collateral"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.ComakerCollateral) *types.ComakerCollateralResponse {
			if data == nil {
				return nil
			}
			return &types.ComakerCollateralResponse{
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
				LoanTransactionID: data.LoanTransactionID,
				LoanTransaction:   LoanTransactionManager(service).ToModel(data.LoanTransaction),
				CollateralID:      data.CollateralID,
				Collateral:        CollateralManager(service).ToModel(data.Collateral),
				Amount:            data.Amount,
				Description:       data.Description,
				MonthsCount:       data.MonthsCount,
				YearCount:         data.YearCount,
			}
		},
		Created: func(data *types.ComakerCollateral) registry.Topics {
			return []string{
				"comaker_collateral.create",
				fmt.Sprintf("comaker_collateral.create.%s", data.ID),
				fmt.Sprintf("comaker_collateral.create.branch.%s", data.BranchID),
				fmt.Sprintf("comaker_collateral.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("comaker_collateral.create.loan_transaction.%s", data.LoanTransactionID),
			}
		},
		Updated: func(data *types.ComakerCollateral) registry.Topics {
			return []string{
				"comaker_collateral.update",
				fmt.Sprintf("comaker_collateral.update.%s", data.ID),
				fmt.Sprintf("comaker_collateral.update.branch.%s", data.BranchID),
				fmt.Sprintf("comaker_collateral.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("comaker_collateral.update.loan_transaction.%s", data.LoanTransactionID),
			}
		},
		Deleted: func(data *types.ComakerCollateral) registry.Topics {
			return []string{
				"comaker_collateral.delete",
				fmt.Sprintf("comaker_collateral.delete.%s", data.ID),
				fmt.Sprintf("comaker_collateral.delete.branch.%s", data.BranchID),
				fmt.Sprintf("comaker_collateral.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("comaker_collateral.delete.loan_transaction.%s", data.LoanTransactionID),
			}
		},
	})
}

func ComakerCollateralCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.ComakerCollateral, error) {
	return ComakerCollateralManager(service).Find(context, &types.ComakerCollateral{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func ComakerCollateralByLoanTransaction(context context.Context, service *horizon.HorizonService, loanTransactionID uuid.UUID) ([]*types.ComakerCollateral, error) {
	return ComakerCollateralManager(service).Find(context, &types.ComakerCollateral{
		LoanTransactionID: loanTransactionID,
	})
}
