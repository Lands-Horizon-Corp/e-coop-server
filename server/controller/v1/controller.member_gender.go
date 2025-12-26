package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) memberGenderController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-gender-history",
		Method:       "GET",
		ResponseType: core.MemberGenderHistory{},
		Note:         "Returns all member gender history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGenderHistory, err := c.core.MemberGenderHistoryCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member gender history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MemberGenderHistoryManager().ToModels(memberGenderHistory))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-gender-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: core.MemberGenderHistoryResponse{},
		Note:         "Returns member gender history for a specific member profile ID.",
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
		memberGenderHistory, err := c.core.MemberGenderHistoryManager().NormalPagination(context, ctx, &core.MemberGenderHistory{
			MemberProfileID: *memberProfileID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member gender history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberGenderHistory)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-gender",
		Method:       "GET",
		ResponseType: core.MemberGenderResponse{},
		Note:         "Returns all member genders for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGender, err := c.core.MemberGenderCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member genders: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MemberGenderManager().ToModels(memberGender))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-gender/search",
		Method:       "GET",
		ResponseType: core.MemberGenderResponse{},
		Note:         "Returns paginated member genders for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGender, err := c.core.MemberGenderManager().NormalPagination(context, ctx, &core.MemberGender{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member genders for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberGender)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-gender",
		Method:       "POST",
		ResponseType: core.MemberGenderResponse{},
		RequestType:  core.MemberGenderRequest{},
		Note:         "Creates a new member gender record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.MemberGenderManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member gender failed (/member-gender), validation error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member gender failed (/member-gender), user org error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberGender := &core.MemberGender{
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := c.core.MemberGenderManager().Create(context, memberGender); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member gender failed (/member-gender), db error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member gender: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member gender (/member-gender): " + memberGender.Name,
			Module:      "MemberGender",
		})

		return ctx.JSON(http.StatusOK, c.core.MemberGenderManager().ToModel(memberGender))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-gender/:member_gender_id",
		Method:       "PUT",
		ResponseType: core.MemberGenderResponse{},
		RequestType:  core.MemberGenderRequest{},
		Note:         "Updates an existing member gender record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGenderID, err := handlers.EngineUUIDParam(ctx, "member_gender_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member gender failed (/member-gender/:member_gender_id), invalid member_gender_id: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_gender_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member gender failed (/member-gender/:member_gender_id), user org error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.core.MemberGenderManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member gender failed (/member-gender/:member_gender_id), validation error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberGender, err := c.core.MemberGenderManager().GetByID(context, *memberGenderID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member gender failed (/member-gender/:member_gender_id), not found: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member gender not found: " + err.Error()})
		}
		memberGender.UpdatedAt = time.Now().UTC()
		memberGender.UpdatedByID = userOrg.UserID
		memberGender.OrganizationID = userOrg.OrganizationID
		memberGender.BranchID = *userOrg.BranchID
		memberGender.Name = req.Name
		memberGender.Description = req.Description
		if err := c.core.MemberGenderManager().UpdateByID(context, memberGender.ID, memberGender); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member gender failed (/member-gender/:member_gender_id), db error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member gender: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member gender (/member-gender/:member_gender_id): " + memberGender.Name,
			Module:      "MemberGender",
		})
		return ctx.JSON(http.StatusOK, c.core.MemberGenderManager().ToModel(memberGender))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/member-gender/:member_gender_id",
		Method: "DELETE",
		Note:   "Deletes a member gender record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGenderID, err := handlers.EngineUUIDParam(ctx, "member_gender_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member gender failed (/member-gender/:member_gender_id), invalid member_gender_id: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_gender_id: " + err.Error()})
		}
		value, err := c.core.MemberGenderManager().GetByID(context, *memberGenderID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member gender failed (/member-gender/:member_gender_id), record not found: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member gender not found: " + err.Error()})
		}
		if err := c.core.MemberGenderManager().Delete(context, *memberGenderID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member gender failed (/member-gender/:member_gender_id), db error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member gender: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member gender (/member-gender/:member_gender_id): " + value.Name,
			Module:      "MemberGender",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/member-gender/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member gender records by their IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member genders failed (/member-gender/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member genders failed (/member-gender/bulk-delete) | no IDs provided",
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := c.core.MemberGenderManager().BulkDelete(context, ids); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member genders failed (/member-gender/bulk-delete) | error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member genders: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member genders (/member-gender/bulk-delete)",
			Module:      "MemberGender",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
