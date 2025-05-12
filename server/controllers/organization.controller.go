package controllers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /organization-category
func (c *Controller) OrganizationCategoryList(ctx echo.Context) error {
	organizationCategory, err := c.organizationCategory.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.OrganizationCategoryModels(organizationCategory))
}

// GET /organization-category/:organization_category_id
func (c *Controller) OrganizationCategoryGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_category_id")
	if err != nil {
		return err
	}
	organizationCategory, err := c.organizationCategory.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.OrganizationCategoryModel(organizationCategory))
}

// POST /organization-category/:organization_id
func (c *Controller) OrganizationCategoryCreate(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	req, err := c.model.OrganizationCategoryValidate(ctx)
	if err != nil {
		return err
	}
	categoryId, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid feedback ID"})
	}

	model := &model.OrganizationCategory{
		OrganizationID: orgId,
		CategoryID:     &categoryId,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := c.organizationCategory.Manager.Create(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.OrganizationCategoryModel(model))
}

// PUT /:organization_category_id/organization/:organization_id
func (c *Controller) OrganizationCategoryUpdate(ctx echo.Context) error {
	orgCategoryId, err := horizon.EngineUUIDParam(ctx, "organization_category_id")
	if err != nil {
		return err
	}
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	req, err := c.model.OrganizationCategoryValidate(ctx)
	if err != nil {
		return err
	}
	catID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid category ID"})
	}
	existing, err := c.organizationCategory.Manager.GetByID(*orgCategoryId)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}
	existing.OrganizationID = orgId
	existing.CategoryID = &catID
	existing.UpdatedAt = time.Now().UTC()
	if err := c.organizationCategory.Manager.Update(existing); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.OrganizationCategoryModel(existing))
}

// DELETE /organization-category/:organization_category_id
func (c *Controller) OrganizationCategoryDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_category_id")
	if err != nil {
		return err
	}
	if err := c.organizationCategory.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET /organization-category/category/:category_id
func (c *Controller) OrganizationCategoryListByCategory(ctx echo.Context) error {
	categoryId, err := horizon.EngineUUIDParam(ctx, "category_id")
	if err != nil {
		return err
	}
	organizationCategory, err := c.organizationCategory.ListByCategory(categoryId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.OrganizationCategoryModels(organizationCategory))
}

// GET organization-category/organizaton/:organization_id
func (c *Controller) OrganizationCategoryListByOrganization(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	organizationCategory, err := c.organizationCategory.ListByOrganization(orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.OrganizationCategoryModels(organizationCategory))
}
