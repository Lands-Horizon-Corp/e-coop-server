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

func MemberAssetManager(service *horizon.HorizonService) *registry.Registry[types.MemberAsset, types.MemberAssetResponse, types.MemberAssetRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberAsset, types.MemberAssetResponse, types.MemberAssetRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberAsset) *types.MemberAssetResponse {
			if data == nil {
				return nil
			}
			return &types.MemberAssetResponse{
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
				MediaID:         data.MediaID,
				Media:           MediaManager(service).ToModel(data.Media),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   MemberProfileManager(service).ToModel(data.MemberProfile),
				Name:            data.Name,
				EntryDate:       data.EntryDate.Format(time.RFC3339),
				Description:     data.Description,
				Cost:            data.Cost,
			}
		},

		Created: func(data *types.MemberAsset) registry.Topics {
			return []string{
				"member_asset.create",
				fmt.Sprintf("member_asset.create.%s", data.ID),
				fmt.Sprintf("member_asset.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_asset.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberAsset) registry.Topics {
			return []string{
				"member_asset.update",
				fmt.Sprintf("member_asset.update.%s", data.ID),
				fmt.Sprintf("member_asset.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_asset.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberAsset) registry.Topics {
			return []string{
				"member_asset.delete",
				fmt.Sprintf("member_asset.delete.%s", data.ID),
				fmt.Sprintf("member_asset.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_asset.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberAssetCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberAsset, error) {
	return MemberAssetManager(service).Find(context, &types.MemberAsset{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
