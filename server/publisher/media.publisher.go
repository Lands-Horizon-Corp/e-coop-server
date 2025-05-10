package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) MediaOnCreate(data *model.Media) {
	go func() {
		b.broadcast.Dispatch([]string{
			"media.create",
			fmt.Sprintf("media.create.%s", data.ID),
		}, b.model.MediaModel(data))
	}()
}

func (b *Publisher) MediaOnUpdate(data *model.Media) {
	go func() {
		b.broadcast.Dispatch([]string{
			"media.update",
			fmt.Sprintf("media.update.%s", data.ID),
		}, b.model.MediaModel(data))
	}()

}

func (b *Publisher) MediaOnDelete(data *model.Media) {
	go func() {
		b.broadcast.Dispatch([]string{
			"media.delete",
			fmt.Sprintf("media.delete.%s", data.ID),
		}, b.model.MediaModel(data))
	}()
}
