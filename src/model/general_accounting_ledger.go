package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

// Enum for general_ledger_source
type GeneralLedgerSource string

const (
	GeneralLedgerSourceWithdraw       GeneralLedgerSource = "withdraw"
	GeneralLedgerSourceDeposit        GeneralLedgerSource = "deposit"
	GeneralLedgerSourceJournal        GeneralLedgerSource = "journal"
	GeneralLedgerSourcePayment        GeneralLedgerSource = "payment"
	GeneralLedgerSourceAdjustment     GeneralLedgerSource = "adjustment"
	GeneralLedgerSourceJournalVoucher GeneralLedgerSource = "journal voucher"
	GeneralLedgerSourceCheckVoucher   GeneralLedgerSource = "check voucher"
)

// Assumes you have TypesOfPaymentType defined elsewhere, as in your payment_type model

type (
	GeneralAccountingLedger struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_accounting_ledger"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_accounting_ledger"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID            *uuid.UUID          `gorm:"type:uuid"`
		Account              *Account            `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`
		TransactionID        *uuid.UUID          `gorm:"type:uuid"`
		Transaction          *Transaction        `gorm:"foreignKey:TransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction,omitempty"`
		TransactionBatchID   *uuid.UUID          `gorm:"type:uuid"`
		TransactionBatch     *TransactionBatch   `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`
		EmployeeUserID       *uuid.UUID          `gorm:"type:uuid"`
		EmployeeUser         *User               `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		MemberProfileID      *uuid.UUID          `gorm:"type:uuid"`
		MemberProfile        *MemberProfile      `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
		MemberJointAccountID *uuid.UUID          `gorm:"type:uuid"`
		MemberJointAccount   *MemberJointAccount `gorm:"foreignKey:MemberJointAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_joint_account,omitempty"`

		TransactionReferenceNumber string `gorm:"type:varchar(50)"`
		ReferenceNumber            string `gorm:"type:varchar(50)"`

		PaymentTypeID *uuid.UUID   `gorm:"type:uuid"`
		PaymentType   *PaymentType `gorm:"foreignKey:PaymentTypeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"payment_type,omitempty"`

		Source            GeneralLedgerSource `gorm:"type:varchar(20)"`
		JournalVoucherID  *uuid.UUID          `gorm:"type:uuid"`
		AdjustmentEntryID *uuid.UUID          `gorm:"type:uuid"`
		AdjustmentEntry   *AdjustmentEntry    `gorm:"foreignKey:AdjustmentEntryID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"adjustment_entry,omitempty"`
		TypeOfPaymentType TypesOfPaymentType  `gorm:"type:varchar(20)"`

		Credit  float64 `gorm:"type:decimal"`
		Debit   float64 `gorm:"type:decimal"`
		Balance float64 `gorm:"type:decimal"`
	}

	GeneralAccountingLedgerResponse struct {
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

		AccountID            *uuid.UUID                  `json:"account_id,omitempty"`
		Account              *AccountResponse            `json:"account,omitempty"`
		TransactionID        *uuid.UUID                  `json:"transaction_id,omitempty"`
		Transaction          *TransactionResponse        `json:"transaction,omitempty"`
		TransactionBatchID   *uuid.UUID                  `json:"transaction_batch_id,omitempty"`
		TransactionBatch     *TransactionBatchResponse   `json:"transaction_batch,omitempty"`
		EmployeeUserID       *uuid.UUID                  `json:"employee_user_id,omitempty"`
		EmployeeUser         *UserResponse               `json:"employee_user,omitempty"`
		MemberProfileID      *uuid.UUID                  `json:"member_profile_id,omitempty"`
		MemberProfile        *MemberProfileResponse      `json:"member_profile,omitempty"`
		MemberJointAccountID *uuid.UUID                  `json:"member_joint_account_id,omitempty"`
		MemberJointAccount   *MemberJointAccountResponse `json:"member_joint_account,omitempty"`

		TransactionReferenceNumber string `json:"transaction_reference_number"`
		ReferenceNumber            string `json:"reference_number"`

		PaymentTypeID *uuid.UUID           `json:"payment_type_id,omitempty"`
		PaymentType   *PaymentTypeResponse `json:"payment_type,omitempty"`

		Source            GeneralLedgerSource      `json:"source"`
		JournalVoucherID  *uuid.UUID               `json:"journal_voucher_id,omitempty"`
		AdjustmentEntryID *uuid.UUID               `json:"adjustment_entry_id,omitempty"`
		AdjustmentEntry   *AdjustmentEntryResponse `json:"adjustment_entry,omitempty"`
		TypeOfPaymentType TypesOfPaymentType       `json:"type_of_payment_type"`

		Credit  float64 `json:"credit"`
		Debit   float64 `json:"debit"`
		Balance float64 `json:"balance"`
	}

	GeneralAccountingLedgerRequest struct {
		OrganizationID             uuid.UUID           `json:"organization_id" validate:"required"`
		BranchID                   uuid.UUID           `json:"branch_id" validate:"required"`
		AccountID                  *uuid.UUID          `json:"account_id,omitempty"`
		TransactionID              *uuid.UUID          `json:"transaction_id,omitempty"`
		TransactionBatchID         *uuid.UUID          `json:"transaction_batch_id,omitempty"`
		EmployeeUserID             *uuid.UUID          `json:"employee_user_id,omitempty"`
		MemberProfileID            *uuid.UUID          `json:"member_profile_id,omitempty"`
		MemberJointAccountID       *uuid.UUID          `json:"member_joint_account_id,omitempty"`
		TransactionReferenceNumber string              `json:"transaction_reference_number,omitempty"`
		ReferenceNumber            string              `json:"reference_number,omitempty"`
		PaymentTypeID              *uuid.UUID          `json:"payment_type_id,omitempty"`
		Source                     GeneralLedgerSource `json:"source,omitempty"`
		JournalVoucherID           *uuid.UUID          `json:"journal_voucher_id,omitempty"`
		AdjustmentEntryID          *uuid.UUID          `json:"adjustment_entry_id,omitempty"`
		TypeOfPaymentType          TypesOfPaymentType  `json:"type_of_payment_type,omitempty"`
		Credit                     float64             `json:"credit,omitempty"`
		Debit                      float64             `json:"debit,omitempty"`
		Balance                    float64             `json:"balance,omitempty"`
	}
)

func (m *Model) GeneralAccountingLedger() {
	m.Migration = append(m.Migration, &GeneralAccountingLedger{})
	m.GeneralAccountingLedgerManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		GeneralAccountingLedger, GeneralAccountingLedgerResponse, GeneralAccountingLedgerRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
			"Account", "Transaction", "TransactionBatch", "EmployeeUser", "MemberProfile", "MemberJointAccount", "PaymentType", "AdjustmentEntry",
		},
		Service: m.provider.Service,
		Resource: func(data *GeneralAccountingLedger) *GeneralAccountingLedgerResponse {
			if data == nil {
				return nil
			}
			return &GeneralAccountingLedgerResponse{
				ID:                         data.ID,
				CreatedAt:                  data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                data.CreatedByID,
				CreatedBy:                  m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                  data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                data.UpdatedByID,
				UpdatedBy:                  m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:             data.OrganizationID,
				Organization:               m.OrganizationManager.ToModel(data.Organization),
				BranchID:                   data.BranchID,
				Branch:                     m.BranchManager.ToModel(data.Branch),
				AccountID:                  data.AccountID,
				Account:                    m.AccountManager.ToModel(data.Account),
				TransactionID:              data.TransactionID,
				Transaction:                m.TransactionManager.ToModel(data.Transaction),
				TransactionBatchID:         data.TransactionBatchID,
				TransactionBatch:           m.TransactionBatchManager.ToModel(data.TransactionBatch),
				EmployeeUserID:             data.EmployeeUserID,
				EmployeeUser:               m.UserManager.ToModel(data.EmployeeUser),
				MemberProfileID:            data.MemberProfileID,
				MemberProfile:              m.MemberProfileManager.ToModel(data.MemberProfile),
				MemberJointAccountID:       data.MemberJointAccountID,
				MemberJointAccount:         m.MemberJointAccountManager.ToModel(data.MemberJointAccount),
				TransactionReferenceNumber: data.TransactionReferenceNumber,
				ReferenceNumber:            data.ReferenceNumber,
				PaymentTypeID:              data.PaymentTypeID,
				PaymentType:                m.PaymentTypeManager.ToModel(data.PaymentType),
				Source:                     data.Source,
				JournalVoucherID:           data.JournalVoucherID,
				AdjustmentEntryID:          data.AdjustmentEntryID,
				AdjustmentEntry:            m.AdjustmentEntryManager.ToModel(data.AdjustmentEntry),
				TypeOfPaymentType:          data.TypeOfPaymentType,
				Credit:                     data.Credit,
				Debit:                      data.Debit,
				Balance:                    data.Balance,
			}
		},
		Created: func(data *GeneralAccountingLedger) []string {
			return []string{
				"general_accounting_ledger.create",
				fmt.Sprintf("general_accounting_ledger.create.%s", data.ID),
			}
		},
		Updated: func(data *GeneralAccountingLedger) []string {
			return []string{
				"general_accounting_ledger.update",
				fmt.Sprintf("general_accounting_ledger.update.%s", data.ID),
			}
		},
		Deleted: func(data *GeneralAccountingLedger) []string {
			return []string{
				"general_accounting_ledger.delete",
				fmt.Sprintf("general_accounting_ledger.delete.%s", data.ID),
			}
		},
	})
}
