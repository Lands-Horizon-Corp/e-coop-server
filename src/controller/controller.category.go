package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) CategoryController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/category",
		Method:   "GET",
		Response: "TCategory[]",
	}, func(ctx echo.Context) error {
		categories, err := c.model.CategoryManager.ListRaw(context.Background())
		if err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, categories)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/category/:category_id",
		Method:   "GET",
		Response: "TCategory",
	}, func(ctx echo.Context) error {
		categoryID, err := horizon.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid category ID")
		}

		category, err := c.model.CategoryManager.GetByIDRaw(context.Background(), *categoryID)
		if err != nil {
			return c.NotFound(ctx, "Category")
		}

		return ctx.JSON(http.StatusOK, category)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/category",
		Method:   "POST",
		Request:  "TCategory",
		Response: "TCategory",
	}, func(ctx echo.Context) error {
		req, err := c.model.CategoryManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		category := &model.Category{
			Name:        req.Name,
			Description: req.Description,
			Color:       req.Color,
			Icon:        req.Icon,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		if err := c.model.CategoryManager.Create(context.Background(), category); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.CategoryManager.ToModel(category))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/category/:category_id",
		Method:   "PUT",
		Request:  "TCategory",
		Response: "TCategory",
	}, func(ctx echo.Context) error {
		categoryID, err := horizon.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid category ID")
		}

		req, err := c.model.CategoryManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		category, err := c.model.CategoryManager.GetByID(context.Background(), *categoryID)
		if err != nil {
			return c.NotFound(ctx, "Category")
		}

		category.Color = req.Color
		category.Name = req.Name
		category.Description = req.Description
		category.Icon = req.Icon
		category.UpdatedAt = time.Now().UTC()

		if err := c.model.CategoryManager.UpdateFields(context.Background(), category.ID, category); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.CategoryManager.ToModel(category))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/category/:category_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		categoryID, err := horizon.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid category ID")
		}

		if err := c.model.CategoryManager.DeleteByID(context.Background(), *categoryID); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/category/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple category records",
	}, func(ctx echo.Context) error {
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
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		for _, rawID := range reqBody.IDs {
			categoryID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.CategoryManager.GetByID(context.Background(), categoryID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("Category with ID %s", rawID))
			}

			if err := c.model.CategoryManager.DeleteByIDWithTx(context.Background(), tx, categoryID); err != nil {
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
