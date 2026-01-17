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
	"gorm.io/gorm"
)

func AccountHistoryManager(service *horizon.HorizonService) *registry.Registry[types.AccountHistory, types.AccountHistoryResponse, types.AccountHistoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.AccountHistory, types.AccountHistoryResponse, types.AccountHistoryRequest,
	]{
		Preloads: []string{"CreatedBy", "CreatedBy.Media", "Account", "Account.Currency", "Organization", "Branch"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.AccountHistory) *types.AccountHistoryResponse {
			if data == nil {
				return nil
			}

			response := &types.AccountHistoryResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				AccountID:      data.AccountID,
				Account:        AccountManager(service).ToModel(data.Account),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),

				Name:                         data.Name,
				Description:                  data.Description,
				Type:                         data.Type,
				MinAmount:                    data.MinAmount,
				MaxAmount:                    data.MaxAmount,
				Index:                        data.Index,
				IsInternal:                   data.IsInternal,
				CashOnHand:                   data.CashOnHand,
				PaidUpShareCapital:           data.PaidUpShareCapital,
				ComputationType:              data.ComputationType,
				FinesAmort:                   data.FinesAmort,
				FinesMaturity:                data.FinesMaturity,
				InterestStandard:             data.InterestStandard,
				InterestSecured:              data.InterestSecured,
				FinesGracePeriodAmortization: data.FinesGracePeriodAmortization,
				AdditionalGracePeriod:        data.AdditionalGracePeriod,
				NoGracePeriodDaily:           data.NoGracePeriodDaily,
				FinesGracePeriodMaturity:     data.FinesGracePeriodMaturity,
				YearlySubscriptionFee:        data.YearlySubscriptionFee,
				CutOffDays:                   data.CutOffDays,
				CutOffMonths:                 data.CutOffMonths,

				LumpsumComputationType:                            data.LumpsumComputationType,
				InterestFinesComputationDiminishing:               data.InterestFinesComputationDiminishing,
				InterestFinesComputationDiminishingStraightYearly: data.InterestFinesComputationDiminishingStraightYearly,
				EarnedUnearnedInterest:                            data.EarnedUnearnedInterest,
				LoanSavingType:                                    data.LoanSavingType,
				InterestDeduction:                                 data.InterestDeduction,
				OtherDeductionEntry:                               data.OtherDeductionEntry,
				InterestSavingTypeDiminishingStraight:             data.InterestSavingTypeDiminishingStraight,
				OtherInformationOfAnAccount:                       data.OtherInformationOfAnAccount,

				GeneralLedgerType:                   data.GeneralLedgerType,
				HeaderRow:                           data.HeaderRow,
				CenterRow:                           data.CenterRow,
				TotalRow:                            data.TotalRow,
				GeneralLedgerGroupingExcludeAccount: data.GeneralLedgerGroupingExcludeAccount,
				Icon:                                data.Icon,

				ShowInGeneralLedgerSourceWithdraw:       data.ShowInGeneralLedgerSourceWithdraw,
				ShowInGeneralLedgerSourceDeposit:        data.ShowInGeneralLedgerSourceDeposit,
				ShowInGeneralLedgerSourceJournal:        data.ShowInGeneralLedgerSourceJournal,
				ShowInGeneralLedgerSourcePayment:        data.ShowInGeneralLedgerSourcePayment,
				ShowInGeneralLedgerSourceAdjustment:     data.ShowInGeneralLedgerSourceAdjustment,
				ShowInGeneralLedgerSourceJournalVoucher: data.ShowInGeneralLedgerSourceJournalVoucher,
				ShowInGeneralLedgerSourceCheckVoucher:   data.ShowInGeneralLedgerSourceCheckVoucher,

				CompassionFund:              data.CompassionFund,
				CompassionFundAmount:        data.CompassionFundAmount,
				CashAndCashEquivalence:      data.CashAndCashEquivalence,
				InterestStandardComputation: data.InterestStandardComputation,

				GeneralLedgerDefinitionID:      data.GeneralLedgerDefinitionID,
				FinancialStatementDefinitionID: data.FinancialStatementDefinitionID,
				AccountClassificationID:        data.AccountClassificationID,
				AccountCategoryID:              data.AccountCategoryID,
				MemberTypeID:                   data.MemberTypeID,
				CurrencyID:                     data.CurrencyID,
				DefaultPaymentTypeID:           data.DefaultPaymentTypeID,
				ComputationSheetID:             data.ComputationSheetID,
				LoanAccountID:                  data.LoanAccountID,

				CohCibFinesGracePeriodEntryCashHand:                data.CohCibFinesGracePeriodEntryCashHand,
				CohCibFinesGracePeriodEntryCashInBank:              data.CohCibFinesGracePeriodEntryCashInBank,
				CohCibFinesGracePeriodEntryDailyAmortization:       data.CohCibFinesGracePeriodEntryDailyAmortization,
				CohCibFinesGracePeriodEntryDailyMaturity:           data.CohCibFinesGracePeriodEntryDailyMaturity,
				CohCibFinesGracePeriodEntryWeeklyAmortization:      data.CohCibFinesGracePeriodEntryWeeklyAmortization,
				CohCibFinesGracePeriodEntryWeeklyMaturity:          data.CohCibFinesGracePeriodEntryWeeklyMaturity,
				CohCibFinesGracePeriodEntryMonthlyAmortization:     data.CohCibFinesGracePeriodEntryMonthlyAmortization,
				CohCibFinesGracePeriodEntryMonthlyMaturity:         data.CohCibFinesGracePeriodEntryMonthlyMaturity,
				CohCibFinesGracePeriodEntrySemiMonthlyAmortization: data.CohCibFinesGracePeriodEntrySemiMonthlyAmortization,
				CohCibFinesGracePeriodEntrySemiMonthlyMaturity:     data.CohCibFinesGracePeriodEntrySemiMonthlyMaturity,
				CohCibFinesGracePeriodEntryQuarterlyAmortization:   data.CohCibFinesGracePeriodEntryQuarterlyAmortization,
				CohCibFinesGracePeriodEntryQuarterlyMaturity:       data.CohCibFinesGracePeriodEntryQuarterlyMaturity,
				CohCibFinesGracePeriodEntrySemiAnnualAmortization:  data.CohCibFinesGracePeriodEntrySemiAnnualAmortization,
				CohCibFinesGracePeriodEntrySemiAnnualMaturity:      data.CohCibFinesGracePeriodEntrySemiAnnualMaturity,
				CohCibFinesGracePeriodEntryAnnualAmortization:      data.CohCibFinesGracePeriodEntryAnnualAmortization,
				CohCibFinesGracePeriodEntryAnnualMaturity:          data.CohCibFinesGracePeriodEntryAnnualMaturity,
				CohCibFinesGracePeriodEntryLumpsumAmortization:     data.CohCibFinesGracePeriodEntryLumpsumAmortization,
				CohCibFinesGracePeriodEntryLumpsumMaturity:         data.CohCibFinesGracePeriodEntryLumpsumMaturity,
			}

			return response
		},
		Created: func(data *types.AccountHistory) registry.Topics {
			return []string{
				"account_history.create",
				fmt.Sprintf("account_history.create.%s", data.ID),
				fmt.Sprintf("account_history.create.account.%s", data.AccountID),
				fmt.Sprintf("account_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_history.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.AccountHistory) registry.Topics {
			return []string{
				"account_history.update",
				fmt.Sprintf("account_history.update.%s", data.ID),
				fmt.Sprintf("account_history.update.account.%s", data.AccountID),
			}
		},
		Deleted: func(data *types.AccountHistory) registry.Topics {
			return []string{
				"account_history.delete",
				fmt.Sprintf("account_history.delete.%s", data.ID),
			}
		},
	})
}

func AccountHistoryToModel(data *types.AccountHistory) *types.Account {
	if data == nil {
		return nil
	}
	return &types.Account{
		ID:               data.AccountID, // Use the original account ID
		AccountHistoryID: &data.ID,       // Use the original account ID
		CreatedAt:        data.CreatedAt,
		UpdatedAt:        data.UpdatedAt,
		DeletedAt:        gorm.DeletedAt{}, // History doesn't track deletion state of original

		OrganizationID: data.OrganizationID,
		Organization:   data.Organization,
		BranchID:       data.BranchID,
		Branch:         data.Branch,

		Name:        data.Name,
		Description: data.Description,
		Type:        data.Type,
		MinAmount:   data.MinAmount,
		MaxAmount:   data.MaxAmount,
		Index:       data.Index,

		IsInternal:         data.IsInternal,
		CashOnHand:         data.CashOnHand,
		PaidUpShareCapital: data.PaidUpShareCapital,

		ComputationType: data.ComputationType,

		FinesAmort:       data.FinesAmort,
		FinesMaturity:    data.FinesMaturity,
		InterestStandard: data.InterestStandard,
		InterestSecured:  data.InterestSecured,

		FinesGracePeriodAmortization: data.FinesGracePeriodAmortization,
		AdditionalGracePeriod:        data.AdditionalGracePeriod,
		NoGracePeriodDaily:           data.NoGracePeriodDaily,
		FinesGracePeriodMaturity:     data.FinesGracePeriodMaturity,
		YearlySubscriptionFee:        data.YearlySubscriptionFee,
		CutOffDays:                   data.CutOffDays,
		CutOffMonths:                 data.CutOffMonths,

		LumpsumComputationType:                            data.LumpsumComputationType,
		InterestFinesComputationDiminishing:               data.InterestFinesComputationDiminishing,
		InterestFinesComputationDiminishingStraightYearly: data.InterestFinesComputationDiminishingStraightYearly,
		EarnedUnearnedInterest:                            data.EarnedUnearnedInterest,
		LoanSavingType:                                    data.LoanSavingType,
		InterestDeduction:                                 data.InterestDeduction,
		OtherDeductionEntry:                               data.OtherDeductionEntry,
		InterestSavingTypeDiminishingStraight:             data.InterestSavingTypeDiminishingStraight,
		OtherInformationOfAnAccount:                       data.OtherInformationOfAnAccount,

		GeneralLedgerType: data.GeneralLedgerType,

		HeaderRow: data.HeaderRow,
		CenterRow: data.CenterRow,
		TotalRow:  data.TotalRow,

		GeneralLedgerGroupingExcludeAccount: data.GeneralLedgerGroupingExcludeAccount,
		Icon:                                data.Icon,

		ShowInGeneralLedgerSourceWithdraw:       data.ShowInGeneralLedgerSourceWithdraw,
		ShowInGeneralLedgerSourceDeposit:        data.ShowInGeneralLedgerSourceDeposit,
		ShowInGeneralLedgerSourceJournal:        data.ShowInGeneralLedgerSourceJournal,
		ShowInGeneralLedgerSourcePayment:        data.ShowInGeneralLedgerSourcePayment,
		ShowInGeneralLedgerSourceAdjustment:     data.ShowInGeneralLedgerSourceAdjustment,
		ShowInGeneralLedgerSourceJournalVoucher: data.ShowInGeneralLedgerSourceJournalVoucher,
		ShowInGeneralLedgerSourceCheckVoucher:   data.ShowInGeneralLedgerSourceCheckVoucher,

		CompassionFund:         data.CompassionFund,
		CompassionFundAmount:   data.CompassionFundAmount,
		CashAndCashEquivalence: data.CashAndCashEquivalence,

		InterestStandardComputation: data.InterestStandardComputation,

		GeneralLedgerDefinitionID:      data.GeneralLedgerDefinitionID,
		FinancialStatementDefinitionID: data.FinancialStatementDefinitionID,
		AccountClassificationID:        data.AccountClassificationID,
		AccountCategoryID:              data.AccountCategoryID,
		MemberTypeID:                   data.MemberTypeID,
		CurrencyID:                     data.CurrencyID,
		DefaultPaymentTypeID:           data.DefaultPaymentTypeID,
		ComputationSheetID:             data.ComputationSheetID,
		LoanAccountID:                  data.LoanAccountID,

		CohCibFinesGracePeriodEntryCashHand:                data.CohCibFinesGracePeriodEntryCashHand,
		CohCibFinesGracePeriodEntryCashInBank:              data.CohCibFinesGracePeriodEntryCashInBank,
		CohCibFinesGracePeriodEntryDailyAmortization:       data.CohCibFinesGracePeriodEntryDailyAmortization,
		CohCibFinesGracePeriodEntryDailyMaturity:           data.CohCibFinesGracePeriodEntryDailyMaturity,
		CohCibFinesGracePeriodEntryWeeklyAmortization:      data.CohCibFinesGracePeriodEntryWeeklyAmortization,
		CohCibFinesGracePeriodEntryWeeklyMaturity:          data.CohCibFinesGracePeriodEntryWeeklyMaturity,
		CohCibFinesGracePeriodEntryMonthlyAmortization:     data.CohCibFinesGracePeriodEntryMonthlyAmortization,
		CohCibFinesGracePeriodEntryMonthlyMaturity:         data.CohCibFinesGracePeriodEntryMonthlyMaturity,
		CohCibFinesGracePeriodEntrySemiMonthlyAmortization: data.CohCibFinesGracePeriodEntrySemiMonthlyAmortization,
		CohCibFinesGracePeriodEntrySemiMonthlyMaturity:     data.CohCibFinesGracePeriodEntrySemiMonthlyMaturity,
		CohCibFinesGracePeriodEntryQuarterlyAmortization:   data.CohCibFinesGracePeriodEntryQuarterlyAmortization,
		CohCibFinesGracePeriodEntryQuarterlyMaturity:       data.CohCibFinesGracePeriodEntryQuarterlyMaturity,
		CohCibFinesGracePeriodEntrySemiAnnualAmortization:  data.CohCibFinesGracePeriodEntrySemiAnnualAmortization,
		CohCibFinesGracePeriodEntrySemiAnnualMaturity:      data.CohCibFinesGracePeriodEntrySemiAnnualMaturity,
		CohCibFinesGracePeriodEntryAnnualAmortization:      data.CohCibFinesGracePeriodEntryAnnualAmortization,
		CohCibFinesGracePeriodEntryAnnualMaturity:          data.CohCibFinesGracePeriodEntryAnnualMaturity,
		CohCibFinesGracePeriodEntryLumpsumAmortization:     data.CohCibFinesGracePeriodEntryLumpsumAmortization,
		CohCibFinesGracePeriodEntryLumpsumMaturity:         data.CohCibFinesGracePeriodEntryLumpsumMaturity,
	}
}

func GetAccountHistory(ctx context.Context, service *horizon.HorizonService, accountID uuid.UUID) ([]*types.AccountHistory, error) {
	filters := []query.ArrFilterSQL{
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
	}

	return AccountHistoryManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func GetAllAccountHistory(ctx context.Context, service *horizon.HorizonService, accountID, organizationID, branchID uuid.UUID) ([]*types.AccountHistory, error) {
	filters := []query.ArrFilterSQL{
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return AccountHistoryManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "created_at", Order: query.SortOrderDesc}, // Latest first
		{Field: "updated_at", Order: query.SortOrderDesc}, // Secondary sort
	})
}

func GetAccountHistoryLatestByTime(
	ctx context.Context,
	service *horizon.HorizonService,
	accountID, organizationID, branchID uuid.UUID,
	asOfDate *time.Time) (*types.Account, error) {
	currentTime := time.Now()
	if asOfDate == nil {
		asOfDate = &currentTime
	}
	filters := []query.ArrFilterSQL{
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "created_at", Op: query.ModeLTE, Value: asOfDate},
	}

	histories, err := AccountHistoryManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "created_at", Order: query.SortOrderDesc}, // Latest first
		{Field: "updated_at", Order: query.SortOrderDesc}, // Secondary sort
	})
	if err != nil {
		return nil, err
	}

	if len(histories) > 0 {
		return AccountHistoryToModel(histories[0]), nil
	}

	return nil, eris.Errorf("no history found for account %s at time %s", accountID, asOfDate.Format(time.RFC3339))
}

func GetAccountHistoryLatestByTimeHistoryID(
	ctx context.Context,
	service *horizon.HorizonService,
	accountID, organizationID, branchID uuid.UUID,
	asOfDate *time.Time) (*uuid.UUID, error) {
	currentTime := time.Now()
	if asOfDate == nil {
		asOfDate = &currentTime
	}
	filters := []query.ArrFilterSQL{
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "created_at", Op: query.ModeLTE, Value: asOfDate},
	}

	history, err := AccountHistoryManager(service).ArrFindOne(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "created_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, eris.Errorf("no history found for account %s at time %s", accountID, asOfDate.Format(time.RFC3339))
	}

	return &history.ID, nil
}

func GetAccountHistoryLatestByTimeHistory(
	ctx context.Context,
	service *horizon.HorizonService,
	accountID, organizationID, branchID uuid.UUID,
	asOfDate *time.Time) (*types.AccountHistory, error) {
	currentTime := time.Now()
	if asOfDate == nil {
		asOfDate = &currentTime
	}
	filters := []query.ArrFilterSQL{
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "created_at", Op: query.ModeLTE, Value: asOfDate},
	}

	histories, err := AccountHistoryManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "created_at", Order: query.SortOrderDesc}, // Latest first
		{Field: "updated_at", Order: query.SortOrderDesc}, // Secondary sort
	})
	if err != nil {
		return nil, err
	}

	if len(histories) > 0 {
		return histories[0], nil
	}

	return nil, eris.Errorf("no history found for account %s at time %s", accountID, asOfDate.Format(time.RFC3339))
}

func GetAccountHistoriesByFiltersAtTime(
	ctx context.Context,
	service *horizon.HorizonService,
	organizationID, branchID uuid.UUID,
	asOfDate *time.Time,
	loanAccountID *uuid.UUID,
	currencyID *uuid.UUID,
) ([]*types.Account, error) {
	currentTime := time.Now()
	if asOfDate == nil {
		asOfDate = &currentTime
	}

	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "created_at", Op: query.ModeLTE, Value: asOfDate},
	}

	if loanAccountID != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "loan_account_id", Op: query.ModeEqual, Value: *loanAccountID,
		})
	}

	if currencyID != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "currency_id", Op: query.ModeEqual, Value: *currencyID,
		})
	}

	histories, err := AccountHistoryManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "account_id", Order: query.SortOrderAsc},
		{Field: "created_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}

	accountMap := make(map[uuid.UUID]*types.AccountHistory)
	for _, history := range histories {
		if existing, found := accountMap[history.AccountID]; !found || history.CreatedAt.After(existing.CreatedAt) {
			accountMap[history.AccountID] = history
		}
	}

	var accounts []*types.Account
	for _, history := range accountMap {
		accounts = append(accounts, AccountHistoryToModel(history))
	}

	return accounts, nil
}
