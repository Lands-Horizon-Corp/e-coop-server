package modelcore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	MemberAccountingLedgerSummary struct {
		TotalDeposits                     float64 `json:"total_deposits"`
		TotalShareCapitalPlusFixedSavings float64 `json:"total_share_capital_plus_fixed_savings"`
		TotalLoans                        float64 `json:"total_loans"`
	}

	MemberAccountingLedgerAccountSummary struct {
		Balance     float64 `json:"balance"`
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
	}
)

func (m *ModelCore) memberAccountingLedger() {
	m.Migration = append(m.Migration, &MemberAccountingLedger{})
	m.MemberAccountingLedgerManager = services.NewRepository(services.RepositoryParams[
		MemberAccountingLedger, MemberAccountingLedgerResponse, MemberAccountingLedgerRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Account", "MemberProfile", "Account",
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

func (m *ModelCore) memberAccountingLedgerCurrentbranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberAccountingLedger, error) {
	return m.MemberAccountingLedgerManager.Find(context, &MemberAccountingLedger{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}

// MemberAccountingLedgerFindForUpdate finds and locks a member accounting ledger for concurrent protection
// Returns nil if not found (without error), allowing for create-or-update patterns
func (m *ModelCore) memberAccountingLedgerFindForUpdate(ctx context.Context, tx *gorm.DB, memberProfileID, accountID, orgID, branchID uuid.UUID) (*MemberAccountingLedger, error) {
	var ledger MemberAccountingLedger
	err := tx.WithContext(ctx).
		Model(&MemberAccountingLedger{}).
		Where("member_profile_id = ? AND account_id = ? AND organization_id = ? AND branch_id = ?",
			memberProfileID, accountID, orgID, branchID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&ledger).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found, but not an error - allows create-or-update pattern
		}
		return nil, err
	}

	return &ledger, nil
}

// MemberAccountingLedgerUpdateOrCreate safely updates existing or creates new member accounting ledger
// with race condition protection and proper transaction handling
//
// This function:
// 1. Attempts to find and lock an existing ledger entry
// 2. If found, updates the balance, last pay time, and increments transaction count
// 3. If not found, creates a new ledger entry with initial values
// 4. Uses SELECT FOR UPDATE to prevent concurrent modifications
//
// Example usage:
//
//	ledger, err := m.memberAccountingLedgerUpdateOrCreate(
//	    ctx, tx, memberID, accountID, orgID, branchID, userID,
//	    newBalance, time.Now())
//	if err != nil {
//	}
func (m *ModelCore) MemberAccountingLedgerUpdateOrCreate(
	ctx context.Context,
	tx *gorm.DB,
	memberProfileID, accountID, orgID, branchID, userID uuid.UUID,
	newBalance float64,
	lastPayTime time.Time,
) (*MemberAccountingLedger, error) {
	// First, try to find and lock existing ledger
	ledger, err := m.memberAccountingLedgerFindForUpdate(ctx, tx, memberProfileID, accountID, orgID, branchID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to find member accounting ledger for update")
	}

	if ledger == nil {
		// Create new member accounting ledger
		ledger = &MemberAccountingLedger{
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     userID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     userID,
			OrganizationID:  orgID,
			BranchID:        branchID,
			MemberProfileID: memberProfileID,
			AccountID:       accountID,
			Balance:         newBalance,
			LastPay:         &lastPayTime,
			// Initialize other fields to zero
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
		ledger.Balance = newBalance
		ledger.LastPay = &lastPayTime
		ledger.UpdatedAt = time.Now().UTC()
		ledger.UpdatedByID = userID
		ledger.Count += 1 // Increment transaction count

		err = tx.WithContext(ctx).Save(ledger).Error
		if err != nil {
			return nil, eris.Wrap(err, "failed to update member accounting ledger")
		}
	}

	return ledger, nil
}
