package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) OrganizationDailyUsageOnCreate(data *model.OrganizationDailyUsage) {
	go func() {
		b.broadcast.Dispatch([]string{
			"organization_daily_usage.create",
			fmt.Sprintf("organization_daily_usage.create.%s", data.ID),
			fmt.Sprintf("organization_daily_usage.create.organization.%s", data.OrganizationID),
		}, b.model.OrganizationDailyUsageModel(data))
	}()
}

func (b *Publisher) OrganizationDailyUsageOnUpdate(data *model.OrganizationDailyUsage) {
	go func() {
		b.broadcast.Dispatch([]string{
			"organization_daily_usage.update",
			fmt.Sprintf("organization_daily_usage.update.%s", data.ID),
			fmt.Sprintf("organization_daily_usage.update.organization.%s", data.OrganizationID),
		}, b.model.OrganizationDailyUsageModel(data))
	}()

}

func (b *Publisher) OrganizationDailyUsageOnDelete(data *model.OrganizationDailyUsage) {
	go func() {
		b.broadcast.Dispatch([]string{
			"organization_daily_usage.delete",
			fmt.Sprintf("organization_daily_usage.delete.%s", data.ID),
			fmt.Sprintf("organization_daily_usage.delete.organization.%s", data.OrganizationID),
		}, b.model.OrganizationDailyUsageModel(data))
	}()
}
