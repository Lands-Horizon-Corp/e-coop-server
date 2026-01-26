package member_profile

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func MemberOccupationController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-occupation-history",
		Method:       "GET",
		ResponseType: types.MemberOccupationHistoryResponse{},
		Note:         "Returns all member occupation history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberOccupationHistory, err := core.MemberOccupationHistoryCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupation history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberOccupationHistoryManager(service).ToModels(memberOccupationHistory))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-occupation-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: types.MemberOccupationHistoryResponse{},
		Note:         "Returns member occupation history for a specific member profile ID.",
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
		memberOccupationHistory, err := core.MemberOccupationHistoryManager(service).NormalPagination(context, ctx, &types.MemberOccupationHistory{
			MemberProfileID: *memberProfileID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupation history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberOccupationHistory)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-occupation",
		Method:       "GET",
		ResponseType: types.MemberOccupationResponse{},
		Note:         "Returns all member occupations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberOccupation, err := core.MemberOccupationCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberOccupationManager(service).ToModels(memberOccupation))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-occupation/search",
		Method:       "GET",
		ResponseType: types.MemberOccupationResponse{},
		Note:         "Returns paginated member occupations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := core.MemberOccupationManager(service).NormalPagination(context, ctx, &types.MemberOccupation{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupations for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-occupation",
		Method:       "POST",
		ResponseType: types.MemberOccupationResponse{},
		RequestType:  types.MemberOccupationRequest{},
		Note:         "Creates a new member occupation record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.MemberOccupationManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member occupation failed (/member-occupation), validation error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member occupation failed (/member-occupation), user org error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberOccupation := &types.MemberOccupation{
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.MemberOccupationManager(service).Create(context, memberOccupation); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member occupation failed (/member-occupation), db error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member occupation: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member occupation (/member-occupation): " + memberOccupation.Name,
			Module:      "MemberOccupation",
		})

		return ctx.JSON(http.StatusOK, core.MemberOccupationManager(service).ToModel(memberOccupation))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-occupation/:member_occupation_id",
		Method:       "PUT",
		ResponseType: types.MemberOccupationResponse{},
		RequestType:  types.MemberOccupationRequest{},
		Note:         "Updates an existing member occupation record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberOccupationID, err := helpers.EngineUUIDParam(ctx, "member_occupation_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), invalid member_occupation_id: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_occupation_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), user org error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := core.MemberOccupationManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), validation error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberOccupation, err := core.MemberOccupationManager(service).GetByID(context, *memberOccupationID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.MemberOccupationManager(service).UpdateByID(context, memberOccupation.ID, memberOccupation); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), db error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member occupation: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member occupation (/member-occupation/:member_occupation_id): " + memberOccupation.Name,
			Module:      "MemberOccupation",
		})
		return ctx.JSON(http.StatusOK, core.MemberOccupationManager(service).ToModel(memberOccupation))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-occupation/:member_occupation_id",
		Method: "DELETE",
		Note:   "Deletes a member occupation record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberOccupationID, err := helpers.EngineUUIDParam(ctx, "member_occupation_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member occupation failed (/member-occupation/:member_occupation_id), invalid member_occupation_id: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_occupation_id: " + err.Error()})
		}
		value, err := core.MemberOccupationManager(service).GetByID(context, *memberOccupationID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member occupation failed (/member-occupation/:member_occupation_id), record not found: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member occupation not found: " + err.Error()})
		}
		if err := core.MemberOccupationManager(service).Delete(context, *memberOccupationID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member occupation failed (/member-occupation/:member_occupation_id), db error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member occupation: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member occupation (/member-occupation/:member_occupation_id): " + value.Name,
			Module:      "MemberOccupation",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/member-occupation/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member occupation records by their IDs.",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member occupations failed (/member-occupation/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.MemberOccupationManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member occupations failed (/member-occupation/bulk-delete) | error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member occupations: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member occupations (/member-occupation/bulk-delete)",
			Module:      "MemberOccupation",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
