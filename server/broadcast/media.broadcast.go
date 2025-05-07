package broadcast

import (
	"fmt"

	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
)

type MediaBroadcast struct {
	broadcast  *horizon.HorizonBroadcast
	collection *collection.MediaCollection
}

func NewMediaBroadcast(
	broadcast *horizon.HorizonBroadcast,
	collection *collection.MediaCollection,
) (*MediaBroadcast, error) {
	return &MediaBroadcast{
		broadcast:  broadcast,
		collection: collection,
	}, nil
}

func (b *MediaBroadcast) OnCreate(data *collection.Media) {
	go func() {
		b.broadcast.Dispatch([]string{
			"media.create",
			fmt.Sprintf("media.create.%s", data.ID),
		}, b.collection.ToModel(data))
	}()
}

func (b *MediaBroadcast) OnUpdate(data *collection.Media) {
	go func() {
		b.broadcast.Dispatch([]string{
			"media.update",
			fmt.Sprintf("media.update.%s", data.ID),
		}, b.collection.ToModel(data))
	}()

}
func (b *MediaBroadcast) OnDelete(data *collection.Media) {
	go func() {
		b.broadcast.Dispatch([]string{
			"media.delete",
			fmt.Sprintf("media.delete.%s", data.ID),
		}, b.collection.ToModel(data))
	}()
}
