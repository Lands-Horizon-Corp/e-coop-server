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
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

func JournalVoucherManager(service *horizon.HorizonService) *registry.Registry[
	types.JournalVoucher, types.JournalVoucherResponse, types.JournalVoucherRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.JournalVoucher, types.JournalVoucherResponse, types.JournalVoucherRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Currency", "PostedBy",
			"EmployeeUser", "EmployeeUser.Media", "TransactionBatch",
			"PrintedBy", "ApprovedBy", "ReleasedBy",
			"PrintedBy.Media", "ApprovedBy.Media", "ReleasedBy.Media",
			"JournalVoucherTags",
			"JournalVoucherEntries", "JournalVoucherEntries.Account", "JournalVoucherEntries.LoanTransaction",
			"JournalVoucherEntries.Account.Currency",
			"JournalVoucherEntries.MemberProfile", "JournalVoucherEntries.EmployeeUser",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.JournalVoucher) *types.JournalVoucherResponse {
			if data == nil {
				return nil
			}

			var postedAt *string
			if data.PostedAt != nil {
				postedAtStr := data.PostedAt.Format(time.RFC3339)
				postedAt = &postedAtStr
			}

			var printedDate, approvedDate, releasedDate *string
			if data.PrintedDate != nil {
				str := data.PrintedDate.Format(time.RFC3339)
				printedDate = &str
			}
			if data.ApprovedDate != nil {
				str := data.ApprovedDate.Format(time.RFC3339)
				approvedDate = &str
			}
			if data.ReleasedDate != nil {
				str := data.ReleasedDate.Format(time.RFC3339)
				releasedDate = &str
			}

			return &types.JournalVoucherResponse{
				ID:                    data.ID,
				CreatedAt:             data.CreatedAt.Format(time.RFC3339),
				CreatedByID:           data.CreatedByID,
				CreatedBy:             UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:             data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:           data.UpdatedByID,
				UpdatedBy:             UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:        data.OrganizationID,
				Organization:          OrganizationManager(service).ToModel(data.Organization),
				BranchID:              data.BranchID,
				Branch:                BranchManager(service).ToModel(data.Branch),
				CurrencyID:            data.CurrencyID,
				Currency:              CurrencyManager(service).ToModel(data.Currency),
				Name:                  data.Name,
				CashVoucherNumber:     data.CashVoucherNumber,
				Date:                  data.Date.Format("2006-01-02"),
				Description:           data.Description,
				Reference:             data.Reference,
				Status:                data.Status,
				PostedAt:              postedAt,
				PostedByID:            data.PostedByID,
				PostedBy:              UserManager(service).ToModel(data.PostedBy),
				EmployeeUserID:        data.EmployeeUserID,
				EmployeeUser:          UserManager(service).ToModel(data.EmployeeUser),
				TransactionBatchID:    data.TransactionBatchID,
				TransactionBatch:      TransactionBatchManager(service).ToModel(data.TransactionBatch),
				PrintedDate:           printedDate,
				PrintedByID:           data.PrintedByID,
				PrintedBy:             UserManager(service).ToModel(data.PrintedBy),
				PrintNumber:           data.PrintNumber,
				ApprovedDate:          approvedDate,
				ApprovedByID:          data.ApprovedByID,
				ApprovedBy:            UserManager(service).ToModel(data.ApprovedBy),
				ReleasedDate:          releasedDate,
				ReleasedByID:          data.ReleasedByID,
				ReleasedBy:            UserManager(service).ToModel(data.ReleasedBy),
				JournalVoucherTags:    JournalVoucherTagManager(service).ToModels(data.JournalVoucherTags),
				JournalVoucherEntries: JournalVoucherEntryManager(service).ToModels(data.JournalVoucherEntries),
				TotalDebit:            data.TotalDebit,
				TotalCredit:           data.TotalCredit,
			}
		},
		Created: func(data *types.JournalVoucher) registry.Topics {
			return []string{
				"journal_voucher.create",
				fmt.Sprintf("journal_voucher.create.%s", data.ID),
				fmt.Sprintf("journal_voucher.create.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.JournalVoucher) registry.Topics {
			return []string{
				"journal_voucher.update",
				fmt.Sprintf("journal_voucher.update.%s", data.ID),
				fmt.Sprintf("journal_voucher.update.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.JournalVoucher) registry.Topics {
			return []string{
				"journal_voucher.delete",
				fmt.Sprintf("journal_voucher.delete.%s", data.ID),
				fmt.Sprintf("journal_voucher.delete.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func JournalVoucherCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.JournalVoucher, error) {
	return JournalVoucherManager(service).Find(context, &types.JournalVoucher{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func ValidateJournalVoucherBalance(entries []*types.JournalVoucherEntry) error {
	totalDebit := decimal.NewFromInt(0)
	totalCredit := decimal.NewFromInt(0)
	for _, entry := range entries {
		totalDebit = totalDebit.Add(decimal.NewFromFloat(entry.Debit))
		totalCredit = totalCredit.Add(decimal.NewFromFloat(entry.Credit))
	}
	if !totalDebit.Equal(totalCredit) {
		return eris.Errorf(
			"journal voucher is not balanced: debit %.2f != credit %.2f",
			totalDebit.InexactFloat64(),
			totalCredit.InexactFloat64(),
		)
	}
	return nil
}

func JournalVoucherDraft(ctx context.Context, service *horizon.HorizonService, branchID, organizationID uuid.UUID) ([]*types.JournalVoucher, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "approved_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "printed_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	return JournalVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func JournalVoucherPrinted(ctx context.Context, service *horizon.HorizonService, branchID, organizationID uuid.UUID) ([]*types.JournalVoucher, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	return JournalVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func JournalVoucherApproved(ctx context.Context, service *horizon.HorizonService, branchID, organizationID uuid.UUID) ([]*types.JournalVoucher, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	return JournalVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func JournalVoucherReleased(ctx context.Context, service *horizon.HorizonService, branchID, organizationID uuid.UUID) ([]*types.JournalVoucher, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
	}

	return JournalVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func JournalVoucherReleasedCurrentDay(ctx context.Context, service *horizon.HorizonService, branchID uuid.UUID, organizationID uuid.UUID) ([]*types.JournalVoucher, error) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeGTE, Value: startOfDay},
		{Field: "released_date", Op: query.ModeLT, Value: endOfDay},
	}

	return JournalVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}
