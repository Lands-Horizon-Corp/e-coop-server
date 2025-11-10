package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// GeneralLedgerType define this type according to your domain (e.g. as string or int)
type GeneralLedgerType string // adjust as needed

// General ledger type constants
const (
	GLTypeAssets      GeneralLedgerType = "Assets"
	GLTypeLiabilities GeneralLedgerType = "Liabilities"
	GLTypeEquity      GeneralLedgerType = "Equity"
	GLTypeRevenue     GeneralLedgerType = "Revenue"
	GLTypeExpenses    GeneralLedgerType = "Expenses"
)

// AccountType represents the type of account in the system
type AccountType string

// Account type constants
const (
	AccountTypeDeposit     AccountType = "Deposit"
	AccountTypeLoan        AccountType = "Loan"
	AccountTypeARLedger    AccountType = "A/R-Ledger"
	AccountTypeARAging     AccountType = "A/R-Aging"
	AccountTypeFines       AccountType = "Fines"
	AccountTypeInterest    AccountType = "Interest"
	AccountTypeSVFLedger   AccountType = "SVF-Ledger"
	AccountTypeWOff        AccountType = "W-Off"
	AccountTypeAPLedger    AccountType = "A/P-Ledger"
	AccountTypeOther       AccountType = "Other"
	AccountTypeTimeDeposit AccountType = "Time Deposit"
)

// LumpsumComputationType represents the type of lumpsum computation
type LumpsumComputationType string

// Lumpsum computation type constants
const (
	LumpsumComputationNone             LumpsumComputationType = "None"
	LumpsumComputationFinesMaturity    LumpsumComputationType = "Compute Fines Maturity"
	LumpsumComputationInterestMaturity LumpsumComputationType = "Compute Interest Maturity / Terms"
	LumpsumComputationAdvanceInterest  LumpsumComputationType = "Compute Advance Interest"
)

// InterestFinesComputationDiminishing represents the type of interest fines computation diminishing
type InterestFinesComputationDiminishing string

// Interest fines computation diminishing constants
const (
	IFCDNone                  InterestFinesComputationDiminishing = "None"
	IFCDByAmortization        InterestFinesComputationDiminishing = "By Amortization"
	IFCDByAmortizationDalyArr InterestFinesComputationDiminishing = "By Amortization Daly on Interest Principal + Interest = Fines(Arr)"
)

// InterestFinesComputationDiminishingStraightYearly represents the type of interest fines computation diminishing straight yearly
type InterestFinesComputationDiminishingStraightYearly string

// Interest fines computation diminishing straight yearly constants
const (
	IFCDSYNone                   InterestFinesComputationDiminishingStraightYearly = "None"
	IFCDSYByDailyInterestBalance InterestFinesComputationDiminishingStraightYearly = "By Daily on Interest based on loan balance by year Principal + Interest Amortization = Fines Fines Grace Period Month end Amortization"
)

// EarnedUnearnedInterest indicates how interest is recorded for an account
// (earned, unearned, formula-based, or advanced interest handling).
type EarnedUnearnedInterest string

// Values for EarnedUnearnedInterest
const (
	EUITypeNone                    EarnedUnearnedInterest = "None"
	EUITypeByFormula               EarnedUnearnedInterest = "By Formula"
	EUITypeByFormulaActualPay      EarnedUnearnedInterest = "By Formula + Actual Pay"
	EUITypeByAdvanceInterestActual EarnedUnearnedInterest = "By Advance Interest + Actual Pay"
)

// LoanSavingType indicates how loan-linked savings are stored and reported.
type LoanSavingType string

// Values for LoanSavingType
const (
	LSTSeparate                 LoanSavingType = "Separate"
	LSTSingleLedger             LoanSavingType = "Single Ledger"
	LSTSingleLedgerIfNotZero    LoanSavingType = "Single Ledger if Not Zero"
	LSTSingleLedgerSemi1530     LoanSavingType = "Single Ledger Semi (15/30)"
	LSTSingleLedgerSemiMaturity LoanSavingType = "Single Ledger Semi Within Maturity"
)

// InterestDeduction indicates whether interest applies above/below a threshold.
type InterestDeduction string

// Values for InterestDeduction
const (
	InterestDeductionAbove InterestDeduction = "Above"
	InterestDeductionBelow InterestDeduction = "Below"
)

// OtherDeductionEntry represents additional deduction categories for accounts.
type OtherDeductionEntry string

// Values for OtherDeductionEntry
const (
	OtherDeductionEntryNone       OtherDeductionEntry = "None"
	OtherDeductionEntryHealthCare OtherDeductionEntry = "Health Care"
)

// InterestSavingTypeDiminishingStraight represents interest saving options for diminishing-straight loans.
type InterestSavingTypeDiminishingStraight string

// Values for InterestSavingTypeDiminishingStraight
const (
	ISTDSSpread     InterestSavingTypeDiminishingStraight = "Spread"
	ISTDS1stPayment InterestSavingTypeDiminishingStraight = "1st Payment"
)

// OtherInformationOfAnAccount represents miscellaneous account flags and metadata.
type OtherInformationOfAnAccount string

// Values for OtherInformationOfAnAccount
const (
	OIOANone               OtherInformationOfAnAccount = "None"
	OIOAJewely             OtherInformationOfAnAccount = "Jewely"
	OIOAGrocery            OtherInformationOfAnAccount = "Grocery"
	OIOATrackLoanDeduction OtherInformationOfAnAccount = "Track Loan Deduction"
	OIOARestructured       OtherInformationOfAnAccount = "Restructured"
	OIOACashInBank         OtherInformationOfAnAccount = "Cash in Bank / Cash in Check Account"
	OIOACashOnHand         OtherInformationOfAnAccount = "Cash on Hand"
)

// InterestStandardComputation indicates the standard way interest is computed for an account.
type InterestStandardComputation string

// Values for InterestStandardComputation
const (
	ISCNone    InterestStandardComputation = "None"
	ISCYearly  InterestStandardComputation = "Yearly"
	ISCMonthly InterestStandardComputation = "Monthly"
)

// ComputationType enumerates the supported computation algorithms for account interest/amortization.
type ComputationType string

// Values for ComputationType
const (
	Straight             ComputationType = "Straight"
	Diminishing          ComputationType = "Diminishing"
	DiminishingAddOn     ComputationType = "DiminishingAddOn"
	DiminishingYearly    ComputationType = "DiminishingYearly"
	DiminishingStraight  ComputationType = "DiminishingStraight"
	DiminishingQuarterly ComputationType = "DiminishingQuarterly"
)

// --- MODEL ---

type (
	// Account represents the Account model.
	Account struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account;uniqueIndex:idx_account_name_org_branch" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account;uniqueIndex:idx_account_name_org_branch" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		GeneralLedgerDefinitionID *uuid.UUID               `gorm:"type:uuid" json:"general_ledger_definition_id"`
		GeneralLedgerDefinition   *GeneralLedgerDefinition `gorm:"foreignKey:GeneralLedgerDefinitionID" json:"general_ledger_definition,omitempty"`

		FinancialStatementDefinitionID *uuid.UUID                    `gorm:"type:uuid" json:"financial_statement_definition_id"`
		FinancialStatementDefinition   *FinancialStatementDefinition `gorm:"foreignKey:FinancialStatementDefinitionID;constraint:OnDelete:SET NULL;" json:"financial_statement_definition,omitempty"`

		AccountClassificationID *uuid.UUID             `gorm:"type:uuid" json:"account_classification_id"`
		AccountClassification   *AccountClassification `gorm:"foreignKey:AccountClassificationID;constraint:OnDelete:SET NULL;" json:"account_classification,omitempty"`

		AccountCategoryID *uuid.UUID       `gorm:"type:uuid" json:"account_category_id"`
		AccountCategory   *AccountCategory `gorm:"foreignKey:AccountCategoryID;constraint:OnDelete:SET NULL;" json:"account_category,omitempty"`

		MemberTypeID *uuid.UUID  `gorm:"type:uuid" json:"member_type_id"`
		MemberType   *MemberType `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:SET NULL;" json:"member_type,omitempty"`

		CurrencyID *uuid.UUID `gorm:"type:uuid" json:"currency_id"`
		Currency   *Currency  `gorm:"foreignKey:CurrencyID;constraint:OnDelete:SET NULL;" json:"currency,omitempty"`

		DefaultPaymentTypeID *uuid.UUID   `gorm:"type:uuid" json:"default_payment_type_id,omitempty"`
		DefaultPaymentType   *PaymentType `gorm:"foreignKey:DefaultPaymentTypeID;constraint:OnDelete:SET NULL;" json:"default_payment_type,omitempty"`

		Name        string `gorm:"type:varchar(255);not null;uniqueIndex:idx_account_name_org_branch" json:"name"`
		Description string `gorm:"type:text;not null" json:"description"`

		MinAmount float64     `gorm:"type:decimal;default:0" json:"min_amount"`
		MaxAmount float64     `gorm:"type:decimal;default:50000" json:"max_amount"`
		Index     int         `gorm:"default:0" json:"index"`
		Type      AccountType `gorm:"type:varchar(50);not null" json:"type"`

		IsInternal         bool `gorm:"default:false" json:"is_internal"`
		CashOnHand         bool `gorm:"default:false" json:"cash_on_hand"`
		PaidUpShareCapital bool `gorm:"default:false" json:"paid_up_share_capital"`

		ComputationType ComputationType `gorm:"type:varchar(50);default:'Straight'" json:"computation_type"`

		FinesAmort       float64 `gorm:"type:decimal;default:0;check:fines_amort >= 0 AND fines_amort <= 100" json:"fines_amort"`
		FinesMaturity    float64 `gorm:"type:decimal;default:0;check:fines_maturity >= 0 AND fines_maturity <= 100" json:"fines_maturity"`
		InterestStandard float64 `gorm:"type:decimal;default:0" json:"interest_standard"`
		InterestSecured  float64 `gorm:"type:decimal;default:0" json:"interest_secured"`

		ComputationSheetID *uuid.UUID        `gorm:"type:uuid" json:"computation_sheet_id"`
		ComputationSheet   *ComputationSheet `gorm:"foreignKey:ComputationSheetID;constraint:OnDelete:SET NULL;" json:"computation_sheet,omitempty"`

		CohCibFinesGracePeriodEntryCashHand                float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_cash_hand >= 0 AND coh_cib_fines_grace_period_entry_cash_hand <= 100" json:"coh_cib_fines_grace_period_entry_cash_hand"`
		CohCibFinesGracePeriodEntryCashInBank              float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_cash_in_bank >= 0 AND coh_cib_fines_grace_period_entry_cash_in_bank <= 100" json:"coh_cib_fines_grace_period_entry_cash_in_bank"`
		CohCibFinesGracePeriodEntryDailyAmortization       float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_daily_amortization >= 0 AND coh_cib_fines_grace_period_entry_daily_amortization <= 100" json:"coh_cib_fines_grace_period_entry_daily_amortization"`
		CohCibFinesGracePeriodEntryDailyMaturity           float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_daily_maturity >= 0 AND coh_cib_fines_grace_period_entry_daily_maturity <= 100" json:"coh_cib_fines_grace_period_entry_daily_maturity"`
		CohCibFinesGracePeriodEntryWeeklyAmortization      float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_weekly_amortization >= 0 AND coh_cib_fines_grace_period_entry_weekly_amortization <= 100" json:"coh_cib_fines_grace_period_entry_weekly_amortization"`
		CohCibFinesGracePeriodEntryWeeklyMaturity          float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_weekly_maturity >= 0 AND coh_cib_fines_grace_period_entry_weekly_maturity <= 100" json:"coh_cib_fines_grace_period_entry_weekly_maturity"`
		CohCibFinesGracePeriodEntryMonthlyAmortization     float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_monthly_amortization >= 0 AND coh_cib_fines_grace_period_entry_monthly_amortization <= 100" json:"coh_cib_fines_grace_period_entry_monthly_amortization"`
		CohCibFinesGracePeriodEntryMonthlyMaturity         float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_monthly_maturity >= 0 AND coh_cib_fines_grace_period_entry_monthly_maturity <= 100" json:"coh_cib_fines_grace_period_entry_monthly_maturity"`
		CohCibFinesGracePeriodEntrySemiMonthlyAmortization float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_semi_monthly_amortization >= 0 AND coh_cib_fines_grace_period_entry_semi_monthly_amortization <= 100" json:"coh_cib_fines_grace_period_entry_semi_monthly_amortization"`
		CohCibFinesGracePeriodEntrySemiMonthlyMaturity     float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_semi_monthly_maturity >= 0 AND coh_cib_fines_grace_period_entry_semi_monthly_maturity <= 100" json:"coh_cib_fines_grace_period_entry_semi_monthly_maturity"`
		CohCibFinesGracePeriodEntryQuarterlyAmortization   float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_quarterly_amortization >= 0 AND coh_cib_fines_grace_period_entry_quarterly_amortization <= 100" json:"coh_cib_fines_grace_period_entry_quarterly_amortization"`
		CohCibFinesGracePeriodEntryQuarterlyMaturity       float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_quarterly_maturity >= 0 AND coh_cib_fines_grace_period_entry_quarterly_maturity <= 100" json:"coh_cib_fines_grace_period_entry_quarterly_maturity"`
		CohCibFinesGracePeriodEntrySemiAnnualAmortization  float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_semi_annual_amortization >= 0 AND coh_cib_fines_grace_period_entry_semi_annual_amortization <= 100" json:"coh_cib_fines_grace_period_entry_semi_annual_amortization"`
		CohCibFinesGracePeriodEntrySemiAnnualMaturity      float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_semi_annual_maturity >= 0 AND coh_cib_fines_grace_period_entry_semi_annual_maturity <= 100" json:"coh_cib_fines_grace_period_entry_semi_annual_maturity"`
		CohCibFinesGracePeriodEntryAnnualAmortization      float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_annual_amortization >= 0 AND coh_cib_fines_grace_period_entry_annual_amortization <= 100" json:"coh_cib_fines_grace_period_entry_annual_amortization"`
		CohCibFinesGracePeriodEntryAnnualMaturity          float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_annual_maturity >= 0 AND coh_cib_fines_grace_period_entry_annual_maturity <= 100" json:"coh_cib_fines_grace_period_entry_annual_maturity"`
		CohCibFinesGracePeriodEntryLumpsumAmortization     float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_lumpsum_amortization >= 0 AND coh_cib_fines_grace_period_entry_lumpsum_amortization <= 100" json:"coh_cib_fines_grace_period_entry_lumpsum_amortization"`
		CohCibFinesGracePeriodEntryLumpsumMaturity         float64 `gorm:"type:decimal;default:0;check:coh_cib_fines_grace_period_entry_lumpsum_maturity >= 0 AND coh_cib_fines_grace_period_entry_lumpsum_maturity <= 100" json:"coh_cib_fines_grace_period_entry_lumpsum_maturity"`

		GeneralLedgerType GeneralLedgerType `gorm:"type:varchar(50)" json:"general_ledger_type"`

		LoanAccountID *uuid.UUID `gorm:"type:uuid" json:"loan_account_id"`
		LoanAccount   *Account   `gorm:"foreignKey:LoanAccountID;constraint:OnDelete:SET NULL;" json:"loan_account,omitempty"`

		FinesGracePeriodAmortization int  `gorm:"type:int;default:0" json:"fines_grace_period_amortization"`
		AdditionalGracePeriod        int  `gorm:"type:int;default:0" json:"additional_grace_period"`
		NoGracePeriodDaily           bool `gorm:"default:false" json:"no_grace_period_daily"`
		FinesGracePeriodMaturity     int  `gorm:"type:int;default:0" json:"fines_grace_period_maturity"`
		YearlySubscriptionFee        int  `gorm:"type:int;default:0" json:"yearly_subscription_fee"`
		CutOffDays                   int  `gorm:"type:int;default:0;check:cut_off_days >= 0 AND cut_off_days <= 30" json:"cut_off_days"`
		CutOffMonths                 int  `gorm:"type:int;default:0;check:cut_off_months >= 0 AND cut_off_months <= 12" json:"cut_off_months"`

		LumpsumComputationType                            LumpsumComputationType                            `gorm:"type:varchar(50);default:'None'" json:"lumpsum_computation_type"`
		InterestFinesComputationDiminishing               InterestFinesComputationDiminishing               `gorm:"type:varchar(100);default:'None'" json:"interest_fines_computation_diminishing"`
		InterestFinesComputationDiminishingStraightYearly InterestFinesComputationDiminishingStraightYearly `gorm:"type:varchar(200);default:'None'" json:"interest_fines_computation_diminishing_straight_yearly"`
		EarnedUnearnedInterest                            EarnedUnearnedInterest                            `gorm:"type:varchar(50);default:'None'" json:"earned_unearned_interest"`
		LoanSavingType                                    LoanSavingType                                    `gorm:"type:varchar(50);default:'Separate'" json:"loan_saving_type"`
		InterestDeduction                                 InterestDeduction                                 `gorm:"type:varchar(10);default:'Above'" json:"interest_deduction"`
		OtherDeductionEntry                               OtherDeductionEntry                               `gorm:"type:varchar(20);default:'None'" json:"other_deduction_entry"`
		InterestSavingTypeDiminishingStraight             InterestSavingTypeDiminishingStraight             `gorm:"type:varchar(20);default:'Spread'" json:"interest_saving_type_diminishing_straight"`
		OtherInformationOfAnAccount                       OtherInformationOfAnAccount                       `gorm:"type:varchar(50);default:'None'" json:"other_information_of_an_account"`

		HeaderRow int `gorm:"type:int;default:0" json:"header_row"`
		CenterRow int `gorm:"type:int;default:0" json:"center_row"`
		TotalRow  int `gorm:"type:int;default:0" json:"total_row"`

		AccountTags                         []*AccountTag `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE;" json:"account_tags,omitempty"`
		GeneralLedgerGroupingExcludeAccount bool          `gorm:"default:false" json:"general_ledger_grouping_exclude_account"`

		Icon string `gorm:"type:varchar(50);default:'account'" json:"icon,omitempty"`

		// General Ledger Source
		ShowInGeneralLedgerSourceWithdraw       bool `gorm:"default:true" json:"show_in_general_ledger_source_withdraw"`
		ShowInGeneralLedgerSourceDeposit        bool `gorm:"default:true" json:"show_in_general_ledger_source_deposit"`
		ShowInGeneralLedgerSourceJournal        bool `gorm:"default:true" json:"show_in_general_ledger_source_journal"`
		ShowInGeneralLedgerSourcePayment        bool `gorm:"default:true" json:"show_in_general_ledger_source_payment"`
		ShowInGeneralLedgerSourceAdjustment     bool `gorm:"default:true" json:"show_in_general_ledger_source_adjustment"`
		ShowInGeneralLedgerSourceJournalVoucher bool `gorm:"default:true" json:"show_in_general_ledger_source_journal_voucher"`
		ShowInGeneralLedgerSourceCheckVoucher   bool `gorm:"default:true" json:"show_in_general_ledger_source_check_voucher"`

		CompassionFund         bool    `gorm:"default:false" json:"compassion_fund"`
		CompassionFundAmount   float64 `gorm:"type:decimal;default:0" json:"compassion_fund_amount"`
		CashAndCashEquivalence bool    `gorm:"default:false" json:"cash_and_cash_equivalence"`
		// AccountResponse

		InterestStandardComputation InterestStandardComputation `gorm:"type:varchar(20);default:'None'" json:"interest_standard_computation"`
		// AccountResponse
	}
)

// --- RESPONSE & REQUEST STRUCTS ---

// AccountResponse represents the response structure for account data
type AccountResponse struct {
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

	GeneralLedgerDefinitionID      *uuid.UUID                            `json:"general_ledger_definition_id,omitempty"`
	GeneralLedgerDefinition        *GeneralLedgerDefinitionResponse      `json:"general_ledger_definition,omitempty"`
	FinancialStatementDefinitionID *uuid.UUID                            `json:"financial_statement_definition_id,omitempty"`
	FinancialStatementDefinition   *FinancialStatementDefinitionResponse `json:"financial_statement_definition,omitempty"`
	AccountClassificationID        *uuid.UUID                            `json:"account_classification_id,omitempty"`
	AccountClassification          *AccountClassificationResponse        `json:"account_classification,omitempty"`
	AccountCategoryID              *uuid.UUID                            `json:"account_category_id,omitempty"`
	AccountCategory                *AccountCategoryResponse              `json:"account_category,omitempty"`
	MemberTypeID                   *uuid.UUID                            `json:"member_type_id,omitempty"`
	MemberType                     *MemberTypeResponse                   `json:"member_type,omitempty"`
	CurrencyID                     *uuid.UUID                            `json:"currency_id,omitempty"`
	Currency                       *CurrencyResponse                     `json:"currency,omitempty"`
	DefaultPaymentTypeID           *uuid.UUID                            `json:"default_payment_type_id,omitempty"`
	DefaultPaymentType             *PaymentTypeResponse                  `json:"default_payment_type,omitempty"`

	Name        string      `json:"name"`
	Description string      `json:"description"`
	MinAmount   float64     `json:"min_amount"`
	MaxAmount   float64     `json:"max_amount"`
	Index       int         `json:"index"`
	Type        AccountType `json:"type"`

	IsInternal         bool `json:"is_internal"`
	CashOnHand         bool `json:"cash_on_hand"`
	PaidUpShareCapital bool `json:"paid_up_share_capital"`

	ComputationType ComputationType `json:"computation_type"`

	FinesAmort       float64 `json:"fines_amort"`
	FinesMaturity    float64 `json:"fines_maturity"`
	InterestStandard float64 `json:"interest_standard"`
	InterestSecured  float64 `json:"interest_secured"`

	ComputationSheetID *uuid.UUID                `json:"computation_sheet_id,omitempty"`
	ComputationSheet   *ComputationSheetResponse `json:"computation_sheet,omitempty"`

	CohCibFinesGracePeriodEntryCashHand                float64 `json:"coh_cib_fines_grace_period_entry_cash_hand"`
	CohCibFinesGracePeriodEntryCashInBank              float64 `json:"coh_cib_fines_grace_period_entry_cash_in_bank"`
	CohCibFinesGracePeriodEntryDailyAmortization       float64 `json:"coh_cib_fines_grace_period_entry_daily_amortization"`
	CohCibFinesGracePeriodEntryDailyMaturity           float64 `json:"coh_cib_fines_grace_period_entry_daily_maturity"`
	CohCibFinesGracePeriodEntryWeeklyAmortization      float64 `json:"coh_cib_fines_grace_period_entry_weekly_amortization"`
	CohCibFinesGracePeriodEntryWeeklyMaturity          float64 `json:"coh_cib_fines_grace_period_entry_weekly_maturity"`
	CohCibFinesGracePeriodEntryMonthlyAmortization     float64 `json:"coh_cib_fines_grace_period_entry_monthly_amortization"`
	CohCibFinesGracePeriodEntryMonthlyMaturity         float64 `json:"coh_cib_fines_grace_period_entry_monthly_maturity"`
	CohCibFinesGracePeriodEntrySemiMonthlyAmortization float64 `json:"coh_cib_fines_grace_period_entry_semi_monthly_amortization"`
	CohCibFinesGracePeriodEntrySemiMonthlyMaturity     float64 `json:"coh_cib_fines_grace_period_entry_semi_monthly_maturity"`
	CohCibFinesGracePeriodEntryQuarterlyAmortization   float64 `json:"coh_cib_fines_grace_period_entry_quarterly_amortization"`
	CohCibFinesGracePeriodEntryQuarterlyMaturity       float64 `json:"coh_cib_fines_grace_period_entry_quarterly_maturity"`
	CohCibFinesGracePeriodEntrySemiAnnualAmortization  float64 `json:"coh_cib_fines_grace_period_entry_semi_annual_amortization"`
	CohCibFinesGracePeriodEntrySemiAnnualMaturity      float64 `json:"coh_cib_fines_grace_period_entry_semi_annual_maturity"`
	CohCibFinesGracePeriodEntryAnnualAmortization      float64 `json:"coh_cib_fines_grace_period_entry_annual_amortization"`
	CohCibFinesGracePeriodEntryAnnualMaturity          float64 `json:"coh_cib_fines_grace_period_entry_annual_maturity"`
	CohCibFinesGracePeriodEntryLumpsumAmortization     float64 `json:"coh_cib_fines_grace_period_entry_lumpsum_amortization"`
	CohCibFinesGracePeriodEntryLumpsumMaturity         float64 `json:"coh_cib_fines_grace_period_entry_lumpsum_maturity"`

	GeneralLedgerType GeneralLedgerType `json:"general_ledger_type"`

	LoanAccountID *uuid.UUID       `json:"loan_account_id,omitempty"`
	LoanAccount   *AccountResponse `json:"loan_account,omitempty"`

	FinesGracePeriodAmortization int  `json:"fines_grace_period_amortization"`
	AdditionalGracePeriod        int  `json:"additional_grace_period"`
	NoGracePeriodDaily           bool `json:"no_grace_period_daily"`
	FinesGracePeriodMaturity     int  `json:"fines_grace_period_maturity"`
	YearlySubscriptionFee        int  `json:"yearly_subscription_fee"`
	CutOffDays                   int  `json:"cut_off_days"`
	CutOffMonths                 int  `json:"cut_off_months"`

	LumpsumComputationType                            LumpsumComputationType                            `json:"lumpsum_computation_type"`
	InterestFinesComputationDiminishing               InterestFinesComputationDiminishing               `json:"interest_fines_computation_diminishing"`
	InterestFinesComputationDiminishingStraightYearly InterestFinesComputationDiminishingStraightYearly `json:"interest_fines_computation_diminishing_straight_yearly"`
	EarnedUnearnedInterest                            EarnedUnearnedInterest                            `json:"earned_unearned_interest"`
	LoanSavingType                                    LoanSavingType                                    `json:"loan_saving_type"`
	InterestDeduction                                 InterestDeduction                                 `json:"interest_deduction"`
	OtherDeductionEntry                               OtherDeductionEntry                               `json:"other_deduction_entry"`
	InterestSavingTypeDiminishingStraight             InterestSavingTypeDiminishingStraight             `json:"interest_saving_type_diminishing_straight"`
	OtherInformationOfAnAccount                       OtherInformationOfAnAccount                       `json:"other_information_of_an_account"`

	HeaderRow int `json:"header_row"`
	CenterRow int `json:"center_row"`
	TotalRow  int `json:"total_row"`

	GeneralLedgerGroupingExcludeAccount bool                  `json:"general_ledger_grouping_exclude_account"`
	AccountTags                         []*AccountTagResponse `json:"account_tags,omitempty"`

	Icon                                    string `json:"icon,omitempty"`
	ShowInGeneralLedgerSourceWithdraw       bool   `json:"show_in_general_ledger_source_withdraw"`
	ShowInGeneralLedgerSourceDeposit        bool   `json:"show_in_general_ledger_source_deposit"`
	ShowInGeneralLedgerSourceJournal        bool   `json:"show_in_general_ledger_source_journal"`
	ShowInGeneralLedgerSourcePayment        bool   `json:"show_in_general_ledger_source_payment"`
	ShowInGeneralLedgerSourceAdjustment     bool   `json:"show_in_general_ledger_source_adjustment"`
	ShowInGeneralLedgerSourceJournalVoucher bool   `json:"show_in_general_ledger_source_journal_voucher"`
	// AccountRequest
	ShowInGeneralLedgerSourceCheckVoucher bool `json:"show_in_general_ledger_source_check_voucher"`

	// AccountRequest
	CompassionFund              bool                        `json:"compassion_fund"`
	CompassionFundAmount        float64                     `json:"compassion_fund_amount"`
	CashAndCashEquivalence      bool                        `json:"cash_and_cash_equivalence"`
	InterestStandardComputation InterestStandardComputation `json:"interest_standard_computation"`
}

// AccountRequest represents the request structure for creating/updating accounts
type AccountRequest struct {
	GeneralLedgerDefinitionID      *uuid.UUID `json:"general_ledger_definition_id,omitempty"`
	FinancialStatementDefinitionID *uuid.UUID `json:"financial_statement_definition_id,omitempty"`
	AccountClassificationID        *uuid.UUID `json:"account_classification_id,omitempty"`
	AccountCategoryID              *uuid.UUID `json:"account_category_id,omitempty"`
	MemberTypeID                   *uuid.UUID `json:"member_type_id,omitempty"`
	CurrencyID                     *uuid.UUID `json:"currency_id" validate:"required"`
	DefaultPaymentTypeID           *uuid.UUID `json:"default_payment_type_id,omitempty"`

	Name        string      `json:"name" validate:"required,min=1,max=255"`
	Description string      `json:"description"`
	MinAmount   float64     `json:"min_amount,omitempty"`
	MaxAmount   float64     `json:"max_amount,omitempty"`
	Index       int         `json:"index,omitempty"`
	Type        AccountType `json:"type" validate:"required"`

	IsInternal         bool `json:"is_internal,omitempty"`
	CashOnHand         bool `json:"cash_on_hand,omitempty"`
	PaidUpShareCapital bool `json:"paid_up_share_capital,omitempty"`

	ComputationType ComputationType `json:"computation_type,omitempty"`

	FinesAmort       float64 `json:"fines_amort,omitempty" validate:"gte=0,lte=100"`
	FinesMaturity    float64 `json:"fines_maturity,omitempty" validate:"gte=0,lte=100"`
	InterestStandard float64 `json:"interest_standard,omitempty" validate:"gte=0,lte=100"`
	InterestSecured  float64 `json:"interest_secured,omitempty" validate:"gte=0,lte=100"`

	ComputationSheetID *uuid.UUID `json:"computation_sheet_id,omitempty"`

	CohCibFinesGracePeriodEntryCashHand                float64 `json:"coh_cib_fines_grace_period_entry_cash_hand,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntryCashInBank              float64 `json:"coh_cib_fines_grace_period_entry_cash_in_bank,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntryDailyAmortization       float64 `json:"coh_cib_fines_grace_period_entry_daily_amortization,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntryDailyMaturity           float64 `json:"coh_cib_fines_grace_period_entry_daily_maturity,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntryWeeklyAmortization      float64 `json:"coh_cib_fines_grace_period_entry_weekly_amortization,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntryWeeklyMaturity          float64 `json:"coh_cib_fines_grace_period_entry_weekly_maturity,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntryMonthlyAmortization     float64 `json:"coh_cib_fines_grace_period_entry_monthly_amortization,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntryMonthlyMaturity         float64 `json:"coh_cib_fines_grace_period_entry_monthly_maturity,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntrySemiMonthlyAmortization float64 `json:"coh_cib_fines_grace_period_entry_semi_monthly_amortization,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntrySemiMonthlyMaturity     float64 `json:"coh_cib_fines_grace_period_entry_semi_monthly_maturity,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntryQuarterlyAmortization   float64 `json:"coh_cib_fines_grace_period_entry_quarterly_amortization,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntryQuarterlyMaturity       float64 `json:"coh_cib_fines_grace_period_entry_quarterly_maturity,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntrySemiAnnualAmortization  float64 `json:"coh_cib_fines_grace_period_entry_semi_annual_amortization,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntrySemiAnnualMaturity      float64 `json:"coh_cib_fines_grace_period_entry_semi_annual_maturity,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntryAnnualAmortization      float64 `json:"coh_cib_fines_grace_period_entry_annual_amortization,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntryAnnualMaturity          float64 `json:"coh_cib_fines_grace_period_entry_annual_maturity,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntryLumpsumAmortization     float64 `json:"coh_cib_fines_grace_period_entry_lumpsum_amortization,omitempty" validate:"gte=0,lte=100"`
	CohCibFinesGracePeriodEntryLumpsumMaturity         float64 `json:"coh_cib_fines_grace_period_entry_lumpsum_maturity,omitempty" validate:"gte=0,lte=100"`

	GeneralLedgerType GeneralLedgerType `json:"general_ledger_type,omitempty"`

	LoanAccountID *uuid.UUID `json:"loan_account_id,omitempty"`

	FinesGracePeriodAmortization int  `json:"fines_grace_period_amortization,omitempty" validate:"gte=0,lte=365"`
	AdditionalGracePeriod        int  `json:"additional_grace_period,omitempty" validate:"gte=0,lte=365"`
	NoGracePeriodDaily           bool `json:"no_grace_period_daily,omitempty"`
	FinesGracePeriodMaturity     int  `json:"fines_grace_period_maturity,omitempty" validate:"gte=0,lte=365"`
	YearlySubscriptionFee        int  `json:"yearly_subscription_fee,omitempty" validate:"gte=0"`
	CutOffDays                   int  `json:"cut_off_days,omitempty" validate:"gte=0,lte=30"`
	CutOffMonths                 int  `json:"cut_off_months,omitempty" validate:"gte=0,lte=12"`

	LumpsumComputationType                            LumpsumComputationType                            `json:"lumpsum_computation_type,omitempty"`
	InterestFinesComputationDiminishing               InterestFinesComputationDiminishing               `json:"interest_fines_computation_diminishing,omitempty"`
	InterestFinesComputationDiminishingStraightYearly InterestFinesComputationDiminishingStraightYearly `json:"interest_fines_computation_diminishing_straight_yearly,omitempty"`
	EarnedUnearnedInterest                            EarnedUnearnedInterest                            `json:"earned_unearned_interest,omitempty"`
	LoanSavingType                                    LoanSavingType                                    `json:"loan_saving_type,omitempty"`
	InterestDeduction                                 InterestDeduction                                 `json:"interest_deduction,omitempty"`
	OtherDeductionEntry                               OtherDeductionEntry                               `json:"other_deduction_entry,omitempty"`
	InterestSavingTypeDiminishingStraight             InterestSavingTypeDiminishingStraight             `json:"interest_saving_type_diminishing_straight,omitempty"`
	OtherInformationOfAnAccount                       OtherInformationOfAnAccount                       `json:"other_information_of_an_account,omitempty"`

	HeaderRow int `json:"header_row,omitempty"`
	CenterRow int `json:"center_row,omitempty"`
	TotalRow  int `json:"total_row,omitempty"`

	GeneralLedgerGroupingExcludeAccount bool                 `json:"general_ledger_grouping_exclude_account,omitempty"`
	AccountTags                         []*AccountTagRequest `json:"account_tags,omitempty"`

	Icon                                    string `json:"icon,omitempty"`
	ShowInGeneralLedgerSourceWithdraw       bool   `json:"show_in_general_ledger_source_withdraw"`
	ShowInGeneralLedgerSourceDeposit        bool   `json:"show_in_general_ledger_source_deposit"`
	ShowInGeneralLedgerSourceJournal        bool   `json:"show_in_general_ledger_source_journal"`
	ShowInGeneralLedgerSourcePayment        bool   `json:"show_in_general_ledger_source_payment"`
	ShowInGeneralLedgerSourceAdjustment     bool   `json:"show_in_general_ledger_source_adjustment"`
	ShowInGeneralLedgerSourceJournalVoucher bool   `json:"show_in_general_ledger_source_journal_voucher"`
	ShowInGeneralLedgerSourceCheckVoucher   bool   `json:"show_in_general_ledger_source_check_voucher"`

	CompassionFund              bool                        `json:"compassion_fund,omitempty"`
	CompassionFundAmount        float64                     `json:"compassion_fund_amount,omitempty"`
	CashAndCashEquivalence      bool                        `json:"cash_and_cash_equivalence,omitempty"`
	InterestStandardComputation InterestStandardComputation `json:"interest_standard_computation,omitempty"`
}

// --- REGISTRATION ---

func (m *Core) account() {
	m.Migration = append(m.Migration, &Account{})
	m.AccountManager = *registry.NewRegistry(registry.RegistryParams[
		Account, AccountResponse, AccountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"AccountClassification", "AccountCategory",
			"AccountTags", "ComputationSheet", "Currency",
			"DefaultPaymentType", "LoanAccount",
		},
		Service: m.provider.Service,
		Resource: func(data *Account) *AccountResponse {
			if data == nil {
				return nil
			}
			return &AccountResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),

				GeneralLedgerDefinitionID:      data.GeneralLedgerDefinitionID,
				GeneralLedgerDefinition:        m.GeneralLedgerDefinitionManager.ToModel(data.GeneralLedgerDefinition),
				FinancialStatementDefinitionID: data.FinancialStatementDefinitionID,
				FinancialStatementDefinition:   m.FinancialStatementDefinitionManager.ToModel(data.FinancialStatementDefinition),
				AccountClassificationID:        data.AccountClassificationID,
				AccountClassification:          m.AccountClassificationManager.ToModel(data.AccountClassification),
				AccountCategoryID:              data.AccountCategoryID,
				AccountCategory:                m.AccountCategoryManager.ToModel(data.AccountCategory),
				MemberTypeID:                   data.MemberTypeID,
				MemberType:                     m.MemberTypeManager.ToModel(data.MemberType),
				CurrencyID:                     data.CurrencyID,
				Currency:                       m.CurrencyManager.ToModel(data.Currency),
				DefaultPaymentTypeID:           data.DefaultPaymentTypeID,
				DefaultPaymentType:             m.PaymentTypeManager.ToModel(data.DefaultPaymentType),

				Name:                                  data.Name,
				Description:                           data.Description,
				MinAmount:                             data.MinAmount,
				MaxAmount:                             data.MaxAmount,
				Index:                                 data.Index,
				Type:                                  data.Type,
				IsInternal:                            data.IsInternal,
				CashOnHand:                            data.CashOnHand,
				PaidUpShareCapital:                    data.PaidUpShareCapital,
				ComputationType:                       data.ComputationType,
				FinesAmort:                            data.FinesAmort,
				FinesMaturity:                         data.FinesMaturity,
				InterestStandard:                      data.InterestStandard,
				InterestSecured:                       data.InterestSecured,
				ComputationSheetID:                    data.ComputationSheetID,
				CohCibFinesGracePeriodEntryCashHand:   data.CohCibFinesGracePeriodEntryCashHand,
				CohCibFinesGracePeriodEntryCashInBank: data.CohCibFinesGracePeriodEntryCashInBank,
				CohCibFinesGracePeriodEntryDailyAmortization:       data.CohCibFinesGracePeriodEntryDailyAmortization,
				CohCibFinesGracePeriodEntryDailyMaturity:           data.CohCibFinesGracePeriodEntryDailyMaturity,
				CohCibFinesGracePeriodEntryWeeklyAmortization:      data.CohCibFinesGracePeriodEntryWeeklyAmortization,
				CohCibFinesGracePeriodEntryWeeklyMaturity:          data.CohCibFinesGracePeriodEntryWeeklyMaturity,
				CohCibFinesGracePeriodEntryMonthlyAmortization:     data.CohCibFinesGracePeriodEntryMonthlyAmortization,
				CohCibFinesGracePeriodEntryMonthlyMaturity:         data.CohCibFinesGracePeriodEntryMonthlyMaturity,
				CohCibFinesGracePeriodEntrySemiMonthlyAmortization: data.CohCibFinesGracePeriodEntrySemiMonthlyAmortization,
				CohCibFinesGracePeriodEntrySemiMonthlyMaturity:     data.CohCibFinesGracePeriodEntrySemiMonthlyMaturity,
				CohCibFinesGracePeriodEntryQuarterlyAmortization:   data.CohCibFinesGracePeriodEntryQuarterlyAmortization,
				CohCibFinesGracePeriodEntryQuarterlyMaturity:       data.CohCibFinesGracePeriodEntryQuarterlyMaturity,
				CohCibFinesGracePeriodEntrySemiAnnualAmortization:  data.CohCibFinesGracePeriodEntrySemiAnnualAmortization,
				CohCibFinesGracePeriodEntrySemiAnnualMaturity:      data.CohCibFinesGracePeriodEntrySemiAnnualMaturity,
				CohCibFinesGracePeriodEntryAnnualAmortization:      data.CohCibFinesGracePeriodEntryAnnualAmortization,
				CohCibFinesGracePeriodEntryAnnualMaturity:          data.CohCibFinesGracePeriodEntryAnnualMaturity,
				CohCibFinesGracePeriodEntryLumpsumAmortization:     data.CohCibFinesGracePeriodEntryLumpsumAmortization,
				CohCibFinesGracePeriodEntryLumpsumMaturity:         data.CohCibFinesGracePeriodEntryLumpsumMaturity,
				GeneralLedgerType:                   data.GeneralLedgerType,
				LoanAccountID:                       data.LoanAccountID,
				LoanAccount:                         m.AccountManager.ToModel(data.LoanAccount),
				FinesGracePeriodAmortization:        data.FinesGracePeriodAmortization,
				AdditionalGracePeriod:               data.AdditionalGracePeriod,
				NoGracePeriodDaily:                  data.NoGracePeriodDaily,
				FinesGracePeriodMaturity:            data.FinesGracePeriodMaturity,
				YearlySubscriptionFee:               data.YearlySubscriptionFee,
				CutOffDays:                          data.CutOffDays,
				CutOffMonths:                        data.CutOffMonths,
				LumpsumComputationType:              data.LumpsumComputationType,
				InterestFinesComputationDiminishing: data.InterestFinesComputationDiminishing,
				InterestFinesComputationDiminishingStraightYearly: data.InterestFinesComputationDiminishingStraightYearly,
				EarnedUnearnedInterest:                            data.EarnedUnearnedInterest,
				LoanSavingType:                                    data.LoanSavingType,
				InterestDeduction:                                 data.InterestDeduction,
				OtherDeductionEntry:                               data.OtherDeductionEntry,
				InterestSavingTypeDiminishingStraight:             data.InterestSavingTypeDiminishingStraight,
				OtherInformationOfAnAccount:                       data.OtherInformationOfAnAccount,
				HeaderRow:                                         data.HeaderRow,
				CenterRow:                                         data.CenterRow,
				TotalRow:                                          data.TotalRow,
				GeneralLedgerGroupingExcludeAccount:               data.GeneralLedgerGroupingExcludeAccount,
				AccountTags:                                       m.AccountTagManager.ToModels(data.AccountTags),
				ComputationSheet:                                  m.ComputationSheetManager.ToModel(data.ComputationSheet),

				Icon:                                    data.Icon,
				ShowInGeneralLedgerSourceWithdraw:       data.ShowInGeneralLedgerSourceWithdraw,
				ShowInGeneralLedgerSourceDeposit:        data.ShowInGeneralLedgerSourceDeposit,
				ShowInGeneralLedgerSourceJournal:        data.ShowInGeneralLedgerSourceJournal,
				ShowInGeneralLedgerSourcePayment:        data.ShowInGeneralLedgerSourcePayment,
				ShowInGeneralLedgerSourceAdjustment:     data.ShowInGeneralLedgerSourceAdjustment,
				ShowInGeneralLedgerSourceJournalVoucher: data.ShowInGeneralLedgerSourceJournalVoucher,
				ShowInGeneralLedgerSourceCheckVoucher:   data.ShowInGeneralLedgerSourceCheckVoucher,

				CompassionFund:              data.CompassionFund,
				CompassionFundAmount:        data.CompassionFundAmount,
				CashAndCashEquivalence:      data.CashAndCashEquivalence,
				InterestStandardComputation: data.InterestStandardComputation,
			}
		},
		Created: func(data *Account) []string {
			return []string{
				"account.create",
				fmt.Sprintf("account.create.%s", data.ID),
				fmt.Sprintf("account.create.branch.%s", data.BranchID),
				fmt.Sprintf("account.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Account) []string {
			return []string{
				"account.update",
				fmt.Sprintf("account.update.%s", data.ID),
				fmt.Sprintf("account.update.branch.%s", data.BranchID),
				fmt.Sprintf("account.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Account) []string {
			return []string{
				"account.delete",
				fmt.Sprintf("account.delete.%s", data.ID),
				fmt.Sprintf("account.delete.branch.%s", data.BranchID),
				fmt.Sprintf("account.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) accountSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()

	branch, err := m.BranchManager.GetByID(context, branchID)
	if err != nil {
		return eris.Wrap(err, "failed to find branch for account seeding")
	}
	currency, err := m.CurrencyFindByAlpha2(context, branch.CountryCode)
	if err != nil {
		return eris.Wrap(err, "failed to find currency for account seeding")
	}

	accounts := []*Account{
		// Regular Savings Accounts
		{
			CreatedAt:         now,
			CreatedByID:       userID,
			UpdatedAt:         now,
			UpdatedByID:       userID,
			OrganizationID:    organizationID,
			BranchID:          branchID,
			Name:              "Regular Savings",
			Description:       "Basic savings account for general purpose savings with standard interest rates.",
			Type:              AccountTypeDeposit,
			MinAmount:         100.00,
			MaxAmount:         1000000.00,
			InterestStandard:  2.5,
			CurrencyID:        &currency.ID,
			GeneralLedgerType: GLTypeLiabilities,
			ComputationType:   Diminishing,
			Index:             1,
			Icon:              "Savings",
		},
		{
			CreatedAt:         now,
			CreatedByID:       userID,
			UpdatedAt:         now,
			UpdatedByID:       userID,
			OrganizationID:    organizationID,
			BranchID:          branchID,
			Name:              "Premium Savings",
			Description:       "High-yield savings account with better interest rates for higher balances.",
			Type:              AccountTypeDeposit,
			MinAmount:         5000.00,
			MaxAmount:         5000000.00,
			InterestStandard:  4.0,
			GeneralLedgerType: GLTypeLiabilities,
			ComputationType:   DiminishingYearly,
			Index:             2,
			CurrencyID:        &currency.ID,
			Icon:              "Crown",
		},
		{
			CreatedAt:         now,
			CreatedByID:       userID,
			UpdatedAt:         now,
			UpdatedByID:       userID,
			OrganizationID:    organizationID,
			BranchID:          branchID,
			Name:              "Junior Savings",
			Description:       "Special savings account designed for minors and young members.",
			Type:              AccountTypeDeposit,
			MinAmount:         50.00,
			MaxAmount:         100000.00,
			InterestStandard:  3.0,
			GeneralLedgerType: GLTypeLiabilities,
			ComputationType:   Diminishing,
			Index:             3,
			CurrencyID:        &currency.ID,
			Icon:              "Cake",
		},
		{
			CreatedAt:         now,
			CreatedByID:       userID,
			UpdatedAt:         now,
			UpdatedByID:       userID,
			OrganizationID:    organizationID,
			BranchID:          branchID,
			Name:              "Senior Citizen Savings",
			Description:       "Special savings account with higher interest rates for senior citizens.",
			Type:              AccountTypeDeposit,
			MinAmount:         500.00,
			MaxAmount:         2000000.00,
			InterestStandard:  3.5,
			GeneralLedgerType: GLTypeLiabilities,
			ComputationType:   DiminishingQuarterly,
			Index:             4,
			CurrencyID:        &currency.ID,
			Icon:              "Umbrella",
		},
		{
			CreatedAt:         now,
			CreatedByID:       userID,
			UpdatedAt:         now,
			UpdatedByID:       userID,
			OrganizationID:    organizationID,
			BranchID:          branchID,
			Name:              "Christmas Savings",
			Description:       "Seasonal savings account for holiday preparations with withdrawal restrictions.",
			Type:              AccountTypeDeposit,
			MinAmount:         200.00,
			MaxAmount:         500000.00,
			InterestStandard:  3.0,
			GeneralLedgerType: GLTypeLiabilities,
			ComputationType:   Diminishing,
			Index:             5,
			CurrencyID:        &currency.ID,
			Icon:              "Calendar",
		},
		{
			CreatedAt:         now,
			CreatedByID:       userID,
			UpdatedAt:         now,
			UpdatedByID:       userID,
			OrganizationID:    organizationID,
			BranchID:          branchID,
			Name:              "Education Savings",
			Description:       "Long-term savings account dedicated to educational expenses.",
			Type:              AccountTypeDeposit,
			MinAmount:         1000.00,
			MaxAmount:         3000000.00,
			InterestStandard:  4.0,
			GeneralLedgerType: GLTypeLiabilities,
			ComputationType:   DiminishingAddOn,
			Index:             6,
			CurrencyID:        &currency.ID,
			Icon:              "Graduation Cap",
		},
		{
			CreatedAt:         now,
			CreatedByID:       userID,
			UpdatedAt:         now,
			UpdatedByID:       userID,
			OrganizationID:    organizationID,
			BranchID:          branchID,
			Name:              "Emergency Fund",
			Description:       "High-liquidity savings account for emergency situations.",
			Type:              AccountTypeDeposit,
			MinAmount:         500.00,
			MaxAmount:         1000000.00,
			InterestStandard:  2.0,
			GeneralLedgerType: GLTypeLiabilities,
			ComputationType:   Diminishing,
			Index:             7,
			CurrencyID:        &currency.ID,
			Icon:              "Shield Check",
		},
		{
			CreatedAt:         now,
			CreatedByID:       userID,
			UpdatedAt:         now,
			UpdatedByID:       userID,
			OrganizationID:    organizationID,
			BranchID:          branchID,
			Name:              "Business Savings",
			Description:       "Savings account designed for small businesses and entrepreneurs.",
			Type:              AccountTypeDeposit,
			MinAmount:         2000.00,
			MaxAmount:         10000000.00,
			InterestStandard:  3.5,
			GeneralLedgerType: GLTypeLiabilities,
			ComputationType:   DiminishingStraight,
			Index:             8,
			CurrencyID:        &currency.ID,
			Icon:              "Brief Case",
		},
		{
			CreatedAt:         now,
			CreatedByID:       userID,
			UpdatedAt:         now,
			UpdatedByID:       userID,
			OrganizationID:    organizationID,
			BranchID:          branchID,
			Name:              "Retirement Savings",
			Description:       "Long-term savings account for retirement planning with tax benefits.",
			Type:              AccountTypeDeposit,
			MinAmount:         1000.00,
			MaxAmount:         5000000.00,
			InterestStandard:  4.5,
			GeneralLedgerType: GLTypeLiabilities,
			ComputationType:   DiminishingYearly,
			Index:             9,
			CurrencyID:        &currency.ID,
			Icon:              "Clock",
		},
	}

	// Create all deposit accounts first
	for _, data := range accounts {
		data.CurrencyID = &currency.ID
		data.ShowInGeneralLedgerSourceWithdraw = true
		data.ShowInGeneralLedgerSourceDeposit = true
		data.ShowInGeneralLedgerSourceJournal = true
		data.ShowInGeneralLedgerSourcePayment = true
		data.ShowInGeneralLedgerSourceAdjustment = true
		data.ShowInGeneralLedgerSourceJournalVoucher = true
		data.ShowInGeneralLedgerSourceCheckVoucher = true
		if err := m.AccountManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed account %s", data.Name)
		}
	}

	// Create loan accounts with their alternative accounts
	loanAccounts := []*Account{
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Emergency Loan",
			Description:                             "Quick access loan for urgent financial needs and unexpected expenses.",
			Type:                                    AccountTypeLoan,
			MinAmount:                               1000.00,
			MaxAmount:                               100000.00,
			InterestStandard:                        8.5, // Already between 0-100
			InterestSecured:                         7.5,
			FinesAmort:                              1.0,
			FinesMaturity:                           2.0,
			GeneralLedgerType:                       GLTypeAssets,
			ComputationType:                         Diminishing,
			Index:                                   10,
			CutOffDays:                              3,
			CutOffMonths:                            0,
			FinesGracePeriodAmortization:            5,
			FinesGracePeriodMaturity:                7,
			AdditionalGracePeriod:                   2,
			LumpsumComputationType:                  LumpsumComputationNone,
			InterestFinesComputationDiminishing:     IFCDByAmortization,
			EarnedUnearnedInterest:                  EUITypeByFormula,
			LoanSavingType:                          LSTSeparate,
			InterestDeduction:                       InterestDeductionAbove,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			Icon:                                    "Rocket",
		},
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Business Loan",
			Description:                             "Capital loan for business expansion, equipment purchase, and working capital needs.",
			Type:                                    AccountTypeLoan,
			MinAmount:                               50000.00,
			MaxAmount:                               5000000.00,
			InterestStandard:                        10.0, // Already between 0-100
			InterestSecured:                         9.0,
			FinesAmort:                              1.5,
			FinesMaturity:                           2.5,
			GeneralLedgerType:                       GLTypeAssets,
			ComputationType:                         DiminishingYearly,
			Index:                                   11,
			CutOffDays:                              7,
			CutOffMonths:                            0,
			FinesGracePeriodAmortization:            10,
			FinesGracePeriodMaturity:                15,
			AdditionalGracePeriod:                   5,
			LumpsumComputationType:                  LumpsumComputationNone,
			InterestFinesComputationDiminishing:     IFCDByAmortization,
			EarnedUnearnedInterest:                  EUITypeByFormula,
			LoanSavingType:                          LSTSeparate,
			InterestDeduction:                       InterestDeductionAbove,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			CurrencyID:                              &currency.ID,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			Icon:                                    "Shop Icon",
		},
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Educational Loan",
			Description:                             "Student loan for tuition fees, educational expenses, and academic development.",
			Type:                                    AccountTypeLoan,
			MinAmount:                               5000.00,
			MaxAmount:                               500000.00,
			InterestStandard:                        6.5, // Already between 0-100
			InterestSecured:                         5.5,
			FinesAmort:                              0.5,
			FinesMaturity:                           1.0,
			GeneralLedgerType:                       GLTypeAssets,
			ComputationType:                         Diminishing,
			Index:                                   12,
			CutOffDays:                              14,
			CutOffMonths:                            0,
			FinesGracePeriodAmortization:            15,
			FinesGracePeriodMaturity:                30,
			AdditionalGracePeriod:                   10,
			LumpsumComputationType:                  LumpsumComputationNone,
			InterestFinesComputationDiminishing:     IFCDNone,
			EarnedUnearnedInterest:                  EUITypeByFormula,
			LoanSavingType:                          LSTSeparate,
			InterestDeduction:                       InterestDeductionBelow,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			Icon:                                    "Book Open",
		},
	}

	// Create loan accounts and their alternative accounts
	for _, loanAccount := range loanAccounts {
		loanAccount.CurrencyID = &currency.ID
		// Create the main loan account
		if err := m.AccountManager.CreateWithTx(context, tx, loanAccount); err != nil {
			return eris.Wrapf(err, "failed to seed loan account %s", loanAccount.Name)
		}

		// Create Interest Account with varying computation types
		var interestComputationType ComputationType
		var interestStandardRate float64

		// Set different computation types and rates based on loan type
		switch loanAccount.Name {
		case "Emergency Loan":
			interestComputationType = Diminishing
			interestStandardRate = 2.5 // 2.5% interest standard
		case "Business Loan":
			interestComputationType = DiminishingStraight
			interestStandardRate = 3.0 // 3% interest standard
		case "Educational Loan":
			interestComputationType = Straight
			interestStandardRate = 1.5 // 1.5% interest standard
		default:
			interestComputationType = Diminishing
			interestStandardRate = 2.0 // 2% default interest standard
		}

		interestAccount := &Account{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			CurrencyID:                              &currency.ID,
			Name:                                    "Interest " + loanAccount.Name,
			Description:                             "Interest account for " + loanAccount.Description,
			Type:                                    AccountTypeInterest,
			MinAmount:                               0.00,
			MaxAmount:                               100.00, // Max percentage is 100%
			InterestStandard:                        interestStandardRate,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         interestComputationType,
			Index:                                   loanAccount.Index + 100, // Offset to avoid conflicts
			LoanAccountID:                           &loanAccount.ID,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			Icon:                                    "Percent",
		}

		if err := m.AccountManager.CreateWithTx(context, tx, interestAccount); err != nil {
			return eris.Wrapf(err, "failed to seed interest account for %s", loanAccount.Name)
		}

		// Create Service Fee Account with varying computation types
		var svfComputationType ComputationType
		var svfStandardRate float64

		// Set different computation types and rates based on loan type
		switch loanAccount.Name {
		case "Emergency Loan":
			svfComputationType = Straight
			svfStandardRate = 1.0 // 1% service fee standard
		case "Business Loan":
			svfComputationType = DiminishingStraight
			svfStandardRate = 1.5 // 1.5% service fee standard
		case "Educational Loan":
			svfComputationType = Diminishing
			svfStandardRate = 0.5 // 0.5% service fee standard
		default:
			svfComputationType = Straight
			svfStandardRate = 1.0 // 1% default service fee standard
		}

		serviceFeeAccount := &Account{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			CurrencyID:                              &currency.ID,
			Name:                                    "Service Fee " + loanAccount.Name,
			Description:                             "Service fee account for " + loanAccount.Description,
			Type:                                    AccountTypeSVFLedger,
			MinAmount:                               0.00,
			MaxAmount:                               100.00, // Max percentage is 100%
			InterestStandard:                        svfStandardRate,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         svfComputationType,
			Index:                                   loanAccount.Index + 200, // Offset to avoid conflicts
			LoanAccountID:                           &loanAccount.ID,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			Icon:                                    "Receipt",
		}

		if err := m.AccountManager.CreateWithTx(context, tx, serviceFeeAccount); err != nil {
			return eris.Wrapf(err, "failed to seed service fee account for %s", loanAccount.Name)
		}

		// Create Fines Account with percentage-based rates and grace periods
		finesAccount := &Account{
			CreatedAt:        now,
			CreatedByID:      userID,
			UpdatedAt:        now,
			UpdatedByID:      userID,
			OrganizationID:   organizationID,
			BranchID:         branchID,
			CurrencyID:       &currency.ID,
			Name:             "Fines " + loanAccount.Name,
			Description:      "Fines account for " + loanAccount.Description,
			Type:             AccountTypeFines,
			MinAmount:        0.00,
			MaxAmount:        100.00, // Max percentage is 100%
			InterestStandard: 0.0,

			// Percentage-based fines rates (0-100%)
			FinesAmort:    2.5, // 2.5% fine on amortization
			FinesMaturity: 5.0, // 5.0% fine on maturity

			// Grace periods for fines
			FinesGracePeriodAmortization: 7,     // 7 days grace period for amortization fines
			FinesGracePeriodMaturity:     15,    // 15 days grace period for maturity fines
			AdditionalGracePeriod:        3,     // 3 additional days
			NoGracePeriodDaily:           false, // Allow daily grace period

			// Computation settings
			GeneralLedgerType: GLTypeRevenue,
			ComputationType:   Straight,
			Index:             loanAccount.Index + 300, // Offset to avoid conflicts
			LoanAccountID:     &loanAccount.ID,

			// Enhanced grace period entries with different frequencies
			CohCibFinesGracePeriodEntryDailyAmortization:       1.0,  // 1% daily amortization fine
			CohCibFinesGracePeriodEntryDailyMaturity:           2.0,  // 2% daily maturity fine
			CohCibFinesGracePeriodEntryWeeklyAmortization:      5.0,  // 5% weekly amortization fine
			CohCibFinesGracePeriodEntryWeeklyMaturity:          8.0,  // 8% weekly maturity fine
			CohCibFinesGracePeriodEntryMonthlyAmortization:     10.0, // 10% monthly amortization fine
			CohCibFinesGracePeriodEntryMonthlyMaturity:         15.0, // 15% monthly maturity fine
			CohCibFinesGracePeriodEntrySemiMonthlyAmortization: 7.5,  // 7.5% semi-monthly amortization fine
			CohCibFinesGracePeriodEntrySemiMonthlyMaturity:     12.0, // 12% semi-monthly maturity fine
			CohCibFinesGracePeriodEntryQuarterlyAmortization:   20.0, // 20% quarterly amortization fine
			CohCibFinesGracePeriodEntryQuarterlyMaturity:       25.0, // 25% quarterly maturity fine
			CohCibFinesGracePeriodEntrySemiAnnualAmortization:  35.0, // 35% semi-annual amortization fine
			CohCibFinesGracePeriodEntrySemiAnnualMaturity:      40.0, // 40% semi-annual maturity fine
			CohCibFinesGracePeriodEntryLumpsumAmortization:     50.0, // 50% lumpsum amortization fine
			CohCibFinesGracePeriodEntryLumpsumMaturity:         60.0, // 60% lumpsum maturity fine

			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			Icon:                                    "Warning",
		}

		if err := m.AccountManager.CreateWithTx(context, tx, finesAccount); err != nil {
			return eris.Wrapf(err, "failed to seed fines account for %s", loanAccount.Name)
		}
	}

	// Create additional standalone fines accounts with different configurations
	standaloneFinesAccounts := []*Account{
		{
			CreatedAt:                    now,
			CreatedByID:                  userID,
			UpdatedAt:                    now,
			UpdatedByID:                  userID,
			OrganizationID:               organizationID,
			BranchID:                     branchID,
			CurrencyID:                   &currency.ID,
			Name:                         "Late Payment Fines",
			Description:                  "Fines for late payment of any cooperative obligations and dues.",
			Type:                         AccountTypeFines,
			MinAmount:                    0.00,
			MaxAmount:                    100.00, // Max percentage is 100%
			InterestStandard:             0.0,
			FinesAmort:                   3.0, // 3% fine on amortization
			FinesMaturity:                7.5, // 7.5% fine on maturity
			FinesGracePeriodAmortization: 5,   // 5 days grace period for amortization fines
			FinesGracePeriodMaturity:     10,  // 10 days grace period for maturity fines
			AdditionalGracePeriod:        2,   // 2 additional days
			NoGracePeriodDaily:           false,
			GeneralLedgerType:            GLTypeRevenue,
			ComputationType:              Straight,
			Index:                        500,
			CohCibFinesGracePeriodEntryDailyAmortization:   2.0,  // 2% daily amortization fine
			CohCibFinesGracePeriodEntryDailyMaturity:       3.5,  // 3.5% daily maturity fine
			CohCibFinesGracePeriodEntryWeeklyAmortization:  7.5,  // 7.5% weekly amortization fine
			CohCibFinesGracePeriodEntryWeeklyMaturity:      12.0, // 12% weekly maturity fine
			CohCibFinesGracePeriodEntryMonthlyAmortization: 15.0, // 15% monthly amortization fine
			CohCibFinesGracePeriodEntryMonthlyMaturity:     22.5, // 22.5% monthly maturity fine
			ShowInGeneralLedgerSourceWithdraw:              true,
			ShowInGeneralLedgerSourceDeposit:               true,
			ShowInGeneralLedgerSourceJournal:               true,
			ShowInGeneralLedgerSourcePayment:               true,
			ShowInGeneralLedgerSourceAdjustment:            true,
			ShowInGeneralLedgerSourceJournalVoucher:        true,
			ShowInGeneralLedgerSourceCheckVoucher:          true,
			OtherInformationOfAnAccount:                    OIOANone,
			Icon:                                           "Clock Cancel",
		},
		{
			CreatedAt:                    now,
			CreatedByID:                  userID,
			UpdatedAt:                    now,
			UpdatedByID:                  userID,
			OrganizationID:               organizationID,
			BranchID:                     branchID,
			CurrencyID:                   &currency.ID,
			Name:                         "Penalty Fines",
			Description:                  "Penalty fines for violations of cooperative rules and regulations.",
			Type:                         AccountTypeFines,
			MinAmount:                    0.00,
			MaxAmount:                    100.00,
			InterestStandard:             0.0,
			FinesAmort:                   5.0,  // 5% fine on amortization
			FinesMaturity:                10.0, // 10% fine on maturity
			FinesGracePeriodAmortization: 3,    // 3 days grace period for amortization fines
			FinesGracePeriodMaturity:     7,    // 7 days grace period for maturity fines
			AdditionalGracePeriod:        1,    // 1 additional day
			NoGracePeriodDaily:           false,
			GeneralLedgerType:            GLTypeRevenue,
			ComputationType:              Straight,
			Index:                        501,
			CohCibFinesGracePeriodEntryDailyAmortization:   3.0,  // 3% daily amortization fine
			CohCibFinesGracePeriodEntryDailyMaturity:       5.0,  // 5% daily maturity fine
			CohCibFinesGracePeriodEntryWeeklyAmortization:  10.0, // 10% weekly amortization fine
			CohCibFinesGracePeriodEntryWeeklyMaturity:      15.0, // 15% weekly maturity fine
			CohCibFinesGracePeriodEntryMonthlyAmortization: 25.0, // 25% monthly amortization fine
			CohCibFinesGracePeriodEntryMonthlyMaturity:     35.0, // 35% monthly maturity fine
			ShowInGeneralLedgerSourceWithdraw:              true,
			ShowInGeneralLedgerSourceDeposit:               true,
			ShowInGeneralLedgerSourceJournal:               true,
			ShowInGeneralLedgerSourcePayment:               true,
			ShowInGeneralLedgerSourceAdjustment:            true,
			ShowInGeneralLedgerSourceJournalVoucher:        true,
			ShowInGeneralLedgerSourceCheckVoucher:          true,
			OtherInformationOfAnAccount:                    OIOANone,
			Icon:                                           "Badge Exclamation",
		},
		{
			CreatedAt:                    now,
			CreatedByID:                  userID,
			UpdatedAt:                    now,
			UpdatedByID:                  userID,
			OrganizationID:               organizationID,
			BranchID:                     branchID,
			CurrencyID:                   &currency.ID,
			Name:                         "Administrative Fines",
			Description:                  "Administrative fines for procedural violations and documentation errors.",
			Type:                         AccountTypeFines,
			MinAmount:                    0.00,
			MaxAmount:                    100.00,
			InterestStandard:             0.0,
			FinesAmort:                   1.5, // 1.5% fine on amortization
			FinesMaturity:                4.0, // 4% fine on maturity
			FinesGracePeriodAmortization: 10,  // 10 days grace period for amortization fines
			FinesGracePeriodMaturity:     20,  // 20 days grace period for maturity fines
			AdditionalGracePeriod:        5,   // 5 additional days
			NoGracePeriodDaily:           false,
			GeneralLedgerType:            GLTypeRevenue,
			ComputationType:              Straight,
			Index:                        502,
			CohCibFinesGracePeriodEntryDailyAmortization:   0.5,  // 0.5% daily amortization fine
			CohCibFinesGracePeriodEntryDailyMaturity:       1.0,  // 1% daily maturity fine
			CohCibFinesGracePeriodEntryWeeklyAmortization:  2.5,  // 2.5% weekly amortization fine
			CohCibFinesGracePeriodEntryWeeklyMaturity:      5.0,  // 5% weekly maturity fine
			CohCibFinesGracePeriodEntryMonthlyAmortization: 8.0,  // 8% monthly amortization fine
			CohCibFinesGracePeriodEntryMonthlyMaturity:     12.0, // 12% monthly maturity fine
			ShowInGeneralLedgerSourceWithdraw:              true,
			ShowInGeneralLedgerSourceDeposit:               true,
			ShowInGeneralLedgerSourceJournal:               true,
			ShowInGeneralLedgerSourcePayment:               true,
			ShowInGeneralLedgerSourceAdjustment:            true,
			ShowInGeneralLedgerSourceJournalVoucher:        true,
			ShowInGeneralLedgerSourceCheckVoucher:          true,
			OtherInformationOfAnAccount:                    OIOANone,
			Icon:                                           "Document File Fill",
		},
	}

	// Create all standalone fines accounts
	for _, finesAccount := range standaloneFinesAccounts {
		if err := m.AccountManager.CreateWithTx(context, tx, finesAccount); err != nil {
			return eris.Wrapf(err, "failed to seed standalone fines account %s", finesAccount.Name)
		}
	}

	// Create additional standalone Interest accounts with different configurations
	standaloneInterestAccounts := []*Account{
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			CurrencyID:                              &currency.ID,
			Name:                                    "General Interest Income",
			Icon:                                    "Trend Up",
			Description:                             "General interest income from various cooperative investments and deposits.",
			Type:                                    AccountTypeInterest,
			MinAmount:                               0.00,
			MaxAmount:                               100.00, // Max percentage is 100%
			InterestStandard:                        2.0,    // 2% interest standard
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Diminishing,
			Index:                                   600,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
		},
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			CurrencyID:                              &currency.ID,
			Name:                                    "Penalty Interest",
			Icon:                                    "Arrow Trend Up",
			Description:                             "Interest penalties for overdue accounts and late payments.",
			Type:                                    AccountTypeInterest,
			MinAmount:                               0.00,
			MaxAmount:                               100.00, // Max percentage is 100%
			InterestStandard:                        5.0,    // 5% penalty interest
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   601,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
		},
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			CurrencyID:                              &currency.ID,
			Name:                                    "Investment Interest",
			Icon:                                    "Pie Chart",
			Description:                             "Interest income from long-term investments and financial instruments.",
			Type:                                    AccountTypeInterest,
			MinAmount:                               0.00,
			MaxAmount:                               100.00, // Max percentage is 100%
			InterestStandard:                        3.5,    // 3.5% investment interest
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         DiminishingStraight,
			Index:                                   602,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
		},
	}

	// Create additional standalone SVF accounts with different configurations
	standaloneSVFAccounts := []*Account{
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			CurrencyID:                              &currency.ID,
			Name:                                    "General Service Fee",
			Icon:                                    "Ticket",
			Description:                             "General service fees for various cooperative services and transactions.",
			Type:                                    AccountTypeSVFLedger,
			MinAmount:                               0.00,
			MaxAmount:                               100.00, // Max percentage is 100%
			InterestStandard:                        1.0,    // 1% service fee standard
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   700,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
		},
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			CurrencyID:                              &currency.ID,
			Name:                                    "Processing Service Fee",
			Icon:                                    "Wrench Icon",
			Description:                             "Service fees for loan processing, account opening, and administrative services.",
			Type:                                    AccountTypeSVFLedger,
			MinAmount:                               0.00,
			MaxAmount:                               100.00, // Max percentage is 100%
			InterestStandard:                        2.0,    // 2% processing fee standard
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Diminishing,
			Index:                                   701,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
		},
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			CurrencyID:                              &currency.ID,
			Name:                                    "Maintenance Service Fee",
			Icon:                                    "Gear",
			Description:                             "Monthly and annual maintenance service fees for account upkeep and services.",
			Type:                                    AccountTypeSVFLedger,
			MinAmount:                               0.00,
			MaxAmount:                               100.00, // Max percentage is 100%
			InterestStandard:                        0.5,    // 0.5% maintenance fee standard
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         DiminishingStraight,
			Index:                                   702,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
		},
	}

	// Create all standalone interest accounts
	for _, interestAccount := range standaloneInterestAccounts {
		if err := m.AccountManager.CreateWithTx(context, tx, interestAccount); err != nil {
			return eris.Wrapf(err, "failed to seed standalone interest account %s", interestAccount.Name)
		}
	}

	// Create all standalone SVF accounts
	for _, svfAccount := range standaloneSVFAccounts {
		if err := m.AccountManager.CreateWithTx(context, tx, svfAccount); err != nil {
			return eris.Wrapf(err, "failed to seed standalone SVF account %s", svfAccount.Name)
		}
	}

	paidUpShareCapital := &Account{
		CreatedAt:                         now,
		CreatedByID:                       userID,
		UpdatedAt:                         now,
		UpdatedByID:                       userID,
		OrganizationID:                    organizationID,
		BranchID:                          branchID,
		CurrencyID:                        &currency.ID,
		Name:                              "Paid Up Share Capital",
		Icon:                              "Star",
		Description:                       "Member's share capital contribution representing ownership stake in the cooperative.",
		Type:                              AccountTypeOther,
		MinAmount:                         100.00,
		MaxAmount:                         1000000.00,
		InterestStandard:                  0.0,
		GeneralLedgerType:                 GLTypeEquity,
		ComputationType:                   Straight,
		Index:                             10,
		PaidUpShareCapital:                true,
		ShowInGeneralLedgerSourceWithdraw: true,
		ShowInGeneralLedgerSourceDeposit:  true,
		ShowInGeneralLedgerSourceJournal:  true,
		ShowInGeneralLedgerSourcePayment:  true,
	}
	if err := m.AccountManager.CreateWithTx(context, tx, paidUpShareCapital); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", paidUpShareCapital.Name)
	}

	// Create essential payment types for account seeding
	var cashOnHandPaymentType *PaymentType

	// Try to find existing Cash On Hand payment type
	cashOnHandPaymentType, _ = m.PaymentTypeManager.FindOne(context, &PaymentType{
		OrganizationID: organizationID,
		BranchID:       branchID,
		Name:           "Cash On Hand",
	})

	// If Cash On Hand payment type doesn't exist, create it
	if cashOnHandPaymentType == nil {
		cashOnHandPaymentType = &PaymentType{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Cash On Hand",
			Description:    "Cash available at the branch for immediate use.",
			Type:           PaymentTypeCash,
			NumberOfDays:   0,
		}

		if err := m.PaymentTypeManager.CreateWithTx(context, tx, cashOnHandPaymentType); err != nil {
			return eris.Wrapf(err, "failed to seed payment type %s", cashOnHandPaymentType.Name)
		}

		// Set this payment type as the default in user organization settings
		userOrganization, err := m.UserOrganizationManager.FindOne(context, &UserOrganization{
			UserID:         userID,
			OrganizationID: organizationID,
			BranchID:       &branchID,
		})
		if err != nil {
			return eris.Wrap(err, "failed to find user organization for setting default payment type")
		}
		userOrganization.SettingsPaymentTypeDefaultValueID = &cashOnHandPaymentType.ID
		if err := m.UserOrganizationManager.UpdateByIDWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
			return eris.Wrap(err, "failed to update user organization with default payment type")
		}

		// Create additional payment types
		paymentTypes := []*PaymentType{
			// Cash types
			{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Forward Cash On Hand",
				Description:    "Physical cash received and forwarded for transactions.",
				NumberOfDays:   0,
				Type:           PaymentTypeCash,
			},
			{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Petty Cash",
				Description:    "Small amount of cash for minor expenses.",
				NumberOfDays:   0,
				Type:           PaymentTypeCash,
			},
			// Online types
			{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "E-Wallet",
				Description:    "Digital wallet for online payments.",
				NumberOfDays:   0,
				Type:           PaymentTypeOnline,
			},
			{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "E-Bank",
				Description:    "Online banking transfer.",
				NumberOfDays:   0,
				Type:           PaymentTypeOnline,
			},
			{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "GCash",
				Description:    "GCash mobile wallet payment.",
				NumberOfDays:   0,
				Type:           PaymentTypeOnline,
			},
			// Check/Bank types
			{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Cheque",
				Description:    "Payment via cheque/check.",
				NumberOfDays:   3,
				Type:           PaymentTypeCheck,
			},
			{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Bank Transfer",
				Description:    "Direct bank-to-bank transfer.",
				NumberOfDays:   1,
				Type:           PaymentTypeCheck,
			},
			{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Manager's Check",
				Description:    "Bank-issued check for secure payments.",
				NumberOfDays:   2,
				Type:           PaymentTypeCheck,
			},
			// Adjustment types
			{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Manual Adjustment",
				Description:    "Manual adjustments for corrections and reconciliation.",
				NumberOfDays:   0,
				Type:           PaymentTypeAdjustment,
			},
			{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Adjustment Entry",
				Description:    "Manual adjustments for corrections and reconciliation.",
				NumberOfDays:   0,
				Type:           PaymentTypeAdjustment,
			},
		}

		for _, data := range paymentTypes {
			if err := m.PaymentTypeManager.CreateWithTx(context, tx, data); err != nil {
				return eris.Wrapf(err, "failed to seed payment type %s", data.Name)
			}
		}
	}

	cashOnHand := &Account{
		CreatedAt:                               now,
		CreatedByID:                             userID,
		UpdatedAt:                               now,
		UpdatedByID:                             userID,
		OrganizationID:                          organizationID,
		BranchID:                                branchID,
		CurrencyID:                              &currency.ID,
		DefaultPaymentTypeID:                    &cashOnHandPaymentType.ID,
		Name:                                    "Cash on Hand",
		Icon:                                    "Hand Coins",
		Description:                             "Physical cash available at the branch for daily operations and transactions.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               10000000.00,
		InterestStandard:                        0.0,
		GeneralLedgerType:                       GLTypeAssets,
		ComputationType:                         Straight,
		Index:                                   11,
		CashOnHand:                              true,
		ShowInGeneralLedgerSourceWithdraw:       false,
		ShowInGeneralLedgerSourceDeposit:        false,
		ShowInGeneralLedgerSourceJournal:        true,
		ShowInGeneralLedgerSourcePayment:        true,
		ShowInGeneralLedgerSourceAdjustment:     true,
		ShowInGeneralLedgerSourceJournalVoucher: true,
		ShowInGeneralLedgerSourceCheckVoucher:   true,
		CashAndCashEquivalence:                  true,

		OtherInformationOfAnAccount: OIOACashOnHand,
	}

	if err := m.AccountManager.CreateWithTx(context, tx, cashOnHand); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", cashOnHand.Name)
	}

	// Cash in Bank Account
	cashInBank := &Account{
		CreatedAt:                               now,
		CreatedByID:                             userID,
		UpdatedAt:                               now,
		UpdatedByID:                             userID,
		OrganizationID:                          organizationID,
		BranchID:                                branchID,
		CurrencyID:                              &currency.ID,
		DefaultPaymentTypeID:                    &cashOnHandPaymentType.ID,
		Name:                                    "Cash in Bank",
		Icon:                                    "Bank",
		Description:                             "Funds deposited in bank accounts for secure storage and banking transactions.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               50000000.00,
		InterestStandard:                        0.0,
		GeneralLedgerType:                       GLTypeAssets,
		ComputationType:                         Straight,
		Index:                                   12,
		CashOnHand:                              false,
		ShowInGeneralLedgerSourceWithdraw:       true,
		ShowInGeneralLedgerSourceDeposit:        true,
		ShowInGeneralLedgerSourceJournal:        true,
		ShowInGeneralLedgerSourcePayment:        true,
		ShowInGeneralLedgerSourceAdjustment:     true,
		ShowInGeneralLedgerSourceJournalVoucher: true,
		ShowInGeneralLedgerSourceCheckVoucher:   true,
		CashAndCashEquivalence:                  true,
		OtherInformationOfAnAccount:             OIOACashInBank,
	}

	if err := m.AccountManager.CreateWithTx(context, tx, cashInBank); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", cashInBank.Name)
	}

	// Cash Online Account (Digital Wallets, Online Banking)
	cashOnline := &Account{
		CreatedAt:                               now,
		CreatedByID:                             userID,
		UpdatedAt:                               now,
		UpdatedByID:                             userID,
		OrganizationID:                          organizationID,
		BranchID:                                branchID,
		DefaultPaymentTypeID:                    &cashOnHandPaymentType.ID,
		Name:                                    "Cash Online",
		Icon:                                    "Smartphone",
		Description:                             "Digital funds available through online banking platforms and digital wallets.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               10000000.00,
		InterestStandard:                        0.0,
		GeneralLedgerType:                       GLTypeAssets,
		ComputationType:                         Straight,
		Index:                                   13,
		CashOnHand:                              false,
		ShowInGeneralLedgerSourceWithdraw:       true,
		ShowInGeneralLedgerSourceDeposit:        true,
		ShowInGeneralLedgerSourceJournal:        true,
		ShowInGeneralLedgerSourcePayment:        true,
		ShowInGeneralLedgerSourceAdjustment:     true,
		ShowInGeneralLedgerSourceJournalVoucher: true,
		ShowInGeneralLedgerSourceCheckVoucher:   true,
		CashAndCashEquivalence:                  true,
		OtherInformationOfAnAccount:             OIOANone,
		CurrencyID:                              &currency.ID,
	}

	if err := m.AccountManager.CreateWithTx(context, tx, cashOnline); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", cashOnline.Name)
	}

	// Petty Cash Account
	pettyCash := &Account{
		CreatedAt:                               now,
		CreatedByID:                             userID,
		UpdatedAt:                               now,
		UpdatedByID:                             userID,
		OrganizationID:                          organizationID,
		BranchID:                                branchID,
		DefaultPaymentTypeID:                    &cashOnHandPaymentType.ID,
		Name:                                    "Petty Cash",
		Icon:                                    "Wallet",
		Description:                             "Small amount of cash kept on hand for minor expenses and incidental purchases.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               100000.00,
		InterestStandard:                        0.0,
		GeneralLedgerType:                       GLTypeAssets,
		ComputationType:                         Straight,
		Index:                                   14,
		CashOnHand:                              true,
		ShowInGeneralLedgerSourceWithdraw:       true,
		ShowInGeneralLedgerSourceDeposit:        true,
		ShowInGeneralLedgerSourceJournal:        true,
		ShowInGeneralLedgerSourcePayment:        true,
		ShowInGeneralLedgerSourceAdjustment:     true,
		ShowInGeneralLedgerSourceJournalVoucher: true,
		ShowInGeneralLedgerSourceCheckVoucher:   true,
		CashAndCashEquivalence:                  true,
		OtherInformationOfAnAccount:             OIOANone,
		CurrencyID:                              &currency.ID,
	}

	if err := m.AccountManager.CreateWithTx(context, tx, pettyCash); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", pettyCash.Name)
	}

	// Cash in Transit Account
	cashInTransit := &Account{
		CreatedAt:                               now,
		CreatedByID:                             userID,
		UpdatedAt:                               now,
		UpdatedByID:                             userID,
		OrganizationID:                          organizationID,
		BranchID:                                branchID,
		DefaultPaymentTypeID:                    &cashOnHandPaymentType.ID,
		Name:                                    "Cash in Transit",
		Icon:                                    "Running",
		Description:                             "Cash deposits or transfers that are in process but not yet cleared or posted.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               5000000.00,
		InterestStandard:                        0.0,
		GeneralLedgerType:                       GLTypeAssets,
		ComputationType:                         Straight,
		Index:                                   15,
		CashOnHand:                              false,
		ShowInGeneralLedgerSourceWithdraw:       true,
		ShowInGeneralLedgerSourceDeposit:        true,
		ShowInGeneralLedgerSourceJournal:        true,
		ShowInGeneralLedgerSourcePayment:        true,
		ShowInGeneralLedgerSourceAdjustment:     true,
		ShowInGeneralLedgerSourceJournalVoucher: true,
		ShowInGeneralLedgerSourceCheckVoucher:   true,
		CashAndCashEquivalence:                  true,
		OtherInformationOfAnAccount:             OIOANone,
		CurrencyID:                              &currency.ID,
	}

	if err := m.AccountManager.CreateWithTx(context, tx, cashInTransit); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", cashInTransit.Name)
	}

	// Foreign Currency Cash Account
	foreignCurrencyCash := &Account{
		CreatedAt:                               now,
		CreatedByID:                             userID,
		UpdatedAt:                               now,
		UpdatedByID:                             userID,
		OrganizationID:                          organizationID,
		BranchID:                                branchID,
		DefaultPaymentTypeID:                    &cashOnHandPaymentType.ID,
		Name:                                    "Foreign Currency Cash",
		Icon:                                    "Globe Asia",
		Description:                             "Cash holdings in foreign currencies for international transactions and exchange.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               2000000.00,
		InterestStandard:                        0.0,
		GeneralLedgerType:                       GLTypeAssets,
		ComputationType:                         Straight,
		Index:                                   16,
		CashOnHand:                              true,
		ShowInGeneralLedgerSourceWithdraw:       true,
		ShowInGeneralLedgerSourceDeposit:        true,
		ShowInGeneralLedgerSourceJournal:        true,
		ShowInGeneralLedgerSourcePayment:        true,
		ShowInGeneralLedgerSourceAdjustment:     true,
		ShowInGeneralLedgerSourceJournalVoucher: true,
		ShowInGeneralLedgerSourceCheckVoucher:   true,
		CashAndCashEquivalence:                  true,
		CurrencyID:                              &currency.ID,
		OtherInformationOfAnAccount:             OIOANone,
	}

	if err := m.AccountManager.CreateWithTx(context, tx, foreignCurrencyCash); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", foreignCurrencyCash.Name)
	}

	// Cash Equivalents - Money Market Account
	moneyMarketFund := &Account{
		CreatedAt:                               now,
		CreatedByID:                             userID,
		UpdatedAt:                               now,
		UpdatedByID:                             userID,
		OrganizationID:                          organizationID,
		BranchID:                                branchID,
		DefaultPaymentTypeID:                    &cashOnHandPaymentType.ID,
		Name:                                    "Money Market Fund",
		Icon:                                    "Chart Bar",
		Description:                             "Short-term, highly liquid investments that can be quickly converted to cash.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               20000000.00,
		InterestStandard:                        1.5,
		GeneralLedgerType:                       GLTypeAssets,
		ComputationType:                         Diminishing,
		Index:                                   17,
		CashOnHand:                              false,
		ShowInGeneralLedgerSourceWithdraw:       true,
		ShowInGeneralLedgerSourceDeposit:        true,
		ShowInGeneralLedgerSourceJournal:        true,
		ShowInGeneralLedgerSourcePayment:        true,
		ShowInGeneralLedgerSourceAdjustment:     true,
		ShowInGeneralLedgerSourceJournalVoucher: true,
		ShowInGeneralLedgerSourceCheckVoucher:   true,
		CashAndCashEquivalence:                  true,
		CurrencyID:                              &currency.ID,
		OtherInformationOfAnAccount:             OIOANone,
	}

	if err := m.AccountManager.CreateWithTx(context, tx, moneyMarketFund); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", moneyMarketFund.Name)
	}

	// Treasury Bills (Short-term)
	treasuryBills := &Account{
		CreatedAt:                               now,
		CreatedByID:                             userID,
		UpdatedAt:                               now,
		UpdatedByID:                             userID,
		OrganizationID:                          organizationID,
		BranchID:                                branchID,
		DefaultPaymentTypeID:                    &cashOnHandPaymentType.ID,
		Name:                                    "Treasury Bills",
		Icon:                                    "Document File Fill",
		Description:                             "Short-term government securities with maturity of less than one year.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               15000000.00,
		InterestStandard:                        2.0,
		GeneralLedgerType:                       GLTypeAssets,
		ComputationType:                         Diminishing,
		Index:                                   18,
		CashOnHand:                              false,
		ShowInGeneralLedgerSourceWithdraw:       true,
		ShowInGeneralLedgerSourceDeposit:        true,
		ShowInGeneralLedgerSourceJournal:        true,
		ShowInGeneralLedgerSourcePayment:        true,
		ShowInGeneralLedgerSourceAdjustment:     true,
		ShowInGeneralLedgerSourceJournalVoucher: true,
		ShowInGeneralLedgerSourceCheckVoucher:   true,
		CashAndCashEquivalence:                  true,
		CurrencyID:                              &currency.ID,
		OtherInformationOfAnAccount:             OIOANone,
	}

	if err := m.AccountManager.CreateWithTx(context, tx, treasuryBills); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", treasuryBills.Name)
	}

	// Fee Accounts - Other Type
	feeAccounts := []*Account{
		// Service Fees
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Service Fee",
			Description:                             "General service fees charged for account maintenance and banking services.",
			Type:                                    AccountTypeOther,
			Icon:                                    "Receipt",
			MinAmount:                               0.00,
			MaxAmount:                               10000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   19,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Transaction Fees
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Transaction Fee",
			Description:                             "Fees charged for various transaction services including transfers and withdrawals.",
			Type:                                    AccountTypeOther,
			Icon:                                    "Transaction Dollar",
			MinAmount:                               0.00,
			MaxAmount:                               1000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   20,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Loan Processing Fee
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Loan Processing Fee",
			Description:                             "One-time fee charged for loan application processing and documentation.",
			Type:                                    AccountTypeOther,
			Icon:                                    "Document File Fill",
			MinAmount:                               0.00,
			MaxAmount:                               50000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   21,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Passbook Fee
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Passbook Fee",
			Description:                             "Fee for issuing new passbooks and passbook replacement services.",
			Type:                                    AccountTypeOther,
			Icon:                                    "Book",
			MinAmount:                               0.00,
			MaxAmount:                               500.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   22,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// ATM Fee
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "ATM Fee",
			Description:                             "Fees charged for ATM usage, card issuance, and ATM-related services.",
			Type:                                    AccountTypeOther,
			Icon:                                    "Credit Card",
			MinAmount:                               0.00,
			MaxAmount:                               200.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   23,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Check Processing Fee
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Check Processing Fee",
			Icon:                                    "Receipt",
			Description:                             "Fees for check processing, clearance, and checkbook issuance services.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               1000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   24,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Documentation Fee
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Documentation Fee",
			Icon:                                    "Document File Fill",
			Description:                             "Fee for preparing legal documents, certificates, and official statements.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               2000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   25,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Late Payment Fee
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Late Payment Fee",
			Icon:                                    "Warning",
			Description:                             "Penalty fees charged for late loan payments and overdue account obligations.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               5000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   26,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Account Closure Fee
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Account Closure Fee",
			Icon:                                    "User Lock",
			Description:                             "Fee charged for closing accounts and terminating membership services.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               1000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   27,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Annual Membership Fee
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Annual Membership Fee",
			Icon:                                    "ID Card",
			Description:                             "Yearly membership fee for maintaining cooperative membership status.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               5000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   28,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Insurance Premium Fee
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Insurance Premium Fee",
			Icon:                                    "Shield",
			Description:                             "Insurance premium fees for loan protection and member insurance coverage.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               20000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   29,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Notarial Fee
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Notarial Fee",
			Icon:                                    "Badge Check",
			Description:                             "Fee for notarial services and document authentication requirements.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               3000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   30,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
	}

	// Operational Expense Accounts - Other Type
	operationalAccounts := []*Account{
		// Computer and IT Maintenance
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Computer Maintenance",
			Icon:                                    "Monitor",
			Description:                             "Expenses for computer hardware maintenance, software updates, and IT support services.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               100000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   31,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// General Maintenance
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "General Maintenance",
			Icon:                                    "Gear",
			Description:                             "General maintenance expenses for equipment, furniture, and operational assets.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               150000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   32,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Electricity Bills
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Electricity Bills",
			Icon:                                    "Sparkle",
			Description:                             "Monthly electricity and power consumption expenses for branch operations.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               50000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   33,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Water Bills
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Water Bills",
			Icon:                                    "Globe",
			Description:                             "Monthly water utility expenses for branch facilities and operations.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               20000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   34,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Building Repairs and Maintenance
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Building Repairs",
			Icon:                                    "Building",
			Description:                             "Costs for building repairs, renovations, and structural maintenance work.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               500000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   35,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Internet and Telecommunications
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Internet and Telecommunications",
			Icon:                                    "Globe",
			Description:                             "Monthly internet, phone, and communication service expenses.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               30000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   36,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Office Supplies
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Office Supplies",
			Icon:                                    "Pencil Outline",
			Description:                             "Expenses for office supplies, stationery, and consumable materials.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               25000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   37,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Security Services
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Security Services",
			Icon:                                    "Shield",
			Description:                             "Expenses for security guards, surveillance systems, and safety equipment.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               80000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   38,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Cleaning Services
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Cleaning Services",
			Icon:                                    "Sparkle",
			Description:                             "Expenses for janitorial services, cleaning supplies, and facility sanitation.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               40000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   39,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Professional Services
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Professional Services",
			Icon:                                    "Brief Case",
			Description:                             "Fees for legal, accounting, consulting, and other professional services.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               200000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   40,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Vehicle Maintenance
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Vehicle Maintenance",
			Icon:                                    "Gear",
			Description:                             "Expenses for company vehicle maintenance, fuel, and transportation costs.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               60000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   41,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Equipment Rental
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Equipment Rental",
			Icon:                                    "Settings",
			Description:                             "Rental expenses for equipment, machinery, and temporary facility needs.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               100000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   42,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Training and Development
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Training and Development",
			Icon:                                    "Graduation Cap",
			Description:                             "Expenses for employee training, seminars, workshops, and professional development.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               75000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   43,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Marketing and Advertising
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Marketing and Advertising",
			Icon:                                    "Sparkle",
			Description:                             "Expenses for promotional activities, advertising campaigns, and marketing materials.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               100000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   44,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Travel and Accommodation
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Travel and Accommodation",
			Icon:                                    "Navigation",
			Description:                             "Business travel expenses including transportation, lodging, and meal allowances.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               80000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   45,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Government Fees and Permits
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Government Fees and Permits",
			Icon:                                    "Building Cog",
			Description:                             "Expenses for business permits, licenses, regulatory fees, and government compliance.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               50000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   46,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Medical and Health Services
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Medical and Health Services",
			Icon:                                    "Shield Check",
			Description:                             "Expenses for employee health benefits, medical services, and workplace health programs.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               150000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   47,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Waste Management
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Waste Management",
			Icon:                                    "Recycle",
			Description:                             "Expenses for garbage collection, waste disposal, and environmental compliance services.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               15000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   48,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
		// Emergency Expenses
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Emergency Expenses",
			Icon:                                    "Warning",
			Description:                             "Unexpected expenses for emergency repairs, urgent purchases, and crisis management.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               200000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   49,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             OIOANone,
			CurrencyID:                              &currency.ID,
		},
	}

	// Additional Cooperative-Specific Accounts
	cooperativeAccounts := []*Account{
		// === EQUITY ACCOUNTS ===
		// Retained Earnings
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Retained Earnings",
			Icon:                                    "PiggyBank",
			Description:                             "Accumulated profits retained for reinvestment in the cooperative.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               50000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeEquity,
			ComputationType:                         Straight,
			Index:                                   50,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Patronage Refund Payable
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Patronage Refund Payable",
			Icon:                                    "Hand Drop Coins",
			Description:                             "Profits to be distributed to members based on their patronage.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               10000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeLiabilities,
			ComputationType:                         Straight,
			Index:                                   51,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Member Equity Withdrawals
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Member Equity Withdrawals",
			Icon:                                    "Hand Withdraw",
			Description:                             "Account for tracking member equity withdrawals and distributions.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               5000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeEquity,
			ComputationType:                         Straight,
			Index:                                   52,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// === REVENUE ACCOUNTS ===
		// Dividend Income
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Dividend Income",
			Icon:                                    "Money Trend",
			Description:                             "Income from investments and dividend distributions.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               2000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   53,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Other Income
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Other Income",
			Icon:                                    "Money",
			Description:                             "Miscellaneous income not categorized elsewhere.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               1000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeRevenue,
			ComputationType:                         Straight,
			Index:                                   54,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// === ASSET ACCOUNTS ===
		// Accounts Receivable
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Accounts Receivable",
			Icon:                                    "Receive Money",
			Description:                             "Money owed to the cooperative by members and other parties.",
			Type:                                    AccountTypeARLedger,
			MinAmount:                               0.00,
			MaxAmount:                               10000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeAssets,
			ComputationType:                         Straight,
			Index:                                   55,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Allowance for Doubtful Accounts
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Allowance for Doubtful Accounts",
			Icon:                                    "Question Circle",
			Description:                             "Reserve for potential uncollectible receivables.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               5000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeAssets,
			ComputationType:                         Straight,
			Index:                                   56,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Inventory
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Inventory",
			Icon:                                    "Store",
			Description:                             "Goods and supplies held for sale or use in operations.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               3000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeAssets,
			ComputationType:                         Straight,
			Index:                                   57,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Prepaid Expenses
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Prepaid Expenses",
			Icon:                                    "Calendar Check",
			Description:                             "Expenses paid in advance for future periods.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               500000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeAssets,
			ComputationType:                         Straight,
			Index:                                   58,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Fixed Assets - Land
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Land",
			Icon:                                    "Park",
			Description:                             "Real estate and land owned by the cooperative.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               50000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeAssets,
			ComputationType:                         Straight,
			Index:                                   59,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Fixed Assets - Building
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Building",
			Icon:                                    "Building",
			Description:                             "Buildings and structures owned by the cooperative.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               30000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeAssets,
			ComputationType:                         Straight,
			Index:                                   60,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Fixed Assets - Equipment
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Equipment",
			Icon:                                    "Gear",
			Description:                             "Machinery, tools, and equipment used in operations.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               5000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeAssets,
			ComputationType:                         Straight,
			Index:                                   61,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Fixed Assets - Furniture and Fixtures
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Furniture and Fixtures",
			Icon:                                    "House",
			Description:                             "Office furniture, fixtures, and fittings.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               1000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeAssets,
			ComputationType:                         Straight,
			Index:                                   62,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Accumulated Depreciation
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Accumulated Depreciation",
			Icon:                                    "Trend Down",
			Description:                             "Cumulative depreciation of fixed assets.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               20000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeAssets,
			ComputationType:                         Straight,
			Index:                                   63,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// === LIABILITY ACCOUNTS ===
		// Accounts Payable
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Accounts Payable",
			Icon:                                    "Bill",
			Description:                             "Amounts owed to suppliers and vendors.",
			Type:                                    AccountTypeAPLedger,
			MinAmount:                               0.00,
			MaxAmount:                               5000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeLiabilities,
			ComputationType:                         Straight,
			Index:                                   64,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Accrued Expenses
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Accrued Expenses",
			Icon:                                    "Clock",
			Description:                             "Expenses incurred but not yet paid.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               2000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeLiabilities,
			ComputationType:                         Straight,
			Index:                                   65,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Taxes Payable
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Taxes Payable",
			Icon:                                    "Receipt",
			Description:                             "Taxes owed to government authorities.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               3000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeLiabilities,
			ComputationType:                         Straight,
			Index:                                   66,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Unearned Revenue
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Unearned Revenue",
			Icon:                                    "Calendar",
			Description:                             "Advance payments received for services not yet rendered.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               1000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeLiabilities,
			ComputationType:                         Straight,
			Index:                                   67,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// === EXPENSE ACCOUNTS ===
		// Salaries and Wages
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Salaries and Wages",
			Icon:                                    "Users 3",
			Description:                             "Employee compensation and payroll expenses.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               5000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   68,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Employee Benefits
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Employee Benefits",
			Icon:                                    "Shield Check",
			Description:                             "Health insurance, retirement, and other employee benefits.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               1000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   69,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Depreciation Expense
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Depreciation Expense",
			Icon:                                    "Trend Down",
			Description:                             "Systematic allocation of asset cost over useful life.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               500000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   70,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Bad Debt Expense
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Bad Debt Expense",
			Icon:                                    "Trash",
			Description:                             "Losses from uncollectible receivables.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               1000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   71,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Interest Expense on Borrowings
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Interest Expense on Borrowings",
			Icon:                                    "Percent",
			Description:                             "Interest paid on loans and borrowings.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               2000000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   72,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Audit and Accounting Fees
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Audit and Accounting Fees",
			Icon:                                    "Finance Reports",
			Description:                             "Professional fees for auditing and accounting services.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               300000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   73,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Bank Charges
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Bank Charges",
			Icon:                                    "Bank",
			Description:                             "Bank service fees, transaction charges, and related costs.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               100000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   74,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
		// Donations and Contributions
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			Name:                                    "Donations and Contributions",
			Icon:                                    "Hand Shake Heart",
			Description:                             "Charitable donations and community contributions.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               500000.00,
			InterestStandard:                        0.0,
			GeneralLedgerType:                       GLTypeExpenses,
			ComputationType:                         Straight,
			Index:                                   75,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             OIOANone,
		},
	}

	// Create all cooperative-specific accounts
	for _, coopAccount := range cooperativeAccounts {
		coopAccount.CurrencyID = &currency.ID
		if err := m.AccountManager.CreateWithTx(context, tx, coopAccount); err != nil {
			return eris.Wrapf(err, "failed to seed cooperative account %s", coopAccount.Name)
		}
	}

	// Create all fee accounts
	for _, feeAccount := range feeAccounts {
		feeAccount.CurrencyID = &currency.ID
		if err := m.AccountManager.CreateWithTx(context, tx, feeAccount); err != nil {
			return eris.Wrapf(err, "failed to seed fee account %s", feeAccount.Name)
		}
	}

	// Create all operational expense accounts
	for _, operationalAccount := range operationalAccounts {
		operationalAccount.CurrencyID = &currency.ID
		if err := m.AccountManager.CreateWithTx(context, tx, operationalAccount); err != nil {
			return eris.Wrapf(err, "failed to seed operational account %s", operationalAccount.Name)
		}
	}

	branch.BranchSetting.PaidUpSharedCapitalAccountID = &paidUpShareCapital.ID
	branch.BranchSetting.CashOnHandAccountID = &cashOnHand.ID
	if err := m.BranchSettingManager.UpdateByIDWithTx(context, tx, branch.BranchSetting.ID, branch.BranchSetting); err != nil {
		return eris.Wrap(err, "failed to update branch settings with paid up share capital and cash on hand accounts")
	}

	unbalanced := &UnbalancedAccount{
		CreatedAt:            now,
		CreatedByID:          userID,
		UpdatedAt:            now,
		UpdatedByID:          userID,
		BranchSettingsID:     branch.BranchSetting.ID,
		CurrencyID:           currency.ID,
		AccountForShortageID: cashOnHand.ID,
		AccountForOverageID:  cashOnHand.ID,
	}
	if err := m.UnbalancedAccountManager.CreateWithTx(context, tx, unbalanced); err != nil {
		return eris.Wrap(err, "failed to create unbalanced account for branch")
	}
	// Set default accounting accounts for user organization
	userOrganization, err := m.UserOrganizationManager.FindOne(context, &UserOrganization{
		UserID:         userID,
		OrganizationID: organizationID,
		BranchID:       &branchID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find user organization for seeding accounting default accounts")
	}

	// Set default accounting accounts - using first suitable account for each type
	var regularSavings *Account
	for _, account := range accounts {
		if account.Name == "Regular Savings" {
			regularSavings = account
			break
		}
	}

	if regularSavings != nil {
		// Use Regular Savings as default for all three accounting operations
		userOrganization.SettingsAccountingPaymentDefaultValueID = &regularSavings.ID
		userOrganization.SettingsAccountingDepositDefaultValueID = &regularSavings.ID
		userOrganization.SettingsAccountingWithdrawDefaultValueID = &regularSavings.ID
	}
	return nil
}

// BeforeUpdate hook for Account model to track changes
func (a *Account) BeforeUpdate(tx *gorm.DB) error {
	// Get the original account data before update
	var original Account
	if err := tx.Unscoped().Where("id = ?", a.ID).First(&original).Error; err != nil {
		// If we can't find the original record, skip history creation to avoid blocking the update
		return nil
	}

	// Create history record with the original data
	now := time.Now().UTC()
	history := &AccountHistory{
		AccountID:      a.ID,
		OrganizationID: original.OrganizationID,
		BranchID:       original.BranchID,
		CreatedByID:    a.CreatedByID,
		CreatedAt:      now,

		// Copy all original data
		Name:                                original.Name,
		Description:                         original.Description,
		Type:                                original.Type,
		MinAmount:                           original.MinAmount,
		MaxAmount:                           original.MaxAmount,
		Index:                               original.Index,
		IsInternal:                          original.IsInternal,
		CashOnHand:                          original.CashOnHand,
		PaidUpShareCapital:                  original.PaidUpShareCapital,
		ComputationType:                     original.ComputationType,
		FinesAmort:                          original.FinesAmort,
		FinesMaturity:                       original.FinesMaturity,
		InterestStandard:                    original.InterestStandard,
		InterestSecured:                     original.InterestSecured,
		FinesGracePeriodAmortization:        original.FinesGracePeriodAmortization,
		AdditionalGracePeriod:               original.AdditionalGracePeriod,
		NoGracePeriodDaily:                  original.NoGracePeriodDaily,
		FinesGracePeriodMaturity:            original.FinesGracePeriodMaturity,
		YearlySubscriptionFee:               original.YearlySubscriptionFee,
		CutOffDays:                          original.CutOffDays,
		CutOffMonths:                        original.CutOffMonths,
		LumpsumComputationType:              original.LumpsumComputationType,
		InterestFinesComputationDiminishing: original.InterestFinesComputationDiminishing,
		InterestFinesComputationDiminishingStraightYearly: original.InterestFinesComputationDiminishingStraightYearly,
		EarnedUnearnedInterest:                            original.EarnedUnearnedInterest,
		LoanSavingType:                                    original.LoanSavingType,
		InterestDeduction:                                 original.InterestDeduction,
		OtherDeductionEntry:                               original.OtherDeductionEntry,
		InterestSavingTypeDiminishingStraight:             original.InterestSavingTypeDiminishingStraight,
		OtherInformationOfAnAccount:                       original.OtherInformationOfAnAccount,
		GeneralLedgerType:                                 original.GeneralLedgerType,
		HeaderRow:                                         original.HeaderRow,
		CenterRow:                                         original.CenterRow,
		TotalRow:                                          original.TotalRow,
		GeneralLedgerGroupingExcludeAccount:               original.GeneralLedgerGroupingExcludeAccount,
		Icon:                                              original.Icon,
		ShowInGeneralLedgerSourceWithdraw:                 original.ShowInGeneralLedgerSourceWithdraw,
		ShowInGeneralLedgerSourceDeposit:                  original.ShowInGeneralLedgerSourceDeposit,
		ShowInGeneralLedgerSourceJournal:                  original.ShowInGeneralLedgerSourceJournal,
		ShowInGeneralLedgerSourcePayment:                  original.ShowInGeneralLedgerSourcePayment,
		ShowInGeneralLedgerSourceAdjustment:               original.ShowInGeneralLedgerSourceAdjustment,
		ShowInGeneralLedgerSourceJournalVoucher:           original.ShowInGeneralLedgerSourceJournalVoucher,
		ShowInGeneralLedgerSourceCheckVoucher:             original.ShowInGeneralLedgerSourceCheckVoucher,
		CompassionFund:                                    original.CompassionFund,
		CompassionFundAmount:                              original.CompassionFundAmount,
		CashAndCashEquivalence:                            original.CashAndCashEquivalence,
		InterestStandardComputation:                       original.InterestStandardComputation,

		// Foreign key references
		GeneralLedgerDefinitionID:      original.GeneralLedgerDefinitionID,
		FinancialStatementDefinitionID: original.FinancialStatementDefinitionID,
		AccountClassificationID:        original.AccountClassificationID,
		AccountCategoryID:              original.AccountCategoryID,
		MemberTypeID:                   original.MemberTypeID,
		CurrencyID:                     original.CurrencyID,
		DefaultPaymentTypeID:           original.DefaultPaymentTypeID,
		ComputationSheetID:             original.ComputationSheetID,
		LoanAccountID:                  original.LoanAccountID,

		// Grace period entries
		CohCibFinesGracePeriodEntryCashHand:                original.CohCibFinesGracePeriodEntryCashHand,
		CohCibFinesGracePeriodEntryCashInBank:              original.CohCibFinesGracePeriodEntryCashInBank,
		CohCibFinesGracePeriodEntryDailyAmortization:       original.CohCibFinesGracePeriodEntryDailyAmortization,
		CohCibFinesGracePeriodEntryDailyMaturity:           original.CohCibFinesGracePeriodEntryDailyMaturity,
		CohCibFinesGracePeriodEntryWeeklyAmortization:      original.CohCibFinesGracePeriodEntryWeeklyAmortization,
		CohCibFinesGracePeriodEntryWeeklyMaturity:          original.CohCibFinesGracePeriodEntryWeeklyMaturity,
		CohCibFinesGracePeriodEntryMonthlyAmortization:     original.CohCibFinesGracePeriodEntryMonthlyAmortization,
		CohCibFinesGracePeriodEntryMonthlyMaturity:         original.CohCibFinesGracePeriodEntryMonthlyMaturity,
		CohCibFinesGracePeriodEntrySemiMonthlyAmortization: original.CohCibFinesGracePeriodEntrySemiMonthlyAmortization,
		CohCibFinesGracePeriodEntrySemiMonthlyMaturity:     original.CohCibFinesGracePeriodEntrySemiMonthlyMaturity,
		CohCibFinesGracePeriodEntryQuarterlyAmortization:   original.CohCibFinesGracePeriodEntryQuarterlyAmortization,
		CohCibFinesGracePeriodEntryQuarterlyMaturity:       original.CohCibFinesGracePeriodEntryQuarterlyMaturity,
		CohCibFinesGracePeriodEntrySemiAnnualAmortization:  original.CohCibFinesGracePeriodEntrySemiAnnualAmortization,
		CohCibFinesGracePeriodEntrySemiAnnualMaturity:      original.CohCibFinesGracePeriodEntrySemiAnnualMaturity,
		CohCibFinesGracePeriodEntryLumpsumAmortization:     original.CohCibFinesGracePeriodEntryLumpsumAmortization,
		CohCibFinesGracePeriodEntryLumpsumMaturity:         original.CohCibFinesGracePeriodEntryLumpsumMaturity,
	}

	// Save the history record
	if err := tx.Create(history).Error; err != nil {
		return err
	}

	return nil
}

// AfterCreate hook for Account model to create initial history record
func (a *Account) AfterCreate(tx *gorm.DB) error {
	history := &AccountHistory{
		AccountID:      a.ID,
		OrganizationID: a.OrganizationID,
		BranchID:       a.BranchID,
		CreatedByID:    a.CreatedByID,

		// Copy all current data
		Name:                                a.Name,
		Description:                         a.Description,
		Type:                                a.Type,
		MinAmount:                           a.MinAmount,
		MaxAmount:                           a.MaxAmount,
		Index:                               a.Index,
		IsInternal:                          a.IsInternal,
		CashOnHand:                          a.CashOnHand,
		PaidUpShareCapital:                  a.PaidUpShareCapital,
		ComputationType:                     a.ComputationType,
		FinesAmort:                          a.FinesAmort,
		FinesMaturity:                       a.FinesMaturity,
		InterestStandard:                    a.InterestStandard,
		InterestSecured:                     a.InterestSecured,
		FinesGracePeriodAmortization:        a.FinesGracePeriodAmortization,
		AdditionalGracePeriod:               a.AdditionalGracePeriod,
		NoGracePeriodDaily:                  a.NoGracePeriodDaily,
		FinesGracePeriodMaturity:            a.FinesGracePeriodMaturity,
		YearlySubscriptionFee:               a.YearlySubscriptionFee,
		CutOffDays:                          a.CutOffDays,
		CutOffMonths:                        a.CutOffMonths,
		LumpsumComputationType:              a.LumpsumComputationType,
		InterestFinesComputationDiminishing: a.InterestFinesComputationDiminishing,
		InterestFinesComputationDiminishingStraightYearly: a.InterestFinesComputationDiminishingStraightYearly,
		EarnedUnearnedInterest:                            a.EarnedUnearnedInterest,
		LoanSavingType:                                    a.LoanSavingType,
		InterestDeduction:                                 a.InterestDeduction,
		OtherDeductionEntry:                               a.OtherDeductionEntry,
		InterestSavingTypeDiminishingStraight:             a.InterestSavingTypeDiminishingStraight,
		OtherInformationOfAnAccount:                       a.OtherInformationOfAnAccount,
		GeneralLedgerType:                                 a.GeneralLedgerType,
		HeaderRow:                                         a.HeaderRow,
		CenterRow:                                         a.CenterRow,
		TotalRow:                                          a.TotalRow,
		GeneralLedgerGroupingExcludeAccount:               a.GeneralLedgerGroupingExcludeAccount,
		Icon:                                              a.Icon,
		ShowInGeneralLedgerSourceWithdraw:                 a.ShowInGeneralLedgerSourceWithdraw,
		ShowInGeneralLedgerSourceDeposit:                  a.ShowInGeneralLedgerSourceDeposit,
		ShowInGeneralLedgerSourceJournal:                  a.ShowInGeneralLedgerSourceJournal,
		ShowInGeneralLedgerSourcePayment:                  a.ShowInGeneralLedgerSourcePayment,
		ShowInGeneralLedgerSourceAdjustment:               a.ShowInGeneralLedgerSourceAdjustment,
		ShowInGeneralLedgerSourceJournalVoucher:           a.ShowInGeneralLedgerSourceJournalVoucher,
		ShowInGeneralLedgerSourceCheckVoucher:             a.ShowInGeneralLedgerSourceCheckVoucher,
		CompassionFund:                                    a.CompassionFund,
		CompassionFundAmount:                              a.CompassionFundAmount,
		CashAndCashEquivalence:                            a.CashAndCashEquivalence,
		InterestStandardComputation:                       a.InterestStandardComputation,

		// Foreign key references
		GeneralLedgerDefinitionID:      a.GeneralLedgerDefinitionID,
		FinancialStatementDefinitionID: a.FinancialStatementDefinitionID,
		AccountClassificationID:        a.AccountClassificationID,
		AccountCategoryID:              a.AccountCategoryID,
		MemberTypeID:                   a.MemberTypeID,
		CurrencyID:                     a.CurrencyID,
		ComputationSheetID:             a.ComputationSheetID,
		LoanAccountID:                  a.LoanAccountID,

		// Grace period entries
		CohCibFinesGracePeriodEntryCashHand:                a.CohCibFinesGracePeriodEntryCashHand,
		CohCibFinesGracePeriodEntryCashInBank:              a.CohCibFinesGracePeriodEntryCashInBank,
		CohCibFinesGracePeriodEntryDailyAmortization:       a.CohCibFinesGracePeriodEntryDailyAmortization,
		CohCibFinesGracePeriodEntryDailyMaturity:           a.CohCibFinesGracePeriodEntryDailyMaturity,
		CohCibFinesGracePeriodEntryWeeklyAmortization:      a.CohCibFinesGracePeriodEntryWeeklyAmortization,
		CohCibFinesGracePeriodEntryWeeklyMaturity:          a.CohCibFinesGracePeriodEntryWeeklyMaturity,
		CohCibFinesGracePeriodEntryMonthlyAmortization:     a.CohCibFinesGracePeriodEntryMonthlyAmortization,
		CohCibFinesGracePeriodEntryMonthlyMaturity:         a.CohCibFinesGracePeriodEntryMonthlyMaturity,
		CohCibFinesGracePeriodEntrySemiMonthlyAmortization: a.CohCibFinesGracePeriodEntrySemiMonthlyAmortization,
		CohCibFinesGracePeriodEntrySemiMonthlyMaturity:     a.CohCibFinesGracePeriodEntrySemiMonthlyMaturity,
		CohCibFinesGracePeriodEntryQuarterlyAmortization:   a.CohCibFinesGracePeriodEntryQuarterlyAmortization,
		CohCibFinesGracePeriodEntryQuarterlyMaturity:       a.CohCibFinesGracePeriodEntryQuarterlyMaturity,
		CohCibFinesGracePeriodEntrySemiAnnualAmortization:  a.CohCibFinesGracePeriodEntrySemiAnnualAmortization,
		CohCibFinesGracePeriodEntrySemiAnnualMaturity:      a.CohCibFinesGracePeriodEntrySemiAnnualMaturity,
		CohCibFinesGracePeriodEntryLumpsumAmortization:     a.CohCibFinesGracePeriodEntryLumpsumAmortization,
		CohCibFinesGracePeriodEntryLumpsumMaturity:         a.CohCibFinesGracePeriodEntryLumpsumMaturity,
		// In the AccountHistoryToModel function (around line 500):
		CohCibFinesGracePeriodEntryAnnualAmortization: a.CohCibFinesGracePeriodEntryAnnualAmortization,
		CohCibFinesGracePeriodEntryAnnualMaturity:     a.CohCibFinesGracePeriodEntryAnnualMaturity,
		// Add DefaultPaymentTypeID
		DefaultPaymentTypeID: a.DefaultPaymentTypeID,
	}

	// Save the history record
	return tx.Create(history).Error
}

// BeforeDelete hook for Account model to track deletion
func (a *Account) BeforeDelete(tx *gorm.DB) error {
	now := time.Now().UTC()

	// Close any open history records for this account
	if err := tx.Model(&AccountHistory{}).
		Where("account_id = ? AND valid_to IS NULL", a.ID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	// Create deletion history record
	history := &AccountHistory{
		AccountID:      a.ID,
		OrganizationID: a.OrganizationID,
		BranchID:       a.BranchID,
		CreatedByID:    a.CreatedByID,
		// In the AccountHistoryToModel function (around line 500):
		CohCibFinesGracePeriodEntryAnnualAmortization:  a.CohCibFinesGracePeriodEntryAnnualAmortization,
		CohCibFinesGracePeriodEntryAnnualMaturity:      a.CohCibFinesGracePeriodEntryAnnualMaturity,
		CohCibFinesGracePeriodEntryLumpsumAmortization: a.CohCibFinesGracePeriodEntryLumpsumAmortization,
		CohCibFinesGracePeriodEntryLumpsumMaturity:     a.CohCibFinesGracePeriodEntryLumpsumMaturity,

		// Copy all current data before deletion
		Name:                                a.Name,
		Description:                         a.Description,
		Type:                                a.Type,
		MinAmount:                           a.MinAmount,
		MaxAmount:                           a.MaxAmount,
		Index:                               a.Index,
		IsInternal:                          a.IsInternal,
		CashOnHand:                          a.CashOnHand,
		PaidUpShareCapital:                  a.PaidUpShareCapital,
		ComputationType:                     a.ComputationType,
		FinesAmort:                          a.FinesAmort,
		FinesMaturity:                       a.FinesMaturity,
		InterestStandard:                    a.InterestStandard,
		InterestSecured:                     a.InterestSecured,
		FinesGracePeriodAmortization:        a.FinesGracePeriodAmortization,
		AdditionalGracePeriod:               a.AdditionalGracePeriod,
		NoGracePeriodDaily:                  a.NoGracePeriodDaily,
		FinesGracePeriodMaturity:            a.FinesGracePeriodMaturity,
		YearlySubscriptionFee:               a.YearlySubscriptionFee,
		CutOffDays:                          a.CutOffDays,
		CutOffMonths:                        a.CutOffMonths,
		LumpsumComputationType:              a.LumpsumComputationType,
		InterestFinesComputationDiminishing: a.InterestFinesComputationDiminishing,
		InterestFinesComputationDiminishingStraightYearly: a.InterestFinesComputationDiminishingStraightYearly,
		EarnedUnearnedInterest:                            a.EarnedUnearnedInterest,
		LoanSavingType:                                    a.LoanSavingType,
		InterestDeduction:                                 a.InterestDeduction,
		OtherDeductionEntry:                               a.OtherDeductionEntry,
		InterestSavingTypeDiminishingStraight:             a.InterestSavingTypeDiminishingStraight,
		OtherInformationOfAnAccount:                       a.OtherInformationOfAnAccount,
		GeneralLedgerType:                                 a.GeneralLedgerType,
		HeaderRow:                                         a.HeaderRow,
		CenterRow:                                         a.CenterRow,
		TotalRow:                                          a.TotalRow,
		GeneralLedgerGroupingExcludeAccount:               a.GeneralLedgerGroupingExcludeAccount,
		Icon:                                              a.Icon,
		ShowInGeneralLedgerSourceWithdraw:                 a.ShowInGeneralLedgerSourceWithdraw,
		ShowInGeneralLedgerSourceDeposit:                  a.ShowInGeneralLedgerSourceDeposit,
		ShowInGeneralLedgerSourceJournal:                  a.ShowInGeneralLedgerSourceJournal,
		ShowInGeneralLedgerSourcePayment:                  a.ShowInGeneralLedgerSourcePayment,
		ShowInGeneralLedgerSourceAdjustment:               a.ShowInGeneralLedgerSourceAdjustment,
		ShowInGeneralLedgerSourceJournalVoucher:           a.ShowInGeneralLedgerSourceJournalVoucher,
		ShowInGeneralLedgerSourceCheckVoucher:             a.ShowInGeneralLedgerSourceCheckVoucher,
		CompassionFund:                                    a.CompassionFund,
		CompassionFundAmount:                              a.CompassionFundAmount,
		CashAndCashEquivalence:                            a.CashAndCashEquivalence,
		InterestStandardComputation:                       a.InterestStandardComputation,

		// Foreign key references
		GeneralLedgerDefinitionID:      a.GeneralLedgerDefinitionID,
		FinancialStatementDefinitionID: a.FinancialStatementDefinitionID,
		AccountClassificationID:        a.AccountClassificationID,
		AccountCategoryID:              a.AccountCategoryID,
		MemberTypeID:                   a.MemberTypeID,
		CurrencyID:                     a.CurrencyID,
		ComputationSheetID:             a.ComputationSheetID,
		LoanAccountID:                  a.LoanAccountID,

		// Grace period entries
		CohCibFinesGracePeriodEntryCashHand:                a.CohCibFinesGracePeriodEntryCashHand,
		CohCibFinesGracePeriodEntryCashInBank:              a.CohCibFinesGracePeriodEntryCashInBank,
		CohCibFinesGracePeriodEntryDailyAmortization:       a.CohCibFinesGracePeriodEntryDailyAmortization,
		CohCibFinesGracePeriodEntryDailyMaturity:           a.CohCibFinesGracePeriodEntryDailyMaturity,
		CohCibFinesGracePeriodEntryWeeklyAmortization:      a.CohCibFinesGracePeriodEntryWeeklyAmortization,
		CohCibFinesGracePeriodEntryWeeklyMaturity:          a.CohCibFinesGracePeriodEntryWeeklyMaturity,
		CohCibFinesGracePeriodEntryMonthlyAmortization:     a.CohCibFinesGracePeriodEntryMonthlyAmortization,
		CohCibFinesGracePeriodEntryMonthlyMaturity:         a.CohCibFinesGracePeriodEntryMonthlyMaturity,
		CohCibFinesGracePeriodEntrySemiMonthlyAmortization: a.CohCibFinesGracePeriodEntrySemiMonthlyAmortization,
		CohCibFinesGracePeriodEntrySemiMonthlyMaturity:     a.CohCibFinesGracePeriodEntrySemiMonthlyMaturity,
		CohCibFinesGracePeriodEntryQuarterlyAmortization:   a.CohCibFinesGracePeriodEntryQuarterlyAmortization,
		CohCibFinesGracePeriodEntryQuarterlyMaturity:       a.CohCibFinesGracePeriodEntryQuarterlyMaturity,
		CohCibFinesGracePeriodEntrySemiAnnualAmortization:  a.CohCibFinesGracePeriodEntrySemiAnnualAmortization,
		// AccountCurrentBranch
		CohCibFinesGracePeriodEntrySemiAnnualMaturity: a.CohCibFinesGracePeriodEntrySemiAnnualMaturity,
		// Add DefaultPaymentTypeID
		DefaultPaymentTypeID: a.DefaultPaymentTypeID,
		// AccountCurrentBranch
	}

	// Save the deletion history record
	return tx.Create(history).Error
}

// AccountCurrentBranch
func (m *Core) AccountCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*Account, error) {
	return m.AccountManager.Find(context, &Account{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

// AccountLockForUpdate locks the account row for update and returns the locked Account.
// It uses a SELECT ... FOR UPDATE style locking to prevent concurrent modifications.
func (m *Core) AccountLockForUpdate(ctx context.Context, tx *gorm.DB, accountID uuid.UUID) (*Account, error) {
	return m.AccountManager.GetByIDLock(ctx, tx, accountID)
}

// AccountLockWithValidation acquires an account lock and validates that the account
// has not been changed compared to originalAccount. It returns the locked Account.
func (m *Core) AccountLockWithValidation(ctx context.Context, tx *gorm.DB, accountID uuid.UUID, originalAccount *Account) (*Account, error) {
	lockedAccount, err := m.AccountManager.GetByIDLock(ctx, tx, accountID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to acquire account lock")
	}

	// Verify account data hasn't changed since initial check (concurrent modification detection)
	if originalAccount != nil {
		if lockedAccount.OrganizationID != originalAccount.OrganizationID ||
			lockedAccount.BranchID != originalAccount.BranchID ||
			lockedAccount.Type != originalAccount.Type {
			return nil, eris.New("account was modified by another transaction")
		}
	}

	return lockedAccount, nil
}

// LoanAccounts retrieves all loan accounts for a given organization and branch.
func (m *Core) LoanAccounts(ctx context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*Account, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
	}

	return m.AccountManager.FindWithSQL(ctx, filters, []registry.FilterSortSQL{
		{Field: "updated_at", Order: filter.SortOrderDesc},
	})
}

// FindAccountsByTypesAndBranch finds all accounts with specified branch, organization and account types (Fines, Interest, or SVFLedger)
func (m *Core) FindAccountsByTypesAndBranch(ctx context.Context, organizationID uuid.UUID, branchID uuid.UUID, currencyID uuid.UUID) ([]*Account, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "currency_id", Op: registry.OpEq, Value: currencyID},
		{Field: "type", Op: registry.OpIn, Value: []AccountType{
			AccountTypeFines,
			AccountTypeInterest,
			AccountTypeSVFLedger,
		}},
	}
	return m.AccountManager.FindWithSQL(ctx, filters, []registry.FilterSortSQL{
		{Field: "updated_at", Order: filter.SortOrderDesc},
	})
}

// FindAccountsBySpecificType finds all accounts with specified branch, organization and a single account type
func (m *Core) FindAccountsBySpecificType(ctx context.Context, organizationID uuid.UUID, branchID uuid.UUID, accountType AccountType) ([]*Account, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "type", Op: registry.OpEq, Value: accountType},
	}

	return m.AccountManager.FindWithSQL(ctx, filters, []registry.FilterSortSQL{
		{Field: "updated_at", Order: filter.SortOrderDesc},
	})
}

// FindAccountsBySpecificTypeByAccountID finds all accounts with specified branch, organization and a single account ID
func (m *Core) FindLoanAccountsByID(ctx context.Context,
	organizationID uuid.UUID, branchID uuid.UUID, accountID uuid.UUID) ([]*Account, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "loan_account_id", Op: registry.OpEq, Value: accountID},
	}

	accounts, err := m.AccountManager.FindWithSQL(ctx, filters, []registry.FilterSortSQL{
		{Field: "updated_at", Order: filter.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}
	return accounts, nil
}
