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

func MemberOtherInformationEntryManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberOtherInformationEntry, types.MemberOtherInformationEntryResponse, types.MemberOtherInformationEntryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberOtherInformationEntry, types.MemberOtherInformationEntryResponse, types.MemberOtherInformationEntryRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberOtherInformationEntry) *types.MemberOtherInformationEntryResponse {
			if data == nil {
				return nil
			}
			return &types.MemberOtherInformationEntryResponse{
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
				EntryDate:      data.EntryDate.Format(time.RFC3339),
			}
		},

		Created: func(data *types.MemberOtherInformationEntry) registry.Topics {
			return []string{
				"member_other_information_entry.create",
				fmt.Sprintf("member_other_information_entry.create.%s", data.ID),
				fmt.Sprintf("member_other_information_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_other_information_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberOtherInformationEntry) registry.Topics {
			return []string{
				"member_other_information_entry.update",
				fmt.Sprintf("member_other_information_entry.update.%s", data.ID),
				fmt.Sprintf("member_other_information_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_other_information_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberOtherInformationEntry) registry.Topics {
			return []string{
				"member_other_information_entry.delete",
				fmt.Sprintf("member_other_information_entry.delete.%s", data.ID),
				fmt.Sprintf("member_other_information_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_other_information_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberOtherInformationEntryCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberOtherInformationEntry, error) {
	return MemberOtherInformationEntryManager(service).Find(context, &types.MemberOtherInformationEntry{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
