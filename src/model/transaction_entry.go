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
	TransactionEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_transaction_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_transaction_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID      *uuid.UUID          `gorm:"type:uuid"`
		MemberProfile        *MemberProfile      `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
		EmployeeUserID       *uuid.UUID          `gorm:"type:uuid"`
		EmployeeUser         *User               `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		TransactionID        *uuid.UUID          `gorm:"type:uuid"`
		Transaction          *Transaction        `gorm:"foreignKey:TransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction,omitempty"`
		MemberJointAccountID *uuid.UUID          `gorm:"type:uuid"`
		MemberJointAccount   *MemberJointAccount `gorm:"foreignKey:MemberJointAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_joint_account,omitempty"`
		GeneralLedgerID      *uuid.UUID          `gorm:"type:uuid"`
		GeneralLedger        *GeneralLedger      `gorm:"foreignKey:GeneralLedgerID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"general_ledger,omitempty"`
		TransactionBatchID   *uuid.UUID          `gorm:"type:uuid"`
		TransactionBatch     *TransactionBatch   `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`
		AccountID            *uuid.UUID          `gorm:"type:uuid"`
		Account              *Account            `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`

		SignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		SignatureMedia   *Media     `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"signature_media,omitempty"`

		ReferenceNumber string  `gorm:"type:varchar(50)"`
		Debit           float64 `gorm:"type:decimal"`
		Credit          float64 `gorm:"type:decimal"`
	}

	TransactionEntryResponse struct {
		ID                   uuid.UUID                   `json:"id"`
		CreatedAt            string                      `json:"created_at"`
		CreatedByID          uuid.UUID                   `json:"created_by_id"`
		CreatedBy            *UserResponse               `json:"created_by,omitempty"`
		UpdatedAt            string                      `json:"updated_at"`
		UpdatedByID          uuid.UUID                   `json:"updated_by_id"`
		UpdatedBy            *UserResponse               `json:"updated_by,omitempty"`
		OrganizationID       uuid.UUID                   `json:"organization_id"`
		Organization         *OrganizationResponse       `json:"organization,omitempty"`
		BranchID             uuid.UUID                   `json:"branch_id"`
		Branch               *BranchResponse             `json:"branch,omitempty"`
		MemberProfileID      *uuid.UUID                  `json:"member_profile_id,omitempty"`
		MemberProfile        *MemberProfileResponse      `json:"member_profile,omitempty"`
		EmployeeUserID       *uuid.UUID                  `json:"employee_user_id,omitempty"`
		EmployeeUser         *UserResponse               `json:"employee_user,omitempty"`
		TransactionID        *uuid.UUID                  `json:"transaction_id,omitempty"`
		Transaction          *TransactionResponse        `json:"transaction,omitempty"`
		MemberJointAccountID *uuid.UUID                  `json:"member_joint_account_id,omitempty"`
		MemberJointAccount   *MemberJointAccountResponse `json:"member_joint_account,omitempty"`
		GeneralLedgerID      *uuid.UUID                  `json:"general_ledger_id,omitempty"`
		GeneralLedger        *GeneralLedgerResponse      `json:"general_ledger,omitempty"`
		TransactionBatchID   *uuid.UUID                  `json:"transaction_batch_id,omitempty"`
		TransactionBatch     *TransactionBatchResponse   `json:"transaction_batch,omitempty"`
		SignatureMediaID     *uuid.UUID                  `json:"signature_media_id,omitempty"`
		SignatureMedia       *MediaResponse              `json:"signature_media,omitempty"`
		AccountID            *uuid.UUID                  `json:"account_id,omitempty"`
		Account              *AccountResponse            `json:"account,omitempty"`
		ReferenceNumber      string                      `json:"reference_number"`
		Debit                float64                     `json:"debit"`
		Credit               float64                     `json:"credit"`
	}

	TransactionEntryRequest struct {
		OrganizationID       uuid.UUID  `json:"organization_id" validate:"required"`
		BranchID             uuid.UUID  `json:"branch_id" validate:"required"`
		MemberProfileID      *uuid.UUID `json:"member_profile_id,omitempty"`
		EmployeeUserID       *uuid.UUID `json:"employee_user_id,omitempty"`
		TransactionID        *uuid.UUID `json:"transaction_id,omitempty"`
		MemberJointAccountID *uuid.UUID `json:"member_joint_account_id,omitempty"`
		GeneralLedgerID      *uuid.UUID `json:"general_ledger_id,omitempty"`
		TransactionBatchID   *uuid.UUID `json:"transaction_batch_id,omitempty"`
		SignatureMediaID     *uuid.UUID `json:"signature_media_id,omitempty"`
		AccountID            *uuid.UUID `json:"account_id,omitempty"`
		ReferenceNumber      string     `json:"reference_number,omitempty"`
		Debit                float64    `json:"debit,omitempty"`
		Credit               float64    `json:"credit,omitempty"`
	}
)

func (m *Model) TransactionEntry() {
	m.Migration = append(m.Migration, &TransactionEntry{})
	m.TransactionEntryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		TransactionEntry, TransactionEntryResponse, TransactionEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
			"MemberProfile", "EmployeeUser", "Transaction", "MemberJointAccount",
			"GeneralLedger", "TransactionBatch", "SignatureMedia", "Account",
		},
		Service: m.provider.Service,
		Resource: func(data *TransactionEntry) *TransactionEntryResponse {
			if data == nil {
				return nil
			}
			return &TransactionEntryResponse{
				ID:                   data.ID,
				CreatedAt:            data.CreatedAt.Format(time.RFC3339),
				CreatedByID:          data.CreatedByID,
				CreatedBy:            m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:            data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:          data.UpdatedByID,
				UpdatedBy:            m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:       data.OrganizationID,
				Organization:         m.OrganizationManager.ToModel(data.Organization),
				BranchID:             data.BranchID,
				Branch:               m.BranchManager.ToModel(data.Branch),
				MemberProfileID:      data.MemberProfileID,
				MemberProfile:        m.MemberProfileManager.ToModel(data.MemberProfile),
				EmployeeUserID:       data.EmployeeUserID,
				EmployeeUser:         m.UserManager.ToModel(data.EmployeeUser),
				TransactionID:        data.TransactionID,
				Transaction:          m.TransactionManager.ToModel(data.Transaction),
				MemberJointAccountID: data.MemberJointAccountID,
				MemberJointAccount:   m.MemberJointAccountManager.ToModel(data.MemberJointAccount),
				GeneralLedgerID:      data.GeneralLedgerID,
				GeneralLedger:        m.GeneralLedgerManager.ToModel(data.GeneralLedger),
				TransactionBatchID:   data.TransactionBatchID,
				TransactionBatch:     m.TransactionBatchManager.ToModel(data.TransactionBatch),
				SignatureMediaID:     data.SignatureMediaID,
				SignatureMedia:       m.MediaManager.ToModel(data.SignatureMedia),
				AccountID:            data.AccountID,
				Account:              m.AccountManager.ToModel(data.Account),
				ReferenceNumber:      data.ReferenceNumber,
				Debit:                data.Debit,
				Credit:               data.Credit,
			}
		},

		Created: func(data *TransactionEntry) []string {
			return []string{
				"transaction_entry.create",
				fmt.Sprintf("transaction_entry.create.%s", data.ID),
				fmt.Sprintf("transaction_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("transaction_entry.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("transaction_entry.create.transaction.%s", data.TransactionID),
			}
		},
		Updated: func(data *TransactionEntry) []string {
			return []string{
				"transaction_entry.update",
				fmt.Sprintf("transaction_entry.update.%s", data.ID),
				fmt.Sprintf("transaction_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("transaction_entry.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("transaction_entry.update.transaction.%s", data.TransactionID),
			}
		},
		Deleted: func(data *TransactionEntry) []string {
			return []string{
				"transaction_entry.delete",
				fmt.Sprintf("transaction_entry.delete.%s", data.ID),
				fmt.Sprintf("transaction_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("transaction_entry.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("transaction_entry.delete.transaction.%s", data.TransactionID),
			}
		},
	})
}

func (m *Model) TransactionEntryCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*TransactionEntry, error) {
	return m.TransactionEntryManager.Find(context, &TransactionEntry{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
