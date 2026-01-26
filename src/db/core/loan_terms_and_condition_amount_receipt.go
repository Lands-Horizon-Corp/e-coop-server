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

func LoanTermsAndConditionAmountReceiptManager(service *horizon.HorizonService) *registry.Registry[
	types.LoanTermsAndConditionAmountReceipt, types.LoanTermsAndConditionAmountReceiptResponse, types.LoanTermsAndConditionAmountReceiptRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.LoanTermsAndConditionAmountReceipt, types.LoanTermsAndConditionAmountReceiptResponse, types.LoanTermsAndConditionAmountReceiptRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "LoanTransaction", "Account",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.LoanTermsAndConditionAmountReceipt) *types.LoanTermsAndConditionAmountReceiptResponse {
			if data == nil {
				return nil
			}
			return &types.LoanTermsAndConditionAmountReceiptResponse{
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
				AccountID:         data.AccountID,
				Account:           AccountManager(service).ToModel(data.Account),
				Amount:            data.Amount,
			}
		},

		Created: func(data *types.LoanTermsAndConditionAmountReceipt) registry.Topics {
			return []string{
				"loan_terms_and_condition_amount_receipt.create",
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.create.%s", data.ID),
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.LoanTermsAndConditionAmountReceipt) registry.Topics {
			return []string{
				"loan_terms_and_condition_amount_receipt.update",
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.update.%s", data.ID),
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.LoanTermsAndConditionAmountReceipt) registry.Topics {
			return []string{
				"loan_terms_and_condition_amount_receipt.delete",
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.delete.%s", data.ID),
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func LoanTermsAndConditionAmountReceiptCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.LoanTermsAndConditionAmountReceipt, error) {
	return LoanTermsAndConditionAmountReceiptManager(service).Find(context, &types.LoanTermsAndConditionAmountReceipt{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
