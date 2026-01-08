package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type NotificationEvent struct {
	Title            string
	Description      string
	NotificationType core.NotificationType
}

func (e *Event) createNotificationForUsers(context context.Context, users []*core.UserOrganization, data NotificationEvent, senderUserID *uuid.UUID) {
	data.Title = handlers.Sanitize(data.Title)
	data.Description = handlers.Sanitize(data.Description)

	if data.Description == "" || data.NotificationType == "" {
		return
	}

	for _, org := range users {
		if senderUserID != nil && (org.UserID == *senderUserID || handlers.UUIDPtrEqual(&org.UserID, senderUserID)) {
			continue
		}

		notification := &core.Notification{
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
			Title:            data.Title,
			Description:      data.Description,
			IsViewed:         false,
			NotificationType: data.NotificationType,
			UserID:           org.UserID,   // Recipient (who receives the notification)
			RecipientID:      senderUserID, // Sender (who sent the notification)
			UserType:         org.UserType,
		}

		if err := e.core.NotificationManager().Create(context, notification); err != nil {
			continue // Continue with other notifications if one fails
		}
	}
}

func (e *Event) filterAdminUsers(users []*core.UserOrganization) []*core.UserOrganization {
	var adminUsers []*core.UserOrganization
	for _, user := range users {
		if user.UserType == core.UserOrganizationTypeEmployee || user.UserType == core.UserOrganizationTypeOwner {
			adminUsers = append(adminUsers, user)
		}
	}
	return adminUsers
}

func (e *Event) Notification(ctx echo.Context, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		user, err := e.CurrentUser(context, ctx)
		if err != nil {
			return
		}

		users := []*core.UserOrganization{{UserID: user.ID, UserType: ""}}
		e.createNotificationForUsers(context, users, data, nil)
	}()
}

func (e *Event) OrganizationAdminsNotification(ctx echo.Context, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userOrg, err := e.CurrentUserOrganization(context, ctx)
		if err != nil {
			return
		}

		userOrganizations, err := e.core.UserOrganizationManager().Find(context, &core.UserOrganization{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       userOrg.BranchID,
		})
		if err != nil {
			return
		}

		adminUsers := e.filterAdminUsers(userOrganizations)
		e.createNotificationForUsers(context, adminUsers, data, &userOrg.UserID)
	}()
}

func (e *Event) OrganizationNotification(ctx echo.Context, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userOrg, err := e.CurrentUserOrganization(context, ctx)
		if err != nil {
			return
		}

		userOrganizations, err := e.core.UserOrganizationManager().Find(context, &core.UserOrganization{
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return
		}

		e.createNotificationForUsers(context, userOrganizations, data, &userOrg.UserID)
	}()
}

func (e *Event) OrganizationDirectNotification(ctx echo.Context, organizationID uuid.UUID, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userOrganizations, err := e.core.UserOrganizationManager().Find(context, &core.UserOrganization{
			OrganizationID: organizationID,
		})
		if err != nil {
			return
		}

		e.createNotificationForUsers(context, userOrganizations, data, nil)
	}()
}

func (e *Event) OrganizationAdminsDirectNotification(ctx echo.Context, organizationID uuid.UUID, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userOrganizations, err := e.core.UserOrganizationManager().Find(context, &core.UserOrganization{
			OrganizationID: organizationID,
		})
		if err != nil {
			return
		}

		adminUsers := e.filterAdminUsers(userOrganizations)
		e.createNotificationForUsers(context, adminUsers, data, nil)
	}()
}
