package member_profile

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func MemberTypeController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-type-history",
		Method:       "GET",
		ResponseType: core.MemberTypeHistoryResponse{},
		Note:         "Returns all member type history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberTypeHistory, err := core.MemberTypeHistoryCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member type history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberTypeHistoryManager(service).ToModels(memberTypeHistory))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-type-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: core.MemberTypeHistoryResponse{},
		Note:         "Returns member type history for a specific member profile ID.",
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
		memberTypeHistory, err := core.MemberTypeHistoryManager(service).NormalPagination(context, ctx, &core.MemberTypeHistory{
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			MemberProfileID: *memberProfileID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member type history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberTypeHistory)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-type",
		Method:       "GET",
		ResponseType: core.MemberTypeResponse{},
		Note:         "Returns all member types for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberType, err := core.MemberTypeCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member types: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberTypeManager(service).ToModels(memberType))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-type/search",
		Method:       "GET",
		ResponseType: core.MemberTypeResponse{},
		Note:         "Returns paginated member types for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := core.MemberTypeManager(service).NormalPagination(context, ctx, &core.MemberType{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member types for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-type",
		Method:       "POST",
		RequestType:  core.MemberTypeRequest{},
		ResponseType: core.MemberTypeResponse{},
		Note:         "Creates a new member type record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.MemberTypeManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member type failed: validation error: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member type failed: user org error: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberType := &core.MemberType{
			Name:           req.Name,
			Description:    req.Description,
			Prefix:         req.Prefix,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.MemberTypeManager(service).Create(context, memberType); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member type failed: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member type: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member type: " + memberType.Name,
			Module:      "MemberType",
		})

		return ctx.JSON(http.StatusOK, core.MemberTypeManager(service).ToModel(memberType))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-type/:member_type_id",
		Method:       "PUT",
		RequestType:  core.MemberTypeRequest{},
		ResponseType: core.MemberTypeResponse{},
		Note:         "Updates an existing member type record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberTypeID, err := helpers.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type failed: invalid member_type_id: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_type_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type failed: user org error: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		req, err := core.MemberTypeManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type failed: validation error: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		memberType, err := core.MemberTypeManager(service).GetByID(context, *memberTypeID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: fmt.Sprintf("Update member type failed: not found (ID: %s): %v", memberTypeID, err),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberType with ID %s not found: %v", memberTypeID, err)})
		}

		memberType.UpdatedAt = time.Now().UTC()
		memberType.UpdatedByID = userOrg.UserID
		memberType.OrganizationID = userOrg.OrganizationID
		memberType.BranchID = *userOrg.BranchID
		memberType.Name = req.Name
		memberType.Description = req.Description
		memberType.Prefix = req.Prefix
		if err := core.MemberTypeManager(service).UpdateByID(context, memberType.ID, memberType); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type failed: update error: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member type: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member type: " + memberType.Name,
			Module:      "MemberType",
		})
		return ctx.JSON(http.StatusOK, core.MemberTypeManager(service).ToModel(memberType))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-type/:member_type_id",
		Method: "DELETE",
		Note:   "Deletes a member type record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberTypeID, err := helpers.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member type failed: invalid member_type_id: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_type_id: " + err.Error()})
		}
		memberType, err := core.MemberTypeManager(service).GetByID(context, *memberTypeID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: fmt.Sprintf("Delete member type failed: not found (ID: %s): %v", memberTypeID, err),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberType with ID %s not found: %v", memberTypeID, err)})
		}
		if err := core.MemberTypeManager(service).Delete(context, *memberTypeID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member type failed: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member type: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member type: " + memberType.Name,
			Module:      "MemberType",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/member-type/bulk-delete",
		Method:      "DELETE",
		RequestType: core.IDSRequest{},
		Note:        "Deletes multiple member type records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member types failed (/member-type/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member types failed (/member-type/bulk-delete) | no IDs provided",
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.MemberTypeManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member types failed (/member-type/bulk-delete) | error: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member types: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member types (/member-type/bulk-delete)",
			Module:      "MemberType",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
