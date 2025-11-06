package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// NotificationEvent represents data required to create a notification.
type NotificationEvent struct {
	Title            string
	Description      string
	NotificationType core.NotificationType
}

// createNotificationForUsers is a helper function that creates notifications for a list of users
func (e *Event) createNotificationForUsers(context context.Context, users []*core.UserOrganization, data NotificationEvent, senderUserID *uuid.UUID) {
	data.Title = handlers.Sanitize(data.Title)
	data.Description = handlers.Sanitize(data.Description)

	if data.Description == "" || data.NotificationType == "" {
		return
	}

	for _, org := range users {
		// Skip sending notification to sender (if provided)
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
			RecipientID:      &org.UserID,
			UserID:           org.UserID,
			UserType:         org.UserType,
		}

		if senderUserID != nil {
			notification.UserID = *senderUserID
			notification.RecipientID = &org.UserID
		}

		if err := e.core.NotificationManager.Create(context, notification); err != nil {
			continue // Continue with other notifications if one fails
		}
	}
}

// filterAdminUsers filters users to only include employees and owners
func (e *Event) filterAdminUsers(users []*core.UserOrganization) []*core.UserOrganization {
	var adminUsers []*core.UserOrganization
	for _, user := range users {
		if user.UserType == core.UserOrganizationTypeEmployee || user.UserType == core.UserOrganizationTypeOwner {
			adminUsers = append(adminUsers, user)
		}
	}
	return adminUsers
}

// Notification creates a notification record asynchronously for the current user
func (e *Event) Notification(ctx echo.Context, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		user, err := e.userToken.CurrentUser(context, ctx)
		if err != nil {
			return
		}

		// Create a fake user org for the single user notification
		users := []*core.UserOrganization{{UserID: user.ID, UserType: ""}}
		e.createNotificationForUsers(context, users, data, nil)
	}()
}

// OrganizationAdminsNotification notifies admin users in the current user's organization and branch
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
			BranchID:       user.BranchID,
		})
		if err != nil {
			return
		}

		adminUsers := e.filterAdminUsers(useOrganizations)
		e.createNotificationForUsers(context, adminUsers, data, &user.UserID)
	}()
}

// OrganizationNotification notifies all users in the current user's organization
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
		if err != nil {
			return
		}

		e.createNotificationForUsers(context, useOrganizations, data, &user.UserID)
	}()
}

// OrganizationDirectNotification creates notifications for all users in an organization by ID
func (e *Event) OrganizationDirectNotification(ctx echo.Context, organizationID uuid.UUID, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		useOrganizations, err := e.core.UserOrganizationManager.Find(context, &core.UserOrganization{
			OrganizationID: organizationID,
		})
		if err != nil {
			return
		}

		e.createNotificationForUsers(context, useOrganizations, data, nil)
	}()
}

// OrganizationAdminsDirectNotification creates notifications for admin users only in an organization by ID
func (e *Event) OrganizationAdminsDirectNotification(ctx echo.Context, organizationID uuid.UUID, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		useOrganizations, err := e.core.UserOrganizationManager.Find(context, &core.UserOrganization{
			OrganizationID: organizationID,
		})
		if err != nil {
			return
		}

		adminUsers := e.filterAdminUsers(useOrganizations)
		e.createNotificationForUsers(context, adminUsers, data, nil)
	}()
}
