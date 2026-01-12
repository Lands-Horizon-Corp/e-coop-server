package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type NotificationEvent struct {
	Title            string
	Description      string
	NotificationType core.NotificationType
}

func createNotificationForUsers(
	context context.Context, service *horizon.HorizonService,
	users []*core.UserOrganization, data NotificationEvent, senderUserID *uuid.UUID) {
	data.Title = helpers.Sanitize(data.Title)
	data.Description = helpers.Sanitize(data.Description)

	if data.Description == "" || data.NotificationType == "" {
		return
	}

	for _, org := range users {
		if senderUserID != nil && (org.UserID == *senderUserID || helpers.UUIDPtrEqual(&org.UserID, senderUserID)) {
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

		if err := core.NotificationManager(service).Create(context, notification); err != nil {
			continue
		}
	}
}

func filterAdminUsers(users []*core.UserOrganization) []*core.UserOrganization {
	var adminUsers []*core.UserOrganization
	for _, user := range users {
		if user.UserType == core.UserOrganizationTypeEmployee || user.UserType == core.UserOrganizationTypeOwner {
			adminUsers = append(adminUsers, user)
		}
	}
	return adminUsers
}

func Notification(ctx echo.Context, service *horizon.HorizonService, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		user, err := CurrentUser(context, service, ctx)
		if err != nil {
			return
		}

		users := []*core.UserOrganization{{UserID: user.ID, UserType: ""}}
		createNotificationForUsers(context, service, users, data, nil)
	}()
}

func OrganizationAdminsNotification(ctx echo.Context, service *horizon.HorizonService, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userOrg, err := CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return
		}

		userOrganizations, err := core.UserOrganizationManager(service).Find(context, &core.UserOrganization{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       userOrg.BranchID,
		})
		if err != nil {
			return
		}

		adminUsers := filterAdminUsers(userOrganizations)
		createNotificationForUsers(context, service, adminUsers, data, &userOrg.UserID)
	}()
}

func OrganizationNotification(ctx echo.Context, service *horizon.HorizonService, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userOrg, err := CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return
		}

		userOrganizations, err := core.UserOrganizationManager(service).Find(context, &core.UserOrganization{
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return
		}

		createNotificationForUsers(context, service, userOrganizations, data, &userOrg.UserID)
	}()
}

func OrganizationDirectNotification(ctx echo.Context, service *horizon.HorizonService, organizationID uuid.UUID, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userOrganizations, err := core.UserOrganizationManager(service).Find(context, &core.UserOrganization{
			OrganizationID: organizationID,
		})
		if err != nil {
			return
		}

		createNotificationForUsers(context, service, userOrganizations, data, nil)
	}()
}

func OrganizationAdminsDirectNotification(ctx echo.Context, service *horizon.HorizonService, organizationID uuid.UUID, data NotificationEvent) {
	go func() {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userOrganizations, err := core.UserOrganizationManager(service).Find(context, &core.UserOrganization{
			OrganizationID: organizationID,
		})
		if err != nil {
			return
		}

		adminUsers := filterAdminUsers(userOrganizations)
		createNotificationForUsers(context, service, adminUsers, data, nil)
	}()
}
