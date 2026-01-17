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

func MutualFundEntryManager(service *horizon.HorizonService) *registry.Registry[
	types.MutualFundEntry, types.MutualFundEntryResponse, types.MutualFundEntryRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MutualFundEntry, types.MutualFundEntryResponse, types.MutualFundEntryRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
			"MemberProfile", "Account", "Account.Currency", "MutualFund"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MutualFundEntry) *types.MutualFundEntryResponse {
			if data == nil {
				return nil
			}
			return &types.MutualFundEntryResponse{
				ID:              data.ID,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:     data.CreatedByID,
				CreatedBy:       UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:     data.UpdatedByID,
				UpdatedBy:       UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:  data.OrganizationID,
				Organization:    OrganizationManager(service).ToModel(data.Organization),
				BranchID:        data.BranchID,
				Branch:          BranchManager(service).ToModel(data.Branch),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   MemberProfileManager(service).ToModel(data.MemberProfile),
				AccountID:       data.AccountID,
				Account:         AccountManager(service).ToModel(data.Account),
				Amount:          data.Amount,
				MutualFundID:    data.MutualFundID,
				MutualFund:      MutualFundManager(service).ToModel(data.MutualFund),
			}
		},
		Created: func(data *types.MutualFundEntry) registry.Topics {
			return []string{
				"mutual_fund_entry.create",
				fmt.Sprintf("mutual_fund_entry.create.%s", data.ID),
				fmt.Sprintf("mutual_fund_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_entry.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_entry.create.member.%s", data.MemberProfileID),
				fmt.Sprintf("mutual_fund_entry.create.account.%s", data.AccountID),
			}
		},
		Updated: func(data *types.MutualFundEntry) registry.Topics {
			return []string{
				"mutual_fund_entry.update",
				fmt.Sprintf("mutual_fund_entry.update.%s", data.ID),
				fmt.Sprintf("mutual_fund_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_entry.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_entry.update.member.%s", data.MemberProfileID),
				fmt.Sprintf("mutual_fund_entry.update.account.%s", data.AccountID),
			}
		},
		Deleted: func(data *types.MutualFundEntry) registry.Topics {
			return []string{
				"mutual_fund_entry.delete",
				fmt.Sprintf("mutual_fund_entry.delete.%s", data.ID),
				fmt.Sprintf("mutual_fund_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund_entry.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund_entry.delete.member.%s", data.MemberProfileID),
				fmt.Sprintf("mutual_fund_entry.delete.account.%s", data.AccountID),
			}
		},
	})
}

func MutualFundEntryCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MutualFundEntry, error) {
	return MutualFundEntryManager(service).Find(context, &types.MutualFundEntry{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func MutualFundEntryByMember(context context.Context, service *horizon.HorizonService,
	memberProfileID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MutualFundEntry, error) {
	return MutualFundEntryManager(service).Find(context, &types.MutualFundEntry{
		MemberProfileID: memberProfileID,
		OrganizationID:  organizationID,
		BranchID:        branchID,
	})
}

func MutualFundEntryByAccount(context context.Context, service *horizon.HorizonService,
	accountID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MutualFundEntry, error) {
	return MutualFundEntryManager(service).Find(context, &types.MutualFundEntry{
		AccountID:      accountID,
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
