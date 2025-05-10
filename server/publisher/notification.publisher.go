package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) NotificationOnCreate(data *model.Notification) {
	go func() {
		b.broadcast.Dispatch([]string{
			"notification.create",
			fmt.Sprintf("footstep.create.user.%s", data.UserID),
			fmt.Sprintf("notification.create.%s", data.ID),
		}, b.model.NotificationModel(data))
	}()
}

func (b *Publisher) NotificationOnUpdate(data *model.Notification) {
	go func() {
		b.broadcast.Dispatch([]string{
			"notification.update",
			fmt.Sprintf("footstep.update.user.%s", data.UserID),
			fmt.Sprintf("notification.update.%s", data.ID),
		}, b.model.NotificationModel(data))
	}()

}

func (b *Publisher) NotificationOnDelete(data *model.Notification) {
	go func() {
		b.broadcast.Dispatch([]string{
			"notification.delete",
			fmt.Sprintf("footstep.delete.user.%s", data.UserID),
			fmt.Sprintf("notification.delete.%s", data.ID),
		}, b.model.NotificationModel(data))
	}()
}
