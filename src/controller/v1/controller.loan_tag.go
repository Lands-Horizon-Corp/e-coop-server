package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func loanTagController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-tag",
		Method:       "GET",
		Note:         "Returns all loan tags for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.LoanTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTags, err := c.core.LoanTagCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No loan tags found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTagManager().ToModels(loanTags))
	})

	req.RegisterWebRoute(handlers.Route{
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
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTags, err := c.core.LoanTagManager().Find(context, &core.LoanTag{
			LoanTransactionID: loanTransactionID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No loan tags found for the specified loan transaction ID in the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTagManager().ToModels(loanTags))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-tag/search",
		Method:       "GET",
		Note:         "Returns a paginated list of loan tags for the current user's organization and branch.",
		ResponseType: core.LoanTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTags, err := c.core.LoanTagManager().NormalPagination(context, ctx, &core.LoanTag{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch loan tags for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, loanTags)
	})

	req.RegisterWebRoute(handlers.Route{
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
		loanTag, err := c.core.LoanTagManager().GetByIDRaw(context, *loanTagID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan tag not found"})
		}
		return ctx.JSON(http.StatusOK, loanTag)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-tag",
		Method:       "POST",
		Note:         "Creates a new loan tag for the current user's organization and branch.",
		RequestType:  core.LoanTagRequest{},
		ResponseType: core.LoanTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.LoanTagManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan tag creation failed (/loan-tag), validation error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan tag data: " + err.Error()})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan tag creation failed (/loan-tag), user org error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed " + err.Error()})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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
			CreatedByID:       userOrg.UserID,
			UpdatedAt:         time.Now().UTC(),
			UpdatedByID:       userOrg.UserID,
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrg.OrganizationID,
		}

		if err := c.core.LoanTagManager().Create(context, loanTag); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan tag creation failed (/loan-tag), db error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan tag: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created loan tag (/loan-tag): " + loanTag.Name,
			Module:      "LoanTag",
		})
		return ctx.JSON(http.StatusCreated, c.core.LoanTagManager().ToModel(loanTag))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-tag/:loan_tag_id",
		Method:       "PUT",
		Note:         "Updates an existing loan tag by its ID.",
		RequestType:  core.LoanTagRequest{},
		ResponseType: core.LoanTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTagID, err := handlers.EngineUUIDParam(ctx, "loan_tag_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan tag update failed (/loan-tag/:loan_tag_id), invalid loan tag ID.",
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan tag ID"})
		}

		req, err := c.core.LoanTagManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan tag update failed (/loan-tag/:loan_tag_id), validation error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan tag data: " + err.Error()})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan tag update failed (/loan-tag/:loan_tag_id), user org error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		loanTag, err := c.core.LoanTagManager().GetByID(context, *loanTagID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		loanTag.UpdatedByID = userOrg.UserID
		if err := c.core.LoanTagManager().UpdateByID(context, loanTag.ID, loanTag); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan tag update failed (/loan-tag/:loan_tag_id), db error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan tag: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated loan tag (/loan-tag/:loan_tag_id): " + loanTag.Name,
			Module:      "LoanTag",
		})
		return ctx.JSON(http.StatusOK, c.core.LoanTagManager().ToModel(loanTag))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/loan-tag/:loan_tag_id",
		Method: "DELETE",
		Note:   "Deletes the specified loan tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTagID, err := handlers.EngineUUIDParam(ctx, "loan_tag_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan tag delete failed (/loan-tag/:loan_tag_id), invalid loan tag ID.",
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan tag ID"})
		}
		loanTag, err := c.core.LoanTagManager().GetByID(context, *loanTagID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan tag delete failed (/loan-tag/:loan_tag_id), not found.",
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan tag not found"})
		}
		if err := c.core.LoanTagManager().Delete(context, *loanTagID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan tag delete failed (/loan-tag/:loan_tag_id), db error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan tag: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted loan tag (/loan-tag/:loan_tag_id): " + loanTag.Name,
			Module:      "LoanTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/loan-tag/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple loan tags by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-tag/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-tag/bulk-delete) | no IDs provided",
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := c.core.LoanTagManager().BulkDelete(context, ids); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-tag/bulk-delete) | error: " + err.Error(),
				Module:      "LoanTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete loan tags: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted loan tags (/loan-tag/bulk-delete)",
			Module:      "LoanTag",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
