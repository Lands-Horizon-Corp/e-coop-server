package event

import (
	"fmt"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/report"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
)

type Event struct {
	core     *core.Core
	provider *server.Provider
	report   *report.Reports

	userOrgCSRF horizon.AuthService[UserOrganizationCSRF]
	userCSRF    horizon.AuthService[UserCSRF]
}

func NewEvent(
	core *core.Core,
	provider *server.Provider,
	report *report.Reports,
) (*Event, error) {

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

	return &Event{
		core:        core,
		provider:    provider,
		report:      report,
		userCSRF:    userCSRF,
		userOrgCSRF: userOrgCSRF,
	}, nil
}
