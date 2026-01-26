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

func CashCheckVoucherEntryManager(service *horizon.HorizonService) *registry.Registry[
	types.CashCheckVoucherEntry, types.CashCheckVoucherEntryResponse, types.CashCheckVoucherEntryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.CashCheckVoucherEntry, types.CashCheckVoucherEntryResponse, types.CashCheckVoucherEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Account.Currency",
			"Account", "EmployeeUser", "TransactionBatch", "CashCheckVoucher",
			"MemberProfile", "MemberProfile.Media", "LoanTransaction",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.CashCheckVoucherEntry) *types.CashCheckVoucherEntryResponse {
			if data == nil {
				return nil
			}
			return &types.CashCheckVoucherEntryResponse{
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
				AccountID:          data.AccountID,
				Account:            AccountManager(service).ToModel(data.Account),
				EmployeeUserID:     data.EmployeeUserID,
				EmployeeUser:       UserManager(service).ToModel(data.EmployeeUser),
				TransactionBatchID: data.TransactionBatchID,
				TransactionBatch:   TransactionBatchManager(service).ToModel(data.TransactionBatch),
				CashCheckVoucherID: data.CashCheckVoucherID,
				CashCheckVoucher:   CashCheckVoucherManager(service).ToModel(data.CashCheckVoucher),
				LoanTransactionID:  data.LoanTransactionID,
				LoanTransaction:    LoanTransactionManager(service).ToModel(data.LoanTransaction),
				MemberProfileID:    data.MemberProfileID,
				MemberProfile:      MemberProfileManager(service).ToModel(data.MemberProfile),
				Debit:              data.Debit,
				Credit:             data.Credit,
				Description:        data.Description,
			}
		},
		Created: func(data *types.CashCheckVoucherEntry) registry.Topics {
			return []string{
				"cash_check_voucher_entry.create",
				fmt.Sprintf("cash_check_voucher_entry.create.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.CashCheckVoucherEntry) registry.Topics {
			return []string{
				"cash_check_voucher_entry.update",
				fmt.Sprintf("cash_check_voucher_entry.update.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.CashCheckVoucherEntry) registry.Topics {
			return []string{
				"cash_check_voucher_entry.delete",
				fmt.Sprintf("cash_check_voucher_entry.delete.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func CashCheckVoucherEntryCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.CashCheckVoucherEntry, error) {
	return CashCheckVoucherEntryManager(service).Find(context, &types.CashCheckVoucherEntry{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
