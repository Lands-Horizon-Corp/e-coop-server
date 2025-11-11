package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/rotisserie/eris"
)

func (e *Event) GeneratedReportDownload(ctx context.Context, generatedReport *core.GeneratedReport) (*core.GeneratedReport, error) {
	if err := e.core.GeneratedReportManager.Create(ctx, generatedReport); err != nil {
		return nil, eris.Wrapf(err, "Failed to create generated report")
	}
	id := generatedReport.ID
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		generatedReport.Status = core.GeneratedReportStatusInProgress
		if updateErr := e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport); updateErr != nil {
			generatedReport.Status = core.GeneratedReportStatusFailed
			generatedReport.SystemMessage = "Failed to update report status to in-progress: " + updateErr.Error()
			e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
			return
		}

		err := e.processReportGeneration(ctx, generatedReport)

		// Get the latest report data
		generatedReport, getErr := e.core.GeneratedReportManager.GetByID(ctx, id)
		if getErr != nil {
			generatedReport.Status = core.GeneratedReportStatusFailed
			generatedReport.SystemMessage = "Failed to retrieve report after processing: " + getErr.Error()
			e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
			return
		}

		// Update status and system message based on processing result
		if err != nil {
			generatedReport.Status = core.GeneratedReportStatusFailed
			generatedReport.SystemMessage = "Report generation failed: " + err.Error()
		} else {
			generatedReport.Status = core.GeneratedReportStatusCompleted
			generatedReport.SystemMessage = "Report generated successfully"
		}

		// Final update with result
		if finalUpdateErr := e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport); finalUpdateErr != nil {
			e.core.FootstepManager.Create(ctx, &core.Footstep{
				OrganizationID: &generatedReport.OrganizationID,
				BranchID:       &generatedReport.BranchID,
				UserID:         &generatedReport.CreatedByID,
				Module:         "generated_report_final_update_failed",
				Description:    "Failed to update generated report with final status after processing: " + finalUpdateErr.Error()})
		}
	}()
	return generatedReport, nil
}

// Add your background processing logic here
func (e *Event) processReportGeneration(ctx context.Context, generatedReport *core.GeneratedReport) error {
	switch generatedReport.GeneratedReportType {
	case core.GeneratedReportTypeExcel:
		extractor := handlers.NewRouteHandlerExtractor[[]byte](generatedReport.URL)
		// data, err :=

		// extractor := handlers.NewRouteHandlerExtractor(generatedReport.URL)
		// if err := extractor.MatchableRoute("/api/v1/bank/search", func(params ...string) error {
		// 	banks, err := e.core.BankManager.NoPaginationWithFields(ctx, generatedReport.FilterSearch, &core.Bank{
		// 		OrganizationID: generatedReport.OrganizationID,
		// 		BranchID:       generatedReport.BranchID,
		// 	})
		// 	if err != nil {
		// 		return err
		// 	}
		// 	return nil
		// }); err != nil {
		// 	return eris.Wrapf(err, "Failed to process bank report generation")
		// }

	case core.GeneratedReportTypePDF:
	}
	return nil
}
