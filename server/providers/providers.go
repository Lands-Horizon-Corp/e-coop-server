package providers

import (
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

type Providers struct {
	model          *model.Model
	authentication *horizon.HorizonAuthentication

	footstep         *model.FootstepCollection
	notification     *model.NotificationCollection
	userOrganization *model.UserOrganizationCollection
	user             *model.UserCollection
}

func NewProviders(
	model *model.Model,
	authentication *horizon.HorizonAuthentication,

	footstep *model.FootstepCollection,
	notification *model.NotificationCollection,
	userOrganization *model.UserOrganizationCollection,
	user *model.UserCollection,
) (*Providers, error) {
	return &Providers{
		footstep:         footstep,
		notification:     notification,
		userOrganization: userOrganization,
		user:             user,
		model:            model,
		authentication:   authentication,
	}, nil
}
