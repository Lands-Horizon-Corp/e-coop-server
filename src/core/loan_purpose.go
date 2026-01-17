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

func LoanPurposeManager(service *horizon.HorizonService) *registry.Registry[
	types.LoanPurpose, types.LoanPurposeResponse, types.LoanPurposeRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.LoanPurpose, types.LoanPurposeResponse, types.LoanPurposeRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.LoanPurpose) *types.LoanPurposeResponse {
			if data == nil {
				return nil
			}
			return &types.LoanPurposeResponse{
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
				Description:    data.Description,
				Icon:           data.Icon,
			}
		},

		Created: func(data *types.LoanPurpose) registry.Topics {
			return []string{
				"loan_purpose.create",
				fmt.Sprintf("loan_purpose.create.%s", data.ID),
				fmt.Sprintf("loan_purpose.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_purpose.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.LoanPurpose) registry.Topics {
			return []string{
				"loan_purpose.update",
				fmt.Sprintf("loan_purpose.update.%s", data.ID),
				fmt.Sprintf("loan_purpose.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_purpose.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.LoanPurpose) registry.Topics {
			return []string{
				"loan_purpose.delete",
				fmt.Sprintf("loan_purpose.delete.%s", data.ID),
				fmt.Sprintf("loan_purpose.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_purpose.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func loanPurposeSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB,
	userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	loanPurposes := []*types.LoanPurpose{
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Home Purchase/Construction",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Vehicle Purchase",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Business Capital",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Education/Tuition Fee",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Medical/Healthcare",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Emergency/Personal",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Agricultural/Farming",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Debt Consolidation",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Home Improvement/Renovation",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Appliance/Electronics",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Wedding/Special Occasion",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Equipment/Machinery",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Investment/Securities",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Other/Miscellaneous",
		},
	}

	for _, data := range loanPurposes {
		if err := LoanPurposeManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed loan purpose %s", data.Description)
		}
	}

	return nil
}

func LoanPurposeCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.LoanPurpose, error) {
	return LoanPurposeManager(service).Find(context, &types.LoanPurpose{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
