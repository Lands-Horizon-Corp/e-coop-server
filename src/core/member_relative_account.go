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

func MemberRelativeAccountManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberRelativeAccount, types.MemberRelativeAccountResponse, types.MemberRelativeAccountRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberRelativeAccount, types.MemberRelativeAccountResponse, types.MemberRelativeAccountRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile", "RelativeMemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberRelativeAccount) *types.MemberRelativeAccountResponse {
			if data == nil {
				return nil
			}
			return &types.MemberRelativeAccountResponse{
				ID:                      data.ID,
				CreatedAt:               data.CreatedAt.Format(time.RFC3339),
				CreatedByID:             data.CreatedByID,
				CreatedBy:               UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:               data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:             data.UpdatedByID,
				UpdatedBy:               UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:          data.OrganizationID,
				Organization:            OrganizationManager(service).ToModel(data.Organization),
				BranchID:                data.BranchID,
				Branch:                  BranchManager(service).ToModel(data.Branch),
				MemberProfileID:         data.MemberProfileID,
				MemberProfile:           MemberProfileManager(service).ToModel(data.MemberProfile),
				RelativeMemberProfileID: data.RelativeMemberProfileID,
				RelativeMemberProfile:   MemberProfileManager(service).ToModel(data.RelativeMemberProfile),
				FamilyRelationship:      data.FamilyRelationship,
				Description:             data.Description,
			}
		},

		Created: func(data *types.MemberRelativeAccount) registry.Topics {
			return []string{
				"member_relative_account.create",
				fmt.Sprintf("member_relative_account.create.%s", data.ID),
				fmt.Sprintf("member_relative_account.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_relative_account.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberRelativeAccount) registry.Topics {
			return []string{
				"member_relative_account.update",
				fmt.Sprintf("member_relative_account.update.%s", data.ID),
				fmt.Sprintf("member_relative_account.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_relative_account.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberRelativeAccount) registry.Topics {
			return []string{
				"member_relative_account.delete",
				fmt.Sprintf("member_relative_account.delete.%s", data.ID),
				fmt.Sprintf("member_relative_account.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_relative_account.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberRelativeAccountCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberRelativeAccount, error) {
	return MemberRelativeAccountManager(service).Find(context, &types.MemberRelativeAccount{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
