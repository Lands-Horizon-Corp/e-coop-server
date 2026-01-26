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

func FinancialStatementDefinitionManager(service *horizon.HorizonService) *registry.Registry[
	types.FinancialStatementDefinition, types.FinancialStatementDefinitionResponse, types.FinancialStatementDefinitionRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.FinancialStatementDefinition, types.FinancialStatementDefinitionResponse, types.FinancialStatementDefinitionRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Branch",
			"Organization",
			"Accounts",
			"FinancialStatementDefinitionEntries", // Parent
			"FinancialStatementDefinitionEntries", // Children level 1
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries",                                                                                                             // Parent of children
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries",                                                                                                             // Children level 2
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries",                                                                         // Parent of level 2
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries",                                                                         // Children level 3
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries",                                     // Parent of level 3
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries",                                     // Children level 4
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries", // Parent of level 4
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries", // Children level 5
			"FinancialStatementDefinitionEntries.Accounts",
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.Accounts",
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.Accounts",
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.Accounts",
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.Accounts",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.FinancialStatementDefinition) *types.FinancialStatementDefinitionResponse {
			if data == nil {
				return nil
			}
			return &types.FinancialStatementDefinitionResponse{
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

				FinancialStatementDefinitionEntriesID: data.FinancialStatementDefinitionEntriesID,
				FinancialStatementDefinitionEntries:   FinancialStatementDefinitionManager(service).ToModels(data.FinancialStatementDefinitionEntries),

				FinancialStatementAccountsGroupingID: data.FinancialStatementAccountsGroupingID,
				FinancialStatementAccountsGrouping:   FinancialStatementAccountsGroupingManager(service).ToModel(data.FinancialStatementAccountsGrouping),
				Accounts:                             AccountManager(service).ToModels(data.Accounts),

				Name:                   data.Name,
				Description:            data.Description,
				Index:                  data.Index,
				NameInTotal:            data.NameInTotal,
				IsPosting:              data.IsPosting,
				FinancialStatementType: data.FinancialStatementType,
			}
		},
		Created: func(data *types.FinancialStatementDefinition) registry.Topics {
			return []string{
				"financial_statement_definition.create",
				fmt.Sprintf("financial_statement_definition.create.%s", data.ID),
				fmt.Sprintf("financial_statement_definition.create.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_definition.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.FinancialStatementDefinition) registry.Topics {
			return []string{
				"financial_statement_definition.update",
				fmt.Sprintf("financial_statement_definition.update.%s", data.ID),
				fmt.Sprintf("financial_statement_definition.update.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_definition.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.FinancialStatementDefinition) registry.Topics {
			return []string{
				"financial_statement_definition.delete",
				fmt.Sprintf("financial_statement_definition.delete.%s", data.ID),
				fmt.Sprintf("financial_statement_definition.delete.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_definition.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func FinancialStatementDefinitionCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.FinancialStatementDefinition, error) {
	return FinancialStatementDefinitionManager(service).Find(context, &types.FinancialStatementDefinition{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
