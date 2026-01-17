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

func CollectorsMemberAccountEntryManager(service *horizon.HorizonService) *registry.Registry[
	types.CollectorsMemberAccountEntry, types.CollectorsMemberAccountEntryResponse, types.CollectorsMemberAccountEntryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.CollectorsMemberAccountEntry, types.CollectorsMemberAccountEntryResponse, types.CollectorsMemberAccountEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"CollectorUser", "MemberProfile", "Account",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.CollectorsMemberAccountEntry) *types.CollectorsMemberAccountEntryResponse {
			if data == nil {
				return nil
			}
			return &types.CollectorsMemberAccountEntryResponse{
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
				CollectorUserID: data.CollectorUserID,
				CollectorUser:   UserManager(service).ToModel(data.CollectorUser),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   MemberProfileManager(service).ToModel(data.MemberProfile),
				AccountID:       data.AccountID,
				Account:         AccountManager(service).ToModel(data.Account),
				Description:     data.Description,
			}
		},
		Created: func(data *types.CollectorsMemberAccountEntry) registry.Topics {
			return []string{
				"collectors_member_account_entry.create",
				fmt.Sprintf("collectors_member_account_entry.create.%s", data.ID),
				fmt.Sprintf("collectors_member_account_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("collectors_member_account_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.CollectorsMemberAccountEntry) registry.Topics {
			return []string{
				"collectors_member_account_entry.update",
				fmt.Sprintf("collectors_member_account_entry.update.%s", data.ID),
				fmt.Sprintf("collectors_member_account_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("collectors_member_account_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.CollectorsMemberAccountEntry) registry.Topics {
			return []string{
				"collectors_member_account_entry.delete",
				fmt.Sprintf("collectors_member_account_entry.delete.%s", data.ID),
				fmt.Sprintf("collectors_member_account_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("collectors_member_account_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func CollectorsMemberAccountEntryCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.CollectorsMemberAccountEntry, error) {
	return CollectorsMemberAccountEntryManager(service).Find(context, &types.CollectorsMemberAccountEntry{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
