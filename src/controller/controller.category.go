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
		category, err := c.model.CategoryManager.ListRaw(context.Background())
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, category)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/category/:category_id",
		Method:   "GET",
		Response: "TCategory",
	}, func(ctx echo.Context) error {
		context := context.Background()
		categoryId, err := horizon.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			return err
		}
		category, err := c.model.CategoryManager.GetByIDRaw(context, *categoryId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, category)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/category",
		Method:   "POST",
		Request:  "TCategory",
		Response: "TCategory",
	}, func(ctx echo.Context) error {
		context := context.Background()
		req, err := c.model.CategoryManager.Validate(ctx)
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
		if err := c.model.CategoryManager.Create(context, model); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.model.CategoryManager.ToModel(model))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/category/category_id",
		Method:   "PUT",
		Request:  "TCategory",
		Response: "TCategory",
	}, func(ctx echo.Context) error {
		context := context.Background()
		categoryId, err := horizon.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			return err
		}
		req, err := c.model.CategoryManager.Validate(ctx)
		if err != nil {
			return err
		}
		category, err := c.model.CategoryManager.GetByID(context, *categoryId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		category.Color = req.Color
		category.Name = req.Name
		category.Description = req.Description
		category.Icon = req.Icon
		category.UpdatedAt = time.Now().UTC()
		if err := c.model.CategoryManager.UpdateFields(context, category.ID, category); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.model.CategoryManager.ToModel(category))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/category/:category_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := context.Background()
		categoryId, err := horizon.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			return err
		}
		category, err := c.model.CategoryManager.GetByID(context, *categoryId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		if err := c.model.CategoryManager.DeleteByID(context, category.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/category/bulk-delete",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		backgroundCtx := context.Background()

		// Bind incoming JSON body to an inline struct
		reqBody := &struct {
			Ids []string `json:"ids"`
		}{}

		if err := ctx.Bind(reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		if len(reqBody.Ids) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "No IDs provided",
			})
		}

		// Start DB transaction
		tx := c.provider.Service.Database.Client().Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		for _, rawID := range reqBody.Ids {
			categoryID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{
					"error": fmt.Sprintf("Invalid UUID: %s", rawID),
				})
			}

			category, err := c.model.CategoryManager.GetByID(backgroundCtx, categoryID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{
					"error": fmt.Sprintf("Category not found for ID: %s", categoryID),
				})
			}

			if err := c.model.CategoryManager.DeleteByIDWithTx(backgroundCtx, tx, category.ID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{
					"error": fmt.Sprintf("Failed to delete category ID %s: %v", categoryID, err),
				})
			}
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to commit transaction: " + err.Error(),
			})
		}

		return ctx.NoContent(http.StatusNoContent)
	})

}
