package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

func UserOrganizationManager(service *horizon.HorizonService) *registry.Registry[types.UserOrganization, types.UserOrganizationResponse, types.UserOrganizationRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.UserOrganization, types.UserOrganizationResponse, types.UserOrganizationRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Branch",
			"Branch.Media",

			"User",
			"User.Media",
			"Organization",
			"Organization.Media",
			"Organization.CoverMedia",
			"Organization.OrganizationCategories",
			"Organization.OrganizationCategories.Category",

			"SettingsAccountingPaymentDefaultValue",
			"SettingsAccountingDepositDefaultValue",
			"SettingsAccountingWithdrawDefaultValue",
			"SettingsPaymentTypeDefaultValue",

			"Branch.BranchSetting",
			"Branch.BranchSetting.Currency",

			"Branch.BranchSetting.CashOnHandAccount",
			"Branch.BranchSetting.CashOnHandAccount.Currency",
			"Branch.BranchSetting.PaidUpSharedCapitalAccount",
			"Branch.BranchSetting.PaidUpSharedCapitalAccount.Currency",
			"Branch.BranchSetting.DefaultMemberGender",
			"Branch.BranchSetting.DefaultMemberType",

			"Branch.BranchSetting.CompassionFundAccount",
			"Branch.BranchSetting.CompassionFundAccount.Currency",

			"Branch.BranchSetting.UnbalancedAccounts.Currency",
			"Branch.BranchSetting.UnbalancedAccounts.AccountForShortage",
			"Branch.BranchSetting.UnbalancedAccounts.AccountForOverage",
			"Branch.BranchSetting.UnbalancedAccounts.CashOnHandAccount",
			"Branch.BranchSetting.UnbalancedAccounts.MemberProfileForShortage",
			"Branch.BranchSetting.UnbalancedAccounts.MemberProfileForOverage",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.UserOrganization) *types.UserOrganizationResponse {
			if data == nil {
				return nil
			}
			if data.Permissions == nil {
				data.Permissions = []string{}
			}
			return &types.UserOrganizationResponse{
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

				UserID:                 data.UserID,
				User:                   UserManager(service).ToModel(data.User),
				UserType:               data.UserType,
				Description:            data.Description,
				ApplicationDescription: data.ApplicationDescription,
				ApplicationStatus:      data.ApplicationStatus,
				DeveloperSecretKey:     "",
				PermissionName:         data.PermissionName,
				PermissionDescription:  data.PermissionDescription,
				Permissions:            data.Permissions,

				UserSettingDescription: data.UserSettingDescription,

				PaymentORUnique:                      data.PaymentORUnique,
				PaymentORAllowUserInput:              data.PaymentORAllowUserInput,
				PaymentORCurrent:                     data.PaymentORCurrent,
				PaymentORStart:                       data.PaymentORStart,
				PaymentOREnd:                         data.PaymentOREnd,
				PaymentORIteration:                   data.PaymentORIteration,
				PaymentORUseDateOR:                   data.PaymentORUseDateOR,
				PaymentPrefix:                        data.PaymentPrefix,
				PaymentPadding:                       data.PaymentPadding,
				SettingsAllowWithdrawNegativeBalance: data.SettingsAllowWithdrawNegativeBalance,
				SettingsAllowWithdrawExactBalance:    data.SettingsAllowWithdrawExactBalance,
				SettingsMaintainingBalance:           data.SettingsMaintainingBalance,
				Status:                               data.Status,
				LastOnlineAt:                         data.LastOnlineAt,
				TimeMachineTime:                      data.TimeMachineTime,

				SettingsAccountingPaymentDefaultValueID:  data.SettingsAccountingPaymentDefaultValueID,
				SettingsAccountingPaymentDefaultValue:    AccountManager(service).ToModel(data.SettingsAccountingPaymentDefaultValue),
				SettingsAccountingDepositDefaultValueID:  data.SettingsAccountingDepositDefaultValueID,
				SettingsAccountingDepositDefaultValue:    AccountManager(service).ToModel(data.SettingsAccountingDepositDefaultValue),
				SettingsAccountingWithdrawDefaultValueID: data.SettingsAccountingWithdrawDefaultValueID,
				SettingsAccountingWithdrawDefaultValue:   AccountManager(service).ToModel(data.SettingsAccountingWithdrawDefaultValue),
				SettingsPaymentTypeDefaultValueID:        data.SettingsPaymentTypeDefaultValueID,
				SettingsPaymentTypeDefaultValue:          PaymentTypeManager(service).ToModel(data.SettingsPaymentTypeDefaultValue),
			}
		},
		Created: func(data *types.UserOrganization) registry.Topics {
			return []string{
				"user_organization.create",
				fmt.Sprintf("user_organization.create.%s", data.ID),
				fmt.Sprintf("user_organization.create.branch.%s", data.BranchID),
				fmt.Sprintf("user_organization.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("user_organization.create.user.%s", data.UserID),
			}
		},
		Updated: func(data *types.UserOrganization) registry.Topics {
			return []string{
				"user_organization.update",
				fmt.Sprintf("user_organization.update.%s", data.ID),
				fmt.Sprintf("user_organization.update.branch.%s", data.BranchID),
				fmt.Sprintf("user_organization.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("user_organization.update.user.%s", data.UserID),
			}
		},
		Deleted: func(data *types.UserOrganization) registry.Topics {
			return []string{
				"user_organization.delete",
				fmt.Sprintf("user_organization.delete.%s", data.ID),
				fmt.Sprintf("user_organization.delete.branch.%s", data.BranchID),
				fmt.Sprintf("user_organization.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("user_organization.delete.user.%s", data.UserID),
			}
		},
	})
}

func GetUserOrganizationByUser(context context.Context, service *horizon.HorizonService, userID uuid.UUID, pending *bool) ([]*types.UserOrganization, error) {
	filter := &types.UserOrganization{
		UserID: userID,
	}
	if pending != nil && *pending {
		filter.ApplicationStatus = "pending"
	}
	return UserOrganizationManager(service).Find(context, filter)
}

func GetUserOrganizationByOrganization(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, pending *bool) ([]*types.UserOrganization, error) {
	filter := &types.UserOrganization{
		OrganizationID: organizationID,
	}
	if pending != nil && *pending {
		filter.ApplicationStatus = "pending"
	}
	return UserOrganizationManager(service).Find(context, filter)
}

func GetUserOrganizationByBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID, pending *bool) ([]*types.UserOrganization, error) {
	filter := &types.UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
	}
	if pending != nil && *pending {
		filter.ApplicationStatus = "pending"
	}
	return UserOrganizationManager(service).Find(context, filter)
}

func CountUserOrganizationPerBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) (int64, error) {
	return UserOrganizationManager(service).Count(context, &types.UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
	})
}

func CountUserOrganizationbranch(context context.Context, service *horizon.HorizonService, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) (int64, error) {
	return UserOrganizationManager(service).Count(context, &types.UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserID:         userID,
	})
}

func UserOrganizationEmployeeCanJoin(context context.Context, service *horizon.HorizonService, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) bool {
	existing, err := CountUserOrganizationbranch(context, service, userID, organizationID, branchID)
	return err == nil && existing == 0
}

func UserOrganizationMemberCanJoin(context context.Context, service *horizon.HorizonService, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) bool {
	existing, err := CountUserOrganizationbranch(context, service, userID, organizationID, branchID)
	if err != nil || existing > 0 {
		return false
	}
	existingOrgCount, err := UserOrganizationManager(service).Count(context, &types.UserOrganization{
		UserID:         userID,
		OrganizationID: organizationID,
	})
	return err == nil && existingOrgCount == 0
}

func Employees(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.UserOrganization, error) {
	return UserOrganizationManager(service).Find(context, &types.UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserType:       types.UserOrganizationTypeEmployee,
	})
}

func Members(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.UserOrganization, error) {
	return UserOrganizationManager(service).Find(context, &types.UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserType:       types.UserOrganizationTypeMember,
	})
}
