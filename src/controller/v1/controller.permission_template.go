package controller_v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelCore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/labstack/echo/v4"
)

// PermissionTemplateController registers all routes related to permission templates.
func (c *Controller) PermissionTemplateController() {
	req := c.provider.Service.Request

	// Fetch all permission templates associated with the current user's branch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/permission-template",
		Method:       "GET",
		ResponseType: modelCore.PermissionTemplateResponse{},
		Note:         "Returns all permission templates for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		permissionTemplates, err := c.modelCore.GetPermissionTemplateByBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve permission templates: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.PermissionTemplateManager.Filtered(context, ctx, permissionTemplates))
	})

	// Fetch all permission templates (paginated) for the current user's branch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/permission-template/search",
		Method:       "GET",
		ResponseType: modelCore.PermissionTemplateResponse{},
		Note:         "Returns all permission templates (paginated) for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		permissionTemplates, err := c.modelCore.GetPermissionTemplateByBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve permission templates: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.PermissionTemplateManager.Pagination(context, ctx, permissionTemplates))
	})

	// Fetch a single permission template by its ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/permission-template/:permission_template_id",
		Method:       "GET",
		ResponseType: modelCore.PermissionTemplateResponse{},
		Note:         "Returns a specific permission template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		permissionTemplateID, err := handlers.EngineUUIDParam(ctx, "permission_template_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid permission_template_id: " + err.Error()})
		}

		permissionTemplate, err := c.modelCore.PermissionTemplateManager.GetByID(context, *permissionTemplateID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Permission template not found: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.modelCore.PermissionTemplateManager.ToModel(permissionTemplate))
	})

	// Create a new permission template.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/permission-template",
		Method:       "POST",
		RequestType:  modelCore.PermissionTemplateRequest{},
		ResponseType: modelCore.PermissionTemplateResponse{},
		Note:         "Creates a new permission template for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		reqData, err := c.modelCore.PermissionTemplateManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create permission template failed: validation error: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create permission template failed: user org error: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		newTemplate := &modelCore.PermissionTemplate{
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

		if err := c.modelCore.PermissionTemplateManager.Create(context, newTemplate); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create permission template failed: create error: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create permission template: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created permission template: " + newTemplate.Name,
			Module:      "PermissionTemplate",
		})

		return ctx.JSON(http.StatusOK, c.modelCore.PermissionTemplateManager.ToModel(newTemplate))
	})

	// Update an existing permission template by its ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/permission-template/:permission_template_id",
		Method:       "PUT",
		RequestType:  modelCore.PermissionTemplateRequest{},
		ResponseType: modelCore.PermissionTemplateResponse{},
		Note:         "Updates an existing permission template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		permissionTemplateID, err := handlers.EngineUUIDParam(ctx, "permission_template_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission template failed: invalid permission_template_id: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid permission_template_id: " + err.Error()})
		}

		reqData, err := c.modelCore.PermissionTemplateManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission template failed: validation error: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		template, err := c.modelCore.PermissionTemplateManager.GetByID(context, *permissionTemplateID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission template failed: not found: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Permission template not found: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission template failed: user org error: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		template.UpdatedAt = time.Now().UTC()
		template.UpdatedByID = userOrg.UserID
		template.OrganizationID = userOrg.OrganizationID
		template.BranchID = *userOrg.BranchID
		template.Name = reqData.Name
		template.Description = reqData.Description
		template.Permissions = reqData.Permissions

		if err := c.modelCore.PermissionTemplateManager.UpdateFields(context, template.ID, template); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission template failed: update error: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update permission template: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated permission template: " + template.Name,
			Module:      "PermissionTemplate",
		})

		return ctx.JSON(http.StatusOK, c.modelCore.PermissionTemplateManager.ToModel(template))
	})

	// Delete a permission template by its ID.
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/permission-template/:permission_template_id",
		Method: "DELETE",
		Note:   "Deletes a permission template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		permissionTemplateID, err := handlers.EngineUUIDParam(ctx, "permission_template_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete permission template failed: invalid permission_template_id: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid permission_template_id: " + err.Error()})
		}

		template, err := c.modelCore.PermissionTemplateManager.GetByID(context, *permissionTemplateID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete permission template failed: not found: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Permission template not found: " + err.Error()})
		}

		if err := c.modelCore.PermissionTemplateManager.DeleteByID(context, *permissionTemplateID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete permission template failed: delete error: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete permission template: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted permission template: " + template.Name,
			Module:      "PermissionTemplate",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
