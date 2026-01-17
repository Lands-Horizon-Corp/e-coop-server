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

func DisbursementManager(service *horizon.HorizonService) *registry.Registry[types.Disbursement, types.DisbursementResponse, types.DisbursementRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.Disbursement, types.DisbursementResponse, types.DisbursementRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Currency",
			"Organization.Media", "Branch.Media",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Disbursement) *types.DisbursementResponse {
			if data == nil {
				return nil
			}
			return &types.DisbursementResponse{
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
				CurrencyID:     data.CurrencyID,
				Currency:       CurrencyManager(service).ToModel(data.Currency),
				Name:           data.Name,
				Icon:           data.Icon,
				Description:    data.Description,
			}
		},
		Created: func(data *types.Disbursement) registry.Topics {
			return []string{
				"disbursement.create",
				fmt.Sprintf("disbursement.create.%s", data.ID),
				fmt.Sprintf("disbursement.create.branch.%s", data.BranchID),
				fmt.Sprintf("disbursement.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.Disbursement) registry.Topics {
			return []string{
				"disbursement.update",
				fmt.Sprintf("disbursement.update.%s", data.ID),
				fmt.Sprintf("disbursement.update.branch.%s", data.BranchID),
				fmt.Sprintf("disbursement.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.Disbursement) registry.Topics {
			return []string{
				"disbursement.delete",
				fmt.Sprintf("disbursement.delete.%s", data.ID),
				fmt.Sprintf("disbursement.delete.branch.%s", data.BranchID),
				fmt.Sprintf("disbursement.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func disbursementSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB,
	userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()

	branch, err := BranchManager(service).GetByID(context, branchID)
	if err != nil {
		return eris.Wrap(err, "failed to find branch for account seeding")
	}

	disbursements := []*types.Disbursement{
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     *branch.CurrencyID,
			Name:           "Petty Cash",

			Description: "Small cash disbursements for minor expenses and miscellaneous operational costs.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     *branch.CurrencyID,
			Name:           "Office Supplies",

			Description: "Purchase of office materials, stationery, and administrative supplies.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     *branch.CurrencyID,
			Name:           "Utilities",
			Description:    "Payment for electricity, water, internet, and other utility services.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     *branch.CurrencyID,
			Name:           "Travel Expenses",
			Description:    "Transportation costs, accommodation, and meal allowances for official travels.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     *branch.CurrencyID,
			Name:           "Meeting Expenses",
			Description:    "Costs associated with meetings, seminars, and training sessions.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     *branch.CurrencyID,
			Name:           "Equipment Purchase",
			Description:    "Acquisition of office equipment, furniture, and technology devices.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     *branch.CurrencyID,
			Name:           "Maintenance & Repairs",
			Description:    "Building maintenance, equipment repairs, and facility improvements.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     *branch.CurrencyID,
			Name:           "Insurance Premium",
			Description:    "Payment for insurance coverage including property, liability, and employee insurance.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     *branch.CurrencyID,
			Name:           "Professional Services",
			Description:    "Fees for legal, accounting, consulting, and other professional services.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     *branch.CurrencyID,
			Name:           "Member Benefits",
			Description:    "Disbursements for member welfare, dividends, and cooperative benefits distribution.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     *branch.CurrencyID,
			Name:           "Loan Disbursement",
			Description:    "Release of approved loans to cooperative members.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     *branch.CurrencyID,
			Name:           "Emergency Fund",
			Description:    "Disbursements from emergency reserves for urgent organizational needs.",
		},
	}

	for _, data := range disbursements {
		if err := DisbursementManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed disbursement %s", data.Name)
		}
	}

	return nil
}

func DisbursementCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID,
	branchID uuid.UUID) ([]*types.Disbursement, error) {
	return DisbursementManager(service).Find(context, &types.Disbursement{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
