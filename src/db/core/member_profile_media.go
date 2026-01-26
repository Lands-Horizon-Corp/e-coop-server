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

func MemberProfileMediaManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberProfileMedia, types.MemberProfileMediaResponse, types.MemberProfileMediaRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberProfileMedia, types.MemberProfileMediaResponse, types.MemberProfileMediaRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberProfileMedia) *types.MemberProfileMediaResponse {
			if data == nil {
				return nil
			}
			return &types.MemberProfileMediaResponse{
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
				MediaID:         data.MediaID,
				Media:           MediaManager(service).ToModel(data.Media),
				Name:            data.Name,
				Description:     data.Description,
			}
		},
		Created: func(data *types.MemberProfileMedia) registry.Topics {
			events := []string{
				"member_profile_media.create",
				fmt.Sprintf("member_profile_media.create.%s", data.ID),
			}
			if data.BranchID != nil {
				events = append(events, fmt.Sprintf("member_profile_media.create.branch.%s", *data.BranchID))
			}
			if data.OrganizationID != nil {
				events = append(events, fmt.Sprintf("member_profile_media.create.organization.%s", *data.OrganizationID))
			}
			return events
		},
		Updated: func(data *types.MemberProfileMedia) registry.Topics {
			events := []string{
				"member_profile_media.update",
				fmt.Sprintf("member_profile_media.update.%s", data.ID),
			}
			if data.BranchID != nil {
				events = append(events, fmt.Sprintf("member_profile_media.update.branch.%s", *data.BranchID))
			}
			if data.OrganizationID != nil {
				events = append(events, fmt.Sprintf("member_profile_media.update.organization.%s", *data.OrganizationID))
			}
			return events
		},
		Deleted: func(data *types.MemberProfileMedia) registry.Topics {
			events := []string{
				"member_profile_media.delete",
				fmt.Sprintf("member_profile_media.delete.%s", data.ID),
			}
			if data.BranchID != nil {
				events = append(events, fmt.Sprintf("member_profile_media.delete.branch.%s", *data.BranchID))
			}
			if data.OrganizationID != nil {
				events = append(events, fmt.Sprintf("member_profile_media.delete.organization.%s", *data.OrganizationID))
			}
			return events
		},
	})
}

func MemberProfileMediaCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID *uuid.UUID, branchID *uuid.UUID) ([]*types.MemberProfileMedia, error) {
	return MemberProfileMediaManager(service).Find(context, &types.MemberProfileMedia{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
