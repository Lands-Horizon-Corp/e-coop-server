package types

import (
	"time"

	"github.com/google/uuid"
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

		GeneralLedgerDefinitionEntries []*GeneralLedgerDefinition `gorm:"foreignKey:GeneralLedgerAccountsGroupingID" json:"general_ledger_definition,omitempty"`
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

		GeneralLedgerDefinitionEntries []*GeneralLedgerDefinitionResponse `json:"general_ledger_definition,omitempty"`
	}

	GeneralLedgerAccountsGroupingRequest struct {
		Debit       float64 `json:"debit" validate:"omitempty,gt=0"`
		Credit      float64 `json:"credit" validate:"omitempty,gt=0"`
		Name        string  `json:"name" validate:"required,min=1,max=50"`
		Description string  `json:"description" validate:"required"`
		FromCode    float64 `json:"from_code,omitempty"`
		ToCode      float64 `json:"to_code,omitempty"`
	}
)
