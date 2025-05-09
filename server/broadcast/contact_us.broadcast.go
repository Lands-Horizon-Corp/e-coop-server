package broadcast

import (
	"fmt"

	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
)

type ContactUsBroadcast struct {
	broadcast  *horizon.HorizonBroadcast
	collection *collection.ContactUsCollection
}

func NewContactUsBroadcast(
	broadcast *horizon.HorizonBroadcast,
	collection *collection.ContactUsCollection,
) (*ContactUsBroadcast, error) {
	return &ContactUsBroadcast{
		broadcast:  broadcast,
		collection: collection,
	}, nil
}

func (b *ContactUsBroadcast) OnCreate(data *collection.ContactUs) {
	go func() {
		b.broadcast.Dispatch([]string{
			"contact_us.create",
			fmt.Sprintf("contact_us.create.%s", data.ID),
		}, b.collection.ToModel(data))
	}()
}

func (b *ContactUsBroadcast) OnUpdate(data *collection.ContactUs) {
	go func() {
		b.broadcast.Dispatch([]string{
			"contact_us.update",
			fmt.Sprintf("contact_us.update.%s", data.ID),
		}, b.collection.ToModel(data))
	}()

}

func (b *ContactUsBroadcast) OnDelete(data *collection.ContactUs) {
	go func() {
		b.broadcast.Dispatch([]string{
			"contact_us.delete",
			fmt.Sprintf("contact_us.delete.%s", data.ID),
		}, b.collection.ToModel(data))
	}()
}
