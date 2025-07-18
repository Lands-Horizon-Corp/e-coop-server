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

// CategoryController manages endpoints for categories.
func (c *Controller) CategoryController() {
	req := c.provider.Service.Request

	// GET /category: List all categories.
	req.RegisterRoute(horizon.Route{
		Route:    "/category",
		Method:   "GET",
		Response: "TCategory[]",
		Note:     "Returns all categories in the system.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categories, err := c.model.CategoryManager.ListRaw(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve categories: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, categories)
	})

	// GET /category/:category_id: Get a specific category by ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/category/:category_id",
		Method:   "GET",
		Response: "TCategory",
		Note:     "Returns a single category by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categoryID, err := horizon.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category ID"})
		}

		category, err := c.model.CategoryManager.GetByIDRaw(context, *categoryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
		}

		return ctx.JSON(http.StatusOK, category)
	})

	// POST /category: Create a new category.
	req.RegisterRoute(horizon.Route{
		Route:    "/category",
		Method:   "POST",
		Request:  "TCategory",
		Response: "TCategory",
		Note:     "Creates a new category.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.CategoryManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category data: " + err.Error()})
		}

		category := &model.Category{
			Name:        req.Name,
			Description: req.Description,
			Color:       req.Color,
			Icon:        req.Icon,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		if err := c.model.CategoryManager.Create(context, category); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create category: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, c.model.CategoryManager.ToModel(category))
	})

	// PUT /category/:category_id: Update a category by ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/category/:category_id",
		Method:   "PUT",
		Request:  "TCategory",
		Response: "TCategory",
		Note:     "Updates an existing category by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categoryID, err := horizon.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category ID"})
		}

		req, err := c.model.CategoryManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category data: " + err.Error()})
		}

		category, err := c.model.CategoryManager.GetByID(context, *categoryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
		}

		category.Color = req.Color
		category.Name = req.Name
		category.Description = req.Description
		category.Icon = req.Icon
		category.UpdatedAt = time.Now().UTC()

		if err := c.model.CategoryManager.UpdateFields(context, category.ID, category); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update category: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.CategoryManager.ToModel(category))
	})

	// DELETE /category/:category_id: Delete a category by ID.
	req.RegisterRoute(horizon.Route{
		Route:  "/category/:category_id",
		Method: "DELETE",
		Note:   "Deletes the specified category by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categoryID, err := horizon.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category ID"})
		}

		if err := c.model.CategoryManager.DeleteByID(context, *categoryID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete category: " + err.Error()})
		}

		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /category/bulk-delete: Bulk delete categories by IDs.
	req.RegisterRoute(horizon.Route{
		Route:   "/category/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple categories by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}

		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		if len(reqBody.IDs) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			categoryID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}

			if _, err := c.model.CategoryManager.GetByID(context, categoryID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Category not found with ID: %s", rawID)})
			}

			if err := c.model.CategoryManager.DeleteByIDWithTx(context, tx, categoryID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete category: " + err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
