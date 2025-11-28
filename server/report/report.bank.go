package report

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
)

func (r *Reports) bankReport(ctx context.Context, data ReportData) (result []byte, err error) {

	switch data.generated.GeneratedReportType {
	case core.GeneratedReportTypeExcel:
		result, err = data.extractor.MatchableRoute("/api/v1/bank/search", func(params ...string) ([]byte, error) {
			return r.core.BankManager.FilterFieldsCSV(ctx, data.generated.FilterSearch, &core.Bank{
				OrganizationID: data.generated.OrganizationID,
				BranchID:       data.generated.BranchID,
			})
		})
	case core.GeneratedReportTypePDF:
	}

	return result, err
}
