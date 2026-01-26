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

func LoanComakerMemberManager(service *horizon.HorizonService) *registry.Registry[
	types.LoanComakerMember, types.LoanComakerMemberResponse, types.LoanComakerMemberRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.LoanComakerMember, types.LoanComakerMemberResponse, types.LoanComakerMemberRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"MemberProfile", "LoanTransaction",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.LoanComakerMember) *types.LoanComakerMemberResponse {
			if data == nil {
				return nil
			}
			return &types.LoanComakerMemberResponse{
				ID:                data.ID,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				CreatedByID:       data.CreatedByID,
				CreatedBy:         UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:       data.UpdatedByID,
				UpdatedBy:         UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:    data.OrganizationID,
				Organization:      OrganizationManager(service).ToModel(data.Organization),
				BranchID:          data.BranchID,
				Branch:            BranchManager(service).ToModel(data.Branch),
				MemberProfileID:   data.MemberProfileID,
				MemberProfile:     MemberProfileManager(service).ToModel(data.MemberProfile),
				LoanTransactionID: data.LoanTransactionID,
				LoanTransaction:   LoanTransactionManager(service).ToModel(data.LoanTransaction),
				Description:       data.Description,
				Amount:            data.Amount,
				MonthsCount:       data.MonthsCount,
				YearCount:         data.YearCount,
			}
		},

		Created: func(data *types.LoanComakerMember) registry.Topics {
			return []string{
				"loan_comaker_member.create",
				fmt.Sprintf("loan_comaker_member.create.%s", data.ID),
				fmt.Sprintf("loan_comaker_member.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_comaker_member.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.LoanComakerMember) registry.Topics {
			return []string{
				"loan_comaker_member.update",
				fmt.Sprintf("loan_comaker_member.update.%s", data.ID),
				fmt.Sprintf("loan_comaker_member.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_comaker_member.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.LoanComakerMember) registry.Topics {
			return []string{
				"loan_comaker_member.delete",
				fmt.Sprintf("loan_comaker_member.delete.%s", data.ID),
				fmt.Sprintf("loan_comaker_member.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_comaker_member.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func LoanComakerMemberCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.LoanComakerMember, error) {
	return LoanComakerMemberManager(service).Find(context, &types.LoanComakerMember{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
