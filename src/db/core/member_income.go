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

func MemberIncomeManager(service *horizon.HorizonService) *registry.Registry[types.MemberIncome, types.MemberIncomeResponse, types.MemberIncomeRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberIncome, types.MemberIncomeResponse, types.MemberIncomeRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberIncome) *types.MemberIncomeResponse {
			if data == nil {
				return nil
			}
			var releaseDateStr *string
			if data.ReleaseDate != nil {
				s := data.ReleaseDate.Format(time.RFC3339)
				releaseDateStr = &s
			}
			return &types.MemberIncomeResponse{
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
				MediaID:         data.MediaID,
				Media:           MediaManager(service).ToModel(data.Media),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   MemberProfileManager(service).ToModel(data.MemberProfile),
				Name:            data.Name,
				Source:          data.Source,
				Amount:          data.Amount,
				ReleaseDate:     releaseDateStr,
			}
		},

		Created: func(data *types.MemberIncome) registry.Topics {
			return []string{
				"member_income.create",
				fmt.Sprintf("member_income.create.%s", data.ID),
				fmt.Sprintf("member_income.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_income.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberIncome) registry.Topics {
			return []string{
				"member_income.update",
				fmt.Sprintf("member_income.update.%s", data.ID),
				fmt.Sprintf("member_income.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_income.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberIncome) registry.Topics {
			return []string{
				"member_income.delete",
				fmt.Sprintf("member_income.delete.%s", data.ID),
				fmt.Sprintf("member_income.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_income.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberIncomeCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID,
	branchID uuid.UUID) ([]*types.MemberIncome, error) {
	return MemberIncomeManager(service).Find(context, &types.MemberIncome{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
