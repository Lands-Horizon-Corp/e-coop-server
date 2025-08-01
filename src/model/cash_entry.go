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
	CashEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_entry" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_entry" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID                 *uuid.UUID               `gorm:"type:uuid" json:"account_id"`
		Account                   *Account                 `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`
		MemberProfileID           *uuid.UUID               `gorm:"type:uuid" json:"member_profile_id"`
		MemberProfile             *MemberProfile           `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
		MemberJointAccountID      *uuid.UUID               `gorm:"type:uuid" json:"member_joint_account_id"`
		MemberJointAccount        *MemberJointAccount      `gorm:"foreignKey:MemberJointAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_joint_account,omitempty"`
		TransactionBatchID        *uuid.UUID               `gorm:"type:uuid" json:"transaction_batch_id"`
		TransactionBatch          *TransactionBatch        `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`
		GeneralAccountingID       *uuid.UUID               `gorm:"type:uuid;column:general_ledger_id" json:"general_ledger_id"`
		GeneralLedger             *GeneralLedger           `gorm:"foreignKey:GeneralAccountingID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"general_ledger,omitempty"`
		TransactionID             *uuid.UUID               `gorm:"type:uuid" json:"transaction_id"`
		Transaction               *Transaction             `gorm:"foreignKey:TransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction,omitempty"`
		EmployeeUserID            *uuid.UUID               `gorm:"type:uuid" json:"employee_user_id"`
		EmployeeUser              *User                    `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		DisbursementTransactionID *uuid.UUID               `gorm:"type:uuid;column:disbursement_transaction_id" json:"disbursement_transaction_id"`
		DisbursementTransaction   *DisbursementTransaction `gorm:"foreignKey:DisbursementTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"disbursement_transaction,omitempty"`

		ReferenceNumber string  `gorm:"type:varchar(255);not null" json:"reference_number"`
		Debit           float64 `gorm:"type:decimal" json:"debit"`
		Credit          float64 `gorm:"type:decimal" json:"credit"`
	}

	CashEntryResponse struct {
		ID                        uuid.UUID                        `json:"id"`
		CreatedAt                 string                           `json:"created_at"`
		CreatedByID               uuid.UUID                        `json:"created_by_id"`
		CreatedBy                 *UserResponse                    `json:"created_by,omitempty"`
		UpdatedAt                 string                           `json:"updated_at"`
		UpdatedByID               uuid.UUID                        `json:"updated_by_id"`
		UpdatedBy                 *UserResponse                    `json:"updated_by,omitempty"`
		OrganizationID            uuid.UUID                        `json:"organization_id"`
		Organization              *OrganizationResponse            `json:"organization,omitempty"`
		BranchID                  uuid.UUID                        `json:"branch_id"`
		Branch                    *BranchResponse                  `json:"branch,omitempty"`
		AccountID                 *uuid.UUID                       `json:"account_id,omitempty"`
		Account                   *AccountResponse                 `json:"account,omitempty"`
		MemberProfileID           *uuid.UUID                       `json:"member_profile_id,omitempty"`
		MemberProfile             *MemberProfileResponse           `json:"member_profile,omitempty"`
		MemberJointAccountID      *uuid.UUID                       `json:"member_joint_account_id,omitempty"`
		MemberJointAccount        *MemberJointAccountResponse      `json:"member_joint_account,omitempty"`
		TransactionBatchID        *uuid.UUID                       `json:"transaction_batch_id,omitempty"`
		TransactionBatch          *TransactionBatchResponse        `json:"transaction_batch,omitempty"`
		GeneralLedgerID           *uuid.UUID                       `json:"general_ledger_id,omitempty"`
		GeneralLedger             *GeneralLedgerResponse           `json:"general_ledger,omitempty"`
		TransactionID             *uuid.UUID                       `json:"transaction_id,omitempty"`
		Transaction               *TransactionResponse             `json:"transaction,omitempty"`
		EmployeeUserID            *uuid.UUID                       `json:"employee_user_id,omitempty"`
		EmployeeUser              *UserResponse                    `json:"employee_user,omitempty"`
		DisbursementTransactionID *uuid.UUID                       `json:"disbursement_transaction_id,omitempty"`
		DisbursementTransaction   *DisbursementTransactionResponse `json:"disbursement_transaction,omitempty"`
		ReferenceNumber           string                           `json:"reference_number"`
		Debit                     float64                          `json:"debit"`
		Credit                    float64                          `json:"credit"`
	}

	CashEntryRequest struct {
		AccountID                 *uuid.UUID `json:"account_id,omitempty"`
		MemberProfileID           *uuid.UUID `json:"member_profile_id,omitempty"`
		MemberJointAccountID      *uuid.UUID `json:"member_joint_account_id,omitempty"`
		TransactionBatchID        *uuid.UUID `json:"transaction_batch_id,omitempty"`
		GeneralLedgerID           *uuid.UUID `json:"general_ledger_id,omitempty"`
		TransactionID             *uuid.UUID `json:"transaction_id,omitempty"`
		EmployeeUserID            *uuid.UUID `json:"employee_user_id,omitempty"`
		DisbursementTransactionID *uuid.UUID `json:"disbursement_transaction_id,omitempty"`
		ReferenceNumber           string     `json:"reference_number" validate:"required,min=1,max=255"`
		Debit                     float64    `json:"debit"`
		Credit                    float64    `json:"credit"`
	}
)

func (m *Model) CashEntry() {
	m.Migration = append(m.Migration, &CashEntry{})
	m.CashEntryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[CashEntry, CashEntryResponse, CashEntryRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization", "Account", "MemberProfile",
			"MemberJointAccount", "TransactionBatch", "GeneralLedger", "Transaction", "EmployeeUser", "DisbursementTransaction"},
		Service: m.provider.Service,
		Resource: func(data *CashEntry) *CashEntryResponse {
			if data == nil {
				return nil
			}
			return &CashEntryResponse{
				ID:                        data.ID,
				CreatedAt:                 data.CreatedAt.Format(time.RFC3339),
				CreatedByID:               data.CreatedByID,
				CreatedBy:                 m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                 data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:               data.UpdatedByID,
				UpdatedBy:                 m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:            data.OrganizationID,
				Organization:              m.OrganizationManager.ToModel(data.Organization),
				BranchID:                  data.BranchID,
				Branch:                    m.BranchManager.ToModel(data.Branch),
				AccountID:                 data.AccountID,
				Account:                   m.AccountManager.ToModel(data.Account),
				MemberProfileID:           data.MemberProfileID,
				MemberProfile:             m.MemberProfileManager.ToModel(data.MemberProfile),
				MemberJointAccountID:      data.MemberJointAccountID,
				MemberJointAccount:        m.MemberJointAccountManager.ToModel(data.MemberJointAccount),
				TransactionBatchID:        data.TransactionBatchID,
				TransactionBatch:          m.TransactionBatchManager.ToModel(data.TransactionBatch),
				GeneralLedgerID:           data.GeneralAccountingID,
				GeneralLedger:             m.GeneralLedgerManager.ToModel(data.GeneralLedger),
				TransactionID:             data.TransactionID,
				Transaction:               m.TransactionManager.ToModel(data.Transaction),
				EmployeeUserID:            data.EmployeeUserID,
				EmployeeUser:              m.UserManager.ToModel(data.EmployeeUser),
				DisbursementTransactionID: data.DisbursementTransactionID,
				DisbursementTransaction:   m.DisbursementTransactionManager.ToModel(data.DisbursementTransaction),
				ReferenceNumber:           data.ReferenceNumber,
				Debit:                     data.Debit,
				Credit:                    data.Credit,
			}
		},
		Created: func(data *CashEntry) []string {
			return []string{
				"cash_entry.create",
				fmt.Sprintf("cash_entry.create.%s", data.ID),
				fmt.Sprintf("cash_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("cash_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *CashEntry) []string {
			return []string{
				"cash_entry.update",
				fmt.Sprintf("cash_entry.update.%s", data.ID),
				fmt.Sprintf("cash_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("cash_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *CashEntry) []string {
			return []string{
				"cash_entry.delete",
				fmt.Sprintf("cash_entry.delete.%s", data.ID),
				fmt.Sprintf("cash_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("cash_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) CashEntryByBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*CashEntry, error) {
	return m.CashEntryManager.Find(context, &CashEntry{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}

func (m *Model) CashEntryByReference(context context.Context, refNum string) (*CashEntry, error) {
	return m.CashEntryManager.FindOne(context, &CashEntry{
		ReferenceNumber: refNum,
	})
}
