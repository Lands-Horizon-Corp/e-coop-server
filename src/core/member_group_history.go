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

func MemberGroupHistoryManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberGroupHistory, types.MemberGroupHistoryResponse, types.MemberGroupHistoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberGroupHistory, types.MemberGroupHistoryResponse, types.MemberGroupHistoryRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile", "MemberGroup"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberGroupHistory) *types.MemberGroupHistoryResponse {
			if data == nil {
				return nil
			}
			return &types.MemberGroupHistoryResponse{
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
				MemberGroupID:   data.MemberGroupID,
				MemberGroup:     MemberGroupManager(service).ToModel(data.MemberGroup),
			}
		},

		Created: func(data *types.MemberGroupHistory) registry.Topics {
			return []string{
				"member_group_history.create",
				fmt.Sprintf("member_group_history.create.%s", data.ID),
				fmt.Sprintf("member_group_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_group_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_group_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *types.MemberGroupHistory) registry.Topics {
			return []string{
				"member_group_history.update",
				fmt.Sprintf("member_group_history.update.%s", data.ID),
				fmt.Sprintf("member_group_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_group_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_group_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *types.MemberGroupHistory) registry.Topics {
			return []string{
				"member_group_history.delete",
				fmt.Sprintf("member_group_history.delete.%s", data.ID),
				fmt.Sprintf("member_group_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_group_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_group_history.delete.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func MemberGroupHistoryCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberGroupHistory, error) {
	return MemberGroupHistoryManager(service).Find(context, &types.MemberGroupHistory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func MemberGroupHistoryMemberProfileID(context context.Context, service *horizon.HorizonService,
	memberProfileID, organizationID, branchID uuid.UUID) ([]*types.MemberGroupHistory, error) {
	return MemberGroupHistoryManager(service).Find(context, &types.MemberGroupHistory{
		OrganizationID:  organizationID,
		BranchID:        branchID,
		MemberProfileID: memberProfileID,
	})
}
