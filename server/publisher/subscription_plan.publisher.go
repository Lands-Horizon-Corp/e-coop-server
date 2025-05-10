package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) SubscriptionPlanOnCreate(data *model.SubscriptionPlan) {
	go func() {
		b.broadcast.Dispatch([]string{
			"subscription_plan.create",
			fmt.Sprintf("subscription_plan.create.%s", data.ID),
		}, b.model.SubscriptionPlanModel(data))
	}()
}

func (b *Publisher) SubscriptionPlanOnUpdate(data *model.SubscriptionPlan) {
	go func() {
		b.broadcast.Dispatch([]string{
			"subscription_plan.update",
			fmt.Sprintf("subscription_plan.update.%s", data.ID),
		}, b.model.SubscriptionPlanModel(data))
	}()

}

func (b *Publisher) SubscriptionPlanOnDelete(data *model.SubscriptionPlan) {
	go func() {
		b.broadcast.Dispatch([]string{
			"subscription_plan.delete",
			fmt.Sprintf("subscription_plan.delete.%s", data.ID),
		}, b.model.SubscriptionPlanModel(data))
	}()
}
