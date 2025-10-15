package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
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
		if userOrg.Status == model_core.UserOrganizationStatusOnline {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrg.Status = model_core.UserOrganizationStatusOnline
		userOrg.LastOnlineAt = time.Now()
		if err := c.model_core.UserOrganizationManager.Update(context, userOrg); err != nil {
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
		if userOrg.Status == model_core.UserOrganizationStatusOffline {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrg.Status = model_core.UserOrganizationStatusOffline
		userOrg.LastOnlineAt = time.Now()
		if err := c.model_core.UserOrganizationManager.Update(context, userOrg); err != nil {
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
		RequestType: model_core.UserOrganizationStatusRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req model_core.UserOrganizationStatusRequest
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
		if err := c.model_core.UserOrganizationManager.Update(context, userOrg); err != nil {
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
		ResponseType: model_core.UserOrganizationStatusResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganizations, err := c.model_core.UserOrganizationManager.Find(context, &model_core.UserOrganization{
			BranchID:       userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organization status"})
		}
		statuses := c.model_core.UserOrganizationManager.Filtered(context, ctx, userOrganizations)

		var (
			offlineUsers   []*model_core.UserOrganizationResponse
			onlineUsers    []*model_core.UserOrganizationResponse
			commutingUsers []*model_core.UserOrganizationResponse
			busyUsers      []*model_core.UserOrganizationResponse
			vacationUsers  []*model_core.UserOrganizationResponse
		)
		onlineMembers, onlineEmployees := 0, 0
		totalMembers, totalEmployees := 0, 0
		for _, org := range statuses {
			switch org.Status {
			case model_core.UserOrganizationStatusOnline:
				onlineUsers = append(onlineUsers, org)
				if org.UserType == model_core.UserOrganizationTypeMember {
					onlineMembers++
					totalMembers++
				}
				if org.UserType == model_core.UserOrganizationTypeEmployee {
					onlineEmployees++
					totalEmployees++
				}
			case model_core.UserOrganizationStatusOffline:
				offlineUsers = append(offlineUsers, org)
			case model_core.UserOrganizationStatusCommuting:
				commutingUsers = append(commutingUsers, org)
			case model_core.UserOrganizationStatusBusy:
				busyUsers = append(busyUsers, org)
			case model_core.UserOrganizationStatusVacation:
				vacationUsers = append(vacationUsers, org)
			}
		}
		timesheets, err := c.model_core.TimeSheetActiveUsers(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve active timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, model_core.UserOrganizationStatusResponse{
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
			ActiveEmployees:      c.model_core.TimesheetManager.Filtered(context, ctx, timesheets),
		})
	})

}
