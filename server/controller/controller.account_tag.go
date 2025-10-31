package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelcore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelcore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AccountTagController handles routes for managing account tags.
func (c *Controller) AccountTagController() {
	req := c.provider.Service.Request

	// GET /account-tag - List current branch's account tags for the authenticated user.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-tag",
		Method:       "GET",
		Note:         "Returns all account tags for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: modelcore.AccountTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		accountTags, err := c.modelcore.AccountTagCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account tags not found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.AccountTagManager.Filtered(context, ctx, accountTags))
	})

	// GET /account-tag/search - Paginated search of account tags for current branch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-tag/search",
		Method:       "GET",
		Note:         "Returns a paginated list of account tags for the current user's organization and branch.",
		ResponseType: modelcore.AccountTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		accountTags, err := c.modelcore.AccountTagCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch account tags for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.AccountTagManager.Pagination(context, ctx, accountTags))
	})

	// GET /account-tag/:account_tag_id - Get specific account tag by ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-tag/:account_tag_id",
		Method:       "GET",
		Note:         "Returns a single account tag by its ID.",
		ResponseType: modelcore.AccountTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountTagID, err := handlers.EngineUUIDParam(ctx, "account_tag_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag ID"})
		}
		accountTag, err := c.modelcore.AccountTagManager.GetByIDRaw(context, *accountTagID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account tag not found"})
		}
		return ctx.JSON(http.StatusOK, accountTag)
	})

	// "/api/v1/account-tag/account/:account_id",
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-tag/account/:account_id",
		Method:       "GET",
		Note:         "Returns all account tags for a specific account ID within the user's organization and branch.",
		ResponseType: modelcore.AccountTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		accountTags, err := c.modelcore.AccountTagManager.Find(context, &modelcore.AccountTag{
			AccountID:      *accountID,
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch account tags: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accountTags)
	})

	// POST /account-tag - Create new account tag.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-tag",
		Method:       "POST",
		Note:         "Creates a new account tag for the user's organization and branch.",
		ResponseType: modelcore.AccountTagResponse{},
		RequestType:  modelcore.AccountTagRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelcore.AccountTagManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to validate data for POST /account-tag: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: "User organization not found or authentication failed for POST /account-tag",
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: "User is not assigned to a branch for POST /account-tag",
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		accountTag := &modelcore.AccountTag{
			AccountID:      req.AccountID,
			Name:           req.Name,
			Description:    req.Description,
			Category:       req.Category,
			Color:          req.Color,
			Icon:           req.Icon,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.modelcore.AccountTagManager.Create(context, accountTag); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to create account tag on POST /account-tag: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create account tag: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create success",
			Description: fmt.Sprintf("Created account tag: %s (ID: %s)", accountTag.Name, accountTag.ID),
			Module:      "account-tag",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.AccountTagManager.ToModel(accountTag))
	})

	// PUT /account-tag/:account_tag_id - Update account tag by ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-tag/:account_tag_id",
		Method:       "PUT",
		Note:         "Updates an existing account tag by its ID.",
		ResponseType: modelcore.AccountTagResponse{},
		RequestType:  modelcore.AccountTagRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountTagID, err := handlers.EngineUUIDParam(ctx, "account_tag_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Invalid account tag ID for PUT /account-tag/:account_tag_id: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag ID"})
		}
		req, err := c.modelcore.AccountTagManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to validate data for PUT /account-tag/:account_tag_id: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: "User organization not found or authentication failed for PUT /account-tag/:account_tag_id",
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		accountTag, err := c.modelcore.AccountTagManager.GetByID(context, *accountTagID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		accountTag.UpdatedByID = user.UserID

		if err := c.modelcore.AccountTagManager.UpdateFields(context, accountTag.ID, accountTag); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to update account tag for PUT /account-tag/:account_tag_id: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account tag: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update success",
			Description: fmt.Sprintf("Updated account tag: %s (ID: %s)", accountTag.Name, accountTag.ID),
			Module:      "account-tag",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.AccountTagManager.ToModel(accountTag))
	})

	// DELETE /account-tag/:account_tag_id - Delete account tag by ID.
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/account-tag/:account_tag_id",
		Method: "DELETE",
		Note:   "Deletes the specified account tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountTagID, err := handlers.EngineUUIDParam(ctx, "account_tag_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Invalid account tag ID for DELETE /account-tag/:account_tag_id: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag ID"})
		}
		if err := c.modelcore.AccountTagManager.DeleteByID(context, *accountTagID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to delete account tag for DELETE /account-tag/:account_tag_id: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account tag: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete success",
			Description: fmt.Sprintf("Deleted account tag ID: %s", accountTagID.String()),
			Module:      "account-tag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /account-tag/bulk-delete - Bulk delete account tags by IDs.
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/account-tag/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple account tags by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: modelcore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: "Invalid request body for DELETE /account-tag/bulk-delete",
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: "No IDs provided for bulk delete on DELETE /account-tag/bulk-delete",
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to start DB transaction for DELETE /account-tag/bulk-delete: %v", tx.Error),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			accountTagID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete error",
					Description: fmt.Sprintf("Invalid UUID in bulk delete: %s", rawID),
					Module:      "account-tag",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			if _, err := c.modelcore.AccountTagManager.GetByID(context, accountTagID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete error",
					Description: fmt.Sprintf("Account tag not found in bulk delete: %s", rawID),
					Module:      "account-tag",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Account tag not found with ID: %s", rawID)})
			}
			if err := c.modelcore.AccountTagManager.DeleteByIDWithTx(context, tx, accountTagID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete error",
					Description: fmt.Sprintf("Failed to delete account tag in bulk delete: %v", err),
					Module:      "account-tag",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account tag: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to commit bulk delete transaction for DELETE /account-tag/bulk-delete: %v", err),
				Module:      "account-tag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete success",
			Description: fmt.Sprintf("Bulk deleted account tags: %v", reqBody.IDs),
			Module:      "account-tag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
