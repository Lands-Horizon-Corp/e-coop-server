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

func MemberProfileArchiveManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberProfileArchive, types.MemberProfileArchiveResponse, types.MemberProfileArchiveRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberProfileArchive, types.MemberProfileArchiveResponse, types.MemberProfileArchiveRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberProfileArchive) *types.MemberProfileArchiveResponse {
			if data == nil {
				return nil
			}
			return &types.MemberProfileArchiveResponse{
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
				Category:        data.Category,
			}
		},
		Created: func(data *types.MemberProfileArchive) registry.Topics {
			events := []string{
				"member_profile_archive.create",
				fmt.Sprintf("member_profile_archive.create.%s", data.ID),
			}
			if data.BranchID != nil {
				events = append(events, fmt.Sprintf("member_profile_archive.create.branch.%s", *data.BranchID))
			}
			if data.OrganizationID != nil {
				events = append(events, fmt.Sprintf("member_profile_archive.create.organization.%s", *data.OrganizationID))
			}
			return events
		},
		Updated: func(data *types.MemberProfileArchive) registry.Topics {
			events := []string{
				"member_profile_archive.update",
				fmt.Sprintf("member_profile_archive.update.%s", data.ID),
			}
			if data.BranchID != nil {
				events = append(events, fmt.Sprintf("member_profile_archive.update.branch.%s", *data.BranchID))
			}
			if data.OrganizationID != nil {
				events = append(events, fmt.Sprintf("member_profile_archive.update.organization.%s", *data.OrganizationID))
			}
			return events
		},
		Deleted: func(data *types.MemberProfileArchive) registry.Topics {
			events := []string{
				"member_profile_archive.delete",
				fmt.Sprintf("member_profile_archive.delete.%s", data.ID),
			}
			if data.BranchID != nil {
				events = append(events, fmt.Sprintf("member_profile_archive.delete.branch.%s", *data.BranchID))
			}
			if data.OrganizationID != nil {
				events = append(events, fmt.Sprintf("member_profile_archive.delete.organization.%s", *data.OrganizationID))
			}
			return events
		},
	})
}

func MemberProfileArchiveCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID *uuid.UUID, branchID *uuid.UUID) ([]*types.MemberProfileArchive, error) {
	return MemberProfileArchiveManager(service).Find(context, &types.MemberProfileArchive{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
