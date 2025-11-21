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
	usecase               *usecase.TransactionService
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
	usecase *usecase.TransactionService,

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

// Generate now returns a slice of results (one entry per handler that produced output).
// Callers that only expect one result can pick the first element of the returned slice.
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

	type handlerEntry struct {
		name string
		fn   func(context.Context, ReportData) ([]byte, error)
	}

	handlersList := []handlerEntry{
		{name: "bankReport", fn: r.bankReport},
		{name: "loanTransactionReport", fn: r.loanTransactionReport},
	}

	for _, h := range handlersList {
		res, err := h.fn(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("handler %s error: %w", h.name, err)
		}
		if res != nil {
			return res, nil
		}
	}
	return nil, nil
}
