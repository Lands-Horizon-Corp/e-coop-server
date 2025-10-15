package event

import (
	"context"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
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
		user, err := e.user_token.CurrentUser(ctx, echoCtx)
		if err != nil {
			return
		}
		data.Title = strings.TrimSpace(data.Title)
		data.Description = strings.TrimSpace(data.Description)
		if data.Description == "" || data.NotificationType == "" {
			return
		}

		if err := e.model_core.NotificationManager.Create(ctx, &model_core.Notification{
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
			UserID:           user.ID,
			Title:            data.Title,
			Description:      data.Description,
			IsViewed:         false,
			NotificationType: data.NotificationType,
		}); err != nil {
			return
		}
	}()
}
