package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// JournalVoucherEntry represents the JournalVoucherEntry model.
	JournalVoucherEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_journal_voucher_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_journal_voucher_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID        uuid.UUID       `gorm:"type:uuid;not null"`
		Account          *Account        `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`
		MemberProfileID  *uuid.UUID      `gorm:"type:uuid"`
		MemberProfile    *MemberProfile  `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
		EmployeeUserID   *uuid.UUID      `gorm:"type:uuid"`
		EmployeeUser     *User           `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		JournalVoucherID uuid.UUID       `gorm:"type:uuid;not null"`
		JournalVoucher   *JournalVoucher `gorm:"foreignKey:JournalVoucherID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"journal_voucher,omitempty"`

		LoanTransactionID *uuid.UUID       `gorm:"type:uuid"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		Description string  `gorm:"type:text"`
		Debit       float64 `gorm:"type:decimal"`
		Credit      float64 `gorm:"type:decimal"`
	}

	// JournalVoucherEntryResponse represents the response structure for journalvoucherentry data

	// JournalVoucherEntryResponse represents the response structure for JournalVoucherEntry.
	JournalVoucherEntryResponse struct {
		ID                uuid.UUID                `json:"id"`
		CreatedAt         string                   `json:"created_at"`
		CreatedByID       uuid.UUID                `json:"created_by_id"`
		CreatedBy         *UserResponse            `json:"created_by,omitempty"`
		UpdatedAt         string                   `json:"updated_at"`
		UpdatedByID       uuid.UUID                `json:"updated_by_id"`
		UpdatedBy         *UserResponse            `json:"updated_by,omitempty"`
		OrganizationID    uuid.UUID                `json:"organization_id"`
		Organization      *OrganizationResponse    `json:"organization,omitempty"`
		BranchID          uuid.UUID                `json:"branch_id"`
		Branch            *BranchResponse          `json:"branch,omitempty"`
		AccountID         uuid.UUID                `json:"account_id"`
		Account           *AccountResponse         `json:"account,omitempty"`
		MemberProfileID   *uuid.UUID               `json:"member_profile_id,omitempty"`
		MemberProfile     *MemberProfileResponse   `json:"member_profile,omitempty"`
		EmployeeUserID    *uuid.UUID               `json:"employee_user_id,omitempty"`
		EmployeeUser      *UserResponse            `json:"employee_user,omitempty"`
		JournalVoucherID  uuid.UUID                `json:"journal_voucher_id"`
		JournalVoucher    *JournalVoucherResponse  `json:"journal_voucher,omitempty"`
		Description       string                   `json:"description"`
		Debit             float64                  `json:"debit"`
		Credit            float64                  `json:"credit"`
		LoanTransactionID *uuid.UUID               `json:"loan_transaction_id,omitempty"`
		LoanTransaction   *LoanTransactionResponse `json:"loan_transaction,omitempty"`
	}

	// JournalVoucherEntryRequest represents the request structure for JournalVoucherEntry.
	JournalVoucherEntryRequest struct {
		ID                *uuid.UUID `json:"id,omitempty"`
		AccountID         uuid.UUID  `json:"account_id" validate:"required"`
		MemberProfileID   *uuid.UUID `json:"member_profile_id,omitempty"`
		EmployeeUserID    *uuid.UUID `json:"employee_user_id,omitempty"`
		Description       string     `json:"description,omitempty"`
		Debit             float64    `json:"debit,omitempty"`
		Credit            float64    `json:"credit,omitempty"`
		LoanTransactionID *uuid.UUID `json:"loan_transaction_id,omitempty"`
	}
)

func (m *Core) journalVoucherEntry() {
	m.Migration = append(m.Migration, &JournalVoucherEntry{})
	m.JournalVoucherEntryManager = *registry.NewRegistry(registry.RegistryParams[
		JournalVoucherEntry, JournalVoucherEntryResponse, JournalVoucherEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy",
			"Account", "MemberProfile", "EmployeeUser", "JournalVoucher",
			"Account.Currency", "LoanTransaction",
		},
		Service: m.provider.Service,
		Resource: func(data *JournalVoucherEntry) *JournalVoucherEntryResponse {
			if data == nil {
				return nil
			}
			return &JournalVoucherEntryResponse{
				ID:                data.ID,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				CreatedByID:       data.CreatedByID,
				CreatedBy:         m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:       data.UpdatedByID,
				UpdatedBy:         m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:    data.OrganizationID,
				Organization:      m.OrganizationManager.ToModel(data.Organization),
				BranchID:          data.BranchID,
				Branch:            m.BranchManager.ToModel(data.Branch),
				AccountID:         data.AccountID,
				Account:           m.AccountManager.ToModel(data.Account),
				MemberProfileID:   data.MemberProfileID,
				MemberProfile:     m.MemberProfileManager.ToModel(data.MemberProfile),
				EmployeeUserID:    data.EmployeeUserID,
				EmployeeUser:      m.UserManager.ToModel(data.EmployeeUser),
				JournalVoucherID:  data.JournalVoucherID,
				JournalVoucher:    m.JournalVoucherManager.ToModel(data.JournalVoucher),
				LoanTransactionID: data.LoanTransactionID,
				LoanTransaction:   m.LoanTransactionManager.ToModel(data.LoanTransaction),
				Description:       data.Description,
				Debit:             data.Debit,
				Credit:            data.Credit,
			}
		},
		Created: func(data *JournalVoucherEntry) []string {
			return []string{
				"journal_voucher_entry.create",
				fmt.Sprintf("journal_voucher_entry.create.%s", data.ID),
				fmt.Sprintf("journal_voucher_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *JournalVoucherEntry) []string {
			return []string{
				"journal_voucher_entry.update",
				fmt.Sprintf("journal_voucher_entry.update.%s", data.ID),
				fmt.Sprintf("journal_voucher_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *JournalVoucherEntry) []string {
			return []string{
				"journal_voucher_entry.delete",
				fmt.Sprintf("journal_voucher_entry.delete.%s", data.ID),
				fmt.Sprintf("journal_voucher_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// JournalVoucherEntryCurrentBranch returns JournalVoucherEntryCurrentBranch for the current branch or organization where applicable.
func (m *Core) JournalVoucherEntryCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*JournalVoucherEntry, error) {
	return m.JournalVoucherEntryManager.Find(context, &JournalVoucherEntry{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
