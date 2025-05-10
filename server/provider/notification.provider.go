package provider

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

func (p *Provider) Notification(c echo.Context, title, description string) {
	go func() {
		claim, err := p.authentication.GetUserFromToken(c)
		if err != nil {
			return
		}
		id, err := uuid.Parse(claim.ID)
		if err != nil {
			return
		}
		user, err := p.repository.UserGetByID(id)
		if err != nil {
			return
		}
		_ = p.repository.NotificationCreate(&model.Notification{
			Title:       title,
			Description: description,
			IsViewed:    false,
			UserID:      user.ID,
		})
	}()
}
