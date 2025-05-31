package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/src/model"
)

type NotificationEvent struct {
	Title            string
	Description      string
	NotificationType string
}

func (e *Event) Notification(ctx context.Context, echoCtx echo.Context, data NotificationEvent) {
	go func() {
		user, err := e.userToken.CSRF.GetCSRF(ctx, echoCtx)
		if err != nil {
			return
		}
		userId, err := uuid.Parse(user.UserID)
		if err != nil {
			return
		}
		if err := e.model.NotificationManager.Create(ctx, &model.Notification{
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
			UserID:           userId,
			Title:            data.Description,
			Description:      data.Description,
			IsViewed:         false,
			NotificationType: data.NotificationType,
		}); err != nil {
			return
		}
	}()
}
