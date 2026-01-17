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

func MemberVerificationManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberVerification, types.MemberVerificationResponse, types.MemberVerificationRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberVerification, types.MemberVerificationResponse, types.MemberVerificationRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile", "VerifiedByUser"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberVerification) *types.MemberVerificationResponse {
			if data == nil {
				return nil
			}
			return &types.MemberVerificationResponse{
				ID:               data.ID,
				CreatedAt:        data.CreatedAt.Format(time.RFC3339),
				CreatedByID:      data.CreatedByID,
				CreatedBy:        UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:        data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:      data.UpdatedByID,
				UpdatedBy:        UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:   data.OrganizationID,
				Organization:     OrganizationManager(service).ToModel(data.Organization),
				BranchID:         *data.BranchID,
				Branch:           BranchManager(service).ToModel(data.Branch),
				MemberProfileID:  *data.MemberProfileID,
				MemberProfile:    MemberProfileManager(service).ToModel(data.MemberProfile),
				VerifiedByUserID: *data.VerifiedByUserID,
				VerifiedByUser:   UserManager(service).ToModel(data.VerifiedByUser),
				Status:           data.Status,
			}
		},

		Created: func(data *types.MemberVerification) registry.Topics {
			return []string{
				"member_verification.create",
				fmt.Sprintf("member_verification.create.%s", data.ID),
				fmt.Sprintf("member_verification.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_verification.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberVerification) registry.Topics {
			return []string{
				"member_verification.update",
				fmt.Sprintf("member_verification.update.%s", data.ID),
				fmt.Sprintf("member_verification.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_verification.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberVerification) registry.Topics {
			return []string{
				"member_verification.delete",
				fmt.Sprintf("member_verification.delete.%s", data.ID),
				fmt.Sprintf("member_verification.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_verification.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberVerificationCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberVerification, error) {
	return MemberVerificationManager(service).Find(context, &types.MemberVerification{
		OrganizationID: organizationID,
		BranchID:       &branchID,
	})
}
