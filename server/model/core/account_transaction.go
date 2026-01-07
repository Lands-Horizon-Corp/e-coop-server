package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountTransactionSource string

const (
	AccountTransactionSourceDailyCollectionBook       AccountTransactionSource = "daily_collection_book"
	AccountTransactionSourceCashCheckDisbursementBook AccountTransactionSource = "cash_check_disbursement_book"
	AccountTransactionSourceGeneralJournal            AccountTransactionSource = "general_journal"
)

type (
	AccountTransaction struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_org_branch_account_tx" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_org_branch_account_tx" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Source AccountTransactionSource `gorm:"type:varchar(50);not null" json:"source"`

		JVNumber    string    `gorm:"type:varchar(255);not null" json:"jv_number"`
		Date        time.Time `gorm:"type:date;not null" json:"date"`
		Description string    `gorm:"type:text" json:"description"`
		Debit       float64   `gorm:"type:numeric(18,2);default:0" json:"debit"`
		Credit      float64   `gorm:"type:numeric(18,2);default:0" json:"credit"`

		Entries []*AccountTransactionEntry `gorm:"foreignKey:AccountTransactionID" json:"entries,omitempty"`
	}

	AccountTransactionResponse struct {
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

		JVNumber    string                   `json:"jv_number"`
		Date        string                   `json:"date"`
		Description string                   `json:"description"`
		Debit       float64                  `json:"debit"`
		Credit      float64                  `json:"credit"`
		Source      AccountTransactionSource `json:"source"`

		Entries []*AccountTransactionEntryResponse `json:"entries"`
	}

	AccountTransactionRequest struct {
		JVNumber    string `json:"jv_number" validate:"required,min=1,max=255"`
		Description string `json:"description,omitempty"`
	}

	AccountTransactionProcessGLRequest struct {
		StartDate time.Time `json:"start_date" validate:"required"`
		EndDate   time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
	}

	AccountTransactionLedgerResponse struct {
		AccountTransactionEntry []*AccountTransactionEntryResponse `json:"account_transaction_entry"`
		Month                   int                                `json:"month"`
		Debit                   float64                            `json:"debit"`
		Credit                  float64                            `json:"credit"`
	}
)

func (m *Core) AccountTransactionManager() *registry.Registry[
	AccountTransaction,
	AccountTransactionResponse,
	AccountTransactionRequest,
] {
	return registry.GetRegistry(registry.RegistryParams[
		AccountTransaction,
		AccountTransactionResponse,
		AccountTransactionRequest,
	]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Entries",
			"Entries.Account",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *AccountTransaction) *AccountTransactionResponse {
			if data == nil {
				return nil
			}

			return &AccountTransactionResponse{
				ID:          data.ID,
				CreatedAt:   data.CreatedAt.Format(time.RFC3339),
				JVNumber:    data.JVNumber,
				Date:        data.Date.Format("2006-01-02"),
				Description: data.Description,
				Entries:     m.AccountTransactionEntryManager().ToModels(data.Entries),
				Source:      data.Source,
			}
		},
		Created: func(data *AccountTransaction) registry.Topics {
			return []string{
				"account_transaction.create",
				fmt.Sprintf("account_transaction.create.%s", data.ID),
				fmt.Sprintf("account_transaction.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_transaction.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *AccountTransaction) registry.Topics {
			return []string{
				"account_transaction.update",
				fmt.Sprintf("account_transaction.update.%s", data.ID),
				fmt.Sprintf("account_transaction.update.branch.%s", data.BranchID),
				fmt.Sprintf("account_transaction.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *AccountTransaction) registry.Topics {
			return []string{
				"account_transaction.delete",
				fmt.Sprintf("account_transaction.delete.%s", data.ID),
				fmt.Sprintf("account_transaction.delete.branch.%s", data.BranchID),
				fmt.Sprintf("account_transaction.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) AccountTransactionByMonthYear(
	ctx context.Context,
	year int,
	month int,
	organizationID uuid.UUID,
	branchID uuid.UUID,
) ([]*AccountTransaction, error) {
	normalizedMonth := ((month - 1) % 12) + 1
	if normalizedMonth <= 0 {
		normalizedMonth += 12
	}
	start := time.Date(year, time.Month(normalizedMonth), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "created_at", Op: query.ModeGTE, Value: start},
		{Field: "created_at", Op: query.ModeLTE, Value: end},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "created_at", Order: query.SortOrderAsc},
	}
	return m.AccountTransactionManager().ArrFind(ctx, filters, sorts, "Entries", "Entries.Account")
}
