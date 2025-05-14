package providers

import (
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

type Providers struct {
	model          *model.Model
	authentication *horizon.HorizonAuthentication
	customAuth     *horizon.HorizonAuthCustom

	footstep         *model.FootstepCollection
	notification     *model.NotificationCollection
	userOrganization *model.UserOrganizationCollection
	user             *model.UserCollection
}

func NewProviders(
	model *model.Model,
	authentication *horizon.HorizonAuthentication,
	customAuth *horizon.HorizonAuthCustom,

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
		customAuth:       customAuth,
	}, nil
}

func (p *Providers) CleanCustomToken(ctx echo.Context) {
	p.customAuth.CleanToken(ctx)
}

func (p *Providers) CleanToken(ctx echo.Context) {
	p.authentication.CleanToken(ctx)
	p.customAuth.CleanToken(ctx)
}
