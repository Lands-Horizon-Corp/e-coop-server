package core

import (
	"context"
	"errors"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func GeneralLedgerManager(service *horizon.HorizonService) *registry.Registry[types.GeneralLedger, types.GeneralLedgerResponse, types.GeneralLedgerRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.GeneralLedger, types.GeneralLedgerResponse, types.GeneralLedgerRequest]{
		Preloads: []string{
			"Account",
			"Account.Currency",
			"EmployeeUser",
			"EmployeeUser.Media",
			"MemberProfile",
			"MemberProfile.Media",
			"MemberJointAccount",
			"MemberJointAccount.PictureMedia",
			"PaymentType",
			"AdjustmentEntry",
			"SignatureMedia",
			"Bank",
			"ProofOfPaymentMedia",
			"Currency",
			"CreatedBy.Media",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.GeneralLedger) *types.GeneralLedgerResponse {
			if data == nil {
				return nil
			}
			if data.AccountID == nil {
				return nil
			}
			accountHistoryID, err := GetAccountHistoryLatestByTimeHistoryID(
				context.Background(),
				service,
				*data.AccountID,
				data.OrganizationID,
				data.BranchID,
				&data.CreatedAt,
			)
			if err != nil {
				accountHistoryID = nil
			}
			return &types.GeneralLedgerResponse{
				ID:                         data.ID,
				EntryDate:                  data.EntryDate.Format(time.RFC3339),
				CreatedAt:                  data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                data.CreatedByID,
				CreatedBy:                  UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:                  data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                data.UpdatedByID,
				UpdatedBy:                  UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:             data.OrganizationID,
				Organization:               OrganizationManager(service).ToModel(data.Organization),
				BranchID:                   data.BranchID,
				Branch:                     BranchManager(service).ToModel(data.Branch),
				AccountID:                  data.AccountID,
				Account:                    AccountManager(service).ToModel(data.Account),
				TransactionID:              data.TransactionID,
				Transaction:                TransactionManager(service).ToModel(data.Transaction),
				TransactionBatchID:         data.TransactionBatchID,
				TransactionBatch:           TransactionBatchManager(service).ToModel(data.TransactionBatch),
				EmployeeUserID:             data.EmployeeUserID,
				EmployeeUser:               UserManager(service).ToModel(data.EmployeeUser),
				MemberProfileID:            data.MemberProfileID,
				MemberProfile:              MemberProfileManager(service).ToModel(data.MemberProfile),
				MemberJointAccountID:       data.MemberJointAccountID,
				MemberJointAccount:         MemberJointAccountManager(service).ToModel(data.MemberJointAccount),
				TransactionReferenceNumber: data.TransactionReferenceNumber,
				ReferenceNumber:            data.ReferenceNumber,
				PaymentTypeID:              data.PaymentTypeID,
				PaymentType:                PaymentTypeManager(service).ToModel(data.PaymentType),
				Source:                     data.Source,
				JournalVoucherID:           data.JournalVoucherID,
				AdjustmentEntryID:          data.AdjustmentEntryID,
				AdjustmentEntry:            AdjustmentEntryManager(service).ToModel(data.AdjustmentEntry),
				LoanTransactionID:          data.LoanTransactionID,
				LoanTransaction:            LoanTransactionManager(service).ToModel(data.LoanTransaction),
				TypeOfPaymentType:          data.TypeOfPaymentType,
				Credit:                     data.Credit,
				Debit:                      data.Debit,
				SignatureMediaID:           data.SignatureMediaID,
				SignatureMedia:             MediaManager(service).ToModel(data.SignatureMedia),

				BankID:                data.BankID,
				Bank:                  BankManager(service).ToModel(data.Bank),
				ProofOfPaymentMediaID: data.ProofOfPaymentMediaID,
				ProofOfPaymentMedia:   MediaManager(service).ToModel(data.ProofOfPaymentMedia),
				CurrencyID:            data.CurrencyID,
				Currency:              CurrencyManager(service).ToModel(data.Currency),
				BankReferenceNumber:   data.BankReferenceNumber,
				Description:           data.Description,
				PrintNumber:           data.PrintNumber,
				AccountHistoryID:      accountHistoryID,
				Balance:               data.Balance}
		},
		Created: func(data *types.GeneralLedger) registry.Topics {
			return []string{}
		},
		Updated: func(data *types.GeneralLedger) registry.Topics {
			return []string{}
		},
		Deleted: func(data *types.GeneralLedger) registry.Topics {
			return []string{}
		},
	})
}

func CreateGeneralLedgerEntry(
	ctx context.Context,
	service *horizon.HorizonService,
	tx *gorm.DB,
	data *types.GeneralLedger,
) error {
	if data == nil {
		return eris.New("general ledger: data is nil")
	}
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: data.OrganizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: data.BranchID},
		{Field: "account_id", Op: query.ModeEqual, Value: data.AccountID},
	}
	if data.Account != nil &&
		data.Account.Type != types.AccountTypeOther &&
		data.MemberProfileID != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "member_profile_id",
			Op:    query.ModeEqual,
			Value: data.MemberProfileID,
		})
	}
	ledger, err := GeneralLedgerManager(service).
		ArrFindOneWithLock(ctx, tx, filters, []query.ArrFilterSortSQL{
			{Field: "created_at", Order: "DESC"},
		})

	previousBalance := decimal.Zero
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// skipped
		} else {
			return eris.Wrap(err, "general ledger: failed to fetch previous ledger")
		}
	} else if ledger != nil {
		previousBalance = decimal.NewFromFloat(ledger.Balance)
	}

	debit := decimal.NewFromFloat(data.Debit)
	credit := decimal.NewFromFloat(data.Credit)

	// Calculate balance change
	var balanceChange decimal.Decimal
	if data.Account == nil {
		balanceChange = debit.Sub(credit)
	} else {
		switch data.Account.GeneralLedgerType {
		case types.GLTypeAssets, types.GLTypeExpenses:
		case types.GLTypeLiabilities, types.GLTypeEquity, types.GLTypeRevenue:
			balanceChange = credit.Sub(debit)
		default:
			balanceChange = credit.Sub(debit)
		}
	}

	newBalance := previousBalance.Add(balanceChange)
	// Load branch settings
	userOrg, err := UserOrganizationManager(service).FindOne(ctx, &types.UserOrganization{
		BranchID:       &data.BranchID,
		OrganizationID: data.OrganizationID,
		UserID:         *data.EmployeeUserID,
	}, "")
	if err != nil {
		return eris.Wrap(err, "general ledger: failed to get branch settings")
	}

	// --- CRITICAL PANIC FIXES START ---
	// You cannot access ledger.Source if ledger is nil!
	if ledger != nil && data.Account != nil {
		if ledger.Source == types.GeneralLedgerSourceWithdraw && userOrg.SettingsMaintainingBalance {
			minAmount := decimal.NewFromFloat(data.Account.MinAmount)
			if newBalance.LessThan(minAmount) {
				return eris.New("general ledger: maintaining balance violation")
			}
		}

		if ledger.Source == types.GeneralLedgerSourceWithdraw && !userOrg.SettingsAllowWithdrawNegativeBalance {
			zero := decimal.NewFromFloat(0)
			if newBalance.LessThan(zero) {
				return eris.New("general ledger: negative balance violation")
			}
		}
	} else {
		// skipping
	}
	data.Balance, _ = newBalance.Float64()
	if err := GeneralLedgerManager(service).CreateWithTx(ctx, tx, data); err != nil {
		return eris.Wrap(err, "general ledger: failed to create entry")
	}

	// Sync member accounting ledger
	if data.Account != nil &&
		data.Account.Type != types.AccountTypeOther &&
		data.MemberProfileID != nil {
		_, err = MemberAccountingLedgerUpdateOrCreate(
			ctx,
			service,
			tx,
			data.Balance,
			types.MemberAccountingLedgerUpdateOrCreateParams{
				MemberProfileID: *data.MemberProfileID,
				AccountID:       *data.AccountID,
				OrganizationID:  data.OrganizationID,
				BranchID:        data.BranchID,
				UserID:          data.CreatedByID,
				LastPayTime:     data.EntryDate,
			},
		)
		if err != nil {
			return eris.Wrap(err, "general ledger: failed to update member accounting ledger")
		}
	}
	return nil
}

func GeneralLedgerPrintMaxNumber(
	ctx context.Context, service *horizon.HorizonService,
	memberProfileID, accountID, branchID, organizationID uuid.UUID,
) (int, error) {
	filters := []query.ArrFilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
	}
	res, err := GeneralLedgerManager(service).ArrGetMaxInt(ctx, "print_number", filters)
	if err != nil {
		return 0, err
	}
	return int(res), nil
}
func GeneralLedgerCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID, branchID uuid.UUID) ([]*types.GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return GeneralLedgerManager(service).ArrFind(context, filters, nil)
}

func GeneralLedgerCurrentMemberAccount(context context.Context, service *horizon.HorizonService, memberProfileID, accountID, organizationID, branchID uuid.UUID) (*types.GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
	}

	return GeneralLedgerManager(service).ArrFindOne(context, filters, nil)
}

func GeneralLedgerExcludeCashonHand(
	ctx context.Context, service *horizon.HorizonService,
	transactionID, organizationID,
	branchID uuid.UUID,
) ([]*types.GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "transaction_id", Op: query.ModeEqual, Value: transactionID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	branchSetting, err := BranchSettingManager(service).FindOne(ctx, &types.BranchSetting{BranchID: branchID})
	if err != nil {
		return nil, err
	}

	if branchSetting.CashOnHandAccountID != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "account_id",
			Op:    query.ModeNotEqual,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}

	return GeneralLedgerManager(service).ArrFind(ctx, filters, nil)
}

func GeneralLedgerExcludeCashonHandWithType(
	ctx context.Context, service *horizon.HorizonService,
	transactionID, organizationID, branchID uuid.UUID,
	paymentType *types.TypeOfPaymentType,
) ([]*types.GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "transaction_id", Op: query.ModeEqual, Value: transactionID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	if paymentType != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "type_of_payment_type",
			Op:    query.ModeEqual,
			Value: *paymentType,
		})
	}

	branchSetting, err := BranchSettingManager(service).FindOne(ctx, &types.BranchSetting{BranchID: branchID})
	if err != nil {
		return nil, err
	}

	if branchSetting.CashOnHandAccountID != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "account_id",
			Op:    query.ModeNotEqual,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}

	return GeneralLedgerManager(service).ArrFind(ctx, filters, nil)
}

func GeneralLedgerExcludeCashonHandWithSource(
	ctx context.Context, service *horizon.HorizonService,
	transactionID, organizationID, branchID uuid.UUID,
	source *types.GeneralLedgerSource,
) ([]*types.GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "transaction_id", Op: query.ModeEqual, Value: transactionID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}
	if source != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "source",
			Op:    query.ModeEqual,
			Value: *source,
		})
	}
	branchSetting, err := BranchSettingManager(service).FindOne(ctx, &types.BranchSetting{BranchID: branchID})
	if err != nil {
		return nil, err
	}
	if branchSetting.CashOnHandAccountID != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "account_id",
			Op:    query.ModeNotEqual,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}
	return GeneralLedgerManager(service).ArrFind(ctx, filters, nil)
}

func GeneralLedgerExcludeCashonHandWithFilters(
	ctx context.Context, service *horizon.HorizonService,
	transactionID, organizationID, branchID uuid.UUID,
	paymentType *types.TypeOfPaymentType,
	source *types.GeneralLedgerSource,
) ([]*types.GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "transaction_id", Op: query.ModeEqual, Value: transactionID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	if paymentType != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "type_of_payment_type",
			Op:    query.ModeEqual,
			Value: *paymentType,
		})
	}

	if source != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "source",
			Op:    query.ModeEqual,
			Value: *source,
		})
	}

	branchSetting, err := BranchSettingManager(service).FindOne(ctx, &types.BranchSetting{BranchID: branchID})
	if err != nil {
		return nil, err
	}

	if branchSetting.CashOnHandAccountID != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "account_id",
			Op:    query.ModeNotEqual,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}

	return GeneralLedgerManager(service).ArrFind(ctx, filters, nil)
}

func GeneralLedgerAlignments(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.GeneralLedgerAccountsGrouping, error) {
	glGroupings, err := GeneralLedgerAccountsGroupingManager(service).Find(context, &types.GeneralLedgerAccountsGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to get general ledger groupings")
	}

	for _, grouping := range glGroupings {
		if grouping != nil {
			grouping.GeneralLedgerDefinitionEntries = []*types.GeneralLedgerDefinition{}
			entries, err := GeneralLedgerDefinitionManager(service).ArrFind(context,
				[]query.ArrFilterSQL{
					{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
					{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
					{Field: "general_ledger_accounts_grouping_id", Op: query.ModeEqual, Value: grouping.ID},
				},
				[]query.ArrFilterSortSQL{
					{Field: "created_at", Order: query.SortOrderAsc},
				},
			)
			if err != nil {
				return nil, eris.Wrap(err, "failed to get general ledger definition entries")
			}

			var filteredEntries []*types.GeneralLedgerDefinition
			for _, entry := range entries {
				if entry.GeneralLedgerDefinitionEntryID == nil {
					filteredEntries = append(filteredEntries, entry)
				}
			}

			grouping.GeneralLedgerDefinitionEntries = filteredEntries
		}
	}
	return glGroupings, nil
}

func GeneralLedgerCurrentMemberAccountEntries(
	ctx context.Context, service *horizon.HorizonService,
	memberProfileID, accountID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
) ([]*types.GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "account_id", Op: query.ModeNotEqual, Value: cashOnHandAccountID},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "entry_date", Order: query.SortOrderDesc},
		{Field: "created_at", Order: query.SortOrderDesc},
	}
	return GeneralLedgerManager(service).ArrFind(ctx, filters, sorts)
}

func GeneralLedgerMemberAccountTotal(
	ctx context.Context, service *horizon.HorizonService,
	memberProfileID, accountID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
) ([]*types.GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "account_id", Op: query.ModeNotEqual, Value: cashOnHandAccountID},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	}
	return GeneralLedgerManager(service).ArrFind(ctx, filters, sorts)
}

func GeneralLedgerMemberProfileEntries(
	ctx context.Context, service *horizon.HorizonService,
	memberProfileID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
) ([]*types.GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeNotEqual, Value: cashOnHandAccountID},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	}
	return GeneralLedgerManager(service).ArrFind(ctx, filters, sorts)
}

func GeneralLedgerMemberProfileEntriesByPaymentType(
	ctx context.Context, service *horizon.HorizonService,
	memberProfileID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
	paymentType types.TypeOfPaymentType,
) ([]*types.GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "type_of_payment_type", Op: query.ModeEqual, Value: paymentType},
		{Field: "account_id", Op: query.ModeNotEqual, Value: cashOnHandAccountID},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	}
	return GeneralLedgerManager(service).ArrFind(ctx, filters, sorts)
}

func GeneralLedgerMemberProfileEntriesBySource(
	ctx context.Context, service *horizon.HorizonService,
	memberProfileID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
	source types.GeneralLedgerSource,
) ([]*types.GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "source", Op: query.ModeEqual, Value: source},
		{Field: "account_id", Op: query.ModeNotEqual, Value: cashOnHandAccountID},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	}
	return GeneralLedgerManager(service).ArrFind(ctx, filters, sorts)
}

func GeneralLedgerByLoanTransaction(
	ctx context.Context, service *horizon.HorizonService,
	loanTransactionID, organizationID, branchID uuid.UUID,
) ([]*types.GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "loan_transaction_id", Op: query.ModeEqual, Value: loanTransactionID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	sorts := []query.ArrFilterSortSQL{
		{Field: "entry_date", Order: "DESC NULLS LAST"},
		{Field: "created_at", Order: "DESC"},
	}

	entries, err := GeneralLedgerManager(service).ArrFind(ctx, filters, sorts, "Account", "EmployeeUser", "EmployeeUser.Media")
	if err != nil {
		return nil, err
	}
	result := []*types.GeneralLedger{}
	for _, entry := range entries {
		if entry.Account.CashAndCashEquivalence {
			continue
		}
		if entry.Account.Type != types.AccountTypeLoan &&
			entry.Account.Type != types.AccountTypeFines &&
			entry.Account.Type != types.AccountTypeInterest &&
			entry.Account.Type != types.AccountTypeSVFLedger {
			continue
		}
		result = append(result, entry)
	}
	return result, nil
}

func GetGeneralLedgerOfMemberByEndOfDay(
	ctx context.Context, service *horizon.HorizonService,
	from, to time.Time,
	accountID, memberProfileID,
	organizationID,
	branchID uuid.UUID,
) ([]*types.GeneralLedger, error) {
	fromStartOfDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location()).UTC()
	toEndOfDay := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location()).UTC()
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "created_at", Op: query.ModeGTE, Value: fromStartOfDay},
		{Field: "created_at", Op: query.ModeLTE, Value: toEndOfDay},
	}

	sorts := []query.ArrFilterSortSQL{
		{Field: "created_at", Order: "DESC"},
		{Field: "entry_date", Order: "DESC"},
	}

	return GeneralLedgerManager(service).ArrFind(ctx, filters, sorts, "Account")
}
func GetDailyEndingBalances(
	ctx context.Context, service *horizon.HorizonService,
	from, to time.Time,
	accountID, memberProfileID, organizationID, branchID uuid.UUID,
) ([]float64, error) {

	if to.Before(from) {

		return nil, eris.New("invalid date range: 'to' date cannot be before 'from' date")
	}

	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

	entries, err := GetGeneralLedgerOfMemberByEndOfDay(ctx, service, from, to, accountID, memberProfileID, organizationID, branchID)
	if err != nil {
		return nil, err
	}

	entriesByDate := make(map[string]*types.GeneralLedger)
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		dateStr := entry.CreatedAt.UTC().Format("2006-01-02")
		if existing, exists := entriesByDate[dateStr]; !exists || entry.CreatedAt.After(existing.CreatedAt) {
			entriesByDate[dateStr] = entry

		}
	}

	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "created_at", Op: query.ModeLT, Value: fromDate},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "created_at", Order: "DESC"},
		{Field: "entry_date", Order: "DESC"},
	}

	startingBalance := 0.0
	lastEntry, err := GeneralLedgerManager(service).ArrFindOne(ctx, filters, sorts, "Account")
	if err == nil {
		if lastEntry != nil {
			startingBalance = lastEntry.Balance
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var dailyBalances []float64
	currentBalance := startingBalance

	for currentDate := fromDate; currentDate.Before(toDate) || currentDate.Equal(toDate); currentDate = currentDate.AddDate(0, 0, 1) {
		dateStr := currentDate.Format("2006-01-02")
		if entry, hasEntry := entriesByDate[dateStr]; hasEntry {
			if entry != nil {
				currentBalance = entry.Balance
			}
		}
		dailyBalances = append(dailyBalances, currentBalance)
	}
	return dailyBalances, nil
}

func DailyBookingCollection(
	ctx context.Context, service *horizon.HorizonService,
	date time.Time,
	organizationID uuid.UUID,
	branchID uuid.UUID,
) ([]*types.GeneralLedger, error) {

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.UTC)

	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "entry_date", Op: query.ModeGTE, Value: startOfDay},
		{Field: "entry_date", Op: query.ModeLTE, Value: endOfDay},
	}

	sorts := []query.ArrFilterSortSQL{
		{Field: "entry_date", Order: query.SortOrderAsc},
	}

	allData, err := GeneralLedgerManager(service).ArrFind(ctx, filters, sorts, "Account", "Account.Currency")
	if err != nil {
		return nil, err
	}
	result := make([]*types.GeneralLedger, 0)
	for _, item := range allData {
		if item.Source == types.GeneralLedgerSourcePayment || item.Source == types.GeneralLedgerSourceDeposit {
			result = append(result, item)
		}
	}
	return result, nil
}

func DailyDisbursementCollection(
	ctx context.Context, service *horizon.HorizonService,
	date time.Time,
	organizationID uuid.UUID,
	branchID uuid.UUID,
) ([]*types.GeneralLedger, error) {

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.UTC)

	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "entry_date", Op: query.ModeGTE, Value: startOfDay},
		{Field: "entry_date", Op: query.ModeLTE, Value: endOfDay},
	}

	sorts := []query.ArrFilterSortSQL{
		{Field: "entry_date", Order: query.SortOrderAsc},
	}

	allData, err := GeneralLedgerManager(service).ArrFind(ctx, filters, sorts, "Account", "Account.Currency")
	if err != nil {
		return nil, err
	}
	result := make([]*types.GeneralLedger, 0)
	for _, item := range allData {
		if item.Source == types.GeneralLedgerSourceWithdraw ||
			item.Source == types.GeneralLedgerSourceCheckVoucher ||
			item.Source == types.GeneralLedgerSourceLoan {
			result = append(result, item)
		}
	}

	return result, nil
}

func DailyJournalCollection(
	ctx context.Context, service *horizon.HorizonService,
	date time.Time,
	organizationID uuid.UUID,
	branchID uuid.UUID,
) ([]*types.GeneralLedger, error) {

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.UTC)

	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "entry_date", Op: query.ModeGTE, Value: startOfDay},
		{Field: "entry_date", Op: query.ModeLTE, Value: endOfDay},
	}

	sorts := []query.ArrFilterSortSQL{
		{Field: "entry_date", Order: query.SortOrderAsc},
	}

	allData, err := GeneralLedgerManager(service).ArrFind(ctx, filters, sorts, "Account", "Account.Currency")
	if err != nil {
		return nil, err
	}
	result := make([]*types.GeneralLedger, 0)
	for _, item := range allData {
		if item.Source == types.GeneralLedgerSourceJournalVoucher ||
			item.Source == types.GeneralLedgerSourceAdjustment {
			result = append(result, item)
		}
	}
	return result, nil
}
