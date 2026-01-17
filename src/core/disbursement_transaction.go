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

func DisbursementTransactionManager(service *horizon.HorizonService) *registry.Registry[
	types.DisbursementTransaction, types.DisbursementTransactionResponse, types.DisbursementTransactionRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.DisbursementTransaction, types.DisbursementTransactionResponse, types.DisbursementTransactionRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"Disbursement", "TransactionBatch", "EmployeeUser",
			"Disbursement.Currency",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.DisbursementTransaction) *types.DisbursementTransactionResponse {
			if data == nil {
				return nil
			}
			return &types.DisbursementTransactionResponse{
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
				DisbursementID:     data.DisbursementID,
				Disbursement:       DisbursementManager(service).ToModel(data.Disbursement),
				TransactionBatchID: data.TransactionBatchID,
				TransactionBatch:   TransactionBatchManager(service).ToModel(data.TransactionBatch),
				EmployeeUserID:     data.EmployeeUserID,
				EmployeeUser:       UserManager(service).ToModel(data.EmployeeUser),
				ReferenceNumber:    data.ReferenceNumber,
				Amount:             data.Amount,
			}
		},
		Created: func(data *types.DisbursementTransaction) registry.Topics {
			return []string{
				"disbursement_transaction.create",
				fmt.Sprintf("disbursement_transaction.create.%s", data.ID),
				fmt.Sprintf("disbursement_transaction.create.branch.%s", data.BranchID),
				fmt.Sprintf("disbursement_transaction.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.DisbursementTransaction) registry.Topics {
			return []string{
				"disbursement_transaction.update",
				fmt.Sprintf("disbursement_transaction.update.%s", data.ID),
				fmt.Sprintf("disbursement_transaction.update.branch.%s", data.BranchID),
				fmt.Sprintf("disbursement_transaction.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.DisbursementTransaction) registry.Topics {
			return []string{
				"disbursement_transaction.delete",
				fmt.Sprintf("disbursement_transaction.delete.%s", data.ID),
				fmt.Sprintf("disbursement_transaction.delete.branch.%s", data.BranchID),
				fmt.Sprintf("disbursement_transaction.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func DisbursementTransactionCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.DisbursementTransaction, error) {
	return DisbursementTransactionManager(service).Find(context, &types.DisbursementTransaction{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
