package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func memberOccupationController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-occupation-history",
		Method:       "GET",
		ResponseType: core.MemberOccupationHistoryResponse{},
		Note:         "Returns all member occupation history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberOccupationHistory, err := c.core.MemberOccupationHistoryCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupation history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MemberOccupationHistoryManager().ToModels(memberOccupationHistory))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-occupation-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: core.MemberOccupationHistoryResponse{},
		Note:         "Returns member occupation history for a specific member profile ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberOccupationHistory, err := c.core.MemberOccupationHistoryManager().NormalPagination(context, ctx, &core.MemberOccupationHistory{
			MemberProfileID: *memberProfileID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupation history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberOccupationHistory)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-occupation",
		Method:       "GET",
		ResponseType: core.MemberOccupationResponse{},
		Note:         "Returns all member occupations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberOccupation, err := c.core.MemberOccupationCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MemberOccupationManager().ToModels(memberOccupation))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-occupation/search",
		Method:       "GET",
		ResponseType: core.MemberOccupationResponse{},
		Note:         "Returns paginated member occupations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.core.MemberOccupationManager().NormalPagination(context, ctx, &core.MemberOccupation{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupations for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-occupation",
		Method:       "POST",
		ResponseType: core.MemberOccupationResponse{},
		RequestType:  core.MemberOccupationRequest{},
		Note:         "Creates a new member occupation record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.MemberOccupationManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member occupation failed (/member-occupation), validation error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member occupation failed (/member-occupation), user org error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberOccupation := &core.MemberOccupation{
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := c.core.MemberOccupationManager().Create(context, memberOccupation); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member occupation failed (/member-occupation), db error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member occupation: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member occupation (/member-occupation): " + memberOccupation.Name,
			Module:      "MemberOccupation",
		})

		return ctx.JSON(http.StatusOK, c.core.MemberOccupationManager().ToModel(memberOccupation))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-occupation/:member_occupation_id",
		Method:       "PUT",
		ResponseType: core.MemberOccupationResponse{},
		RequestType:  core.MemberOccupationRequest{},
		Note:         "Updates an existing member occupation record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberOccupationID, err := handlers.EngineUUIDParam(ctx, "member_occupation_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), invalid member_occupation_id: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_occupation_id: " + err.Error()})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), user org error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.core.MemberOccupationManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), validation error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberOccupation, err := c.core.MemberOccupationManager().GetByID(context, *memberOccupationID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), not found: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member occupation not found: " + err.Error()})
		}
		memberOccupation.UpdatedAt = time.Now().UTC()
		memberOccupation.UpdatedByID = userOrg.UserID
		memberOccupation.OrganizationID = userOrg.OrganizationID
		memberOccupation.BranchID = *userOrg.BranchID
		memberOccupation.Name = req.Name
		memberOccupation.Description = req.Description
		if err := c.core.MemberOccupationManager().UpdateByID(context, memberOccupation.ID, memberOccupation); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), db error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member occupation: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member occupation (/member-occupation/:member_occupation_id): " + memberOccupation.Name,
			Module:      "MemberOccupation",
		})
		return ctx.JSON(http.StatusOK, c.core.MemberOccupationManager().ToModel(memberOccupation))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/member-occupation/:member_occupation_id",
		Method: "DELETE",
		Note:   "Deletes a member occupation record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberOccupationID, err := handlers.EngineUUIDParam(ctx, "member_occupation_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member occupation failed (/member-occupation/:member_occupation_id), invalid member_occupation_id: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_occupation_id: " + err.Error()})
		}
		value, err := c.core.MemberOccupationManager().GetByID(context, *memberOccupationID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member occupation failed (/member-occupation/:member_occupation_id), record not found: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member occupation not found: " + err.Error()})
		}
		if err := c.core.MemberOccupationManager().Delete(context, *memberOccupationID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member occupation failed (/member-occupation/:member_occupation_id), db error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member occupation: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member occupation (/member-occupation/:member_occupation_id): " + value.Name,
			Module:      "MemberOccupation",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/member-occupation/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member occupation records by their IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member occupations failed (/member-occupation/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member occupations failed (/member-occupation/bulk-delete) | no IDs provided",
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := c.core.MemberOccupationManager().BulkDelete(context, ids); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member occupations failed (/member-occupation/bulk-delete) | error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member occupations: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member occupations (/member-occupation/bulk-delete)",
			Module:      "MemberOccupation",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
