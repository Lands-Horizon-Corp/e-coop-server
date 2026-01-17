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

func MemberContactReferenceManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberContactReference, types.MemberContactReferenceResponse, types.MemberContactReferenceRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberContactReference, types.MemberContactReferenceResponse, types.MemberContactReferenceRequest]{
		Preloads: []string{"MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberContactReference) *types.MemberContactReferenceResponse {
			if data == nil {
				return nil
			}
			return &types.MemberContactReferenceResponse{
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
				ContactNumber:   data.ContactNumber,
			}
		},

		Created: func(data *types.MemberContactReference) registry.Topics {
			return []string{
				"member_contact_reference.create",
				fmt.Sprintf("member_contact_reference.create.%s", data.ID),
				fmt.Sprintf("member_contact_reference.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_contact_reference.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberContactReference) registry.Topics {
			return []string{
				"member_contact_reference.update",
				fmt.Sprintf("member_contact_reference.update.%s", data.ID),
				fmt.Sprintf("member_contact_reference.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_contact_reference.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberContactReference) registry.Topics {
			return []string{
				"member_contact_reference.delete",
				fmt.Sprintf("member_contact_reference.delete.%s", data.ID),
				fmt.Sprintf("member_contact_reference.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_contact_reference.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberContactReferenceCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberContactReference, error) {
	return MemberContactReferenceManager(service).Find(context, &types.MemberContactReference{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
