package broadcast

import (
	"horizon.com/server/horizon"
	"horizon.com/server/models/collection"
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
	}, data)
}

func (b *FeedbackBroadcast) OnUpdate(data *collection.Feedback) {
	b.broadcast.Dispatch([]string{
		"feedback.update",
	}, data)

}
func (b *FeedbackBroadcast) OnDelete(data *collection.Feedback) {
	b.broadcast.Dispatch([]string{
		"feedback.delete",
	}, data)
}
