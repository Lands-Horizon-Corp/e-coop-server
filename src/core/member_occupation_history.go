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

func MemberOccupationHistoryManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberOccupationHistory, types.MemberOccupationHistoryResponse, types.MemberOccupationHistoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberOccupationHistory, types.MemberOccupationHistoryResponse, types.MemberOccupationHistoryRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "MemberProfile", "MemberOccupation",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberOccupationHistory) *types.MemberOccupationHistoryResponse {
			if data == nil {
				return nil
			}
			return &types.MemberOccupationHistoryResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       OrganizationManager(service).ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             BranchManager(service).ToModel(data.Branch),
				MemberProfileID:    data.MemberProfileID,
				MemberProfile:      MemberProfileManager(service).ToModel(data.MemberProfile),
				MemberOccupationID: data.MemberOccupationID,
				MemberOccupation:   MemberOccupationManager(service).ToModel(data.MemberOccupation),
			}
		},
		Created: func(data *types.MemberOccupationHistory) registry.Topics {
			return []string{
				"member_occupation_history.create",
				fmt.Sprintf("member_occupation_history.create.%s", data.ID),
				fmt.Sprintf("member_occupation_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_occupation_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_occupation_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *types.MemberOccupationHistory) registry.Topics {
			return []string{
				"member_occupation_history.update",
				fmt.Sprintf("member_occupation_history.update.%s", data.ID),
				fmt.Sprintf("member_occupation_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_occupation_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_occupation_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *types.MemberOccupationHistory) registry.Topics {
			return []string{
				"member_occupation_history.delete",
				fmt.Sprintf("member_occupation_history.delete.%s", data.ID),
				fmt.Sprintf("member_occupation_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_occupation_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_occupation_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func MemberOccupationHistoryCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberOccupationHistory, error) {
	return MemberOccupationHistoryManager(service).Find(context, &types.MemberOccupationHistory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func MemberOccupationHistoryMemberProfileID(context context.Context, service *horizon.HorizonService, memberProfileID,
	organizationID, branchID uuid.UUID) ([]*types.MemberOccupationHistory, error) {
	return MemberOccupationHistoryManager(service).Find(context, &types.MemberOccupationHistory{
		OrganizationID:  organizationID,
		BranchID:        branchID,
		MemberProfileID: memberProfileID,
	})
}
