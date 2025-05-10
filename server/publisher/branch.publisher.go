package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) BranchOnCreate(data *model.Branch) {
	go func() {
		b.broadcast.Dispatch([]string{
			"branch.create",
			fmt.Sprintf("branch.create.%s", data.ID),
			fmt.Sprintf("branch.create.organization.%s", data.OrganizationID),
		}, b.model.BranchModel(data))
	}()
}

func (b *Publisher) BranchOnUpdate(data *model.Branch) {
	go func() {
		b.broadcast.Dispatch([]string{
			"branch.update",
			fmt.Sprintf("branch.update.%s", data.ID),
			fmt.Sprintf("branch.update.organization.%s", data.OrganizationID),
		}, b.model.BranchModel(data))
	}()

}

func (b *Publisher) BranchOnDelete(data *model.Branch) {
	go func() {
		b.broadcast.Dispatch([]string{
			"branch.delete",
			fmt.Sprintf("branch.delete.%s", data.ID),
			fmt.Sprintf("branch.delete.organization.%s", data.OrganizationID),
		}, b.model.BranchModel(data))
	}()
}
