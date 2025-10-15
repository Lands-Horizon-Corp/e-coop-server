package model_core

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	LoanGuaranteedFundPerMonth struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_guaranteed_fund_per_month"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_guaranteed_fund_per_month"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Month              int `gorm:"type:int;default:0"`
		LoanGuaranteedFund int `gorm:"type:int;default:0"`
	}

	LoanGuaranteedFundPerMonthResponse struct {
		ID                 uuid.UUID             `json:"id"`
		CreatedAt          string                `json:"created_at"`
		CreatedByID        uuid.UUID             `json:"created_by_id"`
		CreatedBy          *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt          string                `json:"updated_at"`
		UpdatedByID        uuid.UUID             `json:"updated_by_id"`
		UpdatedBy          *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID     uuid.UUID             `json:"organization_id"`
		Organization       *OrganizationResponse `json:"organization,omitempty"`
		BranchID           uuid.UUID             `json:"branch_id"`
		Branch             *BranchResponse       `json:"branch,omitempty"`
		Month              int                   `json:"month"`
		LoanGuaranteedFund int                   `json:"loan_guaranteed_fund"`
	}

	LoanGuaranteedFundPerMonthRequest struct {
		Month              int `json:"month,omitempty"`
		LoanGuaranteedFund int `json:"loan_guaranteed_fund,omitempty"`
	}
)

func (m *ModelCore) LoanGuaranteedFundPerMonth() {
	m.Migration = append(m.Migration, &LoanGuaranteedFundPerMonth{})
	m.LoanGuaranteedFundPerMonthManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		LoanGuaranteedFundPerMonth, LoanGuaranteedFundPerMonthResponse, LoanGuaranteedFundPerMonthRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanGuaranteedFundPerMonth) *LoanGuaranteedFundPerMonthResponse {
			if data == nil {
				return nil
			}
			return &LoanGuaranteedFundPerMonthResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       m.OrganizationManager.ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             m.BranchManager.ToModel(data.Branch),
				Month:              data.Month,
				LoanGuaranteedFund: data.LoanGuaranteedFund,
			}
		},

		Created: func(data *LoanGuaranteedFundPerMonth) []string {
			return []string{
				"loan_guaranteed_fund_per_month.create",
				fmt.Sprintf("loan_guaranteed_fund_per_month.create.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanGuaranteedFundPerMonth) []string {
			return []string{
				"loan_guaranteed_fund_per_month.update",
				fmt.Sprintf("loan_guaranteed_fund_per_month.update.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanGuaranteedFundPerMonth) []string {
			return []string{
				"loan_guaranteed_fund_per_month.delete",
				fmt.Sprintf("loan_guaranteed_fund_per_month.delete.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) LoanGuaranteedFundPerMonthCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*LoanGuaranteedFundPerMonth, error) {
	return m.LoanGuaranteedFundPerMonthManager.Find(context, &LoanGuaranteedFundPerMonth{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
