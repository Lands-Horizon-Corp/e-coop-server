package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	AccountTransactionEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_org_branch_account_tx_entry" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`

		BranchID uuid.UUID `gorm:"type:uuid;not null;index:idx_org_branch_account_tx_entry" json:"branch_id"`
		Branch   *Branch   `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountTransactionID uuid.UUID           `gorm:"type:uuid;not null;index" json:"account_transaction_id"`
		AccountTransaction   *AccountTransaction `gorm:"foreignKey:AccountTransactionID;constraint:OnDelete:CASCADE;" json:"account_transaction,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid;not null;index" json:"account_id"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT;" json:"account,omitempty"`

		Debit  float64 `gorm:"type:numeric(18,2);default:0" json:"debit"`
		Credit float64 `gorm:"type:numeric(18,2);default:0" json:"credit"`
	}

	AccountTransactionEntryResponse struct {
		ID             uuid.UUID        `json:"id"`
		CreatedAt      string           `json:"created_at"`
		OrganizationID uuid.UUID        `json:"organization_id"`
		BranchID       uuid.UUID        `json:"branch_id"`
		AccountID      uuid.UUID        `json:"account_id"`
		Account        *AccountResponse `json:"account,omitempty"`
		Debit          float64          `json:"debit"`
		Credit         float64          `json:"credit"`
	}
)

func (m *Core) AccountTransactionEntryManager() *registry.Registry[
	AccountTransactionEntry,
	AccountTransactionEntryResponse,
	any,
] {
	return registry.GetRegistry(registry.RegistryParams[
		AccountTransactionEntry,
		AccountTransactionEntryResponse,
		any,
	]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
			"Account",
			"AccountTransaction",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *AccountTransactionEntry) *AccountTransactionEntryResponse {
			if data == nil {
				return nil
			}

			return &AccountTransactionEntryResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				OrganizationID: data.OrganizationID,
				BranchID:       data.BranchID,
				AccountID:      data.AccountID,
				Account:        m.AccountManager().ToModel(data.Account),
				Debit:          data.Debit,
				Credit:         data.Credit,
			}
		},
		Created: func(data *AccountTransactionEntry) registry.Topics {
			return []string{
				"account_transaction_entry.create",
				fmt.Sprintf("account_transaction_entry.create.%s", data.ID),
				fmt.Sprintf("account_transaction_entry.create.transaction.%s", data.AccountTransactionID),
				fmt.Sprintf("account_transaction_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_transaction_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *AccountTransactionEntry) registry.Topics {
			return []string{
				"account_transaction_entry.update",
				fmt.Sprintf("account_transaction_entry.update.%s", data.ID),
				fmt.Sprintf("account_transaction_entry.update.transaction.%s", data.AccountTransactionID),
				fmt.Sprintf("account_transaction_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("account_transaction_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *AccountTransactionEntry) registry.Topics {
			return []string{
				"account_transaction_entry.delete",
				fmt.Sprintf("account_transaction_entry.delete.%s", data.ID),
				fmt.Sprintf("account_transaction_entry.delete.transaction.%s", data.AccountTransactionID),
				fmt.Sprintf("account_transaction_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("account_transaction_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}
