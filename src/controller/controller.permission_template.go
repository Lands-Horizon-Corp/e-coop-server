package controller

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) PermissionTemplateController() {

	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/permission-template",
		Method:   "GET",
		Response: "TPermissionTemplate[]",
	}, func(ctx echo.Context) error {
		context := context.Background()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		permissionTemplate, err := c.model.GetPermissionTemplateByBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.PermissionTemplateManager.ToModels(permissionTemplate))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/permission-template/:permission_template_id",
		Method:   "GET",
		Response: "TPermissionTemplate",
	}, func(ctx echo.Context) error {
		context := context.Background()
		permissionTemplateId, err := horizon.EngineUUIDParam(ctx, "permission_template_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid permission template ID")
		}
		permissionTemplate, err := c.model.PermissionTemplateManager.GetByID(context, *permissionTemplateId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusAccepted, c.model.PermissionTemplateManager.ToModel(permissionTemplate))
	})
}
