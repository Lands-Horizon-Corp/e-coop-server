package model_core

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// --- ENUMS ---

type AccountType string

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

type FinancialStatementType string

const (
	FSTypeAssets      FinancialStatementType = "Assets"
	FSTypeLiabilities FinancialStatementType = "Liabilities"
	FSTypeEquity      FinancialStatementType = "Equity"
	FSTypeRevenue     FinancialStatementType = "Revenue"
	FSTypeExpenses    FinancialStatementType = "Expenses"
)

type LumpsumComputationType string

const (
	LumpsumComputationNone             LumpsumComputationType = "None"
	LumpsumComputationFinesMaturity    LumpsumComputationType = "Compute Fines Maturity"
	LumpsumComputationInterestMaturity LumpsumComputationType = "Compute Interest Maturity / Terms"
	LumpsumComputationAdvanceInterest  LumpsumComputationType = "Compute Advance Interest"
)

type InterestFinesComputationDiminishing string

const (
	IFCDNone                  InterestFinesComputationDiminishing = "None"
	IFCDByAmortization        InterestFinesComputationDiminishing = "By Amortization"
	IFCDByAmortizationDalyArr InterestFinesComputationDiminishing = "By Amortization Daly on Interest Principal + Interest = Fines(Arr)"
)

type InterestFinesComputationDiminishingStraightYearly string

const (
	IFCDSYNone                   InterestFinesComputationDiminishingStraightYearly = "None"
	IFCDSYByDailyInterestBalance InterestFinesComputationDiminishingStraightYearly = "By Daily on Interest based on loan balance by year Principal + Interest Amortization = Fines Fines Grace Period Month end Amortization"
)

type EarnedUnearnedInterest string

const (
	EUITypeNone                    EarnedUnearnedInterest = "None"
	EUITypeByFormula               EarnedUnearnedInterest = "By Formula"
	EUITypeByFormulaActualPay      EarnedUnearnedInterest = "By Formula + Actual Pay"
	EUITypeByAdvanceInterestActual EarnedUnearnedInterest = "By Advance Interest + Actual Pay"
)

type LoanSavingType string

const (
	LSTSeparate                 LoanSavingType = "Separate"
	LSTSingleLedger             LoanSavingType = "Single Ledger"
	LSTSingleLedgerIfNotZero    LoanSavingType = "Single Ledger if Not Zero"
	LSTSingleLedgerSemi1530     LoanSavingType = "Single Ledger Semi (15/30)"
	LSTSingleLedgerSemiMaturity LoanSavingType = "Single Ledger Semi Within Maturity"
)

type InterestDeduction string

const (
	InterestDeductionAbove InterestDeduction = "above"
	InterestDeductionBelow InterestDeduction = "Below"
)

type OtherDeductionEntry string

const (
	OtherDeductionEntryNone       OtherDeductionEntry = "None"
	OtherDeductionEntryHealthCare OtherDeductionEntry = "Health Care"
)

type InterestSavingTypeDiminishingStraight string

const (
	ISTDS_Spread     InterestSavingTypeDiminishingStraight = "Spread"
	ISTDS_1stPayment InterestSavingTypeDiminishingStraight = "1st Payment"
)

type OtherInformationOfAnAccount string

const (
	OIOA_None               OtherInformationOfAnAccount = "None"
	OIOA_Jewely             OtherInformationOfAnAccount = "Jewely"
	OIOA_Grocery            OtherInformationOfAnAccount = "Grocery"
	OIOA_TrackLoanDeduction OtherInformationOfAnAccount = "Track Loan Deduction"
	OIOA_Restructured       OtherInformationOfAnAccount = "Restructured"
	OIOA_CashInBank         OtherInformationOfAnAccount = "Cash in Bank / Cash in Check Account"
	OIOA_CashOnHand         OtherInformationOfAnAccount = "Cash on Hand"
)

type InterestStandardComputation string

const (
	ISC_None     InterestStandardComputation = "None"
	ISC_Yearly   InterestStandardComputation = "Yearly"
	ISC_Mmonthly InterestStandardComputation = "Monthly"
)

type ComputationType string

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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account" json:"branch_id"`
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

		Name        string `gorm:"type:varchar(255);not null" json:"name"`
		Description string `gorm:"type:text;not null" json:"description"`

		MinAmount float64     `gorm:"type:decimal;default:0" json:"min_amount"`
		MaxAmount float64     `gorm:"type:decimal;default:50000" json:"max_amount"`
		Index     int         `gorm:"default:0" json:"index"`
		Type      AccountType `gorm:"type:varchar(50);not null" json:"type"`

		IsInternal         bool `gorm:"default:false" json:"is_internal"`
		CashOnHand         bool `gorm:"default:false" json:"cash_on_hand"`
		PaidUpShareCapital bool `gorm:"default:false" json:"paid_up_share_capital"`

		ComputationType ComputationType `gorm:"type:varchar(50)" json:"computation_type"`

		FinesAmort       float64 `gorm:"type:decimal" json:"fines_amort"`
		FinesMaturity    float64 `gorm:"type:decimal" json:"fines_maturity"`
		InterestStandard float64 `gorm:"type:decimal" json:"interest_standard"`
		InterestSecured  float64 `gorm:"type:decimal" json:"interest_secured"`

		ComputationSheetID *uuid.UUID        `gorm:"type:uuid" json:"computation_sheet_id"`
		ComputationSheet   *ComputationSheet `gorm:"foreignKey:ComputationSheetID;constraint:OnDelete:SET NULL;" json:"computation_sheet,omitempty"`

		CohCibFinesGracePeriodEntryCashHand                float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_cash_hand"`
		CohCibFinesGracePeriodEntryCashInBank              float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_cash_in_bank"`
		CohCibFinesGracePeriodEntryDailyAmortization       float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_daily_amortization"`
		CohCibFinesGracePeriodEntryDailyMaturity           float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_daily_maturity"`
		CohCibFinesGracePeriodEntryWeeklyAmortization      float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_weekly_amortization"`
		CohCibFinesGracePeriodEntryWeeklyMaturity          float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_weekly_maturity"`
		CohCibFinesGracePeriodEntryMonthlyAmortization     float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_monthly_amortization"`
		CohCibFinesGracePeriodEntryMonthlyMaturity         float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_monthly_maturity"`
		CohCibFinesGracePeriodEntrySemiMonthlyAmortization float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_semi_monthly_amortization"`
		CohCibFinesGracePeriodEntrySemiMonthlyMaturity     float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_semi_monthly_maturity"`
		CohCibFinesGracePeriodEntryQuarterlyAmortization   float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_quarterly_amortization"`
		CohCibFinesGracePeriodEntryQuarterlyMaturity       float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_quarterly_maturity"`
		CohCibFinesGracePeriodEntrySemiAnualAmortization   float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_semi_anual_amortization"`
		CohCibFinesGracePeriodEntrySemiAnualMaturity       float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_semi_anual_maturity"`
		CohCibFinesGracePeriodEntryLumpsumAmortization     float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_lumpsum_amortization"`
		CohCibFinesGracePeriodEntryLumpsumMaturity         float64 `gorm:"type:decimal;default:0" json:"coh_cib_fines_grace_period_entry_lumpsum_maturity"`

		FinancialStatementType string `gorm:"type:varchar(50)" json:"financial_statement_type"`
		GeneralLedgerType      string `gorm:"type:varchar(50)" json:"general_ledger_type"`

		AlternativeAccountID *uuid.UUID `gorm:"type:uuid" json:"alternative_account_id"`
		AlternativeAccount   *Account   `gorm:"foreignKey:AlternativeAccountID;constraint:OnDelete:SET NULL;" json:"alternative_account,omitempty"`

		FinesGracePeriodAmortization int  `gorm:"type:int" json:"fines_grace_period_amortization"`
		AdditionalGracePeriod        int  `gorm:"type:int" json:"additional_grace_period"`
		NumberGracePeriodDaily       bool `gorm:"default:false" json:"number_grace_period_daily"`
		FinesGracePeriodMaturity     int  `gorm:"type:int" json:"fines_grace_period_maturity"`
		YearlySubscriptionFee        int  `gorm:"type:int" json:"yearly_subscription_fee"`
		LoanCutOffDays               int  `gorm:"type:int" json:"loan_cut_off_days"`

		LumpsumComputationType                            string `gorm:"type:varchar(50);default:'None'" json:"lumpsum_computation_type"`
		InterestFinesComputationDiminishing               string `gorm:"type:varchar(100);default:'None'" json:"interest_fines_computation_diminishing"`
		InterestFinesComputationDiminishingStraightYearly string `gorm:"type:varchar(200);default:'None'" json:"interest_fines_computation_diminishing_straight_yearly"`
		EarnedUnearnedInterest                            string `gorm:"type:varchar(50);default:'None'" json:"earned_unearned_interest"`
		LoanSavingType                                    string `gorm:"type:varchar(50);default:'Separate'" json:"loan_saving_type"`
		InterestDeduction                                 string `gorm:"type:varchar(10);default:'Above'" json:"interest_deduction"`
		OtherDeductionEntry                               string `gorm:"type:varchar(20);default:'None'" json:"other_deduction_entry"`
		InterestSavingTypeDiminishingStraight             string `gorm:"type:varchar(20);default:'Spread'" json:"interest_saving_type_diminishing_straight"`
		OtherInformationOfAnAccount                       string `gorm:"type:varchar(50);default:'None'" json:"other_information_of_an_account"`

		HeaderRow int `gorm:"type:int" json:"header_row"`
		CenterRow int `gorm:"type:int" json:"center_row"`
		TotalRow  int `gorm:"type:int" json:"total_row"`

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

		InterestStandardComputation InterestStandardComputation `gorm:"type:varchar(20);default:'None'" json:"interest_standard_computation"`
	}
)

// --- RESPONSE & REQUEST STRUCTS ---

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
	CohCibFinesGracePeriodEntrySemiAnualAmortization   float64 `json:"coh_cib_fines_grace_period_entry_semi_anual_amortization"`
	CohCibFinesGracePeriodEntrySemiAnualMaturity       float64 `json:"coh_cib_fines_grace_period_entry_semi_anual_maturity"`
	CohCibFinesGracePeriodEntryLumpsumAmortization     float64 `json:"coh_cib_fines_grace_period_entry_lumpsum_amortization"`
	CohCibFinesGracePeriodEntryLumpsumMaturity         float64 `json:"coh_cib_fines_grace_period_entry_lumpsum_maturity"`

	FinancialStatementType FinancialStatementType `json:"financial_statement_type"`
	GeneralLedgerType      string                 `json:"general_ledger_type"`

	AlternativeAccountID *uuid.UUID       `gorm:"type:uuid" json:"alternative_account_id,omitempty"`
	AlternativeAccount   *AccountResponse `json:"alternative_account,omitempty"`

	FinesGracePeriodAmortization int  `json:"fines_grace_period_amortization"`
	AdditionalGracePeriod        int  `json:"additional_grace_period"`
	NumberGracePeriodDaily       bool `json:"number_grace_period_daily"`
	FinesGracePeriodMaturity     int  `json:"fines_grace_period_maturity"`
	YearlySubscriptionFee        int  `json:"yearly_subscription_fee"`
	LoanCutOffDays               int  `json:"loan_cut_off_days"`

	LumpsumComputationType                            LumpsumComputationType                            `json:"lumpsum_computation_type"`
	InterestFinesComputationDiminishing               InterestFinesComputationDiminishing               `json:"interest_fines_computation_diminishing"`
	InterestFinesComputationDiminishingStraightYearly InterestFinesComputationDiminishingStraightYearly `json:"interest_fines_computation_diminishing_straight_diminishing_yearly"`
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
	ShowInGeneralLedgerSourceCheckVoucher   bool   `json:"show_in_general_ledger_source_check_voucher"`

	CompassionFund              bool                        `json:"compassion_fund"`
	CompassionFundAmount        float64                     `json:"compassion_fund_amount"`
	CashAndCashEquivalence      bool                        `json:"cash_and_cash_equivalence"`
	InterestStandardComputation InterestStandardComputation `json:"interest_standard_computation"`
}

type AccountRequest struct {
	GeneralLedgerDefinitionID      *uuid.UUID `json:"general_ledger_definition_id,omitempty"`
	FinancialStatementDefinitionID *uuid.UUID `json:"financial_statement_definition_id,omitempty"`
	AccountClassificationID        *uuid.UUID `json:"account_classification_id,omitempty"`
	AccountCategoryID              *uuid.UUID `json:"account_category_id,omitempty"`
	MemberTypeID                   *uuid.UUID `json:"member_type_id,omitempty"`
	CurrencyID                     *uuid.UUID `json:"currency_id" validate:"required"`

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

	FinesAmort       float64 `json:"fines_amort,omitempty"`
	FinesMaturity    float64 `json:"fines_maturity,omitempty"`
	InterestStandard float64 `json:"interest_standard,omitempty"`
	InterestSecured  float64 `json:"interest_secured,omitempty"`

	ComputationSheetID *uuid.UUID `json:"computation_sheet_id,omitempty"`

	CohCibFinesGracePeriodEntryCashHand                float64 `json:"coh_cib_fines_grace_period_entry_cash_hand,omitempty"`
	CohCibFinesGracePeriodEntryCashInBank              float64 `json:"coh_cib_fines_grace_period_entry_cash_in_bank,omitempty"`
	CohCibFinesGracePeriodEntryDailyAmortization       float64 `json:"coh_cib_fines_grace_period_entry_daily_amortization,omitempty"`
	CohCibFinesGracePeriodEntryDailyMaturity           float64 `json:"coh_cib_fines_grace_period_entry_daily_maturity,omitempty"`
	CohCibFinesGracePeriodEntryWeeklyAmortization      float64 `json:"coh_cib_fines_grace_period_entry_weekly_amortization,omitempty"`
	CohCibFinesGracePeriodEntryWeeklyMaturity          float64 `json:"coh_cib_fines_grace_period_entry_weekly_maturity,omitempty"`
	CohCibFinesGracePeriodEntryMonthlyAmortization     float64 `json:"coh_cib_fines_grace_period_entry_monthly_amortization,omitempty"`
	CohCibFinesGracePeriodEntryMonthlyMaturity         float64 `json:"coh_cib_fines_grace_period_entry_monthly_maturity,omitempty"`
	CohCibFinesGracePeriodEntrySemiMonthlyAmortization float64 `json:"coh_cib_fines_grace_period_entry_semi_monthly_amortization,omitempty"`
	CohCibFinesGracePeriodEntrySemiMonthlyMaturity     float64 `json:"coh_cib_fines_grace_period_entry_semi_monthly_maturity,omitempty"`
	CohCibFinesGracePeriodEntryQuarterlyAmortization   float64 `json:"coh_cib_fines_grace_period_entry_quarterly_amortization,omitempty"`
	CohCibFinesGracePeriodEntryQuarterlyMaturity       float64 `json:"coh_cib_fines_grace_period_entry_quarterly_maturity,omitempty"`
	CohCibFinesGracePeriodEntrySemiAnualAmortization   float64 `json:"coh_cib_fines_grace_period_entry_semi_anual_amortization,omitempty"`
	CohCibFinesGracePeriodEntrySemiAnualMaturity       float64 `json:"coh_cib_fines_grace_period_entry_semi_anual_maturity,omitempty"`
	CohCibFinesGracePeriodEntryLumpsumAmortization     float64 `json:"coh_cib_fines_grace_period_entry_lumpsum_amortization,omitempty"`
	CohCibFinesGracePeriodEntryLumpsumMaturity         float64 `json:"coh_cib_fines_grace_period_entry_lumpsum_maturity,omitempty"`

	FinancialStatementType FinancialStatementType `json:"financial_statement_type,omitempty"`
	GeneralLedgerType      string                 `json:"general_ledger_type,omitempty"`

	AlternativeAccountID *uuid.UUID `json:"alternative_account_id,omitempty"`

	FinesGracePeriodAmortization int  `json:"fines_grace_period_amortization,omitempty"`
	AdditionalGracePeriod        int  `json:"additional_grace_period,omitempty"`
	NumberGracePeriodDaily       bool `json:"number_grace_period_daily,omitempty"`
	FinesGracePeriodMaturity     int  `json:"fines_grace_period_maturity,omitempty"`
	YearlySubscriptionFee        int  `json:"yearly_subscription_fee,omitempty"`
	LoanCutOffDays               int  `json:"loan_cut_off_days,omitempty"`

	LumpsumComputationType                            LumpsumComputationType                            `json:"lumpsum_computation_type,omitempty"`
	InterestFinesComputationDiminishing               InterestFinesComputationDiminishing               `json:"interest_fines_computation_diminishing,omitempty"`
	InterestFinesComputationDiminishingStraightYearly InterestFinesComputationDiminishingStraightYearly `json:"interest_fines_computation_diminishing_straight_diminishing_yearly,omitempty"`
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

func (m *ModelCore) Account() {
	m.Migration = append(m.Migration, &Account{})
	m.AccountManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		Account, AccountResponse, AccountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"AccountClassification", "AccountCategory",
			"AccountTags", "ComputationSheet", "Currency",
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
				CohCibFinesGracePeriodEntrySemiAnualAmortization:   data.CohCibFinesGracePeriodEntrySemiAnualAmortization,
				CohCibFinesGracePeriodEntrySemiAnualMaturity:       data.CohCibFinesGracePeriodEntrySemiAnualMaturity,
				CohCibFinesGracePeriodEntryLumpsumAmortization:     data.CohCibFinesGracePeriodEntryLumpsumAmortization,
				CohCibFinesGracePeriodEntryLumpsumMaturity:         data.CohCibFinesGracePeriodEntryLumpsumMaturity,
				FinancialStatementType:                             FinancialStatementType(data.FinancialStatementType),
				GeneralLedgerType:                                  data.GeneralLedgerType,
				AlternativeAccountID:                               data.AlternativeAccountID,
				AlternativeAccount:                                 m.AccountManager.ToModel(data.AlternativeAccount),
				FinesGracePeriodAmortization:                       data.FinesGracePeriodAmortization,
				AdditionalGracePeriod:                              data.AdditionalGracePeriod,
				NumberGracePeriodDaily:                             data.NumberGracePeriodDaily,
				FinesGracePeriodMaturity:                           data.FinesGracePeriodMaturity,
				YearlySubscriptionFee:                              data.YearlySubscriptionFee,
				LoanCutOffDays:                                     data.LoanCutOffDays,
				LumpsumComputationType:                             LumpsumComputationType(data.LumpsumComputationType),
				InterestFinesComputationDiminishing:                InterestFinesComputationDiminishing(data.InterestFinesComputationDiminishing),
				InterestFinesComputationDiminishingStraightYearly:  InterestFinesComputationDiminishingStraightYearly(data.InterestFinesComputationDiminishingStraightYearly),
				EarnedUnearnedInterest:                             EarnedUnearnedInterest(data.EarnedUnearnedInterest),
				LoanSavingType:                                     LoanSavingType(data.LoanSavingType),
				InterestDeduction:                                  InterestDeduction(data.InterestDeduction),
				OtherDeductionEntry:                                OtherDeductionEntry(data.OtherDeductionEntry),
				InterestSavingTypeDiminishingStraight:              InterestSavingTypeDiminishingStraight(data.InterestSavingTypeDiminishingStraight),
				OtherInformationOfAnAccount:                        OtherInformationOfAnAccount(data.OtherInformationOfAnAccount),
				HeaderRow:                                          data.HeaderRow,
				CenterRow:                                          data.CenterRow,
				TotalRow:                                           data.TotalRow,
				GeneralLedgerGroupingExcludeAccount:                data.GeneralLedgerGroupingExcludeAccount,
				AccountTags:                                        m.AccountTagManager.ToModels(data.AccountTags),
				ComputationSheet:                                   m.ComputationSheetManager.ToModel(data.ComputationSheet),

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

func (m *ModelCore) AccountSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
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
			CreatedAt:              now,
			CreatedByID:            userID,
			UpdatedAt:              now,
			UpdatedByID:            userID,
			OrganizationID:         organizationID,
			BranchID:               branchID,
			Name:                   "Regular Savings",
			Description:            "Basic savings account for general purpose savings with standard interest rates.",
			Type:                   AccountTypeDeposit,
			MinAmount:              100.00,
			MaxAmount:              1000000.00,
			InterestStandard:       2.5,
			CurrencyID:             &currency.ID,
			FinancialStatementType: string(FSTypeAssets),
			ComputationType:        "Simple Interest",
			Index:                  1,
		},
		{
			CreatedAt:              now,
			CreatedByID:            userID,
			UpdatedAt:              now,
			UpdatedByID:            userID,
			OrganizationID:         organizationID,
			BranchID:               branchID,
			Name:                   "Premium Savings",
			Description:            "High-yield savings account with better interest rates for higher balances.",
			Type:                   AccountTypeDeposit,
			MinAmount:              5000.00,
			MaxAmount:              5000000.00,
			InterestStandard:       4.0,
			FinancialStatementType: string(FSTypeAssets),
			ComputationType:        "Compound Interest",
			Index:                  2,
			CurrencyID:             &currency.ID,
		},
		{
			CreatedAt:              now,
			CreatedByID:            userID,
			UpdatedAt:              now,
			UpdatedByID:            userID,
			OrganizationID:         organizationID,
			BranchID:               branchID,
			Name:                   "Junior Savings",
			Description:            "Special savings account designed for minors and young members.",
			Type:                   AccountTypeDeposit,
			MinAmount:              50.00,
			MaxAmount:              100000.00,
			InterestStandard:       3.0,
			FinancialStatementType: string(FSTypeAssets),
			ComputationType:        "Simple Interest",
			Index:                  3,
			CurrencyID:             &currency.ID,
		},
		{
			CreatedAt:              now,
			CreatedByID:            userID,
			UpdatedAt:              now,
			UpdatedByID:            userID,
			OrganizationID:         organizationID,
			BranchID:               branchID,
			Name:                   "Senior Citizen Savings",
			Description:            "Special savings account with higher interest rates for senior citizens.",
			Type:                   AccountTypeDeposit,
			MinAmount:              500.00,
			MaxAmount:              2000000.00,
			InterestStandard:       3.5,
			FinancialStatementType: string(FSTypeAssets),
			ComputationType:        "Compound Interest",
			Index:                  4,
			CurrencyID:             &currency.ID,
		},
		{
			CreatedAt:              now,
			CreatedByID:            userID,
			UpdatedAt:              now,
			UpdatedByID:            userID,
			OrganizationID:         organizationID,
			BranchID:               branchID,
			Name:                   "Christmas Savings",
			Description:            "Seasonal savings account for holiday preparations with withdrawal restrictions.",
			Type:                   AccountTypeDeposit,
			MinAmount:              200.00,
			MaxAmount:              500000.00,
			InterestStandard:       3.0,
			FinancialStatementType: string(FSTypeAssets),
			ComputationType:        "Simple Interest",
			Index:                  5,
			CurrencyID:             &currency.ID,
		},
		{
			CreatedAt:              now,
			CreatedByID:            userID,
			UpdatedAt:              now,
			UpdatedByID:            userID,
			OrganizationID:         organizationID,
			BranchID:               branchID,
			Name:                   "Education Savings",
			Description:            "Long-term savings account dedicated to educational expenses.",
			Type:                   AccountTypeDeposit,
			MinAmount:              1000.00,
			MaxAmount:              3000000.00,
			InterestStandard:       4.0,
			FinancialStatementType: string(FSTypeAssets),
			ComputationType:        "Compound Interest",
			Index:                  6,
			CurrencyID:             &currency.ID,
		},
		{
			CreatedAt:              now,
			CreatedByID:            userID,
			UpdatedAt:              now,
			UpdatedByID:            userID,
			OrganizationID:         organizationID,
			BranchID:               branchID,
			Name:                   "Emergency Fund",
			Description:            "High-liquidity savings account for emergency situations.",
			Type:                   AccountTypeDeposit,
			MinAmount:              500.00,
			MaxAmount:              1000000.00,
			InterestStandard:       2.0,
			FinancialStatementType: string(FSTypeAssets),
			ComputationType:        "Simple Interest",
			Index:                  7,
			CurrencyID:             &currency.ID,
		},
		{
			CreatedAt:              now,
			CreatedByID:            userID,
			UpdatedAt:              now,
			UpdatedByID:            userID,
			OrganizationID:         organizationID,
			BranchID:               branchID,
			Name:                   "Business Savings",
			Description:            "Savings account designed for small businesses and entrepreneurs.",
			Type:                   AccountTypeDeposit,
			MinAmount:              2000.00,
			MaxAmount:              10000000.00,
			InterestStandard:       3.5,
			FinancialStatementType: string(FSTypeAssets),
			ComputationType:        "Compound Interest",
			Index:                  8,
			CurrencyID:             &currency.ID,
		},
		{
			CreatedAt:              now,
			CreatedByID:            userID,
			UpdatedAt:              now,
			UpdatedByID:            userID,
			OrganizationID:         organizationID,
			BranchID:               branchID,
			Name:                   "Retirement Savings",
			Description:            "Long-term savings account for retirement planning with tax benefits.",
			Type:                   AccountTypeDeposit,
			MinAmount:              1000.00,
			MaxAmount:              5000000.00,
			InterestStandard:       4.5,
			FinancialStatementType: string(FSTypeAssets),
			ComputationType:        "Compound Interest",
			Index:                  9,
			CurrencyID:             &currency.ID,
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
			FinancialStatementType:                  string(FSTypeAssets),
			ComputationType:                         "Diminishing Balance",
			Index:                                   10,
			LoanCutOffDays:                          3,
			FinesGracePeriodAmortization:            5,
			FinesGracePeriodMaturity:                7,
			AdditionalGracePeriod:                   2,
			LumpsumComputationType:                  string(LumpsumComputationNone),
			InterestFinesComputationDiminishing:     string(IFCDByAmortization),
			EarnedUnearnedInterest:                  string(EUITypeByFormula),
			LoanSavingType:                          string(LSTSeparate),
			InterestDeduction:                       string(InterestDeductionAbove),
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
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
			FinancialStatementType:                  string(FSTypeAssets),
			ComputationType:                         "Diminishing Balance",
			Index:                                   11,
			LoanCutOffDays:                          7,
			FinesGracePeriodAmortization:            10,
			FinesGracePeriodMaturity:                15,
			AdditionalGracePeriod:                   5,
			LumpsumComputationType:                  string(LumpsumComputationNone),
			InterestFinesComputationDiminishing:     string(IFCDByAmortization),
			EarnedUnearnedInterest:                  string(EUITypeByFormula),
			LoanSavingType:                          string(LSTSeparate),
			InterestDeduction:                       string(InterestDeductionAbove),
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			CurrencyID:                              &currency.ID,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
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
			FinancialStatementType:                  string(FSTypeAssets),
			ComputationType:                         "Simple Interest",
			Index:                                   12,
			LoanCutOffDays:                          14,
			FinesGracePeriodAmortization:            15,
			FinesGracePeriodMaturity:                30,
			AdditionalGracePeriod:                   10,
			LumpsumComputationType:                  string(LumpsumComputationNone),
			InterestFinesComputationDiminishing:     string(IFCDNone),
			EarnedUnearnedInterest:                  string(EUITypeByFormula),
			LoanSavingType:                          string(LSTSeparate),
			InterestDeduction:                       string(InterestDeductionBelow),
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
		},
	}

	// Create loan accounts and their alternative accounts
	for _, loanAccount := range loanAccounts {
		loanAccount.CurrencyID = &currency.ID
		// Create the main loan account
		if err := m.AccountManager.CreateWithTx(context, tx, loanAccount); err != nil {
			return eris.Wrapf(err, "failed to seed loan account %s", loanAccount.Name)
		}

		// Create Interest Account
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
			MaxAmount:                               1000000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   loanAccount.Index + 100, // Offset to avoid conflicts
			AlternativeAccountID:                    &loanAccount.ID,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
		}

		if err := m.AccountManager.CreateWithTx(context, tx, interestAccount); err != nil {
			return eris.Wrapf(err, "failed to seed interest account for %s", loanAccount.Name)
		}

		// Create Service Fee Account
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
			MaxAmount:                               50000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   loanAccount.Index + 200, // Offset to avoid conflicts
			AlternativeAccountID:                    &loanAccount.ID,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
		}

		if err := m.AccountManager.CreateWithTx(context, tx, serviceFeeAccount); err != nil {
			return eris.Wrapf(err, "failed to seed service fee account for %s", loanAccount.Name)
		}

		// Create Fines Account
		finesAccount := &Account{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			CurrencyID:                              &currency.ID,
			Name:                                    "Fines " + loanAccount.Name,
			Description:                             "Fines account for " + loanAccount.Description,
			Type:                                    AccountTypeFines,
			MinAmount:                               0.00,
			MaxAmount:                               100000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   loanAccount.Index + 300, // Offset to avoid conflicts
			AlternativeAccountID:                    &loanAccount.ID,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
		}

		if err := m.AccountManager.CreateWithTx(context, tx, finesAccount); err != nil {
			return eris.Wrapf(err, "failed to seed fines account for %s", loanAccount.Name)
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
		Description:                       "Member's share capital contribution representing ownership stake in the cooperative.",
		Type:                              AccountTypeOther,
		MinAmount:                         100.00,
		MaxAmount:                         1000000.00,
		InterestStandard:                  0.0,
		FinancialStatementType:            string(FSTypeEquity),
		ComputationType:                   "Fixed Amount",
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
	cashOnHand := &Account{
		CreatedAt:                               now,
		CreatedByID:                             userID,
		UpdatedAt:                               now,
		UpdatedByID:                             userID,
		OrganizationID:                          organizationID,
		BranchID:                                branchID,
		CurrencyID:                              &currency.ID,
		Name:                                    "Cash on Hand",
		Description:                             "Physical cash available at the branch for daily operations and transactions.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               10000000.00,
		InterestStandard:                        0.0,
		FinancialStatementType:                  string(FSTypeAssets),
		ComputationType:                         "None",
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

		OtherInformationOfAnAccount: string(OIOA_CashOnHand),
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
		Name:                                    "Cash in Bank",
		Description:                             "Funds deposited in bank accounts for secure storage and banking transactions.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               50000000.00,
		InterestStandard:                        0.0,
		FinancialStatementType:                  string(FSTypeAssets),
		ComputationType:                         "None",
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
		OtherInformationOfAnAccount:             string(OIOA_CashInBank),
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
		Name:                                    "Cash Online",
		Description:                             "Digital funds available through online banking platforms and digital wallets.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               10000000.00,
		InterestStandard:                        0.0,
		FinancialStatementType:                  string(FSTypeAssets),
		ComputationType:                         "None",
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
		OtherInformationOfAnAccount:             string(OIOA_None),
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
		Name:                                    "Petty Cash",
		Description:                             "Small amount of cash kept on hand for minor expenses and incidental purchases.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               100000.00,
		InterestStandard:                        0.0,
		FinancialStatementType:                  string(FSTypeAssets),
		ComputationType:                         "None",
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
		OtherInformationOfAnAccount:             string(OIOA_None),
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
		Name:                                    "Cash in Transit",
		Description:                             "Cash deposits or transfers that are in process but not yet cleared or posted.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               5000000.00,
		InterestStandard:                        0.0,
		FinancialStatementType:                  string(FSTypeAssets),
		ComputationType:                         "None",
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
		OtherInformationOfAnAccount:             string(OIOA_None),
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
		Name:                                    "Foreign Currency Cash",
		Description:                             "Cash holdings in foreign currencies for international transactions and exchange.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               2000000.00,
		InterestStandard:                        0.0,
		FinancialStatementType:                  string(FSTypeAssets),
		ComputationType:                         "None",
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
		OtherInformationOfAnAccount:             string(OIOA_None),
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
		Name:                                    "Money Market Fund",
		Description:                             "Short-term, highly liquid investments that can be quickly converted to cash.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               20000000.00,
		InterestStandard:                        1.5,
		FinancialStatementType:                  string(FSTypeAssets),
		ComputationType:                         "Simple Interest",
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
		OtherInformationOfAnAccount:             string(OIOA_None),
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
		Name:                                    "Treasury Bills",
		Description:                             "Short-term government securities with maturity of less than one year.",
		Type:                                    AccountTypeOther,
		MinAmount:                               0.00,
		MaxAmount:                               15000000.00,
		InterestStandard:                        2.0,
		FinancialStatementType:                  string(FSTypeAssets),
		ComputationType:                         "Simple Interest",
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
		OtherInformationOfAnAccount:             string(OIOA_None),
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
			MinAmount:                               0.00,
			MaxAmount:                               10000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   19,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			MinAmount:                               0.00,
			MaxAmount:                               1000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   20,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			MinAmount:                               0.00,
			MaxAmount:                               50000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   21,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			CurrencyID:                              &currency.ID,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			MinAmount:                               0.00,
			MaxAmount:                               500.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   22,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			MinAmount:                               0.00,
			MaxAmount:                               200.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   23,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Fees for check processing, clearance, and checkbook issuance services.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               1000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   24,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Fee for preparing legal documents, certificates, and official statements.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               2000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   25,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Penalty fees charged for late loan payments and overdue account obligations.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               5000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   26,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Fee charged for closing accounts and terminating membership services.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               1000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   27,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Yearly membership fee for maintaining cooperative membership status.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               5000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   28,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Insurance premium fees for loan protection and member insurance coverage.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               20000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   29,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Fee for notarial services and document authentication requirements.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               3000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeRevenue),
			ComputationType:                         "Fixed Amount",
			Index:                                   30,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Expenses for computer hardware maintenance, software updates, and IT support services.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               100000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   31,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "General maintenance expenses for equipment, furniture, and operational assets.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               150000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   32,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Monthly electricity and power consumption expenses for branch operations.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               50000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   33,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Monthly water utility expenses for branch facilities and operations.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               20000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   34,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Costs for building repairs, renovations, and structural maintenance work.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               500000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   35,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Monthly internet, phone, and communication service expenses.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               30000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   36,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Expenses for office supplies, stationery, and consumable materials.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               25000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   37,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Expenses for security guards, surveillance systems, and safety equipment.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               80000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   38,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Expenses for janitorial services, cleaning supplies, and facility sanitation.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               40000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   39,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Fees for legal, accounting, consulting, and other professional services.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               200000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   40,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Expenses for company vehicle maintenance, fuel, and transportation costs.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               60000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   41,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Rental expenses for equipment, machinery, and temporary facility needs.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               100000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   42,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Expenses for employee training, seminars, workshops, and professional development.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               75000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   43,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Expenses for promotional activities, advertising campaigns, and marketing materials.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               100000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   44,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Business travel expenses including transportation, lodging, and meal allowances.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               80000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   45,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Expenses for business permits, licenses, regulatory fees, and government compliance.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               50000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   46,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Expenses for employee health benefits, medical services, and workplace health programs.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               150000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   47,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Expenses for garbage collection, waste disposal, and environmental compliance services.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               15000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   48,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
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
			Description:                             "Unexpected expenses for emergency repairs, urgent purchases, and crisis management.",
			Type:                                    AccountTypeOther,
			MinAmount:                               0.00,
			MaxAmount:                               200000.00,
			InterestStandard:                        0.0,
			FinancialStatementType:                  string(FSTypeExpenses),
			ComputationType:                         "Fixed Amount",
			Index:                                   49,
			ShowInGeneralLedgerSourceWithdraw:       true,
			ShowInGeneralLedgerSourceDeposit:        true,
			ShowInGeneralLedgerSourceJournal:        true,
			ShowInGeneralLedgerSourcePayment:        true,
			ShowInGeneralLedgerSourceAdjustment:     true,
			ShowInGeneralLedgerSourceJournalVoucher: true,
			ShowInGeneralLedgerSourceCheckVoucher:   true,
			OtherInformationOfAnAccount:             string(OIOA_None),
			CurrencyID:                              &currency.ID,
		},
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
	if err := m.BranchSettingManager.UpdateFieldsWithTx(context, tx, branch.BranchSetting.ID, branch.BranchSetting); err != nil {
		return eris.Wrapf(err, "failed to update branch %s with paid up share capital account", branch.Name)
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

	// Set default currency to PHP
	userOrganization.SettingsCurrencyDefaultValueID = &currency.ID

	if err := m.UserOrganizationManager.UpdateFieldsWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
		return eris.Wrap(err, "failed to update user organization with default accounting accounts and currency")
	}

	return nil
}

func (m *ModelCore) AccountCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*Account, error) {
	return m.AccountManager.Find(context, &Account{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}

func (m *ModelCore) AccountLockForUpdate(ctx context.Context, tx *gorm.DB, accountID uuid.UUID) (*Account, error) {
	var lockedAccount Account
	err := tx.WithContext(ctx).
		Model(&Account{}).
		Where("id = ?", accountID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&lockedAccount).Error

	if err != nil {
		return nil, err
	}

	return &lockedAccount, nil
}

// account = lockedAccount
func (m *ModelCore) AccountLockWithValidation(ctx context.Context, tx *gorm.DB, accountID uuid.UUID, originalAccount *Account) (*Account, error) {
	lockedAccount, err := m.AccountLockForUpdate(ctx, tx, accountID)
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
