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

func MemberClassificationController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-classification-history",
		Method:       "GET",
		ResponseType: types.MemberClassificationHistoryResponse{},
		Note:         "Returns all member classification history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberClassificationHistory, err := core.MemberClassificationHistoryCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member classification history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberClassificationHistoryManager(service).ToModels(memberClassificationHistory))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-classification-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: types.MemberClassificationHistoryResponse{},
		Note:         "Returns member classification history for a specific member profile ID.",
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
		memberClassificationHistory, err := core.MemberClassificationHistoryManager(service).NormalPagination(context, ctx, &types.MemberClassificationHistory{
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			MemberProfileID: *memberProfileID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member classification history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberClassificationHistory)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-classification",
		Method:       "GET",
		ResponseType: types.MemberClassificationResponse{},
		Note:         "Returns all member classifications for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberClassification, err := core.MemberClassificationCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member classifications: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberClassificationManager(service).ToModels(memberClassification))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-classification/search",
		Method:       "GET",
		ResponseType: types.MemberClassificationResponse{},
		Note:         "Returns paginated member classifications for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := core.MemberClassificationManager(service).NormalPagination(context, ctx, &types.MemberClassification{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member classifications for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-classification",
		Method:       "POST",
		ResponseType: types.MemberClassificationResponse{},
		RequestType:  types.MemberClassificationRequest{},
		Note:         "Creates a new member classification record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.MemberClassificationManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member classification failed (/member-classification), validation error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member classification failed (/member-classification), user org error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberClassification := &types.MemberClassification{
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

		if err := core.MemberClassificationManager(service).Create(context, memberClassification); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member classification failed (/member-classification), db error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member classification: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member classification (/member-classification): " + memberClassification.Name,
			Module:      "MemberClassification",
		})

		return ctx.JSON(http.StatusOK, core.MemberClassificationManager(service).ToModel(memberClassification))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-classification/:member_classification_id",
		Method:       "PUT",
		ResponseType: types.MemberClassificationResponse{},
		RequestType:  types.MemberClassificationRequest{},
		Note:         "Updates an existing member classification record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberClassificationID, err := helpers.EngineUUIDParam(ctx, "member_classification_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), invalid member_classification_id: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_classification_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), user org error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := core.MemberClassificationManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), validation error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberClassification, err := core.MemberClassificationManager(service).GetByID(context, *memberClassificationID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), not found: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member classification not found: " + err.Error()})
		}
		memberClassification.UpdatedAt = time.Now().UTC()
		memberClassification.UpdatedByID = userOrg.UserID
		memberClassification.OrganizationID = userOrg.OrganizationID
		memberClassification.BranchID = *userOrg.BranchID
		memberClassification.Name = req.Name
		memberClassification.Description = req.Description
		memberClassification.Icon = req.Icon
		if err := core.MemberClassificationManager(service).UpdateByID(context, memberClassification.ID, memberClassification); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), db error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member classification: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member classification (/member-classification/:member_classification_id): " + memberClassification.Name,
			Module:      "MemberClassification",
		})
		return ctx.JSON(http.StatusOK, core.MemberClassificationManager(service).ToModel(memberClassification))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-classification/:member_classification_id",
		Method: "DELETE",
		Note:   "Deletes a member classification record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberClassificationID, err := helpers.EngineUUIDParam(ctx, "member_classification_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member classification failed (/member-classification/:member_classification_id), invalid member_classification_id: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_classification_id: " + err.Error()})
		}
		value, err := core.MemberClassificationManager(service).GetByID(context, *memberClassificationID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member classification failed (/member-classification/:member_classification_id), not found: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member classification not found: " + err.Error()})
		}
		if err := core.MemberClassificationManager(service).Delete(context, *memberClassificationID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member classification failed (/member-classification/:member_classification_id), db error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member classification: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member classification (/member-classification/:member_classification_id): " + value.Name,
			Module:      "MemberClassification",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/member-classification/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member classification records by their IDs.",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member classifications failed (/member-classification/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member classifications failed (/member-classification/bulk-delete) | no IDs provided",
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.MemberClassificationManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member classifications failed (/member-classification/bulk-delete) | error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member classifications: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member classifications (/member-classification/bulk-delete)",
			Module:      "MemberClassification",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
