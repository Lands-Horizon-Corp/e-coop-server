package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoanTransactionEntryType string

const (
	LoanTransactionStatic             LoanTransactionEntryType = "static"
	LoanTransactionDeduction          LoanTransactionEntryType = "deduction"
	LoanTransactionAddOn              LoanTransactionEntryType = "add-on"
	LoanTransactionAutomaticDeduction LoanTransactionEntryType = "automatic-deduction"
	LoanTransactionPrevious           LoanTransactionEntryType = "previous"
)

type (
	LoanTransactionEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_transaction_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_transaction_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		LoanTransactionID uuid.UUID        `gorm:"type:uuid;not null;index:idx_loan_transaction_entry_loan_transaction"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		Index int `gorm:"type:int;default:0" json:"index"`

		Type    LoanTransactionEntryType `gorm:"type:varchar(20);not null;default:'static'" json:"type"`
		IsAddOn bool                     `gorm:"type:boolean;not null;default:false" json:"is_add_on"`

		AccountID *uuid.UUID `gorm:"type:uuid"`
		Account   *Account   `gorm:"foreignKey:AccountID;constraint:OnDelete:SET NULL;" json:"account,omitempty"`

		AutomaticLoanDeductionID        *uuid.UUID              `gorm:"type:uuid"`
		AutomaticLoanDeduction          *AutomaticLoanDeduction `gorm:"foreignKey:AutomaticLoanDeductionID;constraint:OnDelete:SET NULL;" json:"automatic_loan_deduction,omitempty"`
		IsAutomaticLoanDeductionDeleted bool                    `gorm:"type:boolean;not null;default:false" json:"is_automatic_loan_deduction_deleted"`

		Name        string  `gorm:"type:varchar(255)" json:"name"`
		Description string  `gorm:"type:varchar(500)" json:"description"`
		Credit      float64 `gorm:"type:decimal"`
		Debit       float64 `gorm:"type:decimal"`

		Amount float64 `gorm:"type:decimal;default:0" json:"amount,omitempty"`
	}

	LoanTransactionEntryResponse struct {
		ID                              uuid.UUID                       `json:"id"`
		CreatedAt                       string                          `json:"created_at"`
		CreatedByID                     uuid.UUID                       `json:"created_by_id"`
		CreatedBy                       *UserResponse                   `json:"created_by,omitempty"`
		UpdatedAt                       string                          `json:"updated_at"`
		UpdatedByID                     uuid.UUID                       `json:"updated_by_id"`
		UpdatedBy                       *UserResponse                   `json:"updated_by,omitempty"`
		OrganizationID                  uuid.UUID                       `json:"organization_id"`
		Organization                    *OrganizationResponse           `json:"organization,omitempty"`
		BranchID                        uuid.UUID                       `json:"branch_id"`
		Branch                          *BranchResponse                 `json:"branch,omitempty"`
		LoanTransactionID               uuid.UUID                       `json:"loan_transaction_id"`
		LoanTransaction                 *LoanTransactionResponse        `json:"loan_transaction,omitempty"`
		Index                           int                             `json:"index"`
		Type                            LoanTransactionEntryType        `json:"type"`
		IsAddOn                         bool                            `json:"is_add_on"`
		AccountID                       *uuid.UUID                      `json:"account_id,omitempty"`
		Account                         *AccountResponse                `json:"account,omitempty"`
		AutomaticLoanDeductionID        *uuid.UUID                      `json:"automatic_loan_deduction_id,omitempty"`
		AutomaticLoanDeduction          *AutomaticLoanDeductionResponse `json:"automatic_loan_deduction,omitempty"`
		IsAutomaticLoanDeductionDeleted bool                            `json:"is_automatic_loan_deduction_deleted"`
		Name                            string                          `json:"name"`
		Description                     string                          `json:"description"`
		Credit                          float64                         `json:"credit"`
		Debit                           float64                         `json:"debit"`
		Amount                          float64                         `json:"amount"`
	}

	LoanTransactionEntryRequest struct {
		ID                *uuid.UUID               `json:"id"`
		LoanTransactionID uuid.UUID                `json:"loan_transaction_id" validate:"required"`
		Index             int                      `json:"index,omitempty"`
		Type              LoanTransactionEntryType `json:"type" validate:"required,oneof=static deduction add-on"`
		IsAddOn           bool                     `json:"is_add_on,omitempty"`
		AccountID         *uuid.UUID               `json:"account_id,omitempty"`
		Name              string                   `json:"name,omitempty"`
		Description       string                   `json:"description,omitempty"`
		Credit            float64                  `json:"credit,omitempty"`
		Debit             float64                  `json:"debit,omitempty"`
	}

	LoanTransactionDeductionRequest struct {
		AccountID   uuid.UUID `json:"account_id" validate:"required"`
		Amount      float64   `json:"amount"`
		Description string    `json:"description,omitempty"`
		IsAddOn     bool      `json:"is_add_on,omitempty"`
	}
)

func (m *Core) loanTransactionEntry() {
	m.Migration = append(m.Migration, &LoanTransactionEntry{})
	m.LoanTransactionEntryManager = *registry.NewRegistry(registry.RegistryParams[
		LoanTransactionEntry, LoanTransactionEntryResponse, LoanTransactionEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "LoanTransaction", "Account", "AutomaticLoanDeduction",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *LoanTransactionEntry) *LoanTransactionEntryResponse {
			if data == nil {
				return nil
			}
			return &LoanTransactionEntryResponse{
				ID:                              data.ID,
				CreatedAt:                       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                     data.CreatedByID,
				CreatedBy:                       m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                     data.UpdatedByID,
				UpdatedBy:                       m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:                  data.OrganizationID,
				Organization:                    m.OrganizationManager.ToModel(data.Organization),
				BranchID:                        data.BranchID,
				Branch:                          m.BranchManager.ToModel(data.Branch),
				LoanTransactionID:               data.LoanTransactionID,
				LoanTransaction:                 m.LoanTransactionManager.ToModel(data.LoanTransaction),
				Index:                           data.Index,
				Type:                            data.Type,
				IsAddOn:                         data.IsAddOn,
				AccountID:                       data.AccountID,
				Account:                         m.AccountManager.ToModel(data.Account),
				AutomaticLoanDeductionID:        data.AutomaticLoanDeductionID,
				AutomaticLoanDeduction:          m.AutomaticLoanDeductionManager.ToModel(data.AutomaticLoanDeduction),
				IsAutomaticLoanDeductionDeleted: data.IsAutomaticLoanDeductionDeleted,

				Name:        data.Name,
				Description: data.Description,
				Credit:      data.Credit,
				Debit:       data.Debit,
				Amount:      data.Amount,
			}
		},

		Created: func(data *LoanTransactionEntry) registry.Topics {
			return []string{
				"loan_transaction_entry.create",
				fmt.Sprintf("loan_transaction_entry.create.%s", data.ID),
				fmt.Sprintf("loan_transaction_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanTransactionEntry) registry.Topics {
			return []string{
				"loan_transaction_entry.update",
				fmt.Sprintf("loan_transaction_entry.update.%s", data.ID),
				fmt.Sprintf("loan_transaction_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanTransactionEntry) registry.Topics {
			return []string{
				"loan_transaction_entry.delete",
				fmt.Sprintf("loan_transaction_entry.delete.%s", data.ID),
				fmt.Sprintf("loan_transaction_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) LoanTransactionEntryCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*LoanTransactionEntry, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return m.LoanTransactionEntryManager.ArrFind(context, filters, nil)
}

func (m *Core) GetCashOnCashEquivalence(ctx context.Context, loanTransactionID, organizationID, branchID uuid.UUID) (*LoanTransactionEntry, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "index", Op: query.ModeEqual, Value: 0},
		{Field: "debit", Op: query.ModeEqual, Value: 0},
		{Field: "loan_transaction_id", Op: query.ModeEqual, Value: loanTransactionID},
	}

	return m.LoanTransactionEntryManager.ArrFindOne(
		ctx, filters, nil, "Account", "Account.DefaultPaymentType",
	)
}

func (m *Core) GetLoanEntryAccount(ctx context.Context, loanTransactionID, organizationID, branchID uuid.UUID) (*LoanTransactionEntry, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "index", Op: query.ModeEqual, Value: 1},
		{Field: "credit", Op: query.ModeEqual, Value: 0},
		{Field: "loan_transaction_id", Op: query.ModeEqual, Value: loanTransactionID},
	}

	return m.LoanTransactionEntryManager.ArrFindOne(ctx, filters, nil)
}
