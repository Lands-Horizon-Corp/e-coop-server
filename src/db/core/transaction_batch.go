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

func TransactionBatchManager(service *horizon.HorizonService) *registry.Registry[types.TransactionBatch, types.TransactionBatchResponse, types.TransactionBatchRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.TransactionBatch, types.TransactionBatchResponse, types.TransactionBatchRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"EmployeeUser",
			"EmployeeUser.Media",
			"Currency",
			"UnbalancedAccount",
			"UnbalancedAccount.AccountForShortage",
			"UnbalancedAccount.AccountForOverage",
			"UnbalancedAccount.CashOnHandAccount",
			"ApprovedBySignatureMedia",
			"PreparedBySignatureMedia",
			"CertifiedBySignatureMedia",
			"VerifiedBySignatureMedia",
			"CheckBySignatureMedia",
			"AcknowledgeBySignatureMedia",
			"NotedBySignatureMedia",
			"PostedBySignatureMedia",
			"PaidBySignatureMedia",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.TransactionBatch) *types.TransactionBatchResponse {
			if data == nil {
				return nil
			}

			var endedAt *string
			if data.EndedAt != nil {
				s := data.EndedAt.Format(time.RFC3339)
				endedAt = &s
			}
			isToday := CheckIsToday(
				service,
				time.Now().UTC(),
				data.OrganizationID,
				data.BranchID,
				*data.EmployeeUserID,
			)
			return &types.TransactionBatchResponse{
				IsToday:                       isToday,
				ID:                            data.ID,
				CreatedAt:                     data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                   data.CreatedByID,
				CreatedBy:                     UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:                     data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                   data.UpdatedByID,
				UpdatedBy:                     UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:                data.OrganizationID,
				Organization:                  OrganizationManager(service).ToModel(data.Organization),
				BranchID:                      data.BranchID,
				Branch:                        BranchManager(service).ToModel(data.Branch),
				EmployeeUserID:                data.EmployeeUserID,
				EmployeeUser:                  UserManager(service).ToModel(data.EmployeeUser),
				BatchName:                     data.BatchName,
				TotalCashCollection:           data.TotalCashCollection,
				TotalDepositEntry:             data.TotalDepositEntry,
				BeginningBalance:              data.BeginningBalance,
				DepositInBank:                 data.DepositInBank,
				CashCountTotal:                data.CashCountTotal,
				GrandTotal:                    data.GrandTotal,
				PettyCash:                     data.PettyCash,
				LoanReleases:                  data.LoanReleases,
				CashCheckVoucherTotal:         data.CashCheckVoucherTotal,
				TimeDepositWithdrawal:         data.TimeDepositWithdrawal,
				SavingsWithdrawal:             data.SavingsWithdrawal,
				TotalCashHandled:              data.TotalCashHandled,
				TotalSupposedRemmitance:       data.TotalSupposedRemmitance,
				TotalCashOnHand:               data.TotalCashOnHand,
				TotalCheckRemittance:          data.TotalCheckRemittance,
				TotalOnlineRemittance:         data.TotalOnlineRemittance,
				TotalDepositInBank:            data.TotalDepositInBank,
				TotalActualRemittance:         data.TotalActualRemittance,
				TotalActualSupposedComparison: data.TotalActualSupposedComparison,
				Description:                   data.Description,
				CanView:                       data.CanView,
				IsClosed:                      data.IsClosed,
				RequestView:                   data.RequestView,

				EmployeeBySignatureMediaID:    data.EmployeeBySignatureMediaID,
				EmployeeBySignatureMedia:      MediaManager(service).ToModel(data.EmployeeBySignatureMedia),
				EmployeeByName:                data.EmployeeByName,
				EmployeeByPosition:            data.EmployeeByPosition,
				ApprovedBySignatureMediaID:    data.ApprovedBySignatureMediaID,
				ApprovedBySignatureMedia:      MediaManager(service).ToModel(data.ApprovedBySignatureMedia),
				ApprovedByName:                data.ApprovedByName,
				ApprovedByPosition:            data.ApprovedByPosition,
				PreparedBySignatureMediaID:    data.PreparedBySignatureMediaID,
				PreparedBySignatureMedia:      MediaManager(service).ToModel(data.PreparedBySignatureMedia),
				PreparedByName:                data.PreparedByName,
				PreparedByPosition:            data.PreparedByPosition,
				CertifiedBySignatureMediaID:   data.CertifiedBySignatureMediaID,
				CertifiedBySignatureMedia:     MediaManager(service).ToModel(data.CertifiedBySignatureMedia),
				CertifiedByName:               data.CertifiedByName,
				CertifiedByPosition:           data.CertifiedByPosition,
				VerifiedBySignatureMediaID:    data.VerifiedBySignatureMediaID,
				VerifiedBySignatureMedia:      MediaManager(service).ToModel(data.VerifiedBySignatureMedia),
				VerifiedByName:                data.VerifiedByName,
				VerifiedByPosition:            data.VerifiedByPosition,
				CheckBySignatureMediaID:       data.CheckBySignatureMediaID,
				CheckBySignatureMedia:         MediaManager(service).ToModel(data.CheckBySignatureMedia),
				CheckByName:                   data.CheckByName,
				CheckByPosition:               data.CheckByPosition,
				AcknowledgeBySignatureMediaID: data.AcknowledgeBySignatureMediaID,
				AcknowledgeBySignatureMedia:   MediaManager(service).ToModel(data.AcknowledgeBySignatureMedia),
				AcknowledgeByName:             data.AcknowledgeByName,
				AcknowledgeByPosition:         data.AcknowledgeByPosition,
				NotedBySignatureMediaID:       data.NotedBySignatureMediaID,
				NotedBySignatureMedia:         MediaManager(service).ToModel(data.NotedBySignatureMedia),
				NotedByName:                   data.NotedByName,
				NotedByPosition:               data.NotedByPosition,
				PostedBySignatureMediaID:      data.PostedBySignatureMediaID,
				PostedBySignatureMedia:        MediaManager(service).ToModel(data.PostedBySignatureMedia),
				PostedByName:                  data.PostedByName,
				PostedByPosition:              data.PostedByPosition,
				PaidBySignatureMediaID:        data.PaidBySignatureMediaID,
				PaidBySignatureMedia:          MediaManager(service).ToModel(data.PaidBySignatureMedia),
				PaidByName:                    data.PaidByName,
				PaidByPosition:                data.PaidByPosition,
				CurrencyID:                    data.CurrencyID,
				Currency:                      CurrencyManager(service).ToModel(data.Currency),
				EndedAt:                       endedAt,

				UnbalancedAccountID: data.UnbalancedAccountID,
				UnbalancedAccount:   UnbalancedAccountManager(service).ToModel(data.UnbalancedAccount),
			}
		},
		Created: func(data *types.TransactionBatch) registry.Topics {
			return []string{
				"transaction_batch.create",
				fmt.Sprintf("transaction_batch.create.%s", data.ID),
				fmt.Sprintf("transaction_batch.create.branch.%s", data.BranchID),
				fmt.Sprintf("transaction_batch.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("transaction_batch.create.user.%s", data.EmployeeUserID),
			}
		},
		Updated: func(data *types.TransactionBatch) registry.Topics {
			return []string{
				"transaction_batch.update",
				fmt.Sprintf("transaction_batch.update.%s", data.ID),
				fmt.Sprintf("transaction_batch.update.branch.%s", data.BranchID),
				fmt.Sprintf("transaction_batch.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("transaction_batch.update.user.%s", data.EmployeeUserID),
			}
		},
		Deleted: func(data *types.TransactionBatch) registry.Topics {
			return []string{
				"transaction_batch.delete",
				fmt.Sprintf("transaction_batch.delete.%s", data.ID),
				fmt.Sprintf("transaction_batch.delete.branch.%s", data.BranchID),
				fmt.Sprintf("transaction_batch.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("transaction_batch.delete.user.%s", data.EmployeeUserID),
			}

		},
	})
}

func TransactionBatchMinimal(context context.Context, service *horizon.HorizonService, id uuid.UUID) (*types.TransactionBatchResponse, error) {
	data, err := TransactionBatchManager(service).GetByID(context, id)
	if err != nil {
		return nil, err
	}

	var endedAt *string
	if data.EndedAt != nil {
		s := data.EndedAt.Format(time.RFC3339)
		endedAt = &s
	}
	return &types.TransactionBatchResponse{
		ID:               data.ID,
		CreatedAt:        data.CreatedAt.Format(time.RFC3339),
		CreatedByID:      data.CreatedByID,
		CreatedBy:        UserManager(service).ToModel(data.CreatedBy),
		UpdatedAt:        data.UpdatedAt.Format(time.RFC3339),
		UpdatedByID:      data.UpdatedByID,
		UpdatedBy:        UserManager(service).ToModel(data.UpdatedBy),
		OrganizationID:   data.OrganizationID,
		Organization:     OrganizationManager(service).ToModel(data.Organization),
		BranchID:         data.BranchID,
		Branch:           BranchManager(service).ToModel(data.Branch),
		EmployeeUserID:   data.EmployeeUserID,
		EmployeeUser:     UserManager(service).ToModel(data.EmployeeUser),
		BatchName:        data.BatchName,
		BeginningBalance: data.BeginningBalance,
		DepositInBank:    data.DepositInBank,
		CashCountTotal:   data.CashCountTotal,
		GrandTotal:       data.GrandTotal,
		Description:      data.Description,
		CanView:          data.CanView,
		IsClosed:         data.IsClosed,
		RequestView:      data.RequestView,
		CurrencyID:       data.CurrencyID,
		Currency:         CurrencyManager(service).ToModel(data.Currency),
		EndedAt:          endedAt,
	}, nil
}

func TransactionBatchCurrent(context context.Context, service *horizon.HorizonService, userID, organizationID, branchID uuid.UUID) (*types.TransactionBatch, error) {

	return TransactionBatchManager(service).ArrFindOne(context, []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "employee_user_id", Op: query.ModeEqual, Value: userID},
		{Field: "is_closed", Op: query.ModeEqual, Value: false},
	}, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func TransactionBatchViewRequests(context context.Context, service *horizon.HorizonService, organizationID, branchID uuid.UUID) ([]*types.TransactionBatch, error) {
	return TransactionBatchManager(service).ArrFind(context, []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "request_view", Op: query.ModeEqual, Value: true},
		{Field: "can_view", Op: query.ModeEqual, Value: false},
		{Field: "is_closed", Op: query.ModeEqual, Value: false},
	}, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func TransactionBatchCurrentDay(ctx context.Context, service *horizon.HorizonService, organizationID, branchID uuid.UUID) ([]*types.TransactionBatch, error) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	return TransactionBatchManager(service).ArrFind(ctx, []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "is_closed", Op: query.ModeEqual, Value: true},
		{Field: "created_at", Op: query.ModeGTE, Value: startOfDay},
		{Field: "created_at", Op: query.ModeLT, Value: endOfDay},
	}, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}
