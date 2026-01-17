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

func ChargesRateSchemeModeOfPaymentManager(service *horizon.HorizonService) *registry.Registry[
	types.ChargesRateSchemeModeOfPayment, types.ChargesRateSchemeModeOfPaymentResponse, types.ChargesRateSchemeModeOfPaymentRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.ChargesRateSchemeModeOfPayment, types.ChargesRateSchemeModeOfPaymentResponse, types.ChargesRateSchemeModeOfPaymentRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "ChargesRateScheme",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.ChargesRateSchemeModeOfPayment) *types.ChargesRateSchemeModeOfPaymentResponse {
			if data == nil {
				return nil
			}
			return &types.ChargesRateSchemeModeOfPaymentResponse{
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
				From:                data.From,
				To:                  data.To,
				Column1:             data.Column1,
				Column2:             data.Column2,
				Column3:             data.Column3,
				Column4:             data.Column4,
				Column5:             data.Column5,
				Column6:             data.Column6,
				Column7:             data.Column7,
				Column8:             data.Column8,
				Column9:             data.Column9,
				Column10:            data.Column10,
				Column11:            data.Column11,
				Column12:            data.Column12,
				Column13:            data.Column13,
				Column14:            data.Column14,
				Column15:            data.Column15,
				Column16:            data.Column16,
				Column17:            data.Column17,
				Column18:            data.Column18,
				Column19:            data.Column19,
				Column20:            data.Column20,
				Column21:            data.Column21,
				Column22:            data.Column22,
			}
		},
		Created: func(data *types.ChargesRateSchemeModeOfPayment) registry.Topics {
			return []string{
				"charges_rate_scheme_model_of_payment.create",
				fmt.Sprintf("charges_rate_scheme_model_of_payment.create.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme_model_of_payment.create.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme_model_of_payment.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.ChargesRateSchemeModeOfPayment) registry.Topics {
			return []string{
				"charges_rate_scheme_model_of_payment.update",
				fmt.Sprintf("charges_rate_scheme_model_of_payment.update.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme_model_of_payment.update.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme_model_of_payment.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.ChargesRateSchemeModeOfPayment) registry.Topics {
			return []string{
				"charges_rate_scheme_model_of_payment.delete",
				fmt.Sprintf("charges_rate_scheme_model_of_payment.delete.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme_model_of_payment.delete.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme_model_of_payment.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func ChargesRateSchemeModeOfPaymentCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.ChargesRateSchemeModeOfPayment, error) {
	return ChargesRateSchemeModeOfPaymentManager(service).Find(context, &types.ChargesRateSchemeModeOfPayment{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
