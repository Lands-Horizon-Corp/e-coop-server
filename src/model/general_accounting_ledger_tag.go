package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

// Enum for tag_category (customize as needed)
type TagCategory string

type (
	GeneralAccountingLedgerTag struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_accounting_ledger_tag"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_accounting_ledger_tag"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		GeneralAccountingLedgerID uuid.UUID                `gorm:"type:uuid;not null"`
		GeneralAccountingLedger   *GeneralAccountingLedger `gorm:"foreignKey:GeneralAccountingLedgerID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"general_accounting_ledger,omitempty"`

		Name        string      `gorm:"type:varchar(50)"`
		Description string      `gorm:"type:text"`
		Category    TagCategory `gorm:"type:varchar(50)"`
		Color       string      `gorm:"type:varchar(20)"`
		Icon        string      `gorm:"type:varchar(20)"`
	}

	GeneralAccountingLedgerTagResponse struct {
		ID                        uuid.UUID                        `json:"id"`
		CreatedAt                 string                           `json:"created_at"`
		CreatedByID               uuid.UUID                        `json:"created_by_id"`
		CreatedBy                 *UserResponse                    `json:"created_by,omitempty"`
		UpdatedAt                 string                           `json:"updated_at"`
		UpdatedByID               uuid.UUID                        `json:"updated_by_id"`
		UpdatedBy                 *UserResponse                    `json:"updated_by,omitempty"`
		OrganizationID            uuid.UUID                        `json:"organization_id"`
		Organization              *OrganizationResponse            `json:"organization,omitempty"`
		BranchID                  uuid.UUID                        `json:"branch_id"`
		Branch                    *BranchResponse                  `json:"branch,omitempty"`
		GeneralAccountingLedgerID uuid.UUID                        `json:"general_accounting_ledger_id"`
		GeneralAccountingLedger   *GeneralAccountingLedgerResponse `json:"general_accounting_ledger,omitempty"`
		Name                      string                           `json:"name"`
		Description               string                           `json:"description"`
		Category                  TagCategory                      `json:"category"`
		Color                     string                           `json:"color"`
		Icon                      string                           `json:"icon"`
	}

	GeneralAccountingLedgerTagRequest struct {
		GeneralAccountingLedgerID uuid.UUID   `json:"general_accounting_ledger_id" validate:"required"`
		Name                      string      `json:"name" validate:"required,min=1,max=50"`
		Description               string      `json:"description,omitempty"`
		Category                  TagCategory `json:"category,omitempty"`
		Color                     string      `json:"color,omitempty"`
		Icon                      string      `json:"icon,omitempty"`
	}
)

func (m *Model) GeneralAccountingLedgerTag() {
	m.Migration = append(m.Migration, &GeneralAccountingLedgerTag{})
	m.GeneralAccountingLedgerTagManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		GeneralAccountingLedgerTag, GeneralAccountingLedgerTagResponse, GeneralAccountingLedgerTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization", "GeneralAccountingLedger",
		},
		Service: m.provider.Service,
		Resource: func(data *GeneralAccountingLedgerTag) *GeneralAccountingLedgerTagResponse {
			if data == nil {
				return nil
			}
			return &GeneralAccountingLedgerTagResponse{
				ID:                        data.ID,
				CreatedAt:                 data.CreatedAt.Format(time.RFC3339),
				CreatedByID:               data.CreatedByID,
				CreatedBy:                 m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                 data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:               data.UpdatedByID,
				UpdatedBy:                 m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:            data.OrganizationID,
				Organization:              m.OrganizationManager.ToModel(data.Organization),
				BranchID:                  data.BranchID,
				Branch:                    m.BranchManager.ToModel(data.Branch),
				GeneralAccountingLedgerID: data.GeneralAccountingLedgerID,
				GeneralAccountingLedger:   m.GeneralAccountingLedgerManager.ToModel(data.GeneralAccountingLedger),
				Name:                      data.Name,
				Description:               data.Description,
				Category:                  data.Category,
				Color:                     data.Color,
				Icon:                      data.Icon,
			}
		},
		Created: func(data *GeneralAccountingLedgerTag) []string {
			return []string{
				"general_accounting_ledger_tag.create",
				fmt.Sprintf("general_accounting_ledger_tag.create.%s", data.ID),
				fmt.Sprintf("general_accounting_ledger_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_accounting_ledger_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *GeneralAccountingLedgerTag) []string {
			return []string{
				"general_accounting_ledger_tag.update",
				fmt.Sprintf("general_accounting_ledger_tag.update.%s", data.ID),
				fmt.Sprintf("general_accounting_ledger_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_accounting_ledger_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *GeneralAccountingLedgerTag) []string {
			return []string{
				"general_accounting_ledger_tag.delete",
				fmt.Sprintf("general_accounting_ledger_tag.delete.%s", data.ID),
				fmt.Sprintf("general_accounting_ledger_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_accounting_ledger_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}
