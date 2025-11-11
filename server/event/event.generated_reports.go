package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
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
		e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
		err := e.processReportGeneration(ctx, id)
		generatedReport, getErr := e.core.GeneratedReportManager.GetByID(ctx, id)
		if getErr != nil {
			return
		}
		if err != nil {
			generatedReport.Status = core.GeneratedReportStatusFailed
		} else {
			generatedReport.Status = core.GeneratedReportStatusCompleted
		}
		e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
	}()

	return generatedReport, nil
}

// Add your background processing logic here
func (e *Event) processReportGeneration(ctx context.Context, reportID uuid.UUID) error {
	// Your actual report generation/upload logic goes here
	// Example:
	// 1. base 64 to real value
	// 2. fetch SQL
	// 3. procession to excel
	return nil
}
