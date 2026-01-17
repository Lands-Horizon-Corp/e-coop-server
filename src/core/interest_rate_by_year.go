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

func InterestRateByYearManager(service *horizon.HorizonService) *registry.Registry[
	types.InterestRateByYear, types.InterestRateByYearResponse, types.InterestRateByYearRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.InterestRateByYear, types.InterestRateByYearResponse, types.InterestRateByYearRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Organization", "Branch", "BrowseReference",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.InterestRateByYear) *types.InterestRateByYearResponse {
			if data == nil {
				return nil
			}
			return &types.InterestRateByYearResponse{
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
				FromYear:          data.FromYear,
				ToYear:            data.ToYear,
				InterestRate:      data.InterestRate,
			}
		},

		Created: func(data *types.InterestRateByYear) registry.Topics {
			return []string{
				"interest_rate_by_year.create",
				fmt.Sprintf("interest_rate_by_year.create.%s", data.ID),
				fmt.Sprintf("interest_rate_by_year.create.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_year.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_year.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.InterestRateByYear) registry.Topics {
			return []string{
				"interest_rate_by_year.update",
				fmt.Sprintf("interest_rate_by_year.update.%s", data.ID),
				fmt.Sprintf("interest_rate_by_year.update.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_year.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_year.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.InterestRateByYear) registry.Topics {
			return []string{
				"interest_rate_by_year.delete",
				fmt.Sprintf("interest_rate_by_year.delete.%s", data.ID),
				fmt.Sprintf("interest_rate_by_year.delete.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_year.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_year.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func InterestRateByYearForBrowseReference(context context.Context,
	service *horizon.HorizonService, browseReferenceID uuid.UUID) ([]*types.InterestRateByYear, error) {
	filters := []query.ArrFilterSQL{
		{Field: "browse_reference_id", Op: query.ModeEqual, Value: browseReferenceID},
	}

	return InterestRateByYearManager(service).ArrFind(context, filters, nil)
}

func InterestRateByYearForRange(context context.Context, service *horizon.HorizonService,
	browseReferenceID uuid.UUID, year int) ([]*types.InterestRateByYear, error) {
	filters := []query.ArrFilterSQL{
		{Field: "browse_reference_id", Op: query.ModeEqual, Value: browseReferenceID},
		{Field: "from_year", Op: query.ModeLTE, Value: year},
		{Field: "to_year", Op: query.ModeGTE, Value: year},
	}

	return InterestRateByYearManager(service).ArrFind(context, filters, nil)
}

func InterestRateByYearCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.InterestRateByYear, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return InterestRateByYearManager(service).ArrFind(context, filters, nil)
}

func GetInterestRateForYear(context context.Context, service *horizon.HorizonService,
	browseReferenceID uuid.UUID, year int) (*types.InterestRateByYear, error) {
	rates, err := InterestRateByYearForRange(context, service, browseReferenceID, year)
	if err != nil {
		return nil, err
	}

	if len(rates) == 0 {
		return nil, nil
	}

	return rates[0], nil
}
