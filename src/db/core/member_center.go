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

func MemberCenterManager(service *horizon.HorizonService) *registry.Registry[types.MemberCenter, types.MemberCenterResponse, types.MemberCenterRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberCenter, types.MemberCenterResponse, types.MemberCenterRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberCenter) *types.MemberCenterResponse {
			if data == nil {
				return nil
			}
			return &types.MemberCenterResponse{
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

		Created: func(data *types.MemberCenter) registry.Topics {
			return []string{
				"member_center.create",
				fmt.Sprintf("member_center.create.%s", data.ID),
				fmt.Sprintf("member_center.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_center.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberCenter) registry.Topics {
			return []string{
				"member_center.update",
				fmt.Sprintf("member_center.update.%s", data.ID),
				fmt.Sprintf("member_center.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_center.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberCenter) registry.Topics {
			return []string{
				"member_center.delete",
				fmt.Sprintf("member_center.delete.%s", data.ID),
				fmt.Sprintf("member_center.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_center.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func memberCenterSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB,
	userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	memberCenter := []*types.MemberCenter{
		{
			Name:           "Main Wellness Center",
			Description:    "Provides health and wellness programs.",
			OrganizationID: organizationID,
			BranchID:       branchID,
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
		},
		{

			Name:           "Training Hub",
			Description:    "Offers skill-building and training for members.",
			OrganizationID: organizationID,
			BranchID:       branchID,
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
		},
		{

			Name:           "Community Support Center",
			Description:    "Focuses on community support services and events.",
			OrganizationID: organizationID,
			BranchID:       branchID,
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
		},
	}
	for _, data := range memberCenter {
		if err := MemberCenterManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member center %s", data.Name)
		}
	}
	return nil
}

func MemberCenterCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberCenter, error) {
	return MemberCenterManager(service).Find(context, &types.MemberCenter{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
