package core

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

func GeneralLedgerDefinitionManager(service *horizon.HorizonService) *registry.Registry[
	types.GeneralLedgerDefinition, types.GeneralLedgerDefinitionResponse, types.GeneralLedgerDefinitionRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.GeneralLedgerDefinition, types.GeneralLedgerDefinitionResponse, types.GeneralLedgerDefinitionRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"Accounts",
			"GeneralLedgerAccountsGrouping",
			"GeneralLedgerDefinitionEntries",
			"GeneralLedgerDefinitionEntries",
			"GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries",                                                                                              // Parent of children
			"GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries",                                                                                              // Children level 2
			"GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries",                                                               // Parent of level 2
			"GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries",                                                               // Children level 3
			"GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries",                                // Parent of level 3
			"GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries",                                // Children level 4
			"GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries", // Parent of level 4
			"GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries", // Children level 5
			"GeneralLedgerDefinitionEntries.Accounts",
			"GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.Accounts",
			"GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.Accounts",
			"GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.Accounts",
			"GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.GeneralLedgerDefinitionEntries.Accounts",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.GeneralLedgerDefinition) *types.GeneralLedgerDefinitionResponse {
			if data == nil {
				return nil
			}
			var deletedAt *string
			if data.DeletedAt.Valid {
				t := data.DeletedAt.Time.Format(time.RFC3339)
				deletedAt = &t
			}
			sort.Slice(data.GeneralLedgerDefinitionEntries, func(i, j int) bool {
				return data.GeneralLedgerDefinitionEntries[i].Index < data.GeneralLedgerDefinitionEntries[j].Index
			})
			sort.Slice(data.Accounts, func(i, j int) bool {
				return data.Accounts[i].Index < data.Accounts[j].Index
			})

			entries := GeneralLedgerDefinitionManager(service).ToModels(data.GeneralLedgerDefinitionEntries)
			if len(entries) == 0 || entries == nil {
				entries = []*types.GeneralLedgerDefinitionResponse{}
			}
			return &types.GeneralLedgerDefinitionResponse{
				ID:             data.ID,
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				DeletedByID:    data.DeletedByID,
				DeletedBy:      UserManager(service).ToModel(data.DeletedBy),

				GeneralLedgerDefinitionEntryID:  data.GeneralLedgerDefinitionEntryID,
				GeneralLedgerDefinitionEntries:  entries,
				GeneralLedgerAccountsGroupingID: data.GeneralLedgerAccountsGroupingID,
				GeneralLedgerAccountsGrouping:   GeneralLedgerAccountsGroupingManager(service).ToModel(data.GeneralLedgerAccountsGrouping),

				Accounts:                        AccountManager(service).ToModels(data.Accounts),
				Name:                            data.Name,
				Description:                     data.Description,
				Index:                           data.Index,
				NameInTotal:                     data.NameInTotal,
				IsPosting:                       data.IsPosting,
				GeneralLedgerType:               data.GeneralLedgerType,
				BeginningBalanceOfTheYearCredit: data.BeginningBalanceOfTheYearCredit,
				BeginningBalanceOfTheYearDebit:  data.BeginningBalanceOfTheYearDebit,
				CreatedAt:                       data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:                       data.UpdatedAt.Format(time.RFC3339),
				DeletedAt:                       deletedAt,
				Depth:                           0,
			}
		},
		Created: func(data *types.GeneralLedgerDefinition) registry.Topics {
			return []string{
				"general_ledger_definition.create",
				fmt.Sprintf("general_ledger_definition.create.%s", data.ID),
				fmt.Sprintf("general_ledger_definition.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_definition.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.GeneralLedgerDefinition) registry.Topics {
			return []string{
				"general_ledger_definition.update",
				fmt.Sprintf("general_ledger_definition.update.%s", data.ID),
				fmt.Sprintf("general_ledger_definition.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_definition.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.GeneralLedgerDefinition) registry.Topics {
			return []string{
				"general_ledger_definition.delete",
				fmt.Sprintf("general_ledger_definition.delete.%s", data.ID),
				fmt.Sprintf("general_ledger_definition.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_definition.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GeneralLedgerDefinitionCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.GeneralLedgerDefinition, error) {
	return GeneralLedgerDefinitionManager(service).Find(context, &types.GeneralLedgerDefinition{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
