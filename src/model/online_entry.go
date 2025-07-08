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
	OnlineEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_online_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_online_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		BankID                    *uuid.UUID               `gorm:"type:uuid"`
		Bank                      *Bank                    `gorm:"foreignKey:BankID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"bank,omitempty"`
		AccountID                 *uuid.UUID               `gorm:"type:uuid"`
		Account                   *Account                 `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`
		MemberProfileID           *uuid.UUID               `gorm:"type:uuid"`
		MemberProfile             *MemberProfile           `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
		MemberJointAccountID      *uuid.UUID               `gorm:"type:uuid"`
		MemberJointAccount        *MemberJointAccount      `gorm:"foreignKey:MemberJointAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_joint_account,omitempty"`
		TransactionBatchID        *uuid.UUID               `gorm:"type:uuid"`
		TransactionBatch          *TransactionBatch        `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`
		GeneralLedgerID           *uuid.UUID               `gorm:"type:uuid"`
		GeneralLedger             *GeneralLedger           `gorm:"foreignKey:GeneralLedgerID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"general_ledger,omitempty"`
		TransactionID             *uuid.UUID               `gorm:"type:uuid"`
		Transaction               *Transaction             `gorm:"foreignKey:TransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction,omitempty"`
		EmployeeUserID            *uuid.UUID               `gorm:"type:uuid"`
		EmployeeUser              *User                    `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		DisbursementTransactionID *uuid.UUID               `gorm:"type:uuid"`
		DisbursementTransaction   *DisbursementTransaction `gorm:"foreignKey:DisbursementTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"disbursement_transaction,omitempty"`

		ReferenceNumber string     `gorm:"type:varchar(255);not null"`
		PaymentDate     *time.Time `gorm:"type:timestamp"`
		Debit           float64    `gorm:"type:decimal"`
		Credit          float64    `gorm:"type:decimal"`
	}

	OnlineEntryResponse struct {
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
		BankID                    *uuid.UUID                       `json:"bank_id,omitempty"`
		Bank                      *BankResponse                    `json:"bank,omitempty"`
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

		ReferenceNumber string  `json:"reference_number"`
		PaymentDate     *string `json:"payment_date,omitempty"`
		Debit           float64 `json:"debit"`
		Credit          float64 `json:"credit"`
	}

	OnlineEntryRequest struct {
		OrganizationID            uuid.UUID  `json:"organization_id" validate:"required"`
		BranchID                  uuid.UUID  `json:"branch_id" validate:"required"`
		BankID                    *uuid.UUID `json:"bank_id,omitempty"`
		AccountID                 *uuid.UUID `json:"account_id,omitempty"`
		MemberProfileID           *uuid.UUID `json:"member_profile_id,omitempty"`
		MemberJointAccountID      *uuid.UUID `json:"member_joint_account_id,omitempty"`
		TransactionBatchID        *uuid.UUID `json:"transaction_batch_id,omitempty"`
		GeneralLedgerID           *uuid.UUID `json:"general_ledger_id,omitempty"`
		TransactionID             *uuid.UUID `json:"transaction_id,omitempty"`
		EmployeeUserID            *uuid.UUID `json:"employee_user_id,omitempty"`
		DisbursementTransactionID *uuid.UUID `json:"disbursement_transaction_id,omitempty"`

		ReferenceNumber string     `json:"reference_number" validate:"required,min=1,max=255"`
		PaymentDate     *time.Time `json:"payment_date,omitempty"`
		Debit           float64    `json:"debit,omitempty"`
		Credit          float64    `json:"credit,omitempty"`
	}
)

func (m *Model) OnlineEntry() {
	m.Migration = append(m.Migration, &OnlineEntry{})
	m.OnlineEntryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		OnlineEntry, OnlineEntryResponse, OnlineEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
			"Bank", "Account", "MemberProfile", "MemberJointAccount",
			"TransactionBatch", "GeneralLedger", "Transaction",
			"EmployeeUser", "DisbursementTransaction",
		},
		Service: m.provider.Service,
		Resource: func(data *OnlineEntry) *OnlineEntryResponse {
			if data == nil {
				return nil
			}
			var paymentDate *string
			if data.PaymentDate != nil {
				s := data.PaymentDate.Format(time.RFC3339)
				paymentDate = &s
			}
			return &OnlineEntryResponse{
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
				BankID:                    data.BankID,
				Bank:                      m.BankManager.ToModel(data.Bank),
				AccountID:                 data.AccountID,
				Account:                   m.AccountManager.ToModel(data.Account),
				MemberProfileID:           data.MemberProfileID,
				MemberProfile:             m.MemberProfileManager.ToModel(data.MemberProfile),
				MemberJointAccountID:      data.MemberJointAccountID,
				MemberJointAccount:        m.MemberJointAccountManager.ToModel(data.MemberJointAccount),
				TransactionBatchID:        data.TransactionBatchID,
				TransactionBatch:          m.TransactionBatchManager.ToModel(data.TransactionBatch),
				GeneralLedgerID:           data.GeneralLedgerID,
				GeneralLedger:             m.GeneralLedgerManager.ToModel(data.GeneralLedger),
				TransactionID:             data.TransactionID,
				Transaction:               m.TransactionManager.ToModel(data.Transaction),
				EmployeeUserID:            data.EmployeeUserID,
				EmployeeUser:              m.UserManager.ToModel(data.EmployeeUser),
				DisbursementTransactionID: data.DisbursementTransactionID,
				DisbursementTransaction:   m.DisbursementTransactionManager.ToModel(data.DisbursementTransaction),
				ReferenceNumber:           data.ReferenceNumber,
				PaymentDate:               paymentDate,
				Debit:                     data.Debit,
				Credit:                    data.Credit,
			}
		},

		Created: func(data *OnlineEntry) []string {
			return []string{
				"online_entry.create",
				fmt.Sprintf("online_entry.create.%s", data.ID),
				fmt.Sprintf("online_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("online_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *OnlineEntry) []string {
			return []string{
				"online_entry.update",
				fmt.Sprintf("online_entry.update.%s", data.ID),
				fmt.Sprintf("online_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("online_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *OnlineEntry) []string {
			return []string{
				"online_entry.delete",
				fmt.Sprintf("online_entry.delete.%s", data.ID),
				fmt.Sprintf("online_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("online_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) OnlineEntryCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*OnlineEntry, error) {
	return m.OnlineEntryManager.Find(context, &OnlineEntry{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
