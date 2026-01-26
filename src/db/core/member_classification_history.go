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

func MemberClassificationHistoryManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberClassificationHistory, types.MemberClassificationHistoryResponse, types.MemberClassificationHistoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberClassificationHistory,
		types.MemberClassificationHistoryResponse,
		types.MemberClassificationHistoryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"Organization", "Branch", "MemberClassification", "MemberProfile",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberClassificationHistory) *types.MemberClassificationHistoryResponse {
			if data == nil {
				return nil
			}
			return &types.MemberClassificationHistoryResponse{
				ID:                     data.ID,
				CreatedAt:              data.CreatedAt.Format(time.RFC3339),
				CreatedByID:            data.CreatedByID,
				CreatedBy:              UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:              data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:            data.UpdatedByID,
				UpdatedBy:              UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:         data.OrganizationID,
				Organization:           OrganizationManager(service).ToModel(data.Organization),
				BranchID:               data.BranchID,
				Branch:                 BranchManager(service).ToModel(data.Branch),
				MemberClassificationID: data.MemberClassificationID,
				MemberClassification:   MemberClassificationManager(service).ToModel(data.MemberClassification),
				MemberProfileID:        data.MemberProfileID,
				MemberProfile:          MemberProfileManager(service).ToModel(data.MemberProfile),
			}
		},
		Created: func(data *types.MemberClassificationHistory) registry.Topics {
			return []string{
				"member_classification_history.create",
				fmt.Sprintf("member_classification_history.create.%s", data.ID),
				fmt.Sprintf("member_classification_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_classification_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *types.MemberClassificationHistory) registry.Topics {
			return []string{
				"member_classification_history.update",
				fmt.Sprintf("member_classification_history.update.%s", data.ID),
				fmt.Sprintf("member_classification_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_classification_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *types.MemberClassificationHistory) registry.Topics {
			return []string{
				"member_classification_history.delete",
				fmt.Sprintf("member_classification_history.delete.%s", data.ID),
				fmt.Sprintf("member_classification_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_classification_history.delete.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func MemberClassificationHistoryCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberClassificationHistory, error) {
	return MemberClassificationHistoryManager(service).Find(context, &types.MemberClassificationHistory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func MemberClassificationHistoryMemberProfileID(context context.Context,
	service *horizon.HorizonService, memberProfileID, organizationID, branchID uuid.UUID) ([]*types.MemberClassificationHistory, error) {
	return MemberClassificationHistoryManager(service).Find(context, &types.MemberClassificationHistory{
		OrganizationID:  organizationID,
		BranchID:        branchID,
		MemberProfileID: memberProfileID,
	})
}
