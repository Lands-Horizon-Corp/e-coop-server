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

func MemberGovernmentBenefitManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberGovernmentBenefit, types.MemberGovernmentBenefitResponse, types.MemberGovernmentBenefitRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberGovernmentBenefit, types.MemberGovernmentBenefitResponse, types.MemberGovernmentBenefitRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile", "FrontMedia", "BackMedia"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberGovernmentBenefit) *types.MemberGovernmentBenefitResponse {
			if data == nil {
				return nil
			}
			var expiryDateStr *string
			if data.ExpiryDate != nil {
				s := data.ExpiryDate.Format("2006-01-02")
				expiryDateStr = &s
			}
			return &types.MemberGovernmentBenefitResponse{
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
				FrontMediaID:    data.FrontMediaID,
				FrontMedia:      MediaManager(service).ToModel(data.FrontMedia),
				BackMediaID:     data.BackMediaID,
				BackMedia:       MediaManager(service).ToModel(data.BackMedia),
				CountryCode:     data.CountryCode,
				Description:     data.Description,
				Name:            data.Name,
				Value:           data.Value,
				ExpiryDate:      expiryDateStr,
			}
		},

		Created: func(data *types.MemberGovernmentBenefit) registry.Topics {
			return []string{
				"member_government_benefit.create",
				fmt.Sprintf("member_government_benefit.create.%s", data.ID),
				fmt.Sprintf("member_government_benefit.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_government_benefit.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberGovernmentBenefit) registry.Topics {
			return []string{
				"member_government_benefit.update",
				fmt.Sprintf("member_government_benefit.update.%s", data.ID),
				fmt.Sprintf("member_government_benefit.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_government_benefit.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberGovernmentBenefit) registry.Topics {
			return []string{
				"member_government_benefit.delete",
				fmt.Sprintf("member_government_benefit.delete.%s", data.ID),
				fmt.Sprintf("member_government_benefit.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_government_benefit.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberGovernmentBenefitCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberGovernmentBenefit, error) {
	return MemberGovernmentBenefitManager(service).Find(context, &types.MemberGovernmentBenefit{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
