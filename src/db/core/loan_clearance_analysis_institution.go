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

func LoanClearanceAnalysisInstitutionManager(service *horizon.HorizonService) *registry.Registry[
	types.LoanClearanceAnalysisInstitution, types.LoanClearanceAnalysisInstitutionResponse, types.LoanClearanceAnalysisInstitutionRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.LoanClearanceAnalysisInstitution, types.LoanClearanceAnalysisInstitutionResponse, types.LoanClearanceAnalysisInstitutionRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "LoanTransaction",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.LoanClearanceAnalysisInstitution) *types.LoanClearanceAnalysisInstitutionResponse {
			if data == nil {
				return nil
			}
			return &types.LoanClearanceAnalysisInstitutionResponse{
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

		Created: func(data *types.LoanClearanceAnalysisInstitution) registry.Topics {
			return []string{
				"loan_clearance_analysis_institution.create",
				fmt.Sprintf("loan_clearance_analysis_institution.create.%s", data.ID),
				fmt.Sprintf("loan_clearance_analysis_institution.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_clearance_analysis_institution.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.LoanClearanceAnalysisInstitution) registry.Topics {
			return []string{
				"loan_clearance_analysis_institution.update",
				fmt.Sprintf("loan_clearance_analysis_institution.update.%s", data.ID),
				fmt.Sprintf("loan_clearance_analysis_institution.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_clearance_analysis_institution.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.LoanClearanceAnalysisInstitution) registry.Topics {
			return []string{
				"loan_clearance_analysis_institution.delete",
				fmt.Sprintf("loan_clearance_analysis_institution.delete.%s", data.ID),
				fmt.Sprintf("loan_clearance_analysis_institution.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_clearance_analysis_institution.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func LoanClearanceAnalysisInstitutionCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.LoanClearanceAnalysisInstitution, error) {
	return LoanClearanceAnalysisInstitutionManager(service).Find(context, &types.LoanClearanceAnalysisInstitution{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
