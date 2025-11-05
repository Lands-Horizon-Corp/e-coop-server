package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// NotificationEvent represents data required to create a notification.
type NotificationEvent struct {
	Title            string
	Description      string
	NotificationType string
}

// Notification creates a notification record asynchronously for the
// current user based on the supplied data.
func (e *Event) Notification(echoCtx echo.Context, data NotificationEvent) {

	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		user, err := e.userToken.CurrentUser(context, echoCtx)
		if err != nil {
			return
		}
		data.Title = handlers.Sanitize(data.Title)
		data.Description = handlers.Sanitize(data.Description)

		if data.Description == "" || data.NotificationType == "" {
			return
		}
		notification := &core.Notification{
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
			UserID:           user.ID,
			Title:            data.Title,
			Description:      data.Description,
			IsViewed:         false,
			NotificationType: data.NotificationType,
		}

		if err := e.core.NotificationManager.Create(context, notification); err != nil {
			return
		}
	}()
}
