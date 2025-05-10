package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) UserOrganizationOnCreate(data *model.UserOrganization) {
	go func() {
		b.broadcast.Dispatch([]string{
			"user_organization.create",
			fmt.Sprintf("user_organization.create.%s", data.ID),
			fmt.Sprintf("user_organization.create.branch.%s", data.BranchID),
			fmt.Sprintf("user_organization.create.user.%s", data.UserID),
			fmt.Sprintf("user_organization.create.organization.%s", data.OrganizationID),
		}, b.model.UserOrganizationModel(data))
	}()
}

func (b *Publisher) UserOrganizationOnUpdate(data *model.UserOrganization) {
	go func() {
		b.broadcast.Dispatch([]string{
			"user_organization.update",
			fmt.Sprintf("user_organization.update.%s", data.ID),
			fmt.Sprintf("user_organization.update.branch.%s", data.BranchID),
			fmt.Sprintf("user_organization.update.user.%s", data.UserID),
			fmt.Sprintf("user_organization.update.organization.%s", data.OrganizationID),
		}, b.model.UserOrganizationModel(data))
	}()

}

func (b *Publisher) UserOrganizationOnDelete(data *model.UserOrganization) {
	go func() {
		b.broadcast.Dispatch([]string{
			"user_organization.delete",
			fmt.Sprintf("user_organization.delete.%s", data.ID),
			fmt.Sprintf("user_organization.delete.branch.%s", data.BranchID),
			fmt.Sprintf("user_organization.delete.user.%s", data.UserID),
			fmt.Sprintf("user_organization.delete.organization.%s", data.OrganizationID),
		}, b.model.UserOrganizationModel(data))
	}()
}
