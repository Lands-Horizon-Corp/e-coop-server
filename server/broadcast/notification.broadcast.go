package broadcast

import (
	"fmt"

	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
)

type NotificationBroadcast struct {
	broadcast  *horizon.HorizonBroadcast
	collection *collection.NotificationCollection
}

func NewNotificationBroadcast(
	broadcast *horizon.HorizonBroadcast,
	collection *collection.NotificationCollection,
) (*NotificationBroadcast, error) {
	return &NotificationBroadcast{
		broadcast:  broadcast,
		collection: collection,
	}, nil
}

func (b *NotificationBroadcast) OnCreate(data *collection.Notification) {
	go func() {
		b.broadcast.Dispatch([]string{
			"notification.create",
			fmt.Sprintf("footstep.create.user.%s", data.UserID),
			fmt.Sprintf("notification.create.%s", data.ID),
		}, b.collection.ToModel(data))
	}()
}

func (b *NotificationBroadcast) OnUpdate(data *collection.Notification) {
	go func() {
		b.broadcast.Dispatch([]string{
			"notification.update",
			fmt.Sprintf("footstep.update.user.%s", data.UserID),
			fmt.Sprintf("notification.update.%s", data.ID),
		}, b.collection.ToModel(data))
	}()

}

func (b *NotificationBroadcast) OnDelete(data *collection.Notification) {
	go func() {
		b.broadcast.Dispatch([]string{
			"notification.delete",
			fmt.Sprintf("footstep.delete.user.%s", data.UserID),
			fmt.Sprintf("notification.delete.%s", data.ID),
		}, b.collection.ToModel(data))
	}()
}
