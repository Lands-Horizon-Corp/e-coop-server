package types

import (
	"time"

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
