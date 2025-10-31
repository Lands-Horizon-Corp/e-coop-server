package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) memberDepartmentController() {
	req := c.provider.Service.Request

	// Get all member department history for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-department-history",
		Method:       "GET",
		ResponseType: modelcore.MemberDepartmentHistory{},
		Note:         "Returns all member department history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberDepartmentHistory, err := c.modelcore.MemberDepartmentHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member department history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.MemberDepartmentHistoryManager.Filtered(context, ctx, memberDepartmentHistory))
	})

	// Get member department history by member profile ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-department-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: modelcore.MemberDepartmentHistoryResponse{},
		Note:         "Returns member department history for a specific member profile ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberDepartmentHistory, err := c.modelcore.MemberDepartmentHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member department history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.MemberDepartmentHistoryManager.Pagination(context, ctx, memberDepartmentHistory))
	})

	// Get all member departments for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-department",
		Method:       "GET",
		ResponseType: modelcore.MemberDepartmentResponse{},
		Note:         "Returns all member departments for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberDepartment, err := c.modelcore.MemberDepartmentCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member departments: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.MemberDepartmentManager.Filtered(context, ctx, memberDepartment))
	})

	// Get paginated member departments
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-department/search",
		Method:       "GET",
		ResponseType: modelcore.MemberDepartmentResponse{},
		Note:         "Returns paginated member departments for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberDepartment, err := c.modelcore.MemberDepartmentCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member departments for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.MemberDepartmentManager.Pagination(context, ctx, memberDepartment))
	})

	// Create a new member department
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-department",
		Method:       "POST",
		ResponseType: modelcore.MemberDepartmentResponse{},
		RequestType:  modelcore.MemberDepartmentRequest{},
		Note:         "Creates a new member department record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelcore.MemberDepartmentManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member department failed (/member-department), validation error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member department failed (/member-department), user org error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberDepartment := &modelcore.MemberDepartment{
			Name:           req.Name,
			Description:    req.Description,
			Icon:           req.Icon,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.modelcore.MemberDepartmentManager.Create(context, memberDepartment); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member department failed (/member-department), db error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member department: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member department (/member-department): " + memberDepartment.Name,
			Module:      "MemberDepartment",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.MemberDepartmentManager.ToModel(memberDepartment))
	})

	// Update an existing member department by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-department/:member_department_id",
		Method:       "PUT",
		ResponseType: modelcore.MemberDepartmentResponse{},
		RequestType:  modelcore.MemberDepartmentRequest{},
		Note:         "Updates an existing member department record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberDepartmentID, err := handlers.EngineUUIDParam(ctx, "member_department_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), invalid member_department_id: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_department_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), user org error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.modelcore.MemberDepartmentManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), validation error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberDepartment, err := c.modelcore.MemberDepartmentManager.GetByID(context, *memberDepartmentID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), not found: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member department not found: " + err.Error()})
		}
		memberDepartment.UpdatedAt = time.Now().UTC()
		memberDepartment.UpdatedByID = user.UserID
		memberDepartment.OrganizationID = user.OrganizationID
		memberDepartment.BranchID = *user.BranchID
		memberDepartment.Name = req.Name
		memberDepartment.Description = req.Description
		memberDepartment.Icon = req.Icon
		if err := c.modelcore.MemberDepartmentManager.UpdateFields(context, memberDepartment.ID, memberDepartment); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member department failed (/member-department/:member_department_id), db error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member department: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member department (/member-department/:member_department_id): " + memberDepartment.Name,
			Module:      "MemberDepartment",
		})
		return ctx.JSON(http.StatusOK, c.modelcore.MemberDepartmentManager.ToModel(memberDepartment))
	})

	// Delete a member department by ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-department/:member_department_id",
		Method: "DELETE",
		Note:   "Deletes a member department record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberDepartmentID, err := handlers.EngineUUIDParam(ctx, "member_department_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member department failed (/member-department/:member_department_id), invalid member_department_id: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_department_id: " + err.Error()})
		}
		value, err := c.modelcore.MemberDepartmentManager.GetByID(context, *memberDepartmentID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member department failed (/member-department/:member_department_id), record not found: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member department not found: " + err.Error()})
		}
		if err := c.modelcore.MemberDepartmentManager.DeleteByID(context, *memberDepartmentID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member department failed (/member-department/:member_department_id), db error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member department: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member department (/member-department/:member_department_id): " + value.Name,
			Module:      "MemberDepartment",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete member departments by IDs
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/member-department/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member department records by their IDs.",
		RequestType: modelcore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member departments failed (/member-department/bulk-delete), invalid request body.",
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member departments failed (/member-department/bulk-delete), no IDs provided.",
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member departments failed (/member-department/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		names := ""
		for _, rawID := range reqBody.IDs {
			memberDepartmentID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member departments failed (/member-department/bulk-delete), invalid UUID: " + rawID,
					Module:      "MemberDepartment",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID '%s': %s", rawID, err.Error())})
			}

			value, err := c.modelcore.MemberDepartmentManager.GetByID(context, memberDepartmentID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member departments failed (/member-department/bulk-delete), not found: " + rawID,
					Module:      "MemberDepartment",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Member department with ID '%s' not found: %s", rawID, err.Error())})
			}

			names += value.Name + ","
			if err := c.modelcore.MemberDepartmentManager.DeleteByIDWithTx(context, tx, memberDepartmentID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member departments failed (/member-department/bulk-delete), db error: " + err.Error(),
					Module:      "MemberDepartment",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete member department with ID '%s': %s", rawID, err.Error())})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member departments failed (/member-department/bulk-delete), commit error: " + err.Error(),
				Module:      "MemberDepartment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member departments (/member-department/bulk-delete): " + names,
			Module:      "MemberDepartment",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
