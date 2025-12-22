package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	LoanGuaranteedFund struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_guaranteed_fund"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_guaranteed_fund"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		SchemeNumber   int     `gorm:"not null;unique"`
		IncreasingRate float64 `gorm:"type:decimal;not null"`
	}

	LoanGuaranteedFundResponse struct {
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
		SchemeNumber   int                   `json:"scheme_number"`
		IncreasingRate float64               `json:"increasing_rate"`
	}

	LoanGuaranteedFundRequest struct {
		SchemeNumber   int     `json:"scheme_number" validate:"required"`
		IncreasingRate float64 `json:"increasing_rate" validate:"required"`
	}
)

func (m *Core) loanGuaranteedFund() {
	m.Migration = append(m.Migration, &LoanGuaranteedFund{})
	m.LoanGuaranteedFundManager().= registry.NewRegistry(registry.RegistryParams[
		LoanGuaranteedFund, LoanGuaranteedFundResponse, LoanGuaranteedFundRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *LoanGuaranteedFund) *LoanGuaranteedFundResponse {
			if data == nil {
				return nil
			}
			return &LoanGuaranteedFundResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager().ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager().ToModel(data.Branch),
				SchemeNumber:   data.SchemeNumber,
				IncreasingRate: data.IncreasingRate,
			}
		},

		Created: func(data *LoanGuaranteedFund) registry.Topics {
			return []string{
				"loan_guaranteed_fund.create",
				fmt.Sprintf("loan_guaranteed_fund.create.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanGuaranteedFund) registry.Topics {
			return []string{
				"loan_guaranteed_fund.update",
				fmt.Sprintf("loan_guaranteed_fund.update.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanGuaranteedFund) registry.Topics {
			return []string{
				"loan_guaranteed_fund.delete",
				fmt.Sprintf("loan_guaranteed_fund.delete.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) LoanGuaranteedFundCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*LoanGuaranteedFund, error) {
	return m.LoanGuaranteedFundManager().Find(context, &LoanGuaranteedFund{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
