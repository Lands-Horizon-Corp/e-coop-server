package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelcore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelcore"
	"github.com/labstack/echo/v4"
)

func (c *Controller) heartbeat(
	req := c.provider.Service.Request

	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/heartbeat/online",
		Method: "POST",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "online-error",
				Description: "User authentication failed or organization not found: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.Status == modelcore.UserOrganizationStatusOnline {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrg.Status = modelcore.UserOrganizationStatusOnline
		userOrg.LastOnlineAt = time.Now()
		if err := c.modelcore.UserOrganizationManager.Update(context, userOrg); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "online-error",
				Description: "Failed to update user organization status: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization status"})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "online-success",
			Description: "User set online status",
			Module:      "User",
		})
		if err := c.provider.Service.Broker.Dispatch(context, []string{
			fmt.Sprintf("user_organization.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_organization.status.organization.%s", userOrg.OrganizationID),
		}, nil); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to dispatch user organization status"})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/heartbeat/offline",
		Method: "POST",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "offline-error",
				Description: "User authentication failed or organization not found: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.Status == modelcore.UserOrganizationStatusOffline {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrg.Status = modelcore.UserOrganizationStatusOffline
		userOrg.LastOnlineAt = time.Now()
		if err := c.modelcore.UserOrganizationManager.Update(context, userOrg); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "offline-error",
				Description: "Failed to update user organization status: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization status"})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "offline-success",
			Description: "User set offline status",
			Module:      "User",
		})
		if err := c.provider.Service.Broker.Dispatch(context, []string{
			fmt.Sprintf("user_organization.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_organization.status.organization.%s", userOrg.OrganizationID),
		}, nil); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to dispatch user organization status"})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/heartbeat/status",
		Method:      "POST",
		RequestType: modelcore.UserOrganizationStatusRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req modelcore.UserOrganizationStatusRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.Status == req.UserOrganizationStatus {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrg.Status = req.UserOrganizationStatus
		userOrg.LastOnlineAt = time.Now()
		if err := c.modelcore.UserOrganizationManager.Update(context, userOrg); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization status"})
		}
		if err := c.provider.Service.Broker.Dispatch(context, []string{
			fmt.Sprintf("user_organization.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_organization.status.organization.%s", userOrg.OrganizationID),
		}, nil); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to dispatch user organization status"})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/heartbeat/status",
		Method:       "GET",
		ResponseType: modelcore.UserOrganizationStatusResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganizations, err := c.modelcore.UserOrganizationManager.Find(context, &modelcore.UserOrganization{
			BranchID:       userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organization status"})
		}
		statuses := c.modelcore.UserOrganizationManager.Filtered(context, ctx, userOrganizations)

		var (
			offlineUsers   []*modelcore.UserOrganizationResponse
			onlineUsers    []*modelcore.UserOrganizationResponse
			commutingUsers []*modelcore.UserOrganizationResponse
			busyUsers      []*modelcore.UserOrganizationResponse
			vacationUsers  []*modelcore.UserOrganizationResponse
		)
		onlineMembers, onlineEmployees := 0, 0
		totalMembers, totalEmployees := 0, 0
		for _, org := range statuses {
			switch org.Status {
			case modelcore.UserOrganizationStatusOnline:
				onlineUsers = append(onlineUsers, org)
				if org.UserType == modelcore.UserOrganizationTypeMember {
					onlineMembers++
					totalMembers++
				}
				if org.UserType == modelcore.UserOrganizationTypeEmployee {
					onlineEmployees++
					totalEmployees++
				}
			case modelcore.UserOrganizationStatusOffline:
				offlineUsers = append(offlineUsers, org)
			case modelcore.UserOrganizationStatusCommuting:
				commutingUsers = append(commutingUsers, org)
			case modelcore.UserOrganizationStatusBusy:
				busyUsers = append(busyUsers, org)
			case modelcore.UserOrganizationStatusVacation:
				vacationUsers = append(vacationUsers, org)
			}
		}
		timesheets, err := c.modelcore.TimeSheetActiveUsers(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve active timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, modelcore.UserOrganizationStatusResponse{
			OfflineUsers:   offlineUsers,
			OnlineUsers:    onlineUsers,
			CommutingUsers: commutingUsers,
			BusyUsers:      busyUsers,
			VacationUsers:  vacationUsers,

			OnlineUsersCount:     len(onlineUsers),
			OnlineMembers:        onlineMembers,
			TotalMembers:         totalMembers,
			OnlineEmployees:      onlineEmployees,
			TotalEmployees:       totalEmployees,
			TotalActiveEmployees: len(timesheets),
			ActiveEmployees:      c.modelcore.TimesheetManager.Filtered(context, ctx, timesheets),
		})
	})

}
