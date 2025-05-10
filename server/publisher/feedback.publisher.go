package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) FeedbackOnCreate(data *model.Feedback) {
	go func() {
		b.broadcast.Dispatch([]string{
			"feedback.create",
			fmt.Sprintf("feedback.create.%s", data.ID),
		}, b.model.FeedbackModel(data))
	}()
}

func (b *Publisher) FeedbackOnUpdate(data *model.Feedback) {
	go func() {
		b.broadcast.Dispatch([]string{
			"feedback.update",
			fmt.Sprintf("feedback.update.%s", data.ID),
		}, b.model.FeedbackModel(data))
	}()

}

func (b *Publisher) FeedbackOnDelete(data *model.Feedback) {
	go func() {
		b.broadcast.Dispatch([]string{
			"feedback.delete",
			fmt.Sprintf("feedback.delete.%s", data.ID),
		}, b.model.FeedbackModel(data))
	}()
}
