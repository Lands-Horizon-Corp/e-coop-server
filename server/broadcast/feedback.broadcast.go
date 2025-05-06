package broadcast

import (
	"fmt"

	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
)

type FeedbackBroadcast struct {
	broadcast *horizon.HorizonBroadcast
}

func NewFeedbackBroadcast(broadcast *horizon.HorizonBroadcast) (*FeedbackBroadcast, error) {
	return &FeedbackBroadcast{
		broadcast: broadcast,
	}, nil
}

func (b *FeedbackBroadcast) OnCreate(data *collection.Feedback) {
	b.broadcast.Dispatch([]string{
		"feedback.create",
		fmt.Sprintf("feedback.create.%s", data.ID),
	}, data)
}

func (b *FeedbackBroadcast) OnUpdate(data *collection.Feedback) {
	b.broadcast.Dispatch([]string{
		"feedback.update",
		fmt.Sprintf("feedback.update.%s", data.ID),
	}, data)

}
func (b *FeedbackBroadcast) OnDelete(data *collection.Feedback) {
	b.broadcast.Dispatch([]string{
		"feedback.delete",
		fmt.Sprintf("feedback.delete.%s", data.ID),
	}, data)
}
