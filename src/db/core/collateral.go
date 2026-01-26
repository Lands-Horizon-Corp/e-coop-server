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

func CollateralManager(service *horizon.HorizonService) *registry.Registry[types.Collateral, types.CollateralResponse, types.CollateralRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.Collateral, types.CollateralResponse, types.CollateralRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Collateral) *types.CollateralResponse {
			if data == nil {
				return nil
			}
			return &types.CollateralResponse{
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
				Icon:           data.Icon,
				Name:           data.Name,
				Description:    data.Description,
			}
		},
		Created: func(data *types.Collateral) registry.Topics {
			return []string{
				"collateral.create",
				fmt.Sprintf("collateral.create.%s", data.ID),
				fmt.Sprintf("collateral.create.branch.%s", data.BranchID),
				fmt.Sprintf("collateral.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.Collateral) registry.Topics {
			return []string{
				"collateral.update",
				fmt.Sprintf("collateral.update.%s", data.ID),
				fmt.Sprintf("collateral.update.branch.%s", data.BranchID),
				fmt.Sprintf("collateral.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.Collateral) registry.Topics {
			return []string{
				"collateral.delete",
				fmt.Sprintf("collateral.delete.%s", data.ID),
				fmt.Sprintf("collateral.delete.branch.%s", data.BranchID),
				fmt.Sprintf("collateral.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func collateralSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB, userID uuid.UUID,
	organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	collaterals := []*types.Collateral{
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Vehicle/Motor Vehicle",
			Description:    "Cars, motorcycles, trucks, and other motor vehicles that can be used as collateral for loans.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Real Estate Property",
			Description:    "Houses, lots, buildings, and other real estate properties including residential and commercial properties.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Jewelry & Precious Items",
			Description:    "Gold, silver, diamonds, and other precious jewelry items that hold significant value.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Electronics & Gadgets",
			Description:    "Laptops, smartphones, tablets, cameras, and other electronic devices of significant value.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Machinery & Equipment",
			Description:    "Industrial machinery, construction equipment, and other business equipment used as collateral.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Savings Account/Time Deposit",
			Description:    "Bank savings accounts, time deposits, and other financial instruments as collateral.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Securities & Bonds",
			Description:    "Government bonds, corporate securities, stocks, and other investment instruments.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Agricultural Assets",
			Description:    "Farm equipment, livestock, agricultural land, and other farming-related assets.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Furniture & Appliances",
			Description:    "Valuable furniture sets, home appliances, and other household items of significant worth.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Co-maker/Guarantor",
			Description:    "Personal guarantees from co-makers or guarantors who pledge to pay the loan if the borrower defaults.",
		},
	}
	for _, data := range collaterals {
		if err := CollateralManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed collateral %s", data.Name)
		}
	}
	return nil
}

func CollateralCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.Collateral, error) {
	return CollateralManager(service).Find(context, &types.Collateral{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
