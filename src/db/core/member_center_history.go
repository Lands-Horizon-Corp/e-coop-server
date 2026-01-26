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

func MemberCenterHistoryManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberCenterHistory, types.MemberCenterHistoryResponse, types.MemberCenterHistoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberCenterHistory, types.MemberCenterHistoryResponse, types.MemberCenterHistoryRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberCenter", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberCenterHistory) *types.MemberCenterHistoryResponse {
			if data == nil {
				return nil
			}
			return &types.MemberCenterHistoryResponse{
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
				MemberCenterID:  data.MemberCenterID,
				MemberCenter:    MemberCenterManager(service).ToModel(data.MemberCenter),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   MemberProfileManager(service).ToModel(data.MemberProfile),
			}
		},

		Created: func(data *types.MemberCenterHistory) registry.Topics {
			return []string{
				"member_center_history.create",
				fmt.Sprintf("member_center_history.create.%s", data.ID),
				fmt.Sprintf("member_center_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_center_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_center_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *types.MemberCenterHistory) registry.Topics {
			return []string{
				"member_center_history.update",
				fmt.Sprintf("member_center_history.update.%s", data.ID),
				fmt.Sprintf("member_center_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_center_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_center_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *types.MemberCenterHistory) registry.Topics {
			return []string{
				"member_center_history.delete",
				fmt.Sprintf("member_center_history.delete.%s", data.ID),
				fmt.Sprintf("member_center_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_center_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_center_history.delete.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func MemberCenterHistoryCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberCenterHistory, error) {
	return MemberCenterHistoryManager(service).Find(context, &types.MemberCenterHistory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func MemberCenterHistoryMemberProfileID(context context.Context, service *horizon.HorizonService, memberProfileID, organizationID, branchID uuid.UUID) ([]*types.MemberCenterHistory, error) {
	return MemberCenterHistoryManager(service).Find(context, &types.MemberCenterHistory{
		OrganizationID:  organizationID,
		BranchID:        branchID,
		MemberProfileID: memberProfileID,
	})
}
