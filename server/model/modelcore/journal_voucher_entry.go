package modelcore

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
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

		CashCheckVoucherNumber string `gorm:"type:varchar(255)"`

		Description string  `gorm:"type:text"`
		Debit       float64 `gorm:"type:decimal"`
		Credit      float64 `gorm:"type:decimal"`
	}

	JournalVoucherEntryResponse struct {
		ID                     uuid.UUID               `json:"id"`
		CreatedAt              string                  `json:"created_at"`
		CreatedByID            uuid.UUID               `json:"created_by_id"`
		CreatedBy              *UserResponse           `json:"created_by,omitempty"`
		UpdatedAt              string                  `json:"updated_at"`
		UpdatedByID            uuid.UUID               `json:"updated_by_id"`
		UpdatedBy              *UserResponse           `json:"updated_by,omitempty"`
		OrganizationID         uuid.UUID               `json:"organization_id"`
		Organization           *OrganizationResponse   `json:"organization,omitempty"`
		BranchID               uuid.UUID               `json:"branch_id"`
		Branch                 *BranchResponse         `json:"branch,omitempty"`
		AccountID              uuid.UUID               `json:"account_id"`
		Account                *AccountResponse        `json:"account,omitempty"`
		MemberProfileID        *uuid.UUID              `json:"member_profile_id,omitempty"`
		MemberProfile          *MemberProfileResponse  `json:"member_profile,omitempty"`
		EmployeeUserID         *uuid.UUID              `json:"employee_user_id,omitempty"`
		EmployeeUser           *UserResponse           `json:"employee_user,omitempty"`
		JournalVoucherID       uuid.UUID               `json:"journal_voucher_id"`
		JournalVoucher         *JournalVoucherResponse `json:"journal_voucher,omitempty"`
		CashCheckVoucherNumber string                  `json:"cash_check_voucher_number"`
		Description            string                  `json:"description"`
		Debit                  float64                 `json:"debit"`
		Credit                 float64                 `json:"credit"`
	}

	JournalVoucherEntryRequest struct {
		ID                     *uuid.UUID `json:"id,omitempty"`
		AccountID              uuid.UUID  `json:"account_id" validate:"required"`
		MemberProfileID        *uuid.UUID `json:"member_profile_id,omitempty"`
		EmployeeUserID         *uuid.UUID `json:"employee_user_id,omitempty"`
		CashCheckVoucherNumber string     `json:"cash_check_voucher_number,omitempty"`
		Description            string     `json:"description,omitempty"`
		Debit                  float64    `json:"debit,omitempty"`
		Credit                 float64    `json:"credit,omitempty"`
	}
)

func (m *ModelCore) journalVoucherEntry() {
	m.migration = append(m.migration, &JournalVoucherEntry{})
	m.journalVoucherEntryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		JournalVoucherEntry, JournalVoucherEntryResponse, JournalVoucherEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy",
			"Account", "MemberProfile", "EmployeeUser", "JournalVoucher",
			"Account.Currency",
		},
		Service: m.provider.Service,
		Resource: func(data *JournalVoucherEntry) *JournalVoucherEntryResponse {
			if data == nil {
				return nil
			}
			return &JournalVoucherEntryResponse{
				ID:                     data.ID,
				CreatedAt:              data.CreatedAt.Format(time.RFC3339),
				CreatedByID:            data.CreatedByID,
				CreatedBy:              m.userManager.ToModel(data.CreatedBy),
				UpdatedAt:              data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:            data.UpdatedByID,
				UpdatedBy:              m.userManager.ToModel(data.UpdatedBy),
				OrganizationID:         data.OrganizationID,
				Organization:           m.organizationManager.ToModel(data.Organization),
				BranchID:               data.BranchID,
				Branch:                 m.branchManager.ToModel(data.Branch),
				AccountID:              data.AccountID,
				Account:                m.accountManager.ToModel(data.Account),
				MemberProfileID:        data.MemberProfileID,
				MemberProfile:          m.memberProfileManager.ToModel(data.MemberProfile),
				EmployeeUserID:         data.EmployeeUserID,
				EmployeeUser:           m.userManager.ToModel(data.EmployeeUser),
				JournalVoucherID:       data.JournalVoucherID,
				JournalVoucher:         m.journalVoucherManager.ToModel(data.JournalVoucher),
				CashCheckVoucherNumber: data.CashCheckVoucherNumber,
				Description:            data.Description,
				Debit:                  data.Debit,
				Credit:                 data.Credit,
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

func (m *ModelCore) journalVoucherEntryCurrentbranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*JournalVoucherEntry, error) {
	return m.journalVoucherEntryManager.Find(context, &JournalVoucherEntry{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
