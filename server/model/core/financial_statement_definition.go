package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// FinancialStatementDefinition represents the FinancialStatementDefinition model.
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

		// Self-referencing relationship for parent-child hierarchy
		FinancialStatementDefinitionEntriesID *uuid.UUID                      `gorm:"type:uuid;column:financial_statement_definition_entries_id;index" json:"parent_definition_id,omitempty"`
		FinancialStatementDefinitionEntries   []*FinancialStatementDefinition `gorm:"foreignKey:FinancialStatementDefinitionEntriesID" json:"financial_statement_definition_entries,omitempty"`

		// Many-to-one relationship with FinancialStatementGrouping
		FinancialStatementGroupingID *uuid.UUID                  `gorm:"type:uuid;index" json:"financial_statement_grouping_id,omitempty"`
		FinancialStatementGrouping   *FinancialStatementGrouping `gorm:"foreignKey:FinancialStatementGroupingID;constraint:OnDelete:SET NULL;" json:"grouping,omitempty"`

		Accounts []*Account `gorm:"foreignKey:FinancialStatementDefinitionID" json:"accounts"`

		Name                   string `gorm:"type:varchar(255);not null;unique"`
		Description            string `gorm:"type:text"`
		Index                  int    `gorm:"default:0"`
		NameInTotal            string `gorm:"type:varchar(255)"`
		IsPosting              bool   `gorm:"default:false"`
		FinancialStatementType string `gorm:"type:varchar(255)"`
	}

	// FinancialStatementDefinitionResponse represents the response structure for financialstatementdefinition data

	// FinancialStatementDefinitionResponse represents the response structure for FinancialStatementDefinition.
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

		FinancialStatementGroupingID *uuid.UUID                          `json:"financial_statement_grouping_id,omitempty"`
		FinancialStatementGrouping   *FinancialStatementGroupingResponse `json:"grouping,omitempty"`
		Accounts                     []*AccountResponse                  `json:"accounts,omitempty"`
		Name                         string                              `json:"name"`
		Description                  string                              `json:"description"`
		Index                        int                                 `json:"index"`
		NameInTotal                  string                              `json:"name_in_total"`
		IsPosting                    bool                                `json:"is_posting"`
		FinancialStatementType       string                              `json:"financial_statement_type"`
	}

	// FinancialStatementDefinitionRequest represents the request structure for creating/updating financialstatementdefinition

	// FinancialStatementDefinitionRequest represents the request structure for FinancialStatementDefinition.
	FinancialStatementDefinitionRequest struct {
		Name                                  string     `json:"name" validate:"required,min=1,max=255"`
		Description                           string     `json:"description,omitempty"`
		Index                                 int        `json:"index,omitempty"`
		NameInTotal                           string     `json:"name_in_total,omitempty"`
		IsPosting                             bool       `json:"is_posting,omitempty"`
		FinancialStatementType                string     `json:"financial_statement_type,omitempty"`
		FinancialStatementDefinitionEntriesID *uuid.UUID `json:"financial_statement_definition_entries_id,omitempty"`
		FinancialStatementGroupingID          *uuid.UUID `json:"financial_statement_grouping_id,omitempty"`
	}
)

func (m *Core) financialStatementDefinition() {
	m.Migration = append(m.Migration, &FinancialStatementDefinition{})
	m.FinancialStatementDefinitionManager = *registry.NewRegistry(registry.RegistryParams[FinancialStatementDefinition, FinancialStatementDefinitionResponse, FinancialStatementDefinitionRequest]{
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
			// Preload accounts for each level
			"FinancialStatementDefinitionEntries.Accounts",
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.Accounts",
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.Accounts",
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.Accounts",
			"FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.FinancialStatementDefinitionEntries.Accounts",
		},
		Service: m.provider.Service,
		Resource: func(data *FinancialStatementDefinition) *FinancialStatementDefinitionResponse {
			if data == nil {
				return nil
			}
			return &FinancialStatementDefinitionResponse{
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

				FinancialStatementDefinitionEntriesID: data.FinancialStatementDefinitionEntriesID,
				FinancialStatementDefinitionEntries:   m.FinancialStatementDefinitionManager.ToModels(data.FinancialStatementDefinitionEntries),

				FinancialStatementGroupingID: data.FinancialStatementGroupingID,
				FinancialStatementGrouping:   m.FinancialStatementGroupingManager.ToModel(data.FinancialStatementGrouping),
				Accounts:                     m.AccountManager.ToModels(data.Accounts),

				Name:                   data.Name,
				Description:            data.Description,
				Index:                  data.Index,
				NameInTotal:            data.NameInTotal,
				IsPosting:              data.IsPosting,
				FinancialStatementType: data.FinancialStatementType,
			}
		},
		Created: func(data *FinancialStatementDefinition) []string {
			return []string{
				"financial_statement_definition.create",
				fmt.Sprintf("financial_statement_definition.create.%s", data.ID),
				fmt.Sprintf("financial_statement_definition.create.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_definition.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *FinancialStatementDefinition) []string {
			return []string{
				"financial_statement_definition.update",
				fmt.Sprintf("financial_statement_definition.update.%s", data.ID),
				fmt.Sprintf("financial_statement_definition.update.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_definition.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *FinancialStatementDefinition) []string {
			return []string{
				"financial_statement_definition.delete",
				fmt.Sprintf("financial_statement_definition.delete.%s", data.ID),
				fmt.Sprintf("financial_statement_definition.delete.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_definition.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// FinancialStatementDefinitionCurrentBranch returns FinancialStatementDefinitionCurrentBranch for the current branch or organization where applicable.
func (m *Core) FinancialStatementDefinitionCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*FinancialStatementDefinition, error) {
	return m.FinancialStatementDefinitionManager.Find(context, &FinancialStatementDefinition{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
