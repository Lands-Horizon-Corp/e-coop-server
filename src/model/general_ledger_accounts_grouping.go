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

		GeneralLedgerDefinitions []*GeneralLedgerDefinition `gorm:"foreignKey:GeneralLedgerAccountsGroupingID" json:"general_ledger_definitions,omitempty"`
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

		GeneralLedgerDefinitions []*GeneralLedgerDefinitionResponse `json:"general_ledger_definitions,omitempty"`
	}

	GeneralLedgerAccountsGroupingRequest struct {
		Debit       string  `json:"debit" validate:"required"`
		Credit      string  `json:"credit" validate:"required"`
		Name        string  `json:"name" validate:"required,min=1,max=50"`
		Description string  `json:"description" validate:"required"`
		FromCode    float64 `json:"from_code,omitempty"`
		ToCode      float64 `json:"to_code,omitempty"`
	}
)

func (m *Model) GeneralLedgerAccountsGrouping() {
	m.Migration = append(m.Migration, &GeneralLedgerAccountsGrouping{})
	m.GeneralLedgerAccountsGroupingManager = horizon_services.NewRepository(horizon_services.RepositoryParams[GeneralLedgerAccountsGrouping, GeneralLedgerAccountsGroupingResponse, GeneralLedgerAccountsGroupingRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
			"GeneralLedgerDefinitions",
			"GeneralLedgerDefinitions.Accounts",
		},
		Service: m.provider.Service,
		Resource: func(data *GeneralLedgerAccountsGrouping) *GeneralLedgerAccountsGroupingResponse {
			if data == nil {
				return nil
			}
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

				GeneralLedgerDefinitions: m.GeneralLedgerDefinitionManager.ToModels(data.GeneralLedgerDefinitions),
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

func (m *Model) GeneralLedgerAccountsGroupingCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*GeneralLedgerAccountsGrouping, error) {
	return m.GeneralLedgerAccountsGroupingManager.Find(context, &GeneralLedgerAccountsGrouping{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
