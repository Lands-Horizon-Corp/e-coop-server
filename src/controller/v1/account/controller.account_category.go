package account

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func AccountCategoryController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-category/search",
		Method:       "GET",
		Note:         "Retrieve all account categories for the current branch.",
		ResponseType: types.AccountCategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}

		result, err := core.AccountCategoryManager(service).NormalPagination(context, ctx, &types.AccountCategory{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account categories: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, result)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-category",
		Method:       "GET",
		Note:         "Retrieve all account categories for the current branch (raw).",
		ResponseType: types.AccountCategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		categories, err := core.AccountCategoryManager(service).Find(context, &types.AccountCategory{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account categories (raw): " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.AccountCategoryManager(service).ToModels(categories))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-category/:account_category_id",
		Method:       "GET",
		Note:         "Get an account category by ID.",
		ResponseType: types.AccountCategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "account_category_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account category ID: " + err.Error()})
		}
		category, err := core.AccountCategoryManager(service).GetByID(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account category not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.AccountCategoryManager(service).ToModel(category))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-category",
		Method:       "POST",
		Note:         "Create a new account category for the current branch.",
		ResponseType: types.AccountCategoryResponse{},
		RequestType: types.AccountCategoryRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.AccountCategoryManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed create account category (/account-category) | validation error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account category validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed create account category (/account-category) | user org error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create account category attempt (/account-category)",
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}

		accountCategory := &types.AccountCategory{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			Name:           req.Name,
			Description:    req.Description,
		}

		if err := core.AccountCategoryManager(service).Create(context, accountCategory); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed create account category (/account-category) | db error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create account category: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created account category (/account-category): " + accountCategory.Name,
			Module:      "AccountCategory",
		})
		return ctx.JSON(http.StatusCreated, core.AccountCategoryManager(service).ToModel(accountCategory))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-category/:account_category_id",
		Method:       "PUT",
		Note:         "Update an account category by ID.",
		ResponseType: types.AccountCategoryResponse{},
		RequestType: types.AccountCategoryRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.AccountCategoryManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed update account category (/account-category/:account_category_id) | validation error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account category validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed update account category (/account-category/:account_category_id) | user org error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update account category attempt (/account-category/:account_category_id)",
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accountCategoryID, err := helpers.EngineUUIDParam(ctx, "account_category_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed update account category (/account-category/:account_category_id) | invalid UUID: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account category ID: " + err.Error()})
		}
		accountCategory, err := core.AccountCategoryManager(service).GetByID(context, *types.AccountCategoryID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.AccountCategoryManager(service).UpdateByID(context, accountCategory.ID, accountCategory); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed update account category (/account-category/:account_category_id) | db error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account category: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated account category (/account-category/:account_category_id): " + accountCategory.Name,
			Module:      "AccountCategory",
		})
		return ctx.JSON(http.StatusOK, core.AccountCategoryManager(service).ToModel(accountCategory))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/account-category/:account_category_id",
		Method: "DELETE",
		Note:   "Delete an account category by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed delete account category (/account-category/:account_category_id) | user org error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete account category attempt (/account-category/:account_category_id)",
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accountCategoryID, err := helpers.EngineUUIDParam(ctx, "account_category_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed delete account category (/account-category/:account_category_id) | invalid UUID: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account category ID: " + err.Error()})
		}
		accountCategory, err := core.AccountCategoryManager(service).GetByID(context, *types.AccountCategoryID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed delete account category (/account-category/:account_category_id) | not found: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account category not found: " + err.Error()})
		}
		if err := core.AccountCategoryManager(service).Delete(context, accountCategory.ID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed delete account category (/account-category/:account_category_id) | db error: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account category: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted account category (/account-category/:account_category_id): " + accountCategory.Name,
			Module:      "AccountCategory",
		})
		return ctx.JSON(http.StatusOK, core.AccountCategoryManager(service).ToModel(accountCategory))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/account-category/bulk-delete",
		Method:      "DELETE",
		Note:        "Bulk delete multiple account categories by IDs.",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account categories (/account-category/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "AccountCategory",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}

		if err := core.AccountCategoryManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account categories (/account-category/bulk-delete) | error: " + err.Error(),
				Module:      "AccountCategory",
			})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted account categories (/account-category/bulk-delete)",
			Module:      "AccountCategory",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
