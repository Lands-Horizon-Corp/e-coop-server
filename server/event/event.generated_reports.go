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
	fmt.Printf("Starting GeneratedReportDownload for report: %+v\n", generatedReport)
	if err := e.core.GeneratedReportManager.Create(ctx, generatedReport); err != nil {
		fmt.Printf("Failed to create generated report: %v\n", err)
		return nil, eris.Wrapf(err, "Failed to create generated report")
	}
	fmt.Printf("Created report with ID: %s\n", generatedReport.ID)
	id := generatedReport.ID
	go func() {
		fmt.Printf("Starting background processing for report ID: %s\n", id)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		generatedReport.Status = core.GeneratedReportStatusInProgress
		fmt.Printf("Setting report status to in-progress\n")
		if updateErr := e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport); updateErr != nil {
			fmt.Printf("Failed to update report status to in-progress: %v\n", updateErr)
			generatedReport.Status = core.GeneratedReportStatusFailed
			generatedReport.SystemMessage = "Failed to update report status to in-progress: " + updateErr.Error()
			e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
			return
		}
		fmt.Printf("Successfully updated report status to in-progress\n")

		fmt.Printf("Starting report data processing...\n")
		data, err := e.processReportGeneration(ctx, generatedReport)
		fmt.Printf("Report data processing completed. Error: %v, Data size: %d bytes\n", err, len(data))

		// Get the latest report data
		fmt.Printf("Retrieving latest report data for ID: %s\n", id)
		generatedReport, getErr := e.core.GeneratedReportManager.GetByID(ctx, id)
		if getErr != nil {
			fmt.Printf("Failed to retrieve report after processing: %v\n", getErr)
			generatedReport.Status = core.GeneratedReportStatusFailed
			generatedReport.SystemMessage = "Failed to retrieve report after processing: " + getErr.Error()
			e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
			return
		}
		fmt.Printf("Successfully retrieved latest report data\n")
		// Upload the generated data to media storage
		if data != nil {
			fmt.Printf("Starting media upload process. Data size: %d bytes\n", len(data))
			fileName := fmt.Sprintf("report_%s.csv", generatedReport.Name)
			contentType := "text/csv"

			// Create initial media record
			fmt.Printf("Creating initial media record with fileName: %s\n", fileName)
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
				fmt.Printf("Failed to create media record: %v\n", mediaErr)
				generatedReport.Status = core.GeneratedReportStatusFailed
				generatedReport.SystemMessage = "Failed to create media record: " + mediaErr.Error()
				e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
				return
			}
			fmt.Printf("Successfully created media record with ID: %s\n", initial.ID)

			// Upload binary data
			fmt.Printf("Starting binary upload to storage...\n")
			storage, uploadErr := e.provider.Service.Storage.UploadFromBinaryWithContentType(
				ctx,
				data,
				fileName,
				contentType, func(progress, _ int64, _ *horizon.Storage) {
					fmt.Printf("Upload progress: %d%%\n", progress)
					initial.Progress = progress
					initial.UpdatedAt = time.Now().UTC()
					initial.Status = "progress"
					_ = e.core.MediaManager.UpdateByID(ctx, initial.ID, initial)
				})

			if uploadErr != nil {
				fmt.Printf("Binary upload failed: %v\n", uploadErr)
				initial.UpdatedAt = time.Now().UTC()
				initial.Status = "error"
				_ = e.core.MediaManager.UpdateByID(ctx, initial.ID, initial)
				generatedReport.Status = core.GeneratedReportStatusFailed
				generatedReport.SystemMessage = "File upload failed: " + uploadErr.Error()
				e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
				return
			}
			fmt.Printf("Binary upload completed successfully\n")

			// Update media record with final details
			fmt.Printf("Updating media record with final details\n")
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
				fmt.Printf("Failed to update media record after upload: %v\n", mediaUpdateErr)
				generatedReport.Status = core.GeneratedReportStatusFailed
				generatedReport.SystemMessage = "Failed to update media record after upload: " + mediaUpdateErr.Error()
				e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
				return
			}
			fmt.Printf("Successfully updated media record\n")

			// Link the media to the generated report
			generatedReport.MediaID = &completed.ID
			fmt.Printf("Linked media ID %s to report\n", completed.ID)
		}

		// Update status and system message based on processing result
		fmt.Printf("Updating final report status. Error from processing: %v\n", err)
		if err != nil {
			generatedReport.Status = core.GeneratedReportStatusFailed
			generatedReport.SystemMessage = "Report generation failed: " + err.Error()
			fmt.Printf("Set report status to failed\n")
		} else {
			generatedReport.Status = core.GeneratedReportStatusCompleted
			generatedReport.SystemMessage = "Report generated successfully"
			fmt.Printf("Set report status to completed\n")
		}

		// Final update with result
		fmt.Printf("Performing final update of report\n")
		if finalUpdateErr := e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport); finalUpdateErr != nil {
			fmt.Printf("Failed to perform final update: %v\n", finalUpdateErr)
			e.core.FootstepManager.Create(ctx, &core.Footstep{
				OrganizationID: &generatedReport.OrganizationID,
				BranchID:       &generatedReport.BranchID,
				UserID:         &generatedReport.CreatedByID,
				Module:         "generated_report_final_update_failed",
				Description:    "Failed to update generated report with final status after processing: " + finalUpdateErr.Error()})
		} else {
			fmt.Printf("Final update completed successfully\n")
		}
	}()
	fmt.Printf("Returning generated report from main function\n")
	return generatedReport, nil
}

// Add your background processing logic here
func (e *Event) processReportGeneration(ctx context.Context, generatedReport *core.GeneratedReport) ([]byte, error) {
	fmt.Printf("Starting processReportGeneration for type: %s, URL: %s\n", generatedReport.GeneratedReportType, generatedReport.URL)
	var data []byte
	var err error
	switch generatedReport.GeneratedReportType {
	case core.GeneratedReportTypeExcel:
		fmt.Printf("Processing Excel report type\n")
		extractor := handlers.NewRouteHandlerExtractor[[]byte](generatedReport.URL)
		// [Start Reports Excel] ===============================================================================================

		fmt.Printf("Matching route: /api/v1/bank/search\n")
		data, err = extractor.MatchableRoute("/api/v1/bank/search", func(params ...string) ([]byte, error) {
			fmt.Printf("Executing bank search with params: %v\n", params)
			fmt.Printf("Filter search query: %s\n", generatedReport.FilterSearch)
			return e.core.BankManager.FilterFieldsCSV(ctx, generatedReport.FilterSearch, &core.Bank{
				OrganizationID: generatedReport.OrganizationID,
				BranchID:       generatedReport.BranchID,
			})
		})
		fmt.Printf("Route matching completed. Data length: %d, Error: %v\n", len(data), err)

		// [End Reports Excel] ===============================================================================================
	case core.GeneratedReportTypePDF:
		fmt.Printf("Processing PDF report type (not implemented)\n")
	default:
		fmt.Printf("Unknown report type: %s\n", generatedReport.GeneratedReportType)
	}
	fmt.Printf("processReportGeneration completed. Data length: %d, Error: %v\n", len(data), err)
	return data, err
}
