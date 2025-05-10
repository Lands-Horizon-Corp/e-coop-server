package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) InvitationCodeOnCreate(data *model.InvitationCode) {
	go func() {
		b.broadcast.Dispatch([]string{
			"invitation_code.create",
			fmt.Sprintf("invitation_code.create.%s", data.ID),
			fmt.Sprintf("invitation_code.create.organization.%s", data.OrganizationID),
			fmt.Sprintf("invitation_code.create.organization.%s", data.BranchID),
			fmt.Sprintf("invitation_code.create.user.%s", data.CreatedByID),
		}, b.model.InvitationCodeModel(data))
	}()
}

func (b *Publisher) InvitationCodeOnUpdate(data *model.InvitationCode) {
	go func() {
		b.broadcast.Dispatch([]string{
			"invitation_code.update",
			fmt.Sprintf("invitation_code.update.%s", data.ID),
			fmt.Sprintf("invitation_code.update.organization.%s", data.OrganizationID),
			fmt.Sprintf("invitation_code.update.organization.%s", data.BranchID),
			fmt.Sprintf("invitation_code.update.user.%s", data.CreatedByID),
		}, b.model.InvitationCodeModel(data))
	}()

}

func (b *Publisher) InvitationCodeOnDelete(data *model.InvitationCode) {
	go func() {
		b.broadcast.Dispatch([]string{
			"invitation_code.delete",
			fmt.Sprintf("invitation_code.delete.%s", data.ID),
			fmt.Sprintf("invitation_code.delete.organization.%s", data.OrganizationID),
			fmt.Sprintf("invitation_code.delete.branch.%s", data.BranchID),
			fmt.Sprintf("invitation_code.delete.user.%s", data.CreatedByID),
		}, b.model.InvitationCodeModel(data))
	}()
}
