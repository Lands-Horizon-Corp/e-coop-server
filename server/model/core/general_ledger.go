package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
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
)

type (
	GeneralLedger struct {
		ID                         uuid.UUID           `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt                  time.Time           `gorm:"not null;default:now();index"`
		CreatedByID                uuid.UUID           `gorm:"type:uuid"`
		CreatedBy                  *User               `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt                  time.Time           `gorm:"not null;default:now()"`
		UpdatedByID                uuid.UUID           `gorm:"type:uuid"`
		UpdatedBy                  *User               `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt                  gorm.DeletedAt      `gorm:"index"`
		DeletedByID                *uuid.UUID          `gorm:"type:uuid"`
		DeletedBy                  *User               `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		OrganizationID             uuid.UUID           `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger;index:idx_org_branch_account_member;index:idx_transaction_batch_entry"`
		Organization               *Organization       `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID                   uuid.UUID           `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger;index:idx_org_branch_account_member;index:idx_transaction_batch_entry"`
		Branch                     *Branch             `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`
		AccountID                  *uuid.UUID          `gorm:"type:uuid;index:idx_org_branch_account_member"`
		Account                    *Account            `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`
		TransactionID              *uuid.UUID          `gorm:"type:uuid"`
		Transaction                *Transaction        `gorm:"foreignKey:TransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction,omitempty"`
		TransactionBatchID         *uuid.UUID          `gorm:"type:uuid;index:idx_transaction_batch_entry"`
		TransactionBatch           *TransactionBatch   `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`
		EmployeeUserID             *uuid.UUID          `gorm:"type:uuid"`
		EmployeeUser               *User               `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		MemberProfileID            *uuid.UUID          `gorm:"type:uuid;index:idx_org_branch_account_member"`
		MemberProfile              *MemberProfile      `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
		MemberJointAccountID       *uuid.UUID          `gorm:"type:uuid"`
		MemberJointAccount         *MemberJointAccount `gorm:"foreignKey:MemberJointAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_joint_account,omitempty"`
		TransactionReferenceNumber string              `gorm:"type:varchar(50)"`
		ReferenceNumber            string              `gorm:"type:varchar(50)"`
		PaymentTypeID              *uuid.UUID          `gorm:"type:uuid"`
		PaymentType                *PaymentType        `gorm:"foreignKey:PaymentTypeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"payment_type,omitempty"`
		Source                     GeneralLedgerSource `gorm:"type:varchar(20)"`
		JournalVoucherID           *uuid.UUID          `gorm:"type:uuid"`
		AdjustmentEntryID          *uuid.UUID          `gorm:"type:uuid"`
		AdjustmentEntry            *AdjustmentEntry    `gorm:"foreignKey:AdjustmentEntryID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"adjustment_entry,omitempty"`
		LoanTransactionID          *uuid.UUID          `gorm:"type:uuid"`
		LoanTransaction            *LoanTransaction    `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`
		TypeOfPaymentType          TypeOfPaymentType   `gorm:"type:varchar(20)" json:"type_of_payment_type,omitempty"`
		Credit                     float64             `gorm:"type:decimal"`
		Debit                      float64             `gorm:"type:decimal"`
		Balance                    float64             `gorm:"type:decimal"`

		SignatureMediaID      *uuid.UUID `gorm:"type:uuid"`
		SignatureMedia        *Media     `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"signature_media,omitempty"`
		EntryDate             time.Time  `gorm:"type:timestamp" json:"entry_date"`
		BankID                *uuid.UUID `gorm:"type:uuid"`
		Bank                  *Bank      `gorm:"foreignKey:BankID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"bank,omitempty"`
		ProofOfPaymentMediaID *uuid.UUID `gorm:"type:uuid"`
		ProofOfPaymentMedia   *Media     `gorm:"foreignKey:ProofOfPaymentMediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"proof_of_payment_media,omitempty"`
		CurrencyID            *uuid.UUID `gorm:"type:uuid"`
		Currency              *Currency  `gorm:"foreignKey:CurrencyID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"currency,omitempty"`
		BankReferenceNumber   string     `gorm:"type:varchar(50)"`
		Description           string     `gorm:"type:text"`
		PrintNumber           int        `gorm:"default:0"`
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

func (m *Core) generalLedger() {
	m.Migration = append(m.Migration, &GeneralLedger{})
	m.GeneralLedgerManager = registry.NewRegistry(registry.RegistryParams[
		GeneralLedger, GeneralLedgerResponse, GeneralLedgerRequest,
	]{
		Preloads: []string{
			"Account",
			"Account.Currency",
			"EmployeeUser",
			"EmployeeUser.Media",
			"MemberProfile",
			"MemberJointAccount",
			"PaymentType",
			"AdjustmentEntry",
			"SignatureMedia",
			"Bank",
			"ProofOfPaymentMedia",
			"Currency",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *GeneralLedger) *GeneralLedgerResponse {
			if data == nil {
				return nil
			}

			accountHistoryID, err := m.GetAccountHistoryLatestByTimeHistoryID(
				context.Background(), *data.AccountID, data.OrganizationID, data.BranchID, &data.CreatedAt,
			)
			if err != nil {
				return nil
			}

			return &GeneralLedgerResponse{
				ID:                         data.ID,
				EntryDate:                  data.EntryDate.Format(time.RFC3339),
				CreatedAt:                  data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                data.CreatedByID,
				CreatedBy:                  m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                  data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                data.UpdatedByID,
				UpdatedBy:                  m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:             data.OrganizationID,
				Organization:               m.OrganizationManager.ToModel(data.Organization),
				BranchID:                   data.BranchID,
				Branch:                     m.BranchManager.ToModel(data.Branch),
				AccountID:                  data.AccountID,
				Account:                    m.AccountManager.ToModel(data.Account),
				TransactionID:              data.TransactionID,
				Transaction:                m.TransactionManager.ToModel(data.Transaction),
				TransactionBatchID:         data.TransactionBatchID,
				TransactionBatch:           m.TransactionBatchManager.ToModel(data.TransactionBatch),
				EmployeeUserID:             data.EmployeeUserID,
				EmployeeUser:               m.UserManager.ToModel(data.EmployeeUser),
				MemberProfileID:            data.MemberProfileID,
				MemberProfile:              m.MemberProfileManager.ToModel(data.MemberProfile),
				MemberJointAccountID:       data.MemberJointAccountID,
				MemberJointAccount:         m.MemberJointAccountManager.ToModel(data.MemberJointAccount),
				TransactionReferenceNumber: data.TransactionReferenceNumber,
				ReferenceNumber:            data.ReferenceNumber,
				PaymentTypeID:              data.PaymentTypeID,
				PaymentType:                m.PaymentTypeManager.ToModel(data.PaymentType),
				Source:                     data.Source,
				JournalVoucherID:           data.JournalVoucherID,
				AdjustmentEntryID:          data.AdjustmentEntryID,
				AdjustmentEntry:            m.AdjustmentEntryManager.ToModel(data.AdjustmentEntry),
				LoanTransactionID:          data.LoanTransactionID,
				LoanTransaction:            m.LoanTransactionManager.ToModel(data.LoanTransaction),
				TypeOfPaymentType:          data.TypeOfPaymentType,
				Credit:                     data.Credit,
				Debit:                      data.Debit,
				SignatureMediaID:           data.SignatureMediaID,
				SignatureMedia:             m.MediaManager.ToModel(data.SignatureMedia),

				BankID:                data.BankID,
				Bank:                  m.BankManager.ToModel(data.Bank),
				ProofOfPaymentMediaID: data.ProofOfPaymentMediaID,
				ProofOfPaymentMedia:   m.MediaManager.ToModel(data.ProofOfPaymentMedia),
				CurrencyID:            data.CurrencyID,
				Currency:              m.CurrencyManager.ToModel(data.Currency),
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

func (m *Core) CreateGeneralLedgerEntry(
	context context.Context, tx *gorm.DB, data *GeneralLedger,
) error {
	if data == nil {
		return eris.New("CreateGeneralLedgerEntry: data is nil")
	}

	fmt.Printf("[DEBUG] Input → OrgID=%v BranchID=%v AccountID=%v Debit=%.2f Credit=%.2f MemberProfileID=%v\n",
		data.OrganizationID, data.BranchID, data.AccountID, data.Debit, data.Credit, data.MemberProfileID)

	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: data.OrganizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: data.BranchID},
		{Field: "account_id", Op: query.ModeEqual, Value: data.AccountID},
	}
	if data.Account != nil && data.Account.Type != AccountTypeOther && data.MemberProfileID != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "member_profile_id", Op: query.ModeEqual, Value: data.MemberProfileID,
		})
		fmt.Printf("[DEBUG] Added member filter → MemberProfileID=%v\n", *data.MemberProfileID)
	}

	ledger, err := m.GeneralLedgerManager.ArrFindOneWithLock(context, tx, filters, []query.ArrFilterSortSQL{
		{Field: "created_at", Order: "DESC"},
	})

	var previousBalance = m.provider.Service.Decimal.NewFromFloat(0)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("[DEBUG] Query error → %v\n", err)
			return err
		}
		fmt.Println("[DEBUG] No previous record (ErrRecordNotFound)")
	} else {
		if ledger != nil {
			fmt.Printf("[DEBUG] Previous ledger found → Balance=%.2f\n", ledger.Balance)
			previousBalance = m.provider.Service.Decimal.NewFromFloat(ledger.Balance)
		} else {
			fmt.Println("[DEBUG] Previous ledger is nil → starting from 0")
		}
	}

	debitDecimal := m.provider.Service.Decimal.NewFromFloat(data.Debit)
	creditDecimal := m.provider.Service.Decimal.NewFromFloat(data.Credit)
	fmt.Printf("[DEBUG] DebitDecimal=%s  CreditDecimal=%s\n", debitDecimal.String(), creditDecimal.String())

	var balanceChange = m.provider.Service.Decimal.NewFromFloat(0)
	if data.Account == nil {
		balanceChange = debitDecimal.Sub(creditDecimal)
		fmt.Println("[DEBUG] Account nil → change = Debit - Credit")
	} else {
		fmt.Printf("[DEBUG] Account GLType=%v\n", data.Account.GeneralLedgerType)
		switch data.Account.GeneralLedgerType {
		case GLTypeAssets, GLTypeExpenses:
			balanceChange = debitDecimal.Sub(creditDecimal)
			fmt.Println("[DEBUG] Asset/Expense → change = Debit - Credit")
		case GLTypeLiabilities, GLTypeEquity, GLTypeRevenue:
			balanceChange = creditDecimal.Sub(debitDecimal)
			fmt.Println("[DEBUG] Liability/Equity/Revenue → change = Credit - Debit")
		default:
			balanceChange = debitDecimal.Sub(creditDecimal)
			fmt.Println("[DEBUG] Default → change = Debit - Credit")
		}
	}

	fmt.Printf("[DEBUG] PreviousBalance=%s  BalanceChange=%s\n", previousBalance.String(), balanceChange.String())

	newBalance := previousBalance.Add(balanceChange)
	fmt.Printf("[DEBUG] NewBalance (decimal)=%s\n", newBalance.String())

	nbf, _ := newBalance.Float64()
	fmt.Printf("[DEBUG] Converted to float64 → Balance=%.8f\n", nbf)

	data.Balance = nbf

	if err := m.GeneralLedgerManager.CreateWithTx(context, tx, data); err != nil {
		fmt.Printf("[DEBUG] Create failed → %v\n", err)
		return eris.Wrap(err, "failed to create general ledger entry")
	}
	fmt.Println("[DEBUG] General ledger entry created OK")

	if data.Account != nil && data.Account.Type != AccountTypeOther && data.MemberProfileID != nil {
		fmt.Printf("[DEBUG] Updating member ledger → MemberID=%v NewBalance=%.2f\n", *data.MemberProfileID, data.Balance)
		_, err = m.MemberAccountingLedgerUpdateOrCreate(
			context,
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
			fmt.Printf("[DEBUG] Member ledger update failed → %v\n", err)
			return eris.Wrap(err, "failed to update or create member accounting ledger")
		}
		fmt.Println("[DEBUG] Member ledger updated OK")
	}

	fmt.Println("[DEBUG] CreateGeneralLedgerEntry finished")
	return nil
}

func (m *Core) GeneralLedgerPrintMaxNumber(
	ctx context.Context,
	memberProfileID, accountID, branchID, organizationID uuid.UUID,
) (int, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
	}
	return m.GeneralLedgerManager.ArrGetMaxInt(ctx, "print_number", filters)
}

func (m *Core) GeneralLedgerCurrentBranch(context context.Context, organizationID, branchID uuid.UUID) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return m.GeneralLedgerManager.ArrFind(context, filters, nil)
}

func (m *Core) GeneralLedgerCurrentMemberAccount(context context.Context, memberProfileID, accountID, organizationID, branchID uuid.UUID) (*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
	}

	return m.GeneralLedgerManager.ArrFindOne(context, filters, nil)
}

func (m *Core) GeneralLedgerExcludeCashonHand(
	ctx context.Context,
	transactionID, organizationID,
	branchID uuid.UUID,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "transaction_id", Op: query.ModeEqual, Value: transactionID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	branchSetting, err := m.BranchSettingManager.FindOne(ctx, &BranchSetting{BranchID: branchID})
	if err != nil {
		return nil, err
	}

	if branchSetting.CashOnHandAccountID != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "account_id",
			Op:    query.ModeNotEqual,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}

	return m.GeneralLedgerManager.ArrFind(ctx, filters, nil)
}

func (m *Core) GeneralLedgerExcludeCashonHandWithType(
	ctx context.Context,
	transactionID, organizationID, branchID uuid.UUID,
	paymentType *TypeOfPaymentType,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "transaction_id", Op: query.ModeEqual, Value: transactionID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	if paymentType != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "type_of_payment_type",
			Op:    query.ModeEqual,
			Value: *paymentType,
		})
	}

	branchSetting, err := m.BranchSettingManager.FindOne(ctx, &BranchSetting{BranchID: branchID})
	if err != nil {
		return nil, err
	}

	if branchSetting.CashOnHandAccountID != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "account_id",
			Op:    query.ModeNotEqual,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}

	return m.GeneralLedgerManager.ArrFind(ctx, filters, nil)
}

func (m *Core) GeneralLedgerExcludeCashonHandWithSource(
	ctx context.Context,
	transactionID, organizationID, branchID uuid.UUID,
	source *GeneralLedgerSource,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "transaction_id", Op: query.ModeEqual, Value: transactionID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}
	if source != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "source",
			Op:    query.ModeEqual,
			Value: *source,
		})
	}
	branchSetting, err := m.BranchSettingManager.FindOne(ctx, &BranchSetting{BranchID: branchID})
	if err != nil {
		return nil, err
	}
	if branchSetting.CashOnHandAccountID != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "account_id",
			Op:    query.ModeNotEqual,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}
	return m.GeneralLedgerManager.ArrFind(ctx, filters, nil)
}

func (m *Core) GeneralLedgerExcludeCashonHandWithFilters(
	ctx context.Context,
	transactionID, organizationID, branchID uuid.UUID,
	paymentType *TypeOfPaymentType,
	source *GeneralLedgerSource,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "transaction_id", Op: query.ModeEqual, Value: transactionID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	if paymentType != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "type_of_payment_type",
			Op:    query.ModeEqual,
			Value: *paymentType,
		})
	}

	if source != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "source",
			Op:    query.ModeEqual,
			Value: *source,
		})
	}

	branchSetting, err := m.BranchSettingManager.FindOne(ctx, &BranchSetting{BranchID: branchID})
	if err != nil {
		return nil, err
	}

	if branchSetting.CashOnHandAccountID != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "account_id",
			Op:    query.ModeNotEqual,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}

	return m.GeneralLedgerManager.ArrFind(ctx, filters, nil)
}

func (m *Core) GeneralLedgerAlignments(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*GeneralLedgerAccountsGrouping, error) {
	glGroupings, err := m.GeneralLedgerAccountsGroupingManager.Find(context, &GeneralLedgerAccountsGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to get general ledger groupings")
	}

	for _, grouping := range glGroupings {
		if grouping != nil {
			grouping.GeneralLedgerDefinitionEntries = []*GeneralLedgerDefinition{}
			entries, err := m.GeneralLedgerDefinitionManager.ArrFind(context,
				[]registry.FilterSQL{
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

func (m *Core) GeneralLedgerCurrentMemberAccountEntries(
	ctx context.Context,
	memberProfileID, accountID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
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
	return m.GeneralLedgerManager.ArrFind(ctx, filters, sorts)
}

func (m *Core) GeneralLedgerMemberAccountTotal(
	ctx context.Context,
	memberProfileID, accountID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "account_id", Op: query.ModeNotEqual, Value: cashOnHandAccountID},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	}
	return m.GeneralLedgerManager.ArrFind(ctx, filters, sorts)
}

func (m *Core) GeneralLedgerMemberProfileEntries(
	ctx context.Context,
	memberProfileID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeNotEqual, Value: cashOnHandAccountID},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	}
	return m.GeneralLedgerManager.ArrFind(ctx, filters, sorts)
}

func (m *Core) GeneralLedgerMemberProfileEntriesByPaymentType(
	ctx context.Context,
	memberProfileID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
	paymentType TypeOfPaymentType,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "type_of_payment_type", Op: query.ModeEqual, Value: paymentType},
		{Field: "account_id", Op: query.ModeNotEqual, Value: cashOnHandAccountID},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	}
	return m.GeneralLedgerManager.ArrFind(ctx, filters, sorts)
}

func (m *Core) GeneralLedgerMemberProfileEntriesBySource(
	ctx context.Context,
	memberProfileID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
	source GeneralLedgerSource,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "source", Op: query.ModeEqual, Value: source},
		{Field: "account_id", Op: query.ModeNotEqual, Value: cashOnHandAccountID},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	}
	return m.GeneralLedgerManager.ArrFind(ctx, filters, sorts)
}

func (m *Core) GeneralLedgerByLoanTransaction(
	ctx context.Context,
	loanTransactionID, organizationID, branchID uuid.UUID,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "loan_transaction_id", Op: query.ModeEqual, Value: loanTransactionID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	sorts := []query.ArrFilterSortSQL{
		{Field: "entry_date", Order: "DESC NULLS LAST"},
		{Field: "created_at", Order: "DESC"},
	}

	entries, err := m.GeneralLedgerManager.ArrFind(ctx, filters, sorts, "Account", "EmployeeUser", "EmployeeUser.Media")
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

func (m *Core) GetGeneralLedgerOfMemberByEndOfDay(
	ctx context.Context,
	from, to time.Time,
	accountID, memberProfileID,
	organizationID,
	branchID uuid.UUID,
) ([]*GeneralLedger, error) {
	fromStartOfDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location()).UTC()
	toEndOfDay := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location()).UTC()
	filters := []registry.FilterSQL{
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

	return m.GeneralLedgerManager.ArrFind(ctx, filters, sorts, "Account")
}
func (m *Core) GetDailyEndingBalances(
	ctx context.Context,
	from, to time.Time,
	accountID, memberProfileID, organizationID, branchID uuid.UUID,
) ([]float64, error) {

	if to.Before(from) {

		return nil, eris.New("invalid date range: 'to' date cannot be before 'from' date")
	}

	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

	entries, err := m.GetGeneralLedgerOfMemberByEndOfDay(ctx, from, to, accountID, memberProfileID, organizationID, branchID)
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

	filters := []registry.FilterSQL{
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
	lastEntry, err := m.GeneralLedgerManager.ArrFindOne(ctx, filters, sorts, "Account")
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
