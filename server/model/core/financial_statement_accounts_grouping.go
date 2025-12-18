package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	FinancialStatementGrouping struct {
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

		Name        string  `gorm:"type:varchar(50);not null"`
		Description string  `gorm:"type:text;not null"`
		Debit       float64 `gorm:"type:decimal;not null"`
		Credit      float64 `gorm:"type:decimal;not null"`

		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		FinancialStatementDefinitionEntries []*FinancialStatementDefinition `gorm:"foreignKey:FinancialStatementGroupingID" json:"financial_statement_definition_entries,omitempty"`
	}

	FinancialStatementGroupingResponse struct {
		ID                                  uuid.UUID                               `json:"id"`
		OrganizationID                      uuid.UUID                               `json:"organization_id"`
		Organization                        *OrganizationResponse                   `json:"organization,omitempty"`
		BranchID                            uuid.UUID                               `json:"branch_id"`
		Branch                              *BranchResponse                         `json:"branch,omitempty"`
		CreatedByID                         uuid.UUID                               `json:"created_by_id"`
		CreatedBy                           *UserResponse                           `json:"created_by,omitempty"`
		UpdatedByID                         uuid.UUID                               `json:"updated_by_id"`
		UpdatedBy                           *UserResponse                           `json:"updated_by,omitempty"`
		DeletedByID                         *uuid.UUID                              `json:"deleted_by_id,omitempty"`
		DeletedBy                           *UserResponse                           `json:"deleted_by,omitempty"`
		IconMediaID                         *uuid.UUID                              `json:"icon_media_id,omitempty"`
		IconMedia                           *MediaResponse                          `json:"icon_media,omitempty"`
		Name                                string                                  `json:"name"`
		Description                         string                                  `json:"description"`
		Debit                               float64                                 `json:"debit"`
		Credit                              float64                                 `json:"credit"`
		CreatedAt                           string                                  `json:"created_at"`
		UpdatedAt                           string                                  `json:"updated_at"`
		DeletedAt                           *string                                 `json:"deleted_at,omitempty"`
		FinancialStatementDefinitionEntries []*FinancialStatementDefinitionResponse `json:"financial_statement_definition_entries,omitempty"`
	}

	FinancialStatementGroupingRequest struct {
		Name        string     `json:"name" validate:"required,min=1,max=50"`
		Description string     `json:"description" validate:"required"`
		Debit       float64    `json:"debit" validate:"omitempty,gt=0"`
		Credit      float64    `json:"credit" validate:"omitempty,gt=0"`
		IconMediaID *uuid.UUID `json:"icon_media_id,omitempty"`
	}
)

func (m *Core) financialStatementGrouping() {
	m.Migration = append(m.Migration, &FinancialStatementGrouping{})
	m.FinancialStatementGroupingManager = *registry.NewRegistry(registry.RegistryParams[FinancialStatementGrouping, FinancialStatementGroupingResponse, FinancialStatementGroupingRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "IconMedia",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *FinancialStatementGrouping) *FinancialStatementGroupingResponse {
			if data == nil {
				return nil
			}
			var deletedAt *string
			if data.DeletedAt.Valid {
				t := data.DeletedAt.Time.Format(time.RFC3339)
				deletedAt = &t
			}
			return &FinancialStatementGroupingResponse{
				ID:                                  data.ID,
				OrganizationID:                      data.OrganizationID,
				Organization:                        m.OrganizationManager.ToModel(data.Organization),
				BranchID:                            data.BranchID,
				Branch:                              m.BranchManager.ToModel(data.Branch),
				CreatedByID:                         data.CreatedByID,
				CreatedBy:                           m.UserManager.ToModel(data.CreatedBy),
				UpdatedByID:                         data.UpdatedByID,
				UpdatedBy:                           m.UserManager.ToModel(data.UpdatedBy),
				DeletedByID:                         data.DeletedByID,
				DeletedBy:                           m.UserManager.ToModel(data.DeletedBy),
				IconMediaID:                         data.IconMediaID,
				IconMedia:                           m.MediaManager.ToModel(data.IconMedia),
				Name:                                data.Name,
				Description:                         data.Description,
				Debit:                               data.Debit,
				Credit:                              data.Credit,
				CreatedAt:                           data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:                           data.UpdatedAt.Format(time.RFC3339),
				DeletedAt:                           deletedAt,
				FinancialStatementDefinitionEntries: m.FinancialStatementDefinitionManager.ToModels(data.FinancialStatementDefinitionEntries),
			}
		},
		Created: func(data *FinancialStatementGrouping) registry.Topics {
			return []string{
				"financial_statement_grouping.create",
				fmt.Sprintf("financial_statement_grouping.create.%s", data.ID),
				fmt.Sprintf("financial_statement_grouping.create.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_grouping.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *FinancialStatementGrouping) registry.Topics {
			return []string{
				"financial_statement_grouping.update",
				fmt.Sprintf("financial_statement_grouping.update.%s", data.ID),
				fmt.Sprintf("financial_statement_grouping.update.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_grouping.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *FinancialStatementGrouping) registry.Topics {
			return []string{
				"financial_statement_grouping.delete",
				fmt.Sprintf("financial_statement_grouping.delete.%s", data.ID),
				fmt.Sprintf("financial_statement_grouping.delete.branch.%s", data.BranchID),
				fmt.Sprintf("financial_statement_grouping.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) financialStatementGroupingSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()

	financialStatementGrouping := []*FinancialStatementGrouping{
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
	for _, data := range financialStatementGrouping {
		if err := m.FinancialStatementGroupingManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed financial statement accounts grouping %s", data.Name)
		}
	}
	return nil
}

func (m *Core) FinancialStatementGroupingCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*FinancialStatementGrouping, error) {
	return m.FinancialStatementGroupingManager.Find(context, &FinancialStatementGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func (m *Core) FinancialStatementGroupingAlignments(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*FinancialStatementGrouping, error) {
	fsGroupings, err := m.FinancialStatementGroupingManager.Find(context, &FinancialStatementGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to get financial statement groupings")
	}
	for _, grouping := range fsGroupings {
		if grouping != nil {
			grouping.FinancialStatementDefinitionEntries = []*FinancialStatementDefinition{}
			entries, err := m.FinancialStatementDefinitionManager.ArrFind(context,
				[]registry.FilterSQL{
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

			var filteredEntries []*FinancialStatementDefinition
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
