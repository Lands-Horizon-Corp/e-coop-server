package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	BranchSetting struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt time.Time      `gorm:"not null;default:now()" json:"created_at"`
		UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

		BranchID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"branch_id"`
		Branch   *Branch   `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		CurrencyID uuid.UUID `gorm:"type:uuid;not null" json:"currency_id"`
		Currency   *Currency `gorm:"foreignKey:CurrencyID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"currency,omitempty"`

		CashOnHandAccountID *uuid.UUID `gorm:"type:uuid" json:"cash_on_hand_account_id,omitempty"`
		CashOnHandAccount   *Account   `gorm:"foreignKey:CashOnHandAccountID;constraint:OnDelete:SET NULL;" json:"cash_on_hand_account,omitempty"`

		PaidUpSharedCapitalAccountID *uuid.UUID `gorm:"type:uuid" json:"paid_up_shared_capital_account_id,omitempty"`
		PaidUpSharedCapitalAccount   *Account   `gorm:"foreignKey:PaidUpSharedCapitalAccountID;constraint:OnDelete:SET NULL;" json:"paid_up_shared_capital_account,omitempty"`

		CompassionFundAccountID *uuid.UUID `gorm:"type:uuid" json:"compassion_fund_account_id,omitempty"`
		CompassionFundAccount   *Account   `gorm:"foreignKey:CompassionFundAccountID;constraint:OnDelete:SET NULL;" json:"compassion_fund_account,omitempty"`

		WithdrawAllowUserInput bool   `gorm:"not null;default:true" json:"withdraw_allow_user_input"`
		WithdrawPrefix         string `gorm:"type:varchar(50);not null;default:'WD'" json:"withdraw_prefix"`
		WithdrawORStart        int    `gorm:"not null;default:0" json:"withdraw_or_start"`
		WithdrawORCurrent      int    `gorm:"not null;default:1" json:"withdraw_or_current"`
		WithdrawOREnd          int    `gorm:"not null;default:9999" json:"withdraw_or_end"`
		WithdrawORIteration    int    `gorm:"not null;default:1" json:"withdraw_or_iteration"`
		WithdrawUseDateOR      bool   `gorm:"not null;default:false" json:"withdraw_use_date_or"`
		WithdrawPadding        int    `gorm:"not null;default:6" json:"withdraw_padding"`
		WithdrawCommonOR       string `gorm:"type:varchar(100)" json:"withdraw_common_or"`

		DepositORStart     int    `gorm:"not null;default:0" json:"deposit_or_start"`
		DepositORCurrent   int    `gorm:"not null;default:1" json:"deposit_or_current"`
		DepositOREnd       int    `gorm:"not null;default:9999" json:"deposit_or_end"`
		DepositORIteration int    `gorm:"not null;default:1" json:"deposit_or_iteration"`
		DepositUseDateOR   bool   `gorm:"not null;default:false" json:"deposit_use_date_or"`
		DepositPadding     int    `gorm:"not null;default:6" json:"deposit_padding"`
		DepositCommonOR    string `gorm:"type:varchar(100)" json:"deposit_common_or"`

		CashCheckVoucherAllowUserInput bool   `gorm:"not null;default:true" json:"cash_check_voucher_allow_user_input"`
		CashCheckVoucherORUnique       bool   `gorm:"not null;default:false" json:"cash_check_voucher_or_unique"`
		CashCheckVoucherPrefix         string `gorm:"type:varchar(50);not null;default:'CCV'" json:"cash_check_voucher_prefix"`
		CashCheckVoucherORStart        int    `gorm:"not null;default:0" json:"cash_check_voucher_or_start"`
		CashCheckVoucherORCurrent      int    `gorm:"not null;default:1" json:"cash_check_voucher_or_current"`
		CashCheckVoucherORIteration    int    `gorm:"not null;default:1" json:"cash_check_voucher_or_iteration"`
		CashCheckVoucherPadding        int    `gorm:"not null;default:6" json:"cash_check_voucher_padding"`

		JournalVoucherAllowUserInput bool   `gorm:"not null;default:true" json:"journal_voucher_allow_user_input"`
		JournalVoucherORUnique       bool   `gorm:"not null;default:false" json:"journal_voucher_or_unique"`
		JournalVoucherPrefix         string `gorm:"type:varchar(50);not null;default:'JV'" json:"journal_voucher_prefix"`
		JournalVoucherORStart        int    `gorm:"not null;default:0" json:"journal_voucher_or_start"`
		JournalVoucherORCurrent      int    `gorm:"not null;default:1" json:"journal_voucher_or_current"`
		JournalVoucherORIteration    int    `gorm:"not null;default:1" json:"journal_voucher_or_iteration"`
		JournalVoucherPadding        int    `gorm:"not null;default:6" json:"journal_voucher_padding"`

		AdjustmentVoucherAllowUserInput bool   `gorm:"not null;default:true" json:"adjustment_voucher_allow_user_input"`
		AdjustmentVoucherORUnique       bool   `gorm:"not null;default:false" json:"adjustment_voucher_or_unique"`
		AdjustmentVoucherPrefix         string `gorm:"type:varchar(50);not null;default:'AV'" json:"adjustment_voucher_prefix"`
		AdjustmentVoucherORStart        int    `gorm:"not null;default:0" json:"adjustment_voucher_or_start"`
		AdjustmentVoucherORCurrent      int    `gorm:"not null;default:1" json:"adjustment_voucher_or_current"`
		AdjustmentVoucherORIteration    int    `gorm:"not null;default:1" json:"adjustment_voucher_or_iteration"`
		AdjustmentVoucherPadding        int    `gorm:"not null;default:6" json:"adjustment_voucher_padding"`

		LoanVoucherAllowUserInput bool   `gorm:"not null;default:true" json:"loan_voucher_allow_user_input"`
		LoanVoucherORUnique       bool   `gorm:"not null;default:false" json:"loan_voucher_or_unique"`
		LoanVoucherPrefix         string `gorm:"type:varchar(50);not null;default:'LV'" json:"loan_voucher_prefix"`
		LoanVoucherORStart        int    `gorm:"not null;default:0" json:"loan_voucher_or_start"`
		LoanVoucherORCurrent      int    `gorm:"not null;default:1" json:"loan_voucher_or_current"`
		LoanVoucherORIteration    int    `gorm:"not null;default:1" json:"loan_voucher_or_iteration"`
		LoanVoucherPadding        int    `gorm:"not null;default:6" json:"loan_voucher_padding"`

		CheckVoucherGeneral               bool    `gorm:"not null;default:false" json:"check_voucher_general"`
		CheckVoucherGeneralAllowUserInput bool    `gorm:"not null;default:true" json:"check_voucher_general_allow_user_input"`
		CheckVoucherGeneralORUnique       bool    `gorm:"not null;default:false" json:"check_voucher_general_or_unique"`
		CheckVoucherGeneralPrefix         string  `gorm:"type:varchar(50);not null;default:'CV'" json:"check_voucher_general_prefix"`
		CheckVoucherGeneralORStart        int     `gorm:"not null;default:0" json:"check_voucher_general_or_start"`
		CheckVoucherGeneralORCurrent      int     `gorm:"not null;default:1" json:"check_voucher_general_or_current"`
		CheckVoucherGeneralORIteration    int     `gorm:"not null;default:1" json:"check_voucher_general_or_iteration"`
		CheckVoucherGeneralPadding        int     `gorm:"not null;default:6" json:"check_voucher_general_padding"`
		TaxInterest                       float64 `gorm:"not null;default:0" json:"tax_interest"`

		DefaultMemberGenderID *uuid.UUID    `gorm:"type:uuid" json:"default_member_gender_id,omitempty"`
		DefaultMemberGender   *MemberGender `gorm:"foreignKey:member_gender;constraint:OnDelete:SET NULL;" json:"default_member_gender,omitempty"`

		DefaultMemberTypeID *uuid.UUID           `gorm:"type:uuid" json:"default_member_type_id,omitempty"`
		DefaultMemberType   *MemberType          `gorm:"foreignKey:DefaultMemberTypeID;constraint:OnDelete:SET NULL;" json:"default_member_type,omitempty"`
		AnnualDivisor       int                  `gorm:"not null;default:360" json:"annual_divisor"`
		UnbalancedAccounts  []*UnbalancedAccount `gorm:"foreignKey:BranchSettingsID;constraint:OnDelete:CASCADE;" json:"unbalanced_accounts,omitempty"`
	}

	BranchSettingRequest struct {
		WithdrawAllowUserInput bool   `json:"withdraw_allow_user_input"`
		WithdrawPrefix         string `json:"withdraw_prefix" validate:"omitempty"`
		WithdrawORStart        int    `json:"withdraw_or_start" validate:"min=0"`
		WithdrawORCurrent      int    `json:"withdraw_or_current" validate:"min=0"`
		WithdrawOREnd          int    `json:"withdraw_or_end" validate:"min=0"`
		WithdrawORIteration    int    `json:"withdraw_or_iteration" validate:"min=0"`
		WithdrawUseDateOR      bool   `json:"withdraw_use_date_or"`
		WithdrawPadding        int    `json:"withdraw_padding" validate:"min=0"`
		WithdrawCommonOR       string `json:"withdraw_common_or" validate:"omitempty"`

		DepositORStart     int    `json:"deposit_or_start" validate:"min=0"`
		DepositORCurrent   int    `json:"deposit_or_current" validate:"min=0"`
		DepositOREnd       int    `json:"deposit_or_end" validate:"min=0"`
		DepositORIteration int    `json:"deposit_or_iteration" validate:"min=0"`
		DepositUseDateOR   bool   `json:"deposit_use_date_or"`
		DepositPadding     int    `json:"deposit_padding" validate:"min=0"`
		DepositCommonOR    string `json:"deposit_common_or" validate:"omitempty"`

		CashCheckVoucherAllowUserInput bool   `json:"cash_check_voucher_allow_user_input"`
		CashCheckVoucherORUnique       bool   `json:"cash_check_voucher_or_unique"`
		CashCheckVoucherPrefix         string `json:"cash_check_voucher_prefix" validate:"omitempty"`
		CashCheckVoucherORStart        int    `json:"cash_check_voucher_or_start" validate:"min=0"`
		CashCheckVoucherORCurrent      int    `json:"cash_check_voucher_or_current" validate:"min=0"`
		CashCheckVoucherORIteration    int    `json:"cash_check_voucher_or_iteration" validate:"min=0"`
		CashCheckVoucherPadding        int    `json:"cash_check_voucher_padding" validate:"min=0"`

		JournalVoucherAllowUserInput bool   `json:"journal_voucher_allow_user_input"`
		JournalVoucherORUnique       bool   `json:"journal_voucher_or_unique"`
		JournalVoucherPrefix         string `json:"journal_voucher_prefix" validate:"omitempty"`
		JournalVoucherORStart        int    `json:"journal_voucher_or_start" validate:"min=0"`
		JournalVoucherORCurrent      int    `json:"journal_voucher_or_current" validate:"min=0"`
		JournalVoucherORIteration    int    `json:"journal_voucher_or_iteration" validate:"min=0"`
		JournalVoucherPadding        int    `json:"journal_voucher_padding" validate:"min=0"`

		AdjustmentVoucherAllowUserInput bool   `json:"adjustment_voucher_allow_user_input"`
		AdjustmentVoucherORUnique       bool   `json:"adjustment_voucher_or_unique"`
		AdjustmentVoucherPrefix         string `json:"adjustment_voucher_prefix" validate:"omitempty"`
		AdjustmentVoucherORStart        int    `json:"adjustment_voucher_or_start" validate:"min=0"`
		AdjustmentVoucherORCurrent      int    `json:"adjustment_voucher_or_current" validate:"min=0"`
		AdjustmentVoucherORIteration    int    `json:"adjustment_voucher_or_iteration" validate:"min=0"`
		AdjustmentVoucherPadding        int    `json:"adjustment_voucher_padding" validate:"min=0"`

		LoanVoucherAllowUserInput bool   `json:"loan_voucher_allow_user_input"`
		LoanVoucherORUnique       bool   `json:"loan_voucher_or_unique"`
		LoanVoucherPrefix         string `json:"loan_voucher_prefix" validate:"omitempty"`
		LoanVoucherORStart        int    `json:"loan_voucher_or_start" validate:"min=0"`
		LoanVoucherORCurrent      int    `json:"loan_voucher_or_current" validate:"min=0"`
		LoanVoucherORIteration    int    `json:"loan_voucher_or_iteration" validate:"min=0"`
		LoanVoucherPadding        int    `json:"loan_voucher_padding" validate:"min=0"`

		CheckVoucherGeneral               bool   `json:"check_voucher_general"`
		CheckVoucherGeneralAllowUserInput bool   `json:"check_voucher_general_allow_user_input"`
		CheckVoucherGeneralORUnique       bool   `json:"check_voucher_general_or_unique"`
		CheckVoucherGeneralPrefix         string `json:"check_voucher_general_prefix" validate:"omitempty"`
		CheckVoucherGeneralORStart        int    `json:"check_voucher_general_or_start" validate:"min=0"`
		CheckVoucherGeneralORCurrent      int    `json:"check_voucher_general_or_current" validate:"min=0"`
		CheckVoucherGeneralORIteration    int    `json:"check_voucher_general_or_iteration" validate:"min=0"`
		CheckVoucherGeneralPadding        int    `json:"check_voucher_general_padding" validate:"min=0"`

		DefaultMemberGenderID *uuid.UUID `json:"default_member_gender_id,omitempty"`
		DefaultMemberTypeID   *uuid.UUID `json:"default_member_type_id,omitempty"`

		AnnualDivisor int     `json:"annual_divisor" validate:"min=0"`
		TaxInterest   float64 `json:"tax_interest" validate:"min=0"`
	}

	BranchSettingsCurrencyRequest struct {
		CurrencyID                   uuid.UUID  `json:"currency_id" validate:"required"`
		PaidUpSharedCapitalAccountID uuid.UUID  `json:"paid_up_shared_capital_account_id" validate:"required"`
		CashOnHandAccountID          uuid.UUID  `json:"cash_on_hand_account_id" validate:"required"`
		CompassionFundAccountID      *uuid.UUID `json:"compassion_fund_account_id,omitempty"`

		UnbalancedAccount          []UnbalancedAccountRequest `json:"unbalanced_accounts"`
		UnbalancedAccountDeleteIDs uuid.UUIDs                 `json:"unbalanced_account_delete_ids,omitempty"`
	}

	BranchSettingResponse struct {
		ID         uuid.UUID         `json:"id"`
		CreatedAt  string            `json:"created_at"`
		UpdatedAt  string            `json:"updated_at"`
		BranchID   uuid.UUID         `json:"branch_id"`
		CurrencyID uuid.UUID         `json:"currency_id"`
		Currency   *CurrencyResponse `json:"currency,omitempty"`

		WithdrawAllowUserInput bool   `json:"withdraw_allow_user_input"`
		WithdrawPrefix         string `json:"withdraw_prefix"`
		WithdrawORStart        int    `json:"withdraw_or_start"`
		WithdrawORCurrent      int    `json:"withdraw_or_current"`
		WithdrawOREnd          int    `json:"withdraw_or_end"`
		WithdrawORIteration    int    `json:"withdraw_or_iteration"`
		WithdrawUseDateOR      bool   `json:"withdraw_use_date_or"`
		WithdrawPadding        int    `json:"withdraw_padding"`
		WithdrawCommonOR       string `json:"withdraw_common_or"`

		DepositORStart     int    `json:"deposit_or_start"`
		DepositORCurrent   int    `json:"deposit_or_current"`
		DepositOREnd       int    `json:"deposit_or_end"`
		DepositORIteration int    `json:"deposit_or_iteration"`
		DepositUseDateOR   bool   `json:"deposit_use_date_or"`
		DepositPadding     int    `json:"deposit_padding"`
		DepositCommonOR    string `json:"deposit_common_or"`

		CashCheckVoucherAllowUserInput bool   `json:"cash_check_voucher_allow_user_input"`
		CashCheckVoucherORUnique       bool   `json:"cash_check_voucher_or_unique"`
		CashCheckVoucherPrefix         string `json:"cash_check_voucher_prefix"`
		CashCheckVoucherORStart        int    `json:"cash_check_voucher_or_start"`
		CashCheckVoucherORCurrent      int    `json:"cash_check_voucher_or_current"`
		CashCheckVoucherORIteration    int    `json:"cash_check_voucher_or_iteration"`
		CashCheckVoucherPadding        int    `json:"cash_check_voucher_padding"`

		JournalVoucherAllowUserInput bool   `json:"journal_voucher_allow_user_input"`
		JournalVoucherORUnique       bool   `json:"journal_voucher_or_unique"`
		JournalVoucherPrefix         string `json:"journal_voucher_prefix"`
		JournalVoucherORStart        int    `json:"journal_voucher_or_start"`
		JournalVoucherORCurrent      int    `json:"journal_voucher_or_current"`
		JournalVoucherORIteration    int    `json:"journal_voucher_or_iteration"`
		JournalVoucherPadding        int    `json:"journal_voucher_padding"`

		AdjustmentVoucherAllowUserInput bool   `json:"adjustment_voucher_allow_user_input"`
		AdjustmentVoucherORUnique       bool   `json:"adjustment_voucher_or_unique"`
		AdjustmentVoucherPrefix         string `json:"adjustment_voucher_prefix"`
		AdjustmentVoucherORStart        int    `json:"adjustment_voucher_or_start"`
		AdjustmentVoucherORCurrent      int    `json:"adjustment_voucher_or_current"`
		AdjustmentVoucherORIteration    int    `json:"adjustment_voucher_or_iteration"`
		AdjustmentVoucherPadding        int    `json:"adjustment_voucher_padding"`

		LoanVoucherAllowUserInput bool   `json:"loan_voucher_allow_user_input"`
		LoanVoucherORUnique       bool   `json:"loan_voucher_or_unique"`
		LoanVoucherPrefix         string `json:"loan_voucher_prefix"`
		LoanVoucherORStart        int    `json:"loan_voucher_or_start"`
		LoanVoucherORCurrent      int    `json:"loan_voucher_or_current"`
		LoanVoucherORIteration    int    `json:"loan_voucher_or_iteration"`
		LoanVoucherPadding        int    `json:"loan_voucher_padding"`

		CheckVoucherGeneral               bool   `json:"check_voucher_general"`
		CheckVoucherGeneralAllowUserInput bool   `json:"check_voucher_general_allow_user_input"`
		CheckVoucherGeneralORUnique       bool   `json:"check_voucher_general_or_unique"`
		CheckVoucherGeneralPrefix         string `json:"check_voucher_general_prefix"`
		CheckVoucherGeneralORStart        int    `json:"check_voucher_general_or_start"`
		CheckVoucherGeneralORCurrent      int    `json:"check_voucher_general_or_current"`
		CheckVoucherGeneralORIteration    int    `json:"check_voucher_general_or_iteration"`
		CheckVoucherGeneralPadding        int    `json:"check_voucher_general_padding"`

		DefaultMemberGenderID *uuid.UUID            `json:"default_member_gender_id"`
		DefaultMemberGender   *MemberGenderResponse `json:"default_member_gender"`

		DefaultMemberTypeID *uuid.UUID          `json:"default_member_type_id,omitempty"`
		DefaultMemberType   *MemberTypeResponse `json:"default_member_type,omitempty"`

		CashOnHandAccountID          *uuid.UUID       `json:"cash_on_hand_account_id,omitempty"`
		CashOnHandAccount            *AccountResponse `json:"cash_on_hand_account,omitempty"`
		PaidUpSharedCapitalAccountID *uuid.UUID       `json:"paid_up_shared_capital_account_id,omitempty"`
		PaidUpSharedCapitalAccount   *AccountResponse `json:"paid_up_shared_capital_account,omitempty"`
		CompassionFundAccountID      *uuid.UUID       `json:"compassion_fund_account_id,omitempty"`
		CompassionFundAccount        *AccountResponse `json:"compassion_fund_account,omitempty"`
		AnnualDivisor                int              `json:"annual_divisor"`

		UnbalancedAccounts []*UnbalancedAccountResponse `json:"unbalanced_accounts,omitempty"`
		TaxInterest        float64                      `json:"tax_interest"`
	}
)

func BranchSettingManager(service *horizon.HorizonService) *registry.Registry[BranchSetting, BranchSettingResponse, BranchSettingRequest] {
	return registry.NewRegistry(registry.RegistryParams[BranchSetting, BranchSettingResponse, BranchSettingRequest]{
		Preloads: []string{
			"Branch",
			"Currency",
			"DefaultMemberType",
			"CashOnHandAccount",
			"PaidUpSharedCapitalAccount",
			"CompassionFundAccount",
			"CompassionFundAccount.Currency",
			"UnbalancedAccounts",
			"UnbalancedAccounts.Currency",

			"UnbalancedAccounts.AccountForShortage",
			"UnbalancedAccounts.AccountForOverage",
			"UnbalancedAccounts.MemberProfileForShortage",
			"UnbalancedAccounts.MemberProfileForOverage",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *BranchSetting) *BranchSettingResponse {
			if data == nil {
				return nil
			}
			return &BranchSettingResponse{
				ID:         data.ID,
				CreatedAt:  data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:  data.UpdatedAt.Format(time.RFC3339),
				BranchID:   data.BranchID,
				CurrencyID: data.CurrencyID,
				Currency:   CurrencyManager(service).ToModel(data.Currency),

				WithdrawAllowUserInput: data.WithdrawAllowUserInput,
				WithdrawPrefix:         data.WithdrawPrefix,
				WithdrawORStart:        data.WithdrawORStart,
				WithdrawORCurrent:      data.WithdrawORCurrent,
				WithdrawOREnd:          data.WithdrawOREnd,
				WithdrawORIteration:    data.WithdrawORIteration,
				WithdrawUseDateOR:      data.WithdrawUseDateOR,
				WithdrawPadding:        data.WithdrawPadding,
				WithdrawCommonOR:       data.WithdrawCommonOR,

				DepositORStart:     data.DepositORStart,
				DepositORCurrent:   data.DepositORCurrent,
				DepositOREnd:       data.DepositOREnd,
				DepositORIteration: data.DepositORIteration,
				DepositUseDateOR:   data.DepositUseDateOR,
				DepositPadding:     data.DepositPadding,
				DepositCommonOR:    data.DepositCommonOR,

				CashCheckVoucherAllowUserInput: data.CashCheckVoucherAllowUserInput,
				CashCheckVoucherORUnique:       data.CashCheckVoucherORUnique,
				CashCheckVoucherPrefix:         data.CashCheckVoucherPrefix,
				CashCheckVoucherORStart:        data.CashCheckVoucherORStart,
				CashCheckVoucherORCurrent:      data.CashCheckVoucherORCurrent,
				CashCheckVoucherORIteration:    data.CashCheckVoucherORIteration,
				CashCheckVoucherPadding:        data.CashCheckVoucherPadding,

				JournalVoucherAllowUserInput: data.JournalVoucherAllowUserInput,
				JournalVoucherORUnique:       data.JournalVoucherORUnique,
				JournalVoucherPrefix:         data.JournalVoucherPrefix,
				JournalVoucherORStart:        data.JournalVoucherORStart,
				JournalVoucherORCurrent:      data.JournalVoucherORCurrent,
				JournalVoucherORIteration:    data.JournalVoucherORIteration,
				JournalVoucherPadding:        data.JournalVoucherPadding,

				AdjustmentVoucherAllowUserInput: data.AdjustmentVoucherAllowUserInput,
				AdjustmentVoucherORUnique:       data.AdjustmentVoucherORUnique,
				AdjustmentVoucherPrefix:         data.AdjustmentVoucherPrefix,
				AdjustmentVoucherORStart:        data.AdjustmentVoucherORStart,
				AdjustmentVoucherORCurrent:      data.AdjustmentVoucherORCurrent,
				AdjustmentVoucherORIteration:    data.AdjustmentVoucherORIteration,
				AdjustmentVoucherPadding:        data.AdjustmentVoucherPadding,

				LoanVoucherAllowUserInput: data.LoanVoucherAllowUserInput,
				LoanVoucherORUnique:       data.LoanVoucherORUnique,
				LoanVoucherPrefix:         data.LoanVoucherPrefix,
				LoanVoucherORStart:        data.LoanVoucherORStart,
				LoanVoucherORCurrent:      data.LoanVoucherORCurrent,
				LoanVoucherORIteration:    data.LoanVoucherORIteration,
				LoanVoucherPadding:        data.LoanVoucherPadding,

				CheckVoucherGeneral:               data.CheckVoucherGeneral,
				CheckVoucherGeneralAllowUserInput: data.CheckVoucherGeneralAllowUserInput,
				CheckVoucherGeneralORUnique:       data.CheckVoucherGeneralORUnique,
				CheckVoucherGeneralPrefix:         data.CheckVoucherGeneralPrefix,
				CheckVoucherGeneralORStart:        data.CheckVoucherGeneralORStart,
				CheckVoucherGeneralORCurrent:      data.CheckVoucherGeneralORCurrent,
				CheckVoucherGeneralORIteration:    data.CheckVoucherGeneralORIteration,
				CheckVoucherGeneralPadding:        data.CheckVoucherGeneralPadding,

				DefaultMemberTypeID: data.DefaultMemberTypeID,
				DefaultMemberType:   MemberTypeManager(service).ToModel(data.DefaultMemberType),

				CashOnHandAccountID:          data.CashOnHandAccountID,
				CashOnHandAccount:            AccountManager(service).ToModel(data.CashOnHandAccount),
				PaidUpSharedCapitalAccountID: data.PaidUpSharedCapitalAccountID,
				PaidUpSharedCapitalAccount:   AccountManager(service).ToModel(data.PaidUpSharedCapitalAccount),
				CompassionFundAccountID:      data.CompassionFundAccountID,
				CompassionFundAccount:        AccountManager(service).ToModel(data.CompassionFundAccount),

				UnbalancedAccounts: UnbalancedAccountManager(service).ToModels(data.UnbalancedAccounts),
				AnnualDivisor:      data.AnnualDivisor,
				TaxInterest:        data.TaxInterest,

				DefaultMemberGenderID: data.DefaultMemberGenderID,
				DefaultMemberGender:   MemberGenderManager(service).ToModel(data.DefaultMemberGender),
			}
		},
		Created: func(data *BranchSetting) registry.Topics {
			return []string{
				"branch_setting.create",
				fmt.Sprintf("branch_setting.create.%s", data.ID),
				fmt.Sprintf("branch_setting.create.branch.%s", data.BranchID),
			}
		},
		Updated: func(data *BranchSetting) registry.Topics {
			return []string{
				"branch_setting.update",
				fmt.Sprintf("branch_setting.update.%s", data.ID),
				fmt.Sprintf("branch_setting.update.branch.%s", data.BranchID),
			}
		},
		Deleted: func(data *BranchSetting) registry.Topics {
			return []string{
				"branch_setting.delete",
				fmt.Sprintf("branch_setting.delete.%s", data.ID),
				fmt.Sprintf("branch_setting.delete.branch.%s", data.BranchID),
			}
		},
	})
}
