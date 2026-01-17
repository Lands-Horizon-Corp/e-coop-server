package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HistoryChangeType string

const (
	HistoryChangeTypeCreated HistoryChangeType = "created"

	HistoryChangeTypeUpdated HistoryChangeType = "updated"

	HistoryChangeTypeDeleted HistoryChangeType = "deleted"
)

type (
	AccountHistory struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`

		AccountID uuid.UUID `gorm:"type:uuid;not null;index:idx_account_history_account" json:"account_id"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"account,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_account_history_org_branch" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_account_history_org_branch" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string      `gorm:"type:varchar(255)" json:"name"`
		Description string      `gorm:"type:text" json:"description"`
		Type        AccountType `gorm:"type:varchar(50)" json:"type"`
		MinAmount   float64     `gorm:"type:decimal" json:"min_amount"`
		MaxAmount   float64     `gorm:"type:decimal" json:"max_amount"`
		Index       float64     `gorm:"default:0" json:"index"`

		IsInternal         bool `gorm:"default:false" json:"is_internal"`
		CashOnHand         bool `gorm:"default:false" json:"cash_on_hand"`
		PaidUpShareCapital bool `gorm:"default:false" json:"paid_up_share_capital"`

		ComputationType ComputationType `gorm:"type:varchar(50)" json:"computation_type"`

		FinesAmort       float64 `gorm:"type:decimal" json:"fines_amort"`
		FinesMaturity    float64 `gorm:"type:decimal" json:"fines_maturity"`
		InterestStandard float64 `gorm:"type:decimal" json:"interest_standard"`
		InterestSecured  float64 `gorm:"type:decimal" json:"interest_secured"`

		FinesGracePeriodAmortization int  `gorm:"type:int" json:"fines_grace_period_amortization"`
		AdditionalGracePeriod        int  `gorm:"type:int" json:"additional_grace_period"`
		NoGracePeriodDaily           bool `gorm:"default:false" json:"no_grace_period_daily"`
		FinesGracePeriodMaturity     int  `gorm:"type:int" json:"fines_grace_period_maturity"`
		YearlySubscriptionFee        int  `gorm:"type:int" json:"yearly_subscription_fee"`
		CutOffDays                   int  `gorm:"type:int;default:0" json:"cut_off_days"`
		CutOffMonths                 int  `gorm:"type:int;default:0" json:"cut_off_months"`

		LumpsumComputationType                            LumpsumComputationType                            `gorm:"type:varchar(50)" json:"lumpsum_computation_type"`
		InterestFinesComputationDiminishing               InterestFinesComputationDiminishing               `gorm:"type:varchar(100)" json:"interest_fines_computation_diminishing"`
		InterestFinesComputationDiminishingStraightYearly InterestFinesComputationDiminishingStraightYearly `gorm:"type:varchar(200)" json:"interest_fines_computation_diminishing_straight_yearly"`
		EarnedUnearnedInterest                            EarnedUnearnedInterest                            `gorm:"type:varchar(50)" json:"earned_unearned_interest"`
		LoanSavingType                                    LoanSavingType                                    `gorm:"type:varchar(50)" json:"loan_saving_type"`
		InterestDeduction                                 InterestDeduction                                 `gorm:"type:varchar(10)" json:"interest_deduction"`
		OtherDeductionEntry                               OtherDeductionEntry                               `gorm:"type:varchar(20)" json:"other_deduction_entry"`
		InterestSavingTypeDiminishingStraight             InterestSavingTypeDiminishingStraight             `gorm:"type:varchar(20)" json:"interest_saving_type_diminishing_straight"`
		OtherInformationOfAnAccount                       OtherInformationOfAnAccount                       `gorm:"type:varchar(50)" json:"other_information_of_an_account"`

		GeneralLedgerType GeneralLedgerType `gorm:"type:varchar(50)" json:"general_ledger_type"`

		HeaderRow int `gorm:"type:int" json:"header_row"`
		CenterRow int `gorm:"type:int" json:"center_row"`
		TotalRow  int `gorm:"type:int" json:"total_row"`

		GeneralLedgerGroupingExcludeAccount bool   `gorm:"default:false" json:"general_ledger_grouping_exclude_account"`
		Icon                                string `gorm:"type:varchar(50)" json:"icon"`

		ShowInGeneralLedgerSourceWithdraw       bool `gorm:"default:true" json:"show_in_general_ledger_source_withdraw"`
		ShowInGeneralLedgerSourceDeposit        bool `gorm:"default:true" json:"show_in_general_ledger_source_deposit"`
		ShowInGeneralLedgerSourceJournal        bool `gorm:"default:true" json:"show_in_general_ledger_source_journal"`
		ShowInGeneralLedgerSourcePayment        bool `gorm:"default:true" json:"show_in_general_ledger_source_payment"`
		ShowInGeneralLedgerSourceAdjustment     bool `gorm:"default:true" json:"show_in_general_ledger_source_adjustment"`
		ShowInGeneralLedgerSourceJournalVoucher bool `gorm:"default:true" json:"show_in_general_ledger_source_journal_voucher"`
		ShowInGeneralLedgerSourceCheckVoucher   bool `gorm:"default:true" json:"show_in_general_ledger_source_check_voucher"`

		CompassionFund         bool    `gorm:"default:false" json:"compassion_fund"`
		CompassionFundAmount   float64 `gorm:"type:decimal" json:"compassion_fund_amount"`
		CashAndCashEquivalence bool    `gorm:"default:false" json:"cash_and_cash_equivalence"`

		InterestStandardComputation InterestStandardComputation `gorm:"type:varchar(20)" json:"interest_standard_computation"`

		GeneralLedgerDefinitionID      *uuid.UUID `gorm:"type:uuid" json:"general_ledger_definition_id,omitempty"`
		FinancialStatementDefinitionID *uuid.UUID `gorm:"type:uuid" json:"financial_statement_definition_id,omitempty"`
		AccountClassificationID        *uuid.UUID `gorm:"type:uuid" json:"account_classification_id,omitempty"`
		AccountCategoryID              *uuid.UUID `gorm:"type:uuid" json:"account_category_id,omitempty"`
		MemberTypeID                   *uuid.UUID `gorm:"type:uuid" json:"member_type_id,omitempty"`
		CurrencyID                     *uuid.UUID `gorm:"type:uuid" json:"currency_id,omitempty"`
		DefaultPaymentTypeID           *uuid.UUID `gorm:"type:uuid" json:"default_payment_type_id,omitempty"`
		ComputationSheetID             *uuid.UUID `gorm:"type:uuid" json:"computation_sheet_id,omitempty"`
		LoanAccountID                  *uuid.UUID `gorm:"type:uuid" json:"loan_account_id,omitempty"`

		CohCibFinesGracePeriodEntryCashHand                float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_cash_hand"`
		CohCibFinesGracePeriodEntryCashInBank              float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_cash_in_bank"`
		CohCibFinesGracePeriodEntryDailyAmortization       float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_daily_amortization"`
		CohCibFinesGracePeriodEntryDailyMaturity           float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_daily_maturity"`
		CohCibFinesGracePeriodEntryWeeklyAmortization      float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_weekly_amortization"`
		CohCibFinesGracePeriodEntryWeeklyMaturity          float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_weekly_maturity"`
		CohCibFinesGracePeriodEntryMonthlyAmortization     float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_monthly_amortization"`
		CohCibFinesGracePeriodEntryMonthlyMaturity         float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_monthly_maturity"`
		CohCibFinesGracePeriodEntrySemiMonthlyAmortization float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_semi_monthly_amortization"`
		CohCibFinesGracePeriodEntrySemiMonthlyMaturity     float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_semi_monthly_maturity"`
		CohCibFinesGracePeriodEntryQuarterlyAmortization   float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_quarterly_amortization"`
		CohCibFinesGracePeriodEntryQuarterlyMaturity       float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_quarterly_maturity"`
		CohCibFinesGracePeriodEntrySemiAnnualAmortization  float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_semi_annual_amortization"`
		CohCibFinesGracePeriodEntrySemiAnnualMaturity      float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_semi_annual_maturity"`
		CohCibFinesGracePeriodEntryAnnualAmortization      float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_annual_amortization"`
		CohCibFinesGracePeriodEntryAnnualMaturity          float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_annual_maturity"`
		CohCibFinesGracePeriodEntryLumpsumAmortization     float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_lumpsum_amortization"`
		CohCibFinesGracePeriodEntryLumpsumMaturity         float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_lumpsum_maturity"`
	}

	AccountHistoryResponse struct {
		ID             uuid.UUID             `json:"id"`
		CreatedAt      string                `json:"created_at"`
		UpdatedAt      string                `json:"updated_at"`
		CreatedByID    uuid.UUID             `json:"created_by_id"`
		CreatedBy      *UserResponse         `json:"created_by,omitempty"`
		AccountID      uuid.UUID             `json:"account_id"`
		Account        *AccountResponse      `json:"account,omitempty"`
		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       uuid.UUID             `json:"branch_id"`
		Branch         *BranchResponse       `json:"branch,omitempty"`

		Name        string      `json:"name"`
		Description string      `json:"description"`
		Type        AccountType `json:"type"`
		MinAmount   float64     `json:"min_amount"`
		MaxAmount   float64     `json:"max_amount"`
		Index       float64     `json:"index"`

		IsInternal         bool `json:"is_internal"`
		CashOnHand         bool `json:"cash_on_hand"`
		PaidUpShareCapital bool `json:"paid_up_share_capital"`

		ComputationType ComputationType `json:"computation_type"`

		FinesAmort       float64 `json:"fines_amort"`
		FinesMaturity    float64 `json:"fines_maturity"`
		InterestStandard float64 `json:"interest_standard"`
		InterestSecured  float64 `json:"interest_secured"`

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

		GeneralLedgerType GeneralLedgerType `json:"general_ledger_type"`

		HeaderRow int `json:"header_row"`
		CenterRow int `json:"center_row"`
		TotalRow  int `json:"total_row"`

		GeneralLedgerGroupingExcludeAccount bool   `json:"general_ledger_grouping_exclude_account"`
		Icon                                string `json:"icon"`

		ShowInGeneralLedgerSourceWithdraw       bool `json:"show_in_general_ledger_source_withdraw"`
		ShowInGeneralLedgerSourceDeposit        bool `json:"show_in_general_ledger_source_deposit"`
		ShowInGeneralLedgerSourceJournal        bool `json:"show_in_general_ledger_source_journal"`
		ShowInGeneralLedgerSourcePayment        bool `json:"show_in_general_ledger_source_payment"`
		ShowInGeneralLedgerSourceAdjustment     bool `json:"show_in_general_ledger_source_adjustment"`
		ShowInGeneralLedgerSourceJournalVoucher bool `json:"show_in_general_ledger_source_journal_voucher"`
		ShowInGeneralLedgerSourceCheckVoucher   bool `json:"show_in_general_ledger_source_check_voucher"`

		CompassionFund              bool                        `json:"compassion_fund"`
		CompassionFundAmount        float64                     `json:"compassion_fund_amount"`
		CashAndCashEquivalence      bool                        `json:"cash_and_cash_equivalence"`
		InterestStandardComputation InterestStandardComputation `json:"interest_standard_computation"`

		GeneralLedgerDefinitionID      *uuid.UUID `json:"general_ledger_definition_id,omitempty"`
		FinancialStatementDefinitionID *uuid.UUID `json:"financial_statement_definition_id,omitempty"`
		AccountClassificationID        *uuid.UUID `json:"account_classification_id,omitempty"`
		AccountCategoryID              *uuid.UUID `json:"account_category_id,omitempty"`
		MemberTypeID                   *uuid.UUID `json:"member_type_id,omitempty"`
		CurrencyID                     *uuid.UUID `json:"currency_id,omitempty"`
		DefaultPaymentTypeID           *uuid.UUID `json:"default_payment_type_id,omitempty"`
		ComputationSheetID             *uuid.UUID `json:"computation_sheet_id,omitempty"`
		LoanAccountID                  *uuid.UUID `json:"loan_account_id,omitempty"`

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
	}

	AccountHistoryRequest struct {
		AccountID uuid.UUID `json:"account_id" validate:"required"`
	}
)
