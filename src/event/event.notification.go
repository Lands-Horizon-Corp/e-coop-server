package event

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/src/model"
)

type NotificationEven struct {
}

func (e *Event) Notification(ctx context.Context, echoCtx echo.Context) {
	go func() {
		e.model.NotificationManager.Create(ctx, &model.Notification{
			// ID
			// CreatedAt
			// UpdatedAt
			// DeletedAt
			// UserID
			// User
			// Title
			// Description
			IsViewed: false,
			// NotificationType
		})
	}()
}
