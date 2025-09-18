package model

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
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

		VoucherNumber string     `gorm:"type:varchar(255);uniqueIndex:idx_voucher_number_branch"`
		Date          time.Time  `gorm:"not null;default:now()"`
		Description   string     `gorm:"type:text"`
		Reference     string     `gorm:"type:varchar(255)"`
		Status        string     `gorm:"type:varchar(50);default:'draft'"` // draft, posted, cancelled
		PostedAt      *time.Time `gorm:"type:timestamp"`
		PostedByID    *uuid.UUID `gorm:"type:uuid"`
		PostedBy      *User      `gorm:"foreignKey:PostedByID;constraint:OnDelete:SET NULL;" json:"posted_by,omitempty"`

		// Relationships
		JournalVoucherEntries []*JournalVoucherEntry `gorm:"foreignKey:JournalVoucherID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"journal_voucher_entries,omitempty"`

		// Computed fields
		TotalDebit  float64 `gorm:"-" json:"total_debit"`
		TotalCredit float64 `gorm:"-" json:"total_credit"`
	}

	JournalVoucherResponse struct {
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
		VoucherNumber  string                `json:"voucher_number"`
		Date           string                `json:"date"`
		Description    string                `json:"description"`
		Reference      string                `json:"reference"`
		Status         string                `json:"status"`
		PostedAt       *string               `json:"posted_at,omitempty"`
		PostedByID     *uuid.UUID            `json:"posted_by_id,omitempty"`
		PostedBy       *UserResponse         `json:"posted_by,omitempty"`

		// Relationships
		JournalVoucherEntries []*JournalVoucherEntryResponse `json:"journal_voucher_entries,omitempty"`

		// Computed fields
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
	}

	JournalVoucherRequest struct {
		VoucherNumber string    `json:"voucher_number" validate:"required"`
		Date          time.Time `json:"date"`
		Description   string    `json:"description,omitempty"`
		Reference     string    `json:"reference,omitempty"`
		Status        string    `json:"status,omitempty"`

		// Nested relationships for creation/update
		JournalVoucherEntries        []*JournalVoucherEntryRequest `json:"journal_voucher_entries,omitempty"`
		JournalVoucherEntriesDeleted []uuid.UUID                   `json:"journal_voucher_entries_deleted,omitempty"`
	}
)

func (m *Model) JournalVoucher() {
	m.Migration = append(m.Migration, &JournalVoucher{})
	m.JournalVoucherManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		JournalVoucher, JournalVoucherResponse, JournalVoucherRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization", "PostedBy",
			"JournalVoucherEntries", "JournalVoucherEntries.Account",
			"JournalVoucherEntries.MemberProfile", "JournalVoucherEntries.EmployeeUser",
		},
		Service: m.provider.Service,
		Resource: func(data *JournalVoucher) *JournalVoucherResponse {
			if data == nil {
				return nil
			}

			// Calculate totals
			totalDebit := 0.0
			totalCredit := 0.0
			for _, entry := range data.JournalVoucherEntries {
				totalDebit += entry.Debit
				totalCredit += entry.Credit
			}

			var postedAt *string
			if data.PostedAt != nil {
				postedAtStr := data.PostedAt.Format(time.RFC3339)
				postedAt = &postedAtStr
			}

			return &JournalVoucherResponse{
				ID:                    data.ID,
				CreatedAt:             data.CreatedAt.Format(time.RFC3339),
				CreatedByID:           data.CreatedByID,
				CreatedBy:             m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:             data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:           data.UpdatedByID,
				UpdatedBy:             m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:        data.OrganizationID,
				Organization:          m.OrganizationManager.ToModel(data.Organization),
				BranchID:              data.BranchID,
				Branch:                m.BranchManager.ToModel(data.Branch),
				VoucherNumber:         data.VoucherNumber,
				Date:                  data.Date.Format("2006-01-02"),
				Description:           data.Description,
				Reference:             data.Reference,
				Status:                data.Status,
				PostedAt:              postedAt,
				PostedByID:            data.PostedByID,
				PostedBy:              m.UserManager.ToModel(data.PostedBy),
				JournalVoucherEntries: m.mapJournalVoucherEntries(data.JournalVoucherEntries),
				TotalDebit:            totalDebit,
				TotalCredit:           totalCredit,
			}
		},
		Created: func(data *JournalVoucher) []string {
			return []string{
				"journal_voucher.create",
				fmt.Sprintf("journal_voucher.create.%s", data.ID),
				fmt.Sprintf("journal_voucher.create.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *JournalVoucher) []string {
			return []string{
				"journal_voucher.update",
				fmt.Sprintf("journal_voucher.update.%s", data.ID),
				fmt.Sprintf("journal_voucher.update.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *JournalVoucher) []string {
			return []string{
				"journal_voucher.delete",
				fmt.Sprintf("journal_voucher.delete.%s", data.ID),
				fmt.Sprintf("journal_voucher.delete.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) JournalVoucherCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*JournalVoucher, error) {
	return m.JournalVoucherManager.Find(context, &JournalVoucher{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}

// Helper function to map journal voucher entries
func (m *Model) mapJournalVoucherEntries(entries []*JournalVoucherEntry) []*JournalVoucherEntryResponse {
	if entries == nil {
		return nil
	}

	var result []*JournalVoucherEntryResponse
	for _, entry := range entries {
		if entry != nil {
			result = append(result, m.JournalVoucherEntryManager.ToModel(entry))
		}
	}
	return result
}

// Helper function to validate journal voucher balance
func (m *Model) ValidateJournalVoucherBalance(entries []*JournalVoucherEntry) error {
	totalDebit := 0.0
	totalCredit := 0.0

	for _, entry := range entries {
		totalDebit += entry.Debit
		totalCredit += entry.Credit
	}

	if totalDebit != totalCredit {
		return fmt.Errorf("journal voucher is not balanced: debit %.2f != credit %.2f", totalDebit, totalCredit)
	}

	return nil
}
