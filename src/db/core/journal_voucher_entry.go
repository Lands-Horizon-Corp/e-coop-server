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

func JournalVoucherEntryManager(service *horizon.HorizonService) *registry.Registry[
	types.JournalVoucherEntry, types.JournalVoucherEntryResponse, types.JournalVoucherEntryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.JournalVoucherEntry, types.JournalVoucherEntryResponse, types.JournalVoucherEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy",
			"Account", "MemberProfile", "EmployeeUser", "JournalVoucher",
			"Account.Currency", "LoanTransaction",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.JournalVoucherEntry) *types.JournalVoucherEntryResponse {
			if data == nil {
				return nil
			}
			return &types.JournalVoucherEntryResponse{
				ID:                data.ID,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				CreatedByID:       data.CreatedByID,
				CreatedBy:         UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:       data.UpdatedByID,
				UpdatedBy:         UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:    data.OrganizationID,
				Organization:      OrganizationManager(service).ToModel(data.Organization),
				BranchID:          data.BranchID,
				Branch:            BranchManager(service).ToModel(data.Branch),
				AccountID:         data.AccountID,
				Account:           AccountManager(service).ToModel(data.Account),
				MemberProfileID:   data.MemberProfileID,
				MemberProfile:     MemberProfileManager(service).ToModel(data.MemberProfile),
				EmployeeUserID:    data.EmployeeUserID,
				EmployeeUser:      UserManager(service).ToModel(data.EmployeeUser),
				JournalVoucherID:  data.JournalVoucherID,
				JournalVoucher:    JournalVoucherManager(service).ToModel(data.JournalVoucher),
				LoanTransactionID: data.LoanTransactionID,
				LoanTransaction:   LoanTransactionManager(service).ToModel(data.LoanTransaction),
				Description:       data.Description,
				Debit:             data.Debit,
				Credit:            data.Credit,
			}
		},
		Created: func(data *types.JournalVoucherEntry) registry.Topics {
			return []string{
				"journal_voucher_entry.create",
				fmt.Sprintf("journal_voucher_entry.create.%s", data.ID),
				fmt.Sprintf("journal_voucher_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.JournalVoucherEntry) registry.Topics {
			return []string{
				"journal_voucher_entry.update",
				fmt.Sprintf("journal_voucher_entry.update.%s", data.ID),
				fmt.Sprintf("journal_voucher_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.JournalVoucherEntry) registry.Topics {
			return []string{
				"journal_voucher_entry.delete",
				fmt.Sprintf("journal_voucher_entry.delete.%s", data.ID),
				fmt.Sprintf("journal_voucher_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func JournalVoucherEntryCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.JournalVoucherEntry, error) {
	return JournalVoucherEntryManager(service).Find(context, &types.JournalVoucherEntry{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
