package v1

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) tagTemplateController() {
	req := c.provider.Service.Request

	// Returns all tag templates for the current user's branch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/tag-template",
		Method:       "GET",
		ResponseType: core.TagTemplateResponse{},
		Note:         "Returns all tag templates for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		templates, err := c.core.TagTemplateManager.Find(context, &core.TagTemplate{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve tag templates: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.TagTemplateManager.Filtered(context, ctx, templates))
	})

	// Returns paginated tag templates for the current user's branch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/tag-template/search",
		Method:       "GET",
		ResponseType: core.TagTemplateResponse{},
		Note:         "Returns paginated tag templates for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.core.TagTemplateManager.Find(context, &core.TagTemplate{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve tag templates for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.TagTemplateManager.Pagination(context, ctx, value))
	})

	// Returns a single tag template by its ID.
	req.RegisterRoute(handlers.Route{
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

	// Creates a new tag template.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/tag-template",
		Method:       "POST",
		ResponseType: core.TagTemplateResponse{},
		RequestType:  core.TagTemplateRequest{},
		Note:         "Creates a new tag template for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.TagTemplateManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create tag template failed: validation error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.core.TagTemplateManager.Create(context, template); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create tag template failed: create error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create tag template: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created tag template: " + template.Name,
			Module:      "TagTemplate",
		})

		return ctx.JSON(http.StatusOK, c.core.TagTemplateManager.ToModel(template))
	})

	// Updates an existing tag template by its ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/tag-template/:tag_template_id",
		Method:       "PUT",
		ResponseType: core.TagTemplateResponse{},
		RequestType:  core.TagTemplateRequest{},
		Note:         "Updates an existing tag template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update tag template failed: invalid tag_template_id: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tag_template_id: " + err.Error()})
		}

		req, err := c.core.TagTemplateManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update tag template failed: validation error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update tag template failed: user org error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		template, err := c.core.TagTemplateManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		template.UpdatedByID = user.UserID
		if err := c.core.TagTemplateManager.UpdateFields(context, template.ID, template); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update tag template failed: update error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update tag template: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated tag template: " + template.Name,
			Module:      "TagTemplate",
		})
		return ctx.JSON(http.StatusOK, c.core.TagTemplateManager.ToModel(template))
	})

	// Deletes a tag template by its ID.
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/tag-template/:tag_template_id",
		Method: "DELETE",
		Note:   "Deletes a tag template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete tag template failed: invalid tag_template_id: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tag_template_id: " + err.Error()})
		}
		template, err := c.core.TagTemplateManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete tag template failed: not found: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "TagTemplate not found: " + err.Error()})
		}
		if err := c.core.TagTemplateManager.DeleteByID(context, *id); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete tag template failed: delete error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete tag template: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted tag template: " + template.Name,
			Module:      "TagTemplate",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Deletes multiple tag templates by their IDs.
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/tag-template/bulk-delete",
		Method:      "DELETE",
		RequestType: core.IDSRequest{},
		Note:        "Deletes multiple tag template records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete tag templates failed: invalid request body.",
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete tag templates failed: no IDs provided.",
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete tag templates failed: begin tx error: " + tx.Error.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}
		var namesSlice []string
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete tag templates failed: invalid UUID: " + rawID,
					Module:      "TagTemplate",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s - %v", rawID, err)})
			}
			template, err := c.core.TagTemplateManager.GetByID(context, id)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete tag templates failed: not found: " + rawID,
					Module:      "TagTemplate",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("TagTemplate with ID %s not found: %v", rawID, err)})
			}
			namesSlice = append(namesSlice, template.Name)
			if err := c.core.TagTemplateManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete tag templates failed: delete error: " + err.Error(),
					Module:      "TagTemplate",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete tag template with ID %s: %v", rawID, err)})
			}
		}
		names := strings.Join(namesSlice, ",")

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete tag templates failed: commit tx error: " + err.Error(),
				Module:      "TagTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted tag templates: " + names,
			Module:      "TagTemplate",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
