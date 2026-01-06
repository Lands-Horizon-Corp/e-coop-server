package report

import (
	"context"
	"fmt"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
)

type Reports struct {
	provider *server.Provider
	core     *core.Core
	token    *tokens.Token
	usecase  *usecase.UsecaseService
}

type ReportData struct {
	generated core.GeneratedReport
	extractor *handlers.RouteHandlerExtractor[[]byte]
	report    handlers.PDFOptions[any]
}

func NewReports(
	provider *server.Provider,
	core *core.Core,

	token *tokens.Token,
	usecase *usecase.UsecaseService,

) (*Reports, error) {
	return &Reports{
		provider: provider,
		core:     core,
		token:    token,
		usecase:  usecase,
	}, nil
}

type handlerEntry struct {
	name string
	fn   func(context.Context, ReportData) ([]byte, error)
}

func reportHandlers(r *Reports) []handlerEntry {
	return []handlerEntry{
		{name: "bankReport", fn: r.bankReport},
		{name: "loanTransactionReport", fn: r.loanTransactionReport},
	}
}

func (r *Reports) Generate(ctx context.Context, generatedReport core.GeneratedReport) (results []byte, err error) {
	extractor := handlers.NewRouteHandlerExtractor[[]byte](generatedReport.URL)
	report := handlers.PDFOptions[any]{
		Name:      generatedReport.Name,
		Template:  generatedReport.Template,
		Height:    generatedReport.Height,
		Width:     generatedReport.Width,
		Unit:      generatedReport.Unit,
		Landscape: generatedReport.Landscape,
	}
	data := ReportData{
		generated: generatedReport,
		extractor: extractor,
		report:    report,
	}

	for _, h := range reportHandlers(r) {
		res, err := h.fn(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("handler %s error: %w", h.name, err)
		}
		if len(res) != 0 {
			return res, nil
		}
	}
	return nil, nil
}
