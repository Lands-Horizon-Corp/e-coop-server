package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// MutualFundTable represents the MutualFundTable model.
	MutualFundTable struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_mutual_fund_table" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_mutual_fund_table" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MutualFundID uuid.UUID   `gorm:"type:uuid;not null;index:idx_mutual_fund_table" json:"mutual_fund_id"`
		MutualFund   *MutualFund `gorm:"foreignKey:MutualFundID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"mutual_fund,omitempty"`

		MonthFrom int     `gorm:"not null" json:"month_from"`
		MonthTo   int     `gorm:"not null" json:"month_to"`
		Amount    float64 `gorm:"type:decimal(15,4);not null" json:"amount"`
	}

	// MutualFundTableResponse represents the response structure for MutualFundTable.
	MutualFundTableResponse struct {
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
		MutualFundID   uuid.UUID             `json:"mutual_fund_id"`
		MutualFund     *MutualFundResponse   `json:"mutual_fund,omitempty"`
		MonthFrom      int                   `json:"month_from"`
		MonthTo        int                   `json:"month_to"`
		Amount         float64               `json:"amount"`
	}

	// MutualFundTableRequest represents the request structure for MutualFundTable.
	MutualFundTableRequest struct {
		ID           *uuid.UUID `json:"id,omitempty"`
		MutualFundID uuid.UUID  `json:"mutual_fund_id" validate:"required"`
		MonthFrom    int        `json:"month_from" validate:"required,gte=1"`
		MonthTo      int        `json:"month_to" validate:"required,gte=1"`
		Amount       float64    `json:"amount" validate:"required,gte=0"`
	}
)

func (m *Core) mutualFundTable() {
	m.Migration = append(m.Migration, &MutualFundTable{})
	m.MutualFundTableManager = *registry.NewRegistry(registry.RegistryParams[MutualFundTable, MutualFundTableResponse, MutualFundTableRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Organization", "Branch", "MutualFund"},
		Database: m.provider.Service.Database.Client(),
Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		}
		Resource: func(data *MutualFundTable) *MutualFundTableResponse {
			if data == nil {
				return nil
			}
			return &MutualFundTableResponse{
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
				MutualFundID:   data.MutualFundID,
				MutualFund:     m.MutualFundManager.ToModel(data.MutualFund),
				MonthFrom:      data.MonthFrom,
				MonthTo:        data.MonthTo,
				Amount:         data.Amount,
			}
		},
		Created: func(data *MutualFundTable) []string {
			return []string{
				"mutual_fund_table.create",
				fmt.Sprintf("mutual_fund_table.create.%s", data.ID),
				fmt.Sprintf("mutual_fund_table.create.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_table.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_table.create.mutual_fund.%s", data.MutualFundID),
			}
		},
		Updated: func(data *MutualFundTable) []string {
			return []string{
				"mutual_fund_table.update",
				fmt.Sprintf("mutual_fund_table.update.%s", data.ID),
				fmt.Sprintf("mutual_fund_table.update.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_table.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_table.update.mutual_fund.%s", data.MutualFundID),
			}
		},
		Deleted: func(data *MutualFundTable) []string {
			return []string{
				"mutual_fund_table.delete",
				fmt.Sprintf("mutual_fund_table.delete.%s", data.ID),
				fmt.Sprintf("mutual_fund_table.delete.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_table.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_table.delete.mutual_fund.%s", data.MutualFundID),
			}
		},
	})
}

// MutualFundTableCurrentBranch retrieves all mutual fund tables associated with the specified organization and branch.
func (m *Core) MutualFundTableCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MutualFundTable, error) {
	return m.MutualFundTableManager.Find(context, &MutualFundTable{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

// MutualFundTableByMutualFund retrieves all table entries for a specific mutual fund.
func (m *Core) MutualFundTableByMutualFund(context context.Context, mutualFundID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*MutualFundTable, error) {
	return m.MutualFundTableManager.Find(context, &MutualFundTable{
		MutualFundID:   mutualFundID,
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
