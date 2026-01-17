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

func MemberTypeManager(service *horizon.HorizonService) *registry.Registry[types.MemberType, types.MemberTypeResponse, types.MemberTypeRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberType, types.MemberTypeResponse, types.MemberTypeRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"BrowseReferences",
			"BrowseReferences.Account",
			"BrowseReferences.MemberType",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberType) *types.MemberTypeResponse {
			if data == nil {
				return nil
			}
			return &types.MemberTypeResponse{
				ID:                         data.ID,
				CreatedAt:                  data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                data.CreatedByID,
				CreatedBy:                  UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:                  data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                data.UpdatedByID,
				UpdatedBy:                  UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:             data.OrganizationID,
				Organization:               OrganizationManager(service).ToModel(data.Organization),
				BranchID:                   data.BranchID,
				Branch:                     BranchManager(service).ToModel(data.Branch),
				Prefix:                     data.Prefix,
				Name:                       data.Name,
				Description:                data.Description,
				BrowseReferenceDescription: data.BrowseReferenceDescription,
				BrowseReferences:           BrowseReferenceManager(service).ToModels(data.BrowseReferences),
			}
		},

		Created: func(data *types.MemberType) registry.Topics {
			return []string{
				"member_type.create",
				fmt.Sprintf("member_type.create.%s", data.ID),
				fmt.Sprintf("member_type.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_type.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberType) registry.Topics {
			return []string{
				"member_type.update",
				fmt.Sprintf("member_type.update.%s", data.ID),
				fmt.Sprintf("member_type.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_type.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberType) registry.Topics {
			return []string{
				"member_type.delete",
				fmt.Sprintf("member_type.delete.%s", data.ID),
				fmt.Sprintf("member_type.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_type.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func memberTypeSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB, userID uuid.UUID,
	organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	memberType := []*types.MemberType{
		{

			Name:           "New",
			Prefix:         "NEW",
			Description:    "Recently registered member, no activity yet.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Active",
			Prefix:         "ACT",
			Description:    "Regularly engaged member with no issues.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Loyal",
			Prefix:         "LOY",
			Description:    "Consistently active over a long period; high retention.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "VIP",
			Prefix:         "VIP",
			Description:    "Very high-value member with premium privileges.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Reported",
			Prefix:         "RPT",
			Description:    "Flagged by community or system for review.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Suspended",
			Prefix:         "SUS",
			Description:    "Temporarily barred from activities pending resolution.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Banned",
			Prefix:         "BAN",
			Description:    "Permanently barred due to policy violations.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Closed",
			Prefix:         "CLS",
			Description:    "Account closed by user request or administrative action.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Alumni",
			Prefix:         "ALM",
			Description:    "Former member with notable contributions.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Pending",
			Prefix:         "PND",
			Description:    "Awaiting verification or approval.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Dormant",
			Prefix:         "DRM",
			Description:    "Inactive for a long period with no recent engagement.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Guest",
			Prefix:         "GST",
			Description:    "Limited access member without full privileges.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Moderator",
			Prefix:         "MOD",
			Description:    "Member with special privileges to manage content or users.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Admin",
			Prefix:         "ADM",
			Description:    "Administrator with full access and control.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
	}
	for _, data := range memberType {
		if err := MemberTypeManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member type %s", data.Name)
		}
	}
	return nil
}

func MemberTypeCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberType, error) {
	return MemberTypeManager(service).Find(context, &types.MemberType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
