package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) accountCategoryController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/account-category/search",
		Method:       "GET",
		Note:         "Retrieve all account categories for the current branch.",
		ResponseType: core.AccountCategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}

		result, err := c.core.AccountCategoryManager.PaginationWithFields(context, ctx, &core.AccountCategory{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account categories: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, result)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/account-category",
		Method:       "GET",
		Note:         "Retrieve all account categories for the current branch (raw).",
		ResponseType: core.AccountCategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		categories, err := c.core.AccountCategoryManager.Find(context, &core.AccountCategory{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account categories (raw): " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.AccountCategoryManager.ToModels(categories))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/account-category/:account_category_id",
		Method:       "GET",
		Note:         "Get an account category by ID.",
		ResponseType: core.AccountCategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "account_category_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account category ID: " + err.Error()})
		}
		category, err := c.core.AccountCategoryManager.GetByID(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account category not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.AccountCategoryManager.ToModel(category))
	})

	// CREATE (POST) - ADD FOOTSTEP

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/account-category",
		Method:       "POST",
		Note:         "Create a new account category for the current branch.",
		ResponseType: core.AccountCategoryResponse{},
		RequestType:  core.AccountCategoryRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.AccountCategoryManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed create account category (/account-category) | validation error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account category validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed create account category (/account-category) | user org error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create account category attempt (/account-category)",
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}

		accountCategory := &core.AccountCategory{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			Name:           req.Name,
			Description:    req.Description,
		}

		if err := c.core.AccountCategoryManager.Create(context, accountCategory); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed create account category (/account-category) | db error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create account category: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created account category (/account-category): " + accountCategory.Name,
			Module:      "AccountCategory",
		})
		return ctx.JSON(http.StatusCreated, c.core.AccountCategoryManager.ToModel(accountCategory))
	})

	// UPDATE (PUT) - ADD FOOTSTEP

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/account-category/:account_category_id",
		Method:       "PUT",
		Note:         "Update an account category by ID.",
		ResponseType: core.AccountCategoryResponse{},
		RequestType:  core.AccountCategoryRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.AccountCategoryManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed update account category (/account-category/:account_category_id) | validation error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account category validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed update account category (/account-category/:account_category_id) | user org error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update account category attempt (/account-category/:account_category_id)",
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accountCategoryID, err := handlers.EngineUUIDParam(ctx, "account_category_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed update account category (/account-category/:account_category_id) | invalid UUID: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account category ID: " + err.Error()})
		}
		accountCategory, err := c.core.AccountCategoryManager.GetByID(context, *accountCategoryID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		if err := c.core.AccountCategoryManager.UpdateByID(context, accountCategory.ID, accountCategory); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed update account category (/account-category/:account_category_id) | db error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account category: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated account category (/account-category/:account_category_id): " + accountCategory.Name,
			Module:      "AccountCategory",
		})
		return ctx.JSON(http.StatusOK, c.core.AccountCategoryManager.ToModel(accountCategory))
	})

	// DELETE (DELETE) - ADD FOOTSTEP

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/account-category/:account_category_id",
		Method: "DELETE",
		Note:   "Delete an account category by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed delete account category (/account-category/:account_category_id) | user org error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete account category attempt (/account-category/:account_category_id)",
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accountCategoryID, err := handlers.EngineUUIDParam(ctx, "account_category_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed delete account category (/account-category/:account_category_id) | invalid UUID: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account category ID: " + err.Error()})
		}
		accountCategory, err := c.core.AccountCategoryManager.GetByID(context, *accountCategoryID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed delete account category (/account-category/:account_category_id) | not found: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account category not found: " + err.Error()})
		}
		if err := c.core.AccountCategoryManager.Delete(context, accountCategory.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed delete account category (/account-category/:account_category_id) | db error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account category: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted account category (/account-category/:account_category_id): " + accountCategory.Name,
			Module:      "AccountCategory",
		})
		return ctx.JSON(http.StatusOK, c.core.AccountCategoryManager.ToModel(accountCategory))
	})

	// BULK DELETE (DELETE) - ADD FOOTSTEP
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/account-category/bulk-delete",
		Method:      "DELETE",
		Note:        "Bulk delete multiple account categories by IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account categories (/account-category/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.core.AccountCategoryManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account categories (/account-category/bulk-delete) | error: " + err.Error(),
				Module:      "AccountCategory",
			})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted account categories (/account-category/bulk-delete)",
			Module:      "AccountCategory",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
