package model

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ENUMS
type LoanModeOfPayment string
type Weekdays string
type LoanCollectorPlace string
type LoanComakerType string
type LoanType string
type LoanAmortizationType string

const (
	LoanModeOfPaymentDaily       LoanModeOfPayment = "daily"
	LoanModeOfPaymentWeekly      LoanModeOfPayment = "weekly"
	LoanModeOfPaymentSemiMonthly LoanModeOfPayment = "semi-monthly"
	LoanModeOfPaymentMonthly     LoanModeOfPayment = "monthly"
	LoanModeOfPaymentQuarterly   LoanModeOfPayment = "quarterly"
	LoanModeOfPaymentSemiAnnual  LoanModeOfPayment = "semi-annual"
	LoanModeOfPaymentLumpsum     LoanModeOfPayment = "lumpsum"

	WeekdayMonday    Weekdays = "monday"
	WeekdayTuesday   Weekdays = "tuesday"
	WeekdayWednesday Weekdays = "wednesday"
	WeekdayThursday  Weekdays = "thursday"
	WeekdayFriday    Weekdays = "friday"
	WeekdaySaturday  Weekdays = "saturday"
	WeekdaySunday    Weekdays = "sunday"

	LoanCollectorPlaceOffice LoanCollectorPlace = "office"
	LoanCollectorPlaceField  LoanCollectorPlace = "field"

	LoanComakerTypeMember  LoanComakerType = "member"
	LoanComakerTypeDeposit LoanComakerType = "deposit"
	LoanComakerTypeOthers  LoanComakerType = "others"

	LoanTypeStandard             LoanType = "standard"
	LoanTypeRestructured         LoanType = "restructured"
	LoanTypeStandardPrevious     LoanType = "standard previous"
	LoanTypeRenewal              LoanType = "renewal"
	LoanTypeRenewalWithoutDeduct LoanType = "renewal without deduction"

	LoanAmortizationTypeSuggested LoanAmortizationType = "suggested"
	LoanAmortizationTypeNone      LoanAmortizationType = "none"
)

// MODEL
type (
	LoanTransaction struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_transaction"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_transaction"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		EmployeeUserID *uuid.UUID `gorm:"type:uuid"`
		EmployeeUser   *User      `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`

		TransactionBatchID    *uuid.UUID        `gorm:"type:uuid"`
		TransactionBatch      *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`
		OfficialReceiptNumber string            `gorm:"type:varchar(255)"`
		Voucher               string            `gorm:"type:varchar(255)"`

		LoanPurposeID *uuid.UUID   `gorm:"type:uuid"`
		LoanPurpose   *LoanPurpose `gorm:"foreignKey:LoanPurposeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_purpose,omitempty"`

		LoanStatusID *uuid.UUID  `gorm:"type:uuid"`
		LoanStatus   *LoanStatus `gorm:"foreignKey:LoanStatusID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_status,omitempty"`

		ModeOfPayment                string `gorm:"type:varchar(255)"`
		ModeOfPaymentWeekly          string `gorm:"type:varchar(255)"`
		ModeOfPaymentSemiMonthlyPay1 int    `gorm:"type:int"`
		ModeOfPaymentSemiMonthlyPay2 int    `gorm:"type:int"`

		ComakerType                            string                  `gorm:"type:varchar(255)"`
		ComakerDepositMemberAccountingLedgerID *uuid.UUID              `gorm:"type:uuid"`
		ComakerDepositMemberAccountingLedger   *MemberAccountingLedger `gorm:"foreignKey:ComakerDepositMemberAccountingLedgerID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"comaker_deposit_member_accounting_ledger,omitempty"`
		ComakerCollateralID                    *uuid.UUID              `gorm:"type:uuid"`
		ComakerCollateral                      *Collateral             `gorm:"foreignKey:ComakerCollateralID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"comaker_collateral,omitempty"`
		ComakerCollateralDescription           string                  `gorm:"type:text"`

		CollectorPlace string `gorm:"type:varchar(255);default:'office'"`

		LoanType       string           `gorm:"type:varchar(255);default:'standard'"`
		PreviousLoanID *uuid.UUID       `gorm:"type:uuid"`
		PreviousLoan   *LoanTransaction `gorm:"foreignKey:PreviousLoanID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"previous_loan,omitempty"`
		Terms          int              `gorm:"not null"`

		AmortizationAmount float64 `gorm:"type:decimal"`
		IsAddOn            bool    `gorm:"type:boolean"`

		Applied1 float64 `gorm:"type:decimal;not null"`
		Applied2 float64 `gorm:"type:decimal"`

		AccountID            *uuid.UUID     `gorm:"type:uuid"`
		Account              *Account       `gorm:"foreignKey:AccountID;constraint:OnDelete:SET NULL;" json:"account,omitempty"`
		MemberProfileID      *uuid.UUID     `gorm:"type:uuid"`
		MemberProfile        *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:SET NULL;" json:"member_profile,omitempty"`
		MemberJointAccountID *uuid.UUID     `gorm:"type:uuid"`
		MemberJointAccount   *Account       `gorm:"foreignKey:MemberJointAccountID;constraint:OnDelete:SET NULL;" json:"member_joint_account,omitempty"`
		SignatureMediaID     *uuid.UUID     `gorm:"type:uuid"`
		SignatureMedia       *Media         `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:SET NULL;" json:"signature_media,omitempty"`

		MountToBeClosed float64 `gorm:"type:decimal"`
		DamayanFund     float64 `gorm:"type:decimal"`
		ShareCapital    float64 `gorm:"type:decimal"`
		LengthOfService string  `gorm:"type:varchar(255)"`

		ExcludeSunday   bool `gorm:"type:boolean;default:false"`
		ExcludeHoliday  bool `gorm:"type:boolean;default:false"`
		ExcludeSaturday bool `gorm:"type:boolean;default:false"`

		RemarksOtherTerms                string `gorm:"type:text"`
		RemarksPayrollDeduction          bool   `gorm:"type:boolean;default:false"`
		RecordOfLoanPaymentsOrLoanStatus string `gorm:"type:varchar(255)"`
		CollateralOffered                string `gorm:"type:text"`

		AppraisedValue            float64 `gorm:"type:decimal"`
		AppraisedValueDescription string  `gorm:"type:text"`

		PrintedDate  *time.Time `gorm:"type:timestamp"`
		ApprovedDate *time.Time `gorm:"type:timestamp"`
		ReleasedDate *time.Time `gorm:"type:timestamp"`

		ApprovedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		ApprovedBySignatureMedia   *Media     `gorm:"foreignKey:ApprovedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"approved_by_signature_media,omitempty"`
		ApprovedByName             string     `gorm:"type:varchar(255)"`
		ApprovedByPosition         string     `gorm:"type:varchar(255)"`

		PreparedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		PreparedBySignatureMedia   *Media     `gorm:"foreignKey:PreparedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"prepared_by_signature_media,omitempty"`
		PreparedByName             string     `gorm:"type:varchar(255)"`
		PreparedByPosition         string     `gorm:"type:varchar(255)"`

		CertifiedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		CertifiedBySignatureMedia   *Media     `gorm:"foreignKey:CertifiedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"certified_by_signature_media,omitempty"`
		CertifiedByName             string     `gorm:"type:varchar(255)"`
		CertifiedByPosition         string     `gorm:"type:varchar(255)"`

		VerifiedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		VerifiedBySignatureMedia   *Media     `gorm:"foreignKey:VerifiedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"verified_by_signature_media,omitempty"`
		VerifiedByName             string     `gorm:"type:varchar(255)"`
		VerifiedByPosition         string     `gorm:"type:varchar(255)"`

		CheckBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		CheckBySignatureMedia   *Media     `gorm:"foreignKey:CheckBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"check_by_signature_media,omitempty"`
		CheckByName             string     `gorm:"type:varchar(255)"`
		CheckByPosition         string     `gorm:"type:varchar(255)"`

		AcknowledgeBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		AcknowledgeBySignatureMedia   *Media     `gorm:"foreignKey:AcknowledgeBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"acknowledge_by_signature_media,omitempty"`
		AcknowledgeByName             string     `gorm:"type:varchar(255)"`
		AcknowledgeByPosition         string     `gorm:"type:varchar(255)"`

		NotedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		NotedBySignatureMedia   *Media     `gorm:"foreignKey:NotedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"noted_by_signature_media,omitempty"`
		NotedByName             string     `gorm:"type:varchar(255)"`
		NotedByPosition         string     `gorm:"type:varchar(255)"`

		PostedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		PostedBySignatureMedia   *Media     `gorm:"foreignKey:PostedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"posted_by_signature_media,omitempty"`
		PostedByName             string     `gorm:"type:varchar(255)"`
		PostedByPosition         string     `gorm:"type:varchar(255)"`

		PaidBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		PaidBySignatureMedia   *Media     `gorm:"foreignKey:PaidBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"paid_by_signature_media,omitempty"`
		PaidByName             string     `gorm:"type:varchar(255)"`
		PaidByPosition         string     `gorm:"type:varchar(255)"`

		// Relationships
		LoanTransactionEntries []*LoanTransactionEntry `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"loan_transaction_entries,omitempty"`
	}

	LoanTransactionResponse struct {
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

		EmployeeUserID *uuid.UUID    `json:"employee_user_id,omitempty"`
		EmployeeUser   *UserResponse `json:"employee_user,omitempty"`

		TransactionBatchID    *uuid.UUID                `json:"transaction_batch_id,omitempty"`
		TransactionBatch      *TransactionBatchResponse `json:"transaction_batch,omitempty"`
		OfficialReceiptNumber string                    `json:"official_receipt_number"`
		Voucher               string                    `json:"voucher"`

		LoanPurposeID *uuid.UUID           `json:"loan_purpose_id,omitempty"`
		LoanPurpose   *LoanPurposeResponse `json:"loan_purpose,omitempty"`

		LoanStatusID *uuid.UUID          `json:"loan_status_id,omitempty"`
		LoanStatus   *LoanStatusResponse `json:"loan_status,omitempty"`

		ModeOfPayment                LoanModeOfPayment `json:"mode_of_payment"`
		ModeOfPaymentWeekly          Weekdays          `json:"mode_of_payment_weekly"`
		ModeOfPaymentSemiMonthlyPay1 int               `json:"mode_of_payment_semi_monthly_pay_1"`
		ModeOfPaymentSemiMonthlyPay2 int               `json:"mode_of_payment_semi_monthly_pay_2"`

		ComakerType                            LoanComakerType                 `json:"comaker_type"`
		ComakerDepositMemberAccountingLedgerID *uuid.UUID                      `json:"comaker_deposit_member_accounting_ledger_id,omitempty"`
		ComakerDepositMemberAccountingLedger   *MemberAccountingLedgerResponse `json:"comaker_deposit_member_accounting_ledger,omitempty"`
		ComakerCollateralID                    *uuid.UUID                      `json:"comaker_collateral_id,omitempty"`
		ComakerCollateral                      *CollateralResponse             `json:"comaker_collateral,omitempty"`
		ComakerCollateralDescription           string                          `json:"comaker_collateral_description"`

		CollectorPlace LoanCollectorPlace `json:"collector_place"`

		LoanType       LoanType                 `json:"loan_type"`
		PreviousLoanID *uuid.UUID               `json:"previous_loan_id,omitempty"`
		PreviousLoan   *LoanTransactionResponse `json:"previous_loan,omitempty"`
		Terms          int                      `json:"terms"`

		AmortizationAmount float64 `json:"amortization_amount"`
		IsAddOn            bool    `json:"is_add_on"`

		Applied1 float64 `json:"applied_1"`
		Applied2 float64 `json:"applied_2"`

		AccountID            *uuid.UUID             `json:"account_id,omitempty"`
		Account              *AccountResponse       `json:"account,omitempty"`
		MemberProfileID      *uuid.UUID             `json:"member_profile_id,omitempty"`
		MemberProfile        *MemberProfileResponse `json:"member_profile,omitempty"`
		MemberJointAccountID *uuid.UUID             `json:"member_joint_account_id,omitempty"`
		MemberJointAccount   *AccountResponse       `json:"member_joint_account,omitempty"`
		SignatureMediaID     *uuid.UUID             `json:"signature_media_id,omitempty"`
		SignatureMedia       *MediaResponse         `json:"signature_media,omitempty"`

		MountToBeClosed float64 `json:"mount_to_be_closed"`
		DamayanFund     float64 `json:"damayan_fund"`
		ShareCapital    float64 `json:"share_capital"`
		LengthOfService string  `json:"length_of_service"`

		ExcludeSunday   bool `json:"exclude_sunday"`
		ExcludeHoliday  bool `json:"exclude_holiday"`
		ExcludeSaturday bool `json:"exclude_saturday"`

		RemarksOtherTerms                string `json:"remarks_other_terms"`
		RemarksPayrollDeduction          bool   `json:"remarks_payroll_deduction"`
		RecordOfLoanPaymentsOrLoanStatus string `json:"record_of_loan_payments_or_loan_status"`
		CollateralOffered                string `json:"collateral_offered"`

		AppraisedValue            float64 `json:"appraised_value"`
		AppraisedValueDescription string  `json:"appraised_value_description"`

		PrintedDate  *time.Time `json:"printed_date,omitempty"`
		ApprovedDate *time.Time `json:"approved_date,omitempty"`
		ReleasedDate *time.Time `json:"released_date,omitempty"`

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

		// Relationships
		LoanTransactionEntries []*LoanTransactionEntryResponse `json:"loan_transaction_entries,omitempty"`
	}

	LoanTransactionRequest struct {
		TransactionBatchID    *uuid.UUID `json:"transaction_batch_id,omitempty"`
		OfficialReceiptNumber string     `json:"official_receipt_number,omitempty"`
		Voucher               string     `json:"voucher,omitempty"`
		LoanPurposeID         *uuid.UUID `json:"loan_purpose_id,omitempty"`
		LoanStatusID          *uuid.UUID `json:"loan_status_id,omitempty"`

		ModeOfPayment                LoanModeOfPayment `json:"mode_of_payment"`
		ModeOfPaymentWeekly          Weekdays          `json:"mode_of_payment_weekly,omitempty"`
		ModeOfPaymentSemiMonthlyPay1 int               `json:"mode_of_payment_semi_monthly_pay_1,omitempty"`
		ModeOfPaymentSemiMonthlyPay2 int               `json:"mode_of_payment_semi_monthly_pay_2,omitempty"`

		ComakerType                            LoanComakerType `json:"comaker_type"`
		ComakerDepositMemberAccountingLedgerID *uuid.UUID      `json:"comaker_deposit_member_accounting_ledger_id,omitempty"`
		ComakerCollateralID                    *uuid.UUID      `json:"comaker_collateral_id,omitempty"`
		ComakerCollateralDescription           string          `json:"comaker_collateral_description,omitempty"`

		CollectorPlace LoanCollectorPlace `json:"collector_place"`

		LoanType       LoanType   `json:"loan_type"`
		PreviousLoanID *uuid.UUID `json:"previous_loan_id,omitempty"`
		Terms          int        `json:"terms"`

		AmortizationAmount float64 `json:"amortization_amount,omitempty"`
		IsAddOn            bool    `json:"is_add_on,omitempty"`

		Applied1 float64 `json:"applied_1"`
		Applied2 float64 `json:"applied_2,omitempty"`

		AccountID            *uuid.UUID `json:"account_id,omitempty"`
		MemberProfileID      *uuid.UUID `json:"member_profile_id,omitempty"`
		MemberJointAccountID *uuid.UUID `json:"member_joint_account_id,omitempty"`
		SignatureMediaID     *uuid.UUID `json:"signature_media_id,omitempty"`

		MountToBeClosed float64 `json:"mount_to_be_closed,omitempty"`
		DamayanFund     float64 `json:"damayan_fund,omitempty"`
		ShareCapital    float64 `json:"share_capital,omitempty"`
		LengthOfService string  `json:"length_of_service,omitempty"`

		ExcludeSunday   bool `json:"exclude_sunday,omitempty"`
		ExcludeHoliday  bool `json:"exclude_holiday,omitempty"`
		ExcludeSaturday bool `json:"exclude_saturday,omitempty"`

		RemarksOtherTerms                string `json:"remarks_other_terms,omitempty"`
		RemarksPayrollDeduction          bool   `json:"remarks_payroll_deduction,omitempty"`
		RecordOfLoanPaymentsOrLoanStatus string `json:"record_of_loan_payments_or_loan_status,omitempty"`
		CollateralOffered                string `json:"collateral_offered,omitempty"`

		AppraisedValue            float64 `json:"appraised_value,omitempty"`
		AppraisedValueDescription string  `json:"appraised_value_description,omitempty"`

		PrintedDate  *time.Time `json:"printed_date,omitempty"`
		ApprovedDate *time.Time `json:"approved_date,omitempty"`
		ReleasedDate *time.Time `json:"released_date,omitempty"`

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
)

func (m *Model) LoanTransaction() {
	m.Migration = append(m.Migration, &LoanTransaction{})
	m.LoanTransactionManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		LoanTransaction, LoanTransactionResponse, LoanTransactionRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization", "EmployeeUser",
			"TransactionBatch", "LoanPurpose", "LoanStatus",
			"ComakerDepositMemberAccountingLedger", "ComakerCollateral", "PreviousLoan",
			"Account", "MemberProfile", "MemberJointAccount", "SignatureMedia",
			"ApprovedBySignatureMedia", "PreparedBySignatureMedia", "CertifiedBySignatureMedia",
			"VerifiedBySignatureMedia", "CheckBySignatureMedia", "AcknowledgeBySignatureMedia",
			"NotedBySignatureMedia", "PostedBySignatureMedia", "PaidBySignatureMedia",
			"LoanTransactionEntries",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanTransaction) *LoanTransactionResponse {
			if data == nil {
				return nil
			}
			return &LoanTransactionResponse{
				ID:                                     data.ID,
				CreatedAt:                              data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                            data.CreatedByID,
				CreatedBy:                              m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                              data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                            data.UpdatedByID,
				UpdatedBy:                              m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:                         data.OrganizationID,
				Organization:                           m.OrganizationManager.ToModel(data.Organization),
				BranchID:                               data.BranchID,
				Branch:                                 m.BranchManager.ToModel(data.Branch),
				EmployeeUserID:                         data.EmployeeUserID,
				EmployeeUser:                           m.UserManager.ToModel(data.EmployeeUser),
				TransactionBatchID:                     data.TransactionBatchID,
				TransactionBatch:                       m.TransactionBatchManager.ToModel(data.TransactionBatch),
				OfficialReceiptNumber:                  data.OfficialReceiptNumber,
				Voucher:                                data.Voucher,
				LoanPurposeID:                          data.LoanPurposeID,
				LoanPurpose:                            m.LoanPurposeManager.ToModel(data.LoanPurpose),
				LoanStatusID:                           data.LoanStatusID,
				LoanStatus:                             m.LoanStatusManager.ToModel(data.LoanStatus),
				ModeOfPayment:                          LoanModeOfPayment(data.ModeOfPayment),
				ModeOfPaymentWeekly:                    Weekdays(data.ModeOfPaymentWeekly),
				ModeOfPaymentSemiMonthlyPay1:           data.ModeOfPaymentSemiMonthlyPay1,
				ModeOfPaymentSemiMonthlyPay2:           data.ModeOfPaymentSemiMonthlyPay2,
				ComakerType:                            LoanComakerType(data.ComakerType),
				ComakerDepositMemberAccountingLedgerID: data.ComakerDepositMemberAccountingLedgerID,
				ComakerDepositMemberAccountingLedger:   m.MemberAccountingLedgerManager.ToModel(data.ComakerDepositMemberAccountingLedger),
				ComakerCollateralID:                    data.ComakerCollateralID,
				ComakerCollateral:                      m.CollateralManager.ToModel(data.ComakerCollateral),
				ComakerCollateralDescription:           data.ComakerCollateralDescription,
				CollectorPlace:                         LoanCollectorPlace(data.CollectorPlace),
				LoanType:                               LoanType(data.LoanType),
				PreviousLoanID:                         data.PreviousLoanID,
				PreviousLoan:                           m.LoanTransactionManager.ToModel(data.PreviousLoan),
				Terms:                                  data.Terms,
				AmortizationAmount:                     data.AmortizationAmount,
				IsAddOn:                                data.IsAddOn,
				Applied1:                               data.Applied1,
				Applied2:                               data.Applied2,
				AccountID:                              data.AccountID,
				Account:                                m.AccountManager.ToModel(data.Account),
				MemberProfileID:                        data.MemberProfileID,
				MemberProfile:                          m.MemberProfileManager.ToModel(data.MemberProfile),
				MemberJointAccountID:                   data.MemberJointAccountID,
				MemberJointAccount:                     m.AccountManager.ToModel(data.MemberJointAccount),
				SignatureMediaID:                       data.SignatureMediaID,
				SignatureMedia:                         m.MediaManager.ToModel(data.SignatureMedia),
				MountToBeClosed:                        data.MountToBeClosed,
				DamayanFund:                            data.DamayanFund,
				ShareCapital:                           data.ShareCapital,
				LengthOfService:                        data.LengthOfService,
				ExcludeSunday:                          data.ExcludeSunday,
				ExcludeHoliday:                         data.ExcludeHoliday,
				ExcludeSaturday:                        data.ExcludeSaturday,
				RemarksOtherTerms:                      data.RemarksOtherTerms,
				RemarksPayrollDeduction:                data.RemarksPayrollDeduction,
				RecordOfLoanPaymentsOrLoanStatus:       data.RecordOfLoanPaymentsOrLoanStatus,
				CollateralOffered:                      data.CollateralOffered,
				AppraisedValue:                         data.AppraisedValue,
				AppraisedValueDescription:              data.AppraisedValueDescription,
				PrintedDate:                            data.PrintedDate,
				ApprovedDate:                           data.ApprovedDate,
				ReleasedDate:                           data.ReleasedDate,
				ApprovedBySignatureMediaID:             data.ApprovedBySignatureMediaID,
				ApprovedBySignatureMedia:               m.MediaManager.ToModel(data.ApprovedBySignatureMedia),
				ApprovedByName:                         data.ApprovedByName,
				ApprovedByPosition:                     data.ApprovedByPosition,
				PreparedBySignatureMediaID:             data.PreparedBySignatureMediaID,
				PreparedBySignatureMedia:               m.MediaManager.ToModel(data.PreparedBySignatureMedia),
				PreparedByName:                         data.PreparedByName,
				PreparedByPosition:                     data.PreparedByPosition,
				CertifiedBySignatureMediaID:            data.CertifiedBySignatureMediaID,
				CertifiedBySignatureMedia:              m.MediaManager.ToModel(data.CertifiedBySignatureMedia),
				CertifiedByName:                        data.CertifiedByName,
				CertifiedByPosition:                    data.CertifiedByPosition,
				VerifiedBySignatureMediaID:             data.VerifiedBySignatureMediaID,
				VerifiedBySignatureMedia:               m.MediaManager.ToModel(data.VerifiedBySignatureMedia),
				VerifiedByName:                         data.VerifiedByName,
				VerifiedByPosition:                     data.VerifiedByPosition,
				CheckBySignatureMediaID:                data.CheckBySignatureMediaID,
				CheckBySignatureMedia:                  m.MediaManager.ToModel(data.CheckBySignatureMedia),
				CheckByName:                            data.CheckByName,
				CheckByPosition:                        data.CheckByPosition,
				AcknowledgeBySignatureMediaID:          data.AcknowledgeBySignatureMediaID,
				AcknowledgeBySignatureMedia:            m.MediaManager.ToModel(data.AcknowledgeBySignatureMedia),
				AcknowledgeByName:                      data.AcknowledgeByName,
				AcknowledgeByPosition:                  data.AcknowledgeByPosition,
				NotedBySignatureMediaID:                data.NotedBySignatureMediaID,
				NotedBySignatureMedia:                  m.MediaManager.ToModel(data.NotedBySignatureMedia),
				NotedByName:                            data.NotedByName,
				NotedByPosition:                        data.NotedByPosition,
				PostedBySignatureMediaID:               data.PostedBySignatureMediaID,
				PostedBySignatureMedia:                 m.MediaManager.ToModel(data.PostedBySignatureMedia),
				PostedByName:                           data.PostedByName,
				PostedByPosition:                       data.PostedByPosition,
				PaidBySignatureMediaID:                 data.PaidBySignatureMediaID,
				PaidBySignatureMedia:                   m.MediaManager.ToModel(data.PaidBySignatureMedia),
				PaidByName:                             data.PaidByName,
				PaidByPosition:                         data.PaidByPosition,
				LoanTransactionEntries:                 m.mapLoanTransactionEntries(data.LoanTransactionEntries),
			}
		},

		Created: func(data *LoanTransaction) []string {
			return []string{
				"loan_transaction.create",
				fmt.Sprintf("loan_transaction.create.%s", data.ID),
				fmt.Sprintf("loan_transaction.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanTransaction) []string {
			return []string{
				"loan_transaction.update",
				fmt.Sprintf("loan_transaction.update.%s", data.ID),
				fmt.Sprintf("loan_transaction.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanTransaction) []string {
			return []string{
				"loan_transaction.delete",
				fmt.Sprintf("loan_transaction.delete.%s", data.ID),
				fmt.Sprintf("loan_transaction.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) LoanTransactionCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*LoanTransaction, error) {
	return m.LoanTransactionManager.Find(context, &LoanTransaction{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}

// Helper function to map loan transaction entries
func (m *Model) mapLoanTransactionEntries(entries []*LoanTransactionEntry) []*LoanTransactionEntryResponse {
	if entries == nil {
		return nil
	}

	var result []*LoanTransactionEntryResponse
	for _, entry := range entries {
		if entry != nil {
			result = append(result, m.LoanTransactionEntryManager.ToModel(entry))
		}
	}
	return result
}
