package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	FinancialStatementDefinition struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_financial_statement_definition"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_financial_statement_definition"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		FinancialStatementDefinitionEntriesID *uuid.UUID                      `gorm:"type:uuid;column:financial_statement_definition_entries_id;index" json:"parent_definition_id,omitempty"`
		FinancialStatementDefinitionEntries   []*FinancialStatementDefinition `gorm:"foreignKey:FinancialStatementDefinitionEntriesID" json:"financial_statement_definition_entries,omitempty"`

		FinancialStatementAccountsGroupingID *uuid.UUID                          `gorm:"type:uuid;index" json:"financial_statement_grouping_id,omitempty"`
		FinancialStatementAccountsGrouping   *FinancialStatementAccountsGrouping `gorm:"foreignKey:FinancialStatementAccountsGroupingID;constraint:OnDelete:SET NULL;" json:"grouping,omitempty"`

		Accounts []*Account `gorm:"foreignKey:FinancialStatementDefinitionID" json:"accounts"`

		Name                   string `gorm:"type:varchar(255);not null;unique"`
		Description            string `gorm:"type:text"`
		Index                  int    `gorm:"default:0"`
		NameInTotal            string `gorm:"type:varchar(255)"`
		IsPosting              bool   `gorm:"default:false"`
		FinancialStatementType string `gorm:"type:varchar(255)"`
	}

	FinancialStatementDefinitionResponse struct {
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

		FinancialStatementDefinitionEntriesID *uuid.UUID                              `json:"financial_statement_definition_entries_id,omitempty"`
		FinancialStatementDefinitionEntries   []*FinancialStatementDefinitionResponse `json:"financial_statement_definition_entries,omitempty"`

		FinancialStatementAccountsGroupingID *uuid.UUID                                  `json:"financial_statement_grouping_id,omitempty"`
		FinancialStatementAccountsGrouping   *FinancialStatementAccountsGroupingResponse `json:"grouping,omitempty"`
		Accounts                             []*AccountResponse                          `json:"accounts,omitempty"`
		Name                                 string                                      `json:"name"`
		Description                          string                                      `json:"description"`
		Index                                int                                         `json:"index"`
		NameInTotal                          string                                      `json:"name_in_total"`
		IsPosting                            bool                                        `json:"is_posting"`
		FinancialStatementType               string                                      `json:"financial_statement_type"`
	}

	FinancialStatementDefinitionRequest struct {
		Name                                  string     `json:"name" validate:"required,min=1,max=255"`
		Description                           string     `json:"description,omitempty"`
		Index                                 int        `json:"index,omitempty"`
		NameInTotal                           string     `json:"name_in_total,omitempty"`
		IsPosting                             bool       `json:"is_posting,omitempty"`
		FinancialStatementType                string     `json:"financial_statement_type,omitempty"`
		FinancialStatementDefinitionEntriesID *uuid.UUID `json:"financial_statement_definition_entries_id,omitempty"`
		FinancialStatementAccountsGroupingID  *uuid.UUID `json:"financial_statement_grouping_id,omitempty"`
	}
)
