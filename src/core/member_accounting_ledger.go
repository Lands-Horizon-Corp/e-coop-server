package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func MemberAccountingLedgerManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberAccountingLedger, types.MemberAccountingLedgerResponse, types.MemberAccountingLedgerRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberAccountingLedger, types.MemberAccountingLedgerResponse, types.MemberAccountingLedgerRequest,
	]{
		Preloads: []string{
			"MemberProfile",
			"Account.Currency",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberAccountingLedger) *types.MemberAccountingLedgerResponse {
			if data == nil {
				return nil
			}
			var lastPay *string
			if data.LastPay != nil {
				s := data.LastPay.Format(time.RFC3339)
				lastPay = &s
			}
			return &types.MemberAccountingLedgerResponse{
				ID:                  data.ID,
				CreatedAt:           data.CreatedAt.Format(time.RFC3339),
				CreatedByID:         data.CreatedByID,
				CreatedBy:           UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:           data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:         data.UpdatedByID,
				UpdatedBy:           UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:      data.OrganizationID,
				Organization:        OrganizationManager(service).ToModel(data.Organization),
				BranchID:            data.BranchID,
				Branch:              BranchManager(service).ToModel(data.Branch),
				MemberProfileID:     data.MemberProfileID,
				MemberProfile:       MemberProfileManager(service).ToModel(data.MemberProfile),
				AccountID:           data.AccountID,
				Account:             AccountManager(service).ToModel(data.Account),
				Count:               data.Count,
				Balance:             data.Balance,
				Interest:            data.Interest,
				Fines:               data.Fines,
				Due:                 data.Due,
				CarriedForwardDue:   data.CarriedForwardDue,
				StoredValueFacility: data.StoredValueFacility,
				PrincipalDue:        data.PrincipalDue,
				LastPay:             lastPay,
			}
		},

		Created: func(data *types.MemberAccountingLedger) registry.Topics {
			return []string{
				"member_accounting_ledger.create",
				fmt.Sprintf("member_accounting_ledger.create.%s", data.ID),
				fmt.Sprintf("member_accounting_ledger.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_accounting_ledger.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_accounting_ledger.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *types.MemberAccountingLedger) registry.Topics {
			return []string{
				"member_accounting_ledger.update",
				fmt.Sprintf("member_accounting_ledger.update.%s", data.ID),
				fmt.Sprintf("member_accounting_ledger.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_accounting_ledger.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_accounting_ledger.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *types.MemberAccountingLedger) registry.Topics {
			return []string{
				"member_accounting_ledger.delete",
				fmt.Sprintf("member_accounting_ledger.delete.%s", data.ID),
				fmt.Sprintf("member_accounting_ledger.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_accounting_ledger.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_accounting_ledger.delete.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func MemberAccountingLedgerCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberAccountingLedger, error) {
	return MemberAccountingLedgerManager(service).Find(context, &types.MemberAccountingLedger{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func MemberAccountingLedgerMemberProfileEntries(ctx context.Context,
	service *horizon.HorizonService, memberProfileID, organizationID, branchID, cashOnHandAccountID uuid.UUID) ([]*types.MemberAccountingLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeNotEqual, Value: cashOnHandAccountID},
	}

	return MemberAccountingLedgerManager(service).ArrFind(ctx, filters, nil)
}

func MemberAccountingLedgerBranchEntries(ctx context.Context, service *horizon.HorizonService,
	organizationID, branchID, cashOnHandAccountID uuid.UUID) ([]*types.MemberAccountingLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeNotEqual, Value: cashOnHandAccountID},
	}
	return MemberAccountingLedgerManager(service).ArrFind(ctx, filters, nil)
}

func MemberAccountingLedgerFindForUpdate(
	ctx context.Context,
	service *horizon.HorizonService,
	tx *gorm.DB,
	memberProfileID,
	accountID,
	organizationID,
	branchID uuid.UUID,
) (*types.MemberAccountingLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}
	ledger, err := MemberAccountingLedgerManager(service).ArrFindOneWithLock(ctx, tx, filters, []query.ArrFilterSortSQL{
		{Field: "created_at", Order: "DESC"},
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return ledger, nil
}

func MemberAccountingLedgerUpdateOrCreate(
	ctx context.Context,
	service *horizon.HorizonService,
	tx *gorm.DB,
	balance float64,
	params types.MemberAccountingLedgerUpdateOrCreateParams,
) (*types.MemberAccountingLedger, error) {

	ledger, err := MemberAccountingLedgerFindForUpdate(
		ctx, service, tx,
		params.MemberProfileID,
		params.AccountID,
		params.OrganizationID,
		params.BranchID,
	)
	if err != nil {
		return nil, eris.Wrap(err, "failed to find member accounting ledger for update")
	}

	if ledger == nil || ledger.ID == uuid.Nil {

		ledger = &types.MemberAccountingLedger{
			CreatedAt:           time.Now().UTC(),
			CreatedByID:         params.UserID,
			UpdatedAt:           time.Now().UTC(),
			UpdatedByID:         params.UserID,
			OrganizationID:      params.OrganizationID,
			BranchID:            params.BranchID,
			MemberProfileID:     params.MemberProfileID,
			AccountID:           params.AccountID,
			Balance:             balance,
			LastPay:             &params.LastPayTime,
			Count:               1,
			Interest:            0,
			Fines:               0,
			Due:                 0,
			CarriedForwardDue:   0,
			StoredValueFacility: 0,
			PrincipalDue:        0,
		}
		if tx == nil {
			return nil, eris.New("database tx is nil")
		}
		err = tx.WithContext(ctx).Create(ledger).Error
		if err != nil {
			return nil, eris.Wrap(err, "failed to create member accounting ledger")
		}
	} else {
		ledger.Balance = balance
		ledger.LastPay = &params.LastPayTime
		ledger.UpdatedAt = time.Now().UTC()
		ledger.UpdatedByID = params.UserID
		ledger.Count++

		if tx == nil {
			return nil, eris.New("database tx is nil")
		}
		err = tx.WithContext(ctx).Save(ledger).Error
		if err != nil {
			return nil, eris.Wrap(err, "failed to update member accounting ledger")
		}
	}

	return ledger, nil
}

func MemberAccountingLedgerFilterByCriteria(
	ctx context.Context,
	service *horizon.HorizonService,
	organizationID,
	branchID uuid.UUID,
	accountID,
	memberTypeID *uuid.UUID,
	includeClosedAccounts bool,
) ([]*types.MemberAccountingLedger, error) {
	result := []*types.MemberAccountingLedger{}
	memberAccountingLedger, err := MemberAccountingLedgerManager(service).ArrFind(ctx, []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
	}, nil, "MemberProfile", "Account.Currency")
	if err != nil {
		return nil, eris.Wrap(err, "failed to find member accounting ledgers by account")
	}
	for _, ledger := range memberAccountingLedger {
		if !includeClosedAccounts && ledger.MemberProfile.IsClosed {
			continue
		}
		if helpers.UUIDPtrEqual(ledger.MemberProfile.MemberTypeID, memberTypeID) {
			result = append(result, ledger)
		}
	}

	return result, nil
}

func MemberAccountingLedgerByBrowseReference(ctx context.Context, service *horizon.HorizonService,
	includeClosedAccounts bool, data []*types.BrowseReference) ([]*types.MemberAccountingLedgerBrowseReference, error) {
	memberAccountingLedger := []*types.MemberAccountingLedgerBrowseReference{}
	for _, browseRef := range data {
		ledgers, err := MemberAccountingLedgerFilterByCriteria(
			ctx, service, browseRef.OrganizationID, browseRef.BranchID,
			browseRef.AccountID,
			browseRef.MemberTypeID, includeClosedAccounts)
		if err != nil {
			return nil, eris.Wrap(err, "failed to filter member accounting ledgers by browse reference")
		}

		for _, ledger := range ledgers {
			memberAccountingLedger = append(memberAccountingLedger, &types.MemberAccountingLedgerBrowseReference{
				MemberAccountingLedger: ledger,
				BrowseReference:        browseRef,
			})
		}
	}

	return memberAccountingLedger, nil
}
