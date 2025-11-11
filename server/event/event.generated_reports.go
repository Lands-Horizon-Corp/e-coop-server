package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func (e *Event) GeneratedReportDownload(ctx context.Context, echoCtx echo.Context, generatedReport *core.GeneratedReport) (*core.GeneratedReport, error) {
	if err := e.core.GeneratedReportManager.Create(ctx, generatedReport); err != nil {
		e.Footstep(echoCtx, FootstepEvent{
			Activity:    "create-error",
			Description: "Generated report creation failed (/generated-report), db error: " + err.Error(),
			Module:      "GeneratedReport",
		})
		return nil, eris.Wrapf(err, "Failed to create generated report")
	}

	go func() {
		// Create a context with 5 minute timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	return generatedReport, nil
}
