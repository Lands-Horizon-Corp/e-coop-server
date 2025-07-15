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

	req.RegisterRoute(horizon.Route{
		Route:    "/tag-template",
		Method:   "GET",
		Response: "TTagTemplate[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		templates, err := c.model.TagTemplateManager.Find(context, &model.TagTemplate{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
		})
		if err != nil {
			return c.NotFound(ctx, "TagTemplate")
		}
		return ctx.JSON(http.StatusOK, c.model.TagTemplateManager.ToModels(templates))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/tag-template/search",
		Method:   "GET",
		Request:  "Filter<ITagTemplate>",
		Response: "Paginated<ITagTemplate>",
		Note:     "Get pagination tag templates",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.TagTemplateManager.Find(context, &model.TagTemplate{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TagTemplateManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/tag-template/:tag_template_id",
		Method:   "GET",
		Response: "TTagTemplate",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid tag template ID")
		}
		template, err := c.model.TagTemplateManager.GetByIDRaw(context, *id)
		if err != nil {
			return c.NotFound(ctx, "TagTemplate")
		}
		return ctx.JSON(http.StatusOK, template)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/tag-template",
		Method:   "POST",
		Request:  "TTagTemplate",
		Response: "TTagTemplate",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.TagTemplateManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
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
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.TagTemplateManager.ToModel(template))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/tag-template/:tag_template_id",
		Method:   "PUT",
		Request:  "TTagTemplate",
		Response: "TTagTemplate",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid tag template ID")
		}

		req, err := c.model.TagTemplateManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		template, err := c.model.TagTemplateManager.GetByID(context, *id)
		if err != nil {
			return c.NotFound(ctx, "TagTemplate")
		}
		template.Name = req.Name
		template.Description = req.Description
		template.Category = req.Category
		template.Color = req.Color
		template.Icon = req.Icon
		template.UpdatedAt = time.Now().UTC()
		template.UpdatedByID = user.UserID
		if err := c.model.TagTemplateManager.UpdateFields(context, template.ID, template); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.TagTemplateManager.ToModel(template))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/tag-template/:tag_template_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "tag_template_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid tag template ID")
		}
		if err := c.model.TagTemplateManager.DeleteByID(context, *id); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/tag-template/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple tag template records",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}
		if len(reqBody.IDs) == 0 {
			return c.BadRequest(ctx, "No IDs provided")
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if _, err := c.model.TagTemplateManager.GetByID(context, id); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("TagTemplate with ID %s", rawID))
			}
			if err := c.model.TagTemplateManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				return c.InternalServerError(ctx, err)
			}
		}
		if err := tx.Commit().Error; err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
