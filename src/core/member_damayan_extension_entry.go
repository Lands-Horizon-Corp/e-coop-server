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

func MemberDamayanExtensionEntryManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberDamayanExtensionEntry, types.MemberDamayanExtensionEntryResponse, types.MemberDamayanExtensionEntryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberDamayanExtensionEntry, types.MemberDamayanExtensionEntryResponse, types.MemberDamayanExtensionEntryRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberDamayanExtensionEntry) *types.MemberDamayanExtensionEntryResponse {
			if data == nil {
				return nil
			}
			var birthdateStr *string
			if data.Birthdate != nil {
				s := data.Birthdate.Format(time.RFC3339)
				birthdateStr = &s
			}
			return &types.MemberDamayanExtensionEntryResponse{
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
				Name:            data.Name,
				Description:     data.Description,
				Birthdate:       birthdateStr,
			}
		},

		Created: func(data *types.MemberDamayanExtensionEntry) registry.Topics {
			return []string{
				"member_damayan_extension_entry.create",
				fmt.Sprintf("member_damayan_extension_entry.create.%s", data.ID),
				fmt.Sprintf("member_damayan_extension_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_damayan_extension_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberDamayanExtensionEntry) registry.Topics {
			return []string{
				"member_damayan_extension_entry.update",
				fmt.Sprintf("member_damayan_extension_entry.update.%s", data.ID),
				fmt.Sprintf("member_damayan_extension_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_damayan_extension_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberDamayanExtensionEntry) registry.Topics {
			return []string{
				"member_damayan_extension_entry.delete",
				fmt.Sprintf("member_damayan_extension_entry.delete.%s", data.ID),
				fmt.Sprintf("member_damayan_extension_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_damayan_extension_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberDamayanExtensionEntryCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberDamayanExtensionEntry, error) {
	return MemberDamayanExtensionEntryManager(service).Find(context, &types.MemberDamayanExtensionEntry{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
