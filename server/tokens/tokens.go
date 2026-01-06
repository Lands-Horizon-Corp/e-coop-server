package tokens

import (
	"fmt"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
)

type Token struct {
	core     *core.Core
	provider *server.Provider

	userOrgCSRF horizon.AuthService[UserOrganizationCSRF]
	userCSRF    horizon.AuthService[UserCSRF]
}

func NewToken(provider *server.Provider, core *core.Core) (*Token, error) {
	appName := provider.Service.Environment.GetString("APP_NAME", "")

	userOrgCSRF := horizon.NewAuthServiceImpl[UserOrganizationCSRF](
		provider.Service.Cache,
		"user-organization-csrf",
		fmt.Sprintf("%s-%s", "X-SECURE-CSRF-USER-ORGANIZATION", appName),
		true,
	)

	userCSRF := horizon.NewAuthServiceImpl[UserCSRF](
		provider.Service.Cache,
		"user-csrf",
		fmt.Sprintf("%s-%s", "X-SECURE-CSRF-USER", appName),
		true,
	)

	return &Token{
		userCSRF:    userCSRF,
		userOrgCSRF: userOrgCSRF,
		core:        core,
		provider:    provider,
	}, nil
}
