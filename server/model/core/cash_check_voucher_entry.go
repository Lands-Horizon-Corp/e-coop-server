package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	CashCheckVoucherEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_check_voucher_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_check_voucher_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID          uuid.UUID         `gorm:"type:uuid;not null"`
		Account            *Account          `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`
		EmployeeUserID     *uuid.UUID        `gorm:"type:uuid"`
		EmployeeUser       *User             `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		TransactionBatchID *uuid.UUID        `gorm:"type:uuid"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`
		CashCheckVoucherID uuid.UUID         `gorm:"type:uuid;not null"`
		CashCheckVoucher   *CashCheckVoucher `gorm:"foreignKey:CashCheckVoucherID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"cash_check_voucher,omitempty"`

		LoanTransactionID *uuid.UUID       `gorm:"type:uuid" json:"loan_transaction_id,omitempty"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid" json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		Debit       float64 `gorm:"type:decimal"`
		Credit      float64 `gorm:"type:decimal"`
		Description string  `gorm:"type:text"`
	}

	CashCheckVoucherEntryResponse struct {
		ID                 uuid.UUID                 `json:"id"`
		CreatedAt          string                    `json:"created_at"`
		CreatedByID        uuid.UUID                 `json:"created_by_id"`
		CreatedBy          *UserResponse             `json:"created_by,omitempty"`
		UpdatedAt          string                    `json:"updated_at"`
		UpdatedByID        uuid.UUID                 `json:"updated_by_id"`
		UpdatedBy          *UserResponse             `json:"updated_by,omitempty"`
		OrganizationID     uuid.UUID                 `json:"organization_id"`
		Organization       *OrganizationResponse     `json:"organization,omitempty"`
		BranchID           uuid.UUID                 `json:"branch_id"`
		Branch             *BranchResponse           `json:"branch,omitempty"`
		AccountID          uuid.UUID                 `json:"account_id"`
		Account            *AccountResponse          `json:"account,omitempty"`
		EmployeeUserID     *uuid.UUID                `json:"employee_user_id,omitempty"`
		EmployeeUser       *UserResponse             `json:"employee_user,omitempty"`
		TransactionBatchID *uuid.UUID                `json:"transaction_batch_id,omitempty"`
		TransactionBatch   *TransactionBatchResponse `json:"transaction_batch,omitempty"`
		CashCheckVoucherID uuid.UUID                 `json:"cash_check_voucher_id"`
		CashCheckVoucher   *CashCheckVoucherResponse `json:"cash_check_voucher,omitempty"`

		LoanTransactionID *uuid.UUID               `json:"loan_transaction_id,omitempty"`
		LoanTransaction   *LoanTransactionResponse `json:"loan_transaction,omitempty"`

		MemberProfileID *uuid.UUID             `json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
		Debit           float64                `json:"debit"`
		Credit          float64                `json:"credit"`
		Description     string                 `json:"description"`
	}

	CashCheckVoucherEntryRequest struct {
		ID                     *uuid.UUID `json:"id,omitempty"`
		AccountID              uuid.UUID  `json:"account_id" validate:"required"`
		CashCheckVoucherNumber string     `json:"cash_check_voucher_number,omitempty"`
		MemberProfileID        *uuid.UUID `json:"member_profile_id,omitempty"`
		LoanTransactionID      *uuid.UUID `json:"loan_transaction_id,omitempty"`

		Debit       float64 `json:"debit,omitempty"`
		Credit      float64 `json:"credit,omitempty"`
		Description string  `json:"description,omitempty"`
	}
)

func (m *Core) cashCheckVoucherEntry() {
	m.Migration = append(m.Migration, &CashCheckVoucherEntry{})
	m.CashCheckVoucherEntryManager().= registry.NewRegistry(registry.RegistryParams[
		CashCheckVoucherEntry, CashCheckVoucherEntryResponse, CashCheckVoucherEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Account.Currency",
			"Account", "EmployeeUser", "TransactionBatch", "CashCheckVoucher",
			"MemberProfile", "MemberProfile.Media", "LoanTransaction",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *CashCheckVoucherEntry) *CashCheckVoucherEntryResponse {
			if data == nil {
				return nil
			}
			return &CashCheckVoucherEntryResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       m.OrganizationManager().ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             m.BranchManager().ToModel(data.Branch),
				AccountID:          data.AccountID,
				Account:            m.AccountManager().ToModel(data.Account),
				EmployeeUserID:     data.EmployeeUserID,
				EmployeeUser:       m.UserManager().ToModel(data.EmployeeUser),
				TransactionBatchID: data.TransactionBatchID,
				TransactionBatch:   m.TransactionBatchManager().ToModel(data.TransactionBatch),
				CashCheckVoucherID: data.CashCheckVoucherID,
				CashCheckVoucher:   m.CashCheckVoucherManager().ToModel(data.CashCheckVoucher),
				LoanTransactionID:  data.LoanTransactionID,
				LoanTransaction:    m.LoanTransactionManager().ToModel(data.LoanTransaction),
				MemberProfileID:    data.MemberProfileID,
				MemberProfile:      m.MemberProfileManager().ToModel(data.MemberProfile),
				Debit:              data.Debit,
				Credit:             data.Credit,
				Description:        data.Description,
			}
		},
		Created: func(data *CashCheckVoucherEntry) registry.Topics {
			return []string{
				"cash_check_voucher_entry.create",
				fmt.Sprintf("cash_check_voucher_entry.create.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *CashCheckVoucherEntry) registry.Topics {
			return []string{
				"cash_check_voucher_entry.update",
				fmt.Sprintf("cash_check_voucher_entry.update.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *CashCheckVoucherEntry) registry.Topics {
			return []string{
				"cash_check_voucher_entry.delete",
				fmt.Sprintf("cash_check_voucher_entry.delete.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) CashCheckVoucherEntryCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*CashCheckVoucherEntry, error) {
	return m.CashCheckVoucherEntryManager().Find(context, &CashCheckVoucherEntry{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
