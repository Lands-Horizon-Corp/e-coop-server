package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	CheckEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_check_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_check_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID                 *uuid.UUID               `gorm:"type:uuid"`
		Account                   *Account                 `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`
		MediaID                   *uuid.UUID               `gorm:"type:uuid"`
		Media                     *Media                   `gorm:"foreignKey:MediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"media,omitempty"`
		BankID                    *uuid.UUID               `gorm:"type:uuid"`
		Bank                      *Bank                    `gorm:"foreignKey:BankID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"bank,omitempty"`
		MemberProfileID           *uuid.UUID               `gorm:"type:uuid"`
		MemberProfile             *MemberProfile           `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
		MemberJointAccountID      *uuid.UUID               `gorm:"type:uuid"`
		MemberJointAccount        *MemberJointAccount      `gorm:"foreignKey:MemberJointAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_joint_account,omitempty"`
		EmployeeUserID            *uuid.UUID               `gorm:"type:uuid"`
		EmployeeUser              *User                    `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		TransactionID             *uuid.UUID               `gorm:"type:uuid"`
		Transaction               *Transaction             `gorm:"foreignKey:TransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction,omitempty"`
		TransactionBatchID        *uuid.UUID               `gorm:"type:uuid"`
		TransactionBatch          *TransactionBatch        `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`
		GeneralAccountingLedgerID *uuid.UUID               `gorm:"type:uuid"`
		GeneralAccountingLedger   *GeneralAccountingLedger `gorm:"foreignKey:GeneralAccountingLedgerID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"general_accounting_ledger,omitempty"`
		DisbursementTransactionID *uuid.UUID               `gorm:"type:uuid"`
		DisbursementTransaction   *DisbursementTransaction `gorm:"foreignKey:DisbursementTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"disbursement_transaction,omitempty"`

		Credit      float64    `gorm:"type:decimal"`
		Debit       float64    `gorm:"type:decimal"`
		CheckNumber string     `gorm:"type:varchar(255);not null"`
		CheckDate   *time.Time `gorm:"type:timestamp"`
	}

	CheckEntryResponse struct {
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
		MediaID                   *uuid.UUID                       `json:"media_id,omitempty"`
		Media                     *MediaResponse                   `json:"media,omitempty"`
		BankID                    *uuid.UUID                       `json:"bank_id,omitempty"`
		Bank                      *BankResponse                    `json:"bank,omitempty"`
		MemberProfileID           *uuid.UUID                       `json:"member_profile_id,omitempty"`
		MemberProfile             *MemberProfileResponse           `json:"member_profile,omitempty"`
		MemberJointAccountID      *uuid.UUID                       `json:"member_joint_account_id,omitempty"`
		MemberJointAccount        *MemberJointAccountResponse      `json:"member_joint_account,omitempty"`
		EmployeeUserID            *uuid.UUID                       `json:"employee_user_id,omitempty"`
		EmployeeUser              *UserResponse                    `json:"employee_user,omitempty"`
		TransactionID             *uuid.UUID                       `json:"transaction_id,omitempty"`
		Transaction               *TransactionResponse             `json:"transaction,omitempty"`
		TransactionBatchID        *uuid.UUID                       `json:"transaction_batch_id,omitempty"`
		TransactionBatch          *TransactionBatchResponse        `json:"transaction_batch,omitempty"`
		GeneralAccountingLedgerID *uuid.UUID                       `json:"general_accounting_ledger_id,omitempty"`
		GeneralAccountingLedger   *GeneralAccountingLedgerResponse `json:"general_accounting_ledger,omitempty"`
		DisbursementTransactionID *uuid.UUID                       `json:"disbursement_transaction_id,omitempty"`
		DisbursementTransaction   *DisbursementTransactionResponse `json:"disbursement_transaction,omitempty"`
		Credit                    float64                          `json:"credit"`
		Debit                     float64                          `json:"debit"`
		CheckNumber               string                           `json:"check_number"`
		CheckDate                 *string                          `json:"check_date,omitempty"`
	}

	CheckEntryRequest struct {
		OrganizationID            uuid.UUID  `json:"organization_id" validate:"required"`
		BranchID                  uuid.UUID  `json:"branch_id" validate:"required"`
		AccountID                 *uuid.UUID `json:"account_id,omitempty"`
		MediaID                   *uuid.UUID `json:"media_id,omitempty"`
		BankID                    *uuid.UUID `json:"bank_id,omitempty"`
		MemberProfileID           *uuid.UUID `json:"member_profile_id,omitempty"`
		MemberJointAccountID      *uuid.UUID `json:"member_joint_account_id,omitempty"`
		EmployeeUserID            *uuid.UUID `json:"employee_user_id,omitempty"`
		TransactionID             *uuid.UUID `json:"transaction_id,omitempty"`
		TransactionBatchID        *uuid.UUID `json:"transaction_batch_id,omitempty"`
		GeneralAccountingLedgerID *uuid.UUID `json:"general_accounting_ledger_id,omitempty"`
		DisbursementTransactionID *uuid.UUID `json:"disbursement_transaction_id,omitempty"`
		Credit                    float64    `json:"credit,omitempty"`
		Debit                     float64    `json:"debit,omitempty"`
		CheckNumber               string     `json:"check_number" validate:"required,min=1,max=255"`
		CheckDate                 *time.Time `json:"check_date,omitempty"`
	}
)

func (m *Model) CheckEntry() {
	m.Migration = append(m.Migration, &CheckEntry{})
	m.CheckEntryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		CheckEntry, CheckEntryResponse, CheckEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
			"Account", "Media", "Bank", "MemberProfile", "MemberJointAccount", "EmployeeUser",
			"Transaction", "TransactionBatch", "GeneralAccountingLedger", "DisbursementTransaction",
		},
		Service: m.provider.Service,
		Resource: func(data *CheckEntry) *CheckEntryResponse {
			if data == nil {
				return nil
			}
			var checkDate *string
			if data.CheckDate != nil {
				s := data.CheckDate.Format(time.RFC3339)
				checkDate = &s
			}
			return &CheckEntryResponse{
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
				MediaID:                   data.MediaID,
				Media:                     m.MediaManager.ToModel(data.Media),
				BankID:                    data.BankID,
				Bank:                      m.BankManager.ToModel(data.Bank),
				MemberProfileID:           data.MemberProfileID,
				MemberProfile:             m.MemberProfileManager.ToModel(data.MemberProfile),
				MemberJointAccountID:      data.MemberJointAccountID,
				MemberJointAccount:        m.MemberJointAccountManager.ToModel(data.MemberJointAccount),
				EmployeeUserID:            data.EmployeeUserID,
				EmployeeUser:              m.UserManager.ToModel(data.EmployeeUser),
				TransactionID:             data.TransactionID,
				Transaction:               m.TransactionManager.ToModel(data.Transaction),
				TransactionBatchID:        data.TransactionBatchID,
				TransactionBatch:          m.TransactionBatchManager.ToModel(data.TransactionBatch),
				GeneralAccountingLedgerID: data.GeneralAccountingLedgerID,
				GeneralAccountingLedger:   m.GeneralAccountingLedgerManager.ToModel(data.GeneralAccountingLedger),
				DisbursementTransactionID: data.DisbursementTransactionID,
				DisbursementTransaction:   m.DisbursementTransactionManager.ToModel(data.DisbursementTransaction),
				Credit:                    data.Credit,
				Debit:                     data.Debit,
				CheckNumber:               data.CheckNumber,
				CheckDate:                 checkDate,
			}
		},
		Created: func(data *CheckEntry) []string {
			return []string{
				"check_entry.create",
				fmt.Sprintf("check_entry.create.%s", data.ID),
			}
		},
		Updated: func(data *CheckEntry) []string {
			return []string{
				"check_entry.update",
				fmt.Sprintf("check_entry.update.%s", data.ID),
			}
		},
		Deleted: func(data *CheckEntry) []string {
			return []string{
				"check_entry.delete",
				fmt.Sprintf("check_entry.delete.%s", data.ID),
			}
		},
	})
}
