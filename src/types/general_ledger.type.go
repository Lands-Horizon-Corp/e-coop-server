package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	GeneralLedgerSourceWithdraw           GeneralLedgerSource = "withdraw"
	GeneralLedgerSourceDeposit            GeneralLedgerSource = "deposit"
	GeneralLedgerSourcePayment            GeneralLedgerSource = "payment"
	GeneralLedgerSourceAdjustment         GeneralLedgerSource = "adjustment"
	GeneralLedgerSourceJournalVoucher     GeneralLedgerSource = "journal voucher"
	GeneralLedgerSourceCheckVoucher       GeneralLedgerSource = "check voucher"
	GeneralLedgerSourceLoan               GeneralLedgerSource = "loan"
	GeneralLedgerSourceSavingsInterest    GeneralLedgerSource = "savings interest"
	GeneralLedgerSourceMutualContribution GeneralLedgerSource = "mutual contribution"
	GeneralLedgerSourcDisbursement        GeneralLedgerSource = "disbursement"
	GeneralLedgerSourcBlotter             GeneralLedgerSource = "blotter"
)

type (
	GeneralLedgerSource string

	GeneralLedger struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now();index" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id,omitempty"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger;index:idx_org_branch_account_member;index:idx_transaction_batch_entry" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`

		BranchID uuid.UUID `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger;index:idx_org_branch_account_member;index:idx_transaction_batch_entry" json:"branch_id"`
		Branch   *Branch   `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID *uuid.UUID `gorm:"type:uuid;index:idx_org_branch_account_member" json:"account_id,omitempty"`
		Account   *Account   `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`

		TransactionID *uuid.UUID   `gorm:"type:uuid" json:"transaction_id,omitempty"`
		Transaction   *Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction,omitempty"`

		TransactionBatchID *uuid.UUID        `gorm:"type:uuid;index:idx_transaction_batch_entry" json:"transaction_batch_id,omitempty"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`

		EmployeeUserID *uuid.UUID `gorm:"type:uuid" json:"employee_user_id,omitempty"`
		EmployeeUser   *User      `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid;index:idx_org_branch_account_member" json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		MemberJointAccountID *uuid.UUID          `gorm:"type:uuid" json:"member_joint_account_id,omitempty"`
		MemberJointAccount   *MemberJointAccount `gorm:"foreignKey:MemberJointAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_joint_account,omitempty"`

		TransactionReferenceNumber string `gorm:"type:varchar(50)" json:"transaction_reference_number"`
		ReferenceNumber            string `gorm:"type:varchar(50)" json:"reference_number"`

		PaymentTypeID *uuid.UUID   `gorm:"type:uuid" json:"payment_type_id,omitempty"`
		PaymentType   *PaymentType `gorm:"foreignKey:PaymentTypeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"payment_type,omitempty"`

		Source GeneralLedgerSource `gorm:"type:varchar(20)" json:"source"`

		JournalVoucherID *uuid.UUID `gorm:"type:uuid" json:"journal_voucher_id,omitempty"`

		AdjustmentEntryID *uuid.UUID       `gorm:"type:uuid" json:"adjustment_entry_id,omitempty"`
		AdjustmentEntry   *AdjustmentEntry `gorm:"foreignKey:AdjustmentEntryID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"adjustment_entry,omitempty"`

		LoanTransactionID *uuid.UUID       `gorm:"type:uuid" json:"loan_transaction_id,omitempty"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		TypeOfPaymentType TypeOfPaymentType `gorm:"type:varchar(20)" json:"type_of_payment_type,omitempty"`

		Credit  float64 `gorm:"type:decimal" json:"credit"`
		Debit   float64 `gorm:"type:decimal" json:"debit"`
		Balance float64 `gorm:"type:decimal" json:"balance"`

		SignatureMediaID *uuid.UUID `gorm:"type:uuid" json:"signature_media_id,omitempty"`
		SignatureMedia   *Media     `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"signature_media,omitempty"`

		EntryDate time.Time `gorm:"type:timestamp;not null;default:now()" json:"entry_date"`

		BankID *uuid.UUID `gorm:"type:uuid" json:"bank_id,omitempty"`
		Bank   *Bank      `gorm:"foreignKey:BankID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"bank,omitempty"`

		ProofOfPaymentMediaID *uuid.UUID `gorm:"type:uuid" json:"proof_of_payment_media_id,omitempty"`
		ProofOfPaymentMedia   *Media     `gorm:"foreignKey:ProofOfPaymentMediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"proof_of_payment_media,omitempty"`

		CurrencyID *uuid.UUID `gorm:"type:uuid" json:"currency_id,omitempty"`
		Currency   *Currency  `gorm:"foreignKey:CurrencyID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"currency,omitempty"`

		BankReferenceNumber string `gorm:"type:varchar(50)" json:"bank_reference_number"`
		Description         string `gorm:"type:text" json:"description"`
		PrintNumber         int    `gorm:"default:0" json:"print_number"`
	}

	GeneralLedgerResponse struct {
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
		LoanTransactionID *uuid.UUID               `json:"loan_transaction_id,omitempty"`
		LoanTransaction   *LoanTransactionResponse `json:"loan_transaction,omitempty"`
		TypeOfPaymentType TypeOfPaymentType        `json:"type_of_payment_type"`

		Credit  float64 `json:"credit"`
		Debit   float64 `json:"debit"`
		Balance float64 `json:"balance"`

		SignatureMediaID *uuid.UUID     `json:"signature_media_id,omitempty"`
		SignatureMedia   *MediaResponse `json:"signature_media,omitempty"`

		EntryDate string `json:"entry_date,omitempty"`

		BankID *uuid.UUID    `json:"bank_id,omitempty"`
		Bank   *BankResponse `json:"bank,omitempty"`

		ProofOfPaymentMediaID *uuid.UUID     `json:"proof_of_payment_media_id,omitempty"`
		ProofOfPaymentMedia   *MediaResponse `json:"proof_of_payment_media,omitempty"`

		CurrencyID *uuid.UUID        `json:"currency_id,omitempty"`
		Currency   *CurrencyResponse `json:"currency,omitempty"`

		BankReferenceNumber string `json:"bank_reference_number,omitempty"`

		Description string `json:"description,omitempty"`
		PrintNumber int    `json:"print_number"`

		AccountHistoryID *uuid.UUID `json:"account_history_id"`
	}

	GeneralLedgerRequest struct {
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
		LoanTransactionID          *uuid.UUID          `json:"loan_transaction_id,omitempty"`

		TypeOfPaymentType     TypeOfPaymentType `json:"type_of_payment_type,omitempty"`
		Credit                float64           `json:"credit,omitempty"`
		Debit                 float64           `json:"debit,omitempty"`
		SignatureMediaID      *uuid.UUID        `json:"signature_media_id,omitempty"`
		EntryDate             *time.Time        `json:"entry_date,omitempty"`
		BankID                *uuid.UUID        `json:"bank_id,omitempty"`
		ProofOfPaymentMediaID *uuid.UUID        `json:"proof_of_payment_media_id,omitempty"`
		CurrencyID            *uuid.UUID        `json:"currency_id,omitempty"`
		BankReferenceNumber   string            `json:"bank_reference_number,omitempty"`
		Description           string            `json:"description,omitempty"`
	}

	PaymentRequest struct {
		Amount                float64    `json:"amount" validate:"required,ne=0"`
		SignatureMediaID      *uuid.UUID `json:"signature_media_id,omitempty"`
		ProofOfPaymentMediaID *uuid.UUID `json:"proof_of_payment_media_id,omitempty"`
		BankID                *uuid.UUID `json:"bank_id,omitempty"`
		BankReferenceNumber   string     `json:"bank_reference_number,omitempty"`
		EntryDate             *time.Time `json:"entry_date,omitempty"`
		AccountID             *uuid.UUID `json:"account_id,omitempty"`
		PaymentTypeID         *uuid.UUID `json:"payment_type_id,omitempty"`
		Description           string     `json:"description,omitempty" validate:"max=255"`
		LoanTransactionID     *uuid.UUID `json:"loan_transaction_id,omitempty"`
	}

	PaymentQuickRequest struct {
		Amount                float64    `json:"amount" validate:"required,ne=0"`
		SignatureMediaID      *uuid.UUID `json:"signature_media_id,omitempty"`
		ProofOfPaymentMediaID *uuid.UUID `json:"proof_of_payment_media_id,omitempty"`
		BankID                *uuid.UUID `json:"bank_id,omitempty"`
		BankReferenceNumber   string     `json:"bank_reference_number,omitempty"`
		EntryDate             *time.Time `json:"entry_date,omitempty"`
		AccountID             *uuid.UUID `json:"account_id,omitempty"`
		PaymentTypeID         *uuid.UUID `json:"payment_type_id,omitempty"`
		Description           string     `json:"description,omitempty" validate:"max=255"`

		MemberProfileID      *uuid.UUID `json:"member_profile_id,omitempty"`
		MemberJointAccountID *uuid.UUID `json:"member_joint_account_id,omitempty"`
		ReferenceNumber      string     `json:"reference_number,omitempty"`
		ORAutoGenerated      bool       `json:"or_auto_generated,omitempty"`
		LoanTransactionID    *uuid.UUID `json:"loan_transaction_id,omitempty"`
	}

	MemberGeneralLedgerTotal struct {
		Balance     float64 `json:"balance"`
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
	}
)
