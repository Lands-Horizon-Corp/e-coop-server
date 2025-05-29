package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
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

func (m *Model) LoanGuaranteedFund() {
	m.Migration = append(m.Migration, &LoanGuaranteedFund{})
	m.LoanGuaranteedFundManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		LoanGuaranteedFund, LoanGuaranteedFundResponse, LoanGuaranteedFundRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanGuaranteedFund) *LoanGuaranteedFundResponse {
			if data == nil {
				return nil
			}
			return &LoanGuaranteedFundResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),
				SchemeNumber:   data.SchemeNumber,
				IncreasingRate: data.IncreasingRate,
			}
		},
		Created: func(data *LoanGuaranteedFund) []string {
			return []string{
				"loan_guaranteed_fund.create",
				fmt.Sprintf("loan_guaranteed_fund.create.%s", data.ID),
			}
		},
		Updated: func(data *LoanGuaranteedFund) []string {
			return []string{
				"loan_guaranteed_fund.update",
				fmt.Sprintf("loan_guaranteed_fund.update.%s", data.ID),
			}
		},
		Deleted: func(data *LoanGuaranteedFund) []string {
			return []string{
				"loan_guaranteed_fund.delete",
				fmt.Sprintf("loan_guaranteed_fund.delete.%s", data.ID),
			}
		},
	})
}
