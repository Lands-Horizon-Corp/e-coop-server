package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func permissionTemplateController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/permission-template",
		Method:       "GET",
		ResponseType: core.PermissionTemplateResponse{},
		Note:         "Returns all permission templates for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		permissionTemplates, err := core.GetPermissionTemplateBybranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve permission templates: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.PermissionTemplateManager(service).ToModels(permissionTemplates))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/permission-template/search",
		Method:       "GET",
		ResponseType: core.PermissionTemplateResponse{},
		Note:         "Returns all permission templates (paginated) for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		permissionTemplates, err := core.PermissionTemplateManager(service).NormalPagination(context, ctx, &core.PermissionTemplate{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve permission templates: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, permissionTemplates)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/permission-template/:permission_template_id",
		Method:       "GET",
		ResponseType: core.PermissionTemplateResponse{},
		Note:         "Returns a specific permission template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		permissionTemplateID, err := helpers.EngineUUIDParam(ctx, "permission_template_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid permission_template_id: " + err.Error()})
		}

		permissionTemplate, err := core.PermissionTemplateManager(service).GetByID(context, *permissionTemplateID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Permission template not found: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, core.PermissionTemplateManager(service).ToModel(permissionTemplate))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/permission-template",
		Method:       "POST",
		RequestType:  core.PermissionTemplateRequest{},
		ResponseType: core.PermissionTemplateResponse{},
		Note:         "Creates a new permission template for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := core.PermissionTemplateManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create permission template failed: validation error: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create permission template failed: user org error: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		newTemplate := &core.PermissionTemplate{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Name:           req.Name,
			Description:    req.Description,
			Permissions:    req.Permissions,
		}

		if err := core.PermissionTemplateManager(service).Create(context, newTemplate); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create permission template failed: create error: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create permission template: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created permission template: " + newTemplate.Name,
			Module:      "PermissionTemplate",
		})

		return ctx.JSON(http.StatusOK, core.PermissionTemplateManager(service).ToModel(newTemplate))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/permission-template/:permission_template_id",
		Method:       "PUT",
		RequestType:  core.PermissionTemplateRequest{},
		ResponseType: core.PermissionTemplateResponse{},
		Note:         "Updates an existing permission template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		permissionTemplateID, err := helpers.EngineUUIDParam(ctx, "permission_template_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission template failed: invalid permission_template_id: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid permission_template_id: " + err.Error()})
		}

		req, err := core.PermissionTemplateManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission template failed: validation error: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		template, err := core.PermissionTemplateManager(service).GetByID(context, *permissionTemplateID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission template failed: not found: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Permission template not found: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		template.Name = req.Name
		template.Description = req.Description
		template.Permissions = req.Permissions

		if err := core.PermissionTemplateManager(service).UpdateByID(context, template.ID, template); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission template failed: update error: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update permission template: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated permission template: " + template.Name,
			Module:      "PermissionTemplate",
		})

		return ctx.JSON(http.StatusOK, core.PermissionTemplateManager(service).ToModel(template))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/permission-template/:permission_template_id",
		Method: "DELETE",
		Note:   "Deletes a permission template by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		permissionTemplateID, err := helpers.EngineUUIDParam(ctx, "permission_template_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete permission template failed: invalid permission_template_id: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid permission_template_id: " + err.Error()})
		}

		template, err := core.PermissionTemplateManager(service).GetByID(context, *permissionTemplateID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete permission template failed: not found: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Permission template not found: " + err.Error()})
		}

		if err := core.PermissionTemplateManager(service).Delete(context, *permissionTemplateID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete permission template failed: delete error: " + err.Error(),
				Module:      "PermissionTemplate",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete permission template: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted permission template: " + template.Name,
			Module:      "PermissionTemplate",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
