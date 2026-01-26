package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

func GeneratedSavingsInterestManager(service *horizon.HorizonService) *registry.Registry[
	types.GeneratedSavingsInterest, types.GeneratedSavingsInterestResponse, types.GeneratedSavingsInterestRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.GeneratedSavingsInterest, types.GeneratedSavingsInterestResponse, types.GeneratedSavingsInterestRequest,
	]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
			"Account",
			"Account.Currency",
			"MemberType",
			"PrintedByUser", "PostedByUser", "PostAccount",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.GeneratedSavingsInterest) *types.GeneratedSavingsInterestResponse {
			if data == nil {
				return nil
			}

			var postedDate *string
			if data.PostedDate != nil {
				formatted := data.PostedDate.Format(time.RFC3339)
				postedDate = &formatted
			}

			var printedDate *string
			if data.PrintedDate != nil {
				formatted := data.PrintedDate.Format(time.RFC3339)
				printedDate = &formatted
			}

			return &types.GeneratedSavingsInterestResponse{
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

				DocumentNo:                      data.DocumentNo,
				LastComputationDate:             data.LastComputationDate.Format(time.RFC3339),
				NewComputationDate:              data.NewComputationDate.Format(time.RFC3339),
				AccountID:                       data.AccountID,
				Account:                         AccountManager(service).ToModel(data.Account),
				MemberTypeID:                    data.MemberTypeID,
				MemberType:                      MemberTypeManager(service).ToModel(data.MemberType),
				SavingsComputationType:          data.SavingsComputationType,
				IncludeClosedAccount:            data.IncludeClosedAccount,
				IncludeExistingComputedInterest: data.IncludeExistingComputedInterest,
				InterestTaxRate:                 data.InterestTaxRate,
				TotalInterest:                   data.TotalInterest,
				TotalTax:                        data.TotalTax,
				PrintedByUserID:                 data.PrintedByUserID,
				PrintedByUser:                   UserManager(service).ToModel(data.PrintedByUser),
				PrintedDate:                     printedDate,
				PostedByUserID:                  data.PostedByUserID,
				PostedByUser:                    UserManager(service).ToModel(data.PostedByUser),
				PostedDate:                      postedDate,
				CheckVoucherNumber:              data.CheckVoucherNumber,
				PostAccountID:                   data.PostAccountID,
				PostAccount:                     AccountManager(service).ToModel(data.PostAccount),
				Entries:                         GeneratedSavingsInterestEntryManager(service).ToModels(data.Entries),
			}
		},

		Created: func(data *types.GeneratedSavingsInterest) registry.Topics {
			return []string{
				"generated_savings_interest.create",
				fmt.Sprintf("generated_savings_interest.create.%s", data.ID),
				fmt.Sprintf("generated_savings_interest.create.branch.%s", data.BranchID),
				fmt.Sprintf("generated_savings_interest.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.GeneratedSavingsInterest) registry.Topics {
			return []string{
				"generated_savings_interest.update",
				fmt.Sprintf("generated_savings_interest.update.%s", data.ID),
				fmt.Sprintf("generated_savings_interest.update.branch.%s", data.BranchID),
				fmt.Sprintf("generated_savings_interest.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.GeneratedSavingsInterest) registry.Topics {
			return []string{
				"generated_savings_interest.delete",
				fmt.Sprintf("generated_savings_interest.delete.%s", data.ID),
				fmt.Sprintf("generated_savings_interest.delete.branch.%s", data.BranchID),
				fmt.Sprintf("generated_savings_interest.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}
