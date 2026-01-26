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
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func GeneralLedgerAccountsGroupingManager(service *horizon.HorizonService) *registry.Registry[
	types.GeneralLedgerAccountsGrouping, types.GeneralLedgerAccountsGroupingResponse, types.GeneralLedgerAccountsGroupingRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.GeneralLedgerAccountsGrouping, types.GeneralLedgerAccountsGroupingResponse, types.GeneralLedgerAccountsGroupingRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.GeneralLedgerAccountsGrouping) *types.GeneralLedgerAccountsGroupingResponse {
			if data == nil {
				return nil
			}
			sort.Slice(data.GeneralLedgerDefinitionEntries, func(i, j int) bool {
				return data.GeneralLedgerDefinitionEntries[i].Index < data.GeneralLedgerDefinitionEntries[j].Index
			})
			return &types.GeneralLedgerAccountsGroupingResponse{
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
				Debit:          data.Debit,
				Credit:         data.Credit,
				Name:           data.Name,
				Description:    data.Description,
				FromCode:       data.FromCode,
				ToCode:         data.ToCode,

				GeneralLedgerDefinitionEntries: GeneralLedgerDefinitionManager(service).ToModels(data.GeneralLedgerDefinitionEntries),
			}
		},
		Created: func(data *types.GeneralLedgerAccountsGrouping) registry.Topics {
			return []string{
				"general_ledger_accounts_grouping.create",
				fmt.Sprintf("general_ledger_accounts_grouping.create.%s", data.ID),
				fmt.Sprintf("general_ledger_accounts_grouping.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_accounts_grouping.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.GeneralLedgerAccountsGrouping) registry.Topics {
			return []string{
				"general_ledger_accounts_grouping.update",
				fmt.Sprintf("general_ledger_accounts_grouping.update.%s", data.ID),
				fmt.Sprintf("general_ledger_accounts_grouping.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_accounts_grouping.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.GeneralLedgerAccountsGrouping) registry.Topics {
			return []string{
				"general_ledger_accounts_grouping.delete",
				fmt.Sprintf("general_ledger_accounts_grouping.delete.%s", data.ID),
				fmt.Sprintf("general_ledger_accounts_grouping.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_accounts_grouping.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func generalLedgerAccountsGroupingSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	generalLedgerAccountsGrouping := []*types.GeneralLedgerAccountsGrouping{
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Assets",
			Description:    "Represents resources owned by the organization that have economic value and can provide future benefits.",
			Debit:          0.00,
			Credit:         0.00,
			FromCode:       1000.00,
			ToCode:         1999.99,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Liabilities, Equity & Reserves",
			Description:    "Encompasses the organization's debts, obligations, member equity contributions, and retained earnings reserves.",
			Debit:          0.00,
			Credit:         0.00,
			FromCode:       2000.00,
			ToCode:         3999.99,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Income",
			Description:    "Revenue generated from the organization's primary operations, services, and other income-generating activities.",
			Debit:          0.00,
			Credit:         0.00,
			FromCode:       4000.00,
			ToCode:         4999.99,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Expenses",
			Description:    "Costs incurred in the normal course of business operations, including administrative, operational, and member service expenses.",
			Debit:          0.00,
			Credit:         0.00,
			FromCode:       5000.00,
			ToCode:         5999.99,
		},
	}

	for i, groupingData := range generalLedgerAccountsGrouping {
		if err := GeneralLedgerAccountsGroupingManager(service).CreateWithTx(context, tx, groupingData); err != nil {
			return eris.Wrapf(err, "failed to seed general ledger accounts grouping %s", groupingData.Name)
		}

		var definitions []*types.GeneralLedgerDefinition

		switch i {
		case 0: // Assets
			currentAssetsParent := &types.GeneralLedgerDefinition{
				CreatedAt:                       now,
				UpdatedAt:                       now,
				CreatedByID:                     userID,
				UpdatedByID:                     userID,
				OrganizationID:                  organizationID,
				BranchID:                        branchID,
				GeneralLedgerAccountsGroupingID: &groupingData.ID,
				Name:                            "Current Assets",
				Description:                     "Assets expected to be converted to cash within one year",
				Index:                           0,
				NameInTotal:                     "Current Assets",
				IsPosting:                       false,
				GeneralLedgerType:               "Assets",
				BeginningBalanceOfTheYearCredit: 0,
				BeginningBalanceOfTheYearDebit:  0,
				GeneralLedgerDefinitionEntryID:  nil,
			}
			if err := GeneralLedgerDefinitionManager(service).CreateWithTx(context, tx, currentAssetsParent); err != nil {
				return eris.Wrapf(err, "failed to seed general ledger definition %s", currentAssetsParent.Name)
			}

			definitions = []*types.GeneralLedgerDefinition{
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  &currentAssetsParent.ID,
					Name:                            "Cash on Hand",
					Description:                     "Physical cash and currency held by the organization",
					Index:                           1,
					NameInTotal:                     "Cash on Hand",
					IsPosting:                       true,
					GeneralLedgerType:               "Assets",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  &currentAssetsParent.ID,
					Name:                            "Cash on Bank",
					Description:                     "Funds deposited in bank accounts",
					Index:                           2,
					NameInTotal:                     "Cash on Bank",
					IsPosting:                       true,
					GeneralLedgerType:               "Assets",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  &currentAssetsParent.ID,
					Name:                            "Accounts Receivable",
					Description:                     "Money owed to the organization by members and customers",
					Index:                           3,
					NameInTotal:                     "Accounts Receivable",
					IsPosting:                       true,
					GeneralLedgerType:               "Assets",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  &currentAssetsParent.ID,
					Name:                            "Inventory",
					Description:                     "Goods and materials held for sale or production",
					Index:                           4,
					NameInTotal:                     "Inventory",
					IsPosting:                       true,
					GeneralLedgerType:               "Assets",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  nil,
					Name:                            "Property, Plant & Equipment",
					Description:                     "Long-term physical assets used in operations",
					Index:                           5,
					NameInTotal:                     "PPE",
					IsPosting:                       true,
					GeneralLedgerType:               "Assets",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
			}
		case 1: // Liabilities, Equity & Reserves
			liabilitiesParent := &types.GeneralLedgerDefinition{
				CreatedAt:                       now,
				UpdatedAt:                       now,
				CreatedByID:                     userID,
				UpdatedByID:                     userID,
				OrganizationID:                  organizationID,
				BranchID:                        branchID,
				GeneralLedgerAccountsGroupingID: &groupingData.ID,
				Name:                            "Current Liabilities",
				Description:                     "Short-term debts and obligations",
				Index:                           0,
				NameInTotal:                     "Current Liabilities",
				IsPosting:                       false,
				GeneralLedgerType:               "Liabilities, Equity & Reserves",
				BeginningBalanceOfTheYearCredit: 0,
				BeginningBalanceOfTheYearDebit:  0,
				GeneralLedgerDefinitionEntryID:  nil,
			}

			equityParent := &types.GeneralLedgerDefinition{
				CreatedAt:                       now,
				UpdatedAt:                       now,
				CreatedByID:                     userID,
				UpdatedByID:                     userID,
				OrganizationID:                  organizationID,
				BranchID:                        branchID,
				GeneralLedgerAccountsGroupingID: &groupingData.ID,
				Name:                            "Member Equity",
				Description:                     "Member ownership and retained earnings",
				Index:                           1,
				NameInTotal:                     "Member Equity",
				IsPosting:                       false,
				GeneralLedgerType:               "Liabilities, Equity & Reserves",
				BeginningBalanceOfTheYearCredit: 0,
				BeginningBalanceOfTheYearDebit:  0,
				GeneralLedgerDefinitionEntryID:  nil,
			}

			if err := GeneralLedgerDefinitionManager(service).CreateWithTx(context, tx, liabilitiesParent); err != nil {
				return eris.Wrapf(err, "failed to seed general ledger definition %s", liabilitiesParent.Name)
			}
			if err := GeneralLedgerDefinitionManager(service).CreateWithTx(context, tx, equityParent); err != nil {
				return eris.Wrapf(err, "failed to seed general ledger definition %s", equityParent.Name)
			}

			definitions = []*types.GeneralLedgerDefinition{
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  &liabilitiesParent.ID,
					Name:                            "Accounts Payable",
					Description:                     "Money owed to suppliers and creditors",
					Index:                           2,
					NameInTotal:                     "Accounts Payable",
					IsPosting:                       true,
					GeneralLedgerType:               "Liabilities, Equity & Reserves",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  &liabilitiesParent.ID,
					Name:                            "Member Deposits",
					Description:                     "Funds deposited by cooperative members",
					Index:                           3,
					NameInTotal:                     "Member Deposits",
					IsPosting:                       true,
					GeneralLedgerType:               "Liabilities, Equity & Reserves",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  &equityParent.ID,
					Name:                            "Share Capital",
					Description:                     "Member contributions to cooperative capital",
					Index:                           4,
					NameInTotal:                     "Share Capital",
					IsPosting:                       true,
					GeneralLedgerType:               "Liabilities, Equity & Reserves",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  &equityParent.ID,
					Name:                            "Retained Earnings",
					Description:                     "Accumulated profits retained in the cooperative",
					Index:                           5,
					NameInTotal:                     "Retained Earnings",
					IsPosting:                       true,
					GeneralLedgerType:               "Liabilities, Equity & Reserves",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
			}

		case 2: // Income
			definitions = []*types.GeneralLedgerDefinition{
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  nil,
					Name:                            "Interest Income",
					Description:                     "Income earned from loans and investments",
					Index:                           1,
					NameInTotal:                     "Interest Income",
					IsPosting:                       true,
					GeneralLedgerType:               "Income",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  nil,
					Name:                            "Service Fees",
					Description:                     "Fees collected for various cooperative services",
					Index:                           2,
					NameInTotal:                     "Service Fees",
					IsPosting:                       true,
					GeneralLedgerType:               "Income",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  nil,
					Name:                            "Membership Fees",
					Description:                     "Fees collected from new and existing members",
					Index:                           3,
					NameInTotal:                     "Membership Fees",
					IsPosting:                       true,
					GeneralLedgerType:               "Income",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
			}
		case 3:
			operatingExpensesParent := &types.GeneralLedgerDefinition{
				CreatedAt:                       now,
				UpdatedAt:                       now,
				CreatedByID:                     userID,
				UpdatedByID:                     userID,
				OrganizationID:                  organizationID,
				BranchID:                        branchID,
				GeneralLedgerAccountsGroupingID: &groupingData.ID,
				Name:                            "Operating Expenses",
				Description:                     "General expenses for daily operations",
				Index:                           0,
				NameInTotal:                     "Operating Expenses",
				IsPosting:                       false,
				GeneralLedgerType:               "Expense",
				BeginningBalanceOfTheYearCredit: 0,
				BeginningBalanceOfTheYearDebit:  0,
				GeneralLedgerDefinitionEntryID:  nil,
			}
			if err := GeneralLedgerDefinitionManager(service).CreateWithTx(context, tx, operatingExpensesParent); err != nil {
				return eris.Wrapf(err, "failed to seed general ledger definition %s", operatingExpensesParent.Name)
			}
			definitions = []*types.GeneralLedgerDefinition{
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  &operatingExpensesParent.ID,
					Name:                            "Salaries and Wages",
					Description:                     "Employee compensation and benefits",
					Index:                           1,
					NameInTotal:                     "Salaries and Wages",
					IsPosting:                       true,
					GeneralLedgerType:               "Expense",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  &operatingExpensesParent.ID,
					Name:                            "Utilities Expense",
					Description:                     "Electricity, water, internet, and other utilities",
					Index:                           2,
					NameInTotal:                     "Utilities",
					IsPosting:                       true,
					GeneralLedgerType:               "Expense",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  &operatingExpensesParent.ID,
					Name:                            "Office Supplies",
					Description:                     "Stationery, printing materials, and office consumables",
					Index:                           3,
					NameInTotal:                     "Office Supplies",
					IsPosting:                       true,
					GeneralLedgerType:               "Expense",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
				{
					CreatedAt:                       now,
					UpdatedAt:                       now,
					CreatedByID:                     userID,
					UpdatedByID:                     userID,
					OrganizationID:                  organizationID,
					BranchID:                        branchID,
					GeneralLedgerAccountsGroupingID: &groupingData.ID,
					GeneralLedgerDefinitionEntryID:  &operatingExpensesParent.ID,
					Name:                            "Rent Expense",
					Description:                     "Monthly rental for office space and facilities",
					Index:                           4,
					NameInTotal:                     "Rent",
					IsPosting:                       true,
					GeneralLedgerType:               "Expense",
					BeginningBalanceOfTheYearCredit: 0,
					BeginningBalanceOfTheYearDebit:  0,
				},
			}
		}

		for _, definitionData := range definitions {
			if err := GeneralLedgerDefinitionManager(service).CreateWithTx(context, tx, definitionData); err != nil {
				return eris.Wrapf(err, "failed to seed general ledger definition %s", definitionData.Name)
			}
		}
	}
	return nil
}

func GeneralLedgerAccountsGroupingCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.GeneralLedgerAccountsGrouping, error) {
	return GeneralLedgerAccountsGroupingManager(service).Find(context, &types.GeneralLedgerAccountsGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
