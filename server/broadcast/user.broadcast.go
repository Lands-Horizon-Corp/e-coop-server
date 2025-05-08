package broadcast

import (
	"fmt"

	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
)

type UserBroadcast struct {
	broadcast  *horizon.HorizonBroadcast
	collection *collection.UserCollection
}

func NewUserBroadcast(
	broadcast *horizon.HorizonBroadcast,
	collection *collection.UserCollection,
) (*UserBroadcast, error) {
	return &UserBroadcast{
		broadcast:  broadcast,
		collection: collection,
	}, nil
}

func (b *UserBroadcast) OnCreate(data *collection.User) {
	go func() {
		b.broadcast.Dispatch([]string{
			"user.create",
			fmt.Sprintf("user.create.%s", data.ID),
		}, b.collection.ToModel(data))
	}()
}

func (b *UserBroadcast) OnUpdate(data *collection.User) {
	go func() {
		b.broadcast.Dispatch([]string{
			"user.update",
			fmt.Sprintf("user.update.%s", data.ID),
		}, b.collection.ToModel(data))
	}()

}

func (b *UserBroadcast) OnDelete(data *collection.User) {
	go func() {
		b.broadcast.Dispatch([]string{
			"user.delete",
			fmt.Sprintf("user.delete.%s", data.ID),
		}, b.collection.ToModel(data))
	}()
}
