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

func AutomaticLoanDeductionManager(service *horizon.HorizonService) *registry.Registry[
	types.AutomaticLoanDeduction, types.AutomaticLoanDeductionResponse, types.AutomaticLoanDeductionRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.AutomaticLoanDeduction, types.AutomaticLoanDeductionResponse, types.AutomaticLoanDeductionRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Account.Currency",
			"Account", "ComputationSheet", "ChargesRateScheme",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.AutomaticLoanDeduction) *types.AutomaticLoanDeductionResponse {
			if data == nil {
				return nil
			}
			return &types.AutomaticLoanDeductionResponse{
				ID:                  data.ID,
				CreatedAt:           data.CreatedAt.Format(time.RFC3339),
				CreatedByID:         data.CreatedByID,
				CreatedBy:           UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:           data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:         data.UpdatedByID,
				UpdatedBy:           UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:      data.OrganizationID,
				Organization:        OrganizationManager(service).ToModel(data.Organization),
				BranchID:            data.BranchID,
				Branch:              BranchManager(service).ToModel(data.Branch),
				AccountID:           data.AccountID,
				Account:             AccountManager(service).ToModel(data.Account),
				ComputationSheetID:  data.ComputationSheetID,
				ComputationSheet:    ComputationSheetManager(service).ToModel(data.ComputationSheet),
				ChargesRateSchemeID: data.ChargesRateSchemeID,
				ChargesRateScheme:   ChargesRateSchemeManager(service).ToModel(data.ChargesRateScheme),
				ChargesPercentage1:  data.ChargesPercentage1,
				ChargesPercentage2:  data.ChargesPercentage2,
				ChargesAmount:       data.ChargesAmount,
				ChargesDivisor:      data.ChargesDivisor,
				MinAmount:           data.MinAmount,
				MaxAmount:           data.MaxAmount,
				Anum:                data.Anum,
				NumberOfMonths:      data.NumberOfMonths,
				AddOn:               data.AddOn,
				AoRest:              data.AoRest,
				ExcludeRenewal:      data.ExcludeRenewal,
				Ct:                  data.Ct,
				Name:                data.Name,
				Description:         data.Description,
			}
		},
		Created: func(data *types.AutomaticLoanDeduction) registry.Topics {
			return []string{
				"automatic_loan_deduction.create",
				fmt.Sprintf("automatic_loan_deduction.create.%s", data.ID),
				fmt.Sprintf("automatic_loan_deduction.create.branch.%s", data.BranchID),
				fmt.Sprintf("automatic_loan_deduction.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.AutomaticLoanDeduction) registry.Topics {
			return []string{
				"automatic_loan_deduction.update",
				fmt.Sprintf("automatic_loan_deduction.update.%s", data.ID),
				fmt.Sprintf("automatic_loan_deduction.update.branch.%s", data.BranchID),
				fmt.Sprintf("automatic_loan_deduction.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.AutomaticLoanDeduction) registry.Topics {
			return []string{
				"automatic_loan_deduction.update",
				fmt.Sprintf("automatic_loan_deduction.delete.%s", data.ID),
				fmt.Sprintf("automatic_loan_deduction.delete.branch.%s", data.BranchID),
				fmt.Sprintf("automatic_loan_deduction.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func AutomaticLoanDeductionCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.AutomaticLoanDeduction, error) {
	return AutomaticLoanDeductionManager(service).Find(context, &types.AutomaticLoanDeduction{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
