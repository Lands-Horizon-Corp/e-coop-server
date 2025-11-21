package report

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (r *Reports) loanTransactionReport(ctx context.Context, data ReportData) (result []byte, err error) {

	switch data.generated.GeneratedReportType {
	case core.GeneratedReportTypeExcel:
	case core.GeneratedReportTypePDF:
		result, err = data.extractor.MatchableRoute("/api/v1/loan-transaction/:loan_transaction_id", func(params ...string) ([]byte, error) {
			loanTransactionID, err := uuid.Parse(params[0])
			if err != nil {
				return nil, eris.Wrapf(err, "Invalid loan transaction ID: %s", params[0])
			}
			loanTransaction, getErr := r.core.LoanTransactionManager.GetByID(ctx, loanTransactionID)
			if getErr != nil {
				return nil, eris.Wrapf(getErr, "Failed to get loan transaction by ID: %s", loanTransactionID)
			}
			pdfBytes, genErr := data.report.Generate(ctx, loanTransaction)
			if genErr != nil {
				return nil, eris.Wrapf(genErr, "Failed to generate PDF for loan transaction: %s", loanTransactionID)
			}
			return pdfBytes, nil
		})
	}
	return result, err
}
