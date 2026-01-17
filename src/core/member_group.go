package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func MemberGroupManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberGroup, types.MemberGroupResponse, types.MemberGroupRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberGroup, types.MemberGroupResponse, types.MemberGroupRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberGroup) *types.MemberGroupResponse {
			if data == nil {
				return nil
			}
			return &types.MemberGroupResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),
				Name:           data.Name,
				Description:    data.Description,
			}
		},

		Created: func(data *types.MemberGroup) registry.Topics {
			return []string{
				"member_group.create",
				fmt.Sprintf("member_group.create.%s", data.ID),
				fmt.Sprintf("member_group.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_group.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberGroup) registry.Topics {
			return []string{
				"member_group.update",
				fmt.Sprintf("member_group.update.%s", data.ID),
				fmt.Sprintf("member_group.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_group.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberGroup) registry.Topics {
			return []string{
				"member_group.delete",
				fmt.Sprintf("member_group.delete.%s", data.ID),
				fmt.Sprintf("member_group.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_group.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func memberGroupSeed(context context.Context, service *horizon.HorizonService,
	tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	memberGroup := []*types.MemberGroup{
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Single Moms",
			Description:    "Support group for single mothers in the community.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Athletes",
			Description:    "Members who actively participate in sports and fitness.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Tech",
			Description:    "Members involved in information technology or development.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Graphics Artists",
			Description:    "Creative members who specialize in digital and graphic design.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Accountants",
			Description:    "Finance-focused members responsible for budgeting and auditing.",
		},
	}
	for _, data := range memberGroup {
		if err := MemberGroupManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member group %s", data.Name)
		}
	}
	return nil
}

func MemberGroupCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberGroup, error) {
	return MemberGroupManager(service).Find(context, &types.MemberGroup{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
