package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

// PermissionTemplateController registers all routes related to permission templates.
func (c *Controller) PermissionTemplateController() {

	req := c.provider.Service.Request
	req.RegisterRoute(horizon.Route{
		Route:    "/permission-template",
		Method:   "GET",
		Response: "TPermissionTemplate[]",
		Note:     "Fetches all permission templates associated with the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		permissionTemplates, err := c.model.GetPermissionTemplateByBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.PermissionTemplateManager.ToModels(permissionTemplates))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/permission-template/:permission_template_id",
		Method:   "GET",
		Response: "TPermissionTemplate",
		Note:     "Fetches a single permission template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		permissionTemplateID, err := horizon.EngineUUIDParam(ctx, "permission_template_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid permission template ID")
		}

		permissionTemplate, err := c.model.PermissionTemplateManager.GetByID(context, *permissionTemplateID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusAccepted, c.model.PermissionTemplateManager.ToModel(permissionTemplate))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/permission-template",
		Method:   "POST",
		Response: "TPermissionTemplate",
		Request:  "TPermissionTemplate",
		Note:     "Creates a new permission template.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Validate request payload
		reqData, err := c.model.PermissionTemplateManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}

		newTemplate := &model.PermissionTemplate{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Name:           reqData.Name,
			Description:    reqData.Description,
			Permissions:    reqData.Permissions,
		}

		if err := c.model.PermissionTemplateManager.Create(context, newTemplate); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.PermissionTemplateManager.ToModel(newTemplate))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/permission-template/:permission_template_id",
		Method:   "PUT",
		Response: "TPermissionTemplate",
		Request:  "TPermissionTemplate",
		Note:     "Updates an existing permission template identified by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		permissionTemplateID, err := horizon.EngineUUIDParam(ctx, "permission_template_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid permission template ID")
		}

		reqData, err := c.model.PermissionTemplateManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		template, err := c.model.PermissionTemplateManager.GetByID(context, *permissionTemplateID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}

		template.UpdatedAt = time.Now().UTC()
		template.UpdatedByID = userOrg.UserID
		template.OrganizationID = userOrg.OrganizationID
		template.BranchID = *userOrg.BranchID
		template.Name = reqData.Name
		template.Description = reqData.Description
		template.Permissions = reqData.Permissions

		if err := c.model.PermissionTemplateManager.UpdateByID(context, template.ID, template); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update permission template: "+err.Error())
		}

		return ctx.JSON(http.StatusOK, c.model.PermissionTemplateManager.ToModel(template))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/permission-template/:permission_template_id",
		Method: "DELETE",
		Note:   "Deletes a specific permission template identified by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		permissionTemplateID, err := horizon.EngineUUIDParam(ctx, "permission_template_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid permission template ID")
		}

		if err := c.model.PermissionTemplateManager.DeleteByID(context, *permissionTemplateID); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
