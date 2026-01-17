package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

func UnbalancedAccountManager(service *horizon.HorizonService) *registry.Registry[
	types.UnbalancedAccount, types.UnbalancedAccountResponse, types.UnbalancedAccountRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.UnbalancedAccount, types.UnbalancedAccountResponse, types.UnbalancedAccountRequest]{
		Preloads: []string{
			"Currency",
			"AccountForShortage", "AccountForOverage",
			"CashOnHandAccount",
			"MemberProfileForShortage", "MemberProfileForOverage",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.UnbalancedAccount) *types.UnbalancedAccountResponse {
			if data == nil {
				return nil
			}
			return &types.UnbalancedAccountResponse{
				ID:               data.ID,
				CreatedAt:        data.CreatedAt.Format(time.RFC3339),
				CreatedByID:      data.CreatedByID,
				CreatedBy:        UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:        data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:      data.UpdatedByID,
				UpdatedBy:        UserManager(service).ToModel(data.UpdatedBy),
				BranchSettingsID: data.BranchSettingsID,
				BranchSettings:   BranchSettingManager(service).ToModel(data.BranchSettings),
				CurrencyID:       data.CurrencyID,
				Currency:         CurrencyManager(service).ToModel(data.Currency),

				AccountForShortageID: data.AccountForShortageID,
				AccountForShortage:   AccountManager(service).ToModel(data.AccountForShortage),

				AccountForOverageID: data.AccountForOverageID,
				AccountForOverage:   AccountManager(service).ToModel(data.AccountForOverage),

				CashOnHandAccountID: data.CashOnHandAccountID,
				CashOnHandAccount:   AccountManager(service).ToModel(data.CashOnHandAccount),

				MemberProfileIDForShortage: data.MemberProfileIDForShortage,
				MemberProfileForShortage:   MemberProfileManager(service).ToModel(data.MemberProfileForShortage),

				MemberProfileIDForOverage: data.MemberProfileIDForOverage,
				MemberProfileForOverage:   MemberProfileManager(service).ToModel(data.MemberProfileForOverage),

				Name:        data.Name,
				Description: data.Description,
			}
		},
		Created: func(data *types.UnbalancedAccount) registry.Topics {
			return []string{
				"unbalanced_account.create",
				fmt.Sprintf("unbalanced_account.create.%s", data.ID),
				fmt.Sprintf("unbalanced_account.create.branch_settings.%s", data.BranchSettingsID),
				fmt.Sprintf("unbalanced_account.create.currency.%s", data.CurrencyID),
			}
		},
		Updated: func(data *types.UnbalancedAccount) registry.Topics {
			return []string{
				"unbalanced_account.update",
				fmt.Sprintf("unbalanced_account.update.%s", data.ID),
				fmt.Sprintf("unbalanced_account.update.branch_settings.%s", data.BranchSettingsID),
				fmt.Sprintf("unbalanced_account.update.currency.%s", data.CurrencyID),
			}
		},
		Deleted: func(data *types.UnbalancedAccount) registry.Topics {
			return []string{
				"unbalanced_account.delete",
				fmt.Sprintf("unbalanced_account.delete.%s", data.ID),
				fmt.Sprintf("unbalanced_account.delete.branch_settings.%s", data.BranchSettingsID),
				fmt.Sprintf("unbalanced_account.delete.currency.%s", data.CurrencyID),
			}
		},
	})
}
