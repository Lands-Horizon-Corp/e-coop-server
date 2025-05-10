package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) OrganizationOnCreate(data *model.Organization) {
	go func() {
		b.broadcast.Dispatch([]string{
			"organization.create",
			fmt.Sprintf("organization.create.%s", data.ID),
			fmt.Sprintf("organization.create.user.%s", data.CreatedByID),
		}, b.model.OrganizationModel(data))
	}()
}

func (b *Publisher) OrganizationOnUpdate(data *model.Organization) {
	go func() {
		b.broadcast.Dispatch([]string{
			"organization.update",
			fmt.Sprintf("organization.update.%s", data.ID),
			fmt.Sprintf("organization.update.user.%s", data.CreatedByID),
		}, b.model.OrganizationModel(data))
	}()

}

func (b *Publisher) OrganizationOnDelete(data *model.Organization) {
	go func() {
		b.broadcast.Dispatch([]string{
			"organization.delete",
			fmt.Sprintf("organization.delete.%s", data.ID),
			fmt.Sprintf("organization.delete.user.%s", data.CreatedByID),
		}, b.model.OrganizationModel(data))
	}()
}
