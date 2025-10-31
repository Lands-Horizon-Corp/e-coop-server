package modelcore

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// ComakerCollateral represents the ComakerCollateral model.
	ComakerCollateral struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_comaker_collateral" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_comaker_collateral" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		LoanTransactionID uuid.UUID        `gorm:"type:uuid;not null;index:idx_loan_transaction_comaker_collateral" json:"loan_transaction_id"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		CollateralID uuid.UUID   `gorm:"type:uuid;not null" json:"collateral_id"`
		Collateral   *Collateral `gorm:"foreignKey:CollateralID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"collateral,omitempty"`

		Amount      float64 `gorm:"type:decimal;not null" json:"amount"`
		Description string  `gorm:"type:text" json:"description"`
		MonthsCount int     `gorm:"type:int;default:0" json:"months_count"`
		YearCount   int     `gorm:"type:int;default:0" json:"year_count"`
	}

	// ComakerCollateralResponse represents the response structure for comakercollateral data

	// ComakerCollateralResponse represents the response structure for ComakerCollateral.
	ComakerCollateralResponse struct {
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

		LoanTransactionID uuid.UUID                `json:"loan_transaction_id"`
		LoanTransaction   *LoanTransactionResponse `json:"loan_transaction,omitempty"`

		CollateralID uuid.UUID           `json:"collateral_id"`
		Collateral   *CollateralResponse `json:"collateral,omitempty"`

		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		MonthsCount int     `json:"months_count"`
		YearCount   int     `json:"year_count"`
	}

	// ComakerCollateralRequest represents the request structure for creating/updating comakercollateral

	// ComakerCollateralRequest represents the request structure for ComakerCollateral.
	ComakerCollateralRequest struct {
		ID                *uuid.UUID `json:"id,omitempty"`
		LoanTransactionID uuid.UUID  `json:"loan_transaction_id" validate:"required"`
		CollateralID      uuid.UUID  `json:"collateral_id" validate:"required"`
		Amount            float64    `json:"amount" validate:"required,min=0"`
		Description       string     `json:"description,omitempty"`
		MonthsCount       int        `json:"months_count,omitempty"`
		YearCount         int        `json:"year_count,omitempty"`
	}
)

func (m *ModelCore) comakerCollateral() {
	m.Migration = append(m.Migration, &ComakerCollateral{})
	m.ComakerCollateralManager = services.NewRepository(services.RepositoryParams[ComakerCollateral, ComakerCollateralResponse, ComakerCollateralRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "LoanTransaction", "Collateral"},
		Service:  m.provider.Service,
		Resource: func(data *ComakerCollateral) *ComakerCollateralResponse {
			if data == nil {
				return nil
			}
			return &ComakerCollateralResponse{
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
				CollateralID:      data.CollateralID,
				Collateral:        m.CollateralManager.ToModel(data.Collateral),
				Amount:            data.Amount,
				Description:       data.Description,
				MonthsCount:       data.MonthsCount,
				YearCount:         data.YearCount,
			}
		},
		Created: func(data *ComakerCollateral) []string {
			return []string{
				"comaker_collateral.create",
				fmt.Sprintf("comaker_collateral.create.%s", data.ID),
				fmt.Sprintf("comaker_collateral.create.branch.%s", data.BranchID),
				fmt.Sprintf("comaker_collateral.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("comaker_collateral.create.loan_transaction.%s", data.LoanTransactionID),
			}
		},
		Updated: func(data *ComakerCollateral) []string {
			return []string{
				"comaker_collateral.update",
				fmt.Sprintf("comaker_collateral.update.%s", data.ID),
				fmt.Sprintf("comaker_collateral.update.branch.%s", data.BranchID),
				fmt.Sprintf("comaker_collateral.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("comaker_collateral.update.loan_transaction.%s", data.LoanTransactionID),
			}
		},
		Deleted: func(data *ComakerCollateral) []string {
			return []string{
				"comaker_collateral.delete",
				fmt.Sprintf("comaker_collateral.delete.%s", data.ID),
				fmt.Sprintf("comaker_collateral.delete.branch.%s", data.BranchID),
				fmt.Sprintf("comaker_collateral.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("comaker_collateral.delete.loan_transaction.%s", data.LoanTransactionID),
			}
		},
	})
}

// ComakerCollateralCurrentBranch retrieves all comaker collaterals for the specified organization and branch
func (m *ModelCore) ComakerCollateralCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*ComakerCollateral, error) {
	return m.ComakerCollateralManager.Find(context, &ComakerCollateral{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

// ComakerCollateralByLoanTransaction retrieves all comaker collaterals for the specified loan transaction
func (m *ModelCore) ComakerCollateralByLoanTransaction(context context.Context, loanTransactionId uuid.UUID) ([]*ComakerCollateral, error) {
	return m.ComakerCollateralManager.Find(context, &ComakerCollateral{
		LoanTransactionID: loanTransactionId,
	})
}
