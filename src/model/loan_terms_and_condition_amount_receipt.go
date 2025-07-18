package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	LoanTermsAndConditionAmountReceipt struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_terms_and_condition_amount_receipt"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_terms_and_condition_amount_receipt"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		LoanTransactionID uuid.UUID        `gorm:"type:uuid;not null"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid;not null"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`

		Amount float64 `gorm:"type:decimal"`
	}

	LoanTermsAndConditionAmountReceiptResponse struct {
		ID                uuid.UUID                `json:"id"`
		CreatedAt         string                   `json:"created_at"`
		CreatedByID       uuid.UUID                `json:"created_by_id"`
		CreatedBy         *UserResponse            `json:"created_by,omitempty"`
		UpdatedAt         string                   `json:"updated_at"`
		UpdatedByID       uuid.UUID                `json:"updated_by_id"`
		UpdatedBy         *UserResponse            `json:"updated_by,omitempty"`
		OrganizationID    uuid.UUID                `json:"organization_id"`
		Organization      *OrganizationResponse    `json:"organization,omitempty"`
		BranchID          uuid.UUID                `json:"branch_id"`
		Branch            *BranchResponse          `json:"branch,omitempty"`
		LoanTransactionID uuid.UUID                `json:"loan_transaction_id"`
		LoanTransaction   *LoanTransactionResponse `json:"loan_transaction,omitempty"`
		AccountID         uuid.UUID                `json:"account_id"`
		Account           *AccountResponse         `json:"account,omitempty"`
		Amount            float64                  `json:"amount"`
	}

	LoanTermsAndConditionAmountReceiptRequest struct {
		LoanTransactionID uuid.UUID `json:"loan_transaction_id" validate:"required"`
		AccountID         uuid.UUID `json:"account_id" validate:"required"`
		Amount            float64   `json:"amount"`
	}
)

func (m *Model) LoanTermsAndConditionAmountReceipt() {
	m.Migration = append(m.Migration, &LoanTermsAndConditionAmountReceipt{})
	m.LoanTermsAndConditionAmountReceiptManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		LoanTermsAndConditionAmountReceipt, LoanTermsAndConditionAmountReceiptResponse, LoanTermsAndConditionAmountReceiptRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization", "LoanTransaction", "Account",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanTermsAndConditionAmountReceipt) *LoanTermsAndConditionAmountReceiptResponse {
			if data == nil {
				return nil
			}
			return &LoanTermsAndConditionAmountReceiptResponse{
				ID:                data.ID,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				CreatedByID:       data.CreatedByID,
				CreatedBy:         m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:       data.UpdatedByID,
				UpdatedBy:         m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:    data.OrganizationID,
				Organization:      m.OrganizationManager.ToModel(data.Organization),
				BranchID:          data.BranchID,
				Branch:            m.BranchManager.ToModel(data.Branch),
				LoanTransactionID: data.LoanTransactionID,
				LoanTransaction:   m.LoanTransactionManager.ToModel(data.LoanTransaction),
				AccountID:         data.AccountID,
				Account:           m.AccountManager.ToModel(data.Account),
				Amount:            data.Amount,
			}
		},

		Created: func(data *LoanTermsAndConditionAmountReceipt) []string {
			return []string{
				"loan_terms_and_condition_amount_receipt.create",
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.create.%s", data.ID),
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanTermsAndConditionAmountReceipt) []string {
			return []string{
				"loan_terms_and_condition_amount_receipt.update",
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.update.%s", data.ID),
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanTermsAndConditionAmountReceipt) []string {
			return []string{
				"loan_terms_and_condition_amount_receipt.delete",
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.delete.%s", data.ID),
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_terms_and_condition_amount_receipt.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) LoanTermsAndConditionAmountReceiptCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*LoanTermsAndConditionAmountReceipt, error) {
	return m.LoanTermsAndConditionAmountReceiptManager.Find(context, &LoanTermsAndConditionAmountReceipt{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
