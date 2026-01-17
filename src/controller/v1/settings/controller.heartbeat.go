package settings

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func Heartbeat(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/heartbeat/online",
		Method: "POST",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "online-error",
				Description: "User authentication failed or organization not found: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.Status == core.UserOrganizationStatusOnline {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrg.Status = core.UserOrganizationStatusOnline
		userOrg.LastOnlineAt = time.Now()
		if err := core.UserOrganizationManager(service).UpdateByID(context, userOrg.ID, userOrg); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "online-error",
				Description: "Failed to update user organization status: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization status"})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "online-success",
			Description: "User set online status",
			Module:      "User",
		})
		if err := service.Broker.Dispatch([]string{
			fmt.Sprintf("user_organization.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_organization.status.organization.%s", userOrg.OrganizationID),
		}, nil); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to dispatch user organization status"})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/heartbeat/offline",
		Method: "POST",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "offline-error",
				Description: "User authentication failed or organization not found: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.Status == core.UserOrganizationStatusOffline {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrg.Status = core.UserOrganizationStatusOffline
		userOrg.LastOnlineAt = time.Now()
		if err := core.UserOrganizationManager(service).UpdateByID(context, userOrg.ID, userOrg); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "offline-error",
				Description: "Failed to update user organization status: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization status"})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "offline-success",
			Description: "User set offline status",
			Module:      "User",
		})
		if err := service.Broker.Dispatch([]string{
			fmt.Sprintf("user_organization.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_organization.status.organization.%s", userOrg.OrganizationID),
		}, nil); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to dispatch user organization status"})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/heartbeat/status",
		Method:      "POST",
		RequestType: types.UserOrganizationStatusRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req types.UserOrganizationStatusRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.Status == req.UserOrganizationStatus {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrg.Status = req.UserOrganizationStatus
		userOrg.LastOnlineAt = time.Now()
		if err := core.UserOrganizationManager(service).UpdateByID(context, userOrg.ID, userOrg); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization status"})
		}
		if err := service.Broker.Dispatch([]string{
			fmt.Sprintf("user_organization.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_organization.status.organization.%s", userOrg.OrganizationID),
		}, nil); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to dispatch user organization status"})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/heartbeat/status",
		Method:       "GET",
		ResponseType: types.UserOrganizationStatusResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganizations, err := core.UserOrganizationManager(service).Find(context, &types.UserOrganization{
			BranchID:       userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organization status"})
		}
		statuses := core.UserOrganizationManager(service).ToModels(userOrganizations)

		var (
			offlineUsers   []*types.UserOrganizationResponse
			onlineUsers    []*types.UserOrganizationResponse
			commutingUsers []*types.UserOrganizationResponse
			busyUsers      []*types.UserOrganizationResponse
			vacationUsers  []*types.UserOrganizationResponse
		)
		onlineMembers, onlineEmployees := 0, 0
		totalMembers, totalEmployees := 0, 0
		for _, org := range statuses {
			switch org.Status {
			case core.UserOrganizationStatusOnline:
				onlineUsers = append(onlineUsers, org)
				if org.UserType == types.UserOrganizationTypeMember {
					onlineMembers++
					totalMembers++
				}
				if org.UserType == types.UserOrganizationTypeEmployee {
					onlineEmployees++
					totalEmployees++
				}
			case core.UserOrganizationStatusOffline:
				offlineUsers = append(offlineUsers, org)
			case core.UserOrganizationStatusCommuting:
				commutingUsers = append(commutingUsers, org)
			case core.UserOrganizationStatusBusy:
				busyUsers = append(busyUsers, org)
			case core.UserOrganizationStatusVacation:
				vacationUsers = append(vacationUsers, org)
			}
		}
		timesheets, err := core.TimeSheetActiveUsers(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve active timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, types.UserOrganizationStatusResponse{
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
			ActiveEmployees:      core.TimesheetManager(service).ToModels(timesheets),
		})
	})

}
