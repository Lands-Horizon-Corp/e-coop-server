package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	FinancialStatementAccountsGrouping struct {
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

		FinancialStatementDefinitionEntries []*FinancialStatementDefinition `gorm:"foreignKey:FinancialStatementAccountsGroupingID" json:"financial_statement_definition_entries,omitempty"`
	}

	FinancialStatementAccountsGroupingResponse struct {
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

	FinancialStatementAccountsGroupingRequest struct {
		Name        string     `json:"name" validate:"required,min=1,max=50"`
		Description string     `json:"description" validate:"required"`
		Debit       float64    `json:"debit" validate:"omitempty,gt=0"`
		Credit      float64    `json:"credit" validate:"omitempty,gt=0"`
		IconMediaID *uuid.UUID `json:"icon_media_id,omitempty"`
	}
)
