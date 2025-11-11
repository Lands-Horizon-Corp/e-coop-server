package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
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

		data, err := e.processReportGeneration(ctx, generatedReport)

		// Get the latest report data
		generatedReport, getErr := e.core.GeneratedReportManager.GetByID(ctx, id)
		if getErr != nil {
			generatedReport.Status = core.GeneratedReportStatusFailed
			generatedReport.SystemMessage = "Failed to retrieve report after processing: " + getErr.Error()
			e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
			return
		}
		// Upload the generated data to media storage
		if data != nil {
			fileName := fmt.Sprintf("report_%s.csv", generatedReport.Name)
			contentType := "text/csv"

			// Create initial media record
			initial := &core.Media{
				FileName:   fileName,
				FileSize:   0,
				FileType:   contentType,
				StorageKey: "",
				BucketName: "",
				Status:     "pending",
				Progress:   0,
				CreatedAt:  time.Now().UTC(),
				UpdatedAt:  time.Now().UTC(),
			}

			if mediaErr := e.core.MediaManager.Create(ctx, initial); mediaErr != nil {
				generatedReport.Status = core.GeneratedReportStatusFailed
				generatedReport.SystemMessage = "Failed to create media record: " + mediaErr.Error()
				e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
				return
			}

			// Upload binary data
			storage, uploadErr := e.provider.Service.Storage.UploadFromBinary(ctx, data, func(progress, _ int64, _ *horizon.Storage) {
				_ = e.core.MediaManager.UpdateByID(ctx, initial.ID, &core.Media{
					Progress:  progress,
					Status:    "progress",
					UpdatedAt: time.Now().UTC(),
				})
			})

			if uploadErr != nil {
				_ = e.core.MediaManager.UpdateByID(ctx, initial.ID, &core.Media{
					ID:        initial.ID,
					Status:    "error",
					UpdatedAt: time.Now().UTC(),
				})
				generatedReport.Status = core.GeneratedReportStatusFailed
				generatedReport.SystemMessage = "File upload failed: " + uploadErr.Error()
				e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
				return
			}

			// Update media record with final details
			completed := &core.Media{
				FileName:   storage.FileName,
				FileType:   storage.FileType,
				FileSize:   storage.FileSize,
				StorageKey: storage.StorageKey,
				BucketName: storage.BucketName,
				Status:     "completed",
				Progress:   100,
				CreatedAt:  initial.CreatedAt,
				UpdatedAt:  time.Now().UTC(),
				ID:         initial.ID,
			}

			if mediaUpdateErr := e.core.MediaManager.UpdateByID(ctx, completed.ID, completed); mediaUpdateErr != nil {
				generatedReport.Status = core.GeneratedReportStatusFailed
				generatedReport.SystemMessage = "Failed to update media record after upload: " + mediaUpdateErr.Error()
				e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
				return
			}

			// Link the media to the generated report
			generatedReport.MediaID = &completed.ID
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
func (e *Event) processReportGeneration(ctx context.Context, generatedReport *core.GeneratedReport) ([]byte, error) {
	var data []byte
	var err error
	switch generatedReport.GeneratedReportType {
	case core.GeneratedReportTypeExcel:
		extractor := handlers.NewRouteHandlerExtractor[[]byte](generatedReport.URL)
		// [Start Reports Excel] ===============================================================================================

		data, err = extractor.MatchableRoute("/api/v1/bank/search", func(params ...string) ([]byte, error) {
			return e.core.BankManager.FilterFieldsCSV(ctx, generatedReport.FilterSearch, &core.Bank{
				OrganizationID: generatedReport.OrganizationID,
				BranchID:       generatedReport.BranchID,
			})
		})

		// [End Reports Excel] ===============================================================================================
	case core.GeneratedReportTypePDF:
	}
	return data, err
}
