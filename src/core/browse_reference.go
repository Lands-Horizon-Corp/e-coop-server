package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

func BrowseReferenceManager(service *horizon.HorizonService) *registry.Registry[
	types.BrowseReference, types.BrowseReferenceResponse, types.BrowseReferenceRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.BrowseReference, types.BrowseReferenceResponse, types.BrowseReferenceRequest,
	]{
		Preloads: []string{
			"Account", "Account.Currency",
			"MemberType", "InterestRatesByYear", "InterestRatesByDate", "InterestRatesByAmount",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.BrowseReference) *types.BrowseReferenceResponse {
			if data == nil {
				return nil
			}
			return &types.BrowseReferenceResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),

				Name:                  data.Name,
				Description:           data.Description,
				InterestRate:          data.InterestRate,
				MinimumBalance:        data.MinimumBalance,
				Charges:               data.Charges,
				AccountID:             data.AccountID,
				Account:               AccountManager(service).ToModel(data.Account),
				MemberTypeID:          data.MemberTypeID,
				MemberType:            MemberTypeManager(service).ToModel(data.MemberType),
				InterestType:          data.InterestType,
				DefaultMinimumBalance: data.DefaultMinimumBalance,
				DefaultInterestRate:   data.DefaultInterestRate,
				InterestRatesByYear:   InterestRateByYearManager(service).ToModels(data.InterestRatesByYear),
				InterestRatesByDate:   InterestRateByDateManager(service).ToModels(data.InterestRatesByDate),
				InterestRatesByAmount: InterestRateByAmountManager(service).ToModels(data.InterestRatesByAmount),
			}
		},

		Created: func(data *types.BrowseReference) registry.Topics {
			return []string{
				"browse_reference.create",
				fmt.Sprintf("browse_reference.create.%s", data.ID),
				fmt.Sprintf("browse_reference.create.branch.%s", data.BranchID),
				fmt.Sprintf("browse_reference.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.BrowseReference) registry.Topics {
			return []string{
				"browse_reference.update",
				fmt.Sprintf("browse_reference.update.%s", data.ID),
				fmt.Sprintf("browse_reference.update.branch.%s", data.BranchID),
				fmt.Sprintf("browse_reference.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.BrowseReference) registry.Topics {
			return []string{
				"browse_reference.delete",
				fmt.Sprintf("browse_reference.delete.%s", data.ID),
				fmt.Sprintf("browse_reference.delete.branch.%s", data.BranchID),
				fmt.Sprintf("browse_reference.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func BrowseReferenceCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID,
	branchID uuid.UUID) ([]*types.BrowseReference, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return BrowseReferenceManager(service).ArrFind(context, filters, nil)
}

func BrowseReferenceByMemberType(context context.Context, service *horizon.HorizonService, memberTypeID,
	organizationID, branchID uuid.UUID) ([]*types.BrowseReference, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "member_type_id", Op: query.ModeEqual, Value: memberTypeID},
	}

	return BrowseReferenceManager(service).ArrFind(context, filters, nil)
}

func BrowseReferenceByInterestType(context context.Context, service *horizon.HorizonService,
	interestType types.InterestType, organizationID, branchID uuid.UUID) ([]*types.BrowseReference, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "interest_type", Op: query.ModeEqual, Value: string(interestType)},
	}

	return BrowseReferenceManager(service).ArrFind(context, filters, nil)
}

func BrowseReferenceByField(
	context context.Context, service *horizon.HorizonService, organizationID, branchID uuid.UUID, accountID, memberTypeID *uuid.UUID,
) ([]*types.BrowseReference, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	if memberTypeID != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "member_type_id", Op: query.ModeEqual, Value: *memberTypeID,
		})
	}

	if accountID != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "account_id", Op: query.ModeEqual, Value: *accountID,
		})
	}

	return BrowseReferenceManager(service).ArrFind(context, filters, nil)
}
