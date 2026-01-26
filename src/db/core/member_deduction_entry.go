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

func MemberDeductionEntryManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberDeductionEntry, types.MemberDeductionEntryResponse, types.MemberDeductionEntryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberDeductionEntry, types.MemberDeductionEntryResponse, types.MemberDeductionEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "MemberProfile", "Account",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberDeductionEntry) *types.MemberDeductionEntryResponse {
			if data == nil {
				return nil
			}
			return &types.MemberDeductionEntryResponse{
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
				AccountID:       data.AccountID,
				Account:         AccountManager(service).ToModel(data.Account),
				Name:            data.Name,
				Description:     data.Description,
				MembershipDate:  data.MembershipDate.Format(time.RFC3339),
			}
		},

		Created: func(data *types.MemberDeductionEntry) registry.Topics {
			return []string{
				"member_deduction_entry.create",
				fmt.Sprintf("member_deduction_entry.create.%s", data.ID),
				fmt.Sprintf("member_deduction_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_deduction_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberDeductionEntry) registry.Topics {
			return []string{
				"member_deduction_entry.update",
				fmt.Sprintf("member_deduction_entry.update.%s", data.ID),
				fmt.Sprintf("member_deduction_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_deduction_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberDeductionEntry) registry.Topics {
			return []string{
				"member_deduction_entry.delete",
				fmt.Sprintf("member_deduction_entry.delete.%s", data.ID),
				fmt.Sprintf("member_deduction_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_deduction_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberDeductionEntryCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberDeductionEntry, error) {
	return MemberDeductionEntryManager(service).Find(context, &types.MemberDeductionEntry{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
