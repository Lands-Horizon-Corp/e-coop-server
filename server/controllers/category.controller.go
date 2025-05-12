package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /category
func (c *Controller) CategoryList(ctx echo.Context) error {
	category, err := c.category.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.CategoryModels(category))
}

// GET /category/:category_id
func (c *Controller) CategoryGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "category_id")
	if err != nil {
		return err
	}
	category, err := c.category.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.CategoryModel(category))
}

// POST /category
func (c *Controller) CategoryCreate(ctx echo.Context) error {
	req, err := c.model.CategoryValidate(ctx)
	if err != nil {
		return err
	}
	model := &model.Category{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := c.category.Manager.Create(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.CategoryModel(model))
}

// PUT /category/category_id
func (c *Controller) CategoryUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "category_id")
	if err != nil {
		return err
	}
	req, err := c.model.CategoryValidate(ctx)
	if err != nil {
		return err
	}
	model := &model.Category{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		UpdatedAt:   time.Now().UTC(),
	}
	if err := c.category.Manager.UpdateByID(*id, model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.CategoryModel(model))
}

// DELETE /category/category_id
func (c *Controller) CategoryDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "category_id")
	if err != nil {
		return err
	}
	if err := c.category.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}
