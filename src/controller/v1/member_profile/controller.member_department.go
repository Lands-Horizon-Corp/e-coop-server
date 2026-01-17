package member_profile

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func MemberDepartmentController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-department-history",
		Method:       "GET",
		ResponseType: types.MemberDepartmentHistory{},
		Note:         "Returns all member department history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberDepartmentHistory, err := core.MemberDepartmentHistoryCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member department history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberDepartmentHistoryManager(service).ToModels(memberDepartmentHistory))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-department-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: types.MemberDepartmentHistoryResponse{},
		Note:         "Returns member department history for a specific member profile ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberDepartmentHistory, err := core.MemberDepartmentHistoryManager(service).NormalPagination(context, ctx, &types.MemberDepartmentHistory{
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			MemberProfileID: *memberProfileID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member department history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberDepartmentHistory)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-department",
		Method:       "GET",
		ResponseType: types.MemberDepartmentResponse{},
		Note:         "Returns all member departments for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberDepartment, err := core.MemberDepartmentCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member departments: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberDepartmentManager(service).ToModels(memberDepartment))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-department/search",
		Method:       "GET",
		ResponseType: types.MemberDepartmentResponse{},
		Note:         "Returns paginated member departments for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberDepartment, err := core.MemberDepartmentManager(service).NormalPagination(context, ctx, &types.MemberDepartment{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member departments for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberDepartment)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-department",
		Method:       "POST",
		ResponseType: types.MemberDepartmentResponse{},
		RequestType:  types.MemberDepartmentRequest{},
		Note:         "Creates a new member department record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.MemberDepartmentManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member department failed (/member-department), validation error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member department failed (/member-department), user org error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberDepartment := &types.MemberDepartment{
			Name:           req.Name,
			Description:    req.Description,
			Icon:           req.Icon,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.MemberDepartmentManager(service).Create(context, memberDepartment); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member department failed (/member-department), db error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member department: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member department (/member-department): " + memberDepartment.Name,
			Module:      "MemberDepartment",
		})

		return ctx.JSON(http.StatusOK, core.MemberDepartmentManager(service).ToModel(memberDepartment))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-department/:member_department_id",
		Method:       "PUT",
		ResponseType: types.MemberDepartmentResponse{},
		RequestType:  types.MemberDepartmentRequest{},
		Note:         "Updates an existing member department record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberDepartmentID, err := helpers.EngineUUIDParam(ctx, "member_department_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), invalid member_department_id: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_department_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), user org error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := core.MemberDepartmentManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), validation error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberDepartment, err := core.MemberDepartmentManager(service).GetByID(context, *memberDepartmentID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), not found: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member department not found: " + err.Error()})
		}
		memberDepartment.UpdatedAt = time.Now().UTC()
		memberDepartment.UpdatedByID = userOrg.UserID
		memberDepartment.OrganizationID = userOrg.OrganizationID
		memberDepartment.BranchID = *userOrg.BranchID
		memberDepartment.Name = req.Name
		memberDepartment.Description = req.Description
		memberDepartment.Icon = req.Icon
		if err := core.MemberDepartmentManager(service).UpdateByID(context, memberDepartment.ID, memberDepartment); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), db error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member department: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member department (/member-department/:member_department_id): " + memberDepartment.Name,
			Module:      "MemberDepartment",
		})
		return ctx.JSON(http.StatusOK, core.MemberDepartmentManager(service).ToModel(memberDepartment))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-department/:member_department_id",
		Method: "DELETE",
		Note:   "Deletes a member department record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberDepartmentID, err := helpers.EngineUUIDParam(ctx, "member_department_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member department failed (/member-department/:member_department_id), invalid member_department_id: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_department_id: " + err.Error()})
		}
		value, err := core.MemberDepartmentManager(service).GetByID(context, *memberDepartmentID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member department failed (/member-department/:member_department_id), record not found: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member department not found: " + err.Error()})
		}
		if err := core.MemberDepartmentManager(service).Delete(context, *memberDepartmentID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member department failed (/member-department/:member_department_id), db error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member department: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member department (/member-department/:member_department_id): " + value.Name,
			Module:      "MemberDepartment",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/member-department/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member department records by their IDs.",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member departments failed (/member-department/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member departments failed (/member-department/bulk-delete) | no IDs provided",
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.MemberDepartmentManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member departments failed (/member-department/bulk-delete) | error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member departments: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member departments (/member-department/bulk-delete)",
			Module:      "MemberDepartment",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
