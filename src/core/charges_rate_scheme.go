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

func ChargesRateSchemeManager(service *horizon.HorizonService) *registry.Registry[
	types.ChargesRateScheme, types.ChargesRateSchemeResponse, types.ChargesRateSchemeRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.ChargesRateScheme, types.ChargesRateSchemeResponse, types.ChargesRateSchemeRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"Currency",

			"MemberType",
			"ChargesRateSchemeAccounts",
			"ChargesRateByRangeOrMinimumAmounts",
			"ChargesRateSchemeModeOfPayments",
			"ChargesRateByTerms",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.ChargesRateScheme) *types.ChargesRateSchemeResponse {
			if data == nil {
				return nil
			}
			return &types.ChargesRateSchemeResponse{
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
				CurrencyID:     data.CurrencyID,
				Currency:       CurrencyManager(service).ToModel(data.Currency),
				Name:           data.Name,
				Description:    data.Description,
				Icon:           data.Icon,
				Type:           data.Type,

				ChargesRateSchemeAccounts: ChargesRateSchemeAccountManager(service).ToModels(data.ChargesRateSchemeAccounts),

				ChargesRateByRangeOrMinimumAmounts: ChargesRateByRangeOrMinimumAmountManager(service).ToModels(data.ChargesRateByRangeOrMinimumAmounts),

				ChargesRateSchemeModeOfPayments: ChargesRateSchemeModeOfPaymentManager(service).ToModels(data.ChargesRateSchemeModeOfPayments),

				ChargesRateByTerms: ChargesRateByTermManager(service).ToModels(data.ChargesRateByTerms),

				MemberTypeID:  data.MemberTypeID,
				MemberType:    MemberTypeManager(service).ToModel(data.MemberType),
				ModeOfPayment: data.ModeOfPayment,

				ModeOfPaymentHeader1:  data.ModeOfPaymentHeader1,
				ModeOfPaymentHeader2:  data.ModeOfPaymentHeader2,
				ModeOfPaymentHeader3:  data.ModeOfPaymentHeader3,
				ModeOfPaymentHeader4:  data.ModeOfPaymentHeader4,
				ModeOfPaymentHeader5:  data.ModeOfPaymentHeader5,
				ModeOfPaymentHeader6:  data.ModeOfPaymentHeader6,
				ModeOfPaymentHeader7:  data.ModeOfPaymentHeader7,
				ModeOfPaymentHeader8:  data.ModeOfPaymentHeader8,
				ModeOfPaymentHeader9:  data.ModeOfPaymentHeader9,
				ModeOfPaymentHeader10: data.ModeOfPaymentHeader10,
				ModeOfPaymentHeader11: data.ModeOfPaymentHeader11,
				ModeOfPaymentHeader12: data.ModeOfPaymentHeader12,
				ModeOfPaymentHeader13: data.ModeOfPaymentHeader13,
				ModeOfPaymentHeader14: data.ModeOfPaymentHeader14,
				ModeOfPaymentHeader15: data.ModeOfPaymentHeader15,
				ModeOfPaymentHeader16: data.ModeOfPaymentHeader16,
				ModeOfPaymentHeader17: data.ModeOfPaymentHeader17,
				ModeOfPaymentHeader18: data.ModeOfPaymentHeader18,
				ModeOfPaymentHeader19: data.ModeOfPaymentHeader19,
				ModeOfPaymentHeader20: data.ModeOfPaymentHeader20,
				ModeOfPaymentHeader21: data.ModeOfPaymentHeader21,
				ModeOfPaymentHeader22: data.ModeOfPaymentHeader22,
				ByTermHeader1:         data.ByTermHeader1,
				ByTermHeader2:         data.ByTermHeader2,
				ByTermHeader3:         data.ByTermHeader3,
				ByTermHeader4:         data.ByTermHeader4,
				ByTermHeader5:         data.ByTermHeader5,
				ByTermHeader6:         data.ByTermHeader6,
				ByTermHeader7:         data.ByTermHeader7,
				ByTermHeader8:         data.ByTermHeader8,
				ByTermHeader9:         data.ByTermHeader9,
				ByTermHeader10:        data.ByTermHeader10,
				ByTermHeader11:        data.ByTermHeader11,
				ByTermHeader12:        data.ByTermHeader12,
				ByTermHeader13:        data.ByTermHeader13,
				ByTermHeader14:        data.ByTermHeader14,
				ByTermHeader15:        data.ByTermHeader15,
				ByTermHeader16:        data.ByTermHeader16,
				ByTermHeader17:        data.ByTermHeader17,
				ByTermHeader18:        data.ByTermHeader18,
				ByTermHeader19:        data.ByTermHeader19,
				ByTermHeader20:        data.ByTermHeader20,
				ByTermHeader21:        data.ByTermHeader21,
				ByTermHeader22:        data.ByTermHeader22,
			}
		},
		Created: func(data *types.ChargesRateScheme) registry.Topics {
			return []string{
				"charges_rate_scheme.create",
				fmt.Sprintf("charges_rate_scheme.create.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme.create.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.ChargesRateScheme) registry.Topics {
			return []string{
				"charges_rate_scheme.update",
				fmt.Sprintf("charges_rate_scheme.update.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme.update.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.ChargesRateScheme) registry.Topics {
			return []string{
				"charges_rate_scheme.delete",
				fmt.Sprintf("charges_rate_scheme.delete.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme.delete.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func ChargesRateSchemeCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.ChargesRateScheme, error) {
	return ChargesRateSchemeManager(service).Find(context, &types.ChargesRateScheme{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
