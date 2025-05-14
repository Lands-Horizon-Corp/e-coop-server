package providers

import (
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

func (p *Providers) Notification(c echo.Context, title, description string, notificationType string) {
	go func() {
		user, err := p.CurrentUser(c)
		if err != nil {
			return
		}
		p.notification.Manager.Create(&model.Notification{
			Title:       title,
			Description: description,
			IsViewed:    false,
			UserID:      user.ID,
			UpdatedAt:   time.Now().UTC(),
			CreatedAt:   time.Now().UTC(),
		})
	}()
}
