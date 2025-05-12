package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /permission-template
func (c *Controller) PermissionTemplateList(ctx echo.Context) error {
	permission_template, err := c.permissionTemplate.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.PermissionTemplateModels(permission_template))
}

// GET /permission-template/:permission_template_id
func (c *Controller) PermissionTemplateGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "permission_template_id")
	if err != nil {
		return err
	}
	permission_template, err := c.permissionTemplate.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.PermissionTemplateModel(permission_template))
}

// PUT /permission-template/permission_template_id
func (c *Controller) PermissionTemplateUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "permission_template_id")
	if err != nil {
		return err
	}
	req, err := c.model.PermissionTemplateValidate(ctx)
	if err != nil {
		return err
	}
	existing, err := c.permissionTemplate.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Permission template not found"})
	}
	userOrganization, err := c.provider.UserOwnerEmployee(ctx, existing.OrganizationID.String(), existing.BranchID.String())
	if err != nil {
		return err
	}
	model := &model.PermissionTemplate{
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		UpdatedAt:   time.Now().UTC(),
		UpdatedByID: userOrganization.UserID,
	}

	if err := c.permissionTemplate.Manager.UpdateByID(*id, model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.PermissionTemplateModel(model))
}

// POST /permission-template/organization/:organization_id/branch/:branch_id
func (c *Controller) PermissionTemplateCreate(ctx echo.Context) error {
	req, err := c.model.PermissionTemplateValidate(ctx)
	if err != nil {
		return err
	}
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	model := &model.PermissionTemplate{
		OrganizationID: *orgId,
		BranchID:       *branchId,
		Name:           req.Name,
		Description:    req.Description,
		Permissions:    req.Permissions,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	userOrganization, err := c.provider.UserOwnerEmployee(ctx, orgId.String(), branchId.String())
	if err != nil {
		return err
	}
	model.CreatedByID = userOrganization.UserID
	model.UpdatedByID = userOrganization.UserID
	if err := c.permissionTemplate.Manager.Create(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.PermissionTemplateModel(model))
}

// DELETE /permission-template/:permission_template_id
func (c *Controller) PermissionTemplateDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "permission_template_id")
	if err != nil {
		return err
	}
	template, err := c.permissionTemplate.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Permission template not found"})
	}

	_, err = c.provider.UserOwner(ctx, template.OrganizationID.String(), template.BranchID.String())
	if err != nil {
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Not authorized to delete this template"})
	}
	if err := c.permissionTemplate.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}
