package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) ContactUsOnCreate(data *model.ContactUs) {
	go func() {
		b.broadcast.Dispatch([]string{
			"contact_us.create",
			fmt.Sprintf("contact_us.create.%s", data.ID),
		}, b.model.ContactUsModel(data))
	}()
}

func (b *Publisher) ContactUsOnUpdate(data *model.ContactUs) {
	go func() {
		b.broadcast.Dispatch([]string{
			"contact_us.update",
			fmt.Sprintf("contact_us.update.%s", data.ID),
		}, b.model.ContactUsModel(data))
	}()

}

func (b *Publisher) ContactUsOnDelete(data *model.ContactUs) {
	go func() {
		b.broadcast.Dispatch([]string{
			"contact_us.delete",
			fmt.Sprintf("contact_us.delete.%s", data.ID),
		}, b.model.ContactUsModel(data))
	}()
}
