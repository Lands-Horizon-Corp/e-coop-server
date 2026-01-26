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

func InterestRateByAmountManager(service *horizon.HorizonService) *registry.Registry[
	types.InterestRateByAmount, types.InterestRateByAmountResponse, types.InterestRateByAmountRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.InterestRateByAmount, types.InterestRateByAmountResponse, types.InterestRateByAmountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Organization", "Branch", "BrowseReference",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.InterestRateByAmount) *types.InterestRateByAmountResponse {
			if data == nil {
				return nil
			}
			return &types.InterestRateByAmountResponse{
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

				BrowseReferenceID: data.BrowseReferenceID,
				BrowseReference:   BrowseReferenceManager(service).ToModel(data.BrowseReference),
				FromAmount:        data.FromAmount,
				ToAmount:          data.ToAmount,
				InterestRate:      data.InterestRate,
			}
		},

		Created: func(data *types.InterestRateByAmount) registry.Topics {
			return []string{
				"interest_rate_by_amount.create",
				fmt.Sprintf("interest_rate_by_amount.create.%s", data.ID),
				fmt.Sprintf("interest_rate_by_amount.create.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_amount.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_amount.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.InterestRateByAmount) registry.Topics {
			return []string{
				"interest_rate_by_amount.update",
				fmt.Sprintf("interest_rate_by_amount.update.%s", data.ID),
				fmt.Sprintf("interest_rate_by_amount.update.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_amount.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_amount.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.InterestRateByAmount) registry.Topics {
			return []string{
				"interest_rate_by_amount.delete",
				fmt.Sprintf("interest_rate_by_amount.delete.%s", data.ID),
				fmt.Sprintf("interest_rate_by_amount.delete.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_amount.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_amount.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func InterestRateByAmountForBrowseReference(context context.Context, service *horizon.HorizonService, browseReferenceID uuid.UUID) ([]*types.InterestRateByAmount, error) {
	filters := []query.ArrFilterSQL{
		{Field: "browse_reference_id", Op: query.ModeEqual, Value: browseReferenceID},
	}

	return InterestRateByAmountManager(service).ArrFind(context, filters, nil)
}

func InterestRateByAmountForRange(context context.Context, service *horizon.HorizonService, browseReferenceID uuid.UUID, amount float64) ([]*types.InterestRateByAmount, error) {
	filters := []query.ArrFilterSQL{
		{Field: "browse_reference_id", Op: query.ModeEqual, Value: browseReferenceID},
		{Field: "from_amount", Op: query.ModeLTE, Value: amount},
		{Field: "to_amount", Op: query.ModeGTE, Value: amount},
	}

	return InterestRateByAmountManager(service).ArrFind(context, filters, nil)
}

func InterestRateByAmountCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.InterestRateByAmount, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return InterestRateByAmountManager(service).ArrFind(context, filters, nil)
}

func GetInterestRateForAmount(context context.Context, service *horizon.HorizonService, browseReferenceID uuid.UUID, amount float64) (*types.InterestRateByAmount, error) {
	rates, err := InterestRateByAmountForRange(context, service, browseReferenceID, amount)
	if err != nil {
		return nil, err
	}

	if len(rates) == 0 {
		return nil, nil
	}

	return rates[0], nil
}
