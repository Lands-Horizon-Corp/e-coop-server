package v1

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CategoryController manages endpoints for categories.
func (c *Controller) categoryController() {
	req := c.provider.Service.Request

	// GET /category: List all categories. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/category",
		Method:       "GET",
		Note:         "Returns all categories in the system.",
		ResponseType: core.CategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categories, err := c.core.CategoryManager.List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve categories: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.CategoryManager.ToModels(categories))
	})

	// GET /category/:category_id: Get a specific category by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/category/:category_id",
		Method:       "GET",
		Note:         "Returns a single category by its ID.",
		ResponseType: core.CategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categoryID, err := handlers.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category ID"})
		}

		category, err := c.core.CategoryManager.GetByIDRaw(context, *categoryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
		}

		return ctx.JSON(http.StatusOK, category)
	})

	// POST /category: Create a new category. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/category",
		Method:       "POST",
		Note:         "Creates a new category.",
		RequestType:  core.CategoryRequest{},
		ResponseType: core.CategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.CategoryManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Category creation failed (/category), validation error: " + err.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category data: " + err.Error()})
		}

		category := &core.Category{
			Name:        req.Name,
			Description: req.Description,
			Color:       req.Color,
			Icon:        req.Icon,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		if err := c.core.CategoryManager.Create(context, category); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Category creation failed (/category), db error: " + err.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create category: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created category (/category): " + category.Name,
			Module:      "Category",
		})

		return ctx.JSON(http.StatusCreated, c.core.CategoryManager.ToModel(category))
	})

	// PUT /category/:category_id: Update a category by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/category/:category_id",
		Method:       "PUT",
		Note:         "Updates an existing category by its ID.",
		RequestType:  core.CategoryRequest{},
		ResponseType: core.CategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categoryID, err := handlers.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Category update failed (/category/:category_id), invalid category ID.",
				Module:      "Category",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category ID"})
		}

		req, err := c.core.CategoryManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Category update failed (/category/:category_id), validation error: " + err.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category data: " + err.Error()})
		}

		category, err := c.core.CategoryManager.GetByID(context, *categoryID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Category update failed (/category/:category_id), not found.",
				Module:      "Category",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
		}

		category.Color = req.Color
		category.Name = req.Name
		category.Description = req.Description
		category.Icon = req.Icon
		category.UpdatedAt = time.Now().UTC()

		if err := c.core.CategoryManager.UpdateByID(context, category.ID, category); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Category update failed (/category/:category_id), db error: " + err.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update category: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated category (/category/:category_id): " + category.Name,
			Module:      "Category",
		})

		return ctx.JSON(http.StatusOK, c.core.CategoryManager.ToModel(category))
	})

	// DELETE /category/:category_id: Delete a category by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/category/:category_id",
		Method: "DELETE",
		Note:   "Deletes the specified category by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categoryID, err := handlers.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Category delete failed (/category/:category_id), invalid category ID.",
				Module:      "Category",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category ID"})
		}

		category, err := c.core.CategoryManager.GetByID(context, *categoryID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Category delete failed (/category/:category_id), not found.",
				Module:      "Category",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
		}

		if err := c.core.CategoryManager.Delete(context, *categoryID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Category delete failed (/category/:category_id), db error: " + err.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete category: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted category (/category/:category_id): " + category.Name,
			Module:      "Category",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /category/bulk-delete: Bulk delete categories by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/category/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple categories by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/category/bulk-delete), invalid request body.",
				Module:      "Category",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/category/bulk-delete), no IDs provided.",
				Module:      "Category",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/category/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		var sb strings.Builder
		for _, rawID := range reqBody.IDs {
			categoryID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/category/bulk-delete), invalid UUID: " + rawID,
					Module:      "Category",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}

			category, err := c.core.CategoryManager.GetByID(context, categoryID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/category/bulk-delete), not found: " + rawID,
					Module:      "Category",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Category not found with ID: %s", rawID)})
			}

			sb.WriteString(category.Name)
			sb.WriteByte(',')

			if err := c.core.CategoryManager.DeleteWithTx(context, tx, categoryID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/category/bulk-delete), db error: " + err.Error(),
					Module:      "Category",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete category: " + err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/category/bulk-delete), commit error: " + err.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted categories (/category/bulk-delete): " + sb.String(),
			Module:      "Category",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
