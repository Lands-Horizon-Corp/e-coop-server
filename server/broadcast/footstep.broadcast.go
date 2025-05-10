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

		b.broadcast.Dispatch([]string{
			"footstep.create",
			fmt.Sprintf("footstep.create.%s", data.ID),
			fmt.Sprintf("footstep.create.banch.%s", data.BranchID),
			fmt.Sprintf("footstep.create.organization.%s", data.OrganizationID),
			fmt.Sprintf("footstep.create.user.%s", data.UserID),
		}, b.collection.ToModel(data))
	}()
}

func (b *FootstepBroadcast) OnUpdate(data *collection.Footstep) {
	go func() {
		b.broadcast.Dispatch([]string{
			"footstep.update",
			fmt.Sprintf("footstep.update.%s", data.ID),
			fmt.Sprintf("footstep.update.banch.%s", data.BranchID),
			fmt.Sprintf("footstep.update.organization.%s", data.OrganizationID),
			fmt.Sprintf("footstep.update.user.%s", data.UserID),
		}, b.collection.ToModel(data))
	}()
}

func (b *FootstepBroadcast) OnDelete(data *collection.Footstep) {
	go func() {

		b.broadcast.Dispatch([]string{
			"footstep.delete",
			fmt.Sprintf("footstep.delete.%s", data.ID),
			fmt.Sprintf("footstep.delete.banch.%s", data.BranchID),
			fmt.Sprintf("footstep.delete.organization.%s", data.OrganizationID),
			fmt.Sprintf("footstep.delete.user.%s", data.UserID),
		}, b.collection.ToModel(data))
	}()
}
