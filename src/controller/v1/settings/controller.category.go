package settings

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func CategoryController(service *horizon.HorizonService) {
	

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/category",
		Method:       "GET",
		Note:         "Returns all categories in the system.",
		ResponseType: types.CategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categories, err := core.CategoryManager(service).List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve categories: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.CategoryManager(service).ToModels(categories))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/category/:category_id",
		Method:       "GET",
		Note:         "Returns a single category by its ID.",
		ResponseType: types.CategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categoryID, err := helpers.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category ID"})
		}

		category, err := core.CategoryManager(service).GetByIDRaw(context, *categoryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
		}

		return ctx.JSON(http.StatusOK, category)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/category",
		Method:       "POST",
		Note:         "Creates a new category.",
		RequestType:  types.CategoryRequest{},
		ResponseType: types.CategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.CategoryManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Category creation failed (/category), validation error: " + err.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category data: " + err.Error()})
		}

		category := &types.Category{
			Name:        req.Name,
			Description: req.Description,
			Color:       req.Color,
			Icon:        req.Icon,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		if err := core.CategoryManager(service).Create(context, category); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Category creation failed (/category), db error: " + err.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create category: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created category (/category): " + category.Name,
			Module:      "Category",
		})

		return ctx.JSON(http.StatusCreated, core.CategoryManager(service).ToModel(category))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/category/:category_id",
		Method:       "PUT",
		Note:         "Updates an existing category by its ID.",
		RequestType:  types.CategoryRequest{},
		ResponseType: types.CategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categoryID, err := helpers.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Category update failed (/category/:category_id), invalid category ID.",
				Module:      "Category",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category ID"})
		}

		req, err := core.CategoryManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Category update failed (/category/:category_id), validation error: " + err.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category data: " + err.Error()})
		}

		category, err := core.CategoryManager(service).GetByID(context, *categoryID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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

		if err := core.CategoryManager(service).UpdateByID(context, category.ID, category); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Category update failed (/category/:category_id), db error: " + err.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update category: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated category (/category/:category_id): " + category.Name,
			Module:      "Category",
		})

		return ctx.JSON(http.StatusOK, core.CategoryManager(service).ToModel(category))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/category/:category_id",
		Method: "DELETE",
		Note:   "Deletes the specified category by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categoryID, err := helpers.EngineUUIDParam(ctx, "category_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Category delete failed (/category/:category_id), invalid category ID.",
				Module:      "Category",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category ID"})
		}

		category, err := core.CategoryManager(service).GetByID(context, *categoryID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Category delete failed (/category/:category_id), not found.",
				Module:      "Category",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
		}

		if err := core.CategoryManager(service).Delete(context, *categoryID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Category delete failed (/category/:category_id), db error: " + err.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete category: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted category (/category/:category_id): " + category.Name,
			Module:      "Category",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/category/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple categories by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/category/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/category/bulk-delete) | no IDs provided",
				Module:      "Category",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.CategoryManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/category/bulk-delete) | error: " + err.Error(),
				Module:      "Category",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete categories: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted categories (/category/bulk-delete)",
			Module:      "Category",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
