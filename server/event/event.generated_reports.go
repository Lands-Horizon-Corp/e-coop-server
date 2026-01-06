package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
)

type ReportData struct {
	generated core.GeneratedReport
	extractor *handlers.RouteHandlerExtractor[[]byte]
	report    handlers.PDFOptions[any]
}

type handlerEntry struct {
	name string
	fn   func(context.Context, ReportData) ([]byte, error)
}

func reportHandlers(r *Event) []handlerEntry {
	return []handlerEntry{
		{name: "bankReport", fn: r.bankReport},
		{name: "loanTransactionReport", fn: r.loanTransactionReport},
	}
}

func (e *Event) GeneratedReportDownload(ctx context.Context, generatedReport *core.GeneratedReport) (*core.GeneratedReport, error) {
	if err := e.core.GeneratedReportManager().Create(ctx, generatedReport); err != nil {
		return nil, eris.Wrapf(err, "Failed to create generated report")
	}
	id := generatedReport.ID
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		generatedReport.Status = core.GeneratedReportStatusInProgress
		if updateErr := e.core.GeneratedReportManager().UpdateByID(ctx, id, generatedReport); updateErr != nil {
			generatedReport.Status = core.GeneratedReportStatusFailed
			generatedReport.SystemMessage = "Failed to update report status to in-progress: " + updateErr.Error()
			if err := e.core.GeneratedReportManager().UpdateByID(ctx, id, generatedReport); err != nil {
				e.provider.Service.Logger.Error("Failed to update generated report status after failure", zap.Error(err))
			}
			return
		}
		data, err := e.GenerateReport(ctx, *generatedReport)

		generatedReport, getErr := e.core.GeneratedReportManager().GetByID(ctx, id)
		if getErr != nil {
			generatedReport.Status = core.GeneratedReportStatusFailed
			generatedReport.SystemMessage = "Failed to retrieve report after processing: " + getErr.Error()
			if err := e.core.GeneratedReportManager().UpdateByID(ctx, id, generatedReport); err != nil {
				e.provider.Service.Logger.Error("Failed to update generated report status after failure", zap.Error(err))
			}
			return
		}
		if data != nil {
			fileExt := "csv"
			contentType := "text/csv"
			if generatedReport.GeneratedReportType == core.GeneratedReportTypePDF {
				fileExt = "pdf"
				contentType = "application/pdf"
			}
			fileName := fmt.Sprintf("report_%s.%s", generatedReport.Name, fileExt)

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

			if mediaErr := e.core.MediaManager().Create(ctx, initial); mediaErr != nil {
				generatedReport.Status = core.GeneratedReportStatusFailed
				generatedReport.SystemMessage = "Failed to create media record: " + mediaErr.Error()
				if err := e.core.GeneratedReportManager().UpdateByID(ctx, id, generatedReport); err != nil {
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
					if err := e.core.MediaManager().UpdateByID(ctx, initial.ID, initial); err != nil {
						e.provider.Service.Logger.Error("Failed to update media progress", zap.Error(err))
					}
				})

			if uploadErr != nil {
				initial.UpdatedAt = time.Now().UTC()
				initial.Status = "error"
				if err := e.core.MediaManager().UpdateByID(ctx, initial.ID, initial); err != nil {
					e.provider.Service.Logger.Error("Failed to update media status after upload error", zap.Error(err))
					return
				}
				generatedReport.Status = core.GeneratedReportStatusFailed
				generatedReport.SystemMessage = "File upload failed: " + uploadErr.Error()
				e.core.GeneratedReportManager().UpdateByID(ctx, id, generatedReport)
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

			if mediaUpdateErr := e.core.MediaManager().UpdateByID(ctx, completed.ID, completed); mediaUpdateErr != nil {
				generatedReport.Status = core.GeneratedReportStatusFailed
				generatedReport.SystemMessage = "Failed to update media record after upload: " + mediaUpdateErr.Error()
				if err := e.core.GeneratedReportManager().UpdateByID(ctx, id, generatedReport); err != nil {
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

		if finalUpdateErr := e.core.GeneratedReportManager().UpdateByID(ctx, id, generatedReport); finalUpdateErr != nil {
			e.core.FootstepManager().Create(ctx, &core.Footstep{
				OrganizationID: &generatedReport.OrganizationID,
				BranchID:       &generatedReport.BranchID,
				UserID:         &generatedReport.CreatedByID,
				Module:         "generated_report_final_update_failed",
				Description:    "Failed to update generated report with final status after processing: " + finalUpdateErr.Error()})
		}
	}()
	return generatedReport, nil
}

func (r *Event) GenerateReport(ctx context.Context, generatedReport core.GeneratedReport) (results []byte, err error) {
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
