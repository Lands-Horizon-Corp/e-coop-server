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

func MemberAddressManager(service *horizon.HorizonService) *registry.Registry[types.MemberAddress, types.MemberAddressResponse, types.MemberAddressRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberAddress, types.MemberAddressResponse, types.MemberAddressRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberAddress) *types.MemberAddressResponse {
			if data == nil {
				return nil
			}
			return &types.MemberAddressResponse{
				ID:              data.ID,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:     *data.CreatedByID,
				CreatedBy:       UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:     *data.UpdatedByID,
				UpdatedBy:       UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:  data.OrganizationID,
				Organization:    OrganizationManager(service).ToModel(data.Organization),
				BranchID:        data.BranchID,
				Branch:          BranchManager(service).ToModel(data.Branch),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   MemberProfileManager(service).ToModel(data.MemberProfile),
				Label:           data.Label,
				City:            data.City,
				CountryCode:     data.CountryCode,
				PostalCode:      data.PostalCode,
				ProvinceState:   data.ProvinceState,
				Barangay:        data.Barangay,
				Landmark:        data.Landmark,
				Address:         data.Address,
				Longitude:       data.Longitude,
				Latitude:        data.Latitude,
			}
		},

		Created: func(data *types.MemberAddress) registry.Topics {
			return []string{
				"member_address.create",
				fmt.Sprintf("member_address.create.%s", data.ID),
				fmt.Sprintf("member_address.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_address.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberAddress) registry.Topics {
			return []string{
				"member_address.update",
				fmt.Sprintf("member_address.update.%s", data.ID),
				fmt.Sprintf("member_address.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_address.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberAddress) registry.Topics {
			return []string{
				"member_address.delete",
				fmt.Sprintf("member_address.delete.%s", data.ID),
				fmt.Sprintf("member_address.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_address.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberAddressCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberAddress, error) {
	return MemberAddressManager(service).Find(context, &types.MemberAddress{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
