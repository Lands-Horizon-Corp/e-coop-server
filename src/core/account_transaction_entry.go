package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
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

		Date     time.Time `gorm:"type:date;not null;index:idx_date_account_tx_entry" json:"date"`
		JVNumber string    `gorm:"type:varchar(255);not null" json:"jv_number"`
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
		Date           string           `json:"date"`
		JVNumber       string           `json:"jv_number"`

		Balance float64 `json:"balance"`
	}
)

func AccountTransactionEntryManager(service *horizon.HorizonService) *registry.Registry[
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
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
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
				Account:        AccountManager(service).ToModel(data.Account),
				Debit:          data.Debit,
				Credit:         data.Credit,
				Date:           data.Date.Format("2006-01-02"),
				JVNumber:       data.JVNumber,
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

func AccountingEntryByAccountMonthYear(
	ctx context.Context,
	service *horizon.HorizonService,
	accountID uuid.UUID,
	month int,
	year int,
	organizationID uuid.UUID,
	branchID uuid.UUID,
) ([]*AccountTransactionEntry, error) {
	normalizedMonth := ((month-1)%12+12)%12 + 1
	startDate := time.Date(year, time.Month(normalizedMonth), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "date", Op: query.ModeGTE, Value: startDate},
		{Field: "date", Op: query.ModeLTE, Value: endDate},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "date", Order: query.SortOrderAsc},
	}
	return AccountTransactionEntryManager(service).ArrFind(ctx, filters, sorts, "Account", "Account.Currency")
}
