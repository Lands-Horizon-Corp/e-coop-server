package event

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/report"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
)

type Event struct {
	core                  *core.Core
	userOrganizationToken *tokens.UserOrganizationToken
	userToken             *tokens.UserToken
	provider              *server.Provider
	usecase               *usecase.UsecaseService
	report                *report.Reports
}

func NewEvent(
	core *core.Core,
	userOrganizationToken *tokens.UserOrganizationToken,
	userToken *tokens.UserToken,
	provider *server.Provider,
	usecase *usecase.UsecaseService,
	report *report.Reports,
) (*Event, error) {
	return &Event{
		userOrganizationToken: userOrganizationToken,
		userToken:             userToken,
		core:                  core,
		provider:              provider,
		usecase:               usecase,
		report:                report,
	}, nil
}
