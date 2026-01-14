package account

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func AccountTagController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-tag",
		Method:       "GET",
		Note:         "Returns all account tags for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.AccountTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		accountTags, err := core.AccountTagCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account tags not found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.AccountTagManager(service).ToModels(accountTags))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-tag/search",
		Method:       "GET",
		Note:         "Returns a paginated list of account tags for the current user's organization and branch.",
		ResponseType: core.AccountTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		accountTags, err := core.AccountTagManager(service).NormalPagination(context, ctx, &core.AccountTag{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch account tags for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accountTags)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-tag/:account_tag_id",
		Method:       "GET",
		Note:         "Returns a single account tag by its ID.",
		ResponseType: core.AccountTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountTagID, err := helpers.EngineUUIDParam(ctx, "account_tag_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag ID"})
		}
		accountTag, err := core.AccountTagManager(service).GetByIDRaw(context, *accountTagID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account tag not found"})
		}
		return ctx.JSON(http.StatusOK, accountTag)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-tag/account/:account_id",
		Method:       "GET",
		Note:         "Returns all account tags for a specific account ID within the user's organization and branch.",
		ResponseType: core.AccountTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		accountTags, err := core.AccountTagManager(service).Find(context, &core.AccountTag{
			AccountID:      *accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch account tags: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accountTags)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-tag",
		Method:       "POST",
		Note:         "Creates a new account tag for the user's organization and branch.",
		ResponseType: core.AccountTagResponse{},
		RequestType:  core.AccountTagRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.AccountTagManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to validate data for POST /account-tag: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: "User organization not found or authentication failed for POST /account-tag",
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: "User is not assigned to a branch for POST /account-tag",
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		accountTag := &core.AccountTag{
			AccountID:      req.AccountID,
			Name:           req.Name,
			Description:    req.Description,
			Category:       req.Category,
			Color:          req.Color,
			Icon:           req.Icon,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.AccountTagManager(service).Create(context, accountTag); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to create account tag on POST /account-tag: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create account tag: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create success",
			Description: fmt.Sprintf("Created account tag: %s (ID: %s)", accountTag.Name, accountTag.ID),
			Module:      "account-tag",
		})

		return ctx.JSON(http.StatusOK, core.AccountTagManager(service).ToModel(accountTag))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-tag/:account_tag_id",
		Method:       "PUT",
		Note:         "Updates an existing account tag by its ID.",
		ResponseType: core.AccountTagResponse{},
		RequestType:  core.AccountTagRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountTagID, err := helpers.EngineUUIDParam(ctx, "account_tag_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Invalid account tag ID for PUT /account-tag/:account_tag_id: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag ID"})
		}
		req, err := core.AccountTagManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to validate data for PUT /account-tag/:account_tag_id: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: "User organization not found or authentication failed for PUT /account-tag/:account_tag_id",
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		accountTag, err := core.AccountTagManager(service).GetByID(context, *accountTagID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Account tag not found for PUT /account-tag/:account_tag_id: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account tag not found"})
		}
		accountTag.AccountID = req.AccountID
		accountTag.Name = req.Name
		accountTag.Description = req.Description
		accountTag.Category = req.Category
		accountTag.Color = req.Color
		accountTag.Icon = req.Icon
		accountTag.UpdatedAt = time.Now().UTC()
		accountTag.UpdatedByID = userOrg.UserID

		if err := core.AccountTagManager(service).UpdateByID(context, accountTag.ID, accountTag); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to update account tag for PUT /account-tag/:account_tag_id: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account tag: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update success",
			Description: fmt.Sprintf("Updated account tag: %s (ID: %s)", accountTag.Name, accountTag.ID),
			Module:      "account-tag",
		})

		return ctx.JSON(http.StatusOK, core.AccountTagManager(service).ToModel(accountTag))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/account-tag/:account_tag_id",
		Method: "DELETE",
		Note:   "Deletes the specified account tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountTagID, err := helpers.EngineUUIDParam(ctx, "account_tag_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Invalid account tag ID for DELETE /account-tag/:account_tag_id: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag ID"})
		}
		if err := core.AccountTagManager(service).Delete(context, *accountTagID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to delete account tag for DELETE /account-tag/:account_tag_id: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account tag: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete success",
			Description: fmt.Sprintf("Deleted account tag ID: %s", accountTagID.String()),
			Module:      "account-tag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/account-tag/bulk-delete",
		Method:      "DELETE",
		Note:        "Bulk delete multiple account tags by IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account tags (/account-tag/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "AccountTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account tags (/account-tag/bulk-delete) | no IDs provided",
				Module:      "AccountTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided."})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.AccountTagManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account tags (/account-tag/bulk-delete) | error: " + err.Error(),
				Module:      "AccountTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete account tags: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted account tags (/account-tag/bulk-delete)",
			Module:      "AccountTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
