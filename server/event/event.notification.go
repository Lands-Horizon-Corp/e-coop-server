package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

type NotificationEvent struct {
	Title            string
	Description      string
	NotificationType string
}

// Only users with a valid CSRF token can trigger notifications
func (e *Event) Notification(ctx context.Context, echoCtx echo.Context, data NotificationEvent) {

	go func() {
		user, err := e.userToken.CurrentUser(ctx, echoCtx)
		if err != nil {
			return
		}
		data.Title = handlers.Sanitize(data.Title)
		data.Description = handlers.Sanitize(data.Description)

		if data.Description == "" || data.NotificationType == "" {
			return
		}
		notification := &modelcore.Notification{
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
			UserID:           user.ID,
			Title:            data.Title,
			Description:      data.Description,
			IsViewed:         false,
			NotificationType: data.NotificationType,
		}

		if err := e.modelcore.NotificationManager.Create(ctx, notification); err != nil {
			return
		}
	}()
}
