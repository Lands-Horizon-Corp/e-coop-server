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

func MemberJointAccountManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberJointAccount, types.MemberJointAccountResponse, types.MemberJointAccountRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberJointAccount, types.MemberJointAccountResponse, types.MemberJointAccountRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"MemberProfile", "PictureMedia", "SignatureMedia",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberJointAccount) *types.MemberJointAccountResponse {
			if data == nil {
				return nil
			}
			return &types.MemberJointAccountResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       OrganizationManager(service).ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             BranchManager(service).ToModel(data.Branch),
				MemberProfileID:    data.MemberProfileID,
				MemberProfile:      MemberProfileManager(service).ToModel(data.MemberProfile),
				PictureMediaID:     data.PictureMediaID,
				PictureMedia:       MediaManager(service).ToModel(data.PictureMedia),
				SignatureMediaID:   data.SignatureMediaID,
				SignatureMedia:     MediaManager(service).ToModel(data.SignatureMedia),
				Description:        data.Description,
				FirstName:          data.FirstName,
				MiddleName:         data.MiddleName,
				LastName:           data.LastName,
				FullName:           data.FullName,
				Suffix:             data.Suffix,
				Birthday:           data.Birthday.Format(time.RFC3339),
				FamilyRelationship: data.FamilyRelationship,
			}
		},

		Created: func(data *types.MemberJointAccount) registry.Topics {
			return []string{
				"member_joint_account.create",
				fmt.Sprintf("member_joint_account.create.%s", data.ID),
				fmt.Sprintf("member_joint_account.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_joint_account.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberJointAccount) registry.Topics {
			return []string{
				"member_joint_account.update",
				fmt.Sprintf("member_joint_account.update.%s", data.ID),
				fmt.Sprintf("member_joint_account.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_joint_account.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberJointAccount) registry.Topics {
			return []string{
				"member_joint_account.delete",
				fmt.Sprintf("member_joint_account.delete.%s", data.ID),
				fmt.Sprintf("member_joint_account.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_joint_account.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberJointAccountCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberJointAccount, error) {
	return MemberJointAccountManager(service).Find(context, &types.MemberJointAccount{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
