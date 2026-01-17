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

func ChargesRateSchemeAccountManager(service *horizon.HorizonService) *registry.Registry[
	types.ChargesRateSchemeAccount, types.ChargesRateSchemeAccountResponse, types.ChargesRateSchemeAccountRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.ChargesRateSchemeAccount, types.ChargesRateSchemeAccountResponse, types.ChargesRateSchemeAccountRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "ChargesRateScheme", "Account"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.ChargesRateSchemeAccount) *types.ChargesRateSchemeAccountResponse {
			if data == nil {
				return nil
			}
			return &types.ChargesRateSchemeAccountResponse{
				ID:                  data.ID,
				CreatedAt:           data.CreatedAt.Format(time.RFC3339),
				CreatedByID:         data.CreatedByID,
				CreatedBy:           UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:           data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:         data.UpdatedByID,
				UpdatedBy:           UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:      data.OrganizationID,
				Organization:        OrganizationManager(service).ToModel(data.Organization),
				BranchID:            data.BranchID,
				Branch:              BranchManager(service).ToModel(data.Branch),
				ChargesRateSchemeID: data.ChargesRateSchemeID,
				ChargesRateScheme:   ChargesRateSchemeManager(service).ToModel(data.ChargesRateScheme),
				AccountID:           data.AccountID,
				Account:             AccountManager(service).ToModel(data.Account),
			}
		},
		Created: func(data *types.ChargesRateSchemeAccount) registry.Topics {
			return []string{
				"charges_rate_scheme_account.create",
				fmt.Sprintf("charges_rate_scheme_account.create.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme_account.create.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme_account.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.ChargesRateSchemeAccount) registry.Topics {
			return []string{
				"charges_rate_scheme_account.update",
				fmt.Sprintf("charges_rate_scheme_account.update.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme_account.update.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme_account.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.ChargesRateSchemeAccount) registry.Topics {
			return []string{
				"charges_rate_scheme_account.delete",
				fmt.Sprintf("charges_rate_scheme_account.delete.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme_account.delete.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme_account.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func ChargesRateSchemeAccountCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID,
	branchID uuid.UUID) ([]*types.ChargesRateSchemeAccount, error) {
	return ChargesRateSchemeAccountManager(service).Find(context, &types.ChargesRateSchemeAccount{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
