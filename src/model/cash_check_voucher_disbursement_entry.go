package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	CashCheckVoucherDisbursementEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_check_voucher_disbursement_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_check_voucher_disbursement_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		PrintedByUserID  *uuid.UUID `gorm:"type:uuid"`
		PrintedByUser    *User      `gorm:"foreignKey:PrintedByUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"printed_by_user,omitempty"`
		ApprovedByUserID *uuid.UUID `gorm:"type:uuid"`
		ApprovedByUser   *User      `gorm:"foreignKey:ApprovedByUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"approved_by_user,omitempty"`
		ReleasedByUserID *uuid.UUID `gorm:"type:uuid"`
		ReleasedByUser   *User      `gorm:"foreignKey:ReleasedByUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"released_by_user,omitempty"`

		EmployeeUserID     *uuid.UUID        `gorm:"type:uuid"`
		EmployeeUser       *User             `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		AccountID          *uuid.UUID        `gorm:"type:uuid"`
		Account            *Account          `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`
		CashCheckVoucherID *uuid.UUID        `gorm:"type:uuid"`
		CashCheckVoucher   *CashCheckVoucher `gorm:"foreignKey:CashCheckVoucherID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"cash_check_voucher,omitempty"`
		TransactionBatchID *uuid.UUID        `gorm:"type:uuid"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`

		PayTo                  string                 `gorm:"type:varchar(255)"`
		CashCheckVoucherNumber string                 `gorm:"type:varchar(255)"`
		Status                 CashCheckVoucherStatus `gorm:"type:varchar(20)"` // Enum as string
		Description            string                 `gorm:"type:text"`
		Amount                 float64                `gorm:"type:decimal"`

		PrintedDate  *time.Time `gorm:"type:timestamp"`
		ApprovedDate *time.Time `gorm:"type:timestamp"`
		ReleasedDate *time.Time `gorm:"type:timestamp"`
	}

	CashCheckVoucherDisbursementEntryResponse struct {
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

		PrintedByUserID  *uuid.UUID    `json:"printed_by_user_id,omitempty"`
		PrintedByUser    *UserResponse `json:"printed_by_user,omitempty"`
		ApprovedByUserID *uuid.UUID    `json:"approved_by_user_id,omitempty"`
		ApprovedByUser   *UserResponse `json:"approved_by_user,omitempty"`
		ReleasedByUserID *uuid.UUID    `json:"released_by_user_id,omitempty"`
		ReleasedByUser   *UserResponse `json:"released_by_user,omitempty"`

		EmployeeUserID     *uuid.UUID                `json:"employee_user_id,omitempty"`
		EmployeeUser       *UserResponse             `json:"employee_user,omitempty"`
		AccountID          *uuid.UUID                `json:"account_id,omitempty"`
		Account            *AccountResponse          `json:"account,omitempty"`
		CashCheckVoucherID *uuid.UUID                `json:"cash_check_voucher_id,omitempty"`
		CashCheckVoucher   *CashCheckVoucherResponse `json:"cash_check_voucher,omitempty"`
		TransactionBatchID *uuid.UUID                `json:"transaction_batch_id,omitempty"`
		TransactionBatch   *TransactionBatchResponse `json:"transaction_batch,omitempty"`

		PayTo                  string                 `json:"pay_to"`
		CashCheckVoucherNumber string                 `json:"cash_check_voucher_number"`
		Status                 CashCheckVoucherStatus `json:"status"`
		Description            string                 `json:"description"`
		Amount                 float64                `json:"amount"`
		PrintedDate            *string                `json:"printed_date,omitempty"`
		ApprovedDate           *string                `json:"approved_date,omitempty"`
		ReleasedDate           *string                `json:"released_date,omitempty"`
	}

	CashCheckVoucherDisbursementEntryRequest struct {
		PrintedByUserID        *uuid.UUID             `json:"printed_by_user_id,omitempty"`
		ApprovedByUserID       *uuid.UUID             `json:"approved_by_user_id,omitempty"`
		ReleasedByUserID       *uuid.UUID             `json:"released_by_user_id,omitempty"`
		EmployeeUserID         *uuid.UUID             `json:"employee_user_id,omitempty"`
		AccountID              *uuid.UUID             `json:"account_id,omitempty"`
		CashCheckVoucherID     *uuid.UUID             `json:"cash_check_voucher_id,omitempty"`
		TransactionBatchID     *uuid.UUID             `json:"transaction_batch_id,omitempty"`
		PayTo                  string                 `json:"pay_to,omitempty"`
		CashCheckVoucherNumber string                 `json:"cash_check_voucher_number,omitempty"`
		Status                 CashCheckVoucherStatus `json:"status,omitempty"`
		Description            string                 `json:"description,omitempty"`
		Amount                 float64                `json:"amount,omitempty"`
		PrintedDate            *time.Time             `json:"printed_date,omitempty"`
		ApprovedDate           *time.Time             `json:"approved_date,omitempty"`
		ReleasedDate           *time.Time             `json:"released_date,omitempty"`
	}
)

func (m *Model) CashCheckVoucherDisbursementEntry() {
	m.Migration = append(m.Migration, &CashCheckVoucherDisbursementEntry{})
	m.CashCheckVoucherDisbursementEntryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		CashCheckVoucherDisbursementEntry, CashCheckVoucherDisbursementEntryResponse, CashCheckVoucherDisbursementEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
			"PrintedByUser", "ApprovedByUser", "ReleasedByUser",
			"EmployeeUser", "Account", "CashCheckVoucher", "TransactionBatch",
		},
		Service: m.provider.Service,
		Resource: func(data *CashCheckVoucherDisbursementEntry) *CashCheckVoucherDisbursementEntryResponse {
			if data == nil {
				return nil
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
			return &CashCheckVoucherDisbursementEntryResponse{
				ID:                     data.ID,
				CreatedAt:              data.CreatedAt.Format(time.RFC3339),
				CreatedByID:            data.CreatedByID,
				CreatedBy:              m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:              data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:            data.UpdatedByID,
				UpdatedBy:              m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:         data.OrganizationID,
				Organization:           m.OrganizationManager.ToModel(data.Organization),
				BranchID:               data.BranchID,
				Branch:                 m.BranchManager.ToModel(data.Branch),
				PrintedByUserID:        data.PrintedByUserID,
				PrintedByUser:          m.UserManager.ToModel(data.PrintedByUser),
				ApprovedByUserID:       data.ApprovedByUserID,
				ApprovedByUser:         m.UserManager.ToModel(data.ApprovedByUser),
				ReleasedByUserID:       data.ReleasedByUserID,
				ReleasedByUser:         m.UserManager.ToModel(data.ReleasedByUser),
				EmployeeUserID:         data.EmployeeUserID,
				EmployeeUser:           m.UserManager.ToModel(data.EmployeeUser),
				AccountID:              data.AccountID,
				Account:                m.AccountManager.ToModel(data.Account),
				CashCheckVoucherID:     data.CashCheckVoucherID,
				CashCheckVoucher:       m.CashCheckVoucherManager.ToModel(data.CashCheckVoucher),
				TransactionBatchID:     data.TransactionBatchID,
				TransactionBatch:       m.TransactionBatchManager.ToModel(data.TransactionBatch),
				PayTo:                  data.PayTo,
				CashCheckVoucherNumber: data.CashCheckVoucherNumber,
				Status:                 data.Status,
				Description:            data.Description,
				Amount:                 data.Amount,
				PrintedDate:            printedDate,
				ApprovedDate:           approvedDate,
				ReleasedDate:           releasedDate,
			}
		},
		Created: func(data *CashCheckVoucherDisbursementEntry) []string {
			return []string{
				"cash_check_voucher_disbursement_entry.create",
				fmt.Sprintf("cash_check_voucher_disbursement_entry.create.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_disbursement_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_disbursement_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *CashCheckVoucherDisbursementEntry) []string {
			return []string{
				"cash_check_voucher_disbursement_entry.update",
				fmt.Sprintf("cash_check_voucher_disbursement_entry.update.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_disbursement_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_disbursement_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *CashCheckVoucherDisbursementEntry) []string {
			return []string{
				"cash_check_voucher_disbursement_entry.delete",
				fmt.Sprintf("cash_check_voucher_disbursement_entry.delete.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_disbursement_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_disbursement_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) CashCheckVoucherDisbursementEntryCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*CashCheckVoucherDisbursementEntry, error) {
	return m.CashCheckVoucherDisbursementEntryManager.Find(context, &CashCheckVoucherDisbursementEntry{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
