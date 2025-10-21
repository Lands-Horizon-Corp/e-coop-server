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
	InterestMaturity struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_maturity"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_maturity"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID *uuid.UUID `gorm:"type:uuid"`
		Account   *Account   `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"account,omitempty"`

		From int     `gorm:"not null;default:0"`
		To   int     `gorm:"not null;default:0"`
		Rate float64 `gorm:"type:decimal;not null;default:0"`
	}

	InterestMaturityResponse struct {
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
		AccountID      *uuid.UUID            `json:"account_id,omitempty"`
		Account        *AccountResponse      `json:"account,omitempty"`
		From           int                   `json:"from"`
		To             int                   `json:"to"`
		Rate           float64               `json:"rate"`
	}

	InterestMaturityRequest struct {
		AccountID *uuid.UUID `json:"account_id,omitempty"`
		From      int        `json:"from" validate:"required"`
		To        int        `json:"to" validate:"required"`
		Rate      float64    `json:"rate" validate:"required"`
	}
)

func (m *ModelCore) InterestMaturity() {
	m.Migration = append(m.Migration, &InterestMaturity{})
	m.InterestMaturityManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		InterestMaturity, InterestMaturityResponse, InterestMaturityRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Account",
		},
		Service: m.provider.Service,
		Resource: func(data *InterestMaturity) *InterestMaturityResponse {
			if data == nil {
				return nil
			}
			return &InterestMaturityResponse{
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
				AccountID:      data.AccountID,
				Account:        m.AccountManager.ToModel(data.Account),
				From:           data.From,
				To:             data.To,
				Rate:           data.Rate,
			}
		},
		Created: func(data *InterestMaturity) []string {
			return []string{
				"interest_maturity.create",
				fmt.Sprintf("interest_maturity.create.%s", data.ID),
				fmt.Sprintf("interest_maturity.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_maturity.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *InterestMaturity) []string {
			return []string{
				"interest_maturity.update",
				fmt.Sprintf("interest_maturity.update.%s", data.ID),
				fmt.Sprintf("interest_maturity.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_maturity.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *InterestMaturity) []string {
			return []string{
				"interest_maturity.delete",
				fmt.Sprintf("interest_maturity.delete.%s", data.ID),
				fmt.Sprintf("interest_maturity.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_maturity.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) InterestMaturityCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*InterestMaturity, error) {
	return m.InterestMaturityManager.Find(context, &InterestMaturity{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
