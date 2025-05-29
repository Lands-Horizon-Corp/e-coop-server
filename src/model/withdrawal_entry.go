package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	WithdrawalEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_withdrawal_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_withdrawal_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID           *uuid.UUID               `gorm:"type:uuid"`
		MemberProfile             *MemberProfile           `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
		TransactionID             *uuid.UUID               `gorm:"type:uuid"`
		Transaction               *Transaction             `gorm:"foreignKey:TransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction,omitempty"`
		MemberJointAccountID      *uuid.UUID               `gorm:"type:uuid"`
		MemberJointAccount        *MemberJointAccount      `gorm:"foreignKey:MemberJointAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_joint_account,omitempty"`
		GeneralAccountingLedgerID *uuid.UUID               `gorm:"type:uuid"`
		GeneralAccountingLedger   *GeneralAccountingLedger `gorm:"foreignKey:GeneralAccountingLedgerID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"general_accounting_ledger,omitempty"`
		TransactionBatchID        *uuid.UUID               `gorm:"type:uuid"`
		TransactionBatch          *TransactionBatch        `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`
		SignatureMediaID          *uuid.UUID               `gorm:"type:uuid"`
		SignatureMedia            *Media                   `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"signature_media,omitempty"`
		AccountID                 *uuid.UUID               `gorm:"type:uuid"`
		Account                   *Account                 `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`
		EmployeeUserID            *uuid.UUID               `gorm:"type:uuid"`
		EmployeeUser              *User                    `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`

		ReferenceNumber string  `gorm:"type:varchar(50)"`
		Debit           float64 `gorm:"type:decimal"`
		Credit          float64 `gorm:"type:decimal"`
	}

	WithdrawalEntryResponse struct {
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
		MemberProfileID           *uuid.UUID                       `json:"member_profile_id,omitempty"`
		MemberProfile             *MemberProfileResponse           `json:"member_profile,omitempty"`
		TransactionID             *uuid.UUID                       `json:"transaction_id,omitempty"`
		Transaction               *TransactionResponse             `json:"transaction,omitempty"`
		MemberJointAccountID      *uuid.UUID                       `json:"member_joint_account_id,omitempty"`
		MemberJointAccount        *MemberJointAccountResponse      `json:"member_joint_account,omitempty"`
		GeneralAccountingLedgerID *uuid.UUID                       `json:"general_accounting_ledger_id,omitempty"`
		GeneralAccountingLedger   *GeneralAccountingLedgerResponse `json:"general_accounting_ledger,omitempty"`
		TransactionBatchID        *uuid.UUID                       `json:"transaction_batch_id,omitempty"`
		TransactionBatch          *TransactionBatchResponse        `json:"transaction_batch,omitempty"`
		SignatureMediaID          *uuid.UUID                       `json:"signature_media_id,omitempty"`
		SignatureMedia            *MediaResponse                   `json:"signature_media,omitempty"`
		AccountID                 *uuid.UUID                       `json:"account_id,omitempty"`
		Account                   *AccountResponse                 `json:"account,omitempty"`
		EmployeeUserID            *uuid.UUID                       `json:"employee_user_id,omitempty"`
		EmployeeUser              *UserResponse                    `json:"employee_user,omitempty"`
		ReferenceNumber           string                           `json:"reference_number"`
		Debit                     float64                          `json:"debit"`
		Credit                    float64                          `json:"credit"`
	}

	WithdrawalEntryRequest struct {
		OrganizationID            uuid.UUID  `json:"organization_id" validate:"required"`
		BranchID                  uuid.UUID  `json:"branch_id" validate:"required"`
		MemberProfileID           *uuid.UUID `json:"member_profile_id,omitempty"`
		TransactionID             *uuid.UUID `json:"transaction_id,omitempty"`
		MemberJointAccountID      *uuid.UUID `json:"member_joint_account_id,omitempty"`
		GeneralAccountingLedgerID *uuid.UUID `json:"general_accounting_ledger_id,omitempty"`
		TransactionBatchID        *uuid.UUID `json:"transaction_batch_id,omitempty"`
		SignatureMediaID          *uuid.UUID `json:"signature_media_id,omitempty"`
		AccountID                 *uuid.UUID `json:"account_id,omitempty"`
		EmployeeUserID            *uuid.UUID `json:"employee_user_id,omitempty"`
		ReferenceNumber           string     `json:"reference_number,omitempty"`
		Debit                     float64    `json:"debit,omitempty"`
		Credit                    float64    `json:"credit,omitempty"`
	}
)

func (m *Model) WithdrawalEntry() {
	m.Migration = append(m.Migration, &WithdrawalEntry{})
	m.WithdrawalEntryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		WithdrawalEntry, WithdrawalEntryResponse, WithdrawalEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
			"MemberProfile", "Transaction", "MemberJointAccount", "GeneralAccountingLedger",
			"TransactionBatch", "SignatureMedia", "Account", "EmployeeUser",
		},
		Service: m.provider.Service,
		Resource: func(data *WithdrawalEntry) *WithdrawalEntryResponse {
			if data == nil {
				return nil
			}
			return &WithdrawalEntryResponse{
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
				MemberProfileID:           data.MemberProfileID,
				MemberProfile:             m.MemberProfileManager.ToModel(data.MemberProfile),
				TransactionID:             data.TransactionID,
				Transaction:               m.TransactionManager.ToModel(data.Transaction),
				MemberJointAccountID:      data.MemberJointAccountID,
				MemberJointAccount:        m.MemberJointAccountManager.ToModel(data.MemberJointAccount),
				GeneralAccountingLedgerID: data.GeneralAccountingLedgerID,
				GeneralAccountingLedger:   m.GeneralAccountingLedgerManager.ToModel(data.GeneralAccountingLedger),
				TransactionBatchID:        data.TransactionBatchID,
				TransactionBatch:          m.TransactionBatchManager.ToModel(data.TransactionBatch),
				SignatureMediaID:          data.SignatureMediaID,
				SignatureMedia:            m.MediaManager.ToModel(data.SignatureMedia),
				AccountID:                 data.AccountID,
				Account:                   m.AccountManager.ToModel(data.Account),
				EmployeeUserID:            data.EmployeeUserID,
				EmployeeUser:              m.UserManager.ToModel(data.EmployeeUser),
				ReferenceNumber:           data.ReferenceNumber,
				Debit:                     data.Debit,
				Credit:                    data.Credit,
			}
		},
		Created: func(data *WithdrawalEntry) []string {
			return []string{
				"withdrawal_entry.create",
				fmt.Sprintf("withdrawal_entry.create.%s", data.ID),
			}
		},
		Updated: func(data *WithdrawalEntry) []string {
			return []string{
				"withdrawal_entry.update",
				fmt.Sprintf("withdrawal_entry.update.%s", data.ID),
			}
		},
		Deleted: func(data *WithdrawalEntry) []string {
			return []string{
				"withdrawal_entry.delete",
				fmt.Sprintf("withdrawal_entry.delete.%s", data.ID),
			}
		},
	})
}
