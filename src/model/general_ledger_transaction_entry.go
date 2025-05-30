package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	GeneralLedgerTransactionEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger_transaction_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger_transaction_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid;not null"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"account,omitempty"`

		GeneralLedgerTransactionEntryID *uuid.UUID                     `gorm:"type:uuid"`
		ParentEntry                     *GeneralLedgerTransactionEntry `gorm:"foreignKey:GeneralLedgerTransactionEntryID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"parent_entry,omitempty"`

		Debit  float64 `gorm:"type:decimal"`
		Credit float64 `gorm:"type:decimal"`
	}

	GeneralLedgerTransactionEntryResponse struct {
		ID                              uuid.UUID                              `json:"id"`
		CreatedAt                       string                                 `json:"created_at"`
		CreatedByID                     uuid.UUID                              `json:"created_by_id"`
		CreatedBy                       *UserResponse                          `json:"created_by,omitempty"`
		UpdatedAt                       string                                 `json:"updated_at"`
		UpdatedByID                     uuid.UUID                              `json:"updated_by_id"`
		UpdatedBy                       *UserResponse                          `json:"updated_by,omitempty"`
		OrganizationID                  uuid.UUID                              `json:"organization_id"`
		Organization                    *OrganizationResponse                  `json:"organization,omitempty"`
		BranchID                        uuid.UUID                              `json:"branch_id"`
		Branch                          *BranchResponse                        `json:"branch,omitempty"`
		AccountID                       uuid.UUID                              `json:"account_id"`
		Account                         *AccountResponse                       `json:"account,omitempty"`
		GeneralLedgerTransactionEntryID *uuid.UUID                             `json:"general_ledger_transaction_entry_id,omitempty"`
		ParentEntry                     *GeneralLedgerTransactionEntryResponse `json:"parent_entry,omitempty"`
		Debit                           float64                                `json:"debit"`
		Credit                          float64                                `json:"credit"`
	}

	GeneralLedgerTransactionEntryRequest struct {
		AccountID                       uuid.UUID  `json:"account_id" validate:"required"`
		GeneralLedgerTransactionEntryID *uuid.UUID `json:"general_ledger_transaction_entry_id,omitempty"`
		Debit                           float64    `json:"debit,omitempty"`
		Credit                          float64    `json:"credit,omitempty"`
	}
)

func (m *Model) GeneralLedgerTransactionEntry() {
	m.Migration = append(m.Migration, &GeneralLedgerTransactionEntry{})
	m.GeneralLedgerTransactionEntryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		GeneralLedgerTransactionEntry, GeneralLedgerTransactionEntryResponse, GeneralLedgerTransactionEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization", "Account", "ParentEntry",
		},
		Service: m.provider.Service,
		Resource: func(data *GeneralLedgerTransactionEntry) *GeneralLedgerTransactionEntryResponse {
			if data == nil {
				return nil
			}
			return &GeneralLedgerTransactionEntryResponse{
				ID:                              data.ID,
				CreatedAt:                       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                     data.CreatedByID,
				CreatedBy:                       m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                     data.UpdatedByID,
				UpdatedBy:                       m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:                  data.OrganizationID,
				Organization:                    m.OrganizationManager.ToModel(data.Organization),
				BranchID:                        data.BranchID,
				Branch:                          m.BranchManager.ToModel(data.Branch),
				AccountID:                       data.AccountID,
				Account:                         m.AccountManager.ToModel(data.Account),
				GeneralLedgerTransactionEntryID: data.GeneralLedgerTransactionEntryID,
				ParentEntry:                     m.GeneralLedgerTransactionEntryManager.ToModel(data.ParentEntry),
				Debit:                           data.Debit,
				Credit:                          data.Credit,
			}
		},
		Created: func(data *GeneralLedgerTransactionEntry) []string {
			return []string{
				"general_ledger_transaction_entry.create",
				fmt.Sprintf("general_ledger_transaction_entry.create.%s", data.ID),
				fmt.Sprintf("general_ledger_transaction_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_transaction_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *GeneralLedgerTransactionEntry) []string {
			return []string{
				"general_ledger_transaction_entry.update",
				fmt.Sprintf("general_ledger_transaction_entry.update.%s", data.ID),
				fmt.Sprintf("general_ledger_transaction_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_transaction_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *GeneralLedgerTransactionEntry) []string {
			return []string{
				"general_ledger_transaction_entry.delete",
				fmt.Sprintf("general_ledger_transaction_entry.delete.%s", data.ID),
				fmt.Sprintf("general_ledger_transaction_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_transaction_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}
