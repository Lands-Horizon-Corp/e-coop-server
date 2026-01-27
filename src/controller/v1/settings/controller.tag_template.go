package settings

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

func TagTemplateController(service *horizon.HorizonService) {
	

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/tag-template",
		Method:       "GET",
		ResponseType: types.TagTemplateResponse{},
		Note:         "Returns all tag templates for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		templates, err := core.TagTemplateManager(service).Find(context, &types.TagTemplate{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve tag templates: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.TagTemplateManager(service).ToModels(templates))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/tag-template/search",
		Method:       "GET",
		ResponseType: types.TagTemplateResponse{},
		Note:         "Returns paginated tag templates for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := core.TagTemplateManager(service).NormalPagination(context, ctx, &types.TagTemplate{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve tag templates for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/tag-template/:tag_template_id",
		Method:       "GET",
		ResponseType: types.TagTemplateResponse{},
		Note:         "Returns a specific tag template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tag_template_id: " + err.Error()})
		}
		template, err := core.TagTemplateManager(service).GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "TagTemplate not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, template)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/tag-template",
		Method:       "POST",
		ResponseType: types.TagTemplateResponse{},
		RequestType:  types.TagTemplateRequest{},
		Note:         "Creates a new tag template for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.TagTemplateManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create tag template failed: validation error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create tag template failed: user org error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		template := &types.TagTemplate{
			Name:           req.Name,
			Description:    req.Description,
			Category:       req.Category,
			Color:          req.Color,
			Icon:           req.Icon,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.TagTemplateManager(service).Create(context, template); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create tag template failed: create error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create tag template: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created tag template: " + template.Name,
			Module:      "TagTemplate",
		})

		return ctx.JSON(http.StatusOK, core.TagTemplateManager(service).ToModel(template))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/tag-template/:tag_template_id",
		Method:       "PUT",
		ResponseType: types.TagTemplateResponse{},
		RequestType:  types.TagTemplateRequest{},
		Note:         "Updates an existing tag template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update tag template failed: invalid tag_template_id: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tag_template_id: " + err.Error()})
		}

		req, err := core.TagTemplateManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update tag template failed: validation error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update tag template failed: user org error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		template, err := core.TagTemplateManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update tag template failed: not found: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "TagTemplate not found: " + err.Error()})
		}
		template.Name = req.Name
		template.Description = req.Description
		template.Category = req.Category
		template.Color = req.Color
		template.Icon = req.Icon
		template.UpdatedAt = time.Now().UTC()
		template.UpdatedByID = userOrg.UserID
		if err := core.TagTemplateManager(service).UpdateByID(context, template.ID, template); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update tag template failed: update error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update tag template: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated tag template: " + template.Name,
			Module:      "TagTemplate",
		})
		return ctx.JSON(http.StatusOK, core.TagTemplateManager(service).ToModel(template))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/tag-template/:tag_template_id",
		Method: "DELETE",
		Note:   "Deletes a tag template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete tag template failed: invalid tag_template_id: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tag_template_id: " + err.Error()})
		}
		template, err := core.TagTemplateManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete tag template failed: not found: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "TagTemplate not found: " + err.Error()})
		}
		if err := core.TagTemplateManager(service).Delete(context, *id); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete tag template failed: delete error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete tag template: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted tag template: " + template.Name,
			Module:      "TagTemplate",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/tag-template/bulk-delete",
		Method:      "DELETE",
		RequestType: types.IDSRequest{},
		Note:        "Deletes multiple tag template records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Tag template bulk delete failed (/tag-template/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Tag template bulk delete failed (/tag-template/bulk-delete) | no IDs provided",
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.TagTemplateManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Tag template bulk delete failed (/tag-template/bulk-delete) | error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete tag templates: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted tag templates (/tag-template/bulk-delete)",
			Module:      "TagTemplate",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
