package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func accountClassificationSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()

	classifications := []*types.AccountClassification{
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
		if err := AccountClassificationManager(service).CreateWithTx(context, tx, classification); err != nil {
			return fmt.Errorf("failed to seed account classification %s: %w", classification.Name, err)
		}
	}
	return nil
}
func AccountClassificationManager(service *horizon.HorizonService) *registry.Registry[types.AccountClassification, types.AccountClassificationResponse, types.AccountClassificationRequest] {
	return registry.GetRegistry(registry.RegistryParams[
		types.AccountClassification, types.AccountClassificationResponse, types.AccountClassificationRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.AccountClassification) *types.AccountClassificationResponse {
			if data == nil {
				return nil
			}
			return &types.AccountClassificationResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),
				Name:           data.Name,
				Description:    data.Description,
			}
		},
		Created: func(data *types.AccountClassification) registry.Topics {
			return []string{
				"account_classification.create",
				fmt.Sprintf("account_classification.create.%s", data.ID),
				fmt.Sprintf("account_classification.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_classification.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.AccountClassification) registry.Topics {
			return []string{
				"account_classification.update",
				fmt.Sprintf("account_classification.update.%s", data.ID),
				fmt.Sprintf("account_classification.update.branch.%s", data.BranchID),
				fmt.Sprintf("account_classification.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.AccountClassification) registry.Topics {
			return []string{
				"account_classification.delete",
				fmt.Sprintf("account_classification.delete.%s", data.ID),
				fmt.Sprintf("account_classification.delete.branch.%s", data.BranchID),
				fmt.Sprintf("account_classification.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func AccountClassificationCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.AccountClassification, error) {
	return AccountClassificationManager(service).Find(context, &types.AccountClassification{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
