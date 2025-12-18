package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) tagTemplateController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/tag-template",
		Method:       "GET",
		ResponseType: core.TagTemplateResponse{},
		Note:         "Returns all tag templates for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		templates, err := c.core.TagTemplateManager.Find(context, &core.TagTemplate{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve tag templates: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.TagTemplateManager.ToModels(templates))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/tag-template/search",
		Method:       "GET",
		ResponseType: core.TagTemplateResponse{},
		Note:         "Returns paginated tag templates for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.core.TagTemplateManager.NormalPagination(context, ctx, &core.TagTemplate{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve tag templates for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/tag-template/:tag_template_id",
		Method:       "GET",
		ResponseType: core.TagTemplateResponse{},
		Note:         "Returns a specific tag template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tag_template_id: " + err.Error()})
		}
		template, err := c.core.TagTemplateManager.GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "TagTemplate not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, template)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/tag-template",
		Method:       "POST",
		ResponseType: core.TagTemplateResponse{},
		RequestType:  core.TagTemplateRequest{},
		Note:         "Creates a new tag template for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.TagTemplateManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create tag template failed: validation error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create tag template failed: user org error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		template := &core.TagTemplate{
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

		if err := c.core.TagTemplateManager.Create(context, template); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create tag template failed: create error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create tag template: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created tag template: " + template.Name,
			Module:      "TagTemplate",
		})

		return ctx.JSON(http.StatusOK, c.core.TagTemplateManager.ToModel(template))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/tag-template/:tag_template_id",
		Method:       "PUT",
		ResponseType: core.TagTemplateResponse{},
		RequestType:  core.TagTemplateRequest{},
		Note:         "Updates an existing tag template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update tag template failed: invalid tag_template_id: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tag_template_id: " + err.Error()})
		}

		req, err := c.core.TagTemplateManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update tag template failed: validation error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update tag template failed: user org error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		template, err := c.core.TagTemplateManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		if err := c.core.TagTemplateManager.UpdateByID(context, template.ID, template); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update tag template failed: update error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update tag template: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated tag template: " + template.Name,
			Module:      "TagTemplate",
		})
		return ctx.JSON(http.StatusOK, c.core.TagTemplateManager.ToModel(template))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/tag-template/:tag_template_id",
		Method: "DELETE",
		Note:   "Deletes a tag template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete tag template failed: invalid tag_template_id: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tag_template_id: " + err.Error()})
		}
		template, err := c.core.TagTemplateManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete tag template failed: not found: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "TagTemplate not found: " + err.Error()})
		}
		if err := c.core.TagTemplateManager.Delete(context, *id); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete tag template failed: delete error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete tag template: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted tag template: " + template.Name,
			Module:      "TagTemplate",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/tag-template/bulk-delete",
		Method:      "DELETE",
		RequestType: core.IDSRequest{},
		Note:        "Deletes multiple tag template records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Tag template bulk delete failed (/tag-template/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		if err := c.core.TagTemplateManager.BulkDelete(context, ids); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Tag template bulk delete failed (/tag-template/bulk-delete) | error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete tag templates: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted tag templates (/tag-template/bulk-delete)",
			Module:      "TagTemplate",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
