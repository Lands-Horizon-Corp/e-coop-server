package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// LoanTagController registers routes for managing loan tags.
func (c *Controller) loanTagController() {
	req := c.provider.Service.Request

	// GET /loan-tag: List all loan tags for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-tag",
		Method:       "GET",
		Note:         "Returns all loan tags for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.LoanTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTags, err := c.core.LoanTagCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No loan tags found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTagManager.ToModels(loanTags))
	})

	// GET /api/v1/loan-tag/loan-transaction/:loan_transaction_id: List loan tags by loan transaction ID for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-tag/loan-transaction/:loan_transaction_id",
		Method:       "GET",
		Note:         "Returns all loan tags for the specified loan transaction ID within the current user's organization and branch.",
		ResponseType: core.LoanTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTags, err := c.core.LoanTagManager.Find(context, &core.LoanTag{
			LoanTransactionID: loanTransactionID,
			OrganizationID:    user.OrganizationID,
			BranchID:          *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No loan tags found for the specified loan transaction ID in the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTagManager.ToModels(loanTags))
	})

	// GET /loan-tag/search: Paginated search of loan tags for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-tag/search",
		Method:       "GET",
		Note:         "Returns a paginated list of loan tags for the current user's organization and branch.",
		ResponseType: core.LoanTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTags, err := c.core.LoanTagManager.PaginationWithFields(context, ctx, &core.LoanTag{
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch loan tags for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, loanTags)
	})

	// GET /loan-tag/:loan_tag_id: Get specific loan tag by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-tag/:loan_tag_id",
		Method:       "GET",
		Note:         "Returns a single loan tag by its ID.",
		ResponseType: core.LoanTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTagID, err := handlers.EngineUUIDParam(ctx, "loan_tag_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan tag ID"})
		}
		loanTag, err := c.core.LoanTagManager.GetByIDRaw(context, *loanTagID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan tag not found"})
		}
		return ctx.JSON(http.StatusOK, loanTag)
	})

	// POST /loan-tag: Create a new loan tag. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-tag",
		Method:       "POST",
		Note:         "Creates a new loan tag for the current user's organization and branch.",
		RequestType:  core.LoanTagRequest{},
		ResponseType: core.LoanTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.LoanTagManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan tag creation failed (/loan-tag), validation error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan tag creation failed (/loan-tag), user org error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed " + err.Error()})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan tag creation failed (/loan-tag), user not assigned to branch.",
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		loanTag := &core.LoanTag{
			LoanTransactionID: req.LoanTransactionID,
			Name:              req.Name,
			Description:       req.Description,
			Category:          req.Category,
			Color:             req.Color,
			Icon:              req.Icon,
			CreatedAt:         time.Now().UTC(),
			CreatedByID:       user.UserID,
			UpdatedAt:         time.Now().UTC(),
			UpdatedByID:       user.UserID,
			BranchID:          *user.BranchID,
			OrganizationID:    user.OrganizationID,
		}

		if err := c.core.LoanTagManager.Create(context, loanTag); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan tag creation failed (/loan-tag), db error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan tag: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created loan tag (/loan-tag): " + loanTag.Name,
			Module:      "LoanTag",
		})
		return ctx.JSON(http.StatusCreated, c.core.LoanTagManager.ToModel(loanTag))
	})

	// PUT /loan-tag/:loan_tag_id: Update loan tag by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-tag/:loan_tag_id",
		Method:       "PUT",
		Note:         "Updates an existing loan tag by its ID.",
		RequestType:  core.LoanTagRequest{},
		ResponseType: core.LoanTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTagID, err := handlers.EngineUUIDParam(ctx, "loan_tag_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan tag update failed (/loan-tag/:loan_tag_id), invalid loan tag ID.",
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan tag ID"})
		}

		req, err := c.core.LoanTagManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan tag update failed (/loan-tag/:loan_tag_id), validation error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan tag update failed (/loan-tag/:loan_tag_id), user org error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		loanTag, err := c.core.LoanTagManager.GetByID(context, *loanTagID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan tag update failed (/loan-tag/:loan_tag_id), loan tag not found.",
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan tag not found"})
		}
		loanTag.LoanTransactionID = req.LoanTransactionID
		loanTag.Name = req.Name
		loanTag.Description = req.Description
		loanTag.Category = req.Category
		loanTag.Color = req.Color
		loanTag.Icon = req.Icon
		loanTag.UpdatedAt = time.Now().UTC()
		loanTag.UpdatedByID = user.UserID
		if err := c.core.LoanTagManager.UpdateByID(context, loanTag.ID, loanTag); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan tag update failed (/loan-tag/:loan_tag_id), db error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan tag: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated loan tag (/loan-tag/:loan_tag_id): " + loanTag.Name,
			Module:      "LoanTag",
		})
		return ctx.JSON(http.StatusOK, c.core.LoanTagManager.ToModel(loanTag))
	})

	// DELETE /loan-tag/:loan_tag_id: Delete a loan tag by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/loan-tag/:loan_tag_id",
		Method: "DELETE",
		Note:   "Deletes the specified loan tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTagID, err := handlers.EngineUUIDParam(ctx, "loan_tag_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan tag delete failed (/loan-tag/:loan_tag_id), invalid loan tag ID.",
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan tag ID"})
		}
		loanTag, err := c.core.LoanTagManager.GetByID(context, *loanTagID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan tag delete failed (/loan-tag/:loan_tag_id), not found.",
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan tag not found"})
		}
		if err := c.core.LoanTagManager.Delete(context, *loanTagID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan tag delete failed (/loan-tag/:loan_tag_id), db error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan tag: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted loan tag (/loan-tag/:loan_tag_id): " + loanTag.Name,
			Module:      "LoanTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Simplified bulk-delete handler for loan tags (mirrors the feedback/holiday pattern)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/loan-tag/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple loan tags by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-tag/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-tag/bulk-delete) | no IDs provided",
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		if err := c.core.LoanTagManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-tag/bulk-delete) | error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete loan tags: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted loan tags (/loan-tag/bulk-delete)",
			Module:      "LoanTag",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
