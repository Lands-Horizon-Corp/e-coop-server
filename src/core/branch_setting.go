package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

func BranchSettingManager(service *horizon.HorizonService) *registry.Registry[
	types.BranchSetting, types.BranchSettingResponse, types.BranchSettingRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.BranchSetting, types.BranchSettingResponse, types.BranchSettingRequest]{
		Preloads: []string{
			"Branch",
			"Currency",
			"DefaultMemberType",
			"DefaultMemberGender",
			"CashOnHandAccount",
			"PaidUpSharedCapitalAccount",
			"CompassionFundAccount",
			"CompassionFundAccount.Currency",
			"AccountWallet",
			"AccountWallet.Currency",
			"UnbalancedAccounts",
			"UnbalancedAccounts.Currency",
			"UnbalancedAccounts.AccountForShortage",
			"UnbalancedAccounts.AccountForOverage",
			"UnbalancedAccounts.MemberProfileForOverage",
			"UnbalancedAccounts.MemberProfileForShortage",
			"UnbalancedAccounts.CashOnHandAccount",
		},

		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.BranchSetting) *types.BranchSettingResponse {
			if data == nil {
				return nil
			}
			return &types.BranchSettingResponse{
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
				CashCheckVoucherPadding:        data.CashCheckVoucherPadding,

				JournalVoucherAllowUserInput: data.JournalVoucherAllowUserInput,
				JournalVoucherORUnique:       data.JournalVoucherORUnique,
				JournalVoucherPrefix:         data.JournalVoucherPrefix,
				JournalVoucherORStart:        data.JournalVoucherORStart,
				JournalVoucherORCurrent:      data.JournalVoucherORCurrent,
				JournalVoucherPadding:        data.JournalVoucherPadding,

				AdjustmentVoucherAllowUserInput: data.AdjustmentVoucherAllowUserInput,
				AdjustmentVoucherORUnique:       data.AdjustmentVoucherORUnique,
				AdjustmentVoucherPrefix:         data.AdjustmentVoucherPrefix,
				AdjustmentVoucherORStart:        data.AdjustmentVoucherORStart,
				AdjustmentVoucherORCurrent:      data.AdjustmentVoucherORCurrent,
				AdjustmentVoucherPadding:        data.AdjustmentVoucherPadding,

				LoanVoucherAllowUserInput: data.LoanVoucherAllowUserInput,
				LoanVoucherORUnique:       data.LoanVoucherORUnique,
				LoanVoucherPrefix:         data.LoanVoucherPrefix,
				LoanVoucherORStart:        data.LoanVoucherORStart,
				LoanVoucherORCurrent:      data.LoanVoucherORCurrent,
				LoanVoucherPadding:        data.LoanVoucherPadding,

				CheckVoucherGeneral:               data.CheckVoucherGeneral,
				CheckVoucherGeneralAllowUserInput: data.CheckVoucherGeneralAllowUserInput,
				CheckVoucherGeneralORUnique:       data.CheckVoucherGeneralORUnique,
				CheckVoucherGeneralPrefix:         data.CheckVoucherGeneralPrefix,
				CheckVoucherGeneralORStart:        data.CheckVoucherGeneralORStart,
				CheckVoucherGeneralORCurrent:      data.CheckVoucherGeneralORCurrent,
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

				LoanAppliedEqualToBalance: data.LoanAppliedEqualToBalance,
				AccountWalletID:           data.AccountWalletID,
				AccountWallet:             AccountManager(service).ToModel(data.AccountWallet),
			}
		},
		Created: func(data *types.BranchSetting) registry.Topics {
			return []string{
				"branch_setting.create",
				fmt.Sprintf("branch_setting.create.%s", data.ID),
				fmt.Sprintf("branch_setting.create.branch.%s", data.BranchID),
			}
		},
		Updated: func(data *types.BranchSetting) registry.Topics {
			return []string{
				"branch_setting.update",
				fmt.Sprintf("branch_setting.update.%s", data.ID),
				fmt.Sprintf("branch_setting.update.branch.%s", data.BranchID),
			}
		},
		Deleted: func(data *types.BranchSetting) registry.Topics {
			return []string{
				"branch_setting.delete",
				fmt.Sprintf("branch_setting.delete.%s", data.ID),
				fmt.Sprintf("branch_setting.delete.branch.%s", data.BranchID),
			}
		},
	})
}
