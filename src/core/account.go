package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type GeneralLedgerType string

const (
	GLTypeAssets      GeneralLedgerType = "Assets"
	GLTypeLiabilities GeneralLedgerType = "Liabilities"
	GLTypeEquity      GeneralLedgerType = "Equity"
	GLTypeRevenue     GeneralLedgerType = "Revenue"
	GLTypeExpenses    GeneralLedgerType = "Expenses"
)

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
	InterestDeductionAbove InterestDeduction = "Above"
	InterestDeductionBelow InterestDeduction = "Below"
)

type OtherDeductionEntry string

const (
	OtherDeductionEntryNone       OtherDeductionEntry = "None"
	OtherDeductionEntryHealthCare OtherDeductionEntry = "Health Care"
)

type InterestSavingTypeDiminishingStraight string

const (
	ISTDSSpread     InterestSavingTypeDiminishingStraight = "Spread"
	ISTDS1stPayment InterestSavingTypeDiminishingStraight = "1st Payment"
)

type OtherInformationOfAnAccount string

const (
	OIOANone               OtherInformationOfAnAccount = "None"
	OIOAJewely             OtherInformationOfAnAccount = "Jewely"
	OIOAGrocery            OtherInformationOfAnAccount = "Grocery"
	OIOATrackLoanDeduction OtherInformationOfAnAccount = "Track Loan Deduction"
	OIOARestructured       OtherInformationOfAnAccount = "Restructured"
	OIOACashInBank         OtherInformationOfAnAccount = "Cash in Bank / Cash in Check Account"
	OIOACashOnHand         OtherInformationOfAnAccount = "Cash on Hand"
)

type InterestStandardComputation string

const (
	ISCNone    InterestStandardComputation = "None"
	ISCYearly  InterestStandardComputation = "Yearly"
	ISCMonthly InterestStandardComputation = "Monthly"
)

type ComputationType string

const (
	Straight            ComputationType = "Straight"
	Diminishing         ComputationType = "Diminishing"
	DiminishingStraight ComputationType = "DiminishingStraight"
)

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

		DefaultPaymentTypeID *uuid.UUID   `gorm:"type:uuid" json:"default_payment_type_id"`
		DefaultPaymentType   *PaymentType `gorm:"foreignKey:DefaultPaymentTypeID;constraint:OnDelete:SET NULL;" json:"default_payment_type,omitempty"`

		Name        string `gorm:"type:varchar(255);not null;uniqueIndex:idx_account_name_org_branch" json:"name"`
		Description string `gorm:"type:text;not null" json:"description"`

		MinAmount float64     `gorm:"type:decimal;default:0" json:"min_amount"`
		MaxAmount float64     `gorm:"type:decimal;default:50000" json:"max_amount"`
		Index     float64     `gorm:"default:0" json:"index"`
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
		AccountHistoryID            *uuid.UUID                  `json:"account_history_id"`
		InterestAmortization        float64                     `gorm:"type:decimal;default:0" json:"interest_amortization,omitempty"`
		InterestMaturity            float64                     `gorm:"type:decimal;default:0" json:"interest_maturity,omitempty"`

		IsTaxable bool `gorm:"default:true" json:"is_taxable"`
	}
)

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
	Index       float64     `json:"index"`
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
	ShowInGeneralLedgerSourceCheckVoucher   bool   `json:"show_in_general_ledger_source_check_voucher"`

	CompassionFund              bool                        `json:"compassion_fund"`
	CompassionFundAmount        float64                     `json:"compassion_fund_amount"`
	CashAndCashEquivalence      bool                        `json:"cash_and_cash_equivalence"`
	InterestStandardComputation InterestStandardComputation `json:"interest_standard_computation"`
	AccountHistoryID            *uuid.UUID                  `json:"account_history_id"`

	InterestAmortization float64 `json:"interest_amortization,omitempty"`
	InterestMaturity     float64 `json:"interest_maturity,omitempty"`
	IsTaxable            bool    `json:"is_taxable"`
}

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
	Index       float64     `json:"index,omitempty"`
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
	InterestAmortization        float64                     `json:"interest_amortization,omitempty"`
	InterestMaturity            float64                     `json:"interest_maturity,omitempty"`
	IsTaxable                   bool                        `json:"is_taxable,omitempty"`
}

func AccountManager(service *horizon.HorizonService) *registry.Registry[Account, AccountResponse, AccountRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		Account, AccountResponse, AccountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"AccountClassification", "AccountCategory",
			"AccountTags", "ComputationSheet", "Currency",
			"DefaultPaymentType", "LoanAccount",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *Account) *AccountResponse {
			if data == nil {
				return nil
			}
			return &AccountResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),

				GeneralLedgerDefinitionID:      data.GeneralLedgerDefinitionID,
				GeneralLedgerDefinition:        GeneralLedgerDefinitionManager(service).ToModel(data.GeneralLedgerDefinition),
				FinancialStatementDefinitionID: data.FinancialStatementDefinitionID,
				FinancialStatementDefinition:   FinancialStatementDefinitionManager(service).ToModel(data.FinancialStatementDefinition),
				AccountClassificationID:        data.AccountClassificationID,
				AccountClassification:          AccountClassificationManager(service).ToModel(data.AccountClassification),
				AccountCategoryID:              data.AccountCategoryID,
				AccountCategory:                AccountCategoryManager(service).ToModel(data.AccountCategory),
				MemberTypeID:                   data.MemberTypeID,
				MemberType:                     MemberTypeManager(service).ToModel(data.MemberType),
				CurrencyID:                     data.CurrencyID,
				Currency:                       CurrencyManager(service).ToModel(data.Currency),
				DefaultPaymentTypeID:           data.DefaultPaymentTypeID,
				DefaultPaymentType:             PaymentTypeManager(service).ToModel(data.DefaultPaymentType),

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
				LoanAccount:                         AccountManager(service).ToModel(data.LoanAccount),
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
				AccountTags:                                       AccountTagManager(service).ToModels(data.AccountTags),
				ComputationSheet:                                  ComputationSheetManager(service).ToModel(data.ComputationSheet),

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
				AccountHistoryID:            data.AccountHistoryID,
				InterestAmortization:        data.InterestAmortization,
				InterestMaturity:            data.InterestMaturity,
				IsTaxable:                   data.IsTaxable,
			}
		},
		Created: func(data *Account) registry.Topics {
			return []string{
				"account.create",
				fmt.Sprintf("account.create.%s", data.ID),
				fmt.Sprintf("account.create.branch.%s", data.BranchID),
				fmt.Sprintf("account.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Account) registry.Topics {
			return []string{
				"account.update",
				fmt.Sprintf("account.update.%s", data.ID),
				fmt.Sprintf("account.update.branch.%s", data.BranchID),
				fmt.Sprintf("account.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Account) registry.Topics {
			return []string{
				"account.delete",
				fmt.Sprintf("account.delete.%s", data.ID),
				fmt.Sprintf("account.delete.branch.%s", data.BranchID),
				fmt.Sprintf("account.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func accountSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()

	branch, err := BranchManager(service).GetByID(context, branchID)
	if err != nil {
		return eris.Wrap(err, "failed to find branch for account seeding")
	}

	accounts := []*Account{
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
			CurrencyID:        branch.CurrencyID,
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
			Index:             2,
			CurrencyID:        branch.CurrencyID,
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
			Index:             3,
			CurrencyID:        branch.CurrencyID,
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
			Index:             4,
			CurrencyID:        branch.CurrencyID,
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
			CurrencyID:        branch.CurrencyID,
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
			ComputationType:   Straight,
			Index:             6,
			CurrencyID:        branch.CurrencyID,
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
			CurrencyID:        branch.CurrencyID,
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
			CurrencyID:        branch.CurrencyID,
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
			Index:             9,
			CurrencyID:        branch.CurrencyID,
			Icon:              "Clock",
		},
	}

	for _, data := range accounts {
		data.CurrencyID = branch.CurrencyID
		data.ShowInGeneralLedgerSourceWithdraw = true
		data.ShowInGeneralLedgerSourceDeposit = true
		data.ShowInGeneralLedgerSourceJournal = true
		data.ShowInGeneralLedgerSourcePayment = true
		data.ShowInGeneralLedgerSourceAdjustment = true
		data.ShowInGeneralLedgerSourceJournalVoucher = true
		data.ShowInGeneralLedgerSourceCheckVoucher = true
		if err := AccountManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed account %s", data.Name)
		}

		if err := CreateAccountHistory(context, service, tx, data); err != nil {
			return eris.Wrapf(err, "history: failed to create history for seeded account %s (ID: %s, tx: %v)", data.Name, data.ID, tx != nil)
		}
	}

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
			CurrencyID:                              branch.CurrencyID,
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
			CurrencyID:                              branch.CurrencyID,
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
			CurrencyID:                              branch.CurrencyID,
			Icon:                                    "Book Open",
		},
	}

	for _, loanAccount := range loanAccounts {
		loanAccount.CurrencyID = branch.CurrencyID
		if err := AccountManager(service).CreateWithTx(context, tx, loanAccount); err != nil {
			return eris.Wrapf(err, "failed to seed loan account %s", loanAccount.Name)
		}

		if err := CreateAccountHistory(context, service, tx, loanAccount); err != nil {
			return eris.Wrapf(err, "history: failed to create history for seeded loan account %s", loanAccount.Name)
		}

		var interestComputationType ComputationType
		var interestStandardRate float64

		switch loanAccount.Name {
		case "Emergency Loan":
			interestComputationType = Diminishing
			interestStandardRate = 2.5 // 2.5% interest standard
		case "Business Loan":
			interestComputationType = DiminishingStraight
			interestStandardRate = 3.0 // 3% interest standard
		case "Educational Loan":
			interestComputationType = Diminishing
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
			CurrencyID:                              branch.CurrencyID,
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

		if err := AccountManager(service).CreateWithTx(context, tx, interestAccount); err != nil {
			return eris.Wrapf(err, "failed to seed interest account for %s", loanAccount.Name)
		}

		if err := CreateAccountHistory(context, service, tx, interestAccount); err != nil {
			return eris.Wrapf(err, "history: failed to create history for seeded interest account for %s", loanAccount.Name)
		}

		var svfComputationType ComputationType
		var svfStandardRate float64

		switch loanAccount.Name {
		case "Emergency Loan":
			svfComputationType = Straight
			svfStandardRate = 1.0
		case "Business Loan":
			svfComputationType = DiminishingStraight
			svfStandardRate = 1.5
		case "Educational Loan":
			svfComputationType = Diminishing
			svfStandardRate = 0.5
		default:
			svfComputationType = DiminishingStraight
			svfStandardRate = 1.0
		}

		serviceFeeAccount := &Account{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			CurrencyID:                              branch.CurrencyID,
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

		if err := AccountManager(service).CreateWithTx(context, tx, serviceFeeAccount); err != nil {
			return eris.Wrapf(err, "failed to seed service fee account for %s", loanAccount.Name)
		}

		if err := CreateAccountHistory(context, service, tx, serviceFeeAccount); err != nil {
			return eris.Wrapf(err, "history: failed to create history for seeded service fee account for %s", loanAccount.Name)
		}

		finesAccount := &Account{
			CreatedAt:        now,
			CreatedByID:      userID,
			UpdatedAt:        now,
			UpdatedByID:      userID,
			OrganizationID:   organizationID,
			BranchID:         branchID,
			CurrencyID:       branch.CurrencyID,
			Name:             "Fines " + loanAccount.Name,
			Description:      "Fines account for " + loanAccount.Description,
			Type:             AccountTypeFines,
			MinAmount:        0.00,
			MaxAmount:        100.00, // Max percentage is 100%
			InterestStandard: 0.0,

			FinesAmort:    2.5, // 2.5% fine on amortization
			FinesMaturity: 5.0, // 5.0% fine on maturity

			FinesGracePeriodAmortization: 7,     // 7 days grace period for amortization fines
			FinesGracePeriodMaturity:     15,    // 15 days grace period for maturity fines
			AdditionalGracePeriod:        3,     // 3 additional days
			NoGracePeriodDaily:           false, // Allow daily grace period

			GeneralLedgerType: GLTypeRevenue,
			ComputationType:   Straight,
			Index:             loanAccount.Index + 300, // Offset to avoid conflicts
			LoanAccountID:     &loanAccount.ID,

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

		if err := AccountManager(service).CreateWithTx(context, tx, finesAccount); err != nil {
			return eris.Wrapf(err, "failed to seed fines account for %s", loanAccount.Name)
		}

		if err := CreateAccountHistory(context, service, tx, finesAccount); err != nil {
			return eris.Wrapf(err, "history: failed to create history for seeded fines account for %s", loanAccount.Name)
		}
	}

	standaloneFinesAccounts := []*Account{
		{
			CreatedAt:                    now,
			CreatedByID:                  userID,
			UpdatedAt:                    now,
			UpdatedByID:                  userID,
			OrganizationID:               organizationID,
			BranchID:                     branchID,
			CurrencyID:                   branch.CurrencyID,
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
			CurrencyID:                   branch.CurrencyID,
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
			CurrencyID:                   branch.CurrencyID,
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

	for _, finesAccount := range standaloneFinesAccounts {
		if err := AccountManager(service).CreateWithTx(context, tx, finesAccount); err != nil {
			return eris.Wrapf(err, "failed to seed standalone fines account %s", finesAccount.Name)
		}

		if err := CreateAccountHistory(context, service, tx, finesAccount); err != nil {
			return eris.Wrapf(err, "history: failed to create history for seeded standalone fines account %s", finesAccount.Name)
		}
	}

	standaloneInterestAccounts := []*Account{
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			CurrencyID:                              branch.CurrencyID,
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
			CurrencyID:                              branch.CurrencyID,
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
			CurrencyID:                              branch.CurrencyID,
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

	standaloneSVFAccounts := []*Account{
		{
			CreatedAt:                               now,
			CreatedByID:                             userID,
			UpdatedAt:                               now,
			UpdatedByID:                             userID,
			OrganizationID:                          organizationID,
			BranchID:                                branchID,
			CurrencyID:                              branch.CurrencyID,
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
			CurrencyID:                              branch.CurrencyID,
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
			CurrencyID:                              branch.CurrencyID,
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

	for _, interestAccount := range standaloneInterestAccounts {
		if err := AccountManager(service).CreateWithTx(context, tx, interestAccount); err != nil {
			return eris.Wrapf(err, "failed to seed standalone interest account %s", interestAccount.Name)
		}

		if err := CreateAccountHistory(context, service, tx, interestAccount); err != nil {
			return eris.Wrapf(err, "history: failed to create history for seeded standalone interest account %s", interestAccount.Name)
		}
	}

	for _, svfAccount := range standaloneSVFAccounts {
		if err := AccountManager(service).CreateWithTx(context, tx, svfAccount); err != nil {
			return eris.Wrapf(err, "failed to seed standalone SVF account %s", svfAccount.Name)
		}

		if err := CreateAccountHistory(context, service, tx, svfAccount); err != nil {
			return eris.Wrapf(err, "history: failed to create history for seeded standalone SVF account %s", svfAccount.Name)
		}
	}

	paidUpShareCapital := &Account{
		CreatedAt:                         now,
		CreatedByID:                       userID,
		UpdatedAt:                         now,
		UpdatedByID:                       userID,
		OrganizationID:                    organizationID,
		BranchID:                          branchID,
		CurrencyID:                        branch.CurrencyID,
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
	if err := AccountManager(service).CreateWithTx(context, tx, paidUpShareCapital); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", paidUpShareCapital.Name)
	}

	if err := CreateAccountHistory(context, service, tx, paidUpShareCapital); err != nil {
		return eris.Wrapf(err, "history: failed to create history for seeded account %s", paidUpShareCapital.Name)
	}

	var cashOnHandPaymentType *PaymentType

	cashOnHandPaymentType, _ = PaymentTypeManager(service).FindOne(context, &PaymentType{
		OrganizationID: organizationID,
		BranchID:       branchID,
		Name:           "Cash On Hand",
	})

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

		if err := PaymentTypeManager(service).CreateWithTx(context, tx, cashOnHandPaymentType); err != nil {
			return eris.Wrapf(err, "failed to seed payment type %s", cashOnHandPaymentType.Name)
		}
		paymentTypes := []*PaymentType{
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
			if err := PaymentTypeManager(service).CreateWithTx(context, tx, data); err != nil {
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
		CurrencyID:                              branch.CurrencyID,
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

	if err := AccountManager(service).CreateWithTx(context, tx, cashOnHand); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", cashOnHand.Name)
	}

	if err := CreateAccountHistory(context, service, tx, cashOnHand); err != nil {
		return eris.Wrapf(err, "history: failed to create history for seeded account %s", cashOnHand.Name)
	}

	cashInBank := &Account{
		CreatedAt:                               now,
		CreatedByID:                             userID,
		UpdatedAt:                               now,
		UpdatedByID:                             userID,
		OrganizationID:                          organizationID,
		BranchID:                                branchID,
		CurrencyID:                              branch.CurrencyID,
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

	if err := AccountManager(service).CreateWithTx(context, tx, cashInBank); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", cashInBank.Name)
	}

	if err := CreateAccountHistory(context, service, tx, cashInBank); err != nil {
		return eris.Wrapf(err, "history: failed to create history for seeded account %s", cashInBank.Name)
	}

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
		CurrencyID:                              branch.CurrencyID,
	}

	if err := AccountManager(service).CreateWithTx(context, tx, cashOnline); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", cashOnline.Name)
	}

	if err := CreateAccountHistory(context, service, tx, cashOnline); err != nil {
		return eris.Wrapf(err, "history: failed to create history for seeded account %s", cashOnline.Name)
	}

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
		CurrencyID:                              branch.CurrencyID,
	}

	if err := AccountManager(service).CreateWithTx(context, tx, pettyCash); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", pettyCash.Name)
	}

	if err := CreateAccountHistory(context, service, tx, pettyCash); err != nil {
		return eris.Wrapf(err, "history: failed to create history for seeded account %s", pettyCash.Name)
	}

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
		CurrencyID:                              branch.CurrencyID,
	}

	if err := AccountManager(service).CreateWithTx(context, tx, cashInTransit); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", cashInTransit.Name)
	}

	if err := CreateAccountHistory(context, service, tx, cashInTransit); err != nil {
		return eris.Wrapf(err, "history: failed to create history for seeded account %s", cashInTransit.Name)
	}

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
		CurrencyID:                              branch.CurrencyID,
		OtherInformationOfAnAccount:             OIOANone,
	}

	if err := AccountManager(service).CreateWithTx(context, tx, foreignCurrencyCash); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", foreignCurrencyCash.Name)
	}

	if err := CreateAccountHistory(context, service, tx, foreignCurrencyCash); err != nil {
		return eris.Wrapf(err, "history: failed to create history for seeded account %s", foreignCurrencyCash.Name)
	}

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
		CurrencyID:                              branch.CurrencyID,
		OtherInformationOfAnAccount:             OIOANone,
	}

	if err := AccountManager(service).CreateWithTx(context, tx, moneyMarketFund); err != nil {
		return eris.Wrapf(err, "failed to seed account %s", moneyMarketFund.Name)
	}

	if err := CreateAccountHistory(context, service, tx, moneyMarketFund); err != nil {
		return eris.Wrapf(err, "history: failed to create history for seeded account %s", moneyMarketFund.Name)
	}

	feeAccounts := []*Account{
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
	}

	operationalAccounts := []*Account{
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
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
			CurrencyID:                              branch.CurrencyID,
		},
	}

	cooperativeAccounts := []*Account{
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
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
			CurrencyID:                              branch.CurrencyID,
			OtherInformationOfAnAccount:             OIOANone,
		},
	}

	for _, coopAccount := range cooperativeAccounts {
		coopAccount.CurrencyID = branch.CurrencyID
		if err := AccountManager(service).CreateWithTx(context, tx, coopAccount); err != nil {
			return eris.Wrapf(err, "failed to seed cooperative account %s", coopAccount.Name)
		}
		if err := CreateAccountHistory(context, service, tx, coopAccount); err != nil {
			return eris.Wrapf(err, "history: failed to seed cooperative account %s", coopAccount.Name)
		}

	}

	for _, feeAccount := range feeAccounts {
		feeAccount.CurrencyID = branch.CurrencyID
		if err := AccountManager(service).CreateWithTx(context, tx, feeAccount); err != nil {
			return eris.Wrapf(err, "failed to seed fee account %s", feeAccount.Name)
		}
		if err := CreateAccountHistory(context, service, tx, feeAccount); err != nil {
			return eris.Wrapf(err, "history: failed to seed fee account %s", feeAccount.Name)
		}
	}

	for _, operationalAccount := range operationalAccounts {
		operationalAccount.CurrencyID = branch.CurrencyID
		if err := AccountManager(service).CreateWithTx(context, tx, operationalAccount); err != nil {
			return eris.Wrapf(err, "failed to seed operational account %s", operationalAccount.Name)
		}
		if err := CreateAccountHistory(context, service, tx, operationalAccount); err != nil {
			return eris.Wrapf(err, "history: failed to seed operational account %s", operationalAccount.Name)
		}
	}
	compassionFund := &Account{
		CreatedAt:         now,
		CreatedByID:       userID,
		UpdatedAt:         now,
		UpdatedByID:       userID,
		OrganizationID:    organizationID,
		BranchID:          branchID,
		Name:              "Compassion Fund",
		Description:       "Special deposit account for emergency assistance and member welfare support.",
		Type:              AccountTypeDeposit,
		MinAmount:         100.00,
		MaxAmount:         1000000.00,
		InterestStandard:  2.5,
		GeneralLedgerType: GLTypeLiabilities,
		ComputationType:   Straight,
		Index:             10,
		CurrencyID:        branch.CurrencyID,
		Icon:              "Heart",
	}
	if err := AccountManager(service).CreateWithTx(context, tx, compassionFund); err != nil {
		return eris.Wrap(err, "failed to create compassion fund account")
	}
	if err := CreateAccountHistory(context, service, tx, compassionFund); err != nil {
		return eris.Wrap(err, "history: failed to create compassion fund account")
	}

	branch.BranchSetting.CompassionFundAccountID = &compassionFund.ID
	branch.BranchSetting.PaidUpSharedCapitalAccountID = &paidUpShareCapital.ID
	branch.BranchSetting.CashOnHandAccountID = &cashOnHand.ID
	if err := BranchSettingManager(service).UpdateByIDWithTx(context, tx, branch.BranchSetting.ID, branch.BranchSetting); err != nil {
		return eris.Wrap(err, "failed to update branch settings with paid up share capital and cash on hand accounts")
	}

	unbalanced := &UnbalancedAccount{
		CreatedAt:            now,
		CreatedByID:          userID,
		UpdatedAt:            now,
		UpdatedByID:          userID,
		BranchSettingsID:     branch.BranchSetting.ID,
		CurrencyID:           *branch.CurrencyID,
		AccountForShortageID: cashOnHand.ID,
		AccountForOverageID:  cashOnHand.ID,
	}
	if err := UnbalancedAccountManager(service).CreateWithTx(context, tx, unbalanced); err != nil {
		return eris.Wrap(err, "failed to create unbalanced account for branch")
	}

	var regularSavings *Account
	for _, account := range accounts {
		if account.Name == "Regular Savings" {
			regularSavings = account
			break
		}
	}

	userOrganization, err := UserOrganizationManager(service).FindOne(context, &UserOrganization{
		UserID:         userID,
		OrganizationID: organizationID,
		BranchID:       &branchID,
	})
	if regularSavings != nil {
		userOrganization.SettingsAccountingPaymentDefaultValueID = &regularSavings.ID
		userOrganization.SettingsAccountingDepositDefaultValueID = &regularSavings.ID
		userOrganization.SettingsAccountingWithdrawDefaultValueID = &regularSavings.ID
	}
	if err != nil {
		return eris.Wrap(err, "failed to find user organization for setting default payment type")
	}
	userOrganization.SettingsPaymentTypeDefaultValueID = &cashOnHandPaymentType.ID
	if err := UserOrganizationManager(service).UpdateByIDWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
		return eris.Wrap(err, "failed to update user organization with default payment type")
	}
	return nil
}

func CreateAccountHistory(ctx context.Context, service *horizon.HorizonService, tx *gorm.DB, account *Account) error {
	now := time.Now().UTC()
	history := &AccountHistory{
		AccountID:      account.ID,
		OrganizationID: account.OrganizationID,
		BranchID:       account.BranchID,
		CreatedByID:    account.CreatedByID,
		CreatedAt:      now,
		LoanAccountID:  account.LoanAccountID,

		Name:                                account.Name,
		Description:                         account.Description,
		Type:                                account.Type,
		MinAmount:                           account.MinAmount,
		MaxAmount:                           account.MaxAmount,
		Index:                               account.Index,
		IsInternal:                          account.IsInternal,
		CashOnHand:                          account.CashOnHand,
		PaidUpShareCapital:                  account.PaidUpShareCapital,
		ComputationType:                     account.ComputationType,
		FinesAmort:                          account.FinesAmort,
		FinesMaturity:                       account.FinesMaturity,
		InterestStandard:                    account.InterestStandard,
		InterestSecured:                     account.InterestSecured,
		FinesGracePeriodAmortization:        account.FinesGracePeriodAmortization,
		AdditionalGracePeriod:               account.AdditionalGracePeriod,
		NoGracePeriodDaily:                  account.NoGracePeriodDaily,
		FinesGracePeriodMaturity:            account.FinesGracePeriodMaturity,
		YearlySubscriptionFee:               account.YearlySubscriptionFee,
		CutOffDays:                          account.CutOffDays,
		CutOffMonths:                        account.CutOffMonths,
		LumpsumComputationType:              account.LumpsumComputationType,
		InterestFinesComputationDiminishing: account.InterestFinesComputationDiminishing,
		InterestFinesComputationDiminishingStraightYearly: account.InterestFinesComputationDiminishingStraightYearly,
		EarnedUnearnedInterest:                            account.EarnedUnearnedInterest,
		LoanSavingType:                                    account.LoanSavingType,
		InterestDeduction:                                 account.InterestDeduction,
		OtherDeductionEntry:                               account.OtherDeductionEntry,
		InterestSavingTypeDiminishingStraight:             account.InterestSavingTypeDiminishingStraight,
		OtherInformationOfAnAccount:                       account.OtherInformationOfAnAccount,
		GeneralLedgerType:                                 account.GeneralLedgerType,
		HeaderRow:                                         account.HeaderRow,
		CenterRow:                                         account.CenterRow,
		TotalRow:                                          account.TotalRow,
		GeneralLedgerGroupingExcludeAccount:               account.GeneralLedgerGroupingExcludeAccount,
		Icon:                                              account.Icon,
		ShowInGeneralLedgerSourceWithdraw:                 account.ShowInGeneralLedgerSourceWithdraw,
		ShowInGeneralLedgerSourceDeposit:                  account.ShowInGeneralLedgerSourceDeposit,
		ShowInGeneralLedgerSourceJournal:                  account.ShowInGeneralLedgerSourceJournal,
		ShowInGeneralLedgerSourcePayment:                  account.ShowInGeneralLedgerSourcePayment,
		ShowInGeneralLedgerSourceAdjustment:               account.ShowInGeneralLedgerSourceAdjustment,
		ShowInGeneralLedgerSourceJournalVoucher:           account.ShowInGeneralLedgerSourceJournalVoucher,
		ShowInGeneralLedgerSourceCheckVoucher:             account.ShowInGeneralLedgerSourceCheckVoucher,
		CompassionFund:                                    account.CompassionFund,
		CompassionFundAmount:                              account.CompassionFundAmount,
		CashAndCashEquivalence:                            account.CashAndCashEquivalence,
		InterestStandardComputation:                       account.InterestStandardComputation,

		GeneralLedgerDefinitionID:      account.GeneralLedgerDefinitionID,
		FinancialStatementDefinitionID: account.FinancialStatementDefinitionID,
		AccountClassificationID:        account.AccountClassificationID,
		AccountCategoryID:              account.AccountCategoryID,
		MemberTypeID:                   account.MemberTypeID,
		CurrencyID:                     account.CurrencyID,
		DefaultPaymentTypeID:           account.DefaultPaymentTypeID,
		ComputationSheetID:             account.ComputationSheetID,

		CohCibFinesGracePeriodEntryCashHand:                account.CohCibFinesGracePeriodEntryCashHand,
		CohCibFinesGracePeriodEntryCashInBank:              account.CohCibFinesGracePeriodEntryCashInBank,
		CohCibFinesGracePeriodEntryDailyAmortization:       account.CohCibFinesGracePeriodEntryDailyAmortization,
		CohCibFinesGracePeriodEntryDailyMaturity:           account.CohCibFinesGracePeriodEntryDailyMaturity,
		CohCibFinesGracePeriodEntryWeeklyAmortization:      account.CohCibFinesGracePeriodEntryWeeklyAmortization,
		CohCibFinesGracePeriodEntryWeeklyMaturity:          account.CohCibFinesGracePeriodEntryWeeklyMaturity,
		CohCibFinesGracePeriodEntryMonthlyAmortization:     account.CohCibFinesGracePeriodEntryMonthlyAmortization,
		CohCibFinesGracePeriodEntryMonthlyMaturity:         account.CohCibFinesGracePeriodEntryMonthlyMaturity,
		CohCibFinesGracePeriodEntrySemiMonthlyAmortization: account.CohCibFinesGracePeriodEntrySemiMonthlyAmortization,
		CohCibFinesGracePeriodEntrySemiMonthlyMaturity:     account.CohCibFinesGracePeriodEntrySemiMonthlyMaturity,
		CohCibFinesGracePeriodEntryQuarterlyAmortization:   account.CohCibFinesGracePeriodEntryQuarterlyAmortization,
		CohCibFinesGracePeriodEntryQuarterlyMaturity:       account.CohCibFinesGracePeriodEntryQuarterlyMaturity,
		CohCibFinesGracePeriodEntrySemiAnnualAmortization:  account.CohCibFinesGracePeriodEntrySemiAnnualAmortization,
		CohCibFinesGracePeriodEntrySemiAnnualMaturity:      account.CohCibFinesGracePeriodEntrySemiAnnualMaturity,
		CohCibFinesGracePeriodEntryAnnualAmortization:      account.CohCibFinesGracePeriodEntryAnnualAmortization,
		CohCibFinesGracePeriodEntryAnnualMaturity:          account.CohCibFinesGracePeriodEntryAnnualMaturity,
		CohCibFinesGracePeriodEntryLumpsumAmortization:     account.CohCibFinesGracePeriodEntryLumpsumAmortization,
		CohCibFinesGracePeriodEntryLumpsumMaturity:         account.CohCibFinesGracePeriodEntryLumpsumMaturity,
	}

	if tx == nil {
		return eris.New("transaction is nil in CreateAccountHistory - cannot create history without transaction context")
	}

	return AccountHistoryManager(service).CreateWithTx(ctx, tx, history)
}

func CreateAccountHistoryBeforeUpdate(ctx context.Context, service *horizon.HorizonService, tx *gorm.DB, accountID uuid.UUID, updatedBy uuid.UUID) error {
	original, err := AccountManager(service).GetByID(ctx, accountID)
	if err != nil {
		return nil
	}

	now := time.Now().UTC()
	history := &AccountHistory{
		AccountID:      accountID,
		OrganizationID: original.OrganizationID,
		BranchID:       original.BranchID,
		CreatedByID:    updatedBy,
		CreatedAt:      now,
		LoanAccountID:  original.LoanAccountID,

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

		GeneralLedgerDefinitionID:      original.GeneralLedgerDefinitionID,
		FinancialStatementDefinitionID: original.FinancialStatementDefinitionID,
		AccountClassificationID:        original.AccountClassificationID,
		AccountCategoryID:              original.AccountCategoryID,
		MemberTypeID:                   original.MemberTypeID,
		CurrencyID:                     original.CurrencyID,
		DefaultPaymentTypeID:           original.DefaultPaymentTypeID,
		ComputationSheetID:             original.ComputationSheetID,

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
		CohCibFinesGracePeriodEntryAnnualAmortization:      original.CohCibFinesGracePeriodEntryAnnualAmortization,
		CohCibFinesGracePeriodEntryAnnualMaturity:          original.CohCibFinesGracePeriodEntryAnnualMaturity,
		CohCibFinesGracePeriodEntryLumpsumAmortization:     original.CohCibFinesGracePeriodEntryLumpsumAmortization,
		CohCibFinesGracePeriodEntryLumpsumMaturity:         original.CohCibFinesGracePeriodEntryLumpsumMaturity,
	}

	if tx != nil {
		return AccountHistoryManager(service).CreateWithTx(ctx, tx, history)
	}
	return AccountHistoryManager(service).Create(ctx, history)
}

func AccountCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*Account, error) {
	return AccountManager(service).Find(context, &Account{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func AccountLockForUpdate(ctx context.Context, service *horizon.HorizonService, tx *gorm.DB, accountID uuid.UUID) (*Account, error) {
	return AccountManager(service).GetByIDLock(ctx, tx, accountID)
}

func AccountLockWithValidation(ctx context.Context, service *horizon.HorizonService, tx *gorm.DB, accountID uuid.UUID, originalAccount *Account) (*Account, error) {
	lockedAccount, err := AccountManager(service).GetByIDLock(ctx, tx, accountID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to acquire account lock")
	}

	if originalAccount != nil {
		if lockedAccount.OrganizationID != originalAccount.OrganizationID ||
			lockedAccount.BranchID != originalAccount.BranchID ||
			lockedAccount.Type != originalAccount.Type {
			return nil, eris.New("account was modified by another transaction")
		}
	}

	return lockedAccount, nil
}

func LoanAccounts(ctx context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*Account, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return AccountManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func FindAccountsByTypesAndBranch(ctx context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID, currencyID uuid.UUID) ([]*Account, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "currency_id", Op: query.ModeEqual, Value: currencyID},
		{Field: "type", Op: query.ModeInside, Value: []AccountType{
			AccountTypeFines,
			AccountTypeInterest,
			AccountTypeSVFLedger,
		}},
	}
	return AccountManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func FindAccountsBySpecificType(ctx context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID, accountType AccountType) ([]*Account, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "type", Op: query.ModeEqual, Value: accountType},
	}

	return AccountManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func FindLoanAccountsByID(ctx context.Context,
	service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID, accountID uuid.UUID) ([]*Account, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "loan_account_id", Op: query.ModeEqual, Value: accountID},
	}

	accounts, err := AccountManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func AccountDeleteCheck(ctx context.Context, service *horizon.HorizonService, accountID uuid.UUID) error {
	hasEntries, err := GeneralLedgerManager(service).ArrExists(ctx, []query.ArrFilterSQL{
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
	})
	if err != nil {
		return eris.Wrap(err, "failed to check general ledger entries for account")
	}

	if hasEntries {
		return eris.New("cannot delete account: account has existing general ledger entries")
	}

	account, err := AccountManager(service).GetByID(ctx, accountID)
	if err != nil {
		return eris.Wrap(err, "failed to retrieve account for validation")
	}

	branchSetting, err := BranchSettingManager(service).FindOne(ctx, &BranchSetting{
		BranchID: account.BranchID,
	})
	if err != nil && !eris.Is(err, gorm.ErrRecordNotFound) {
		return eris.Wrap(err, "failed to retrieve branch settings")
	}

	if branchSetting != nil && branchSetting.CashOnHandAccountID != nil &&
		*branchSetting.CashOnHandAccountID == accountID {
		return eris.New("cannot delete account: it is currently set as the Cash on Hand account in branch settings")
	}

	if branchSetting != nil && branchSetting.PaidUpSharedCapitalAccountID != nil &&
		*branchSetting.PaidUpSharedCapitalAccountID == accountID {
		return eris.New("cannot delete account: it is currently set as the Paid Up Share Capital account in branch settings")
	}

	UnbalancedAccount, err := UnbalancedAccountManager(service).FindOne(ctx, &UnbalancedAccount{
		BranchSettingsID: branchSetting.ID,
	})
	if err != nil && !eris.Is(err, gorm.ErrRecordNotFound) {
		return eris.Wrap(err, "failed to check unbalanced account references")
	}

	if UnbalancedAccount != nil {
		if UnbalancedAccount.AccountForShortageID == accountID {
			return eris.New("cannot delete account: it is currently set as the shortage account in branch settings")
		}
		if UnbalancedAccount.AccountForOverageID == accountID {
			return eris.New("cannot delete account: it is currently set as the overage account in branch settings")
		}
	}

	linkedAccounts, err := FindLoanAccountsByID(ctx, service, account.OrganizationID, account.BranchID, accountID)
	if err != nil && !eris.Is(err, gorm.ErrRecordNotFound) {
		return eris.Wrap(err, "failed to check linked loan accounts")
	}

	if len(linkedAccounts) > 0 {
		return eris.Errorf("cannot delete account: %d other accounts (Interest/Fines/SVF) are linked to this loan account. Please delete or unlink them first", len(linkedAccounts))
	}

	return nil
}

func AccountDeleteCheckIncludingDeleted(ctx context.Context, service *horizon.HorizonService, accountID uuid.UUID) error {
	hasEntries, err := GeneralLedgerManager(service).ExistsIncludingDeleted(ctx, []query.ArrFilterSQL{
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
	})
	if err != nil {
		return eris.Wrap(err, "failed to check general ledger entries for account")
	}

	if hasEntries {
		return eris.New("cannot delete account: account has existing general ledger entries (including deleted)")
	}

	return nil
}
