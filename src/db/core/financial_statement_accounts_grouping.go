package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func FinancialStatementAccountsGroupingManager(service *horizon.HorizonService) *registry.Registry[
	types.FinancialStatementAccountsGrouping, types.FinancialStatementAccountsGroupingResponse, types.FinancialStatementAccountsGroupingRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.FinancialStatementAccountsGrouping, types.FinancialStatementAccountsGroupingResponse, types.FinancialStatementAccountsGroupingRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "IconMedia",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.FinancialStatementAccountsGrouping) *types.FinancialStatementAccountsGroupingResponse {
			if data == nil {
				return nil
			}
			var deletedAt *string
			if data.DeletedAt.Valid {
				t := data.DeletedAt.Time.Format(time.RFC3339)
				deletedAt = &t
			}
			return &types.FinancialStatementAccountsGroupingResponse{
				ID:                                  data.ID,
				OrganizationID:                      data.OrganizationID,
				Organization:                        OrganizationManager(service).ToModel(data.Organization),
				BranchID:                            data.BranchID,
				Branch:                              BranchManager(service).ToModel(data.Branch),
				CreatedByID:                         data.CreatedByID,
				CreatedBy:                           UserManager(service).ToModel(data.CreatedBy),
				UpdatedByID:                         data.UpdatedByID,
				UpdatedBy:                           UserManager(service).ToModel(data.UpdatedBy),
				DeletedByID:                         data.DeletedByID,
				DeletedBy:                           UserManager(service).ToModel(data.DeletedBy),
				IconMediaID:                         data.IconMediaID,
				IconMedia:                           MediaManager(service).ToModel(data.IconMedia),
				Name:                                data.Name,
				Description:                         data.Description,
				Debit:                               data.Debit,
				Credit:                              data.Credit,
				CreatedAt:                           data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:                           data.UpdatedAt.Format(time.RFC3339),
				DeletedAt:                           deletedAt,
				FinancialStatementDefinitionEntries: FinancialStatementDefinitionManager(service).ToModels(data.FinancialStatementDefinitionEntries),
			}
		},
		Created: func(data *types.FinancialStatementAccountsGrouping) registry.Topics {
			return []string{
				"financial_statement_grouping.create",
				fmt.Sprintf("financial_statement_grouping.create.%s", data.ID),
				fmt.Sprintf("financial_statement_grouping.create.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_grouping.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.FinancialStatementAccountsGrouping) registry.Topics {
			return []string{
				"financial_statement_grouping.update",
				fmt.Sprintf("financial_statement_grouping.update.%s", data.ID),
				fmt.Sprintf("financial_statement_grouping.update.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_grouping.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.FinancialStatementAccountsGrouping) registry.Topics {
			return []string{
				"financial_statement_grouping.delete",
				fmt.Sprintf("financial_statement_grouping.delete.%s", data.ID),
				fmt.Sprintf("financial_statement_grouping.delete.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_grouping.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func FinancialStatementAccountsGroupingSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB, userID uuid.UUID,
	organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()

	FinancialStatementAccountsGrouping := []*types.FinancialStatementAccountsGrouping{
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Assets",
			Description:    "Resources owned by the cooperative that have economic value and can provide future benefits.",
			Debit:          1.0,
			Credit:         0.0,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Liabilities",
			Description:    "Debts and obligations owed by the cooperative to external parties.",
			Debit:          0.0,
			Credit:         1.0,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Liabilities, Equity & Reserves",
			Description:    "Ownership interest of members in the cooperative, including contributed capital and retained earnings.",
			Debit:          0.0,
			Credit:         1.0,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Income",
			Description:    "Income generated from the cooperative's operations and other income-generating activities.",
			Debit:          0.0,
			Credit:         1.0,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Expenses",
			Description:    "Costs incurred in the normal course of business operations and other business activities.",
			Debit:          1.0,
			Credit:         0.0,
		},
	}
	for _, data := range FinancialStatementAccountsGrouping {
		if err := FinancialStatementAccountsGroupingManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed financial statement accounts grouping %s", data.Name)
		}
	}
	return nil
}

func FinancialStatementAccountsGroupingCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.FinancialStatementAccountsGrouping, error) {
	return FinancialStatementAccountsGroupingManager(service).Find(context, &types.FinancialStatementAccountsGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func FinancialStatementAccountsGroupingAlignments(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.FinancialStatementAccountsGrouping, error) {
	fsGroupings, err := FinancialStatementAccountsGroupingManager(service).Find(context, &types.FinancialStatementAccountsGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to get financial statement groupings")
	}
	for _, grouping := range fsGroupings {
		if grouping != nil {
			grouping.FinancialStatementDefinitionEntries = []*types.FinancialStatementDefinition{}
			entries, err := FinancialStatementDefinitionManager(service).ArrFind(context,
				[]query.ArrFilterSQL{
					{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
					{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
					{Field: "financial_statement_grouping_id", Op: query.ModeEqual, Value: grouping.ID},
				},
				[]query.ArrFilterSortSQL{
					{Field: "created_at", Order: query.SortOrderAsc},
				},
			)
			if err != nil {
				return nil, eris.Wrap(err, "failed to get financial statement definition entries")
			}

			var filteredEntries []*types.FinancialStatementDefinition
			for _, entry := range entries {
				if entry.FinancialStatementDefinitionEntriesID == nil {
					filteredEntries = append(filteredEntries, entry)
				}
			}

			grouping.FinancialStatementDefinitionEntries = filteredEntries
		}
	}
	return fsGroupings, nil
}
