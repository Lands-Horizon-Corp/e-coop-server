package core

import (
	"context"
	"errors"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type GeneralLedgerSource string

const (
	GeneralLedgerSourceWithdraw           GeneralLedgerSource = "withdraw"
	GeneralLedgerSourceDeposit            GeneralLedgerSource = "deposit"
	GeneralLedgerSourceJournal            GeneralLedgerSource = "journal"
	GeneralLedgerSourcePayment            GeneralLedgerSource = "payment"
	GeneralLedgerSourceAdjustment         GeneralLedgerSource = "adjustment"
	GeneralLedgerSourceJournalVoucher     GeneralLedgerSource = "journal voucher"
	GeneralLedgerSourceCheckVoucher       GeneralLedgerSource = "check voucher"
	GeneralLedgerSourceLoan               GeneralLedgerSource = "loan"
	GeneralLedgerSourceSavingsInterest    GeneralLedgerSource = "savings interest"
	GeneralLedgerSourceMutualContribution GeneralLedgerSource = "mutual contribution"
	GeneralLedgerSourcDisbursement        GeneralLedgerSource = "disbursement"
	GeneralLedgerSourcBlotter             GeneralLedgerSource = "blotter"
)

type (
	GeneralLedger struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now();index" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id,omitempty"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger;index:idx_org_branch_account_member;index:idx_transaction_batch_entry" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`

		BranchID uuid.UUID `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger;index:idx_org_branch_account_member;index:idx_transaction_batch_entry" json:"branch_id"`
		Branch   *Branch   `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID *uuid.UUID `gorm:"type:uuid;index:idx_org_branch_account_member" json:"account_id,omitempty"`
		Account   *Account   `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`

		TransactionID *uuid.UUID   `gorm:"type:uuid" json:"transaction_id,omitempty"`
		Transaction   *Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction,omitempty"`

		TransactionBatchID *uuid.UUID        `gorm:"type:uuid;index:idx_transaction_batch_entry" json:"transaction_batch_id,omitempty"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`

		EmployeeUserID *uuid.UUID `gorm:"type:uuid" json:"employee_user_id,omitempty"`
		EmployeeUser   *User      `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid;index:idx_org_branch_account_member" json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		MemberJointAccountID *uuid.UUID          `gorm:"type:uuid" json:"member_joint_account_id,omitempty"`
		MemberJointAccount   *MemberJointAccount `gorm:"foreignKey:MemberJointAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_joint_account,omitempty"`

		TransactionReferenceNumber string `gorm:"type:varchar(50)" json:"transaction_reference_number"`
		ReferenceNumber            string `gorm:"type:varchar(50)" json:"reference_number"`

		PaymentTypeID *uuid.UUID   `gorm:"type:uuid" json:"payment_type_id,omitempty"`
		PaymentType   *PaymentType `gorm:"foreignKey:PaymentTypeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"payment_type,omitempty"`

		Source GeneralLedgerSource `gorm:"type:varchar(20)" json:"source"`

		JournalVoucherID *uuid.UUID `gorm:"type:uuid" json:"journal_voucher_id,omitempty"`

		AdjustmentEntryID *uuid.UUID       `gorm:"type:uuid" json:"adjustment_entry_id,omitempty"`
		AdjustmentEntry   *AdjustmentEntry `gorm:"foreignKey:AdjustmentEntryID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"adjustment_entry,omitempty"`

		LoanTransactionID *uuid.UUID       `gorm:"type:uuid" json:"loan_transaction_id,omitempty"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		TypeOfPaymentType TypeOfPaymentType `gorm:"type:varchar(20)" json:"type_of_payment_type,omitempty"`

		Credit  float64 `gorm:"type:decimal" json:"credit"`
		Debit   float64 `gorm:"type:decimal" json:"debit"`
		Balance float64 `gorm:"type:decimal" json:"balance"`

		SignatureMediaID *uuid.UUID `gorm:"type:uuid" json:"signature_media_id,omitempty"`
		SignatureMedia   *Media     `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"signature_media,omitempty"`

		EntryDate time.Time `gorm:"type:timestamp;not null;default:now()" json:"entry_date"`

		BankID *uuid.UUID `gorm:"type:uuid" json:"bank_id,omitempty"`
		Bank   *Bank      `gorm:"foreignKey:BankID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"bank,omitempty"`

		ProofOfPaymentMediaID *uuid.UUID `gorm:"type:uuid" json:"proof_of_payment_media_id,omitempty"`
		ProofOfPaymentMedia   *Media     `gorm:"foreignKey:ProofOfPaymentMediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"proof_of_payment_media,omitempty"`

		CurrencyID *uuid.UUID `gorm:"type:uuid" json:"currency_id,omitempty"`
		Currency   *Currency  `gorm:"foreignKey:CurrencyID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"currency,omitempty"`

		BankReferenceNumber string `gorm:"type:varchar(50)" json:"bank_reference_number"`
		Description         string `gorm:"type:text" json:"description"`
		PrintNumber         int    `gorm:"default:0" json:"print_number"`
	}

	GeneralLedgerResponse struct {
		ID             uuid.UUID             `json:"id"`
		CreatedAt      string                `json:"created_at"`
		CreatedByID    uuid.UUID             `json:"created_by_id"`
		CreatedBy      *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt      string                `json:"updated_at"`
		UpdatedByID    uuid.UUID             `json:"updated_by_id"`
		UpdatedBy      *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       uuid.UUID             `json:"branch_id"`
		Branch         *BranchResponse       `json:"branch,omitempty"`

		AccountID            *uuid.UUID                  `json:"account_id,omitempty"`
		Account              *AccountResponse            `json:"account,omitempty"`
		TransactionID        *uuid.UUID                  `json:"transaction_id,omitempty"`
		Transaction          *TransactionResponse        `json:"transaction,omitempty"`
		TransactionBatchID   *uuid.UUID                  `json:"transaction_batch_id,omitempty"`
		TransactionBatch     *TransactionBatchResponse   `json:"transaction_batch,omitempty"`
		EmployeeUserID       *uuid.UUID                  `json:"employee_user_id,omitempty"`
		EmployeeUser         *UserResponse               `json:"employee_user,omitempty"`
		MemberProfileID      *uuid.UUID                  `json:"member_profile_id,omitempty"`
		MemberProfile        *MemberProfileResponse      `json:"member_profile,omitempty"`
		MemberJointAccountID *uuid.UUID                  `json:"member_joint_account_id,omitempty"`
		MemberJointAccount   *MemberJointAccountResponse `json:"member_joint_account,omitempty"`

		TransactionReferenceNumber string `json:"transaction_reference_number"`
		ReferenceNumber            string `json:"reference_number"`

		PaymentTypeID *uuid.UUID           `json:"payment_type_id,omitempty"`
		PaymentType   *PaymentTypeResponse `json:"payment_type,omitempty"`

		Source            GeneralLedgerSource      `json:"source"`
		JournalVoucherID  *uuid.UUID               `json:"journal_voucher_id,omitempty"`
		AdjustmentEntryID *uuid.UUID               `json:"adjustment_entry_id,omitempty"`
		AdjustmentEntry   *AdjustmentEntryResponse `json:"adjustment_entry,omitempty"`
		LoanTransactionID *uuid.UUID               `json:"loan_transaction_id,omitempty"`
		LoanTransaction   *LoanTransactionResponse `json:"loan_transaction,omitempty"`
		TypeOfPaymentType TypeOfPaymentType        `json:"type_of_payment_type"`

		Credit  float64 `json:"credit"`
		Debit   float64 `json:"debit"`
		Balance float64 `json:"balance"`

		SignatureMediaID *uuid.UUID     `json:"signature_media_id,omitempty"`
		SignatureMedia   *MediaResponse `json:"signature_media,omitempty"`

		EntryDate string `json:"entry_date,omitempty"`

		BankID *uuid.UUID    `json:"bank_id,omitempty"`
		Bank   *BankResponse `json:"bank,omitempty"`

		ProofOfPaymentMediaID *uuid.UUID     `json:"proof_of_payment_media_id,omitempty"`
		ProofOfPaymentMedia   *MediaResponse `json:"proof_of_payment_media,omitempty"`

		CurrencyID *uuid.UUID        `json:"currency_id,omitempty"`
		Currency   *CurrencyResponse `json:"currency,omitempty"`

		BankReferenceNumber string `json:"bank_reference_number,omitempty"`

		Description string `json:"description,omitempty"`
		PrintNumber int    `json:"print_number"`

		AccountHistoryID *uuid.UUID `json:"account_history_id"`
	}

	GeneralLedgerRequest struct {
		OrganizationID             uuid.UUID           `json:"organization_id" validate:"required"`
		BranchID                   uuid.UUID           `json:"branch_id" validate:"required"`
		AccountID                  *uuid.UUID          `json:"account_id,omitempty"`
		TransactionID              *uuid.UUID          `json:"transaction_id,omitempty"`
		TransactionBatchID         *uuid.UUID          `json:"transaction_batch_id,omitempty"`
		EmployeeUserID             *uuid.UUID          `json:"employee_user_id,omitempty"`
		MemberProfileID            *uuid.UUID          `json:"member_profile_id,omitempty"`
		MemberJointAccountID       *uuid.UUID          `json:"member_joint_account_id,omitempty"`
		TransactionReferenceNumber string              `json:"transaction_reference_number,omitempty"`
		ReferenceNumber            string              `json:"reference_number,omitempty"`
		PaymentTypeID              *uuid.UUID          `json:"payment_type_id,omitempty"`
		Source                     GeneralLedgerSource `json:"source,omitempty"`
		JournalVoucherID           *uuid.UUID          `json:"journal_voucher_id,omitempty"`
		AdjustmentEntryID          *uuid.UUID          `json:"adjustment_entry_id,omitempty"`
		LoanTransactionID          *uuid.UUID          `json:"loan_transaction_id,omitempty"`

		TypeOfPaymentType     TypeOfPaymentType `json:"type_of_payment_type,omitempty"`
		Credit                float64           `json:"credit,omitempty"`
		Debit                 float64           `json:"debit,omitempty"`
		SignatureMediaID      *uuid.UUID        `json:"signature_media_id,omitempty"`
		EntryDate             *time.Time        `json:"entry_date,omitempty"`
		BankID                *uuid.UUID        `json:"bank_id,omitempty"`
		ProofOfPaymentMediaID *uuid.UUID        `json:"proof_of_payment_media_id,omitempty"`
		CurrencyID            *uuid.UUID        `json:"currency_id,omitempty"`
		BankReferenceNumber   string            `json:"bank_reference_number,omitempty"`
		Description           string            `json:"description,omitempty"`
	}

	PaymentRequest struct {
		Amount                float64    `json:"amount" validate:"required,ne=0"`
		SignatureMediaID      *uuid.UUID `json:"signature_media_id,omitempty"`
		ProofOfPaymentMediaID *uuid.UUID `json:"proof_of_payment_media_id,omitempty"`
		BankID                *uuid.UUID `json:"bank_id,omitempty"`
		BankReferenceNumber   string     `json:"bank_reference_number,omitempty"`
		EntryDate             *time.Time `json:"entry_date,omitempty"`
		AccountID             *uuid.UUID `json:"account_id,omitempty"`
		PaymentTypeID         *uuid.UUID `json:"payment_type_id,omitempty"`
		Description           string     `json:"description,omitempty" validate:"max=255"`
		LoanTransactionID     *uuid.UUID `json:"loan_transaction_id,omitempty"`
	}

	PaymentQuickRequest struct {
		Amount                float64    `json:"amount" validate:"required,ne=0"`
		SignatureMediaID      *uuid.UUID `json:"signature_media_id,omitempty"`
		ProofOfPaymentMediaID *uuid.UUID `json:"proof_of_payment_media_id,omitempty"`
		BankID                *uuid.UUID `json:"bank_id,omitempty"`
		BankReferenceNumber   string     `json:"bank_reference_number,omitempty"`
		EntryDate             *time.Time `json:"entry_date,omitempty"`
		AccountID             *uuid.UUID `json:"account_id,omitempty"`
		PaymentTypeID         *uuid.UUID `json:"payment_type_id,omitempty"`
		Description           string     `json:"description,omitempty" validate:"max=255"`

		MemberProfileID      *uuid.UUID `json:"member_profile_id,omitempty"`
		MemberJointAccountID *uuid.UUID `json:"member_joint_account_id,omitempty"`
		ReferenceNumber      string     `json:"reference_number,omitempty"`
		ORAutoGenerated      bool       `json:"or_auto_generated,omitempty"`
		LoanTransactionID    *uuid.UUID `json:"loan_transaction_id,omitempty"`
	}

	MemberGeneralLedgerTotal struct {
		Balance     float64 `json:"balance"`
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
	}
)

func GeneralLedgerManager(service *horizon.HorizonService) *registry.Registry[GeneralLedger, GeneralLedgerResponse, GeneralLedgerRequest] {
	return registry.NewRegistry(registry.RegistryParams[GeneralLedger, GeneralLedgerResponse, GeneralLedgerRequest]{
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
		Resource: func(data *GeneralLedger) *GeneralLedgerResponse {
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
			return &GeneralLedgerResponse{
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
		Created: func(data *GeneralLedger) registry.Topics {
			return []string{}
		},
		Updated: func(data *GeneralLedger) registry.Topics {
			return []string{}
		},
		Deleted: func(data *GeneralLedger) registry.Topics {
			return []string{}
		},
	})
}

func CreateGeneralLedgerEntry(
	context context.Context, service *horizon.HorizonService, tx *gorm.DB, data *GeneralLedger,
) error {
	if data == nil {
		return eris.New("CreateGeneralLedgerEntry: data is nil")
	}

	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: data.OrganizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: data.BranchID},
		{Field: "account_id", Op: query.ModeEqual, Value: data.AccountID},
	}
	if data.Account != nil && data.Account.Type != AccountTypeOther && data.MemberProfileID != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "member_profile_id", Op: query.ModeEqual, Value: data.MemberProfileID,
		})
	}

	ledger, err := GeneralLedgerManager(service).ArrFindOneWithLock(context, tx, filters, []query.ArrFilterSortSQL{
		{Field: "created_at", Order: "DESC"},
	})

	previousBalance := decimal.NewFromFloat(0)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	} else {
		if ledger != nil {
			previousBalance = decimal.NewFromFloat(ledger.Balance)
		}
	}

	debitDecimal := decimal.NewFromFloat(data.Debit)
	creditDecimal := decimal.NewFromFloat(data.Credit)

	var balanceChange decimal.Decimal
	if data.Account == nil {
		balanceChange = debitDecimal.Sub(creditDecimal)
	} else {
		switch data.Account.GeneralLedgerType {
		case GLTypeAssets, GLTypeExpenses:
			balanceChange = debitDecimal.Sub(creditDecimal)
		case GLTypeLiabilities, GLTypeEquity, GLTypeRevenue:
			balanceChange = creditDecimal.Sub(debitDecimal)
		default:
			balanceChange = debitDecimal.Sub(creditDecimal)
		}
	}

	newBalance := previousBalance.Add(balanceChange)
	data.Balance, _ = newBalance.Float64()

	if err := GeneralLedgerManager(service).CreateWithTx(context, tx, data); err != nil {
		return eris.Wrap(err, "failed to create general ledger entry")
	}

	if data.Account != nil && data.Account.Type != AccountTypeOther && data.MemberProfileID != nil {
		_, err = MemberAccountingLedgerUpdateOrCreate(
			context,
			service,
			tx,
			data.Balance,
			MemberAccountingLedgerUpdateOrCreateParams{
				MemberProfileID: *data.MemberProfileID,
				AccountID:       *data.AccountID,
				OrganizationID:  data.OrganizationID,
				BranchID:        data.BranchID,
				UserID:          data.CreatedByID,
				DebitAmount:     data.Debit,
				CreditAmount:    data.Credit,
				LastPayTime:     data.EntryDate,
			},
		)
		if err != nil {
			return eris.Wrap(err, "failed to update or create member accounting ledger")
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
func GeneralLedgerCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID, branchID uuid.UUID) ([]*GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return GeneralLedgerManager(service).ArrFind(context, filters, nil)
}

func GeneralLedgerCurrentMemberAccount(context context.Context, service *horizon.HorizonService, memberProfileID, accountID, organizationID, branchID uuid.UUID) (*GeneralLedger, error) {
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
) ([]*GeneralLedger, error) {
	filters := []query.ArrFilterSQL{
		{Field: "transaction_id", Op: query.ModeEqual, Value: transactionID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	branchSetting, err := BranchSettingManager(service).FindOne(ctx, &BranchSetting{BranchID: branchID})
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
	paymentType *TypeOfPaymentType,
) ([]*GeneralLedger, error) {
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

	branchSetting, err := BranchSettingManager(service).FindOne(ctx, &BranchSetting{BranchID: branchID})
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
	source *GeneralLedgerSource,
) ([]*GeneralLedger, error) {
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
	branchSetting, err := BranchSettingManager(service).FindOne(ctx, &BranchSetting{BranchID: branchID})
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
	paymentType *TypeOfPaymentType,
	source *GeneralLedgerSource,
) ([]*GeneralLedger, error) {
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

	branchSetting, err := BranchSettingManager(service).FindOne(ctx, &BranchSetting{BranchID: branchID})
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

func GeneralLedgerAlignments(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*GeneralLedgerAccountsGrouping, error) {
	glGroupings, err := GeneralLedgerAccountsGroupingManager(service).Find(context, &GeneralLedgerAccountsGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to get general ledger groupings")
	}

	for _, grouping := range glGroupings {
		if grouping != nil {
			grouping.GeneralLedgerDefinitionEntries = []*GeneralLedgerDefinition{}
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

			var filteredEntries []*GeneralLedgerDefinition
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
) ([]*GeneralLedger, error) {
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
) ([]*GeneralLedger, error) {
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
) ([]*GeneralLedger, error) {
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
	paymentType TypeOfPaymentType,
) ([]*GeneralLedger, error) {
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
	source GeneralLedgerSource,
) ([]*GeneralLedger, error) {
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
) ([]*GeneralLedger, error) {
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
	result := []*GeneralLedger{}
	for _, entry := range entries {
		if entry.Account.CashAndCashEquivalence {
			continue
		}
		if !(entry.Account.Type == AccountTypeLoan ||
			entry.Account.Type == AccountTypeFines ||
			entry.Account.Type == AccountTypeInterest ||
			entry.Account.Type == AccountTypeSVFLedger) {
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
) ([]*GeneralLedger, error) {
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

	entriesByDate := make(map[string]*GeneralLedger)
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
) ([]*GeneralLedger, error) {

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
	result := make([]*GeneralLedger, 0)
	for _, item := range allData {
		if item.Source == GeneralLedgerSourcePayment || item.Source == GeneralLedgerSourceDeposit {
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
) ([]*GeneralLedger, error) {

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
	result := make([]*GeneralLedger, 0)
	for _, item := range allData {
		if item.Source == GeneralLedgerSourceWithdraw ||
			item.Source == GeneralLedgerSourceCheckVoucher ||
			item.Source == GeneralLedgerSourceLoan {
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
) ([]*GeneralLedger, error) {

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
	result := make([]*GeneralLedger, 0)
	for _, item := range allData {
		if item.Source == GeneralLedgerSourceJournalVoucher ||
			item.Source == GeneralLedgerSourceAdjustment {
			result = append(result, item)
		}
	}
	return result, nil
}
