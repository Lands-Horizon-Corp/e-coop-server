package broadcast

import (
	"fmt"

	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
)

type FootstepBroadcast struct {
	broadcast  *horizon.HorizonBroadcast
	collection *collection.FootstepCollection
}

func NewFootstepBroadcast(
	broadcast *horizon.HorizonBroadcast,
	collection *collection.FootstepCollection,
) (*FootstepBroadcast, error) {
	return &FootstepBroadcast{
		broadcast:  broadcast,
		collection: collection,
	}, nil
}
func (b *FootstepBroadcast) OnCreate(data *collection.Footstep) {
	go func() {
		channels := []string{
			"footstep.create",
			fmt.Sprintf("footstep.create.%s", data.ID),
		}
		if data.UserID != nil {
			channels = append(channels, fmt.Sprintf("footstep.create.user.%s", data.UserID))
		}
		b.broadcast.Dispatch(channels, b.collection.ToModel(data))
	}()
}

func (b *FootstepBroadcast) OnUpdate(data *collection.Footstep) {
	go func() {
		channels := []string{
			"footstep.update",
			fmt.Sprintf("footstep.update.%s", data.ID),
		}
		if data.UserID != nil {
			channels = append(channels, fmt.Sprintf("footstep.update.user.%s", data.UserID))
		}
		b.broadcast.Dispatch(channels, b.collection.ToModel(data))
	}()
}

func (b *FootstepBroadcast) OnDelete(data *collection.Footstep) {
	go func() {
		channels := []string{
			"footstep.delete",
			fmt.Sprintf("footstep.delete.%s", data.ID),
		}
		if data.UserID != nil {
			channels = append(channels, fmt.Sprintf("footstep.delete.user.%s", data.UserID))
		}
		b.broadcast.Dispatch(channels, b.collection.ToModel(data))
	}()
}
