package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
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

		// Self-referencing relationship for parent-child hierarchy
		ParentDefinitionID *uuid.UUID                      `gorm:"type:uuid;column:financial_statement_definition_id"`
		ParentDefinition   *FinancialStatementDefinition   `gorm:"foreignKey:ParentDefinitionID" json:"parent_definition,omitempty"`
		ChildDefinitions   []*FinancialStatementDefinition `gorm:"foreignKey:ParentDefinitionID" json:"child_definitions,omitempty"`

		// Many-to-one relationship with FinancialStatementGrouping
		FinancialStatementGroupingID *uuid.UUID                  `gorm:"type:uuid;index" json:"financial_statement_grouping_id,omitempty"`
		FinancialStatementGrouping   *FinancialStatementGrouping `gorm:"foreignKey:FinancialStatementGroupingID;constraint:OnDelete:SET NULL;" json:"grouping,omitempty"`

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

		ParentDefinitionID *uuid.UUID                              `json:"parent_definition_id,omitempty"`
		ParentDefinition   *FinancialStatementDefinitionResponse   `json:"parent_definition,omitempty"`
		ChildDefinitions   []*FinancialStatementDefinitionResponse `json:"child_definitions,omitempty"`

		FinancialStatementGroupingID *uuid.UUID                          `json:"financial_statement_grouping_id,omitempty"`
		FinancialStatementGrouping   *FinancialStatementGroupingResponse `json:"grouping,omitempty"`

		Name                   string `json:"name"`
		Description            string `json:"description"`
		Index                  int    `json:"index"`
		NameInTotal            string `json:"name_in_total"`
		IsPosting              bool   `json:"is_posting"`
		FinancialStatementType string `json:"financial_statement_type"`
	}

	FinancialStatementDefinitionRequest struct {
		Name                           string     `json:"name" validate:"required,min=1,max=255"`
		Description                    string     `json:"description,omitempty"`
		Index                          int        `json:"index,omitempty"`
		NameInTotal                    string     `json:"name_in_total,omitempty"`
		IsPosting                      bool       `json:"is_posting,omitempty"`
		FinancialStatementType         string     `json:"financial_statement_type,omitempty"`
		FinancialStatementDefinitionID *uuid.UUID `json:"financial_statement_definition_id,omitempty"`
	}
)

func (m *Model) FinancialStatementDefinition() {
	m.Migration = append(m.Migration, &FinancialStatementDefinition{})
	m.FinancialStatementDefinitionManager = horizon_services.NewRepository(horizon_services.RepositoryParams[FinancialStatementDefinition, FinancialStatementDefinitionResponse, FinancialStatementDefinitionRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization", "ParentDefinition", "ChildDefinitions", "FinancialStatementGrouping"},
		Service:  m.provider.Service,
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

				ParentDefinitionID:           data.ParentDefinitionID,
				ParentDefinition:             m.FinancialStatementDefinitionManager.ToModel(data.ParentDefinition),
				ChildDefinitions:             m.FinancialStatementDefinitionManager.ToModels(data.ChildDefinitions),
				FinancialStatementGroupingID: data.FinancialStatementGroupingID,
				FinancialStatementGrouping:   m.FinancialStatementGroupingManager.ToModel(data.FinancialStatementGrouping),

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

func (m *Model) FinancialStatementDefinitionCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*FinancialStatementDefinition, error) {
	return m.FinancialStatementDefinitionManager.Find(context, &FinancialStatementDefinition{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
