package model

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Funds struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_funds" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_funds" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID *uuid.UUID `gorm:"type:uuid" json:"account_id"`
		Account   *Account   `gorm:"foreignKey:AccountID;constraint:OnDelete:SET NULL;" json:"account,omitempty"`

		Type        string  `gorm:"type:varchar(255)" json:"type"`
		Description string  `gorm:"type:text" json:"description"`
		Icon        *string `gorm:"type:varchar(255)" json:"icon,omitempty"`
		GLBooks     string  `gorm:"type:varchar(255)" json:"gl_books"`
	}

	FundsResponse struct {
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
		Type           string                `json:"type"`
		Description    string                `json:"description"`
		Icon           *string               `json:"icon,omitempty"`
		GLBooks        string                `json:"gl_books"`
	}

	FundsRequest struct {
		AccountID   *uuid.UUID `json:"account_id,omitempty"`
		Type        string     `json:"type" validate:"required,min=1,max=255"`
		Description string     `json:"description,omitempty"`
		Icon        *string    `json:"icon,omitempty"`
		GLBooks     string     `json:"gl_books,omitempty"`
	}
)

func (m *Model) Funds() {
	m.Migration = append(m.Migration, &Funds{})
	m.FundsManager = horizon_services.NewRepository(horizon_services.RepositoryParams[Funds, FundsResponse, FundsRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization", "Account"},
		Service:  m.provider.Service,
		Resource: func(data *Funds) *FundsResponse {
			if data == nil {
				return nil
			}
			return &FundsResponse{
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
				Type:           data.Type,
				Description:    data.Description,
				Icon:           data.Icon,
				GLBooks:        data.GLBooks,
			}
		},
		Created: func(data *Funds) []string {
			return []string{
				"funds.create",
				fmt.Sprintf("funds.create.%s", data.ID),
				fmt.Sprintf("funds.create.branch.%s", data.BranchID),
				fmt.Sprintf("funds.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Funds) []string {
			return []string{
				"funds.update",
				fmt.Sprintf("funds.update.%s", data.ID),
				fmt.Sprintf("funds.update.branch.%s", data.BranchID),
				fmt.Sprintf("funds.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Funds) []string {
			return []string{
				"funds.delete",
				fmt.Sprintf("funds.delete.%s", data.ID),
				fmt.Sprintf("funds.delete.branch.%s", data.BranchID),
				fmt.Sprintf("funds.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) FundsCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*Funds, error) {
	return m.FundsManager.Find(context, &Funds{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
