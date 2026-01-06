package event

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/report"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/tokens"
)

type Event struct {
	core     *core.Core
	token    *tokens.Token
	provider *server.Provider
	report   *report.Reports
}

func NewEvent(
	core *core.Core,
	token *tokens.Token,
	provider *server.Provider,
	report *report.Reports,
) (*Event, error) {
	return &Event{
		token:    token,
		core:     core,
		provider: provider,
		report:   report,
	}, nil
}
