package report

import (
	"context"

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

func (r *Reports) Generate(context context.Context, generatedReport core.GeneratedReport) (result []byte, err error) {
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
	result, err = r.bankReport(context, data)
	result, err = r.loanTransactionReport(context, data)
	return result, err
}
