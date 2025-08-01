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
	FinesMaturity struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_fines_maturity"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_fines_maturity"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID *uuid.UUID `gorm:"type:uuid"`
		Account   *Account   `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"account,omitempty"`

		From int     `gorm:"not null;default:0"`
		To   int     `gorm:"not null;default:0"`
		Rate float64 `gorm:"type:decimal;not null;default:0"`
	}

	FinesMaturityResponse struct {
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

	FinesMaturityRequest struct {
		AccountID *uuid.UUID `json:"account_id,omitempty"`
		From      int        `json:"from" validate:"required"`
		To        int        `json:"to" validate:"required"`
		Rate      float64    `json:"rate" validate:"required"`
	}
)

func (m *Model) FinesMaturity() {
	m.Migration = append(m.Migration, &FinesMaturity{})
	m.FinesMaturityManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		FinesMaturity, FinesMaturityResponse, FinesMaturityRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization", "Account",
		},
		Service: m.provider.Service,
		Resource: func(data *FinesMaturity) *FinesMaturityResponse {
			if data == nil {
				return nil
			}
			return &FinesMaturityResponse{
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

		Created: func(data *FinesMaturity) []string {
			return []string{
				"fines_maturity.create",
				fmt.Sprintf("fines_maturity.create.%s", data.ID),
				fmt.Sprintf("fines_maturity.create.branch.%s", data.BranchID),
				fmt.Sprintf("fines_maturity.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *FinesMaturity) []string {
			return []string{
				"fines_maturity.update",
				fmt.Sprintf("fines_maturity.update.%s", data.ID),
				fmt.Sprintf("fines_maturity.update.branch.%s", data.BranchID),
				fmt.Sprintf("fines_maturity.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *FinesMaturity) []string {
			return []string{
				"fines_maturity.delete",
				fmt.Sprintf("fines_maturity.delete.%s", data.ID),
				fmt.Sprintf("fines_maturity.delete.branch.%s", data.BranchID),
				fmt.Sprintf("fines_maturity.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) FinesMaturityCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*FinesMaturity, error) {
	return m.FinesMaturityManager.Find(context, &FinesMaturity{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
