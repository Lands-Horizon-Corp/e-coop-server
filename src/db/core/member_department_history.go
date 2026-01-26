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

func MemberDepartmentHistoryManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberDepartmentHistory, types.MemberDepartmentHistoryResponse, types.MemberDepartmentHistoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberDepartmentHistory,
		types.MemberDepartmentHistoryResponse,
		types.MemberDepartmentHistoryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"Organization", "Branch", "MemberDepartment", "MemberProfile",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberDepartmentHistory) *types.MemberDepartmentHistoryResponse {
			if data == nil {
				return nil
			}
			return &types.MemberDepartmentHistoryResponse{
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
				MemberDepartmentID: data.MemberDepartmentID,
				MemberDepartment:   MemberDepartmentManager(service).ToModel(data.MemberDepartment),
				MemberProfileID:    data.MemberProfileID,
				MemberProfile:      MemberProfileManager(service).ToModel(data.MemberProfile),
			}
		},
		Created: func(data *types.MemberDepartmentHistory) registry.Topics {
			return []string{
				"member_department_history.create",
				fmt.Sprintf("member_department_history.create.%s", data.ID),
				fmt.Sprintf("member_department_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_department_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_department_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *types.MemberDepartmentHistory) registry.Topics {
			return []string{
				"member_department_history.update",
				fmt.Sprintf("member_department_history.update.%s", data.ID),
				fmt.Sprintf("member_department_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_department_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_department_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *types.MemberDepartmentHistory) registry.Topics {
			return []string{
				"member_department_history.delete",
				fmt.Sprintf("member_department_history.delete.%s", data.ID),
				fmt.Sprintf("member_department_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_department_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_department_history.delete.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func MemberDepartmentHistoryCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberDepartmentHistory, error) {
	return MemberDepartmentHistoryManager(service).Find(context, &types.MemberDepartmentHistory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func MemberDepartmentHistoryMemberProfileID(context context.Context, service *horizon.HorizonService,
	memberProfileID, organizationID, branchID uuid.UUID) ([]*types.MemberDepartmentHistory, error) {
	return MemberDepartmentHistoryManager(service).Find(context, &types.MemberDepartmentHistory{
		OrganizationID:  organizationID,
		BranchID:        branchID,
		MemberProfileID: memberProfileID,
	})
}
