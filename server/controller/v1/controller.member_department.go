package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) memberDepartmentController() {
	req := c.provider.Service.Request

	// Get all member department history for the current branch
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-department-history",
		Method:       "GET",
		ResponseType: core.MemberDepartmentHistory{},
		Note:         "Returns all member department history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberDepartmentHistory, err := c.core.MemberDepartmentHistoryCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member department history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MemberDepartmentHistoryManager.ToModels(memberDepartmentHistory))
	})

	// Get member department history by member profile ID
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-department-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: core.MemberDepartmentHistoryResponse{},
		Note:         "Returns member department history for a specific member profile ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberDepartmentHistory, err := c.core.MemberDepartmentHistoryManager.PaginationWithFields(context, ctx, &core.MemberDepartmentHistory{
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			MemberProfileID: *memberProfileID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member department history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberDepartmentHistory)
	})

	// Get all member departments for the current branch
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-department",
		Method:       "GET",
		ResponseType: core.MemberDepartmentResponse{},
		Note:         "Returns all member departments for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberDepartment, err := c.core.MemberDepartmentCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member departments: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MemberDepartmentManager.ToModels(memberDepartment))
	})

	// Get paginated member departments
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-department/search",
		Method:       "GET",
		ResponseType: core.MemberDepartmentResponse{},
		Note:         "Returns paginated member departments for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberDepartment, err := c.core.MemberDepartmentManager.PaginationWithFields(context, ctx, &core.MemberDepartment{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member departments for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberDepartment)
	})

	// Create a new member department
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-department",
		Method:       "POST",
		ResponseType: core.MemberDepartmentResponse{},
		RequestType:  core.MemberDepartmentRequest{},
		Note:         "Creates a new member department record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.MemberDepartmentManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member department failed (/member-department), validation error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member department failed (/member-department), user org error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberDepartment := &core.MemberDepartment{
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

		if err := c.core.MemberDepartmentManager.Create(context, memberDepartment); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member department failed (/member-department), db error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member department: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member department (/member-department): " + memberDepartment.Name,
			Module:      "MemberDepartment",
		})

		return ctx.JSON(http.StatusOK, c.core.MemberDepartmentManager.ToModel(memberDepartment))
	})

	// Update an existing member department by ID
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-department/:member_department_id",
		Method:       "PUT",
		ResponseType: core.MemberDepartmentResponse{},
		RequestType:  core.MemberDepartmentRequest{},
		Note:         "Updates an existing member department record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberDepartmentID, err := handlers.EngineUUIDParam(ctx, "member_department_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), invalid member_department_id: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_department_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), user org error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.core.MemberDepartmentManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), validation error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberDepartment, err := c.core.MemberDepartmentManager.GetByID(context, *memberDepartmentID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		if err := c.core.MemberDepartmentManager.UpdateByID(context, memberDepartment.ID, memberDepartment); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), db error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member department: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member department (/member-department/:member_department_id): " + memberDepartment.Name,
			Module:      "MemberDepartment",
		})
		return ctx.JSON(http.StatusOK, c.core.MemberDepartmentManager.ToModel(memberDepartment))
	})

	// Delete a member department by ID
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/member-department/:member_department_id",
		Method: "DELETE",
		Note:   "Deletes a member department record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberDepartmentID, err := handlers.EngineUUIDParam(ctx, "member_department_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member department failed (/member-department/:member_department_id), invalid member_department_id: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_department_id: " + err.Error()})
		}
		value, err := c.core.MemberDepartmentManager.GetByID(context, *memberDepartmentID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member department failed (/member-department/:member_department_id), record not found: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member department not found: " + err.Error()})
		}
		if err := c.core.MemberDepartmentManager.Delete(context, *memberDepartmentID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member department failed (/member-department/:member_department_id), db error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member department: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member department (/member-department/:member_department_id): " + value.Name,
			Module:      "MemberDepartment",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Simplified bulk-delete handler for member departments (mirrors the feedback/holiday pattern)
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/member-department/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member department records by their IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member departments failed (/member-department/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member departments failed (/member-department/bulk-delete) | no IDs provided",
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		// Delegate deletion to the manager. Manager should handle transactions, validations and DeletedBy bookkeeping.
		if err := c.core.MemberDepartmentManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member departments failed (/member-department/bulk-delete) | error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member departments: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member departments (/member-department/bulk-delete)",
			Module:      "MemberDepartment",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
