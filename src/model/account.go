package model

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

		Name        string `gorm:"type:varchar(255);not null" json:"name"`
		Description string `gorm:"type:text;not null" json:"description"`

		MinAmount float64     `gorm:"type:decimal;default:0" json:"min_amount"`
		MaxAmount float64     `gorm:"type:decimal;default:50000" json:"max_amount"`
		Index     int         `gorm:"default:0" json:"index"`
		Type      AccountType `gorm:"type:varchar(50);not null" json:"type"`

		IsInternal         bool `gorm:"default:false" json:"is_internal"`
		CashOnHand         bool `gorm:"default:false" json:"cash_on_hand"`
		PaidUpShareCapital bool `gorm:"default:false" json:"paid_up_share_capital"`

		ComputationType string `gorm:"type:varchar(50)" json:"computation_type"`

		FinesAmort       float64 `gorm:"type:decimal" json:"fines_amort"`
		FinesMaturity    float64 `gorm:"type:decimal" json:"fines_maturity"`
		InterestStandard float64 `gorm:"type:decimal" json:"interest_standard"`
		InterestSecured  float64 `gorm:"type:decimal" json:"interest_secured"`

		ComputationSheetID *uuid.UUID `gorm:"type:uuid" json:"computation_sheet_id"`

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

		AlternativeCode string `gorm:"type:varchar(50)" json:"alternative_code"`

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

		CompassionFund       bool    `gorm:"default:false" json:"compassion_fund"`
		CompassionFundAmount float64 `gorm:"type:decimal;default:0" json:"compassion_fund_amount"`
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

	Name        string      `json:"name"`
	Description string      `json:"description"`
	MinAmount   float64     `json:"min_amount"`
	MaxAmount   float64     `json:"max_amount"`
	Index       int         `json:"index"`
	Type        AccountType `json:"type"`

	IsInternal         bool `json:"is_internal"`
	CashOnHand         bool `json:"cash_on_hand"`
	PaidUpShareCapital bool `json:"paid_up_share_capital"`

	ComputationType string `json:"computation_type"`

	FinesAmort       float64 `json:"fines_amort"`
	FinesMaturity    float64 `json:"fines_maturity"`
	InterestStandard float64 `json:"interest_standard"`
	InterestSecured  float64 `json:"interest_secured"`

	ComputationSheetID *uuid.UUID `json:"computation_sheet_id,omitempty"`

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

	AlternativeCode string `json:"alternative_code"`

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

	CompassionFund       bool    `json:"compassion_fund"`
	CompassionFundAmount float64 `json:"compassion_fund_amount"`
}

type AccountRequest struct {
	GeneralLedgerDefinitionID      *uuid.UUID `json:"general_ledger_definition_id,omitempty"`
	FinancialStatementDefinitionID *uuid.UUID `json:"financial_statement_definition_id,omitempty"`
	AccountClassificationID        *uuid.UUID `json:"account_classification_id,omitempty"`
	AccountCategoryID              *uuid.UUID `json:"account_category_id,omitempty"`
	MemberTypeID                   *uuid.UUID `json:"member_type_id,omitempty"`

	Name        string      `json:"name" validate:"required,min=1,max=255"`
	Description string      `json:"description"`
	MinAmount   float64     `json:"min_amount,omitempty"`
	MaxAmount   float64     `json:"max_amount,omitempty"`
	Index       int         `json:"index,omitempty"`
	Type        AccountType `json:"type" validate:"required"`

	IsInternal         bool `json:"is_internal,omitempty"`
	CashOnHand         bool `json:"cash_on_hand,omitempty"`
	PaidUpShareCapital bool `json:"paid_up_share_capital,omitempty"`

	ComputationType string `json:"computation_type,omitempty"`

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

	AlternativeCode string `json:"alternative_code,omitempty"`

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

	CompassionFund       bool    `json:"compassion_fund,omitempty"`
	CompassionFundAmount float64 `json:"compassion_fund_amount,omitempty"`
}

// --- REGISTRATION ---

func (m *Model) Account() {
	m.Migration = append(m.Migration, &Account{})
	m.AccountManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		Account, AccountResponse, AccountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"AccountClassification", "AccountCategory",
			"AccountTags",
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
				AlternativeCode:                                    data.AlternativeCode,
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

				Icon:                                    data.Icon,
				ShowInGeneralLedgerSourceWithdraw:       data.ShowInGeneralLedgerSourceWithdraw,
				ShowInGeneralLedgerSourceDeposit:        data.ShowInGeneralLedgerSourceDeposit,
				ShowInGeneralLedgerSourceJournal:        data.ShowInGeneralLedgerSourceJournal,
				ShowInGeneralLedgerSourcePayment:        data.ShowInGeneralLedgerSourcePayment,
				ShowInGeneralLedgerSourceAdjustment:     data.ShowInGeneralLedgerSourceAdjustment,
				ShowInGeneralLedgerSourceJournalVoucher: data.ShowInGeneralLedgerSourceJournalVoucher,
				ShowInGeneralLedgerSourceCheckVoucher:   data.ShowInGeneralLedgerSourceCheckVoucher,

				CompassionFund:       data.CompassionFund,
				CompassionFundAmount: data.CompassionFundAmount,
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

func (m *Model) AccountSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	accounts := []*Account{
		// Regular Savings Accounts
		{
			CreatedAt:        now,
			CreatedByID:      userID,
			UpdatedAt:        now,
			UpdatedByID:      userID,
			OrganizationID:   organizationID,
			BranchID:         branchID,
			Name:             "Regular Savings",
			Description:      "Basic savings account for general purpose savings with standard interest rates.",
			Type:             AccountTypeDeposit,
			MinAmount:        100.00,
			MaxAmount:        1000000.00,
			InterestStandard: 2.5,

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
		},

		// Loan Accounts
		{
			CreatedAt:                           now,
			CreatedByID:                         userID,
			UpdatedAt:                           now,
			UpdatedByID:                         userID,
			OrganizationID:                      organizationID,
			BranchID:                            branchID,
			Name:                                "Emergency Loan",
			Description:                         "Quick access loan for urgent financial needs and unexpected expenses.",
			Type:                                AccountTypeLoan,
			MinAmount:                           1000.00,
			MaxAmount:                           100000.00,
			InterestStandard:                    8.5,
			InterestSecured:                     7.5,
			FinesAmort:                          1.0,
			FinesMaturity:                       2.0,
			FinancialStatementType:              string(FSTypeAssets),
			ComputationType:                     "Diminishing Balance",
			Index:                               10,
			LoanCutOffDays:                      3,
			FinesGracePeriodAmortization:        5,
			FinesGracePeriodMaturity:            7,
			AdditionalGracePeriod:               2,
			LumpsumComputationType:              string(LumpsumComputationNone),
			InterestFinesComputationDiminishing: string(IFCDByAmortization),
			EarnedUnearnedInterest:              string(EUITypeByFormula),
			LoanSavingType:                      string(LSTSeparate),
			InterestDeduction:                   string(InterestDeductionAbove),
		},
		{
			CreatedAt:                           now,
			CreatedByID:                         userID,
			UpdatedAt:                           now,
			UpdatedByID:                         userID,
			OrganizationID:                      organizationID,
			BranchID:                            branchID,
			Name:                                "Business Loan",
			Description:                         "Capital loan for business expansion, equipment purchase, and working capital needs.",
			Type:                                AccountTypeLoan,
			MinAmount:                           50000.00,
			MaxAmount:                           5000000.00,
			InterestStandard:                    10.0,
			InterestSecured:                     9.0,
			FinesAmort:                          1.5,
			FinesMaturity:                       2.5,
			FinancialStatementType:              string(FSTypeAssets),
			ComputationType:                     "Diminishing Balance",
			Index:                               11,
			LoanCutOffDays:                      7,
			FinesGracePeriodAmortization:        10,
			FinesGracePeriodMaturity:            15,
			AdditionalGracePeriod:               5,
			LumpsumComputationType:              string(LumpsumComputationNone),
			InterestFinesComputationDiminishing: string(IFCDByAmortization),
			EarnedUnearnedInterest:              string(EUITypeByFormula),
			LoanSavingType:                      string(LSTSeparate),
			InterestDeduction:                   string(InterestDeductionAbove),
		},
		{
			CreatedAt:                           now,
			CreatedByID:                         userID,
			UpdatedAt:                           now,
			UpdatedByID:                         userID,
			OrganizationID:                      organizationID,
			BranchID:                            branchID,
			Name:                                "Educational Loan",
			Description:                         "Student loan for tuition fees, educational expenses, and academic development.",
			Type:                                AccountTypeLoan,
			MinAmount:                           5000.00,
			MaxAmount:                           500000.00,
			InterestStandard:                    6.5,
			InterestSecured:                     5.5,
			FinesAmort:                          0.5,
			FinesMaturity:                       1.0,
			FinancialStatementType:              string(FSTypeAssets),
			ComputationType:                     "Simple Interest",
			Index:                               12,
			LoanCutOffDays:                      14,
			FinesGracePeriodAmortization:        15,
			FinesGracePeriodMaturity:            30,
			AdditionalGracePeriod:               10,
			LumpsumComputationType:              string(LumpsumComputationNone),
			InterestFinesComputationDiminishing: string(IFCDNone),
			EarnedUnearnedInterest:              string(EUITypeByFormula),
			LoanSavingType:                      string(LSTSeparate),
			InterestDeduction:                   string(InterestDeductionBelow),
		},
	}
	for _, data := range accounts {
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
	paidUpShareCapital := &Account{
		CreatedAt:                         now,
		CreatedByID:                       userID,
		UpdatedAt:                         now,
		UpdatedByID:                       userID,
		OrganizationID:                    organizationID,
		BranchID:                          branchID,
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
		OtherInformationOfAnAccount:             string(OIOA_CashOnHand),
	}

	if err := m.AccountManager.CreateWithTx(context, tx, cashOnHand); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", cashOnHand.Name)
	}
	branch, err := m.BranchManager.GetByID(context, branchID)
	if err != nil {
		return eris.Wrapf(err, "failed to find branch with ID %s", branchID)
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

	if err := m.UserOrganizationManager.UpdateFieldsWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
		return eris.Wrap(err, "failed to update user organization with default accounting accounts")
	}

	return nil
}

func (m *Model) AccountCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*Account, error) {
	return m.AccountManager.Find(context, &Account{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}

func (m *Model) AccountLockForUpdate(ctx context.Context, tx *gorm.DB, accountID uuid.UUID) (*Account, error) {
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
func (m *Model) AccountLockWithValidation(ctx context.Context, tx *gorm.DB, accountID uuid.UUID, originalAccount *Account) (*Account, error) {
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
