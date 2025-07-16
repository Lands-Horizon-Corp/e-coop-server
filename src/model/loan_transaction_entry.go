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

		LoanTransactionID uuid.UUID        `gorm:"type:uuid;not null"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		Description string  `gorm:"type:text"`
		Credit      float64 `gorm:"type:decimal"`
		Debit       float64 `gorm:"type:decimal"`
	}

	LoanTransactionEntryResponse struct {
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
		Description       string                   `json:"description"`
		Credit            float64                  `json:"credit"`
		Debit             float64                  `json:"debit"`
	}

	LoanTransactionEntryRequest struct {
		LoanTransactionID uuid.UUID `json:"loan_transaction_id" validate:"required"`
		Description       string    `json:"description,omitempty"`
		Credit            float64   `json:"credit,omitempty"`
		Debit             float64   `json:"debit,omitempty"`
	}
)

func (m *Model) LoanTransactionEntry() {
	m.Migration = append(m.Migration, &LoanTransactionEntry{})
	m.LoanTransactionEntryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		LoanTransactionEntry, LoanTransactionEntryResponse, LoanTransactionEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization", "LoanTransaction",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanTransactionEntry) *LoanTransactionEntryResponse {
			if data == nil {
				return nil
			}
			return &LoanTransactionEntryResponse{
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
				Description:       data.Description,
				Credit:            data.Credit,
				Debit:             data.Debit,
			}
		},

		Created: func(data *LoanTransactionEntry) []string {
			return []string{
				"loan_transaction_entry.create",
				fmt.Sprintf("loan_transaction_entry.create.%s", data.ID),
				fmt.Sprintf("loan_transaction_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanTransactionEntry) []string {
			return []string{
				"loan_transaction_entry.update",
				fmt.Sprintf("loan_transaction_entry.update.%s", data.ID),
				fmt.Sprintf("loan_transaction_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanTransactionEntry) []string {
			return []string{
				"loan_transaction_entry.delete",
				fmt.Sprintf("loan_transaction_entry.delete.%s", data.ID),
				fmt.Sprintf("loan_transaction_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) LoanTransactionEntryCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*LoanTransactionEntry, error) {
	return m.LoanTransactionEntryManager.Find(context, &LoanTransactionEntry{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
