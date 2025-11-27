package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	// MemberAccountingLedger represents a member's accounting ledger entry in the database
	MemberAccountingLedger struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt   time.Time      `gorm:"not null;default:now()"`
		CreatedByID uuid.UUID      `gorm:"type:uuid"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_accounting_ledger"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_accounting_ledger"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
		AccountID       uuid.UUID      `gorm:"type:uuid;not null"`
		Account         *Account       `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"account,omitempty"`

		Count               int        `gorm:"type:int"`
		Balance             float64    `gorm:"type:decimal"`
		Interest            float64    `gorm:"type:decimal"`
		Fines               float64    `gorm:"type:decimal"`
		Due                 float64    `gorm:"type:decimal"`
		CarriedForwardDue   float64    `gorm:"type:decimal"`
		StoredValueFacility float64    `gorm:"type:decimal"`
		PrincipalDue        float64    `gorm:"type:decimal"`
		LastPay             *time.Time `gorm:"type:timestamp"`
	}

	// MemberAccountingLedgerResponse represents the response structure for member accounting ledger data
	MemberAccountingLedgerResponse struct {
		ID                  uuid.UUID              `json:"id"`
		CreatedAt           string                 `json:"created_at"`
		CreatedByID         uuid.UUID              `json:"created_by_id"`
		CreatedBy           *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt           string                 `json:"updated_at"`
		UpdatedByID         uuid.UUID              `json:"updated_by_id"`
		UpdatedBy           *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID      uuid.UUID              `json:"organization_id"`
		Organization        *OrganizationResponse  `json:"organization,omitempty"`
		BranchID            uuid.UUID              `json:"branch_id"`
		Branch              *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID     uuid.UUID              `json:"member_profile_id"`
		MemberProfile       *MemberProfileResponse `json:"member_profile,omitempty"`
		AccountID           uuid.UUID              `json:"account_id"`
		Account             *AccountResponse       `json:"account,omitempty"`
		Count               int                    `json:"count"`
		Balance             float64                `json:"balance"`
		Interest            float64                `json:"interest"`
		Fines               float64                `json:"fines"`
		Due                 float64                `json:"due"`
		CarriedForwardDue   float64                `json:"carried_forward_due"`
		StoredValueFacility float64                `json:"stored_value_facility"`
		PrincipalDue        float64                `json:"principal_due"`
		LastPay             *string                `json:"last_pay,omitempty"`
	}

	// MemberAccountingLedgerRequest represents the request structure for member accounting ledger data
	MemberAccountingLedgerRequest struct {
		OrganizationID      uuid.UUID  `json:"organization_id" validate:"required"`
		BranchID            uuid.UUID  `json:"branch_id" validate:"required"`
		MemberProfileID     uuid.UUID  `json:"member_profile_id" validate:"required"`
		AccountID           uuid.UUID  `json:"account_id" validate:"required"`
		Count               int        `json:"count,omitempty"`
		Balance             float64    `json:"balance,omitempty"`
		Interest            float64    `json:"interest,omitempty"`
		Fines               float64    `json:"fines,omitempty"`
		Due                 float64    `json:"due,omitempty"`
		CarriedForwardDue   float64    `json:"carried_forward_due,omitempty"`
		StoredValueFacility float64    `json:"stored_value_facility,omitempty"`
		PrincipalDue        float64    `json:"principal_due,omitempty"`
		LastPay             *time.Time `json:"last_pay,omitempty"`
	}

	MemberAccountingLedgerUpdateOrCreateParams struct {
		MemberProfileID uuid.UUID `validate:"required"`
		AccountID       uuid.UUID `validate:"required"`
		OrganizationID  uuid.UUID `validate:"required"`
		BranchID        uuid.UUID `validate:"required"`
		UserID          uuid.UUID `validate:"required"`
		DebitAmount     float64
		CreditAmount    float64
		LastPayTime     time.Time `validate:"required"`
	}

	// MemberAccountingLedgerAccountSummary represents an account summary for member accounting ledger
	MemberAccountingLedgerAccountSummary struct {
		Balance     float64 `json:"balance"`
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
	}

	MemberAccountingLedgerBrowseReference struct {
		MemberAccountingLedger *MemberAccountingLedger
		BrowseReference        *BrowseReference
	}
)

func (m *Core) memberAccountingLedger() {
	m.Migration = append(m.Migration, &MemberAccountingLedger{})
	m.MemberAccountingLedgerManager = *registry.NewRegistry(registry.RegistryParams[
		MemberAccountingLedger, MemberAccountingLedgerResponse, MemberAccountingLedgerRequest,
	]{
		Preloads: []string{
			"MemberProfile",
			"Account.Currency",
		},
		Service: m.provider.Service,
		Resource: func(data *MemberAccountingLedger) *MemberAccountingLedgerResponse {
			if data == nil {
				return nil
			}
			var lastPay *string
			if data.LastPay != nil {
				s := data.LastPay.Format(time.RFC3339)
				lastPay = &s
			}
			return &MemberAccountingLedgerResponse{
				ID:                  data.ID,
				CreatedAt:           data.CreatedAt.Format(time.RFC3339),
				CreatedByID:         data.CreatedByID,
				CreatedBy:           m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:           data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:         data.UpdatedByID,
				UpdatedBy:           m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:      data.OrganizationID,
				Organization:        m.OrganizationManager.ToModel(data.Organization),
				BranchID:            data.BranchID,
				Branch:              m.BranchManager.ToModel(data.Branch),
				MemberProfileID:     data.MemberProfileID,
				MemberProfile:       m.MemberProfileManager.ToModel(data.MemberProfile),
				AccountID:           data.AccountID,
				Account:             m.AccountManager.ToModel(data.Account),
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

		Created: func(data *MemberAccountingLedger) []string {
			return []string{
				"member_accounting_ledger.create",
				fmt.Sprintf("member_accounting_ledger.create.%s", data.ID),
				fmt.Sprintf("member_accounting_ledger.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_accounting_ledger.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_accounting_ledger.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *MemberAccountingLedger) []string {
			return []string{
				"member_accounting_ledger.update",
				fmt.Sprintf("member_accounting_ledger.update.%s", data.ID),
				fmt.Sprintf("member_accounting_ledger.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_accounting_ledger.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_accounting_ledger.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *MemberAccountingLedger) []string {
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

// MemberAccountingLedgerCurrentBranch retrieves member accounting ledgers for the current branch
func (m *Core) MemberAccountingLedgerCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberAccountingLedger, error) {
	return m.MemberAccountingLedgerManager.Find(context, &MemberAccountingLedger{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

// MemberAccountingLedgerMemberProfileEntries retrieves member accounting ledger entries for a specific member profile
// excluding the cash on hand account
func (m *Core) MemberAccountingLedgerMemberProfileEntries(ctx context.Context, memberProfileID, organizationID, branchID, cashOnHandAccountID uuid.UUID) ([]*MemberAccountingLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: registry.OpEq, Value: memberProfileID},
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "account_id", Op: registry.OpNe, Value: cashOnHandAccountID},
	}

	return m.MemberAccountingLedgerManager.FindWithSQL(ctx, filters, nil)
}

// MemberAccountingLedgerBranchEntries retrieves member accounting ledger entries for a specific branch
// excluding the cash on hand account
func (m *Core) MemberAccountingLedgerBranchEntries(ctx context.Context, organizationID, branchID, cashOnHandAccountID uuid.UUID) ([]*MemberAccountingLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "account_id", Op: registry.OpNe, Value: cashOnHandAccountID},
	}
	return m.MemberAccountingLedgerManager.FindWithSQL(ctx, filters, nil)
}

// MemberAccountingLedgerFindForUpdate finds and locks a member accounting ledger for concurrent protection
// Returns nil if not found (without error), allowing for create-or-update patterns
// MemberAccountingLedgerFindForUpdate returns MemberAccountingLedgerFindForUpdate for the current branch or organization where applicable.
func (m *Core) MemberAccountingLedgerFindForUpdate(
	ctx context.Context,
	tx *gorm.DB,
	memberProfileID,
	accountID,
	organizationID,
	branchID uuid.UUID,
) (*MemberAccountingLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: registry.OpEq, Value: memberProfileID},
		{Field: "account_id", Op: registry.OpEq, Value: accountID},
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
	}
	ledger, err := m.MemberAccountingLedgerManager.FindOneWithSQLLock(ctx, tx, filters, nil)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found, but not an error - allows create-or-update pattern
		}
		return nil, err
	}

	return ledger, nil
}

func (m *Core) MemberAccountingLedgerUpdateOrCreate(
	ctx context.Context,
	tx *gorm.DB,
	balance float64,
	params MemberAccountingLedgerUpdateOrCreateParams,
) (*MemberAccountingLedger, error) {
	// Validate: Either debit or credit must be non-zero, but not both
	if (params.DebitAmount == 0 && params.CreditAmount == 0) || (params.DebitAmount != 0 && params.CreditAmount != 0) {
		return nil, eris.New("exactly one of debit or credit must be non-zero")
	}

	// First, try to find and lock existing ledger
	ledger, err := m.MemberAccountingLedgerFindForUpdate(
		ctx, tx,
		params.MemberProfileID,
		params.AccountID,
		params.OrganizationID,
		params.BranchID,
	)
	if err != nil {
		return nil, eris.Wrap(err, "failed to find member accounting ledger for update")
	}

	if ledger == nil {
		// Create new member accounting ledger
		ledger = &MemberAccountingLedger{
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

		err = tx.WithContext(ctx).Create(ledger).Error
		if err != nil {
			return nil, eris.Wrap(err, "failed to create member accounting ledger")
		}
	} else {
		// Update existing member accounting ledger
		ledger.Balance = balance
		ledger.LastPay = &params.LastPayTime
		ledger.UpdatedAt = time.Now().UTC()
		ledger.UpdatedByID = params.UserID
		ledger.Count++

		err = tx.WithContext(ctx).Save(ledger).Error
		if err != nil {
			return nil, eris.Wrap(err, "failed to update member accounting ledger")
		}
	}

	return ledger, nil
}

// MemberAccountingLedgerFilterByCriteria filters member accounting ledgers based on account and member type criteria
func (m *Core) MemberAccountingLedgerFilterByCriteria(
	ctx context.Context,
	organizationID,
	branchID uuid.UUID,
	accountID,
	memberTypeID *uuid.UUID,
	includeClosedAccounts bool,
) ([]*MemberAccountingLedger, error) {
	result := []*MemberAccountingLedger{}
	memberAccountingLedger, err := m.MemberAccountingLedgerManager.FindWithSQL(ctx, []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "account_id", Op: registry.OpEq, Value: accountID},
	}, nil, "MemberProfile", "Account.Currency")
	if err != nil {
		return nil, eris.Wrap(err, "failed to find member accounting ledgers by account")
	}
	for _, ledger := range memberAccountingLedger {
		if includeClosedAccounts == false && ledger.MemberProfile.IsClosed {
			continue
		}
		if handlers.UUIDPtrEqual(ledger.MemberProfile.MemberTypeID, memberTypeID) {
			result = append(result, ledger)
		}
	}

	return result, nil
}

func (m *Core) MemberAccountingLedgerByBrowseReference(ctx context.Context, includeClosedAccounts bool, data []*BrowseReference) ([]*MemberAccountingLedgerBrowseReference, error) {
	memberAccountingLedger := []*MemberAccountingLedgerBrowseReference{}
	for _, browseRef := range data {
		ledgers, err := m.MemberAccountingLedgerFilterByCriteria(
			ctx, browseRef.OrganizationID, browseRef.BranchID,
			browseRef.AccountID,
			browseRef.MemberTypeID, includeClosedAccounts)
		if err != nil {
			return nil, eris.Wrap(err, "failed to filter member accounting ledgers by browse reference")
		}

		// Create MemberAccountingLedgerBrowseReference for each ledger
		for _, ledger := range ledgers {
			memberAccountingLedger = append(memberAccountingLedger, &MemberAccountingLedgerBrowseReference{
				MemberAccountingLedger: ledger,
				BrowseReference:        browseRef,
			})
		}
	}

	return memberAccountingLedger, nil
}
