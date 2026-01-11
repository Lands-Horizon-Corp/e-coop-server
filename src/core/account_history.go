package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
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

func (m *Core) AccountHistoryManager() *registry.Registry[AccountHistory, AccountHistoryResponse, AccountHistoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		AccountHistory, AccountHistoryResponse, AccountHistoryRequest,
	]{
		Preloads: []string{"CreatedBy", "CreatedBy.Media", "Account", "Account.Currency", "Organization", "Branch"},
		Database: m.provider.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *AccountHistory) *AccountHistoryResponse {
			if data == nil {
				return nil
			}

			response := &AccountHistoryResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager().ToModel(data.CreatedBy),
				AccountID:      data.AccountID,
				Account:        m.AccountManager().ToModel(data.Account),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager().ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager().ToModel(data.Branch),

				Name:                         data.Name,
				Description:                  data.Description,
				Type:                         data.Type,
				MinAmount:                    data.MinAmount,
				MaxAmount:                    data.MaxAmount,
				Index:                        data.Index,
				IsInternal:                   data.IsInternal,
				CashOnHand:                   data.CashOnHand,
				PaidUpShareCapital:           data.PaidUpShareCapital,
				ComputationType:              data.ComputationType,
				FinesAmort:                   data.FinesAmort,
				FinesMaturity:                data.FinesMaturity,
				InterestStandard:             data.InterestStandard,
				InterestSecured:              data.InterestSecured,
				FinesGracePeriodAmortization: data.FinesGracePeriodAmortization,
				AdditionalGracePeriod:        data.AdditionalGracePeriod,
				NoGracePeriodDaily:           data.NoGracePeriodDaily,
				FinesGracePeriodMaturity:     data.FinesGracePeriodMaturity,
				YearlySubscriptionFee:        data.YearlySubscriptionFee,
				CutOffDays:                   data.CutOffDays,
				CutOffMonths:                 data.CutOffMonths,

				LumpsumComputationType:                            data.LumpsumComputationType,
				InterestFinesComputationDiminishing:               data.InterestFinesComputationDiminishing,
				InterestFinesComputationDiminishingStraightYearly: data.InterestFinesComputationDiminishingStraightYearly,
				EarnedUnearnedInterest:                            data.EarnedUnearnedInterest,
				LoanSavingType:                                    data.LoanSavingType,
				InterestDeduction:                                 data.InterestDeduction,
				OtherDeductionEntry:                               data.OtherDeductionEntry,
				InterestSavingTypeDiminishingStraight:             data.InterestSavingTypeDiminishingStraight,
				OtherInformationOfAnAccount:                       data.OtherInformationOfAnAccount,

				GeneralLedgerType:                   data.GeneralLedgerType,
				HeaderRow:                           data.HeaderRow,
				CenterRow:                           data.CenterRow,
				TotalRow:                            data.TotalRow,
				GeneralLedgerGroupingExcludeAccount: data.GeneralLedgerGroupingExcludeAccount,
				Icon:                                data.Icon,

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

				GeneralLedgerDefinitionID:      data.GeneralLedgerDefinitionID,
				FinancialStatementDefinitionID: data.FinancialStatementDefinitionID,
				AccountClassificationID:        data.AccountClassificationID,
				AccountCategoryID:              data.AccountCategoryID,
				MemberTypeID:                   data.MemberTypeID,
				CurrencyID:                     data.CurrencyID,
				DefaultPaymentTypeID:           data.DefaultPaymentTypeID,
				ComputationSheetID:             data.ComputationSheetID,
				LoanAccountID:                  data.LoanAccountID,

				CohCibFinesGracePeriodEntryCashHand:                data.CohCibFinesGracePeriodEntryCashHand,
				CohCibFinesGracePeriodEntryCashInBank:              data.CohCibFinesGracePeriodEntryCashInBank,
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
			}

			return response
		},
		Created: func(data *AccountHistory) registry.Topics {
			return []string{
				"account_history.create",
				fmt.Sprintf("account_history.create.%s", data.ID),
				fmt.Sprintf("account_history.create.account.%s", data.AccountID),
				fmt.Sprintf("account_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_history.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *AccountHistory) registry.Topics {
			return []string{
				"account_history.update",
				fmt.Sprintf("account_history.update.%s", data.ID),
				fmt.Sprintf("account_history.update.account.%s", data.AccountID),
			}
		},
		Deleted: func(data *AccountHistory) registry.Topics {
			return []string{
				"account_history.delete",
				fmt.Sprintf("account_history.delete.%s", data.ID),
			}
		},
	})
}

func (m *Core) AccountHistoryToModel(data *AccountHistory) *Account {
	if data == nil {
		return nil
	}
	return &Account{
		ID:               data.AccountID, // Use the original account ID
		AccountHistoryID: &data.ID,       // Use the original account ID
		CreatedAt:        data.CreatedAt,
		UpdatedAt:        data.UpdatedAt,
		DeletedAt:        gorm.DeletedAt{}, // History doesn't track deletion state of original

		OrganizationID: data.OrganizationID,
		Organization:   data.Organization,
		BranchID:       data.BranchID,
		Branch:         data.Branch,

		Name:        data.Name,
		Description: data.Description,
		Type:        data.Type,
		MinAmount:   data.MinAmount,
		MaxAmount:   data.MaxAmount,
		Index:       data.Index,

		IsInternal:         data.IsInternal,
		CashOnHand:         data.CashOnHand,
		PaidUpShareCapital: data.PaidUpShareCapital,

		ComputationType: data.ComputationType,

		FinesAmort:       data.FinesAmort,
		FinesMaturity:    data.FinesMaturity,
		InterestStandard: data.InterestStandard,
		InterestSecured:  data.InterestSecured,

		FinesGracePeriodAmortization: data.FinesGracePeriodAmortization,
		AdditionalGracePeriod:        data.AdditionalGracePeriod,
		NoGracePeriodDaily:           data.NoGracePeriodDaily,
		FinesGracePeriodMaturity:     data.FinesGracePeriodMaturity,
		YearlySubscriptionFee:        data.YearlySubscriptionFee,
		CutOffDays:                   data.CutOffDays,
		CutOffMonths:                 data.CutOffMonths,

		LumpsumComputationType:                            data.LumpsumComputationType,
		InterestFinesComputationDiminishing:               data.InterestFinesComputationDiminishing,
		InterestFinesComputationDiminishingStraightYearly: data.InterestFinesComputationDiminishingStraightYearly,
		EarnedUnearnedInterest:                            data.EarnedUnearnedInterest,
		LoanSavingType:                                    data.LoanSavingType,
		InterestDeduction:                                 data.InterestDeduction,
		OtherDeductionEntry:                               data.OtherDeductionEntry,
		InterestSavingTypeDiminishingStraight:             data.InterestSavingTypeDiminishingStraight,
		OtherInformationOfAnAccount:                       data.OtherInformationOfAnAccount,

		GeneralLedgerType: data.GeneralLedgerType,

		HeaderRow: data.HeaderRow,
		CenterRow: data.CenterRow,
		TotalRow:  data.TotalRow,

		GeneralLedgerGroupingExcludeAccount: data.GeneralLedgerGroupingExcludeAccount,
		Icon:                                data.Icon,

		ShowInGeneralLedgerSourceWithdraw:       data.ShowInGeneralLedgerSourceWithdraw,
		ShowInGeneralLedgerSourceDeposit:        data.ShowInGeneralLedgerSourceDeposit,
		ShowInGeneralLedgerSourceJournal:        data.ShowInGeneralLedgerSourceJournal,
		ShowInGeneralLedgerSourcePayment:        data.ShowInGeneralLedgerSourcePayment,
		ShowInGeneralLedgerSourceAdjustment:     data.ShowInGeneralLedgerSourceAdjustment,
		ShowInGeneralLedgerSourceJournalVoucher: data.ShowInGeneralLedgerSourceJournalVoucher,
		ShowInGeneralLedgerSourceCheckVoucher:   data.ShowInGeneralLedgerSourceCheckVoucher,

		CompassionFund:         data.CompassionFund,
		CompassionFundAmount:   data.CompassionFundAmount,
		CashAndCashEquivalence: data.CashAndCashEquivalence,

		InterestStandardComputation: data.InterestStandardComputation,

		GeneralLedgerDefinitionID:      data.GeneralLedgerDefinitionID,
		FinancialStatementDefinitionID: data.FinancialStatementDefinitionID,
		AccountClassificationID:        data.AccountClassificationID,
		AccountCategoryID:              data.AccountCategoryID,
		MemberTypeID:                   data.MemberTypeID,
		CurrencyID:                     data.CurrencyID,
		DefaultPaymentTypeID:           data.DefaultPaymentTypeID,
		ComputationSheetID:             data.ComputationSheetID,
		LoanAccountID:                  data.LoanAccountID,

		CohCibFinesGracePeriodEntryCashHand:                data.CohCibFinesGracePeriodEntryCashHand,
		CohCibFinesGracePeriodEntryCashInBank:              data.CohCibFinesGracePeriodEntryCashInBank,
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
	}
}

func (m *Core) GetAccountHistory(ctx context.Context, accountID uuid.UUID) ([]*AccountHistory, error) {
	filters := []registry.FilterSQL{
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
	}

	return m.AccountHistoryManager().ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func (m *Core) GetAllAccountHistory(ctx context.Context, accountID, organizationID, branchID uuid.UUID) ([]*AccountHistory, error) {
	filters := []registry.FilterSQL{
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return m.AccountHistoryManager().ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "created_at", Order: query.SortOrderDesc}, // Latest first
		{Field: "updated_at", Order: query.SortOrderDesc}, // Secondary sort
	})
}

func (m *Core) GetAccountHistoryLatestByTime(
	ctx context.Context,
	accountID, organizationID, branchID uuid.UUID,
	asOfDate *time.Time) (*Account, error) {
	currentTime := time.Now()
	if asOfDate == nil {
		asOfDate = &currentTime
	}
	filters := []registry.FilterSQL{
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "created_at", Op: query.ModeLTE, Value: asOfDate},
	}

	histories, err := m.AccountHistoryManager().ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "created_at", Order: query.SortOrderDesc}, // Latest first
		{Field: "updated_at", Order: query.SortOrderDesc}, // Secondary sort
	})
	if err != nil {
		return nil, err
	}

	if len(histories) > 0 {
		return m.AccountHistoryToModel(histories[0]), nil
	}

	return nil, eris.Errorf("no history found for account %s at time %s", accountID, asOfDate.Format(time.RFC3339))
}

func (m *Core) GetAccountHistoryLatestByTimeHistoryID(
	ctx context.Context,
	accountID, organizationID, branchID uuid.UUID,
	asOfDate *time.Time) (*uuid.UUID, error) {
	currentTime := time.Now()
	if asOfDate == nil {
		asOfDate = &currentTime
	}
	filters := []registry.FilterSQL{
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "created_at", Op: query.ModeLTE, Value: asOfDate},
	}

	history, err := m.AccountHistoryManager().ArrFindOne(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "created_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, eris.Errorf("no history found for account %s at time %s", accountID, asOfDate.Format(time.RFC3339))
	}

	return &history.ID, nil
}

func (m *Core) GetAccountHistoryLatestByTimeHistory(
	ctx context.Context,
	accountID, organizationID, branchID uuid.UUID,
	asOfDate *time.Time) (*AccountHistory, error) {
	currentTime := time.Now()
	if asOfDate == nil {
		asOfDate = &currentTime
	}
	filters := []registry.FilterSQL{
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "created_at", Op: query.ModeLTE, Value: asOfDate},
	}

	histories, err := m.AccountHistoryManager().ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "created_at", Order: query.SortOrderDesc}, // Latest first
		{Field: "updated_at", Order: query.SortOrderDesc}, // Secondary sort
	})
	if err != nil {
		return nil, err
	}

	if len(histories) > 0 {
		return histories[0], nil
	}

	return nil, eris.Errorf("no history found for account %s at time %s", accountID, asOfDate.Format(time.RFC3339))
}

func (m *Core) GetAccountHistoriesByFiltersAtTime(
	ctx context.Context,
	organizationID, branchID uuid.UUID,
	asOfDate *time.Time,
	loanAccountID *uuid.UUID,
	currencyID *uuid.UUID,
) ([]*Account, error) {
	currentTime := time.Now()
	if asOfDate == nil {
		asOfDate = &currentTime
	}

	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "created_at", Op: query.ModeLTE, Value: asOfDate},
	}

	if loanAccountID != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "loan_account_id", Op: query.ModeEqual, Value: *loanAccountID,
		})
	}

	if currencyID != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "currency_id", Op: query.ModeEqual, Value: *currencyID,
		})
	}

	histories, err := m.AccountHistoryManager().ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "account_id", Order: query.SortOrderAsc},
		{Field: "created_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}

	accountMap := make(map[uuid.UUID]*AccountHistory)
	for _, history := range histories {
		if existing, found := accountMap[history.AccountID]; !found || history.CreatedAt.After(existing.CreatedAt) {
			accountMap[history.AccountID] = history
		}
	}

	var accounts []*Account
	for _, history := range accountMap {
		accounts = append(accounts, m.AccountHistoryToModel(history))
	}

	return accounts, nil
}
