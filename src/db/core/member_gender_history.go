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

func MemberGenderHistoryManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberGenderHistory, types.MemberGenderHistoryResponse, types.MemberGenderHistoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberGenderHistory, types.MemberGenderHistoryResponse, types.MemberGenderHistoryRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile", "MemberGender"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberGenderHistory) *types.MemberGenderHistoryResponse {
			if data == nil {
				return nil
			}
			return &types.MemberGenderHistoryResponse{
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
				MemberGenderID:  data.MemberGenderID,
				MemberGender:    MemberGenderManager(service).ToModel(data.MemberGender),
			}
		},

		Created: func(data *types.MemberGenderHistory) registry.Topics {
			return []string{
				"member_gender_history.create",
				fmt.Sprintf("member_gender_history.create.%s", data.ID),
				fmt.Sprintf("member_gender_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_gender_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_gender_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *types.MemberGenderHistory) registry.Topics {
			return []string{
				"member_gender_history.update",
				fmt.Sprintf("member_gender_history.update.%s", data.ID),
				fmt.Sprintf("member_gender_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_gender_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_gender_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *types.MemberGenderHistory) registry.Topics {
			return []string{
				"member_gender_history.delete",
				fmt.Sprintf("member_gender_history.delete.%s", data.ID),
				fmt.Sprintf("member_gender_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_gender_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_gender_history.delete.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func MemberGenderHistoryCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberGenderHistory, error) {
	return MemberGenderHistoryManager(service).Find(context, &types.MemberGenderHistory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func MemberGenderHistoryMemberProfileID(context context.Context, service *horizon.HorizonService,
	memberProfileID, organizationID, branchID uuid.UUID) ([]*types.MemberGenderHistory, error) {
	return MemberGenderHistoryManager(service).Find(context, &types.MemberGenderHistory{
		OrganizationID:  organizationID,
		BranchID:        branchID,
		MemberProfileID: memberProfileID,
	})
}
