package model_core

import (
	"context"
	"fmt"
	"sort"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	GeneralLedgerAccountsGrouping struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt   time.Time      `gorm:"not null;default:now()"`
		CreatedByID uuid.UUID      `gorm:"type:uuid"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger_accounts_grouping"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger_accounts_grouping"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Debit       float64 `gorm:"type:decimal"`
		Credit      float64 `gorm:"type:decimal"`
		Name        string  `gorm:"type:varchar(50);not null"`
		Description string  `gorm:"type:text;not null"`
		FromCode    float64 `gorm:"type:decimal;default:0"`
		ToCode      float64 `gorm:"type:decimal;default:0"`

		GeneralLedgerDefinitionEntries []*GeneralLedgerDefinition `gorm:"foreignKey:GeneralLedgerAccountsGroupingID" json:"general_ledger_definition,omitempty"`
	}

	GeneralLedgerAccountsGroupingResponse struct {
		ID             uuid.UUID             `json:"id"`
		CreatedAt      string                `json:"created_at"`
		CreatedByID    uuid.UUID             `json:"created_by_id"`
		CreatedBy      *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt      string                `json:"updated_at"`
		UpdatedByID    uuid.UUID             `json:"updated_by_id"`
		UpdatedBy      *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       uuid.UUID             `json:"branch_id"`
		Branch         *BranchResponse       `json:"branch,omitempty"`
		Debit          float64               `json:"debit"`
		Credit         float64               `json:"credit"`
		Name           string                `json:"name"`
		Description    string                `json:"description"`
		FromCode       float64               `json:"from_code"`
		ToCode         float64               `json:"to_code"`

		GeneralLedgerDefinitionEntries []*GeneralLedgerDefinitionResponse `json:"general_ledger_definition,omitempty"`
	}

	GeneralLedgerAccountsGroupingRequest struct {
		Debit       float64 `json:"debit" validate:"omitempty,gt=0"`
		Credit      float64 `json:"credit" validate:"omitempty,gt=0"`
		Name        string  `json:"name" validate:"required,min=1,max=50"`
		Description string  `json:"description" validate:"required"`
		FromCode    float64 `json:"from_code,omitempty"`
		ToCode      float64 `json:"to_code,omitempty"`
	}
)

func (m *ModelCore) GeneralLedgerAccountsGrouping() {
	m.Migration = append(m.Migration, &GeneralLedgerAccountsGrouping{})
	m.GeneralLedgerAccountsGroupingManager = horizon_services.NewRepository(horizon_services.RepositoryParams[GeneralLedgerAccountsGrouping, GeneralLedgerAccountsGroupingResponse, GeneralLedgerAccountsGroupingRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
		},
		Service: m.provider.Service,
		Resource: func(data *GeneralLedgerAccountsGrouping) *GeneralLedgerAccountsGroupingResponse {
			if data == nil {
				return nil
			}
			sort.Slice(data.GeneralLedgerDefinitionEntries, func(i, j int) bool {
				return data.GeneralLedgerDefinitionEntries[i].Index < data.GeneralLedgerDefinitionEntries[j].Index
			})
			return &GeneralLedgerAccountsGroupingResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),
				Debit:          data.Debit,
				Credit:         data.Credit,
				Name:           data.Name,
				Description:    data.Description,
				FromCode:       data.FromCode,
				ToCode:         data.ToCode,

				GeneralLedgerDefinitionEntries: m.GeneralLedgerDefinitionManager.ToModels(data.GeneralLedgerDefinitionEntries),
			}
		},
		Created: func(data *GeneralLedgerAccountsGrouping) []string {
			return []string{
				"general_ledger_accounts_grouping.create",
				fmt.Sprintf("general_ledger_accounts_grouping.create.%s", data.ID),
				fmt.Sprintf("general_ledger_accounts_grouping.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_accounts_grouping.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *GeneralLedgerAccountsGrouping) []string {
			return []string{
				"general_ledger_accounts_grouping.update",
				fmt.Sprintf("general_ledger_accounts_grouping.update.%s", data.ID),
				fmt.Sprintf("general_ledger_accounts_grouping.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_accounts_grouping.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *GeneralLedgerAccountsGrouping) []string {
			return []string{
				"general_ledger_accounts_grouping.delete",
				fmt.Sprintf("general_ledger_accounts_grouping.delete.%s", data.ID),
				fmt.Sprintf("general_ledger_accounts_grouping.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_accounts_grouping.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) GeneralLedgerAccountsGroupingSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	generalLedgerAccountsGrouping := []*GeneralLedgerAccountsGrouping{
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

	// Create groupings and their definitions
	for i, groupingData := range generalLedgerAccountsGrouping {
		if err := m.GeneralLedgerAccountsGroupingManager.CreateWithTx(context, tx, groupingData); err != nil {
			return eris.Wrapf(err, "failed to seed general ledger accounts grouping %s", groupingData.Name)
		}

		// Create definitions for each grouping
		var definitions []*GeneralLedgerDefinition

		switch i {
		case 0: // Assets
			currentAssetsParent := &GeneralLedgerDefinition{
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
			if err := m.GeneralLedgerDefinitionManager.CreateWithTx(context, tx, currentAssetsParent); err != nil {
				return eris.Wrapf(err, "failed to seed general ledger definition %s", currentAssetsParent.Name)
			}

			// Now create children with ParentID reference
			definitions = []*GeneralLedgerDefinition{
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
			// Create parent for liabilities
			liabilitiesParent := &GeneralLedgerDefinition{
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

			// Create parent for equity
			equityParent := &GeneralLedgerDefinition{
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

			// Create parents first
			if err := m.GeneralLedgerDefinitionManager.CreateWithTx(context, tx, liabilitiesParent); err != nil {
				return eris.Wrapf(err, "failed to seed general ledger definition %s", liabilitiesParent.Name)
			}
			if err := m.GeneralLedgerDefinitionManager.CreateWithTx(context, tx, equityParent); err != nil {
				return eris.Wrapf(err, "failed to seed general ledger definition %s", equityParent.Name)
			}

			definitions = []*GeneralLedgerDefinition{
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
			definitions = []*GeneralLedgerDefinition{
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
			operatingExpensesParent := &GeneralLedgerDefinition{
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
			// Create parent first
			if err := m.GeneralLedgerDefinitionManager.CreateWithTx(context, tx, operatingExpensesParent); err != nil {
				return eris.Wrapf(err, "failed to seed general ledger definition %s", operatingExpensesParent.Name)
			}
			definitions = []*GeneralLedgerDefinition{
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
			if err := m.GeneralLedgerDefinitionManager.CreateWithTx(context, tx, definitionData); err != nil {
				return eris.Wrapf(err, "failed to seed general ledger definition %s", definitionData.Name)
			}
		}
	}
	return nil
}

func (m *ModelCore) GeneralLedgerAccountsGroupingCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*GeneralLedgerAccountsGrouping, error) {
	return m.GeneralLedgerAccountsGroupingManager.Find(context, &GeneralLedgerAccountsGrouping{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
