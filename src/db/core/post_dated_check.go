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

func PostDatedCheckManager(service *horizon.HorizonService) *registry.Registry[
	types.PostDatedCheck, types.PostDatedCheckResponse, types.PostDatedCheckRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.PostDatedCheck, types.PostDatedCheckResponse, types.PostDatedCheckRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "MemberProfile", "Bank",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.PostDatedCheck) *types.PostDatedCheckResponse {
			if data == nil {
				return nil
			}
			return &types.PostDatedCheckResponse{
				ID:                  data.ID,
				CreatedAt:           data.CreatedAt.Format(time.RFC3339),
				CreatedByID:         data.CreatedByID,
				CreatedBy:           UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:           data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:         data.UpdatedByID,
				UpdatedBy:           UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:      data.OrganizationID,
				Organization:        OrganizationManager(service).ToModel(data.Organization),
				BranchID:            data.BranchID,
				Branch:              BranchManager(service).ToModel(data.Branch),
				MemberProfileID:     data.MemberProfileID,
				MemberProfile:       MemberProfileManager(service).ToModel(data.MemberProfile),
				FullName:            data.FullName,
				PassbookNumber:      data.PassbookNumber,
				CheckNumber:         data.CheckNumber,
				CheckDate:           data.CheckDate.Format(time.RFC3339),
				ClearDays:           data.ClearDays,
				DateCleared:         data.DateCleared.Format(time.RFC3339),
				BankID:              data.BankID,
				Bank:                BankManager(service).ToModel(data.Bank),
				Amount:              data.Amount,
				ReferenceNumber:     data.ReferenceNumber,
				OfficialReceiptDate: data.OfficialReceiptDate.Format(time.RFC3339),
				CollateralUserID:    data.CollateralUserID,
				Description:         data.Description,
			}
		},

		Created: func(data *types.PostDatedCheck) registry.Topics {
			return []string{
				"post_dated_check.create",
				fmt.Sprintf("post_dated_check.create.%s", data.ID),
				fmt.Sprintf("post_dated_check.create.branch.%s", data.BranchID),
				fmt.Sprintf("post_dated_check.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.PostDatedCheck) registry.Topics {
			return []string{
				"post_dated_check.update",
				fmt.Sprintf("post_dated_check.update.%s", data.ID),
				fmt.Sprintf("post_dated_check.update.branch.%s", data.BranchID),
				fmt.Sprintf("post_dated_check.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.PostDatedCheck) registry.Topics {
			return []string{
				"post_dated_check.delete",
				fmt.Sprintf("post_dated_check.delete.%s", data.ID),
				fmt.Sprintf("post_dated_check.delete.branch.%s", data.BranchID),
				fmt.Sprintf("post_dated_check.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func PostDatedCheckCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID,
	branchID uuid.UUID) ([]*types.PostDatedCheck, error) {
	return PostDatedCheckManager(service).Find(context, &types.PostDatedCheck{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
