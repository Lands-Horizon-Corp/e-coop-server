package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func MutualFundManager(service *horizon.HorizonService) *registry.Registry[types.MutualFund, types.MutualFundResponse, types.MutualFundRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MutualFund, types.MutualFundResponse, types.MutualFundRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"MemberProfile",
			"MemberType",
			"AdditionalMembers",
			"AdditionalMembers.MemberType",
			"MutualFundTables",
			"Account", "Account.Currency"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MutualFund) *types.MutualFundResponse {
			if data == nil {
				return nil
			}
			var printedDate *string
			if data.PrintedDate != nil {
				formatted := data.PrintedDate.Format(time.RFC3339)
				printedDate = &formatted
			}
			return &types.MutualFundResponse{
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

				MemberTypeID: data.MemberTypeID,
				MemberType:   MemberTypeManager(service).ToModel(data.MemberType),

				AdditionalMembers: MutualFundAdditionalMembersManager(service).ToModels(data.AdditionalMembers),
				MutualFundTables:  MutualFundTableManager(service).ToModels(data.MutualFundTables),
				Name:              data.Name,
				Description:       data.Description,
				DateOfDeath:       data.DateOfDeath.Format(time.RFC3339),
				ExtensionOnly:     data.ExtensionOnly,
				Amount:            data.Amount,
				ComputationType:   data.ComputationType,
				AccountID:         data.AccountID,
				Account:           data.Account,

				PrintedByUserID: data.PrintedByUserID,
				PrintedByUser:   UserManager(service).ToModel(data.PrintedByUser),
				PrintedDate:     printedDate,

				PostAccountID:  data.PostAccountID,
				PostAccount:    data.PostAccount,
				PostedDate:     data.PostedDate,
				PostedByUserID: data.PostedByUserID,
			}
		},
		Created: func(data *types.MutualFund) registry.Topics {
			return []string{
				"mutual_fund.create",
				fmt.Sprintf("mutual_fund.create.%s", data.ID),
				fmt.Sprintf("mutual_fund.create.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund.create.member.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *types.MutualFund) registry.Topics {
			return []string{
				"mutual_fund.update",
				fmt.Sprintf("mutual_fund.update.%s", data.ID),
				fmt.Sprintf("mutual_fund.update.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund.update.member.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *types.MutualFund) registry.Topics {
			return []string{
				"mutual_fund.delete",
				fmt.Sprintf("mutual_fund.delete.%s", data.ID),
				fmt.Sprintf("mutual_fund.delete.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_fund.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("mutual_fund.delete.member.%s", data.MemberProfileID),
			}
		},
	})
}

func MutualFundCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MutualFund, error) {
	return MutualFundManager(service).Find(context, &types.MutualFund{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func MutualFundByMember(context context.Context, service *horizon.HorizonService,
	memberProfileID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MutualFund, error) {
	return MutualFundManager(service).Find(context, &types.MutualFund{
		MemberProfileID: memberProfileID,
		OrganizationID:  organizationID,
		BranchID:        branchID,
	})
}
func CreateMutualFundValue(
	ctx context.Context,
	service *horizon.HorizonService,
	req *types.MutualFundRequest,
	userOrg *types.UserOrganization,
) (*types.MutualFund, error) {

	now := time.Now().UTC()

	var additionalMembers []*types.MutualFundAdditionalMembers
	for _, additionalMember := range req.MutualFundAdditionalMembers {
		additionalMembers = append(additionalMembers, &types.MutualFundAdditionalMembers{
			ID:              uuid.New(),
			MemberTypeID:    additionalMember.MemberTypeID,
			NumberOfMembers: additionalMember.NumberOfMembers,
			Ratio:           additionalMember.Ratio,
			CreatedAt:       now,
			CreatedByID:     userOrg.UserID,
			UpdatedAt:       now,
			UpdatedByID:     userOrg.UserID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
		})
	}

	var mutualFundTables []*types.MutualFundTable
	for _, table := range req.MutualFundTables {
		mutualFundTables = append(mutualFundTables, &types.MutualFundTable{
			ID:             uuid.New(),
			MonthFrom:      table.MonthFrom,
			MonthTo:        table.MonthTo,
			Amount:         table.Amount,
			CreatedAt:      now,
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      now,
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
	}

	account, err := AccountManager(service).GetByID(ctx, req.AccountID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get account")
	}

	memberProfile, err := MemberProfileManager(service).GetByID(ctx, req.MemberProfileID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get member profile")
	}

	mutualFund := &types.MutualFund{
		ID:                uuid.New(),
		MemberProfileID:   req.MemberProfileID,
		MemberProfile:     memberProfile,
		MemberTypeID:      req.MemberTypeID,
		Name:              req.Name,
		Description:       req.Description,
		DateOfDeath:       req.DateOfDeath,
		ExtensionOnly:     req.ExtensionOnly,
		Amount:            req.Amount,
		ComputationType:   req.ComputationType,
		AccountID:         req.AccountID,
		Account:           account,
		CreatedAt:         now,
		CreatedByID:       userOrg.UserID,
		UpdatedAt:         now,
		UpdatedByID:       userOrg.UserID,
		BranchID:          *userOrg.BranchID,
		OrganizationID:    userOrg.OrganizationID,
		AdditionalMembers: additionalMembers,
		MutualFundTables:  mutualFundTables,
	}

	for _, additionalMember := range additionalMembers {
		additionalMember.MutualFundID = mutualFund.ID
	}

	for _, table := range mutualFundTables {
		table.MutualFundID = mutualFund.ID
	}

	return mutualFund, nil
}
