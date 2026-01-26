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

func MemberCloseRemarkManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberCloseRemark, types.MemberCloseRemarkResponse, types.MemberCloseRemarkRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberCloseRemark, types.MemberCloseRemarkResponse, types.MemberCloseRemarkRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberCloseRemark) *types.MemberCloseRemarkResponse {
			if data == nil {
				return nil
			}
			return &types.MemberCloseRemarkResponse{
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
				MemberProfileID: *data.MemberProfileID,
				MemberProfile:   MemberProfileManager(service).ToModel(data.MemberProfile),
				Reason:          data.Reason,
				Description:     data.Description,
			}
		},

		Created: func(data *types.MemberCloseRemark) registry.Topics {
			return []string{
				"member_close_remark.create",
				fmt.Sprintf("member_close_remark.create.%s", data.ID),
				fmt.Sprintf("member_close_remark.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_close_remark.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberCloseRemark) registry.Topics {
			return []string{
				"member_close_remark.update",
				fmt.Sprintf("member_close_remark.update.%s", data.ID),
				fmt.Sprintf("member_close_remark.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_close_remark.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberCloseRemark) registry.Topics {
			return []string{
				"member_close_remark.delete",
				fmt.Sprintf("member_close_remark.delete.%s", data.ID),
				fmt.Sprintf("member_close_remark.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_close_remark.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberCloseRemarkCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberCloseRemark, error) {
	return MemberCloseRemarkManager(service).Find(context, &types.MemberCloseRemark{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
