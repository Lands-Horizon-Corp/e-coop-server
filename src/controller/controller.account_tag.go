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

// AccountTagController handles routes for managing account tags.
func (c *Controller) AccountTagController() {
	req := c.provider.Service.Request

	// GET /account-tag - List current branch's account tags for the authenticated user.
	req.RegisterRoute(horizon.Route{
		Route:    "/account-tag",
		Method:   "GET",
		Response: "TAccountTag[]",
		Note:     "Returns all account tags for the current user's organization and branch. Returns empty if not authenticated.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		accountTags, err := c.model.AccountTagCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account tags not found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountTagManager.ToModels(accountTags))
	})

	// GET /account-tag/search - Paginated search of account tags for current branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/account-tag/search",
		Method:   "GET",
		Request:  "Filter<IAccountTag>",
		Response: "Paginated<IAccountTag>",
		Note:     "Returns a paginated list of account tags for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		accountTags, err := c.model.AccountTagCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch account tags for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountTagManager.Pagination(context, ctx, accountTags))
	})

	// GET /account-tag/:account_tag_id - Get specific account tag by ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/account-tag/:account_tag_id",
		Method:   "GET",
		Response: "TAccountTag",
		Note:     "Returns a single account tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountTagID, err := horizon.EngineUUIDParam(ctx, "account_tag_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag ID"})
		}
		accountTag, err := c.model.AccountTagManager.GetByIDRaw(context, *accountTagID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account tag not found"})
		}
		return ctx.JSON(http.StatusOK, accountTag)
	})

	// POST /account-tag - Create new account tag.
	req.RegisterRoute(horizon.Route{
		Route:    "/account-tag",
		Method:   "POST",
		Request:  "TAccountTag",
		Response: "TAccountTag",
		Note:     "Creates a new account tag for the user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.AccountTagManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		accountTag := &model.AccountTag{
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

		if err := c.model.AccountTagManager.Create(context, accountTag); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create account tag: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.AccountTagManager.ToModel(accountTag))
	})

	// PUT /account-tag/:account_tag_id - Update account tag by ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/account-tag/:account_tag_id",
		Method:   "PUT",
		Request:  "TAccountTag",
		Response: "TAccountTag",
		Note:     "Updates an existing account tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountTagID, err := horizon.EngineUUIDParam(ctx, "account_tag_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag ID"})
		}
		req, err := c.model.AccountTagManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		accountTag, err := c.model.AccountTagManager.GetByID(context, *accountTagID)
		if err != nil {
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

		if err := c.model.AccountTagManager.UpdateFields(context, accountTag.ID, accountTag); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account tag: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountTagManager.ToModel(accountTag))
	})

	// DELETE /account-tag/:account_tag_id - Delete account tag by ID.
	req.RegisterRoute(horizon.Route{
		Route:  "/account-tag/:account_tag_id",
		Method: "DELETE",
		Note:   "Deletes the specified account tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountTagID, err := horizon.EngineUUIDParam(ctx, "account_tag_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account tag ID"})
		}
		if err := c.model.AccountTagManager.DeleteByID(context, *accountTagID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account tag: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /account-tag/bulk-delete - Bulk delete account tags by IDs.
	req.RegisterRoute(horizon.Route{
		Route:   "/account-tag/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple account tags by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
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
			accountTagID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			if _, err := c.model.AccountTagManager.GetByID(context, accountTagID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Account tag not found with ID: %s", rawID)})
			}
			if err := c.model.AccountTagManager.DeleteByIDWithTx(context, tx, accountTagID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account tag: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete transaction: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
