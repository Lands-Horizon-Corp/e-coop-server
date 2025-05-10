package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) OrganizationCategoryOnCreate(data *model.OrganizationCategory) {
	go func() {
		b.broadcast.Dispatch([]string{
			"organization_category.create",
			fmt.Sprintf("organization_category.create.%s", data.ID),
			fmt.Sprintf("organization_category.create.organization.%s", data.OrganizationID),
		}, b.model.OrganizationCategoryModel(data))
	}()
}

func (b *Publisher) OrganizationCategoryOnUpdate(data *model.OrganizationCategory) {
	go func() {
		b.broadcast.Dispatch([]string{
			"organization_category.update",
			fmt.Sprintf("organization_category.update.%s", data.ID),
			fmt.Sprintf("organization_category.update.organization.%s", data.OrganizationID),
		}, b.model.OrganizationCategoryModel(data))
	}()

}

func (b *Publisher) OrganizationCategoryOnDelete(data *model.OrganizationCategory) {
	go func() {
		b.broadcast.Dispatch([]string{
			"organization_category.delete",
			fmt.Sprintf("organization_category.delete.%s", data.ID),
			fmt.Sprintf("organization_category.delete.organization.%s", data.OrganizationID),
		}, b.model.OrganizationCategoryModel(data))
	}()
}
