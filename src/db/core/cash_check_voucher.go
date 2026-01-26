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

func CashCheckVoucherManager(service *horizon.HorizonService) *registry.Registry[
	types.CashCheckVoucher, types.CashCheckVoucherResponse, types.CashCheckVoucherRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.CashCheckVoucher, types.CashCheckVoucherResponse, types.CashCheckVoucherRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Currency",
			"EmployeeUser", "TransactionBatch", "EmployeeUser.Media",
			"PrintedBy", "ApprovedBy", "ReleasedBy",
			"PrintedBy.Media", "ApprovedBy.Media", "ReleasedBy.Media",
			"CashCheckVoucherTags", "CashCheckVoucherEntries",
			"CashCheckVoucherEntries.MemberProfile", "CashCheckVoucherEntries.Account", "CashCheckVoucherEntries.LoanTransaction", "CashCheckVoucherEntries.MemberProfile.Media",
			"CashCheckVoucherEntries.Account.Currency",
			"ApprovedBySignatureMedia", "PreparedBySignatureMedia", "CertifiedBySignatureMedia",
			"VerifiedBySignatureMedia", "CheckBySignatureMedia", "AcknowledgeBySignatureMedia",
			"NotedBySignatureMedia", "PostedBySignatureMedia", "PaidBySignatureMedia",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.CashCheckVoucher) *types.CashCheckVoucherResponse {
			if data == nil {
				return nil
			}
			var printedDate, approvedDate, releasedDate, entryDate *string
			if data.EntryDate != nil {
				str := data.EntryDate.Format(time.RFC3339)
				entryDate = &str
			}
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
			return &types.CashCheckVoucherResponse{
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
				CurrencyID:     data.CurrencyID,
				Currency:       CurrencyManager(service).ToModel(data.Currency),

				EmployeeUserID:     data.EmployeeUserID,
				EmployeeUser:       UserManager(service).ToModel(data.EmployeeUser),
				TransactionBatchID: data.TransactionBatchID,
				TransactionBatch:   TransactionBatchManager(service).ToModel(data.TransactionBatch),
				PrintedByID:        data.PrintedByID,
				PrintedBy:          UserManager(service).ToModel(data.PrintedBy),
				ApprovedByID:       data.ApprovedByID,
				ApprovedBy:         UserManager(service).ToModel(data.ApprovedBy),
				ReleasedByID:       data.ReleasedByID,
				ReleasedBy:         UserManager(service).ToModel(data.ReleasedBy),

				PayTo: data.PayTo,

				Status:            data.Status,
				Description:       data.Description,
				CashVoucherNumber: data.CashVoucherNumber,
				TotalDebit:        data.TotalDebit,
				TotalCredit:       data.TotalCredit,
				PrintCount:        data.PrintCount,
				EntryDate:         entryDate,
				PrintedDate:       printedDate,
				ApprovedDate:      approvedDate,
				ReleasedDate:      releasedDate,

				ApprovedBySignatureMediaID: data.ApprovedBySignatureMediaID,
				ApprovedBySignatureMedia:   MediaManager(service).ToModel(data.ApprovedBySignatureMedia),
				ApprovedByName:             data.ApprovedByName,
				ApprovedByPosition:         data.ApprovedByPosition,

				PreparedBySignatureMediaID: data.PreparedBySignatureMediaID,
				PreparedBySignatureMedia:   MediaManager(service).ToModel(data.PreparedBySignatureMedia),
				PreparedByName:             data.PreparedByName,
				PreparedByPosition:         data.PreparedByPosition,

				CertifiedBySignatureMediaID: data.CertifiedBySignatureMediaID,
				CertifiedBySignatureMedia:   MediaManager(service).ToModel(data.CertifiedBySignatureMedia),
				CertifiedByName:             data.CertifiedByName,
				CertifiedByPosition:         data.CertifiedByPosition,

				VerifiedBySignatureMediaID: data.VerifiedBySignatureMediaID,
				VerifiedBySignatureMedia:   MediaManager(service).ToModel(data.VerifiedBySignatureMedia),
				VerifiedByName:             data.VerifiedByName,
				VerifiedByPosition:         data.VerifiedByPosition,

				CheckBySignatureMediaID: data.CheckBySignatureMediaID,
				CheckBySignatureMedia:   MediaManager(service).ToModel(data.CheckBySignatureMedia),
				CheckByName:             data.CheckByName,
				CheckByPosition:         data.CheckByPosition,

				AcknowledgeBySignatureMediaID: data.AcknowledgeBySignatureMediaID,
				AcknowledgeBySignatureMedia:   MediaManager(service).ToModel(data.AcknowledgeBySignatureMedia),
				AcknowledgeByName:             data.AcknowledgeByName,
				AcknowledgeByPosition:         data.AcknowledgeByPosition,

				NotedBySignatureMediaID: data.NotedBySignatureMediaID,
				NotedBySignatureMedia:   MediaManager(service).ToModel(data.NotedBySignatureMedia),
				NotedByName:             data.NotedByName,
				NotedByPosition:         data.NotedByPosition,

				PostedBySignatureMediaID: data.PostedBySignatureMediaID,
				PostedBySignatureMedia:   MediaManager(service).ToModel(data.PostedBySignatureMedia),
				PostedByName:             data.PostedByName,
				PostedByPosition:         data.PostedByPosition,

				PaidBySignatureMediaID: data.PaidBySignatureMediaID,
				PaidBySignatureMedia:   MediaManager(service).ToModel(data.PaidBySignatureMedia),
				PaidByName:             data.PaidByName,
				PaidByPosition:         data.PaidByPosition,

				CashCheckVoucherTags:    CashCheckVoucherTagManager(service).ToModels(data.CashCheckVoucherTags),
				CashCheckVoucherEntries: CashCheckVoucherEntryManager(service).ToModels(data.CashCheckVoucherEntries),

				Name: data.Name,
			}
		},
		Created: func(data *types.CashCheckVoucher) registry.Topics {
			return []string{
				"cash_check_voucher.create",
				fmt.Sprintf("cash_check_voucher.create.%s", data.ID),
				fmt.Sprintf("cash_check_voucher.create.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.CashCheckVoucher) registry.Topics {
			return []string{
				"cash_check_voucher.update",
				fmt.Sprintf("cash_check_voucher.update.%s", data.ID),
				fmt.Sprintf("cash_check_voucher.update.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.CashCheckVoucher) registry.Topics {
			return []string{
				"cash_check_voucher.delete",
				fmt.Sprintf("cash_check_voucher.delete.%s", data.ID),
				fmt.Sprintf("cash_check_voucher.delete.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func CashCheckVoucherCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.CashCheckVoucher, error) {
	return CashCheckVoucherManager(service).Find(context, &types.CashCheckVoucher{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func CashCheckVoucherDraft(ctx context.Context, service *horizon.HorizonService,
	branchID, organizationID uuid.UUID) ([]*types.CashCheckVoucher, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "approved_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "printed_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	cashCheckVouchers, err := CashCheckVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}
	return cashCheckVouchers, nil
}

func CashCheckVoucherPrinted(ctx context.Context, service *horizon.HorizonService,
	branchID, organizationID uuid.UUID) ([]*types.CashCheckVoucher, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	cashCheckVouchers, err := CashCheckVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}
	return cashCheckVouchers, nil
}

func CashCheckVoucherApproved(ctx context.Context, service *horizon.HorizonService,
	branchID, organizationID uuid.UUID) ([]*types.CashCheckVoucher, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	cashCheckVouchers, err := CashCheckVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}
	return cashCheckVouchers, nil
}

func CashCheckVoucherReleased(ctx context.Context, service *horizon.HorizonService,
	branchID, organizationID uuid.UUID) ([]*types.CashCheckVoucher, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
	}

	cashCheckVouchers, err := CashCheckVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}
	return cashCheckVouchers, nil
}

func CashCheckVoucherReleasedCurrentDay(ctx context.Context, service *horizon.HorizonService,
	branchID uuid.UUID, organizationID uuid.UUID) ([]*types.CashCheckVoucher, error) {
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

	cashCheckVouchers, err := CashCheckVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}
	return cashCheckVouchers, nil
}
