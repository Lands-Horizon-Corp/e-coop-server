package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	AccountCategory struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account_category" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account_category" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255)" json:"name"`
		Description string `gorm:"type:text" json:"description"`
	}

	AccountCategoryResponse struct {
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
		Name           string                `json:"name"`
		Description    string                `json:"description"`
	}

	AccountCategoryRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description,omitempty"`
	}
)

func (m *Core) accountCategory() {
	m.Migration = append(m.Migration, &AccountCategory{})
	m.AccountCategoryManager = *registry.NewRegistry(registry.RegistryParams[
		AccountCategory, AccountCategoryResponse, AccountCategoryRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *AccountCategory) *AccountCategoryResponse {
			if data == nil {
				return nil
			}
			return &AccountCategoryResponse{
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
				Name:           data.Name,
				Description:    data.Description,
			}
		},
		Created: func(data *AccountCategory) registry.Topics {
			return []string{
				"account_category.create",
				fmt.Sprintf("account_category.create.%s", data.ID),
				fmt.Sprintf("account_category.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_category.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *AccountCategory) registry.Topics {
			return []string{
				"account_category.update",
				fmt.Sprintf("account_category.update.%s", data.ID),
				fmt.Sprintf("account_category.update.branch.%s", data.BranchID),
				fmt.Sprintf("account_category.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *AccountCategory) registry.Topics {
			return []string{
				"account_category.delete",
				fmt.Sprintf("account_category.delete.%s", data.ID),
				fmt.Sprintf("account_category.delete.branch.%s", data.BranchID),
				fmt.Sprintf("account_category.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) accountCategorySeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	accountCategories := []*AccountCategory{
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Savings Accounts",
			Description:    "Regular savings accounts for members including basic, premium, and specialized savings products.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Time Deposits",
			Description:    "Fixed-term deposit accounts with predetermined interest rates and maturity periods.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Loan Accounts",
			Description:    "Various loan products including personal, business, housing, and emergency loans.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Share Capital",
			Description:    "Member equity accounts representing ownership stake in the cooperative.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Special Purpose Accounts",
			Description:    "Accounts for specific purposes like Christmas savings, education fund, emergency fund.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Cash and Cash Equivalents",
			Description:    "Accounts for managing physical cash, petty cash, and other liquid assets.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Investment Accounts",
			Description:    "Accounts for managing cooperative investments in securities and other financial instruments.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Youth and Student Accounts",
			Description:    "Specialized accounts designed for minors, students, and young members.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Senior Citizen Accounts",
			Description:    "Accounts with special benefits and features for senior citizen members.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Business and Corporate Accounts",
			Description:    "Accounts designed for business members and corporate entities.",
		},
	}

	for _, data := range accountCategories {
		if err := m.AccountCategoryManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed account category %s", data.Name)
		}
	}

	return nil
}

func (m *Core) AccountCategoryCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*AccountCategory, error) {
	return m.AccountCategoryManager.Find(context, &AccountCategory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
