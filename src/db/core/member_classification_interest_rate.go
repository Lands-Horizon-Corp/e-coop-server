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

func MemberClassificationInterestRateManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberClassificationInterestRate, types.MemberClassificationInterestRateResponse, types.MemberClassificationInterestRateRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberClassificationInterestRate, types.MemberClassificationInterestRateResponse, types.MemberClassificationInterestRateRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"InterestRateScheme", "MemberClassification", "InterestRateByTermsHeader",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberClassificationInterestRate) *types.MemberClassificationInterestRateResponse {
			if data == nil {
				return nil
			}
			return &types.MemberClassificationInterestRateResponse{
				ID:                     data.ID,
				CreatedAt:              data.CreatedAt.Format(time.RFC3339),
				CreatedByID:            data.CreatedByID,
				CreatedBy:              UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:              data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:            data.UpdatedByID,
				UpdatedBy:              UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:         data.OrganizationID,
				Organization:           OrganizationManager(service).ToModel(data.Organization),
				BranchID:               data.BranchID,
				Branch:                 BranchManager(service).ToModel(data.Branch),
				Name:                   data.Name,
				Description:            data.Description,
				InterestRateSchemeID:   data.InterestRateSchemeID,
				InterestRateScheme:     InterestRateSchemeManager(service).ToModel(data.InterestRateScheme),
				MemberClassificationID: data.MemberClassificationID,
				MemberClassification:   MemberClassificationManager(service).ToModel(data.MemberClassification),

				Header1:  data.Header1,
				Header2:  data.Header2,
				Header3:  data.Header3,
				Header4:  data.Header4,
				Header5:  data.Header5,
				Header6:  data.Header6,
				Header7:  data.Header7,
				Header8:  data.Header8,
				Header9:  data.Header9,
				Header10: data.Header10,
				Header11: data.Header11,
				Header12: data.Header12,
				Header13: data.Header13,
				Header14: data.Header14,
				Header15: data.Header15,
				Header16: data.Header16,
				Header17: data.Header17,
				Header18: data.Header18,
				Header19: data.Header19,
				Header20: data.Header20,
				Header21: data.Header21,
				Header22: data.Header22,
			}
		},
		Created: func(data *types.MemberClassificationInterestRate) registry.Topics {
			return []string{
				"member_classification_interest_rate.create",
				fmt.Sprintf("member_classification_interest_rate.create.%s", data.ID),
				fmt.Sprintf("member_classification_interest_rate.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_interest_rate.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberClassificationInterestRate) registry.Topics {
			return []string{
				"member_classification_interest_rate.update",
				fmt.Sprintf("member_classification_interest_rate.update.%s", data.ID),
				fmt.Sprintf("member_classification_interest_rate.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_interest_rate.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberClassificationInterestRate) registry.Topics {
			return []string{
				"member_classification_interest_rate.delete",
				fmt.Sprintf("member_classification_interest_rate.delete.%s", data.ID),
				fmt.Sprintf("member_classification_interest_rate.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_interest_rate.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberClassificationInterestRateCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberClassificationInterestRate, error) {
	return MemberClassificationInterestRateManager(service).Find(context, &types.MemberClassificationInterestRate{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
