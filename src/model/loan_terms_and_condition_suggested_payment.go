package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	LoanTermsAndConditionSuggestedPayment struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_terms_and_condition_suggested_payment"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_terms_and_condition_suggested_payment"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		LoanTransactionID uuid.UUID        `gorm:"type:uuid;not null"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`
	}

	LoanTermsAndConditionSuggestedPaymentResponse struct {
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
	}

	LoanTermsAndConditionSuggestedPaymentRequest struct {
		LoanTransactionID uuid.UUID `json:"loan_transaction_id" validate:"required"`
	}
)

func (m *Model) LoanTermsAndConditionSuggestedPayment() {
	m.Migration = append(m.Migration, &LoanTermsAndConditionSuggestedPayment{})
	m.LoanTermsAndConditionSuggestedPaymentManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		LoanTermsAndConditionSuggestedPayment, LoanTermsAndConditionSuggestedPaymentResponse, LoanTermsAndConditionSuggestedPaymentRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization", "LoanTransaction",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanTermsAndConditionSuggestedPayment) *LoanTermsAndConditionSuggestedPaymentResponse {
			if data == nil {
				return nil
			}
			return &LoanTermsAndConditionSuggestedPaymentResponse{
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
			}
		},
		Created: func(data *LoanTermsAndConditionSuggestedPayment) []string {
			return []string{
				"loan_terms_and_condition_suggested_payment.create",
				fmt.Sprintf("loan_terms_and_condition_suggested_payment.create.%s", data.ID),
			}
		},
		Updated: func(data *LoanTermsAndConditionSuggestedPayment) []string {
			return []string{
				"loan_terms_and_condition_suggested_payment.update",
				fmt.Sprintf("loan_terms_and_condition_suggested_payment.update.%s", data.ID),
			}
		},
		Deleted: func(data *LoanTermsAndConditionSuggestedPayment) []string {
			return []string{
				"loan_terms_and_condition_suggested_payment.delete",
				fmt.Sprintf("loan_terms_and_condition_suggested_payment.delete.%s", data.ID),
			}
		},
	})
}
