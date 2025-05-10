package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) FootstepOnCreate(data *model.Footstep) {
	go func() {
		b.broadcast.Dispatch([]string{
			fmt.Sprintf("footstep.create.%s", data.ID),
			fmt.Sprintf("footstep.create.banch.%s", data.BranchID),
			fmt.Sprintf("footstep.create.organization.%s", data.OrganizationID),
			fmt.Sprintf("footstep.create.user.%s", data.UserID),
		}, b.model.FootstepModel(data))
	}()
}

func (b *Publisher) FootstepOnUpdate(data *model.Footstep) {
	go func() {
		b.broadcast.Dispatch([]string{
			"footstep.update",
			fmt.Sprintf("footstep.update.%s", data.ID),
			fmt.Sprintf("footstep.update.banch.%s", data.BranchID),
			fmt.Sprintf("footstep.update.organization.%s", data.OrganizationID),
			fmt.Sprintf("footstep.update.user.%s", data.UserID),
		}, b.model.FootstepModel(data))
	}()

}

func (b *Publisher) FootstepOnDelete(data *model.Footstep) {
	go func() {
		b.broadcast.Dispatch([]string{
			"footstep.delete",
			fmt.Sprintf("footstep.delete.%s", data.ID),
			fmt.Sprintf("footstep.delete.banch.%s", data.BranchID),
			fmt.Sprintf("footstep.delete.organization.%s", data.OrganizationID),
			fmt.Sprintf("footstep.delete.user.%s", data.UserID),
		}, b.model.FootstepModel(data))
	}()
}
