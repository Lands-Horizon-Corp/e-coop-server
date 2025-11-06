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
	NotificationType core.NotificationType
}

// Notification creates a notification record asynchronously for the
// current user based on the supplied data.
func (e *Event) Notification(ctx echo.Context, data NotificationEvent) {

	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		user, err := e.userToken.CurrentUser(context, ctx)
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

func (e *Event) OrganizationAdminsNotification(ctx echo.Context, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		user, err := e.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return
		}
		useOrganizations, err := e.core.UserOrganizationManager.Find(context, &core.UserOrganization{
			OrganizationID: user.OrganizationID,
		})
		data.Title = handlers.Sanitize(data.Title)
		data.Description = handlers.Sanitize(data.Description)

		if data.Description == "" || data.NotificationType == "" {
			return
		}
		for _, orgs := range useOrganizations {
			if orgs.UserType != core.UserOrganizationTypeEmployee && orgs.UserType != core.UserOrganizationTypeOwner {
				continue
			}
			notification := &core.Notification{
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),

				Title:            data.Title,
				Description:      data.Description,
				IsViewed:         false,
				NotificationType: data.NotificationType,
				RecipientID:      &orgs.UserID,
				UserID:           user.UserID,
			}
			if err := e.core.NotificationManager.Create(context, notification); err != nil {
				return
			}
		}
	}()
}

func (e *Event) OrganizationNotification(ctx echo.Context, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		user, err := e.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return
		}
		useOrganizations, err := e.core.UserOrganizationManager.Find(context, &core.UserOrganization{
			OrganizationID: user.OrganizationID,
		})
		data.Title = handlers.Sanitize(data.Title)
		data.Description = handlers.Sanitize(data.Description)

		if data.Description == "" || data.NotificationType == "" {
			return
		}
		for _, orgs := range useOrganizations {
			notification := &core.Notification{
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),

				Title:            data.Title,
				Description:      data.Description,
				IsViewed:         false,
				NotificationType: data.NotificationType,
				RecipientID:      &orgs.UserID,
				UserID:           user.UserID,
			}
			if err := e.core.NotificationManager.Create(context, notification); err != nil {
				return
			}
		}
	}()
}
