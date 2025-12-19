package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
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
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
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

		Created: func(data *MemberAccountingLedger) registry.Topics {
			return []string{
				"member_accounting_ledger.create",
				fmt.Sprintf("member_accounting_ledger.create.%s", data.ID),
				fmt.Sprintf("member_accounting_ledger.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_accounting_ledger.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_accounting_ledger.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *MemberAccountingLedger) registry.Topics {
			return []string{
				"member_accounting_ledger.update",
				fmt.Sprintf("member_accounting_ledger.update.%s", data.ID),
				fmt.Sprintf("member_accounting_ledger.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_accounting_ledger.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_accounting_ledger.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *MemberAccountingLedger) registry.Topics {
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

func (m *Core) MemberAccountingLedgerCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberAccountingLedger, error) {
	return m.MemberAccountingLedgerManager.Find(context, &MemberAccountingLedger{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func (m *Core) MemberAccountingLedgerMemberProfileEntries(ctx context.Context, memberProfileID, organizationID, branchID, cashOnHandAccountID uuid.UUID) ([]*MemberAccountingLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeNotEqual, Value: cashOnHandAccountID},
	}

	return m.MemberAccountingLedgerManager.ArrFind(ctx, filters, nil)
}

func (m *Core) MemberAccountingLedgerBranchEntries(ctx context.Context, organizationID, branchID, cashOnHandAccountID uuid.UUID) ([]*MemberAccountingLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeNotEqual, Value: cashOnHandAccountID},
	}
	return m.MemberAccountingLedgerManager.ArrFind(ctx, filters, nil)
}

func (m *Core) MemberAccountingLedgerFindForUpdate(
	ctx context.Context,
	tx *gorm.DB,
	memberProfileID,
	accountID,
	organizationID,
	branchID uuid.UUID,
) (*MemberAccountingLedger, error) {
	filters := []registry.FilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}
	ledger, err := m.MemberAccountingLedgerManager.ArrFindOneWithLock(ctx, tx, filters, nil)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
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
	if (params.DebitAmount == 0 && params.CreditAmount == 0) || (params.DebitAmount != 0 && params.CreditAmount != 0) {
		return nil, eris.New("exactly one of debit or credit must be non-zero")
	}
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

	// Create new ledger if not found
	if ledger == nil {
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

		if tx == nil {
			return nil, eris.New("database tx is nil")
		}
		err = tx.WithContext(ctx).Create(ledger).Error
		if err != nil {
			return nil, eris.Wrap(err, "failed to create member accounting ledger")
		}
	} else {
		// Update existing ledger
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

func (m *Core) MemberAccountingLedgerFilterByCriteria(
	ctx context.Context,
	organizationID,
	branchID uuid.UUID,
	accountID,
	memberTypeID *uuid.UUID,
	includeClosedAccounts bool,
) ([]*MemberAccountingLedger, error) {
	result := []*MemberAccountingLedger{}
	memberAccountingLedger, err := m.MemberAccountingLedgerManager.ArrFind(ctx, []registry.FilterSQL{
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

		for _, ledger := range ledgers {
			memberAccountingLedger = append(memberAccountingLedger, &MemberAccountingLedgerBrowseReference{
				MemberAccountingLedger: ledger,
				BrowseReference:        browseRef,
			})
		}
	}

	return memberAccountingLedger, nil
}
