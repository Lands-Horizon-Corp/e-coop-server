package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// GeneralLedgerSource represents the source type of a general ledger entry
type GeneralLedgerSource string

// General ledger source constants
const (
	GeneralLedgerSourceWithdraw       GeneralLedgerSource = "withdraw"
	GeneralLedgerSourceDeposit        GeneralLedgerSource = "deposit"
	GeneralLedgerSourceJournal        GeneralLedgerSource = "journal"
	GeneralLedgerSourcePayment        GeneralLedgerSource = "payment"
	GeneralLedgerSourceAdjustment     GeneralLedgerSource = "adjustment"
	GeneralLedgerSourceJournalVoucher GeneralLedgerSource = "journal voucher"
	GeneralLedgerSourceCheckVoucher   GeneralLedgerSource = "check voucher"
)

// Assumes you have TypeOfPaymentType defined elsewhere, as in your payment_type model

type (
	// GeneralLedger represents the GeneralLedger model.
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
		TypeOfPaymentType          TypeOfPaymentType   `gorm:"type:varchar(20)" json:"type_of_payment_type,omitempty"`
		Credit                     float64             `gorm:"type:decimal"`
		Debit                      float64             `gorm:"type:decimal"`
		Balance                    float64             `gorm:"type:decimal"`
		SignatureMediaID           *uuid.UUID          `gorm:"type:uuid"`
		SignatureMedia             *Media              `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"signature_media,omitempty"`
		EntryDate                  *time.Time          `gorm:"type:date" json:"entry_date"`
		BankID                     *uuid.UUID          `gorm:"type:uuid"`
		Bank                       *Bank               `gorm:"foreignKey:BankID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"bank,omitempty"`
		ProofOfPaymentMediaID      *uuid.UUID          `gorm:"type:uuid"`
		ProofOfPaymentMedia        *Media              `gorm:"foreignKey:ProofOfPaymentMediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"proof_of_payment_media,omitempty"`
		CurrencyID                 *uuid.UUID          `gorm:"type:uuid"`
		Currency                   *Currency           `gorm:"foreignKey:CurrencyID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"currency,omitempty"`
		BankReferenceNumber        string              `gorm:"type:varchar(50)"`
		Description                string              `gorm:"type:text"`
		PrintNumber                int                 `gorm:"default:0"`
	}

	// GeneralLedgerResponse represents the response structure for GeneralLedger.
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
		TypeOfPaymentType TypeOfPaymentType        `json:"type_of_payment_type"`

		Credit  float64 `json:"credit"`
		Debit   float64 `json:"debit"`
		Balance float64 `json:"balance"`

		SignatureMediaID *uuid.UUID     `json:"signature_media_id,omitempty"`
		SignatureMedia   *MediaResponse `json:"signature_media,omitempty"`

		EntryDate *time.Time `json:"entry_date,omitempty"`

		BankID *uuid.UUID    `json:"bank_id,omitempty"`
		Bank   *BankResponse `json:"bank,omitempty"`

		ProofOfPaymentMediaID *uuid.UUID     `json:"proof_of_payment_media_id,omitempty"`
		ProofOfPaymentMedia   *MediaResponse `json:"proof_of_payment_media,omitempty"`

		CurrencyID *uuid.UUID        `json:"currency_id,omitempty"`
		Currency   *CurrencyResponse `json:"currency,omitempty"`

		BankReferenceNumber string `json:"bank_reference_number,omitempty"`

		Description string `json:"description,omitempty"`
		PrintNumber int    `json:"print_number"`
	}

	// GeneralLedgerRequest represents the request structure for creating/updating generalledger

	// GeneralLedgerRequest represents the request structure for GeneralLedger.
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
		TypeOfPaymentType          TypeOfPaymentType   `json:"type_of_payment_type,omitempty"`
		Credit                     float64             `json:"credit,omitempty"`
		Debit                      float64             `json:"debit,omitempty"`
		Balance                    float64             `json:"balance,omitempty"`
		SignatureMediaID           *uuid.UUID          `json:"signature_media_id,omitempty"`
		EntryDate                  *time.Time          `json:"entry_date,omitempty"`
		BankID                     *uuid.UUID          `json:"bank_id,omitempty"`
		ProofOfPaymentMediaID      *uuid.UUID          `json:"proof_of_payment_media_id,omitempty"`
		CurrencyID                 *uuid.UUID          `json:"currency_id,omitempty"`
		BankReferenceNumber        string              `json:"bank_reference_number,omitempty"`
		Description                string              `json:"description,omitempty"`
	}

	// PaymentRequest represents the request structure for creating/updating payment

	// PaymentRequest represents the request structure for Payment.
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
	}

	// PaymentQuickRequest represents the request structure for creating/updating paymentquick

	// PaymentQuickRequest represents the request structure for PaymentQuick.
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
	}

	// MemberGeneralLedgerTotal represents the MemberGeneralLedgerTotal model.
	MemberGeneralLedgerTotal struct {
		Balance     float64 `json:"balance"`
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
	}
)

func (m *Core) generalLedger() {
	m.Migration = append(m.Migration, &GeneralLedger{})
	m.GeneralLedgerManager = *registry.NewRegistry(registry.RegistryParams[
		GeneralLedger, GeneralLedgerResponse, GeneralLedgerRequest,
	]{
		Preloads: []string{
			"Account",
			"EmployeeUser",
			"MemberProfile",
			"MemberJointAccount",
			"PaymentType",
			"AdjustmentEntry",
			"SignatureMedia",
			"Bank",
			"ProofOfPaymentMedia",
			"Currency",
		},
		Service: m.provider.Service,
		Resource: func(data *GeneralLedger) *GeneralLedgerResponse {
			if data == nil {
				return nil
			}
			return &GeneralLedgerResponse{
				ID:                         data.ID,
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
				TypeOfPaymentType:          data.TypeOfPaymentType,
				Credit:                     data.Credit,
				Debit:                      data.Debit,
				Balance:                    data.Balance,

				SignatureMediaID:      data.SignatureMediaID,
				SignatureMedia:        m.MediaManager.ToModel(data.SignatureMedia),
				EntryDate:             data.EntryDate,
				BankID:                data.BankID,
				Bank:                  m.BankManager.ToModel(data.Bank),
				ProofOfPaymentMediaID: data.ProofOfPaymentMediaID,
				ProofOfPaymentMedia:   m.MediaManager.ToModel(data.ProofOfPaymentMedia),
				CurrencyID:            data.CurrencyID,
				Currency:              m.CurrencyManager.ToModel(data.Currency),
				BankReferenceNumber:   data.BankReferenceNumber,
				Description:           data.Description,
				PrintNumber:           data.PrintNumber,
			}
		},
		Created: func(data *GeneralLedger) []string {
			return []string{
				"general_ledger.create",
				fmt.Sprintf("general_ledger.create.%s", data.ID),
				fmt.Sprintf("general_ledger.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *GeneralLedger) []string {
			return []string{
				"general_ledger.update",
				fmt.Sprintf("general_ledger.update.%s", data.ID),
				fmt.Sprintf("general_ledger.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *GeneralLedger) []string {
			return []string{
				"general_ledger.delete",
				fmt.Sprintf("general_ledger.delete.%s", data.ID),
				fmt.Sprintf("general_ledger.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// GeneralLedgerCurrentMemberAccountForUpdate retrieves the general ledger entry for a member account with row locking for updates
func (m *Core) GeneralLedgerCurrentMemberAccountForUpdate(
	ctx context.Context, tx *gorm.DB, memberProfileID, accountID, organizationID, branchID uuid.UUID,
) (*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "account_id", Op: registry.OpEq, Value: accountID},
		{Field: "member_profile_id", Op: registry.OpEq, Value: memberProfileID},
	}
	sorts := []registry.FilterSortSQL{
		{Field: "entry_date", Order: "DESC NULLS LAST"},
		{Field: "created_at", Order: "DESC"},
	}
	ledger, err := m.GeneralLedgerManager.FindOneWithSQLLock(ctx, tx, filters, sorts)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ledger, nil
}

// GeneralLedgerCurrentSubsidiaryAccountForUpdate retrieves the general ledger entry for a subsidiary account with row locking for updates
func (m *Core) GeneralLedgerCurrentSubsidiaryAccountForUpdate(
	ctx context.Context, tx *gorm.DB, accountID, organizationID, branchID uuid.UUID,
) (*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "account_id", Op: registry.OpEq, Value: accountID},
		{Field: "member_profile_id", Op: registry.OpIsNull, Value: nil},
	}
	sorts := []registry.FilterSortSQL{
		{Field: "entry_date", Order: "DESC NULLS LAST"},
		{Field: "created_at", Order: "DESC"},
	}
	ledger, err := m.GeneralLedgerManager.FindOneWithSQLLock(ctx, tx, filters, sorts)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ledger, nil
}

// GeneralLedgerCashOnHandOnUpdate retrieves the general ledger entry for a cash on hand account with row locking for updates
func (m *Core) GeneralLedgerCashOnHandOnUpdate(
	ctx context.Context, tx *gorm.DB, accountID, organizationID, branchID uuid.UUID,
) (*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "account_id", Op: registry.OpEq, Value: accountID},
	}

	sorts := []registry.FilterSortSQL{
		{Field: "entry_date", Order: "DESC NULLS LAST"},
		{Field: "created_at", Order: "DESC"},
	}

	ledger, err := m.GeneralLedgerManager.FindOneWithSQLLock(ctx, tx, filters, sorts)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ledger, nil
}

// GeneralLedgerPrintMaxNumber retrieves the maximum print number for a member's account ledger entries
// GeneralLedgerPrintMaxNumber retrieves the maximum print number for a member's account ledger entries
func (m *Core) GeneralLedgerPrintMaxNumber(
	ctx context.Context,
	memberProfileID, accountID, branchID, organizationID uuid.UUID,
) (int, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: registry.OpEq, Value: memberProfileID},
		{Field: "account_id", Op: registry.OpEq, Value: accountID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
	}
	return m.GeneralLedgerManager.GetMax(ctx, "print_number", filters)
}

// GeneralLedgerCurrentBranch retrieves general ledger entries for the current branch
func (m *Core) GeneralLedgerCurrentBranch(context context.Context, organizationID, branchID uuid.UUID) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
	}

	return m.GeneralLedgerManager.FindWithSQL(context, filters, nil)
}

// GeneralLedgerCurrentMemberAccount retrieves the general ledger entry for a specific member account
func (m *Core) GeneralLedgerCurrentMemberAccount(context context.Context, memberProfileID, accountID, organizationID, branchID uuid.UUID) (*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "account_id", Op: registry.OpEq, Value: accountID},
		{Field: "member_profile_id", Op: registry.OpEq, Value: memberProfileID},
	}

	return m.GeneralLedgerManager.FindOneWithSQL(context, filters, nil)
}

// GeneralLedgerExcludeCashonHand retrieves general ledger entries excluding cash on hand accounts
func (m *Core) GeneralLedgerExcludeCashonHand(
	ctx context.Context,
	transactionID, organizationID,
	branchID uuid.UUID,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "transaction_id", Op: registry.OpEq, Value: transactionID},
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
	}

	branchSetting, err := m.BranchSettingManager.FindOne(ctx, &BranchSetting{BranchID: branchID})
	if err != nil {
		return nil, err
	}

	if branchSetting.CashOnHandAccountID != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "account_id",
			Op:    registry.OpNe,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}

	return m.GeneralLedgerManager.FindWithSQL(ctx, filters, nil)
}

// GeneralLedgerExcludeCashonHandWithType retrieves general ledger entries excluding cash on hand accounts with payment type filter
func (m *Core) GeneralLedgerExcludeCashonHandWithType(
	ctx context.Context,
	transactionID, organizationID, branchID uuid.UUID,
	paymentType *TypeOfPaymentType,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "transaction_id", Op: registry.OpEq, Value: transactionID},
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
	}

	// Add payment type filter if provided
	if paymentType != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "type_of_payment_type",
			Op:    registry.OpEq,
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
			Op:    registry.OpNe,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}

	return m.GeneralLedgerManager.FindWithSQL(ctx, filters, nil)
}

// GeneralLedgerExcludeCashonHandWithSource retrieves general ledger entries excluding cash on hand accounts with source filter
func (m *Core) GeneralLedgerExcludeCashonHandWithSource(
	ctx context.Context,
	transactionID, organizationID, branchID uuid.UUID,
	source *GeneralLedgerSource,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "transaction_id", Op: registry.OpEq, Value: transactionID},
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
	}
	// Add source filter if provided
	if source != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "source",
			Op:    registry.OpEq,
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
			Op:    registry.OpNe,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}
	return m.GeneralLedgerManager.FindWithSQL(ctx, filters, nil)
}

// GeneralLedgerExcludeCashonHandWithFilters retrieves general ledger entries excluding cash on hand accounts with payment type and source filters
func (m *Core) GeneralLedgerExcludeCashonHandWithFilters(
	ctx context.Context,
	transactionID, organizationID, branchID uuid.UUID,
	paymentType *TypeOfPaymentType,
	source *GeneralLedgerSource,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "transaction_id", Op: registry.OpEq, Value: transactionID},
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
	}

	// Add payment type filter if provided
	if paymentType != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "type_of_payment_type",
			Op:    registry.OpEq,
			Value: *paymentType,
		})
	}

	// Add source filter if provided
	if source != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "source",
			Op:    registry.OpEq,
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
			Op:    registry.OpNe,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}

	return m.GeneralLedgerManager.FindWithSQL(ctx, filters, nil)
}

// GeneralLedgerAlignments retrieves general ledger groupings with their definition entries for a given organization and branch
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
			entries, err := m.GeneralLedgerDefinitionManager.FindWithSQL(context,
				[]registry.FilterSQL{
					{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
					{Field: "branch_id", Op: registry.OpEq, Value: branchID},
					{Field: "general_ledger_accounts_grouping_id", Op: registry.OpEq, Value: grouping.ID},
				},
				[]registry.FilterSortSQL{
					{Field: "created_at", Order: filter.SortOrderAsc},
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

// GeneralLedgerCurrentMemberAccountEntries retrieves all general ledger entries for a specific member account
func (m *Core) GeneralLedgerCurrentMemberAccountEntries(
	ctx context.Context,
	memberProfileID, accountID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: registry.OpEq, Value: memberProfileID},
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "account_id", Op: registry.OpEq, Value: accountID},
		{Field: "account_id", Op: registry.OpNe, Value: cashOnHandAccountID},
	}
	sorts := []registry.FilterSortSQL{
		{Field: "entry_date", Order: filter.SortOrderDesc},
		{Field: "created_at", Order: filter.SortOrderDesc},
	}
	return m.GeneralLedgerManager.FindWithSQL(ctx, filters, sorts)
}

// GeneralLedgerMemberAccountTotal retrieves all general ledger entries for computing totals (excludes cash on hand)
func (m *Core) GeneralLedgerMemberAccountTotal(
	ctx context.Context,
	memberProfileID, accountID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: registry.OpEq, Value: memberProfileID},
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "account_id", Op: registry.OpEq, Value: accountID},
		{Field: "account_id", Op: registry.OpNe, Value: cashOnHandAccountID},
	}
	sorts := []registry.FilterSortSQL{
		{Field: "updated_at", Order: filter.SortOrderDesc},
	}
	return m.GeneralLedgerManager.FindWithSQL(ctx, filters, sorts)
}

// GeneralLedgerMemberProfileEntries retrieves all general ledger entries for a member profile excluding cash on hand
func (m *Core) GeneralLedgerMemberProfileEntries(
	ctx context.Context,
	memberProfileID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: registry.OpEq, Value: memberProfileID},
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "account_id", Op: registry.OpNe, Value: cashOnHandAccountID},
	}
	sorts := []registry.FilterSortSQL{
		{Field: "updated_at", Order: filter.SortOrderDesc},
	}
	return m.GeneralLedgerManager.FindWithSQL(ctx, filters, sorts)
}

// GeneralLedgerMemberProfileEntriesByPaymentType retrieves all general ledger entries for a member profile by payment type, excluding cash on hand
func (m *Core) GeneralLedgerMemberProfileEntriesByPaymentType(
	ctx context.Context,
	memberProfileID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
	paymentType TypeOfPaymentType,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: registry.OpEq, Value: memberProfileID},
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "type_of_payment_type", Op: registry.OpEq, Value: paymentType},
		{Field: "account_id", Op: registry.OpNe, Value: cashOnHandAccountID},
	}
	sorts := []registry.FilterSortSQL{
		{Field: "updated_at", Order: filter.SortOrderDesc},
	}
	return m.GeneralLedgerManager.FindWithSQL(ctx, filters, sorts)
}

// GeneralLedgerMemberProfileEntriesBySource retrieves all general ledger entries for a member profile by source, excluding cash on hand
func (m *Core) GeneralLedgerMemberProfileEntriesBySource(
	ctx context.Context,
	memberProfileID, organizationID, branchID, cashOnHandAccountID uuid.UUID,
	source GeneralLedgerSource,
) ([]*GeneralLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: registry.OpEq, Value: memberProfileID},
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "source", Op: registry.OpEq, Value: source},
		{Field: "account_id", Op: registry.OpNe, Value: cashOnHandAccountID},
	}
	sorts := []registry.FilterSortSQL{
		{Field: "updated_at", Order: filter.SortOrderDesc},
	}
	return m.GeneralLedgerManager.FindWithSQL(ctx, filters, sorts)
}
