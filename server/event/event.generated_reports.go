package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
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
			if err := e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport); err != nil {
				e.provider.Service.Logger.Error("Failed to update generated report status after failure", zap.Error(err))
			}
			return
		}
		data, err := e.processReportGeneration(ctx, generatedReport)

		// Get the latest report data
		generatedReport, getErr := e.core.GeneratedReportManager.GetByID(ctx, id)
		if getErr != nil {
			generatedReport.Status = core.GeneratedReportStatusFailed
			generatedReport.SystemMessage = "Failed to retrieve report after processing: " + getErr.Error()
			if err := e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport); err != nil {
				e.provider.Service.Logger.Error("Failed to update generated report status after failure", zap.Error(err))
			}
			return
		}
		// Upload the generated data to media storage
		if data != nil {
			// choose file extension / content type based on generated report type
			fileExt := "csv"
			contentType := "text/csv"
			if generatedReport.GeneratedReportType == core.GeneratedReportTypePDF {
				fileExt = "pdf"
				contentType = "application/pdf"
			}
			fileName := fmt.Sprintf("report_%s.%s", generatedReport.Name, fileExt)

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
				if err := e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport); err != nil {
					e.provider.Service.Logger.Error("Failed to update generated report status after failure", zap.Error(err))
				}
				return
			}

			storage, uploadErr := e.provider.Service.Storage.UploadFromBinaryWithContentType(
				ctx,
				data,
				fileName,
				contentType, func(progress, _ int64, _ *horizon.Storage) {
					initial.Progress = progress
					initial.UpdatedAt = time.Now().UTC()
					initial.Status = "progress"
					if err := e.core.MediaManager.UpdateByID(ctx, initial.ID, initial); err != nil {
						e.provider.Service.Logger.Error("Failed to update media progress", zap.Error(err))
					}
				})

			if uploadErr != nil {
				initial.UpdatedAt = time.Now().UTC()
				initial.Status = "error"
				if err := e.core.MediaManager.UpdateByID(ctx, initial.ID, initial); err != nil {
					e.provider.Service.Logger.Error("Failed to update media status after upload error", zap.Error(err))
					return
				}
				generatedReport.Status = core.GeneratedReportStatusFailed
				generatedReport.SystemMessage = "File upload failed: " + uploadErr.Error()
				e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport)
				return
			}
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
				if err := e.core.GeneratedReportManager.UpdateByID(ctx, id, generatedReport); err != nil {
					e.provider.Service.Logger.Error("Failed to update generated report status after failure", zap.Error(err))
				}
				return
			}
			generatedReport.MediaID = &completed.ID
		}

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
	extractor := handlers.NewRouteHandlerExtractor[[]byte](generatedReport.URL)

	switch generatedReport.GeneratedReportType {
	case core.GeneratedReportTypeExcel:
		// [Start Reports Excel] ===============================================================================================
		data, err = extractor.MatchableRoute("/api/v1/bank/search", func(params ...string) ([]byte, error) {
			return e.core.BankManager.FilterFieldsCSV(ctx, generatedReport.FilterSearch, &core.Bank{
				OrganizationID: generatedReport.OrganizationID,
				BranchID:       generatedReport.BranchID,
			})
		})
		// [End Reports Excel] ===============================================================================================
	case core.GeneratedReportTypePDF:
		// [Start Reports PDF] ===============================================================================================
		report := handlers.PDFOptions[any]{
			Name:      generatedReport.Name,
			Template:  generatedReport.Template,
			Height:    generatedReport.Height,
			Width:     generatedReport.Width,
			Unit:      generatedReport.Unit,
			Landscape: generatedReport.Landscape,
		}

		// api/v1/payment/general-ledger/:general-ledger_id
		data, err = extractor.MatchableRoute("/api/v1/loan-transaction/:loan_transaction_id", func(params ...string) ([]byte, error) {
			loanTransactionID, err := uuid.Parse(params[0])
			if err != nil {
				return nil, eris.Wrapf(err, "Invalid loan transaction ID: %s", params[0])
			}
			loanTransaction, getErr := e.core.LoanTransactionManager.GetByID(ctx, loanTransactionID)
			if getErr != nil {
				return nil, eris.Wrapf(getErr, "Failed to get loan transaction by ID: %s", loanTransactionID)
			}
			pdfBytes, genErr := report.Generate(ctx, loanTransaction)
			if genErr != nil {
				return nil, eris.Wrapf(genErr, "Failed to generate PDF for loan transaction: %s", loanTransactionID)
			}
			return pdfBytes, nil
		})
		// [End Reports PDF] ===============================================================================================
	default:
	}
	return data, err
}
