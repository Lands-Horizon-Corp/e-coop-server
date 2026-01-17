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

func MemberMutualFundHistoryManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberMutualFundHistory, types.MemberMutualFundHistoryResponse, types.MemberMutualFundHistoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberMutualFundHistory, types.MemberMutualFundHistoryResponse, types.MemberMutualFundHistoryRequest]{
		Preloads: []string{"Organization", "Branch", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberMutualFundHistory) *types.MemberMutualFundHistoryResponse {
			if data == nil {
				return nil
			}
			return &types.MemberMutualFundHistoryResponse{
				ID:              data.ID,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
				OrganizationID:  data.OrganizationID,
				Organization:    OrganizationManager(service).ToModel(data.Organization),
				BranchID:        data.BranchID,
				Branch:          BranchManager(service).ToModel(data.Branch),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   MemberProfileManager(service).ToModel(data.MemberProfile),
				Title:           data.Title,
				Amount:          data.Amount,
				Description:     data.Description,
			}
		},

		Created: func(data *types.MemberMutualFundHistory) registry.Topics {
			return []string{
				"member_mutual_fund_history.create",
				fmt.Sprintf("member_mutual_fund_history.create.%s", data.ID),
				fmt.Sprintf("member_mutual_fund_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_mutual_fund_history.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberMutualFundHistory) registry.Topics {
			return []string{
				"member_mutual_fund_history.update",
				fmt.Sprintf("member_mutual_fund_history.update.%s", data.ID),
				fmt.Sprintf("member_mutual_fund_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_mutual_fund_history.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberMutualFundHistory) registry.Topics {
			return []string{
				"member_mutual_fund_history.delete",
				fmt.Sprintf("member_mutual_fund_history.delete.%s", data.ID),
				fmt.Sprintf("member_mutual_fund_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_mutual_fund_history.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberMutualFundHistoryCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberMutualFundHistory, error) {
	return MemberMutualFundHistoryManager(service).Find(context, &types.MemberMutualFundHistory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
