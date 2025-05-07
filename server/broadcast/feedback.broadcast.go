package broadcast

import (
	"fmt"

	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
)

type FeedbackBroadcast struct {
	broadcast  *horizon.HorizonBroadcast
	collection *collection.FeedbackCollection
}

func NewFeedbackBroadcast(
	broadcast *horizon.HorizonBroadcast,
	collection *collection.FeedbackCollection,
) (*FeedbackBroadcast, error) {
	return &FeedbackBroadcast{
		broadcast:  broadcast,
		collection: collection,
	}, nil
}

func (b *FeedbackBroadcast) OnCreate(data *collection.Feedback) {
	go func() {
		b.broadcast.Dispatch([]string{
			"feedback.create",
			fmt.Sprintf("feedback.create.%s", data.ID),
		}, b.collection.ToModel(data))
	}()
}

func (b *FeedbackBroadcast) OnUpdate(data *collection.Feedback) {
	go func() {
		b.broadcast.Dispatch([]string{
			"feedback.update",
			fmt.Sprintf("feedback.update.%s", data.ID),
		}, b.collection.ToModel(data))
	}()

}
func (b *FeedbackBroadcast) OnDelete(data *collection.Feedback) {
	go func() {
		b.broadcast.Dispatch([]string{
			"feedback.delete",
			fmt.Sprintf("feedback.delete.%s", data.ID),
		}, b.collection.ToModel(data))
	}()
}
