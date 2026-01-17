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

func OrganizationDailyUsageManager(service *horizon.HorizonService) *registry.Registry[
	types.OrganizationDailyUsage, types.OrganizationDailyUsageResponse, types.OrganizationDailyUsageRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.OrganizationDailyUsage, types.OrganizationDailyUsageResponse, types.OrganizationDailyUsageRequest]{
		Preloads: []string{"Organization"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.OrganizationDailyUsage) *types.OrganizationDailyUsageResponse {
			if data == nil {
				return nil
			}
			return &types.OrganizationDailyUsageResponse{
				ID:             data.ID,
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				TotalMembers:   data.TotalMembers,
				TotalBranches:  data.TotalBranches,
				TotalEmployees: data.TotalEmployees,

				CashTransactionCount:   data.CashTransactionCount,
				CheckTransactionCount:  data.CheckTransactionCount,
				OnlineTransactionCount: data.OnlineTransactionCount,

				CashTransactionAmount:   data.CashTransactionAmount,
				CheckTransactionAmount:  data.CheckTransactionAmount,
				OnlineTransactionAmount: data.OnlineTransactionAmount,

				TotalEmailSend:        data.TotalEmailSend,
				TotalMessageSend:      data.TotalMessageSend,
				TotalUploadSize:       data.TotalUploadSize,
				TotalReportRenderTime: data.TotalReportRenderTime,

				CreatedAt: data.CreatedAt.Format(time.RFC3339),
				UpdatedAt: data.UpdatedAt.Format(time.RFC3339),
			}
		},

		Created: func(data *types.OrganizationDailyUsage) registry.Topics {
			return []string{
				"organization_daily_usage.create",
				fmt.Sprintf("organization_daily_usage.create.%s", data.ID),
				fmt.Sprintf("organization_daily_usage.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.OrganizationDailyUsage) registry.Topics {
			return []string{
				"organization_daily_usage.update",
				fmt.Sprintf("organization_daily_usage.update.%s", data.ID),
				fmt.Sprintf("organization_daily_usage.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.OrganizationDailyUsage) registry.Topics {
			return []string{
				"organization_daily_usage.delete",
				fmt.Sprintf("organization_daily_usage.delete.%s", data.ID),
				fmt.Sprintf("organization_daily_usage.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GetOrganizationDailyUsageByOrganization(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID) ([]*types.OrganizationDailyUsage, error) {
	return OrganizationDailyUsageManager(service).Find(context, &types.OrganizationDailyUsage{
		OrganizationID: organizationID,
	})
}
