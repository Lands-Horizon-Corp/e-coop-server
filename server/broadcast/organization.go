package broadcast

import (
	"fmt"

	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
)

type OrganizationBroadcast struct {
	broadcast  *horizon.HorizonBroadcast
	collection *collection.OrganizationCollection
}

func NewOrganizationBroadcast(
	broadcast *horizon.HorizonBroadcast,
	collection *collection.OrganizationCollection,
) (*OrganizationBroadcast, error) {
	return &OrganizationBroadcast{
		broadcast:  broadcast,
		collection: collection,
	}, nil
}

func (b *OrganizationBroadcast) OnCreate(data *collection.Organization) {
	go func() {
		b.broadcast.Dispatch([]string{
			"Organization.create",
			fmt.Sprintf("Organization.create.%s", data.ID),
		}, b.collection.ToModel(data))
	}()
}

func (b *OrganizationBroadcast) OnUpdate(data *collection.Organization) {
	go func() {
		b.broadcast.Dispatch([]string{
			"Organization.update",
			fmt.Sprintf("Organization.update.%s", data.ID),
		}, b.collection.ToModel(data))
	}()

}

func (b *OrganizationBroadcast) OnDelete(data *collection.Organization) {
	go func() {
		b.broadcast.Dispatch([]string{
			"Organization.delete",
			fmt.Sprintf("Organization.delete.%s", data.ID),
		}, b.collection.ToModel(data))
	}()
}
