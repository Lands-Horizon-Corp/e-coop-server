package core

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoanModeOfPayment string

type Weekdays string

type LoanCollectorPlace string

type LoanComakerType string

type LoanType string

type LoanAmortizationType string

type LoanAdjustmentType string

const (
	LoanModeOfPaymentDaily       LoanModeOfPayment = "daily"
	LoanModeOfPaymentWeekly      LoanModeOfPayment = "weekly"
	LoanModeOfPaymentSemiMonthly LoanModeOfPayment = "semi-monthly"
	LoanModeOfPaymentMonthly     LoanModeOfPayment = "monthly"
	LoanModeOfPaymentQuarterly   LoanModeOfPayment = "quarterly"
	LoanModeOfPaymentSemiAnnual  LoanModeOfPayment = "semi-annual"
	LoanModeOfPaymentLumpsum     LoanModeOfPayment = "lumpsum"
	LoanModeOfPaymentFixedDays   LoanModeOfPayment = "fixed-days"

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

const (
	LoanAdjustmentTypeDeduct   LoanAdjustmentType = "deduct"
	LoanAdjustmentTypeAdd      LoanAdjustmentType = "add"
	LoanAdjustmentTypeAdjusted LoanAdjustmentType = "adjusted"
)

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
		CheckNumber           string            `gorm:"type:varchar(255)"`
		CheckDate             *time.Time        `gorm:"type:timestamp"`
		Voucher               string            `gorm:"type:varchar(255)"`

		LoanPurposeID *uuid.UUID   `gorm:"type:uuid"`
		LoanPurpose   *LoanPurpose `gorm:"foreignKey:LoanPurposeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_purpose,omitempty"`

		LoanStatusID *uuid.UUID  `gorm:"type:uuid"`
		LoanStatus   *LoanStatus `gorm:"foreignKey:LoanStatusID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_status,omitempty"`

		ModeOfPayment                LoanModeOfPayment `gorm:"type:varchar(255)"`
		ModeOfPaymentWeekly          Weekdays          `gorm:"type:varchar(255)"`
		ModeOfPaymentSemiMonthlyPay1 int               `gorm:"type:int"`
		ModeOfPaymentSemiMonthlyPay2 int               `gorm:"type:int"`
		ModeOfPaymentFixedDays       int               `gorm:"type:int;default:0" json:"mode_of_payment_fixed_days"`
		ModeOfPaymentMonthlyExactDay bool              `gorm:"type:boolean;default:false" json:"mode_of_payment_monthly_exact_day"`

		ComakerType                            LoanComakerType         `gorm:"type:varchar(255)"`
		ComakerDepositMemberAccountingLedgerID *uuid.UUID              `gorm:"type:uuid"`
		ComakerDepositMemberAccountingLedger   *MemberAccountingLedger `gorm:"foreignKey:ComakerDepositMemberAccountingLedgerID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"comaker_deposit_member_accounting_ledger,omitempty"`

		CollectorPlace LoanCollectorPlace `gorm:"type:varchar(255);default:'office'"`

		LoanType       LoanType         `gorm:"type:varchar(255);default:'standard'"`
		PreviousLoanID *uuid.UUID       `gorm:"type:uuid"`
		PreviousLoan   *LoanTransaction `gorm:"foreignKey:PreviousLoanID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"previous_loan,omitempty"`
		Terms          int              `gorm:"not null"`

		Amortization float64 `gorm:"type:decimal"`
		IsAddOn      bool    `gorm:"type:boolean"`

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
		PrintNumber  int        `gorm:"type:int;default:0"`

		ReleasedByID *uuid.UUID `gorm:"type:uuid"`
		ReleasedBy   *User      `gorm:"foreignKey:ReleasedByID;constraint:OnDelete:SET NULL;" json:"released_by,omitempty"`
		PrintedByID  *uuid.UUID `gorm:"type:uuid"`
		PrintedBy    *User      `gorm:"foreignKey:PrintedByID;constraint:OnDelete:SET NULL;" json:"printed_by,omitempty"`
		ApprovedByID *uuid.UUID `gorm:"type:uuid"`
		ApprovedBy   *User      `gorm:"foreignKey:ApprovedByID;constraint:OnDelete:SET NULL;" json:"approved_by,omitempty"`

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

		LoanTags                              []*LoanTag                               `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"loan_tags,omitempty"`
		LoanTransactionEntries                []*LoanTransactionEntry                  `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"loan_transaction_entries,omitempty"`
		LoanClearanceAnalysis                 []*LoanClearanceAnalysis                 `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"loan_clearance_analysis,omitempty"`
		LoanClearanceAnalysisInstitution      []*LoanClearanceAnalysisInstitution      `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"loan_clearance_analysis_institution,omitempty"`
		LoanTermsAndConditionSuggestedPayment []*LoanTermsAndConditionSuggestedPayment `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"loan_terms_and_condition_suggested_payment,omitempty"`
		LoanTermsAndConditionAmountReceipt    []*LoanTermsAndConditionAmountReceipt    `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"loan_terms_and_condition_amount_receipt,omitempty"`
		ComakerMemberProfiles                 []*ComakerMemberProfile                  `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"comaker_member_profiles,omitempty"`
		ComakerCollaterals                    []*ComakerCollateral                     `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"comaker_collaterals,omitempty"`
		LoanAccounts                          []*LoanAccount                           `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"loan_accounts,omitempty"`

		Count          int        `gorm:"type:int;default:0" json:"count"`
		Balance        float64    `gorm:"type:decimal;default:0" json:"balance"`
		LastPay        *time.Time `gorm:"type:timestamp" json:"last_pay,omitempty"`
		Fines          float64    `gorm:"type:decimal;default:0" json:"fines"`
		Interest       float64    `gorm:"type:decimal;default:0" json:"interest"`
		TotalDebit     float64    `gorm:"total_debit;type:decimal;default:0" json:"total_debit"`
		TotalCredit    float64    `gorm:"total_credit;type:decimal;default:0" json:"total_credit"`
		TotalPrincipal float64    `gorm:"total_principal;type:decimal;default:0" json:"total_principal"`
		TotalAddOn     float64    `gorm:"total_add_on;type:decimal;default:0" json:"total_add_on"`
		AmountGranted  float64    `gorm:"amount_granted;type:decimal;default:0" json:"amount_granted"`
		Processing     bool       `gorm:"default:false" json:"processing"`
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
		CheckDate             *time.Time                `json:"check_date,omitempty"`
		CheckNumber           string                    `json:"check_number"`

		LoanPurposeID *uuid.UUID           `json:"loan_purpose_id,omitempty"`
		LoanPurpose   *LoanPurposeResponse `json:"loan_purpose,omitempty"`

		LoanStatusID *uuid.UUID          `json:"loan_status_id,omitempty"`
		LoanStatus   *LoanStatusResponse `json:"loan_status,omitempty"`

		ModeOfPayment                LoanModeOfPayment `json:"mode_of_payment"`
		ModeOfPaymentWeekly          Weekdays          `json:"mode_of_payment_weekly"`
		ModeOfPaymentSemiMonthlyPay1 int               `json:"mode_of_payment_semi_monthly_pay_1"`
		ModeOfPaymentSemiMonthlyPay2 int               `json:"mode_of_payment_semi_monthly_pay_2"`
		ModeOfPaymentFixedDays       int               `json:"mode_of_payment_fixed_days"`
		ModeOfPaymentMonthlyExactDay bool              `json:"mode_of_payment_monthly_exact_day"`

		ComakerType                            LoanComakerType                 `json:"comaker_type"`
		ComakerDepositMemberAccountingLedgerID *uuid.UUID                      `json:"comaker_deposit_member_accounting_ledger_id,omitempty"`
		ComakerDepositMemberAccountingLedger   *MemberAccountingLedgerResponse `json:"comaker_deposit_member_accounting_ledger,omitempty"`

		CollectorPlace LoanCollectorPlace `json:"collector_place"`

		LoanType       LoanType                 `json:"loan_type"`
		PreviousLoanID *uuid.UUID               `json:"previous_loan_id,omitempty"`
		PreviousLoan   *LoanTransactionResponse `json:"previous_loan,omitempty"`
		Terms          int                      `json:"terms"`

		Amortization float64 `json:"amortization"`
		IsAddOn      bool    `json:"is_add_on"`

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
		PrintNumber  int        `json:"print_number"`
		ApprovedDate *time.Time `json:"approved_date,omitempty"`
		ReleasedDate *time.Time `json:"released_date,omitempty"`

		ReleasedByID *uuid.UUID    `json:"released_by_id,omitempty"`
		ReleasedBy   *UserResponse `json:"released_by,omitempty"`
		PrintedByID  *uuid.UUID    `json:"printed_by_id,omitempty"`
		PrintedBy    *UserResponse `json:"printed_by,omitempty"`
		ApprovedByID *uuid.UUID    `json:"approved_by_id,omitempty"`
		ApprovedBy   *UserResponse `json:"approved_by,omitempty"`

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

		LoanTags                              []*LoanTagResponse                               `json:"loan_tags,omitempty"`
		LoanTransactionEntries                []*LoanTransactionEntryResponse                  `json:"loan_transaction_entries,omitempty"`
		LoanClearanceAnalysis                 []*LoanClearanceAnalysisResponse                 `json:"loan_clearance_analysis,omitempty"`
		LoanClearanceAnalysisInstitution      []*LoanClearanceAnalysisInstitutionResponse      `json:"loan_clearance_analysis_institution,omitempty"`
		LoanTermsAndConditionSuggestedPayment []*LoanTermsAndConditionSuggestedPaymentResponse `json:"loan_terms_and_condition_suggested_payment,omitempty"`
		LoanTermsAndConditionAmountReceipt    []*LoanTermsAndConditionAmountReceiptResponse    `json:"loan_terms_and_condition_amount_receipt,omitempty"`
		ComakerMemberProfiles                 []*ComakerMemberProfileResponse                  `json:"comaker_member_profiles,omitempty"`
		ComakerCollaterals                    []*ComakerCollateralResponse                     `json:"comaker_collaterals,omitempty"`
		LoanAccounts                          []*LoanAccountResponse                           `json:"loan_accounts,omitempty"`

		Count       int        `json:"count"`
		Balance     float64    `json:"balance"`
		LastPay     *time.Time `json:"last_pay,omitempty"`
		Fines       float64    `json:"fines"`
		Interest    float64    `json:"interest"`
		TotalDebit  float64    `json:"total_debit"`
		TotalCredit float64    `json:"total_credit"`

		Processing bool `json:"processing"`
	}

	LoanTransactionTotalResponse struct {
		Balance     float64 `json:"balance"`
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
	}

	LoanTransactionAdjustmentRequest struct {
		Voucher        string             `json:"voucher,omitempty"`
		LoanAccount    uuid.UUID          `json:"loan_account_id"`
		AdjustmentType LoanAdjustmentType `json:"adjustment_type"`
		Amount         float64            `json:"amount"`
	}
	LoanTransactionRequest struct {
		OfficialReceiptNumber string     `json:"official_receipt_number,omitempty"`
		Voucher               string     `json:"voucher,omitempty"`
		LoanPurposeID         *uuid.UUID `json:"loan_purpose_id,omitempty"`
		LoanStatusID          *uuid.UUID `json:"loan_status_id,omitempty"`

		ModeOfPayment                LoanModeOfPayment `json:"mode_of_payment"`
		ModeOfPaymentWeekly          Weekdays          `json:"mode_of_payment_weekly,omitempty"`
		ModeOfPaymentSemiMonthlyPay1 int               `json:"mode_of_payment_semi_monthly_pay_1,omitempty"`
		ModeOfPaymentSemiMonthlyPay2 int               `json:"mode_of_payment_semi_monthly_pay_2,omitempty"`
		ModeOfPaymentFixedDays       int               `json:"mode_of_payment_fixed_days,omitempty"`
		ModeOfPaymentMonthlyExactDay bool              `json:"mode_of_payment_monthly_exact_day,omitempty"`

		ComakerType                            LoanComakerType `json:"comaker_type"`
		ComakerDepositMemberAccountingLedgerID *uuid.UUID      `json:"comaker_deposit_member_accounting_ledger_id,omitempty"`

		CollectorPlace LoanCollectorPlace `json:"collector_place"`

		LoanType       LoanType   `json:"loan_type"`
		PreviousLoanID *uuid.UUID `json:"previous_loan_id,omitempty"`
		Terms          int        `json:"terms"`

		Amortization float64 `json:"amortization,omitempty"`
		IsAddOn      bool    `json:"is_add_on,omitempty"`

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
		PrintNumber  int        `json:"print_number,omitempty"`
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

		LoanTags []*LoanTagRequest `json:"loan_tags,omitempty"`

		LoanClearanceAnalysis                 []*LoanClearanceAnalysisRequest                 `json:"loan_clearance_analysis,omitempty"`
		LoanClearanceAnalysisInstitution      []*LoanClearanceAnalysisInstitutionRequest      `json:"loan_clearance_analysis_institution,omitempty"`
		LoanTermsAndConditionSuggestedPayment []*LoanTermsAndConditionSuggestedPaymentRequest `json:"loan_terms_and_condition_suggested_payment,omitempty"`
		LoanTermsAndConditionAmountReceipt    []*LoanTermsAndConditionAmountReceiptRequest    `json:"loan_terms_and_condition_amount_receipt,omitempty"`
		ComakerMemberProfiles                 []*ComakerMemberProfileRequest                  `json:"comaker_member_profiles,omitempty"`
		ComakerCollaterals                    []*ComakerCollateralRequest                     `json:"comaker_collaterals,omitempty"`

		LoanTagsDeleted uuid.UUIDs `json:"loan_tags_deleted,omitempty"`

		LoanClearanceAnalysisDeleted                 uuid.UUIDs `json:"loan_clearance_analysis_deleted,omitempty"`
		LoanClearanceAnalysisInstitutionDeleted      uuid.UUIDs `json:"loan_clearance_analysis_institution_deleted,omitempty"`
		LoanTermsAndConditionSuggestedPaymentDeleted uuid.UUIDs `json:"loan_terms_and_condition_suggested_payment_deleted,omitempty"`
		LoanTermsAndConditionAmountReceiptDeleted    uuid.UUIDs `json:"loan_terms_and_condition_amount_receipt_deleted,omitempty"`
		ComakerMemberProfilesDeleted                 uuid.UUIDs `json:"comaker_member_profiles_deleted,omitempty"`
		ComakerCollateralsDeleted                    uuid.UUIDs `json:"comaker_collaterals_deleted,omitempty"`
	}

	LoanTransactionPrintRequest struct {
		Voucher     string     `json:"voucher"`
		CheckNumber string     `json:"check_number"`
		CheckDate   *time.Time `json:"check_date"`
	}

	LoanTransactionSignatureRequest struct {
		ApprovedBySignatureMediaID *uuid.UUID `json:"approved_by_signature_media_id,omitempty" validate:"omitempty,uuid"`
		ApprovedByName             string     `json:"approved_by_name,omitempty" validate:"omitempty,max=255"`
		ApprovedByPosition         string     `json:"approved_by_position,omitempty" validate:"omitempty,max=255"`

		PreparedBySignatureMediaID *uuid.UUID `json:"prepared_by_signature_media_id,omitempty" validate:"omitempty,uuid"`
		PreparedByName             string     `json:"prepared_by_name,omitempty" validate:"omitempty,max=255"`
		PreparedByPosition         string     `json:"prepared_by_position,omitempty" validate:"omitempty,max=255"`

		CertifiedBySignatureMediaID *uuid.UUID `json:"certified_by_signature_media_id,omitempty" validate:"omitempty,uuid"`
		CertifiedByName             string     `json:"certified_by_name,omitempty" validate:"omitempty,max=255"`
		CertifiedByPosition         string     `json:"certified_by_position,omitempty" validate:"omitempty,max=255"`

		VerifiedBySignatureMediaID *uuid.UUID `json:"verified_by_signature_media_id,omitempty" validate:"omitempty,uuid"`
		VerifiedByName             string     `json:"verified_by_name,omitempty" validate:"omitempty,max=255"`
		VerifiedByPosition         string     `json:"verified_by_position,omitempty" validate:"omitempty,max=255"`

		CheckBySignatureMediaID *uuid.UUID `json:"check_by_signature_media_id,omitempty" validate:"omitempty,uuid"`
		CheckByName             string     `json:"check_by_name,omitempty" validate:"omitempty,max=255"`
		CheckByPosition         string     `json:"check_by_position,omitempty" validate:"omitempty,max=255"`

		AcknowledgeBySignatureMediaID *uuid.UUID `json:"acknowledge_by_signature_media_id,omitempty" validate:"omitempty,uuid"`
		AcknowledgeByName             string     `json:"acknowledge_by_name,omitempty" validate:"omitempty,max=255"`
		AcknowledgeByPosition         string     `json:"acknowledge_by_position,omitempty" validate:"omitempty,max=255"`

		NotedBySignatureMediaID *uuid.UUID `json:"noted_by_signature_media_id,omitempty" validate:"omitempty,uuid"`
		NotedByName             string     `json:"noted_by_name,omitempty" validate:"omitempty,max=255"`
		NotedByPosition         string     `json:"noted_by_position,omitempty" validate:"omitempty,max=255"`

		PostedBySignatureMediaID *uuid.UUID `json:"posted_by_signature_media_id,omitempty" validate:"omitempty,uuid"`
		PostedByName             string     `json:"posted_by_name,omitempty" validate:"omitempty,max=255"`
		PostedByPosition         string     `json:"posted_by_position,omitempty" validate:"omitempty,max=255"`

		PaidBySignatureMediaID *uuid.UUID `json:"paid_by_signature_media_id,omitempty" validate:"omitempty,uuid"`
		PaidByName             string     `json:"paid_by_name,omitempty" validate:"omitempty,max=255"`
		PaidByPosition         string     `json:"paid_by_position,omitempty" validate:"omitempty,max=255"`
	}

	LoanTransactionSuggestedRequest struct {
		Amount        float64           `json:"amount" validate:"required,gt=0"`
		Principal     float64           `json:"principal" validate:"required,gt=0"`
		ModeOfPayment LoanModeOfPayment `json:"mode_of_payment"`
		FixedDays     int               `json:"fixed_days,omitempty" validate:"omitempty"`
	}

	LoanTransactionSuggestedResponse struct {
		Terms int `json:"terms"`
	}
)

func (m *LoanTransaction) ReadableReleaseDate() string {
	if m.ReleasedDate != nil {
		return handlers.ToReadableDate(*m.ReleasedDate)
	}
	return ""
}

func (m *LoanTransaction) ReadableDueDate() string {
	if m.ReleasedDate == nil {
		return ""
	}
	due := m.nextDueDate(*m.ReleasedDate)
	return handlers.ToReadableDate(due)
}

func (m *LoanTransaction) nextDueDate(from time.Time) time.Time {
	var due time.Time
	switch m.ModeOfPayment {
	case LoanModeOfPaymentDaily:
		due = from.AddDate(0, 0, 1)
	case LoanModeOfPaymentWeekly:
		var target time.Weekday
		switch m.ModeOfPaymentWeekly {
		case WeekdaySunday:
			target = time.Sunday
		case WeekdayMonday:
			target = time.Monday
		case WeekdayTuesday:
			target = time.Tuesday
		case WeekdayWednesday:
			target = time.Wednesday
		case WeekdayThursday:
			target = time.Thursday
		case WeekdayFriday:
			target = time.Friday
		case WeekdaySaturday:
			target = time.Saturday
		default:
			target = from.Weekday()
		}
		d := from.AddDate(0, 0, 1)
		for d.Weekday() != target {
			d = d.AddDate(0, 0, 1)
		}
		due = d

	case LoanModeOfPaymentSemiMonthly:
		day := from.Day()
		year, month := from.Year(), from.Month()
		if day < 15 {
			due = time.Date(year, month, 15, from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), from.Location())
		} else {
			firstNext := time.Date(year, month+1, 1, from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), from.Location())
			last := firstNext.AddDate(0, 0, -1)
			due = time.Date(last.Year(), last.Month(), last.Day(), from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), from.Location())
		}

	case LoanModeOfPaymentMonthly:
		due = handlers.AddMonthsPreserveDay(from, 1)

	case LoanModeOfPaymentQuarterly:
		due = handlers.AddMonthsPreserveDay(from, 3)

	case LoanModeOfPaymentSemiAnnual:
		due = handlers.AddMonthsPreserveDay(from, 6)

	case LoanModeOfPaymentLumpsum:
		if m.Terms > 0 {
			due = handlers.AddMonthsPreserveDay(from, m.Terms)
		} else {
			due = handlers.AddMonthsPreserveDay(from, 1)
		}

	case LoanModeOfPaymentFixedDays:
		if m.ModeOfPaymentFixedDays > 0 {
			due = from.AddDate(0, 0, m.ModeOfPaymentFixedDays)
		} else {
			due = from.AddDate(0, 0, 1)
		}

	default:
		due = from.AddDate(0, 0, 1)
	}

	for {
		if m.ExcludeSaturday && due.Weekday() == time.Saturday {
			due = due.AddDate(0, 0, 1)
			continue
		}
		if m.ExcludeSunday && due.Weekday() == time.Sunday {
			due = due.AddDate(0, 0, 1)
			continue
		}
		break
	}

	return due
}

func (m *Core) LoanWeeklyIota(weekday Weekdays) int {
	switch weekday {
	case WeekdaySunday:
		return 0
	case WeekdayMonday:
		return 1
	case WeekdayTuesday:
		return 2
	case WeekdayWednesday:
		return 3
	case WeekdayThursday:
		return 4
	case WeekdayFriday:
		return 5
	case WeekdaySaturday:
		return 6
	default:
		return -1
	}
}

func (m *Core) loanTransaction() {
	m.Migration = append(m.Migration, &LoanTransaction{})
	m.LoanTransactionManager = registry.NewRegistry(registry.RegistryParams[
		LoanTransaction, LoanTransactionResponse, LoanTransactionRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "EmployeeUser",
			"TransactionBatch", "LoanPurpose", "LoanStatus",
			"ComakerDepositMemberAccountingLedger", "PreviousLoan", "ComakerDepositMemberAccountingLedger.Account",
			"Account",
			"Account.Currency",
			"MemberProfile", "MemberJointAccount", "SignatureMedia", "MemberProfile.Media",
			"MemberProfile.SignatureMedia", "MemberProfile.MemberType",
			"ReleasedBy", "PrintedBy", "ApprovedBy",
			"ApprovedBySignatureMedia", "PreparedBySignatureMedia", "CertifiedBySignatureMedia",
			"VerifiedBySignatureMedia", "CheckBySignatureMedia", "AcknowledgeBySignatureMedia",
			"NotedBySignatureMedia", "PostedBySignatureMedia", "PaidBySignatureMedia",
			"LoanTags",
			"LoanTransactionEntries",
			"LoanTransactionEntries.Account",
			"LoanTransactionEntries.Account.Currency",
			"LoanTransactionEntries.AutomaticLoanDeduction",
			"LoanClearanceAnalysis",
			"LoanClearanceAnalysisInstitution",
			"LoanTermsAndConditionSuggestedPayment",
			"LoanTermsAndConditionAmountReceipt", "LoanTermsAndConditionAmountReceipt.Account",
			"ComakerMemberProfiles", "ComakerMemberProfiles.MemberProfile", "ComakerMemberProfiles.MemberProfile.Media",
			"ComakerCollaterals", "ComakerCollaterals.Collateral",
			"PreviousLoan.Account",
			"ReleasedBy", "PrintedBy", "ApprovedBy",
			"LoanAccounts", "LoanAccounts.Account", "LoanAccounts.Account.Currency",
			"ReleasedBy.Media", "PrintedBy.Media", "ApprovedBy.Media",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
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
				CheckDate:                              data.CheckDate,
				CheckNumber:                            data.CheckNumber,
				LoanPurposeID:                          data.LoanPurposeID,
				LoanPurpose:                            m.LoanPurposeManager.ToModel(data.LoanPurpose),
				LoanStatusID:                           data.LoanStatusID,
				LoanStatus:                             m.LoanStatusManager.ToModel(data.LoanStatus),
				ModeOfPayment:                          data.ModeOfPayment,
				ModeOfPaymentWeekly:                    data.ModeOfPaymentWeekly,
				ModeOfPaymentSemiMonthlyPay1:           data.ModeOfPaymentSemiMonthlyPay1,
				ModeOfPaymentSemiMonthlyPay2:           data.ModeOfPaymentSemiMonthlyPay2,
				ModeOfPaymentFixedDays:                 data.ModeOfPaymentFixedDays,
				ModeOfPaymentMonthlyExactDay:           data.ModeOfPaymentMonthlyExactDay,
				ComakerType:                            data.ComakerType,
				ComakerDepositMemberAccountingLedgerID: data.ComakerDepositMemberAccountingLedgerID,
				ComakerDepositMemberAccountingLedger:   m.MemberAccountingLedgerManager.ToModel(data.ComakerDepositMemberAccountingLedger),
				CollectorPlace:                         data.CollectorPlace,
				LoanType:                               data.LoanType,
				PreviousLoanID:                         data.PreviousLoanID,
				PreviousLoan:                           m.LoanTransactionManager.ToModel(data.PreviousLoan),
				Terms:                                  data.Terms,
				Amortization:                           data.Amortization,
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
				PrintNumber:                            data.PrintNumber,
				ApprovedDate:                           data.ApprovedDate,
				ReleasedDate:                           data.ReleasedDate,
				ReleasedByID:                           data.ReleasedByID,
				ReleasedBy:                             m.UserManager.ToModel(data.ReleasedBy),
				PrintedByID:                            data.PrintedByID,
				PrintedBy:                              m.UserManager.ToModel(data.PrintedBy),
				ApprovedByID:                           data.ApprovedByID,
				ApprovedBy:                             m.UserManager.ToModel(data.ApprovedBy),
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
				LoanTags:                               m.LoanTagManager.ToModels(data.LoanTags),
				LoanTransactionEntries:                 m.mapLoanTransactionEntries(data.LoanTransactionEntries),
				LoanClearanceAnalysis:                  m.LoanClearanceAnalysisManager.ToModels(data.LoanClearanceAnalysis),
				LoanClearanceAnalysisInstitution:       m.LoanClearanceAnalysisInstitutionManager.ToModels(data.LoanClearanceAnalysisInstitution),
				LoanTermsAndConditionSuggestedPayment:  m.LoanTermsAndConditionSuggestedPaymentManager.ToModels(data.LoanTermsAndConditionSuggestedPayment),
				LoanTermsAndConditionAmountReceipt:     m.LoanTermsAndConditionAmountReceiptManager.ToModels(data.LoanTermsAndConditionAmountReceipt),
				ComakerMemberProfiles:                  m.ComakerMemberProfileManager.ToModels(data.ComakerMemberProfiles),
				ComakerCollaterals:                     m.ComakerCollateralManager.ToModels(data.ComakerCollaterals),
				LoanAccounts:                           m.LoanAccountManager.ToModels(data.LoanAccounts),
				Count:                                  data.Count,
				Balance:                                data.Balance,
				LastPay:                                data.LastPay,
				Fines:                                  data.Fines,
				Interest:                               data.Interest,
				TotalDebit:                             data.TotalDebit,
				TotalCredit:                            data.TotalCredit,
				Processing:                             data.Processing,
			}
		},

		Created: func(data *LoanTransaction) registry.Topics {
			return []string{
				"loan_transaction.create",
				fmt.Sprintf("loan_transaction.create.%s", data.ID),
				fmt.Sprintf("loan_transaction.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanTransaction) registry.Topics {
			return []string{
				"loan_transaction.update",
				fmt.Sprintf("loan_transaction.update.%s", data.ID),
				fmt.Sprintf("loan_transaction.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanTransaction) registry.Topics {
			return []string{
				"loan_transaction.delete",
				fmt.Sprintf("loan_transaction.delete.%s", data.ID),
				fmt.Sprintf("loan_transaction.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) LoanTransactionCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*LoanTransaction, error) {
	return m.LoanTransactionManager.Find(context, &LoanTransaction{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func (m *Core) mapLoanTransactionEntries(entries []*LoanTransactionEntry) []*LoanTransactionEntryResponse {
	if entries == nil {
		return nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Index < entries[j].Index
	})
	var result []*LoanTransactionEntryResponse
	for _, entry := range entries {
		if entry != nil {
			result = append(result, m.LoanTransactionEntryManager.ToModel(entry))
		}
	}
	return result
}

func (m *Core) LoanTransactionWithDatesNotNull(ctx context.Context, memberID, branchID, organizationID uuid.UUID) ([]*LoanTransaction, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
	}

	return m.LoanTransactionManager.ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func (m *Core) LoanTransactionsMemberAccount(ctx context.Context, memberID, accountID, branchID, organizationID uuid.UUID) ([]*LoanTransaction, error) {

	account, err := m.AccountManager.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.Type != AccountTypeLoan {
		accountID = *account.LoanAccountID
	}
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
	}

	return m.LoanTransactionManager.ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func (m *Core) LoanTransactionDraft(ctx context.Context, branchID, organizationID uuid.UUID) ([]*LoanTransaction, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "approved_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "printed_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	return m.LoanTransactionManager.ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func (m *Core) LoanTransactionPrinted(ctx context.Context, branchID, organizationID uuid.UUID) ([]*LoanTransaction, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	return m.LoanTransactionManager.ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func (m *Core) LoanTransactionApproved(ctx context.Context, branchID, organizationID uuid.UUID) ([]*LoanTransaction, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	return m.LoanTransactionManager.ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func (m *Core) LoanTransactionReleased(ctx context.Context, branchID, organizationID uuid.UUID) ([]*LoanTransaction, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
	}

	return m.LoanTransactionManager.ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func (m *Core) LoanTransactionReleasedCurrentDay(ctx context.Context, branchID, organizationID uuid.UUID) ([]*LoanTransaction, error) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "created_at", Op: query.ModeGTE, Value: startOfDay},
		{Field: "created_at", Op: query.ModeLT, Value: endOfDay},
	}

	return m.LoanTransactionManager.ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}
