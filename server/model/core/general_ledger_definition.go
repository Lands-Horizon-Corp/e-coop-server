package core

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	GeneralLedgerDefinition struct {
		ID             uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger_definition"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger_definition"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		CreatedByID uuid.UUID  `gorm:"type:uuid"`
		CreatedBy   *User      `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedByID uuid.UUID  `gorm:"type:uuid"`
		UpdatedBy   *User      `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedByID *uuid.UUID `gorm:"type:uuid"`
		DeletedBy   *User      `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		GeneralLedgerDefinitionEntryID *uuid.UUID                 `gorm:"type:uuid" json:"general_ledger_definition_entries_id,omitempty"`
		GeneralLedgerDefinitionEntries []*GeneralLedgerDefinition `gorm:"foreignKey:GeneralLedgerDefinitionEntryID" json:"general_ledger_definition_entries,omitempty"`

		GeneralLedgerAccountsGroupingID *uuid.UUID                     `gorm:"type:uuid" json:"general_ledger_accounts_grouping_id,omitempty"`
		GeneralLedgerAccountsGrouping   *GeneralLedgerAccountsGrouping `gorm:"foreignKey:GeneralLedgerAccountsGroupingID" json:"general_ledger_accounts_grouping,omitempty"`

		Accounts []*Account `gorm:"foreignKey:GeneralLedgerDefinitionID" json:"accounts"`

		Name              string            `gorm:"type:varchar(255);not null;"`
		Description       string            `gorm:"type:text"`
		Index             int               `gorm:"default:0"`
		NameInTotal       string            `gorm:"type:varchar(255)"`
		IsPosting         bool              `gorm:"default:false"`
		GeneralLedgerType GeneralLedgerType `gorm:"type:varchar(50);not null"`

		BeginningBalanceOfTheYearCredit int `gorm:"default:0"`
		BeginningBalanceOfTheYearDebit  int `gorm:"default:0"`

		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`
	}


	GeneralLedgerDefinitionResponse struct {
		ID             uuid.UUID             `json:"id"`
		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       uuid.UUID             `json:"branch_id"`
		Branch         *BranchResponse       `json:"branch,omitempty"`
		CreatedByID    uuid.UUID             `json:"created_by_id"`
		CreatedBy      *UserResponse         `json:"created_by,omitempty"`
		UpdatedByID    uuid.UUID             `json:"updated_by_id"`
		UpdatedBy      *UserResponse         `json:"updated_by,omitempty"`
		DeletedByID    *uuid.UUID            `json:"deleted_by_id,omitempty"`
		DeletedBy      *UserResponse         `json:"deleted_by,omitempty"`

		GeneralLedgerDefinitionEntryID *uuid.UUID                         `json:"general_ledger_definition_entries_id,omitempty"`
		GeneralLedgerDefinitionEntries []*GeneralLedgerDefinitionResponse `json:"general_ledger_definition,omitempty"`

		GeneralLedgerAccountsGroupingID *uuid.UUID                             `json:"general_ledger_accounts_grouping_id,omitempty"`
		GeneralLedgerAccountsGrouping   *GeneralLedgerAccountsGroupingResponse `json:"general_ledger_accounts_grouping,omitempty"`

		Accounts []*AccountResponse `json:"accounts"`

		Name                            string            `json:"name"`
		Description                     string            `json:"description"`
		Index                           int               `json:"index"`
		NameInTotal                     string            `json:"name_in_total"`
		IsPosting                       bool              `json:"is_posting"`
		GeneralLedgerType               GeneralLedgerType `json:"general_ledger_type"`
		BeginningBalanceOfTheYearCredit int               `json:"beginning_balance_of_the_year_credit"`
		BeginningBalanceOfTheYearDebit  int               `json:"beginning_balance_of_the_year_debit"`
		CreatedAt                       string            `json:"created_at"`
		UpdatedAt                       string            `json:"updated_at"`
		DeletedAt                       *string           `json:"deleted_at,omitempty"`
		Depth                           int               `json:"depth"`
	}


	GeneralLedgerDefinitionRequest struct {
		Name                            string            `json:"name" validate:"required,min=1,max=255"`
		Description                     string            `json:"description,omitempty"`
		Index                           int               `json:"index,omitempty"`
		NameInTotal                     string            `json:"name_in_total,omitempty"`
		IsPosting                       bool              `json:"is_posting,omitempty"`
		GeneralLedgerType               GeneralLedgerType `json:"general_ledger_type" validate:"required"`
		BeginningBalanceOfTheYearCredit int               `json:"beginning_balance_of_the_year_credit,omitempty"`
		BeginningBalanceOfTheYearDebit  int               `json:"beginning_balance_of_the_year_debit,omitempty"`
		GeneralLedgerDefinitionEntryID  *uuid.UUID        `json:"general_ledger_definition_entries_id,omitempty"`
		GeneralLedgerAccountsGroupingID *uuid.UUID        `json:"general_ledger_accounts_grouping_id,omitempty"`
	}
)

func (m *Core) generalLedgerDefinition() {
	m.Migration = append(m.Migration, &GeneralLedgerDefinition{})
	m.GeneralLedgerDefinitionManager = *registry.NewRegistry(registry.RegistryParams[GeneralLedgerDefinition, GeneralLedgerDefinitionResponse, GeneralLedgerDefinitionRequest]{
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
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *GeneralLedgerDefinition) *GeneralLedgerDefinitionResponse {
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

			entries := m.GeneralLedgerDefinitionManager.ToModels(data.GeneralLedgerDefinitionEntries)
			if len(entries) == 0 || entries == nil {
				entries = []*GeneralLedgerDefinitionResponse{}
			}
			return &GeneralLedgerDefinitionResponse{
				ID:             data.ID,
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager.ToModel(data.UpdatedBy),
				DeletedByID:    data.DeletedByID,
				DeletedBy:      m.UserManager.ToModel(data.DeletedBy),

				GeneralLedgerDefinitionEntryID:  data.GeneralLedgerDefinitionEntryID,
				GeneralLedgerDefinitionEntries:  entries,
				GeneralLedgerAccountsGroupingID: data.GeneralLedgerAccountsGroupingID,
				GeneralLedgerAccountsGrouping:   m.GeneralLedgerAccountsGroupingManager.ToModel(data.GeneralLedgerAccountsGrouping),

				Accounts:                        m.AccountManager.ToModels(data.Accounts),
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
		Created: func(data *GeneralLedgerDefinition) registry.Topics {
			return []string{
				"general_ledger_definition.create",
				fmt.Sprintf("general_ledger_definition.create.%s", data.ID),
				fmt.Sprintf("general_ledger_definition.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_definition.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *GeneralLedgerDefinition) registry.Topics {
			return []string{
				"general_ledger_definition.update",
				fmt.Sprintf("general_ledger_definition.update.%s", data.ID),
				fmt.Sprintf("general_ledger_definition.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_definition.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *GeneralLedgerDefinition) registry.Topics {
			return []string{
				"general_ledger_definition.delete",
				fmt.Sprintf("general_ledger_definition.delete.%s", data.ID),
				fmt.Sprintf("general_ledger_definition.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_definition.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) GeneralLedgerDefinitionCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*GeneralLedgerDefinition, error) {
	return m.GeneralLedgerDefinitionManager.Find(context, &GeneralLedgerDefinition{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
