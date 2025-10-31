package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelCore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/labstack/echo/v4"
)

func (c *Controller) Heartbeat() {
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
		if userOrg.Status == modelCore.UserOrganizationStatusOnline {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrg.Status = modelCore.UserOrganizationStatusOnline
		userOrg.LastOnlineAt = time.Now()
		if err := c.modelCore.UserOrganizationManager.Update(context, userOrg); err != nil {
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
		if userOrg.Status == modelCore.UserOrganizationStatusOffline {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrg.Status = modelCore.UserOrganizationStatusOffline
		userOrg.LastOnlineAt = time.Now()
		if err := c.modelCore.UserOrganizationManager.Update(context, userOrg); err != nil {
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
		RequestType: modelCore.UserOrganizationStatusRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req modelCore.UserOrganizationStatusRequest
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
		if err := c.modelCore.UserOrganizationManager.Update(context, userOrg); err != nil {
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
		ResponseType: modelCore.UserOrganizationStatusResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganizations, err := c.modelCore.UserOrganizationManager.Find(context, &modelCore.UserOrganization{
			BranchID:       userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organization status"})
		}
		statuses := c.modelCore.UserOrganizationManager.Filtered(context, ctx, userOrganizations)

		var (
			offlineUsers   []*modelCore.UserOrganizationResponse
			onlineUsers    []*modelCore.UserOrganizationResponse
			commutingUsers []*modelCore.UserOrganizationResponse
			busyUsers      []*modelCore.UserOrganizationResponse
			vacationUsers  []*modelCore.UserOrganizationResponse
		)
		onlineMembers, onlineEmployees := 0, 0
		totalMembers, totalEmployees := 0, 0
		for _, org := range statuses {
			switch org.Status {
			case modelCore.UserOrganizationStatusOnline:
				onlineUsers = append(onlineUsers, org)
				if org.UserType == modelCore.UserOrganizationTypeMember {
					onlineMembers++
					totalMembers++
				}
				if org.UserType == modelCore.UserOrganizationTypeEmployee {
					onlineEmployees++
					totalEmployees++
				}
			case modelCore.UserOrganizationStatusOffline:
				offlineUsers = append(offlineUsers, org)
			case modelCore.UserOrganizationStatusCommuting:
				commutingUsers = append(commutingUsers, org)
			case modelCore.UserOrganizationStatusBusy:
				busyUsers = append(busyUsers, org)
			case modelCore.UserOrganizationStatusVacation:
				vacationUsers = append(vacationUsers, org)
			}
		}
		timesheets, err := c.modelCore.TimeSheetActiveUsers(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve active timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, modelCore.UserOrganizationStatusResponse{
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
			ActiveEmployees:      c.modelCore.TimesheetManager.Filtered(context, ctx, timesheets),
		})
	})

}
