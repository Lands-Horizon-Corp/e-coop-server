package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

func GeneralAccountingLedgerTagManager(service *horizon.HorizonService) *registry.Registry[
	types.GeneralAccountingLedgerTag, types.GeneralAccountingLedgerTagResponse, types.GeneralAccountingLedgerTagRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.GeneralAccountingLedgerTag, types.GeneralAccountingLedgerTagResponse, types.GeneralAccountingLedgerTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "GeneralLedger",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.GeneralAccountingLedgerTag) *types.GeneralAccountingLedgerTagResponse {
			if data == nil {
				return nil
			}
			return &types.GeneralAccountingLedgerTagResponse{
				ID:              data.ID,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:     data.CreatedByID,
				CreatedBy:       UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:     data.UpdatedByID,
				UpdatedBy:       UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:  data.OrganizationID,
				Organization:    OrganizationManager(service).ToModel(data.Organization),
				BranchID:        data.BranchID,
				Branch:          BranchManager(service).ToModel(data.Branch),
				GeneralLedgerID: data.GeneralLedgerID,
				GeneralLedger:   GeneralLedgerManager(service).ToModel(data.GeneralLedger),
				Name:            data.Name,
				Description:     data.Description,
				Category:        data.Category,
				Color:           data.Color,
				Icon:            data.Icon,
			}
		},
		Created: func(data *types.GeneralAccountingLedgerTag) registry.Topics {
			return []string{
				"general_ledger_tag.create",
				fmt.Sprintf("general_ledger_tag.create.%s", data.ID),
				fmt.Sprintf("general_ledger_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.GeneralAccountingLedgerTag) registry.Topics {
			return []string{
				"general_ledger_tag.update",
				fmt.Sprintf("general_ledger_tag.update.%s", data.ID),
				fmt.Sprintf("general_ledger_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.GeneralAccountingLedgerTag) registry.Topics {
			return []string{
				"general_ledger_tag.delete",
				fmt.Sprintf("general_ledger_tag.delete.%s", data.ID),
				fmt.Sprintf("general_ledger_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GeneralAccountingLedgerTagCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.GeneralAccountingLedgerTag, error) {
	return GeneralAccountingLedgerTagManager(service).Find(context, &types.GeneralAccountingLedgerTag{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
