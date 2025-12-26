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
	JournalVoucher struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_journal_voucher"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_journal_voucher"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`
		CurrencyID     uuid.UUID     `gorm:"type:uuid;not null"`
		Currency       *Currency     `gorm:"foreignKey:CurrencyID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"currency,omitempty"`

		Name              string     `gorm:"type:varchar(255)"`
		CashVoucherNumber string     `gorm:"type:varchar(255)"`
		Date              time.Time  `gorm:"not null;default:now()"`
		Description       string     `gorm:"type:text"`
		Reference         string     `gorm:"type:varchar(255)"`
		Status            string     `gorm:"type:varchar(50);default:'draft'"` // draft, posted, cancelled
		PostedAt          *time.Time `gorm:"type:timestamp"`
		PostedByID        *uuid.UUID `gorm:"type:uuid"`
		PostedBy          *User      `gorm:"foreignKey:PostedByID;constraint:OnDelete:SET NULL;" json:"posted_by,omitempty"`

		EmployeeUserID     *uuid.UUID        `gorm:"type:uuid" json:"employee_user_id,omitempty"`
		EmployeeUser       *User             `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		TransactionBatchID *uuid.UUID        `gorm:"type:uuid" json:"transaction_batch_id,omitempty"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`

		PrintedDate  *time.Time `gorm:"type:timestamp"`
		PrintedByID  *uuid.UUID `gorm:"type:uuid"`
		PrintedBy    *User      `gorm:"foreignKey:PrintedByID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"printed_by,omitempty"`
		PrintNumber  int        `gorm:"type:int;default:0"`
		ApprovedDate *time.Time `gorm:"type:timestamp"`
		ApprovedByID *uuid.UUID `gorm:"type:uuid"`
		ApprovedBy   *User      `gorm:"foreignKey:ApprovedByID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"approved_by,omitempty"`
		ReleasedDate *time.Time `gorm:"type:timestamp"`
		ReleasedByID *uuid.UUID `gorm:"type:uuid"`
		ReleasedBy   *User      `gorm:"foreignKey:ReleasedByID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"released_by,omitempty"`

		JournalVoucherTags []*JournalVoucherTag `gorm:"foreignKey:JournalVoucherID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"journal_voucher_tags,omitempty"`

		JournalVoucherEntries []*JournalVoucherEntry `gorm:"foreignKey:JournalVoucherID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"journal_voucher_entries,omitempty"`

		TotalDebit  float64 `gorm:"type:decimal" json:"total_debit"`
		TotalCredit float64 `gorm:"type:decimal" json:"total_credit"`
	}

	JournalVoucherResponse struct {
		ID                uuid.UUID             `json:"id"`
		CreatedAt         string                `json:"created_at"`
		CreatedByID       uuid.UUID             `json:"created_by_id"`
		CreatedBy         *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt         string                `json:"updated_at"`
		UpdatedByID       uuid.UUID             `json:"updated_by_id"`
		UpdatedBy         *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID    uuid.UUID             `json:"organization_id"`
		Organization      *OrganizationResponse `json:"organization,omitempty"`
		BranchID          uuid.UUID             `json:"branch_id"`
		Branch            *BranchResponse       `json:"branch,omitempty"`
		CurrencyID        uuid.UUID             `json:"currency_id"`
		Currency          *CurrencyResponse     `json:"currency,omitempty"`
		Name              string                `json:"name"`
		VoucherNumber     string                `json:"voucher_number"`
		CashVoucherNumber string                `json:"cash_voucher_number"`
		Date              string                `json:"date"`
		Description       string                `json:"description"`
		Reference         string                `json:"reference"`
		Status            string                `json:"status"`
		PostedAt          *string               `json:"posted_at,omitempty"`
		PostedByID        *uuid.UUID            `json:"posted_by_id,omitempty"`
		PostedBy          *UserResponse         `json:"posted_by,omitempty"`

		EmployeeUserID     *uuid.UUID                `json:"employee_user_id,omitempty"`
		EmployeeUser       *UserResponse             `json:"employee_user,omitempty"`
		TransactionBatchID *uuid.UUID                `json:"transaction_batch_id,omitempty"`
		TransactionBatch   *TransactionBatchResponse `json:"transaction_batch,omitempty"`

		PrintedDate  *string       `json:"printed_date,omitempty"`
		PrintedByID  *uuid.UUID    `json:"printed_by_id,omitempty"`
		PrintedBy    *UserResponse `json:"printed_by,omitempty"`
		PrintNumber  int           `json:"print_number"`
		ApprovedDate *string       `json:"approved_date,omitempty"`
		ApprovedByID *uuid.UUID    `json:"approved_by_id,omitempty"`
		ApprovedBy   *UserResponse `json:"approved_by,omitempty"`
		ReleasedDate *string       `json:"released_date,omitempty"`
		ReleasedByID *uuid.UUID    `json:"released_by_id,omitempty"`
		ReleasedBy   *UserResponse `json:"released_by,omitempty"`

		JournalVoucherTags []*JournalVoucherTagResponse `json:"journal_voucher_tags,omitempty"`

		JournalVoucherEntries []*JournalVoucherEntryResponse `json:"journal_voucher_entries,omitempty"`

		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
	}

	JournalVoucherRequest struct {
		Name              string    `json:"name" validate:"required"`
		CashVoucherNumber string    `json:"cash_voucher_number,omitempty"`
		Date              time.Time `json:"date"`
		Description       string    `json:"description,omitempty"`
		Reference         string    `json:"reference,omitempty"`
		Status            string    `json:"status,omitempty"`
		CurrencyID        uuid.UUID `json:"currency_id" validate:"required"`

		JournalVoucherEntries        []*JournalVoucherEntryRequest `json:"journal_voucher_entries,omitempty"`
		JournalVoucherEntriesDeleted uuid.UUIDs                    `json:"journal_voucher_entries_deleted,omitempty"`
	}

	JournalVoucherPrintRequest struct {
		CashVoucherNumber string `json:"cash_voucher_number,omitempty"`
	}
)

func (m *Core) JournalVoucherManager() *registry.Registry[JournalVoucher, JournalVoucherResponse, JournalVoucherRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		JournalVoucher, JournalVoucherResponse, JournalVoucherRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Currency", "PostedBy",
			"EmployeeUser", "EmployeeUser.Media", "TransactionBatch",
			"PrintedBy", "ApprovedBy", "ReleasedBy",
			"PrintedBy.Media", "ApprovedBy.Media", "ReleasedBy.Media",
			"JournalVoucherTags",
			"JournalVoucherEntries", "JournalVoucherEntries.Account", "JournalVoucherEntries.LoanTransaction",
			"JournalVoucherEntries.Account.Currency",
			"JournalVoucherEntries.MemberProfile", "JournalVoucherEntries.EmployeeUser",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *JournalVoucher) *JournalVoucherResponse {
			if data == nil {
				return nil
			}

			var postedAt *string
			if data.PostedAt != nil {
				postedAtStr := data.PostedAt.Format(time.RFC3339)
				postedAt = &postedAtStr
			}

			var printedDate, approvedDate, releasedDate *string
			if data.PrintedDate != nil {
				str := data.PrintedDate.Format(time.RFC3339)
				printedDate = &str
			}
			if data.ApprovedDate != nil {
				str := data.ApprovedDate.Format(time.RFC3339)
				approvedDate = &str
			}
			if data.ReleasedDate != nil {
				str := data.ReleasedDate.Format(time.RFC3339)
				releasedDate = &str
			}

			return &JournalVoucherResponse{
				ID:                    data.ID,
				CreatedAt:             data.CreatedAt.Format(time.RFC3339),
				CreatedByID:           data.CreatedByID,
				CreatedBy:             m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:             data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:           data.UpdatedByID,
				UpdatedBy:             m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID:        data.OrganizationID,
				Organization:          m.OrganizationManager().ToModel(data.Organization),
				BranchID:              data.BranchID,
				Branch:                m.BranchManager().ToModel(data.Branch),
				CurrencyID:            data.CurrencyID,
				Currency:              m.CurrencyManager().ToModel(data.Currency),
				Name:                  data.Name,
				CashVoucherNumber:     data.CashVoucherNumber,
				Date:                  data.Date.Format("2006-01-02"),
				Description:           data.Description,
				Reference:             data.Reference,
				Status:                data.Status,
				PostedAt:              postedAt,
				PostedByID:            data.PostedByID,
				PostedBy:              m.UserManager().ToModel(data.PostedBy),
				EmployeeUserID:        data.EmployeeUserID,
				EmployeeUser:          m.UserManager().ToModel(data.EmployeeUser),
				TransactionBatchID:    data.TransactionBatchID,
				TransactionBatch:      m.TransactionBatchManager().ToModel(data.TransactionBatch),
				PrintedDate:           printedDate,
				PrintedByID:           data.PrintedByID,
				PrintedBy:             m.UserManager().ToModel(data.PrintedBy),
				PrintNumber:           data.PrintNumber,
				ApprovedDate:          approvedDate,
				ApprovedByID:          data.ApprovedByID,
				ApprovedBy:            m.UserManager().ToModel(data.ApprovedBy),
				ReleasedDate:          releasedDate,
				ReleasedByID:          data.ReleasedByID,
				ReleasedBy:            m.UserManager().ToModel(data.ReleasedBy),
				JournalVoucherTags:    m.JournalVoucherTagManager().ToModels(data.JournalVoucherTags),
				JournalVoucherEntries: m.JournalVoucherEntryManager().ToModels(data.JournalVoucherEntries),
				TotalDebit:            data.TotalDebit,
				TotalCredit:           data.TotalCredit,
			}
		},
		Created: func(data *JournalVoucher) registry.Topics {
			return []string{
				"journal_voucher.create",
				fmt.Sprintf("journal_voucher.create.%s", data.ID),
				fmt.Sprintf("journal_voucher.create.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *JournalVoucher) registry.Topics {
			return []string{
				"journal_voucher.update",
				fmt.Sprintf("journal_voucher.update.%s", data.ID),
				fmt.Sprintf("journal_voucher.update.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *JournalVoucher) registry.Topics {
			return []string{
				"journal_voucher.delete",
				fmt.Sprintf("journal_voucher.delete.%s", data.ID),
				fmt.Sprintf("journal_voucher.delete.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) JournalVoucherCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*JournalVoucher, error) {
	return m.JournalVoucherManager().Find(context, &JournalVoucher{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func (m *Core) ValidateJournalVoucherBalance(entries []*JournalVoucherEntry) error {
	totalDebit := 0.0
	totalCredit := 0.0

	for _, entry := range entries {
		totalDebit = m.provider.Service.Decimal.Add(totalDebit, entry.Debit)
		totalCredit = m.provider.Service.Decimal.Add(totalCredit, entry.Credit)
	}

	if !m.provider.Service.Decimal.IsEqual(totalDebit, totalCredit) {
		return eris.Errorf("journal voucher is not balanced: debit %.2f != credit %.2f", totalDebit, totalCredit)
	}

	return nil
}

func (m *Core) JournalVoucherDraft(ctx context.Context, branchID, organizationID uuid.UUID) ([]*JournalVoucher, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "approved_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "printed_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	return m.JournalVoucherManager().ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func (m *Core) JournalVoucherPrinted(ctx context.Context, branchID, organizationID uuid.UUID) ([]*JournalVoucher, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	return m.JournalVoucherManager().ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func (m *Core) JournalVoucherApproved(ctx context.Context, branchID, organizationID uuid.UUID) ([]*JournalVoucher, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	return m.JournalVoucherManager().ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func (m *Core) JournalVoucherReleased(ctx context.Context, branchID, organizationID uuid.UUID) ([]*JournalVoucher, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
	}

	return m.JournalVoucherManager().ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func (m *Core) JournalVoucherReleasedCurrentDay(ctx context.Context, branchID uuid.UUID, organizationID uuid.UUID) ([]*JournalVoucher, error) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeGTE, Value: startOfDay},
		{Field: "released_date", Op: query.ModeLT, Value: endOfDay},
	}

	return m.JournalVoucherManager().ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}
