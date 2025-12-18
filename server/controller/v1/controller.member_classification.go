package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) memberClassificationController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-classification-history",
		Method:       "GET",
		ResponseType: core.MemberClassificationHistoryResponse{},
		Note:         "Returns all member classification history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberClassificationHistory, err := c.core.MemberClassificationHistoryCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member classification history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MemberClassificationHistoryManager.ToModels(memberClassificationHistory))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-classification-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: core.MemberClassificationHistoryResponse{},
		Note:         "Returns member classification history for a specific member profile ID.",
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
		memberClassificationHistory, err := c.core.MemberClassificationHistoryManager.NormalPagination(context, ctx, &core.MemberClassificationHistory{
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			MemberProfileID: *memberProfileID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member classification history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberClassificationHistory)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-classification",
		Method:       "GET",
		ResponseType: core.MemberClassificationResponse{},
		Note:         "Returns all member classifications for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberClassification, err := c.core.MemberClassificationCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member classifications: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MemberClassificationManager.ToModels(memberClassification))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-classification/search",
		Method:       "GET",
		ResponseType: core.MemberClassificationResponse{},
		Note:         "Returns paginated member classifications for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.core.MemberClassificationManager.NormalPagination(context, ctx, &core.MemberClassification{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member classifications for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-classification",
		Method:       "POST",
		ResponseType: core.MemberClassificationResponse{},
		RequestType:  core.MemberClassificationRequest{},
		Note:         "Creates a new member classification record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.MemberClassificationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member classification failed (/member-classification), validation error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member classification failed (/member-classification), user org error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberClassification := &core.MemberClassification{
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

		if err := c.core.MemberClassificationManager.Create(context, memberClassification); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member classification failed (/member-classification), db error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member classification: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member classification (/member-classification): " + memberClassification.Name,
			Module:      "MemberClassification",
		})

		return ctx.JSON(http.StatusOK, c.core.MemberClassificationManager.ToModel(memberClassification))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-classification/:member_classification_id",
		Method:       "PUT",
		ResponseType: core.MemberClassificationResponse{},
		RequestType:  core.MemberClassificationRequest{},
		Note:         "Updates an existing member classification record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberClassificationID, err := handlers.EngineUUIDParam(ctx, "member_classification_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), invalid member_classification_id: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_classification_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), user org error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.core.MemberClassificationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), validation error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberClassification, err := c.core.MemberClassificationManager.GetByID(context, *memberClassificationID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		if err := c.core.MemberClassificationManager.UpdateByID(context, memberClassification.ID, memberClassification); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), db error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member classification: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member classification (/member-classification/:member_classification_id): " + memberClassification.Name,
			Module:      "MemberClassification",
		})
		return ctx.JSON(http.StatusOK, c.core.MemberClassificationManager.ToModel(memberClassification))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/member-classification/:member_classification_id",
		Method: "DELETE",
		Note:   "Deletes a member classification record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberClassificationID, err := handlers.EngineUUIDParam(ctx, "member_classification_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member classification failed (/member-classification/:member_classification_id), invalid member_classification_id: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_classification_id: " + err.Error()})
		}
		value, err := c.core.MemberClassificationManager.GetByID(context, *memberClassificationID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member classification failed (/member-classification/:member_classification_id), not found: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member classification not found: " + err.Error()})
		}
		if err := c.core.MemberClassificationManager.Delete(context, *memberClassificationID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member classification failed (/member-classification/:member_classification_id), db error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member classification: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member classification (/member-classification/:member_classification_id): " + value.Name,
			Module:      "MemberClassification",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/member-classification/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member classification records by their IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member classifications failed (/member-classification/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		if err := c.core.MemberClassificationManager.BulkDelete(context, ids); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member classifications failed (/member-classification/bulk-delete) | error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member classifications: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member classifications (/member-classification/bulk-delete)",
			Module:      "MemberClassification",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
