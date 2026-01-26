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

func AdjustmentEntryManager(service *horizon.HorizonService) *registry.Registry[
	types.AdjustmentEntry, types.AdjustmentEntryResponse, types.AdjustmentEntryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.AdjustmentEntry, types.AdjustmentEntryResponse, types.AdjustmentEntryRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"TransactionBatch",
			"SignatureMedia",
			"Account",
			"MemberProfile",
			"EmployeeUser",
			"PaymentType",
			"AdjustmentTags",
			"LoanTransaction",
			"Account.Currency",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.AdjustmentEntry) *types.AdjustmentEntryResponse {
			if data == nil {
				return nil
			}
			var entryDateStr *string
			if data.EntryDate != nil {
				str := data.EntryDate.Format("2006-01-02")
				entryDateStr = &str
			}
			return &types.AdjustmentEntryResponse{
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
				SignatureMediaID:   data.SignatureMediaID,
				SignatureMedia:     MediaManager(service).ToModel(data.SignatureMedia),
				AccountID:          data.AccountID,
				Account:            AccountManager(service).ToModel(data.Account),
				MemberProfileID:    data.MemberProfileID,
				MemberProfile:      MemberProfileManager(service).ToModel(data.MemberProfile),
				EmployeeUserID:     data.EmployeeUserID,
				EmployeeUser:       UserManager(service).ToModel(data.EmployeeUser),
				PaymentTypeID:      data.PaymentTypeID,
				PaymentType:        PaymentTypeManager(service).ToModel(data.PaymentType),
				TypeOfPaymentType:  data.TypeOfPaymentType,
				Description:        data.Description,
				ReferenceNumber:    data.ReferenceNumber,
				EntryDate:          entryDateStr,
				Debit:              data.Debit,
				Credit:             data.Credit,
				AdjustmentTags:     AdjustmentTagManager(service).ToModels(data.AdjustmentTags),
				LoanTransactionID:  data.LoanTransactionID,
				LoanTransaction:    LoanTransactionManager(service).ToModel(data.LoanTransaction),
			}
		},
		Created: func(data *types.AdjustmentEntry) registry.Topics {
			return []string{
				"adjustment_entry.create",
				fmt.Sprintf("adjustment_entry.create.%s", data.ID),
				fmt.Sprintf("adjustment_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("adjustment_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.AdjustmentEntry) registry.Topics {
			return []string{
				"adjustment_entry.update",
				fmt.Sprintf("adjustment_entry.update.%s", data.ID),
				fmt.Sprintf("adjustment_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("adjustment_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.AdjustmentEntry) registry.Topics {
			return []string{
				"adjustment_entry.delete",
				fmt.Sprintf("adjustment_entry.delete.%s", data.ID),
				fmt.Sprintf("adjustment_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("adjustment_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func AdjustmentEntryCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.AdjustmentEntry, error) {
	return AdjustmentEntryManager(service).Find(context, &types.AdjustmentEntry{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
