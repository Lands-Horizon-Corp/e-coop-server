package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) Heartbeat() {
	req := c.provider.Service.Request

	req.RegisterRoute(handlers.Route{
		Route:  "/heartbeat/online",
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
		userOrg.Status = model.UserOrganizationStatusOnline
		userOrg.LastOnlineAt = time.Now()
		if err := c.model.UserOrganizationManager.Update(context, userOrg); err != nil {
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
		c.provider.Service.Broker.Dispatch(context, []string{
			fmt.Sprintf("user_organization.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_organization.status.organization.%s", userOrg.OrganizationID),
		}, nil)
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:  "/heartbeat/offline",
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
		userOrg.Status = model.UserOrganizationStatusOffline
		userOrg.LastOnlineAt = time.Now()
		if err := c.model.UserOrganizationManager.Update(context, userOrg); err != nil {
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
		c.provider.Service.Broker.Dispatch(context, []string{
			fmt.Sprintf("user_organization.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_organization.status.organization.%s", userOrg.OrganizationID),
		}, nil)
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:       "/heartbeat/status",
		Method:      "POST",
		RequestType: model.UserOrganizationStatusRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		var req model.UserOrganizationStatusRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg.Status = req.UserOrganizationStatus
		userOrg.LastOnlineAt = time.Now()
		if err := c.model.UserOrganizationManager.Update(context, userOrg); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization status"})
		}
		c.provider.Service.Broker.Dispatch(context, []string{
			fmt.Sprintf("user_organization.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_organization.status.organization.%s", userOrg.OrganizationID),
		}, nil)
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/heartbeat/status",
		Method:       "GET",
		ResponseType: model.UserOrganizationStatusResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganizations, err := c.model.UserOrganizationManager.Find(context, &model.UserOrganization{
			BranchID:       userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organization status"})
		}
		statuses := c.model.UserOrganizationManager.Filtered(context, ctx, userOrganizations)

		var (
			offlineUsers   []*model.UserOrganizationResponse
			onlineUsers    []*model.UserOrganizationResponse
			commutingUsers []*model.UserOrganizationResponse
			busyUsers      []*model.UserOrganizationResponse
			vacationUsers  []*model.UserOrganizationResponse
		)
		onlineMembers, onlineEmployees := 0, 0
		totalMembers, totalEmployees := 0, 0
		for _, org := range statuses {
			switch org.Status {
			case model.UserOrganizationStatusOnline:
				onlineUsers = append(onlineUsers, org)
				if org.UserType == "member" {
					onlineMembers++
					totalMembers++
				}
				if org.UserType == "employee" {
					onlineEmployees++
					totalEmployees++
				}
			case model.UserOrganizationStatusOffline:
				offlineUsers = append(offlineUsers, org)
			case model.UserOrganizationStatusCommuting:
				commutingUsers = append(commutingUsers, org)
			case model.UserOrganizationStatusBusy:
				busyUsers = append(busyUsers, org)
			case model.UserOrganizationStatusVacation:
				vacationUsers = append(vacationUsers, org)
			}
		}
		timesheets, err := c.model.TimeSheetActiveUsers(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve active timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, model.UserOrganizationStatusResponse{
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
			ActiveEmployees:      c.model.TimesheetManager.Filtered(context, ctx, timesheets),
		})
	})

}
