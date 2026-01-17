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

func LoanTermsAndConditionSuggestedPaymentManager(service *horizon.HorizonService) *registry.Registry[
	types.LoanTermsAndConditionSuggestedPayment, types.LoanTermsAndConditionSuggestedPaymentResponse, types.LoanTermsAndConditionSuggestedPaymentRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.LoanTermsAndConditionSuggestedPayment, types.LoanTermsAndConditionSuggestedPaymentResponse, types.LoanTermsAndConditionSuggestedPaymentRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "LoanTransaction",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.LoanTermsAndConditionSuggestedPayment) *types.LoanTermsAndConditionSuggestedPaymentResponse {
			if data == nil {
				return nil
			}
			return &types.LoanTermsAndConditionSuggestedPaymentResponse{
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
				Name:              data.Name,
				Description:       data.Description,
			}
		},

		Created: func(data *types.LoanTermsAndConditionSuggestedPayment) registry.Topics {
			return []string{
				"loan_terms_and_condition_suggested_payment.create",
				fmt.Sprintf("loan_terms_and_condition_suggested_payment.create.%s", data.ID),
				fmt.Sprintf("loan_terms_and_condition_suggested_payment.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_terms_and_condition_suggested_payment.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.LoanTermsAndConditionSuggestedPayment) registry.Topics {
			return []string{
				"loan_terms_and_condition_suggested_payment.update",
				fmt.Sprintf("loan_terms_and_condition_suggested_payment.update.%s", data.ID),
				fmt.Sprintf("loan_terms_and_condition_suggested_payment.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_terms_and_condition_suggested_payment.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.LoanTermsAndConditionSuggestedPayment) registry.Topics {
			return []string{
				"loan_terms_and_condition_suggested_payment.delete",
				fmt.Sprintf("loan_terms_and_condition_suggested_payment.delete.%s", data.ID),
				fmt.Sprintf("loan_terms_and_condition_suggested_payment.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_terms_and_condition_suggested_payment.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func LoanTermsAndConditionSuggestedPaymentCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.LoanTermsAndConditionSuggestedPayment, error) {
	return LoanTermsAndConditionSuggestedPaymentManager(service).Find(context, &types.LoanTermsAndConditionSuggestedPayment{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
