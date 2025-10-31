package modelcore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Enum for general_ledger_source
type GeneralLedgerSource string

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

	// GeneralLedgerResponse represents the response structure for generalledger data

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

func (m *ModelCore) generalLedger() {
	m.Migration = append(m.Migration, &GeneralLedger{})
	m.GeneralLedgerManager = services.NewRepository(services.RepositoryParams[
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

// General Ledger Queries
// GeneralLedgerCurrentBranch returns GeneralLedgerCurrentBranch for the current branch or organization where applicable.
func (m *ModelCore) GeneralLedgerCurrentBranch(context context.Context, orgID, branchID uuid.UUID) ([]*GeneralLedger, error) {
	return m.GeneralLedgerManager.Find(context, &GeneralLedger{
		OrganizationID: orgID,
		BranchID:       branchID,
	})
}

// General Ledger Queries with locking for update
// GeneralLedgerCurrentMemberAccount returns GeneralLedgerCurrentMemberAccount for the current branch or organization where applicable.
func (m *ModelCore) GeneralLedgerCurrentMemberAccount(context context.Context, memberProfileId, accountId, orgID, branchID uuid.UUID) (*GeneralLedger, error) {
	return m.GeneralLedgerManager.FindOne(context, &GeneralLedger{
		OrganizationID:  orgID,
		BranchID:        branchID,
		AccountID:       &accountId,
		MemberProfileID: &memberProfileId,
	})
}

// General Ledger Queries with locking for update
// GeneralLedgerCurrentMemberAccountForUpdate returns GeneralLedgerCurrentMemberAccountForUpdate for the current branch or organization where applicable.
func (m *ModelCore) GeneralLedgerCurrentMemberAccountForUpdate(
	ctx context.Context, tx *gorm.DB, memberProfileId, accountId, orgID, branchID uuid.UUID,
) (*GeneralLedger, error) {
	var ledger GeneralLedger
	err := tx.
		WithContext(ctx).
		Model(&GeneralLedger{}).
		Where("organization_id = ? AND branch_id = ? AND account_id = ? AND member_profile_id = ?", orgID, branchID, accountId, memberProfileId).
		Order("entry_date DESC NULLS LAST, created_at DESC").
		Limit(1).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Take(&ledger).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &ledger, err
}

// General Ledger Queries with locking for update
// GeneralLedgerCurrentSubsidiaryAccountForUpdate returns GeneralLedgerCurrentSubsidiaryAccountForUpdate for the current branch or organization where applicable.
func (m *ModelCore) GeneralLedgerCurrentSubsidiaryAccountForUpdate(
	ctx context.Context, tx *gorm.DB, accountId, orgID, branchID uuid.UUID,
) (*GeneralLedger, error) {
	var ledger GeneralLedger
	err := tx.
		WithContext(ctx).
		Model(&GeneralLedger{}).
		Where("organization_id = ? AND branch_id = ? AND account_id = ? AND member_profile_id IS NULL", orgID, branchID, accountId).
		Order("entry_date DESC NULLS LAST, created_at DESC").
		Limit(1).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Take(&ledger).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &ledger, err
}

// General Ledger Queries with locking for update
// GeneralLedgerCashOnHandOnUpdate returns GeneralLedgerCashOnHandOnUpdate for the current branch or organization where applicable.
func (m *ModelCore) GeneralLedgerCashOnHandOnUpdate(
	ctx context.Context, tx *gorm.DB, accountId, orgID, branchID uuid.UUID,
) (*GeneralLedger, error) {
	var ledger GeneralLedger
	err := tx.
		WithContext(ctx).
		Model(&GeneralLedger{}).
		Where("organization_id = ? AND branch_id = ? AND account_id = ?", orgID, branchID, accountId).
		Order("entry_date DESC NULLS LAST, created_at DESC").
		Limit(1).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Take(&ledger).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &ledger, err
}

// General Ledger Queries
// GeneralLedgerPrintMaxNumber returns GeneralLedgerPrintMaxNumber for the current branch or organization where applicable.
func (m *ModelCore) GeneralLedgerPrintMaxNumber(
	ctx context.Context,
	memberProfileID, accountID, branchID, orgID uuid.UUID,
) (int, error) {
	var maxPrintNumber int
	err := m.GeneralLedgerManager.Client().
		Where("member_profile_id = ? AND account_id = ? AND branch_id = ? AND organization_id = ?", memberProfileID, accountID, branchID, orgID).
		Select("COALESCE(MAX(print_number), 0)").
		Scan(&maxPrintNumber).Error
	if err != nil {
		return 0, err
	}
	return maxPrintNumber, nil
}

// General Ledger Queries excluding Cash on Hand account
// GeneralLedgerExcludeCashonHand returns GeneralLedgerExcludeCashonHand for the current branch or organization where applicable.
func (m *ModelCore) GeneralLedgerExcludeCashonHand(
	ctx context.Context,
	transactionId, orgID,
	branchID uuid.UUID,
) ([]*GeneralLedger, error) {
	filters := []services.Filter{
		{Field: "transaction_id", Op: services.OpEq, Value: transactionId},
		{Field: "organization_id", Op: services.OpEq, Value: orgID},
		{Field: "branch_id", Op: services.OpEq, Value: branchID},
	}
	branchSetting, err := m.BranchSettingManager.FindOne(ctx, &BranchSetting{BranchID: branchID})
	if err != nil {
		return nil, err
	}
	if branchSetting.CashOnHandAccountID != nil {
		filters = append(filters, services.Filter{
			Field: "account_id",
			Op:    services.OpNe,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}
	return m.GeneralLedgerManager.FindWithFilters(ctx, filters)
}

// General Ledger Queries excluding Cash on Hand account with TypeOfPaymentType filter
// GeneralLedgerExcludeCashonHandWithType returns GeneralLedgerExcludeCashonHandWithType for the current branch or organization where applicable.
func (m *ModelCore) GeneralLedgerExcludeCashonHandWithType(
	ctx context.Context,
	transactionId, orgID, branchID uuid.UUID,
	paymentType *TypeOfPaymentType,
) ([]*GeneralLedger, error) {
	filters := []services.Filter{
		{Field: "transaction_id", Op: services.OpEq, Value: transactionId},
		{Field: "organization_id", Op: services.OpEq, Value: orgID},
		{Field: "branch_id", Op: services.OpEq, Value: branchID},
	}

	// Add payment type filter if provided
	if paymentType != nil {
		filters = append(filters, services.Filter{
			Field: "type_of_payment_type",
			Op:    services.OpEq,
			Value: *paymentType,
		})
	}

	branchSetting, err := m.BranchSettingManager.FindOne(ctx, &BranchSetting{BranchID: branchID})
	if err != nil {
		return nil, err
	}
	if branchSetting.CashOnHandAccountID != nil {
		filters = append(filters, services.Filter{
			Field: "account_id",
			Op:    services.OpNe,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}
	return m.GeneralLedgerManager.FindWithFilters(ctx, filters)
}

// General Ledger Queries excluding Cash on Hand account with Source filter
// GeneralLedgerExcludeCashonHandWithSource returns GeneralLedgerExcludeCashonHandWithSource for the current branch or organization where applicable.
func (m *ModelCore) GeneralLedgerExcludeCashonHandWithSource(
	ctx context.Context,
	transactionId, orgID, branchID uuid.UUID,
	source *GeneralLedgerSource,
) ([]*GeneralLedger, error) {
	filters := []services.Filter{
		{Field: "transaction_id", Op: services.OpEq, Value: transactionId},
		{Field: "organization_id", Op: services.OpEq, Value: orgID},
		{Field: "branch_id", Op: services.OpEq, Value: branchID},
	}

	// Add source filter if provided
	if source != nil {
		filters = append(filters, services.Filter{
			Field: "source",
			Op:    services.OpEq,
			Value: *source,
		})
	}

	branchSetting, err := m.BranchSettingManager.FindOne(ctx, &BranchSetting{BranchID: branchID})
	if err != nil {
		return nil, err
	}
	if branchSetting.CashOnHandAccountID != nil {
		filters = append(filters, services.Filter{
			Field: "account_id",
			Op:    services.OpNe,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}
	return m.GeneralLedgerManager.FindWithFilters(ctx, filters)
}

// General Ledger Queries excluding Cash on Hand account with TypeOfPaymentType and Source filters
// GeneralLedgerExcludeCashonHandWithFilters returns GeneralLedgerExcludeCashonHandWithFilters for the current branch or organization where applicable.
func (m *ModelCore) GeneralLedgerExcludeCashonHandWithFilters(
	ctx context.Context,
	transactionId, orgID, branchID uuid.UUID,
	paymentType *TypeOfPaymentType,
	source *GeneralLedgerSource,
) ([]*GeneralLedger, error) {
	filters := []services.Filter{
		{Field: "transaction_id", Op: services.OpEq, Value: transactionId},
		{Field: "organization_id", Op: services.OpEq, Value: orgID},
		{Field: "branch_id", Op: services.OpEq, Value: branchID},
	}

	// Add payment type filter if provided
	if paymentType != nil {
		filters = append(filters, services.Filter{
			Field: "type_of_payment_type",
			Op:    services.OpEq,
			Value: *paymentType,
		})
	}

	// Add source filter if provided
	if source != nil {
		filters = append(filters, services.Filter{
			Field: "source",
			Op:    services.OpEq,
			Value: *source,
		})
	}

	branchSetting, err := m.BranchSettingManager.FindOne(ctx, &BranchSetting{BranchID: branchID})
	if err != nil {
		return nil, err
	}
	if branchSetting.CashOnHandAccountID != nil {
		filters = append(filters, services.Filter{
			Field: "account_id",
			Op:    services.OpNe,
			Value: *branchSetting.CashOnHandAccountID,
		})
	}
	return m.GeneralLedgerManager.FindWithFilters(ctx, filters)
}
