package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) UserOnCreate(data *model.User) {
	go func() {
		b.broadcast.Dispatch([]string{
			"user.create",
			fmt.Sprintf("user.create.%s", data.ID),
		}, b.model.UserModel(data))
	}()
}

func (b *Publisher) UserOnUpdate(data *model.User) {
	go func() {
		b.broadcast.Dispatch([]string{
			"user.update",
			fmt.Sprintf("user.update.%s", data.ID),
		}, b.model.UserModel(data))
	}()

}

func (b *Publisher) UserOnDelete(data *model.User) {
	go func() {
		b.broadcast.Dispatch([]string{
			"user.delete",
			fmt.Sprintf("user.delete.%s", data.ID),
		}, b.model.UserModel(data))
	}()
}
