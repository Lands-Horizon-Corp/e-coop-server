package event

import (
	"context"
	"fmt"
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
	fmt.Println("=== NOTIFICATION DEBUG START ===")
	fmt.Printf("DEBUG: Notification called with data: Title='%s', Description='%s', Type='%s'\n",
		data.Title, data.Description, data.NotificationType)

	go func() {
		fmt.Println("DEBUG: Inside goroutine, getting current user...")
		user, err := e.user_token.CurrentUser(ctx, echoCtx)
		if err != nil {
			fmt.Printf("ERROR: Failed to get current user: %s\n", err.Error())
			return
		}
		fmt.Printf("DEBUG: Got user - ID: %s, UserName: %s\n", user.ID.String(), user.UserName)

		// Trim and validate data
		originalTitle := data.Title
		originalDesc := data.Description
		data.Title = strings.TrimSpace(data.Title)
		data.Description = strings.TrimSpace(data.Description)

		fmt.Printf("DEBUG: Data validation - Title: '%s' -> '%s'\n", originalTitle, data.Title)
		fmt.Printf("DEBUG: Data validation - Description: '%s' -> '%s'\n", originalDesc, data.Description)
		fmt.Printf("DEBUG: Data validation - NotificationType: '%s'\n", data.NotificationType)

		if data.Description == "" || data.NotificationType == "" {
			fmt.Println("WARNING: Notification validation failed - empty description or notification type")
			fmt.Printf("  Description empty: %t, NotificationType empty: %t\n",
				data.Description == "", data.NotificationType == "")
			return
		}

		fmt.Println("DEBUG: Validation passed, creating notification...")

		notification := &model_core.Notification{
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
			UserID:           user.ID,
			Title:            data.Title,
			Description:      data.Description,
			IsViewed:         false,
			NotificationType: data.NotificationType,
		}

		fmt.Printf("DEBUG: Notification object created: %+v\n", notification)

		if err := e.model_core.NotificationManager.Create(ctx, notification); err != nil {
			fmt.Printf("ERROR: Failed to create notification in database: %s\n", err.Error())
			fmt.Println("ERROR: Full error details:")
			fmt.Printf("  Error type: %T\n", err)
			fmt.Printf("  Error string: %s\n", err.Error())
			return
		}

		fmt.Println("SUCCESS: Notification created successfully!")
		fmt.Printf("SUCCESS: Notification for user %s ('%s') - Title: '%s'\n",
			user.UserName, user.ID.String(), data.Title)
		fmt.Println("=== NOTIFICATION DEBUG END ===")
	}()
}
