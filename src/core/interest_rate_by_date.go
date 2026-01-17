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

func InterestRateByDateManager(service *horizon.HorizonService) *registry.Registry[
	types.InterestRateByDate, types.InterestRateByDateResponse, types.InterestRateByDateRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.InterestRateByDate, types.InterestRateByDateResponse, types.InterestRateByDateRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Organization", "Branch", "BrowseReference",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.InterestRateByDate) *types.InterestRateByDateResponse {
			if data == nil {
				return nil
			}
			return &types.InterestRateByDateResponse{
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
				FromDate:          data.FromDate.Format(time.RFC3339),
				ToDate:            data.ToDate.Format(time.RFC3339),
				InterestRate:      data.InterestRate,
			}
		},

		Created: func(data *types.InterestRateByDate) registry.Topics {
			return []string{
				"interest_rate_by_date.create",
				fmt.Sprintf("interest_rate_by_date.create.%s", data.ID),
				fmt.Sprintf("interest_rate_by_date.create.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_date.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_date.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.InterestRateByDate) registry.Topics {
			return []string{
				"interest_rate_by_date.update",
				fmt.Sprintf("interest_rate_by_date.update.%s", data.ID),
				fmt.Sprintf("interest_rate_by_date.update.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_date.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_date.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.InterestRateByDate) registry.Topics {
			return []string{
				"interest_rate_by_date.delete",
				fmt.Sprintf("interest_rate_by_date.delete.%s", data.ID),
				fmt.Sprintf("interest_rate_by_date.delete.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_date.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_date.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func InterestRateByDateForBrowseReference(context context.Context,
	service *horizon.HorizonService, browseReferenceID uuid.UUID) ([]*types.InterestRateByDate, error) {
	filters := []query.ArrFilterSQL{
		{Field: "browse_reference_id", Op: query.ModeEqual, Value: browseReferenceID},
	}

	return InterestRateByDateManager(service).ArrFind(context, filters, nil)
}

func InterestRateByDateForRange(context context.Context, service *horizon.HorizonService,
	browseReferenceID uuid.UUID, date time.Time) ([]*types.InterestRateByDate, error) {
	filters := []query.ArrFilterSQL{
		{Field: "browse_reference_id", Op: query.ModeEqual, Value: browseReferenceID},
		{Field: "from_date", Op: query.ModeLTE, Value: date},
		{Field: "to_date", Op: query.ModeGTE, Value: date},
	}

	return InterestRateByDateManager(service).ArrFind(context, filters, nil)
}

func InterestRateByDateCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.InterestRateByDate, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return InterestRateByDateManager(service).ArrFind(context, filters, nil)
}

func GetInterestRateForDate(context context.Context, service *horizon.HorizonService,
	browseReferenceID uuid.UUID, date time.Time) (*types.InterestRateByDate, error) {
	rates, err := InterestRateByDateForRange(context, service, browseReferenceID, date)
	if err != nil {
		return nil, err
	}

	if len(rates) == 0 {
		return nil, nil
	}

	return rates[0], nil
}
