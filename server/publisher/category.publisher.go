package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) CategoryOnCreate(data *model.Category) {
	go func() {
		b.broadcast.Dispatch([]string{
			"category.create",
			fmt.Sprintf("category.create.%s", data.ID),
		}, b.model.CategoryModel(data))
	}()
}

func (b *Publisher) CategoryOnUpdate(data *model.Category) {
	go func() {
		b.broadcast.Dispatch([]string{
			"category.update",
			fmt.Sprintf("category.update.%s", data.ID),
		}, b.model.CategoryModel(data))
	}()

}

func (b *Publisher) CategoryOnDelete(data *model.Category) {
	go func() {
		b.broadcast.Dispatch([]string{
			"category.delete",
			fmt.Sprintf("category.delete.%s", data.ID),
		}, b.model.CategoryModel(data))
	}()
}
