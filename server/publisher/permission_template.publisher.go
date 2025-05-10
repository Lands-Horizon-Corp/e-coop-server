package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) PermissionTemplateOnCreate(data *model.PermissionTemplate) {
	go func() {
		b.broadcast.Dispatch([]string{
			"permission_template.create",
			fmt.Sprintf("permission_template.create.%s", data.ID),
			fmt.Sprintf("permission_template.create.organization.%s", data.OrganizationID),
			fmt.Sprintf("permission_template.delete.branch.%s", data.BranchID),
		}, b.model.PermissionTemplateModel(data))
	}()
}

func (b *Publisher) PermissionTemplateOnUpdate(data *model.PermissionTemplate) {
	go func() {
		b.broadcast.Dispatch([]string{
			"permission_template.update",
			fmt.Sprintf("permission_template.update.%s", data.ID),
			fmt.Sprintf("permission_template.update.organization.%s", data.OrganizationID),
			fmt.Sprintf("permission_template.delete.branch.%s", data.BranchID),
		}, b.model.PermissionTemplateModel(data))
	}()

}

func (b *Publisher) PermissionTemplateOnDelete(data *model.PermissionTemplate) {
	go func() {
		b.broadcast.Dispatch([]string{
			"permission_template.delete",
			fmt.Sprintf("permission_template.delete.%s", data.ID),
			fmt.Sprintf("permission_template.delete.organization.%s", data.OrganizationID),
			fmt.Sprintf("permission_template.delete.branch.%s", data.BranchID),
		}, b.model.PermissionTemplateModel(data))
	}()
}
