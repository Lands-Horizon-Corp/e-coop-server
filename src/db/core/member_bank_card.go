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

func MemberBankCardManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberBankCard, types.MemberBankCardResponse, types.MemberBankCardRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberBankCard, types.MemberBankCardResponse, types.MemberBankCardRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Bank", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberBankCard) *types.MemberBankCardResponse {
			if data == nil {
				return nil
			}
			return &types.MemberBankCardResponse{
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
				BankID:          data.BankID,
				Bank:            BankManager(service).ToModel(data.Bank),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   MemberProfileManager(service).ToModel(data.MemberProfile),
				AccountNumber:   data.AccountNumber,
				CardName:        data.CardName,
				ExpirationDate:  data.ExpirationDate.Format(time.RFC3339),
				IsDefault:       data.IsDefault,
			}
		},

		Created: func(data *types.MemberBankCard) registry.Topics {
			return []string{
				"member_bank_card.create",
				fmt.Sprintf("member_bank_card.create.%s", data.ID),
				fmt.Sprintf("member_bank_card.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_bank_card.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberBankCard) registry.Topics {
			return []string{
				"member_bank_card.update",
				fmt.Sprintf("member_bank_card.update.%s", data.ID),
				fmt.Sprintf("member_bank_card.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_bank_card.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberBankCard) registry.Topics {
			return []string{
				"member_bank_card.delete",
				fmt.Sprintf("member_bank_card.delete.%s", data.ID),
				fmt.Sprintf("member_bank_card.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_bank_card.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberBankCardCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberBankCard, error) {
	return MemberBankCardManager(service).Find(context, &types.MemberBankCard{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
