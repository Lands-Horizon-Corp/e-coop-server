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

func BatchFundingManager(service *horizon.HorizonService) *registry.Registry[types.BatchFunding, types.BatchFundingResponse, types.BatchFundingRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.BatchFunding, types.BatchFundingResponse, types.BatchFundingRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"TransactionBatch", "ProvidedByUser", "SignatureMedia", "Currency",
			"ProvidedByUser.Media",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.BatchFunding) *types.BatchFundingResponse {
			if data == nil {
				return nil
			}
			return &types.BatchFundingResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       OrganizationManager(service).ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             BranchManager(service).ToModel(data.Branch),
				TransactionBatchID: data.TransactionBatchID,
				TransactionBatch:   TransactionBatchManager(service).ToModel(data.TransactionBatch),
				ProvidedByUserID:   data.ProvidedByUserID,
				ProvidedByUser:     UserManager(service).ToModel(data.ProvidedByUser),
				SignatureMediaID:   data.SignatureMediaID,
				SignatureMedia:     MediaManager(service).ToModel(data.SignatureMedia),
				CurrencyID:         data.CurrencyID,
				Currency:           CurrencyManager(service).ToModel(data.Currency),
				Name:               data.Name,
				Amount:             data.Amount,
				Description:        data.Description,
			}
		},
		Created: func(data *types.BatchFunding) registry.Topics {
			return []string{
				"batch_funding.create",
				fmt.Sprintf("batch_funding.create.%s", data.ID),
				fmt.Sprintf("batch_funding.create.branch.%s", data.BranchID),
				fmt.Sprintf("batch_funding.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.BatchFunding) registry.Topics {
			return []string{
				"batch_funding.update",
				fmt.Sprintf("batch_funding.update.%s", data.ID),
				fmt.Sprintf("batch_funding.update.branch.%s", data.BranchID),
				fmt.Sprintf("batch_funding.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.BatchFunding) registry.Topics {
			return []string{
				"batch_funding.delete",
				fmt.Sprintf("batch_funding.delete.%s", data.ID),
				fmt.Sprintf("batch_funding.delete.branch.%s", data.BranchID),
				fmt.Sprintf("batch_funding.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func BatchFundingCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID,
	branchID uuid.UUID) ([]*types.BatchFunding, error) {
	return BatchFundingManager(service).Find(context, &types.BatchFunding{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
