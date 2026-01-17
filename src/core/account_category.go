package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func AccountCategoryManager(service *horizon.HorizonService) *registry.Registry[types.AccountCategory, types.AccountCategoryResponse, types.AccountCategoryRequest] {
	return registry.GetRegistry(
		registry.RegistryParams[types.AccountCategory, types.AccountCategoryResponse, types.AccountCategoryRequest]{
			Preloads: []string{"CreatedBy", "UpdatedBy"},
			Database: service.Database.Client(),
			Dispatch: func(topics registry.Topics, payload any) error {
				return service.Broker.Dispatch(topics, payload)
			},
			Resource: func(data *types.AccountCategory) *types.AccountCategoryResponse {
				if data == nil {
					return nil
				}
				return &types.AccountCategoryResponse{
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
			Created: func(data *types.AccountCategory) registry.Topics {
				return []string{
					"account_category.create",
					fmt.Sprintf("account_category.create.%s", data.ID),
					fmt.Sprintf("account_category.create.branch.%s", data.BranchID),
					fmt.Sprintf("account_category.create.organization.%s", data.OrganizationID),
				}
			},
			Updated: func(data *types.AccountCategory) registry.Topics {
				return []string{
					"account_category.update",
					fmt.Sprintf("account_category.update.%s", data.ID),
					fmt.Sprintf("account_category.update.branch.%s", data.BranchID),
					fmt.Sprintf("account_category.update.organization.%s", data.OrganizationID),
				}
			},
			Deleted: func(data *types.AccountCategory) registry.Topics {
				return []string{
					"account_category.delete",
					fmt.Sprintf("account_category.delete.%s", data.ID),
					fmt.Sprintf("account_category.delete.branch.%s", data.BranchID),
					fmt.Sprintf("account_category.delete.organization.%s", data.OrganizationID),
				}
			},
		})
}

func accountCategorySeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	accountCategories := []*types.AccountCategory{
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
		if err := AccountCategoryManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed account category %s", data.Name)
		}
	}

	return nil
}

func AccountCategoryCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.AccountCategory, error) {
	return AccountCategoryManager(service).Find(context, &types.AccountCategory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
