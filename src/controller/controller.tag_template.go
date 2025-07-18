package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) TagTemplateController() {
	req := c.provider.Service.Request

	// Returns all tag templates for the current user's branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/tag-template",
		Method:   "GET",
		Response: "TTagTemplate[]",
		Note:     "Returns all tag templates for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		templates, err := c.model.TagTemplateManager.Find(context, &model.TagTemplate{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve tag templates: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TagTemplateManager.ToModels(templates))
	})

	// Returns paginated tag templates for the current user's branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/tag-template/search",
		Method:   "GET",
		Request:  "Filter<ITagTemplate>",
		Response: "Paginated<ITagTemplate>",
		Note:     "Returns paginated tag templates for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.model.TagTemplateManager.Find(context, &model.TagTemplate{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve tag templates for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TagTemplateManager.Pagination(context, ctx, value))
	})

	// Returns a single tag template by its ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/tag-template/:tag_template_id",
		Method:   "GET",
		Response: "TTagTemplate",
		Note:     "Returns a specific tag template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tag_template_id: " + err.Error()})
		}
		template, err := c.model.TagTemplateManager.GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "TagTemplate not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, template)
	})

	// Creates a new tag template.
	req.RegisterRoute(horizon.Route{
		Route:    "/tag-template",
		Method:   "POST",
		Request:  "TTagTemplate",
		Response: "TTagTemplate",
		Note:     "Creates a new tag template for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.TagTemplateManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		template := &model.TagTemplate{
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

		if err := c.model.TagTemplateManager.Create(context, template); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create tag template: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.TagTemplateManager.ToModel(template))
	})

	// Updates an existing tag template by its ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/tag-template/:tag_template_id",
		Method:   "PUT",
		Request:  "TTagTemplate",
		Response: "TTagTemplate",
		Note:     "Updates an existing tag template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tag_template_id: " + err.Error()})
		}

		req, err := c.model.TagTemplateManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		template, err := c.model.TagTemplateManager.GetByID(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "TagTemplate not found: " + err.Error()})
		}
		template.Name = req.Name
		template.Description = req.Description
		template.Category = req.Category
		template.Color = req.Color
		template.Icon = req.Icon
		template.UpdatedAt = time.Now().UTC()
		template.UpdatedByID = user.UserID
		if err := c.model.TagTemplateManager.UpdateFields(context, template.ID, template); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update tag template: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TagTemplateManager.ToModel(template))
	})

	// Deletes a tag template by its ID.
	req.RegisterRoute(horizon.Route{
		Route:  "/tag-template/:tag_template_id",
		Method: "DELETE",
		Note:   "Deletes a tag template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tag_template_id: " + err.Error()})
		}
		if err := c.model.TagTemplateManager.DeleteByID(context, *id); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete tag template: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// Deletes multiple tag templates by their IDs.
	req.RegisterRoute(horizon.Route{
		Route:   "/tag-template/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple tag template records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s - %v", rawID, err)})
			}
			if _, err := c.model.TagTemplateManager.GetByID(context, id); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("TagTemplate with ID %s not found: %v", rawID, err)})
			}
			if err := c.model.TagTemplateManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete tag template with ID %s: %v", rawID, err)})
			}
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
