package modelcore

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- ENUMS FOR HISTORY ---

type HistoryChangeType string

const (
	HistoryChangeTypeCreated HistoryChangeType = "created"
	HistoryChangeTypeUpdated HistoryChangeType = "updated"
	HistoryChangeTypeDeleted HistoryChangeType = "deleted"
)

// --- ACCOUNT HISTORY MODEL ---

type (
	AccountHistory struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`

		// Reference to the original account
		AccountID uuid.UUID `gorm:"type:uuid;not null;index:idx_account_history_account" json:"account_id"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"account,omitempty"`

		// Organization and branch for filtering
		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_account_history_org_branch" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_account_history_org_branch" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		// History metadata
		ChangeType    HistoryChangeType `gorm:"type:varchar(50);not null" json:"change_type"`
		ValidFrom     time.Time         `gorm:"not null;index:idx_account_history_valid_from" json:"valid_from"`
		ValidTo       *time.Time        `gorm:"index:idx_account_history_valid_to" json:"valid_to,omitempty"`
		ChangeReason  string            `gorm:"type:text" json:"change_reason,omitempty"`
		ChangedFields string            `gorm:"type:text" json:"changed_fields,omitempty"`

		// Snapshot of account data at the time of change
		Name        string      `gorm:"type:varchar(255)" json:"name"`
		Description string      `gorm:"type:text" json:"description"`
		Type        AccountType `gorm:"type:varchar(50)" json:"type"`
		MinAmount   float64     `gorm:"type:decimal" json:"min_amount"`
		MaxAmount   float64     `gorm:"type:decimal" json:"max_amount"`
		Index       int         `gorm:"default:0" json:"index"`

		IsInternal         bool `gorm:"default:false" json:"is_internal"`
		CashOnHand         bool `gorm:"default:false" json:"cash_on_hand"`
		PaidUpShareCapital bool `gorm:"default:false" json:"paid_up_share_capital"`

		ComputationType ComputationType `gorm:"type:varchar(50)" json:"computation_type"`

		// Interest and fees snapshot
		FinesAmort       float64 `gorm:"type:decimal" json:"fines_amort"`
		FinesMaturity    float64 `gorm:"type:decimal" json:"fines_maturity"`
		InterestStandard float64 `gorm:"type:decimal" json:"interest_standard"`
		InterestSecured  float64 `gorm:"type:decimal" json:"interest_secured"`

		// Grace periods snapshot
		FinesGracePeriodAmortization int  `gorm:"type:int" json:"fines_grace_period_amortization"`
		AdditionalGracePeriod        int  `gorm:"type:int" json:"additional_grace_period"`
		NumberGracePeriodDaily       bool `gorm:"default:false" json:"number_grace_period_daily"`
		FinesGracePeriodMaturity     int  `gorm:"type:int" json:"fines_grace_period_maturity"`
		YearlySubscriptionFee        int  `gorm:"type:int" json:"yearly_subscription_fee"`
		LoanCutOffDays               int  `gorm:"type:int" json:"loan_cut_off_days"`

		// Configuration snapshot
		LumpsumComputationType                            LumpsumComputationType                            `gorm:"type:varchar(50)" json:"lumpsum_computation_type"`
		InterestFinesComputationDiminishing               InterestFinesComputationDiminishing               `gorm:"type:varchar(100)" json:"interest_fines_computation_diminishing"`
		InterestFinesComputationDiminishingStraightYearly InterestFinesComputationDiminishingStraightYearly `gorm:"type:varchar(200)" json:"interest_fines_computation_diminishing_straight_yearly"`
		EarnedUnearnedInterest                            EarnedUnearnedInterest                            `gorm:"type:varchar(50)" json:"earned_unearned_interest"`
		LoanSavingType                                    LoanSavingType                                    `gorm:"type:varchar(50)" json:"loan_saving_type"`
		InterestDeduction                                 InterestDeduction                                 `gorm:"type:varchar(10)" json:"interest_deduction"`
		OtherDeductionEntry                               OtherDeductionEntry                               `gorm:"type:varchar(20)" json:"other_deduction_entry"`
		InterestSavingTypeDiminishingStraight             InterestSavingTypeDiminishingStraight             `gorm:"type:varchar(20)" json:"interest_saving_type_diminishing_straight"`
		OtherInformationOfAnAccount                       OtherInformationOfAnAccount                       `gorm:"type:varchar(50)" json:"other_information_of_an_account"`

		FinancialStatementType FinancialStatementType `gorm:"type:varchar(50)" json:"financial_statement_type"`
		GeneralLedgerType      GeneralLedgerType      `gorm:"type:varchar(50)" json:"general_ledger_type"`

		// Display configuration snapshot
		HeaderRow int `gorm:"type:int" json:"header_row"`
		CenterRow int `gorm:"type:int" json:"center_row"`
		TotalRow  int `gorm:"type:int" json:"total_row"`

		GeneralLedgerGroupingExcludeAccount bool   `gorm:"default:false" json:"general_ledger_grouping_exclude_account"`
		Icon                                string `gorm:"type:varchar(50)" json:"icon"`

		// General Ledger Source flags snapshot
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

		// Foreign key references (stored as IDs for history)
		GeneralLedgerDefinitionID      *uuid.UUID `gorm:"type:uuid" json:"general_ledger_definition_id,omitempty"`
		FinancialStatementDefinitionID *uuid.UUID `gorm:"type:uuid" json:"financial_statement_definition_id,omitempty"`
		AccountClassificationID        *uuid.UUID `gorm:"type:uuid" json:"account_classification_id,omitempty"`
		AccountCategoryID              *uuid.UUID `gorm:"type:uuid" json:"account_category_id,omitempty"`
		MemberTypeID                   *uuid.UUID `gorm:"type:uuid" json:"member_type_id,omitempty"`
		CurrencyID                     *uuid.UUID `gorm:"type:uuid" json:"currency_id,omitempty"`
		ComputationSheetID             *uuid.UUID `gorm:"type:uuid" json:"computation_sheet_id,omitempty"`
		AlternativeAccountID           *uuid.UUID `gorm:"type:uuid" json:"alternative_account_id,omitempty"`

		// Grace period entries snapshot
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
		CohCibFinesGracePeriodEntrySemiAnualAmortization   float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_semi_anual_amortization"`
		CohCibFinesGracePeriodEntrySemiAnualMaturity       float64 `gorm:"type:decimal" json:"coh_cib_fines_grace_period_entry_semi_anual_maturity"`
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

		ChangeType    HistoryChangeType `json:"change_type"`
		ValidFrom     string            `json:"valid_from"`
		ValidTo       *string           `json:"valid_to,omitempty"`
		ChangeReason  string            `json:"change_reason,omitempty"`
		ChangedFields string            `json:"changed_fields,omitempty"`

		// Account snapshot data
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Type        AccountType `json:"type"`
		MinAmount   float64     `json:"min_amount"`
		MaxAmount   float64     `json:"max_amount"`
		Index       int         `json:"index"`

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
		NumberGracePeriodDaily       bool `json:"number_grace_period_daily"`
		FinesGracePeriodMaturity     int  `json:"fines_grace_period_maturity"`
		YearlySubscriptionFee        int  `json:"yearly_subscription_fee"`
		LoanCutOffDays               int  `json:"loan_cut_off_days"`

		LumpsumComputationType                            LumpsumComputationType                            `json:"lumpsum_computation_type"`
		InterestFinesComputationDiminishing               InterestFinesComputationDiminishing               `json:"interest_fines_computation_diminishing"`
		InterestFinesComputationDiminishingStraightYearly InterestFinesComputationDiminishingStraightYearly `json:"interest_fines_computation_diminishing_straight_yearly"`
		EarnedUnearnedInterest                            EarnedUnearnedInterest                            `json:"earned_unearned_interest"`
		LoanSavingType                                    LoanSavingType                                    `json:"loan_saving_type"`
		InterestDeduction                                 InterestDeduction                                 `json:"interest_deduction"`
		OtherDeductionEntry                               OtherDeductionEntry                               `json:"other_deduction_entry"`
		InterestSavingTypeDiminishingStraight             InterestSavingTypeDiminishingStraight             `json:"interest_saving_type_diminishing_straight"`
		OtherInformationOfAnAccount                       OtherInformationOfAnAccount                       `json:"other_information_of_an_account"`

		FinancialStatementType FinancialStatementType `json:"financial_statement_type"`
		GeneralLedgerType      GeneralLedgerType      `json:"general_ledger_type"`

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

		// Foreign key references
		GeneralLedgerDefinitionID      *uuid.UUID `json:"general_ledger_definition_id,omitempty"`
		FinancialStatementDefinitionID *uuid.UUID `json:"financial_statement_definition_id,omitempty"`
		AccountClassificationID        *uuid.UUID `json:"account_classification_id,omitempty"`
		AccountCategoryID              *uuid.UUID `json:"account_category_id,omitempty"`
		MemberTypeID                   *uuid.UUID `json:"member_type_id,omitempty"`
		CurrencyID                     *uuid.UUID `json:"currency_id,omitempty"`
		ComputationSheetID             *uuid.UUID `json:"computation_sheet_id,omitempty"`
		AlternativeAccountID           *uuid.UUID `json:"alternative_account_id,omitempty"`

		// Grace period entries
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
	}

	AccountHistoryRequest struct {
		AccountID     uuid.UUID         `json:"account_id" validate:"required"`
		ChangeType    HistoryChangeType `json:"change_type" validate:"required"`
		ValidFrom     time.Time         `json:"valid_from" validate:"required"`
		ValidTo       *time.Time        `json:"valid_to,omitempty"`
		ChangeReason  string            `json:"change_reason,omitempty"`
		ChangedFields string            `json:"changed_fields,omitempty"`
	}
)

// --- REGISTRATION ---

func (m *ModelCore) accountHistory() {
	m.Migration = append(m.Migration, &AccountHistory{})
	m.AccountHistoryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		AccountHistory, AccountHistoryResponse, AccountHistoryRequest,
	]{
		Preloads: []string{"CreatedBy", "CreatedBy.Media", "Account", "Organization", "Branch"},
		Service:  m.provider.Service,
		Resource: func(data *AccountHistory) *AccountHistoryResponse {
			if data == nil {
				return nil
			}

			response := &AccountHistoryResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				AccountID:      data.AccountID,
				Account:        m.AccountManager.ToModel(data.Account),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),
				ChangeType:     data.ChangeType,
				ValidFrom:      data.ValidFrom.Format(time.RFC3339),
				ChangeReason:   data.ChangeReason,
				ChangedFields:  data.ChangedFields,

				// Account snapshot data
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
				NumberGracePeriodDaily:       data.NumberGracePeriodDaily,
				FinesGracePeriodMaturity:     data.FinesGracePeriodMaturity,
				YearlySubscriptionFee:        data.YearlySubscriptionFee,
				LoanCutOffDays:               data.LoanCutOffDays,

				LumpsumComputationType:                            LumpsumComputationType(data.LumpsumComputationType),
				InterestFinesComputationDiminishing:               InterestFinesComputationDiminishing(data.InterestFinesComputationDiminishing),
				InterestFinesComputationDiminishingStraightYearly: InterestFinesComputationDiminishingStraightYearly(data.InterestFinesComputationDiminishingStraightYearly),
				EarnedUnearnedInterest:                            EarnedUnearnedInterest(data.EarnedUnearnedInterest),
				LoanSavingType:                                    LoanSavingType(data.LoanSavingType),
				InterestDeduction:                                 InterestDeduction(data.InterestDeduction),
				OtherDeductionEntry:                               OtherDeductionEntry(data.OtherDeductionEntry),
				InterestSavingTypeDiminishingStraight:             InterestSavingTypeDiminishingStraight(data.InterestSavingTypeDiminishingStraight),
				OtherInformationOfAnAccount:                       OtherInformationOfAnAccount(data.OtherInformationOfAnAccount),

				FinancialStatementType:              FinancialStatementType(data.FinancialStatementType),
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

				// Foreign key references
				GeneralLedgerDefinitionID:      data.GeneralLedgerDefinitionID,
				FinancialStatementDefinitionID: data.FinancialStatementDefinitionID,
				AccountClassificationID:        data.AccountClassificationID,
				AccountCategoryID:              data.AccountCategoryID,
				MemberTypeID:                   data.MemberTypeID,
				CurrencyID:                     data.CurrencyID,
				ComputationSheetID:             data.ComputationSheetID,
				AlternativeAccountID:           data.AlternativeAccountID,

				// Grace period entries
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
				CohCibFinesGracePeriodEntrySemiAnualAmortization:   data.CohCibFinesGracePeriodEntrySemiAnualAmortization,
				CohCibFinesGracePeriodEntrySemiAnualMaturity:       data.CohCibFinesGracePeriodEntrySemiAnualMaturity,
				CohCibFinesGracePeriodEntryLumpsumAmortization:     data.CohCibFinesGracePeriodEntryLumpsumAmortization,
				CohCibFinesGracePeriodEntryLumpsumMaturity:         data.CohCibFinesGracePeriodEntryLumpsumMaturity,
			}

			// Handle ValidTo field
			if data.ValidTo != nil {
				validTo := data.ValidTo.Format(time.RFC3339)
				response.ValidTo = &validTo
			}

			return response
		},
		Created: func(data *AccountHistory) []string {
			return []string{
				"account_history.create",
				fmt.Sprintf("account_history.create.%s", data.ID),
				fmt.Sprintf("account_history.create.account.%s", data.AccountID),
				fmt.Sprintf("account_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_history.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *AccountHistory) []string {
			return []string{
				"account_history.update",
				fmt.Sprintf("account_history.update.%s", data.ID),
				fmt.Sprintf("account_history.update.account.%s", data.AccountID),
			}
		},
		Deleted: func(data *AccountHistory) []string {
			return []string{
				"account_history.delete",
				fmt.Sprintf("account_history.delete.%s", data.ID),
			}
		},
	})
}

// Get history for a specific account
func (m *ModelCore) getAccountHistory(ctx context.Context, accountID uuid.UUID) ([]*AccountHistory, error) {
	filters := []horizon_services.Filter{
		{Field: "account_id", Op: horizon_services.OpEq, Value: accountID},
	}

	return m.AccountHistoryManager.FindWithFilters(ctx, filters)
}

func (m *ModelCore) getAccountAtTime(ctx context.Context, accountID uuid.UUID, asOfDate time.Time) (*AccountHistory, error) {
	filters := []horizon_services.Filter{
		{Field: "account_id", Op: horizon_services.OpEq, Value: accountID},
		{Field: "valid_from", Op: horizon_services.OpLte, Value: asOfDate},
		{Field: "valid_to", Op: horizon_services.OpGt, Value: asOfDate},
	}

	histories, err := m.AccountHistoryManager.FindWithFilters(ctx, filters)
	if err != nil {
		return nil, err
	}

	if len(histories) > 0 {
		return histories[0], nil
	}

	// If no history with valid_to > asOfDate, get the latest one before asOfDate
	filters = []horizon_services.Filter{
		{Field: "account_id", Op: horizon_services.OpEq, Value: accountID},
		{Field: "valid_from", Op: horizon_services.OpLte, Value: asOfDate},
	}

	histories, err = m.AccountHistoryManager.FindWithFilters(ctx, filters)
	if err != nil {
		return nil, err
	}

	if len(histories) > 0 {
		return histories[0], nil
	}

	return nil, fmt.Errorf("no history found for account %s at time %s", accountID, asOfDate.Format(time.RFC3339))
}

// Get all accounts that had changes within a date range
func (m *ModelCore) getAccountsChangedInRange(ctx context.Context, orgID, branchID uuid.UUID, startDate, endDate time.Time) ([]*AccountHistory, error) {
	filters := []horizon_services.Filter{
		{Field: "organization_id", Op: horizon_services.OpEq, Value: orgID},
		{Field: "branch_id", Op: horizon_services.OpEq, Value: branchID},
		{Field: "valid_from", Op: horizon_services.OpGte, Value: startDate},
		{Field: "valid_from", Op: horizon_services.OpLte, Value: endDate},
	}

	return m.AccountHistoryManager.FindWithFilters(ctx, filters)
}

// Close open history records when updating valid_to
func (m *ModelCore) closeAccountHistory(ctx context.Context, accountID uuid.UUID, closedAt time.Time) error {
	// Since there's no UpdateWhere method, we'll need to find and update individually
	filters := []horizon_services.Filter{
		{Field: "account_id", Op: horizon_services.OpEq, Value: accountID},
		{Field: "valid_to", Op: horizon_services.OpIsNull, Value: nil},
	}

	histories, err := m.AccountHistoryManager.FindWithFilters(ctx, filters)
	if err != nil {
		return err
	}

	for _, history := range histories {
		history.ValidTo = &closedAt
		if err := m.AccountHistoryManager.Update(ctx, history); err != nil {
			return err
		}
	}

	return nil
}
