package event

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/report"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
)

type Event struct {
	core     *core.Core
	token    *tokens.Token
	provider *server.Provider
	usecase  *usecase.UsecaseService
	report   *report.Reports
}

func NewEvent(
	core *core.Core,
	token *tokens.Token,
	provider *server.Provider,
	usecase *usecase.UsecaseService,
	report *report.Reports,
) (*Event, error) {
	return &Event{
		token:    token,
		core:     core,
		provider: provider,
		usecase:  usecase,
		report:   report,
	}, nil
}
