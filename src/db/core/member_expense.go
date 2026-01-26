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

func MemberExpenseManager(service *horizon.HorizonService) *registry.Registry[types.MemberExpense, types.MemberExpenseResponse, types.MemberExpenseRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberExpense, types.MemberExpenseResponse, types.MemberExpenseRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberExpense) *types.MemberExpenseResponse {
			if data == nil {
				return nil
			}
			return &types.MemberExpenseResponse{
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
				Amount:          data.Amount,
				Description:     data.Description,
			}
		},

		Created: func(data *types.MemberExpense) registry.Topics {
			return []string{
				"member_expense.create",
				fmt.Sprintf("member_expense.create.%s", data.ID),
				fmt.Sprintf("member_expense.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_expense.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberExpense) registry.Topics {
			return []string{
				"member_expense.update",
				fmt.Sprintf("member_expense.update.%s", data.ID),
				fmt.Sprintf("member_expense.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_expense.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberExpense) registry.Topics {
			return []string{
				"member_expense.delete",
				fmt.Sprintf("member_expense.delete.%s", data.ID),
				fmt.Sprintf("member_expense.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_expense.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberExpenseCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberExpense, error) {
	return MemberExpenseManager(service).Find(context, &types.MemberExpense{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
