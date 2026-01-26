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

func ComakerMemberProfileManager(service *horizon.HorizonService) *registry.Registry[
	types.ComakerMemberProfile, types.ComakerMemberProfileResponse, types.ComakerMemberProfileRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.ComakerMemberProfile, types.ComakerMemberProfileResponse, types.ComakerMemberProfileRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "LoanTransaction", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.ComakerMemberProfile) *types.ComakerMemberProfileResponse {
			if data == nil {
				return nil
			}
			return &types.ComakerMemberProfileResponse{
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
				LoanTransactionID: data.LoanTransactionID,
				LoanTransaction:   LoanTransactionManager(service).ToModel(data.LoanTransaction),
				MemberProfileID:   data.MemberProfileID,
				MemberProfile:     MemberProfileManager(service).ToModel(data.MemberProfile),
				Amount:            data.Amount,
				Description:       data.Description,
				MonthsCount:       data.MonthsCount,
				YearCount:         data.YearCount,
			}
		},
		Created: func(data *types.ComakerMemberProfile) registry.Topics {
			return []string{
				"comaker_member_profile.create",
				fmt.Sprintf("comaker_member_profile.create.%s", data.ID),
				fmt.Sprintf("comaker_member_profile.create.branch.%s", data.BranchID),
				fmt.Sprintf("comaker_member_profile.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("comaker_member_profile.create.loan_transaction.%s", data.LoanTransactionID),
			}
		},
		Updated: func(data *types.ComakerMemberProfile) registry.Topics {
			return []string{
				"comaker_member_profile.update",
				fmt.Sprintf("comaker_member_profile.update.%s", data.ID),
				fmt.Sprintf("comaker_member_profile.update.branch.%s", data.BranchID),
				fmt.Sprintf("comaker_member_profile.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("comaker_member_profile.update.loan_transaction.%s", data.LoanTransactionID),
			}
		},
		Deleted: func(data *types.ComakerMemberProfile) registry.Topics {
			return []string{
				"comaker_member_profile.delete",
				fmt.Sprintf("comaker_member_profile.delete.%s", data.ID),
				fmt.Sprintf("comaker_member_profile.delete.branch.%s", data.BranchID),
				fmt.Sprintf("comaker_member_profile.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("comaker_member_profile.delete.loan_transaction.%s", data.LoanTransactionID),
			}
		},
	})
}

func ComakerMemberProfileCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.ComakerMemberProfile, error) {
	return ComakerMemberProfileManager(service).Find(context, &types.ComakerMemberProfile{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func ComakerMemberProfileByLoanTransaction(context context.Context, service *horizon.HorizonService, loanTransactionID uuid.UUID) ([]*types.ComakerMemberProfile, error) {
	return ComakerMemberProfileManager(service).Find(context, &types.ComakerMemberProfile{
		LoanTransactionID: loanTransactionID,
	})
}
