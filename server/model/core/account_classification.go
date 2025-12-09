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
	// AccountClassification represents the classification of an account in the database
	AccountClassification struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account_classification" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account_classification" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255)" json:"name"`
		Description string `gorm:"type:text" json:"description"`
	}

	// AccountClassificationResponse represents the response structure for account classification data
	AccountClassificationResponse struct {
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

	// AccountClassificationRequest represents the request structure for creating or updating account classification data
	AccountClassificationRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description,omitempty"`
	}
)

func (m *Core) accountClassificationSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()

	classifications := []*AccountClassification{
		// Financial Statement Classifications
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Assets",
			Description:    "Resources owned by the organization that have economic value including cash, receivables, and equipment",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Liabilities",
			Description:    "Obligations owed to external parties including deposits, borrowings, and accounts payable",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Equity",
			Description:    "Members' ownership and capital contributions to the cooperative",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Income",
			Description:    "Revenue and earnings from operations including interest income and service fees",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Expenses",
			Description:    "Costs incurred in operations including salaries, utilities, and administrative expenses",
		},
		// Deposit Account Classifications
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Savings Account",
			Description:    "Personal savings accounts for members with flexible withdrawal terms",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Current Account",
			Description:    "Business or transactional accounts for frequent operations",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Fixed Deposit",
			Description:    "Time deposits with fixed terms and higher interest rates",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Recurring Deposit",
			Description:    "Regular monthly deposits with compound interest benefits",
		},
		// Loan Account Classifications
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Personal Loan",
			Description:    "Unsecured loans for personal use and family needs",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Housing Loan",
			Description:    "Long-term secured loans for home purchase or construction",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Business Loan",
			Description:    "Commercial loans for business operations and expansion",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Agricultural Loan",
			Description:    "Specialized loans for farming and agricultural activities",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Livelihood Loan",
			Description:    "Microfinance loans for income-generating activities",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Emergency Loan",
			Description:    "Quick disbursement loans for urgent financial needs",
		},
		// Account Status Classifications
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Active Account",
			Description:    "Accounts that are regularly operated and maintained",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Dormant Account",
			Description:    "Accounts with no activity for more than 12 months",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Frozen Account",
			Description:    "Accounts with operations stopped due to legal or compliance reasons",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Closed Account",
			Description:    "Accounts formally terminated by the member or cooperative",
		},
		// Ownership Classifications
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Individual Account",
			Description:    "Accounts owned by a single person",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Joint Account",
			Description:    "Accounts shared by two or more individuals",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Corporate Account",
			Description:    "Accounts opened by companies and organizations",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Trust Account",
			Description:    "Fiduciary accounts managed on behalf of someone else",
		},
		// Regulatory Classifications for Asset Quality
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Standard Account",
			Description:    "Loan accounts with payments made on time and no default",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Substandard Account",
			Description:    "Loan accounts with payments overdue for a short period",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Doubtful Account",
			Description:    "Loan accounts with long overdue payments and high risk of non-repayment",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Loss Account",
			Description:    "Loan accounts considered uncollectible and written off",
		},
		// Cooperative-Specific Classifications
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Member Share Capital",
			Description:    "Equity contributions by cooperative members",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Statutory Reserve",
			Description:    "Mandatory reserves required by cooperative regulations",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Surplus Account",
			Description:    "Retained earnings from cooperative operations",
		},
		// Investment and Special Accounts
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Investment Account",
			Description:    "Accounts for mutual funds, bonds, and other investments",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Suspense Account",
			Description:    "Temporary accounts for pending transactions and adjustments",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Clearing Account",
			Description:    "Accounts for clearing and settlement of inter-branch transactions",
		},
	}

	for _, classification := range classifications {
		if err := m.AccountClassificationManager.CreateWithTx(context, tx, classification); err != nil {
			return fmt.Errorf("failed to seed account classification %s: %w", classification.Name, err)
		}
	}
	return nil
}
func (m *Core) accountClassification() {
	m.Migration = append(m.Migration, &AccountClassification{})
	m.AccountClassificationManager = *registry.NewRegistry(registry.RegistryParams[
		AccountClassification, AccountClassificationResponse, AccountClassificationRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Database: m.provider.Service.Database.Client(),
Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		}
		Resource: func(data *AccountClassification) *AccountClassificationResponse {
			if data == nil {
				return nil
			}
			return &AccountClassificationResponse{
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
		Created: func(data *AccountClassification) []string {
			return []string{
				"account_classification.create",
				fmt.Sprintf("account_classification.create.%s", data.ID),
				fmt.Sprintf("account_classification.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_classification.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *AccountClassification) []string {
			return []string{
				"account_classification.update",
				fmt.Sprintf("account_classification.update.%s", data.ID),
				fmt.Sprintf("account_classification.update.branch.%s", data.BranchID),
				fmt.Sprintf("account_classification.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *AccountClassification) []string {
			return []string{
				"account_classification.delete",
				fmt.Sprintf("account_classification.delete.%s", data.ID),
				fmt.Sprintf("account_classification.delete.branch.%s", data.BranchID),
				fmt.Sprintf("account_classification.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// AccountClassificationCurrentBranch retrieves all account classifications for the specified organization and branch
func (m *Core) AccountClassificationCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*AccountClassification, error) {
	return m.AccountClassificationManager.Find(context, &AccountClassification{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
