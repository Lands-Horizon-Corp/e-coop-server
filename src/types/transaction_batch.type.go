package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TransactionBatchBalanced        TransactionBatchBalanceStatus = "balanced"
	TransactionBatchBalanceOverage  TransactionBatchBalanceStatus = "balance overage"
	TransactionBatchBalanceShortage TransactionBatchBalanceStatus = "balance shortage"
)

type (
	TransactionBatchBalanceStatus string
	TransactionBatch              struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_transaction_batch"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_transaction_batch"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		EmployeeUserID *uuid.UUID `gorm:"type:uuid" json:"employee_user_id,omitempty"`
		EmployeeUser   *User      `gorm:"foreignKey:EmployeeUserID" json:"employee_user,omitempty"`

		BatchName               string  `gorm:"type:varchar(50)"`
		TotalCashCollection     float64 `gorm:"type:decimal"`
		TotalDepositEntry       float64 `gorm:"type:decimal"`
		BeginningBalance        float64 `gorm:"type:decimal"`
		DepositInBank           float64 `gorm:"type:decimal"`
		CashCountTotal          float64 `gorm:"type:decimal"`
		GrandTotal              float64 `gorm:"type:decimal"`
		PettyCash               float64 `gorm:"type:decimal"`
		LoanReleases            float64 `gorm:"type:decimal"`
		CashCheckVoucherTotal   float64 `gorm:"type:decimal"`
		TimeDepositWithdrawal   float64 `gorm:"type:decimal"`
		SavingsWithdrawal       float64 `gorm:"type:decimal"`
		TotalCashHandled        float64 `gorm:"type:decimal"`
		TotalSupposedRemmitance float64 `gorm:"type:decimal"`

		TotalCashOnHand               float64 `gorm:"type:decimal"`
		TotalCheckRemittance          float64 `gorm:"type:decimal"`
		TotalOnlineRemittance         float64 `gorm:"type:decimal"`
		TotalDepositInBank            float64 `gorm:"type:decimal"`
		TotalActualRemittance         float64 `gorm:"type:decimal"`
		TotalActualSupposedComparison float64 `gorm:"type:decimal"`

		Description string `gorm:"type:text"`

		CanView     bool `gorm:"not null;default:false"`
		IsClosed    bool `gorm:"not null;default:false"`
		RequestView bool `gorm:"not null;default:false" json:"request_view"`

		EmployeeBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		EmployeeBySignatureMedia   *Media     `gorm:"foreignKey:EmployeeBySignatureMediaID" json:"employee_by_signature_media,omitempty"`
		EmployeeByName             string     `gorm:"type:varchar(255)"`
		EmployeeByPosition         string     `gorm:"type:varchar(255)"`

		ApprovedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		ApprovedBySignatureMedia   *Media     `gorm:"foreignKey:ApprovedBySignatureMediaID" json:"approved_by_signature_media,omitempty"`
		ApprovedByName             string     `gorm:"type:varchar(255)"`
		ApprovedByPosition         string     `gorm:"type:varchar(255)"`

		PreparedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		PreparedBySignatureMedia   *Media     `gorm:"foreignKey:PreparedBySignatureMediaID" json:"prepared_by_signature_media,omitempty"`
		PreparedByName             string     `gorm:"type:varchar(255)"`
		PreparedByPosition         string     `gorm:"type:varchar(255)"`

		CertifiedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		CertifiedBySignatureMedia   *Media     `gorm:"foreignKey:CertifiedBySignatureMediaID" json:"certified_by_signature_media,omitempty"`
		CertifiedByName             string     `gorm:"type:varchar(255)"`
		CertifiedByPosition         string     `gorm:"type:varchar(255)"`

		VerifiedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		VerifiedBySignatureMedia   *Media     `gorm:"foreignKey:VerifiedBySignatureMediaID" json:"verified_by_signature_media,omitempty"`
		VerifiedByName             string     `gorm:"type:varchar(255)"`
		VerifiedByPosition         string     `gorm:"type:varchar(255)"`

		CheckBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		CheckBySignatureMedia   *Media     `gorm:"foreignKey:CheckBySignatureMediaID" json:"check_by_signature_media,omitempty"`
		CheckByName             string     `gorm:"type:varchar(255)"`
		CheckByPosition         string     `gorm:"type:varchar(255)"`

		AcknowledgeBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		AcknowledgeBySignatureMedia   *Media     `gorm:"foreignKey:AcknowledgeBySignatureMediaID" json:"acknowledge_by_signature_media,omitempty"`
		AcknowledgeByName             string     `gorm:"type:varchar(255)"`
		AcknowledgeByPosition         string     `gorm:"type:varchar(255)"`

		NotedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		NotedBySignatureMedia   *Media     `gorm:"foreignKey:NotedBySignatureMediaID" json:"noted_by_signature_media,omitempty"`
		NotedByName             string     `gorm:"type:varchar(255)"`
		NotedByPosition         string     `gorm:"type:varchar(255)"`

		PostedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		PostedBySignatureMedia   *Media     `gorm:"foreignKey:PostedBySignatureMediaID" json:"posted_by_signature_media,omitempty"`
		PostedByName             string     `gorm:"type:varchar(255)"`
		PostedByPosition         string     `gorm:"type:varchar(255)"`

		PaidBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		PaidBySignatureMedia   *Media     `gorm:"foreignKey:PaidBySignatureMediaID" json:"paid_by_signature_media,omitempty"`
		PaidByName             string     `gorm:"type:varchar(255)"`
		PaidByPosition         string     `gorm:"type:varchar(255)"`

		CurrencyID uuid.UUID `gorm:"type:uuid;not null" json:"currency_id"`
		Currency   *Currency `gorm:"foreignKey:CurrencyID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"currency,omitempty"`

		EndedAt *time.Time `gorm:"type:timestamp"`

		UnbalancedAccountID uuid.UUID          `gorm:"type:uuid;not null" json:"unbalanced_account_id"`
		UnbalancedAccount   *UnbalancedAccount `gorm:"foreignKey:UnbalancedAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"unbalanced_account,omitempty"`
	}

	TransactionBatchResponse struct {
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
		EmployeeUserID *uuid.UUID            `json:"employee_user_id,omitempty"`
		EmployeeUser   *UserResponse         `json:"employee_user,omitempty"`

		BatchName                     string  `json:"batch_name"`
		TotalCashCollection           float64 `json:"total_cash_collection"`
		TotalDepositEntry             float64 `json:"total_deposit_entry"`
		BeginningBalance              float64 `json:"beginning_balance"`
		DepositInBank                 float64 `json:"deposit_in_bank"`
		CashCountTotal                float64 `json:"cash_count_total"`
		GrandTotal                    float64 `json:"grand_total"`
		PettyCash                     float64 `json:"petty_cash"`
		LoanReleases                  float64 `json:"loan_releases"`
		CashCheckVoucherTotal         float64 `json:"cash_check_voucher_total"`
		TimeDepositWithdrawal         float64 `json:"time_deposit_withdrawal"`
		SavingsWithdrawal             float64 `json:"savings_withdrawal"`
		TotalCashHandled              float64 `json:"total_cash_handled"`
		TotalSupposedRemmitance       float64 `json:"total_supposed_remitance"`
		TotalCashOnHand               float64 `json:"total_cash_on_hand"`
		TotalCheckRemittance          float64 `json:"total_check_remittance"`
		TotalOnlineRemittance         float64 `json:"total_online_remittance"`
		TotalDepositInBank            float64 `json:"total_deposit_in_bank"`
		TotalActualRemittance         float64 `json:"total_actual_remittance"`
		TotalActualSupposedComparison float64 `json:"total_actual_supposed_comparison"`
		Description                   string  `json:"description"`

		CanView     bool `json:"can_view"`
		IsClosed    bool `json:"is_closed"`
		RequestView bool `json:"request_view"`

		EmployeeBySignatureMediaID *uuid.UUID     `json:"employee_by_signature_media_id,omitempty"`
		EmployeeBySignatureMedia   *MediaResponse `json:"employee_by_signature_media,omitempty"`
		EmployeeByName             string         `json:"employee_by_name"`
		EmployeeByPosition         string         `json:"employee_by_position"`

		ApprovedBySignatureMediaID *uuid.UUID     `json:"approved_by_signature_media_id,omitempty"`
		ApprovedBySignatureMedia   *MediaResponse `json:"approved_by_signature_media,omitempty"`
		ApprovedByName             string         `json:"approved_by_name"`
		ApprovedByPosition         string         `json:"approved_by_position"`

		PreparedBySignatureMediaID *uuid.UUID     `json:"prepared_by_signature_media_id,omitempty"`
		PreparedBySignatureMedia   *MediaResponse `json:"prepared_by_signature_media,omitempty"`
		PreparedByName             string         `json:"prepared_by_name"`
		PreparedByPosition         string         `json:"prepared_by_position"`

		CertifiedBySignatureMediaID *uuid.UUID     `json:"certified_by_signature_media_id,omitempty"`
		CertifiedBySignatureMedia   *MediaResponse `json:"certified_by_signature_media,omitempty"`
		CertifiedByName             string         `json:"certified_by_name"`
		CertifiedByPosition         string         `json:"certified_by_position"`

		VerifiedBySignatureMediaID *uuid.UUID     `json:"verified_by_signature_media_id,omitempty"`
		VerifiedBySignatureMedia   *MediaResponse `json:"verified_by_signature_media,omitempty"`
		VerifiedByName             string         `json:"verified_by_name"`
		VerifiedByPosition         string         `json:"verified_by_position"`

		CheckBySignatureMediaID *uuid.UUID     `json:"check_by_signature_media_id,omitempty"`
		CheckBySignatureMedia   *MediaResponse `json:"check_by_signature_media,omitempty"`
		CheckByName             string         `json:"check_by_name"`
		CheckByPosition         string         `json:"check_by_position"`

		AcknowledgeBySignatureMediaID *uuid.UUID     `json:"acknowledge_by_signature_media_id,omitempty"`
		AcknowledgeBySignatureMedia   *MediaResponse `json:"acknowledge_by_signature_media,omitempty"`
		AcknowledgeByName             string         `json:"acknowledge_by_name"`
		AcknowledgeByPosition         string         `json:"acknowledge_by_position"`

		NotedBySignatureMediaID *uuid.UUID     `json:"noted_by_signature_media_id,omitempty"`
		NotedBySignatureMedia   *MediaResponse `json:"noted_by_signature_media,omitempty"`
		NotedByName             string         `json:"noted_by_name"`
		NotedByPosition         string         `json:"noted_by_position"`

		PostedBySignatureMediaID *uuid.UUID     `json:"posted_by_signature_media_id,omitempty"`
		PostedBySignatureMedia   *MediaResponse `json:"posted_by_signature_media,omitempty"`
		PostedByName             string         `json:"posted_by_name"`
		PostedByPosition         string         `json:"posted_by_position"`

		PaidBySignatureMediaID *uuid.UUID     `json:"paid_by_signature_media_id,omitempty"`
		PaidBySignatureMedia   *MediaResponse `json:"paid_by_signature_media,omitempty"`
		PaidByName             string         `json:"paid_by_name"`
		PaidByPosition         string         `json:"paid_by_position"`

		CurrencyID uuid.UUID         `json:"currency_id"`
		Currency   *CurrencyResponse `json:"currency,omitempty"`

		EndedAt *string `json:"ended_at,omitempty"`

		UnbalancedAccountID uuid.UUID                  `json:"unbalanced_account_id"`
		UnbalancedAccount   *UnbalancedAccountResponse `json:"unbalanced_account,omitempty"`
	}

	TransactionBatchRequest struct {
		OrganizationID                uuid.UUID  `json:"organization_id" validate:"required"`
		BranchID                      uuid.UUID  `json:"branch_id" validate:"required"`
		EmployeeUserID                *uuid.UUID `json:"employee_user_id,omitempty"`
		BatchName                     string     `json:"batch_name" validate:"required,min=1,max=50"`
		TotalCashCollection           float64    `json:"total_cash_collection,omitempty"`
		TotalDepositEntry             float64    `json:"total_deposit_entry,omitempty"`
		BeginningBalance              float64    `json:"beginning_balance,omitempty"`
		DepositInBank                 float64    `json:"deposit_in_bank,omitempty"`
		CashCountTotal                float64    `json:"cash_count_total,omitempty"`
		GrandTotal                    float64    `json:"grand_total,omitempty"`
		PettyCash                     float64    `json:"petty_cash,omitempty"`
		CashCheckVoucherTotal         float64    `json:"cash_check_voucher_total,omitempty"`
		LoanReleases                  float64    `json:"loan_releases,omitempty"`
		TimeDepositWithdrawal         float64    `json:"time_deposit_withdrawal,omitempty"`
		SavingsWithdrawal             float64    `json:"savings_withdrawal,omitempty"`
		TotalCashHandled              float64    `json:"total_cash_handled,omitempty"`
		TotalSupposedRemmitance       float64    `json:"total_supposed_remitance,omitempty"`
		TotalCashOnHand               float64    `json:"total_cash_on_hand,omitempty"`
		TotalCheckRemittance          float64    `json:"total_check_remittance,omitempty"`
		TotalOnlineRemittance         float64    `json:"total_online_remittance,omitempty"`
		TotalDepositInBank            float64    `json:"total_deposit_in_bank,omitempty"`
		TotalActualRemittance         float64    `json:"total_actual_remittance,omitempty"`
		TotalActualSupposedComparison float64    `json:"total_actual_supposed_comparison,omitempty"`
		Description                   string     `json:"description,omitempty"`
		CanView                       bool       `json:"can_view,omitempty"`
		IsClosed                      bool       `json:"is_closed,omitempty"`
		RequestView                   bool       `json:"request_view,omitempty"`

		EmployeeBySignatureMediaID    *uuid.UUID     `json:"employee_by_signature_media_id,omitempty"`
		EmployeeBySignatureMedia      *MediaResponse `json:"employee_by_signature_media,omitempty"`
		EmployeeByName                string         `json:"employee_by_name"`
		EmployeeByPosition            string         `json:"employee_by_position"`
		ApprovedBySignatureMediaID    *uuid.UUID     `json:"approved_by_signature_media_id,omitempty"`
		ApprovedByName                string         `json:"approved_by_name,omitempty"`
		ApprovedByPosition            string         `json:"approved_by_position,omitempty"`
		PreparedBySignatureMediaID    *uuid.UUID     `json:"prepared_by_signature_media_id,omitempty"`
		PreparedByName                string         `json:"prepared_by_name,omitempty"`
		PreparedByPosition            string         `json:"prepared_by_position,omitempty"`
		CertifiedBySignatureMediaID   *uuid.UUID     `json:"certified_by_signature_media_id,omitempty"`
		CertifiedByName               string         `json:"certified_by_name,omitempty"`
		CertifiedByPosition           string         `json:"certified_by_position,omitempty"`
		VerifiedBySignatureMediaID    *uuid.UUID     `json:"verified_by_signature_media_id,omitempty"`
		VerifiedByName                string         `json:"verified_by_name,omitempty"`
		VerifiedByPosition            string         `json:"verified_by_position,omitempty"`
		CheckBySignatureMediaID       *uuid.UUID     `json:"check_by_signature_media_id,omitempty"`
		CheckByName                   string         `json:"check_by_name,omitempty"`
		CheckByPosition               string         `json:"check_by_position,omitempty"`
		AcknowledgeBySignatureMediaID *uuid.UUID     `json:"acknowledge_by_signature_media_id,omitempty"`
		AcknowledgeByName             string         `json:"acknowledge_by_name,omitempty"`
		AcknowledgeByPosition         string         `json:"acknowledge_by_position,omitempty"`
		NotedBySignatureMediaID       *uuid.UUID     `json:"noted_by_signature_media_id,omitempty"`
		NotedByName                   string         `json:"noted_by_name,omitempty"`
		NotedByPosition               string         `json:"noted_by_position,omitempty"`
		PostedBySignatureMediaID      *uuid.UUID     `json:"posted_by_signature_media_id,omitempty"`
		PostedByName                  string         `json:"posted_by_name,omitempty"`
		PostedByPosition              string         `json:"posted_by_position,omitempty"`
		PaidBySignatureMediaID        *uuid.UUID     `json:"paid_by_signature_media_id,omitempty"`
		PaidByName                    string         `json:"paid_by_name,omitempty"`
		PaidByPosition                string         `json:"paid_by_position,omitempty"`
		CurrencyID                    uuid.UUID      `json:"currency_id,omitempty"`
		EndedAt                       *time.Time     `json:"ended_at,omitempty"`
	}

	TransactionBatchSignatureRequest struct {
		EmployeeBySignatureMediaID *uuid.UUID `json:"employee_by_signature_media_id,omitempty"`
		EmployeeByName             string     `json:"employee_by_name,omitempty"`
		EmployeeByPosition         string     `json:"employee_by_position,omitempty"`

		ApprovedBySignatureMediaID *uuid.UUID `json:"approved_by_signature_media_id,omitempty"`
		ApprovedByName             string     `json:"approved_by_name,omitempty"`
		ApprovedByPosition         string     `json:"approved_by_position,omitempty"`

		PreparedBySignatureMediaID *uuid.UUID `json:"prepared_by_signature_media_id,omitempty"`
		PreparedByName             string     `json:"prepared_by_name,omitempty"`
		PreparedByPosition         string     `json:"prepared_by_position,omitempty"`

		CertifiedBySignatureMediaID *uuid.UUID `json:"certified_by_signature_media_id,omitempty"`
		CertifiedByName             string     `json:"certified_by_name,omitempty"`
		CertifiedByPosition         string     `json:"certified_by_position,omitempty"`

		VerifiedBySignatureMediaID *uuid.UUID `json:"verified_by_signature_media_id,omitempty"`
		VerifiedByName             string     `json:"verified_by_name,omitempty"`
		VerifiedByPosition         string     `json:"verified_by_position,omitempty"`

		CheckBySignatureMediaID *uuid.UUID `json:"check_by_signature_media_id,omitempty"`
		CheckByName             string     `json:"check_by_name,omitempty"`
		CheckByPosition         string     `json:"check_by_position,omitempty"`

		AcknowledgeBySignatureMediaID *uuid.UUID `json:"acknowledge_by_signature_media_id,omitempty"`
		AcknowledgeByName             string     `json:"acknowledge_by_name,omitempty"`
		AcknowledgeByPosition         string     `json:"acknowledge_by_position,omitempty"`

		NotedBySignatureMediaID *uuid.UUID `json:"noted_by_signature_media_id,omitempty"`
		NotedByName             string     `json:"noted_by_name,omitempty"`
		NotedByPosition         string     `json:"noted_by_position,omitempty"`

		PostedBySignatureMediaID *uuid.UUID `json:"posted_by_signature_media_id,omitempty"`
		PostedByName             string     `json:"posted_by_name,omitempty"`
		PostedByPosition         string     `json:"posted_by_position,omitempty"`

		PaidBySignatureMediaID *uuid.UUID `json:"paid_by_signature_media_id,omitempty"`
		PaidByName             string     `json:"paid_by_name,omitempty"`
		PaidByPosition         string     `json:"paid_by_position,omitempty"`
	}
	TransactionBatchEndRequest struct {
		EmployeeBySignatureMediaID *uuid.UUID `json:"employee_by_signature_media_id,omitempty"`
		EmployeeByName             string     `json:"employee_by_name"`
		EmployeeByPosition         string     `json:"employee_by_position"`
	}

	TransactionBatchDepositInBankRequest struct {
		DepositInBank float64 `json:"deposit_in_bank" validate:"min=0"`
	}
)
