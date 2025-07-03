package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

// --- ENUMS ---

type AccountType string

const (
	AccountTypeDeposit   AccountType = "Deposit"
	AccountTypeLoan      AccountType = "Loan"
	AccountTypeARLedger  AccountType = "A/R-Ledger"
	AccountTypeARAging   AccountType = "A/R-Aging"
	AccountTypeFines     AccountType = "Fines"
	AccountTypeInterest  AccountType = "Interest"
	AccountTypeSVFLedger AccountType = "SVF-Ledger"
	AccountTypeWOff      AccountType = "W-Off"
	AccountTypeAPLedger  AccountType = "A/P-Ledger"
	AccountTypeOther     AccountType = "Other"
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

		GeneralLedgerGroupingExcludeAccount bool `gorm:"default:false" json:"general_ledger_grouping_exclude_account"`
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

	GeneralLedgerGroupingExcludeAccount bool `json:"general_ledger_grouping_exclude_account"`
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

	GeneralLedgerGroupingExcludeAccount bool `json:"general_ledger_grouping_exclude_account,omitempty"`
}

// --- REGISTRATION ---

func (m *Model) Account() {
	m.Migration = append(m.Migration, &Account{})
	m.AccountManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		Account, AccountResponse, AccountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
			"AccountClassification", "AccountCategory",
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

func (m *Model) AccountCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*Account, error) {
	return m.AccountManager.Find(context, &Account{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
