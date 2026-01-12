package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	Collateral struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_collateral"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_collateral"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Icon        string `gorm:"type:varchar(255)"` // from React Icons, e.g. "FaCar", "MdHome"
		Name        string `gorm:"type:varchar(255);not null"`
		Description string `gorm:"type:text"`
	}

	CollateralResponse struct {
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
		Icon           string                `json:"icon"`
		Name           string                `json:"name"`
		Description    string                `json:"description"`
	}

	CollateralRequest struct {
		Icon        string `json:"icon,omitempty"`
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description,omitempty"`
	}
)

func CollateralManager(service *horizon.HorizonService) *registry.Registry[Collateral, CollateralResponse, CollateralRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		Collateral, CollateralResponse, CollateralRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *Collateral) *CollateralResponse {
			if data == nil {
				return nil
			}
			return &CollateralResponse{
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
		Created: func(data *Collateral) registry.Topics {
			return []string{
				"collateral.create",
				fmt.Sprintf("collateral.create.%s", data.ID),
				fmt.Sprintf("collateral.create.branch.%s", data.BranchID),
				fmt.Sprintf("collateral.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Collateral) registry.Topics {
			return []string{
				"collateral.update",
				fmt.Sprintf("collateral.update.%s", data.ID),
				fmt.Sprintf("collateral.update.branch.%s", data.BranchID),
				fmt.Sprintf("collateral.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Collateral) registry.Topics {
			return []string{
				"collateral.delete",
				fmt.Sprintf("collateral.delete.%s", data.ID),
				fmt.Sprintf("collateral.delete.branch.%s", data.BranchID),
				fmt.Sprintf("collateral.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func collateralSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	collaterals := []*Collateral{
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

func CollateralCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*Collateral, error) {
	return CollateralManager(service).Find(context, &Collateral{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
