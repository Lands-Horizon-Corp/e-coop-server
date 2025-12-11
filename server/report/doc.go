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
	// Services
	provider *server.Provider
	core     *core.Core
	// Tokens
	userOrganizationToken *tokens.UserOrganizationToken
	userToken             *tokens.UserToken
	usecase               *usecase.UsecaseService
}

type ReportData struct {
	generated core.GeneratedReport
	extractor *handlers.RouteHandlerExtractor[[]byte]
	report    handlers.PDFOptions[any]
}

func NewReports(
	// Services
	provider *server.Provider,
	core *core.Core,

	// Tokens
	userOrganizationToken *tokens.UserOrganizationToken,
	userToken *tokens.UserToken,
	usecase *usecase.UsecaseService,

) (*Reports, error) {
	return &Reports{
		// Services
		provider: provider,
		core:     core,

		userOrganizationToken: userOrganizationToken,
		userToken:             userToken,
		usecase:               usecase,
	}, nil
}

// handlerEntry is package-level so the handler registration function can return it.
type handlerEntry struct {
	name string
	fn   func(context.Context, ReportData) ([]byte, error)
}

// reportHandlers returns the ordered list of handlers to run for reports.
// Edit this function to add/remove report handlers — only the list needs updating.
func reportHandlers(r *Reports) []handlerEntry {
	return []handlerEntry{
		{name: "bankReport", fn: r.bankReport},
		{name: "loanTransactionReport", fn: r.loanTransactionReport},
	}
}

// Generate now returns one or more results produced by handlers.
// The handler list is defined in reportHandlers so you only need to edit that.
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
