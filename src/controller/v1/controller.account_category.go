package controller_v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) AccountCategoryController() {
	req := c.provider.Service.Request

	// SEARCH (GET) - NO FOOTSTEP

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-category/search",
		Method:       "GET",
		Note:         "Retrieve all account categories for the current branch.",
		ResponseType: model_core.AccountCategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != model_core.UserOrganizationTypeOwner && userOrg.UserType != model_core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		categories, err := c.model_core.AccountCategoryManager.Find(context, &model_core.AccountCategory{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account categories: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model_core.AccountCategoryManager.Pagination(context, ctx, categories))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-category",
		Method:       "GET",
		Note:         "Retrieve all account categories for the current branch (raw).",
		ResponseType: model_core.AccountCategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != model_core.UserOrganizationTypeOwner && userOrg.UserType != model_core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		categories, err := c.model_core.AccountCategoryManager.Find(context, &model_core.AccountCategory{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account categories (raw): " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model_core.AccountCategoryManager.Filtered(context, ctx, categories))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-category/:account_category_id",
		Method:       "GET",
		Note:         "Get an account category by ID.",
		ResponseType: model_core.AccountCategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "account_category_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account category ID: " + err.Error()})
		}
		category, err := c.model_core.AccountCategoryManager.GetByID(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account category not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model_core.AccountCategoryManager.ToModel(category))
	})

	// CREATE (POST) - ADD FOOTSTEP

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-category",
		Method:       "POST",
		Note:         "Create a new account category for the current branch.",
		ResponseType: model_core.AccountCategoryResponse{},
		RequestType:  model_core.AccountCategoryRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model_core.AccountCategoryManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed create account category (/account-category) | validation error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account category validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed create account category (/account-category) | user org error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != model_core.UserOrganizationTypeOwner && userOrg.UserType != model_core.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create account category attempt (/account-category)",
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}

		accountCategory := &model_core.AccountCategory{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			Name:           req.Name,
			Description:    req.Description,
		}

		if err := c.model_core.AccountCategoryManager.Create(context, accountCategory); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed create account category (/account-category) | db error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create account category: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created account category (/account-category): " + accountCategory.Name,
			Module:      "AccountCategory",
		})
		return ctx.JSON(http.StatusCreated, c.model_core.AccountCategoryManager.ToModel(accountCategory))
	})

	// UPDATE (PUT) - ADD FOOTSTEP

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-category/:account_category_id",
		Method:       "PUT",
		Note:         "Update an account category by ID.",
		ResponseType: model_core.AccountCategoryResponse{},
		RequestType:  model_core.AccountCategoryRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model_core.AccountCategoryManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed update account category (/account-category/:account_category_id) | validation error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account category validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed update account category (/account-category/:account_category_id) | user org error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != model_core.UserOrganizationTypeOwner && userOrg.UserType != model_core.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update account category attempt (/account-category/:account_category_id)",
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accountCategoryID, err := handlers.EngineUUIDParam(ctx, "account_category_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed update account category (/account-category/:account_category_id) | invalid UUID: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account category ID: " + err.Error()})
		}
		accountCategory, err := c.model_core.AccountCategoryManager.GetByID(context, *accountCategoryID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed update account category (/account-category/:account_category_id) | not found: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account category not found: " + err.Error()})
		}
		accountCategory.UpdatedByID = userOrg.UserID
		accountCategory.UpdatedAt = time.Now().UTC()
		accountCategory.Name = req.Name
		accountCategory.Description = req.Description
		accountCategory.BranchID = *userOrg.BranchID
		accountCategory.OrganizationID = userOrg.OrganizationID
		if err := c.model_core.AccountCategoryManager.UpdateFields(context, accountCategory.ID, accountCategory); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed update account category (/account-category/:account_category_id) | db error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account category: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated account category (/account-category/:account_category_id): " + accountCategory.Name,
			Module:      "AccountCategory",
		})
		return ctx.JSON(http.StatusOK, c.model_core.AccountCategoryManager.ToModel(accountCategory))
	})

	// DELETE (DELETE) - ADD FOOTSTEP

	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/account-category/:account_category_id",
		Method: "DELETE",
		Note:   "Delete an account category by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed delete account category (/account-category/:account_category_id) | user org error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != model_core.UserOrganizationTypeOwner && userOrg.UserType != model_core.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete account category attempt (/account-category/:account_category_id)",
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accountCategoryID, err := handlers.EngineUUIDParam(ctx, "account_category_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed delete account category (/account-category/:account_category_id) | invalid UUID: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account category ID: " + err.Error()})
		}
		accountCategory, err := c.model_core.AccountCategoryManager.GetByID(context, *accountCategoryID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed delete account category (/account-category/:account_category_id) | not found: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account category not found: " + err.Error()})
		}
		if err := c.model_core.AccountCategoryManager.DeleteByID(context, accountCategory.ID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed delete account category (/account-category/:account_category_id) | db error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account category: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted account category (/account-category/:account_category_id): " + accountCategory.Name,
			Module:      "AccountCategory",
		})
		return ctx.JSON(http.StatusOK, c.model_core.AccountCategoryManager.ToModel(accountCategory))
	})

	// BULK DELETE (DELETE) - ADD FOOTSTEP

	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/account-category/bulk-delete",
		Method:      "DELETE",
		Note:        "Bulk delete multiple account categories by IDs.",
		RequestType: model_core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model_core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account categories (/account-category/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account categories (/account-category/bulk-delete) | no IDs provided",
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided."})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account categories (/account-category/bulk-delete) | begin tx error: " + tx.Error.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Failed bulk delete account categories (/account-category/bulk-delete) | invalid UUID: " + rawID + " - " + err.Error(),
					Module:      "AccountCategory",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid UUID: " + rawID + " - " + err.Error()})
			}
			if _, err := c.model_core.AccountCategoryManager.GetByID(context, id); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Failed bulk delete account categories (/account-category/bulk-delete) | not found: " + rawID + " - " + err.Error(),
					Module:      "AccountCategory",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account category with ID " + rawID + " not found: " + err.Error()})
			}
			if err := c.model_core.AccountCategoryManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Failed bulk delete account categories (/account-category/bulk-delete) | delete error: " + rawID + " - " + err.Error(),
					Module:      "AccountCategory",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account category with ID " + rawID + ": " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account categories (/account-category/bulk-delete) | commit error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity: "bulk-delete-success",
			Description: "Bulk deleted account categories (/account-category/bulk-delete): IDs=" + func() string {
				b := ""
				for _, id := range reqBody.IDs {
					b += id + ","
				}
				return b
			}(),
			Module: "AccountCategory",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
