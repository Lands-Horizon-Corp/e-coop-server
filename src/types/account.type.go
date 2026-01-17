package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	GLTypeAssets      GeneralLedgerType = "Assets"
	GLTypeLiabilities GeneralLedgerType = "Liabilities"
	GLTypeEquity      GeneralLedgerType = "Equity"
	GLTypeRevenue     GeneralLedgerType = "Revenue"
	GLTypeExpenses    GeneralLedgerType = "Expenses"

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

	LumpsumComputationNone             LumpsumComputationType = "None"
	LumpsumComputationFinesMaturity    LumpsumComputationType = "Compute Fines Maturity"
	LumpsumComputationInterestMaturity LumpsumComputationType = "Compute Interest Maturity / Terms"
	LumpsumComputationAdvanceInterest  LumpsumComputationType = "Compute Advance Interest"

	IFCDNone                  InterestFinesComputationDiminishing = "None"
	IFCDByAmortization        InterestFinesComputationDiminishing = "By Amortization"
	IFCDByAmortizationDalyArr InterestFinesComputationDiminishing = "By Amortization Daly on Interest Principal + Interest = Fines(Arr)"

	IFCDSYNone                   InterestFinesComputationDiminishingStraightYearly = "None"
	IFCDSYByDailyInterestBalance InterestFinesComputationDiminishingStraightYearly = "By Daily on Interest based on loan balance by year Principal + Interest Amortization = Fines Fines Grace Period Month end Amortization"

	EUITypeNone                    EarnedUnearnedInterest = "None"
	EUITypeByFormula               EarnedUnearnedInterest = "By Formula"
	EUITypeByFormulaActualPay      EarnedUnearnedInterest = "By Formula + Actual Pay"
	EUITypeByAdvanceInterestActual EarnedUnearnedInterest = "By Advance Interest + Actual Pay"

	LSTSeparate                 LoanSavingType = "Separate"
	LSTSingleLedger             LoanSavingType = "Single Ledger"
	LSTSingleLedgerIfNotZero    LoanSavingType = "Single Ledger if Not Zero"
	LSTSingleLedgerSemi1530     LoanSavingType = "Single Ledger Semi (15/30)"
	LSTSingleLedgerSemiMaturity LoanSavingType = "Single Ledger Semi Within Maturity"

	InterestDeductionAbove InterestDeduction = "Above"
	InterestDeductionBelow InterestDeduction = "Below"

	OtherDeductionEntryNone       OtherDeductionEntry = "None"
	OtherDeductionEntryHealthCare OtherDeductionEntry = "Health Care"

	ISTDSSpread     InterestSavingTypeDiminishingStraight = "Spread"
	ISTDS1stPayment InterestSavingTypeDiminishingStraight = "1st Payment"

	OIOANone               OtherInformationOfAnAccount = "None"
	OIOAJewely             OtherInformationOfAnAccount = "Jewely"
	OIOAGrocery            OtherInformationOfAnAccount = "Grocery"
	OIOATrackLoanDeduction OtherInformationOfAnAccount = "Track Loan Deduction"
	OIOARestructured       OtherInformationOfAnAccount = "Restructured"
	OIOACashInBank         OtherInformationOfAnAccount = "Cash in Bank / Cash in Check Account"
	OIOACashOnHand         OtherInformationOfAnAccount = "Cash on Hand"

	ISCNone    InterestStandardComputation = "None"
	ISCYearly  InterestStandardComputation = "Yearly"
	ISCMonthly InterestStandardComputation = "Monthly"

	Straight            ComputationType = "Straight"
	Diminishing         ComputationType = "Diminishing"
	DiminishingStraight ComputationType = "DiminishingStraight"
)

type (
	GeneralLedgerType                                 string
	AccountType                                       string
	LumpsumComputationType                            string
	InterestFinesComputationDiminishing               string
	InterestFinesComputationDiminishingStraightYearly string
	EarnedUnearnedInterest                            string
	LoanSavingType                                    string
	InterestDeduction                                 string
	OtherDeductionEntry                               string
	InterestSavingTypeDiminishingStraight             string
	OtherInformationOfAnAccount                       string
	InterestStandardComputation                       string
	ComputationType                                   string

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
