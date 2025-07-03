package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

// Define AccountingPrinciple type as needed (e.g., string, int, or custom type)
type AccountingPrinciple string

type (
	FinancialStatementsrouping struct {
		ID             uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_financial_statement_grouping"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_financial_statement_grouping"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		CreatedByID uuid.UUID  `gorm:"type:uuid"`
		CreatedBy   *User      `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedByID uuid.UUID  `gorm:"type:uuid"`
		UpdatedBy   *User      `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedByID *uuid.UUID `gorm:"type:uuid"`
		DeletedBy   *User      `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		IconMediaID *uuid.UUID `gorm:"type:uuid"`
		IconMedia   *Media     `gorm:"foreignKey:IconMediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"icon_media,omitempty"`

		Name        string              `gorm:"type:varchar(50);not null"`
		Description string              `gorm:"type:text;not null"`
		Debit       AccountingPrinciple `gorm:"type:varchar(50);not null"`
		Credit      AccountingPrinciple `gorm:"type:varchar(50);not null"`
		Code        float64             `gorm:"type:decimal;not null"`

		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		FinancialStatementDefinitions []*FinancialStatementDefinition `gorm:"foreignKey:FinancialStatementAccountsGroupingID" json:"financial_statement_definitions,omitempty"`
	}

	FinancialStatementGroupingResponse struct {
		ID                            uuid.UUID                               `json:"id"`
		OrganizationID                uuid.UUID                               `json:"organization_id"`
		Organization                  *OrganizationResponse                   `json:"organization,omitempty"`
		BranchID                      uuid.UUID                               `json:"branch_id"`
		Branch                        *BranchResponse                         `json:"branch,omitempty"`
		CreatedByID                   uuid.UUID                               `json:"created_by_id"`
		CreatedBy                     *UserResponse                           `json:"created_by,omitempty"`
		UpdatedByID                   uuid.UUID                               `json:"updated_by_id"`
		UpdatedBy                     *UserResponse                           `json:"updated_by,omitempty"`
		DeletedByID                   *uuid.UUID                              `json:"deleted_by_id,omitempty"`
		DeletedBy                     *UserResponse                           `json:"deleted_by,omitempty"`
		IconMediaID                   *uuid.UUID                              `json:"icon_media_id,omitempty"`
		IconMedia                     *MediaResponse                          `json:"icon_media,omitempty"`
		Name                          string                                  `json:"name"`
		Description                   string                                  `json:"description"`
		Debit                         AccountingPrinciple                     `json:"debit"`
		Credit                        AccountingPrinciple                     `json:"credit"`
		Code                          float64                                 `json:"code"`
		CreatedAt                     string                                  `json:"created_at"`
		UpdatedAt                     string                                  `json:"updated_at"`
		DeletedAt                     *string                                 `json:"deleted_at,omitempty"`
		FinancialStatementDefinitions []*FinancialStatementDefinitionResponse `json:"financial_statement_definitions,omitempty"`
	}

	FinancialStatementGroupingRequest struct {
		OrganizationID uuid.UUID           `json:"organization_id" validate:"required"`
		BranchID       uuid.UUID           `json:"branch_id" validate:"required"`
		Name           string              `json:"name" validate:"required,min=1,max=50"`
		Description    string              `json:"description" validate:"required"`
		Debit          AccountingPrinciple `json:"debit" validate:"required"`
		Credit         AccountingPrinciple `json:"credit" validate:"required"`
		Code           float64             `json:"code" validate:"required"`
		IconMediaID    *uuid.UUID          `json:"icon_media_id,omitempty"`
	}
)

func (m *Model) FinancialStatementAccountsGrouping() {
	m.Migration = append(m.Migration, &FinancialStatementsrouping{})
	m.FinancialStatementGroupingManager = horizon_services.NewRepository(horizon_services.RepositoryParams[FinancialStatementsrouping, FinancialStatementGroupingResponse, FinancialStatementGroupingRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization", "IconMedia",
			"FinancialStatementDefinitions",
		},
		Service: m.provider.Service,
		Resource: func(data *FinancialStatementsrouping) *FinancialStatementGroupingResponse {
			if data == nil {
				return nil
			}
			var deletedAt *string
			if data.DeletedAt.Valid {
				t := data.DeletedAt.Time.Format(time.RFC3339)
				deletedAt = &t
			}
			return &FinancialStatementGroupingResponse{
				ID:                            data.ID,
				OrganizationID:                data.OrganizationID,
				Organization:                  m.OrganizationManager.ToModel(data.Organization),
				BranchID:                      data.BranchID,
				Branch:                        m.BranchManager.ToModel(data.Branch),
				CreatedByID:                   data.CreatedByID,
				CreatedBy:                     m.UserManager.ToModel(data.CreatedBy),
				UpdatedByID:                   data.UpdatedByID,
				UpdatedBy:                     m.UserManager.ToModel(data.UpdatedBy),
				DeletedByID:                   data.DeletedByID,
				DeletedBy:                     m.UserManager.ToModel(data.DeletedBy),
				IconMediaID:                   data.IconMediaID,
				IconMedia:                     m.MediaManager.ToModel(data.IconMedia),
				Name:                          data.Name,
				Description:                   data.Description,
				Debit:                         data.Debit,
				Credit:                        data.Credit,
				Code:                          data.Code,
				CreatedAt:                     data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:                     data.UpdatedAt.Format(time.RFC3339),
				DeletedAt:                     deletedAt,
				FinancialStatementDefinitions: m.FinancialStatementDefinitionManager.ToModels(data.FinancialStatementDefinitions),
			}
		},
		Created: func(data *FinancialStatementsrouping) []string {
			return []string{
				"financial_statement_grouping.create",
				fmt.Sprintf("financial_statement_grouping.create.%s", data.ID),
				fmt.Sprintf("financial_statement_grouping.create.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_grouping.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *FinancialStatementsrouping) []string {
			return []string{
				"financial_statement_grouping.update",
				fmt.Sprintf("financial_statement_grouping.update.%s", data.ID),
				fmt.Sprintf("financial_statement_grouping.update.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_grouping.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *FinancialStatementsrouping) []string {
			return []string{
				"financial_statement_grouping.delete",
				fmt.Sprintf("financial_statement_grouping.delete.%s", data.ID),
				fmt.Sprintf("financial_statement_grouping.delete.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_grouping.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) FinancialStatementAccountsGroupingCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*FinancialStatementsrouping, error) {
	return m.FinancialStatementGroupingManager.Find(context, &FinancialStatementsrouping{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
