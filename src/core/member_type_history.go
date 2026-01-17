package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

func MemberTypeHistoryManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberTypeHistory, types.MemberTypeHistoryResponse, types.MemberTypeHistoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberTypeHistory, types.MemberTypeHistoryResponse, types.MemberTypeHistoryRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberType", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberTypeHistory) *types.MemberTypeHistoryResponse {
			if data == nil {
				return nil
			}
			return &types.MemberTypeHistoryResponse{
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
				MemberTypeID:    data.MemberTypeID,
				MemberType:      MemberTypeManager(service).ToModel(data.MemberType),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   MemberProfileManager(service).ToModel(data.MemberProfile),
			}
		},

		Created: func(data *types.MemberTypeHistory) registry.Topics {
			return []string{
				"member_type_history.create",
				fmt.Sprintf("member_type_history.create.%s", data.ID),
				fmt.Sprintf("member_type_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_type_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *types.MemberTypeHistory) registry.Topics {
			return []string{
				"member_type_history.update",
				fmt.Sprintf("member_type_history.update.%s", data.ID),
				fmt.Sprintf("member_type_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_type_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *types.MemberTypeHistory) registry.Topics {
			return []string{
				"member_type_history.delete",
				fmt.Sprintf("member_type_history.delete.%s", data.ID),
				fmt.Sprintf("member_type_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_type_history.delete.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func MemberTypeHistoryCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberTypeHistory, error) {
	return MemberTypeHistoryManager(service).Find(context, &types.MemberTypeHistory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func MemberTypeHistoryMemberProfileID(context context.Context, service *horizon.HorizonService, memberProfileID, organizationID,
	branchID uuid.UUID) ([]*types.MemberTypeHistory, error) {
	return MemberTypeHistoryManager(service).Find(context, &types.MemberTypeHistory{
		OrganizationID:  organizationID,
		BranchID:        branchID,
		MemberProfileID: memberProfileID,
	})
}

func GetMemberTypeHistoryLatest(
	context context.Context, service *horizon.HorizonService,
	memberProfileID, memberTypeID, organizationID, branchID uuid.UUID,
) (*types.MemberTypeHistory, error) {
	filters := []query.ArrFilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "member_type_id", Op: query.ModeEqual, Value: memberTypeID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	sorts := []query.ArrFilterSortSQL{
		{Field: "created_at", Order: "DESC"},
	}

	return MemberTypeHistoryManager(service).ArrFindOne(context, filters, sorts)
}
