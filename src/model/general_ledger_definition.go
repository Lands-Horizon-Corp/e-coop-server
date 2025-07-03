package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

// GeneralLedgerType define this type according to your domain (e.g. as string or int)
type GeneralLedgerType string // adjust as needed

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

		ParentID *uuid.UUID                 `gorm:"type:uuid" json:"parent_id,omitempty"`
		Parent   *GeneralLedgerDefinition   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
		Children []*GeneralLedgerDefinition `gorm:"foreignKey:ParentID" json:"children,omitempty"`

		Accounts []*Account `gorm:"foreignKey:GeneralLedgerDefinitionID" json:"accounts"`

		GeneralLedgerAccountsGroupingID *uuid.UUID                     `gorm:"type:uuid" json:"general_ledger_accounts_grouping_id,omitempty"`
		GeneralLedgerAccountsGrouping   *GeneralLedgerAccountsGrouping `gorm:"foreignKey:GeneralLedgerAccountsGroupingID;constraint:OnDelete:SET NULL;" json:"grouping,omitempty"`

		Name              string            `gorm:"type:varchar(255);not null;unique"`
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

		GeneralLedgerDefinitionEntryID *uuid.UUID                         `json:"general_ledger_definition_entry_id,omitempty"`
		GeneralLedgerDefinitionEntry   *GeneralLedgerDefinitionResponse   `json:"general_ledger_definition_entry,omitempty"`
		GeneralLedgerDefinitionEntries []*GeneralLedgerDefinitionResponse `json:"general_ledger_definition_entries,omitempty"`

		Accounts []*AccountResponse `json:"accounts"`

		GeneralLedgerAccountsGroupingID *uuid.UUID                             `json:"general_ledger_accounts_grouping_id,omitempty"`
		GeneralLedgerAccountsGrouping   *GeneralLedgerAccountsGroupingResponse `json:"grouping,omitempty"`

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
	}

	GeneralLedgerDefinitionRequest struct {
		OrganizationID                   uuid.UUID         `json:"organization_id" validate:"required"`
		BranchID                         uuid.UUID         `json:"branch_id" validate:"required"`
		Name                             string            `json:"name" validate:"required,min=1,max=255"`
		Description                      string            `json:"description,omitempty"`
		Index                            int               `json:"index,omitempty"`
		NameInTotal                      string            `json:"name_in_total,omitempty"`
		IsPosting                        bool              `json:"is_posting,omitempty"`
		GeneralLedgerType                GeneralLedgerType `json:"general_ledger_type" validate:"required"`
		BeginningBalanceOfTheYearCredit  int               `json:"beginning_balance_of_the_year_credit,omitempty"`
		BeginningBalanceOfTheYearDebit   int               `json:"beginning_balance_of_the_year_debit,omitempty"`
		GeneralLedgerDefinitionEntriesID *uuid.UUID        `json:"general_ledger_definition_entries_id,omitempty"`

		GeneralLedgerAccountsGroupingID *uuid.UUID `json:"general_ledger_accounts_grouping_id,omitempty"`
	}
)

func (m *Model) GeneralLedgerDefinition() {
	m.Migration = append(m.Migration, &GeneralLedgerDefinition{})
	m.GeneralLedgerDefinitionManager = horizon_services.NewRepository(horizon_services.RepositoryParams[GeneralLedgerDefinition, GeneralLedgerDefinitionResponse, GeneralLedgerDefinitionRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
			"Parent",
			"Children",
			"Accounts",
			"GeneralLedgerAccountsGrouping",
		},
		Service: m.provider.Service,
		Resource: func(data *GeneralLedgerDefinition) *GeneralLedgerDefinitionResponse {
			if data == nil {
				return nil
			}
			var deletedAt *string
			if data.DeletedAt.Valid {
				t := data.DeletedAt.Time.Format(time.RFC3339)
				deletedAt = &t
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

				GeneralLedgerDefinitionEntryID: data.ParentID,
				GeneralLedgerDefinitionEntry:   m.GeneralLedgerDefinitionManager.ToModel(data.Parent),
				GeneralLedgerDefinitionEntries: m.GeneralLedgerDefinitionManager.ToModels(data.Children),

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
				GeneralLedgerAccountsGroupingID: data.GeneralLedgerAccountsGroupingID,
				GeneralLedgerAccountsGrouping:   m.GeneralLedgerAccountsGroupingManager.ToModel(data.GeneralLedgerAccountsGrouping),
			}
		},
		Created: func(data *GeneralLedgerDefinition) []string {
			return []string{
				"general_ledger_definition.create",
				fmt.Sprintf("general_ledger_definition.create.%s", data.ID),
				fmt.Sprintf("general_ledger_definition.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_definition.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *GeneralLedgerDefinition) []string {
			return []string{
				"general_ledger_definition.update",
				fmt.Sprintf("general_ledger_definition.update.%s", data.ID),
				fmt.Sprintf("general_ledger_definition.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_definition.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *GeneralLedgerDefinition) []string {
			return []string{
				"general_ledger_definition.delete",
				fmt.Sprintf("general_ledger_definition.delete.%s", data.ID),
				fmt.Sprintf("general_ledger_definition.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_definition.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) GeneralLedgerDefinitionCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*GeneralLedgerDefinition, error) {
	return m.GeneralLedgerDefinitionManager.Find(context, &GeneralLedgerDefinition{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
