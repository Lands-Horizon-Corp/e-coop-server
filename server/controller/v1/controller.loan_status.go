package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// LoanStatusController manages endpoints for loan status records.
func (c *Controller) loanStatusController() {
	req := c.provider.Service.Request

	// GET /loan-status: List all loan statuses for the current user's branch. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-status",
		Method:       "GET",
		ResponseType: core.LoanStatusResponse{},
		Note:         "Returns all loan statuses for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		statuses, err := c.core.LoanStatusCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No loan status records found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanStatusManager.ToModels(statuses))
	})

	// GET /loan-status/search: Paginated search of loan statuses for the current branch. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-status/search",
		Method:       "GET",
		ResponseType: core.LoanStatusResponse{},
		Note:         "Returns a paginated list of loan statuses for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value, err := c.core.LoanStatusManager.NormalPagination(context, ctx, &core.LoanStatus{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch loan status records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	// GET /loan-status/:loan_status_id: Get a specific loan status record by ID. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-status/:loan_status_id",
		Method:       "GET",
		ResponseType: core.LoanStatusResponse{},
		Note:         "Returns a loan status record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status ID"})
		}
		status, err := c.core.LoanStatusManager.GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan status record not found"})
		}
		return ctx.JSON(http.StatusOK, status)
	})

	// POST /loan-status: Create a new loan status record. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-status",
		Method:       "POST",
		ResponseType: core.LoanStatusResponse{},
		RequestType:  core.LoanStatusRequest{},
		Note:         "Creates a new loan status record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.LoanStatusManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan status creation failed (/loan-status), validation error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan status creation failed (/loan-status), user org error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan status creation failed (/loan-status), user not assigned to branch.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		status := &core.LoanStatus{
			Name:           req.Name,
			Icon:           req.Icon,
			Color:          req.Color,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}
		if err := c.core.LoanStatusManager.Create(context, status); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan status creation failed (/loan-status), db error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan status record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created loan status (/loan-status): " + status.Name,
			Module:      "LoanStatus",
		})
		return ctx.JSON(http.StatusCreated, c.core.LoanStatusManager.ToModel(status))
	})

	// PUT /loan-status/:loan_status_id: Update a loan status record by ID. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-status/:loan_status_id",
		Method:       "PUT",
		ResponseType: core.LoanStatusResponse{},
		RequestType:  core.LoanStatusRequest{},
		Note:         "Updates an existing loan status record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), invalid loan status ID.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status ID"})
		}
		req, err := c.core.LoanStatusManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), validation error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), user org error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), user not assigned to branch.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		status, err := c.core.LoanStatusManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), not found.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan status record not found"})
		}
		status.Name = req.Name
		status.Icon = req.Icon
		status.Color = req.Color
		status.Description = req.Description
		status.UpdatedAt = time.Now().UTC()
		status.UpdatedByID = userOrg.UserID
		if err := c.core.LoanStatusManager.UpdateByID(context, status.ID, status); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), db error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan status record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated loan status (/loan-status/:loan_status_id): " + status.Name,
			Module:      "LoanStatus",
		})
		return ctx.JSON(http.StatusOK, c.core.LoanStatusManager.ToModel(status))
	})

	// DELETE /loan-status/:loan_status_id: Delete a loan status record by ID. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/loan-status/:loan_status_id",
		Method: "DELETE",
		Note:   "Deletes the specified loan status record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan status delete failed (/loan-status/:loan_status_id), invalid loan status ID.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status ID"})
		}
		status, err := c.core.LoanStatusManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan status delete failed (/loan-status/:loan_status_id), not found.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan status record not found"})
		}
		if err := c.core.LoanStatusManager.Delete(context, *id); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan status delete failed (/loan-status/:loan_status_id), db error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan status record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted loan status (/loan-status/:loan_status_id): " + status.Name,
			Module:      "LoanStatus",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Simplified bulk-delete handler for loan statuses (mirrors the feedback/holiday pattern)
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/loan-status/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple loan status records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan status bulk delete failed (/loan-status/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan status bulk delete failed (/loan-status/bulk-delete) | no IDs provided",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := c.core.LoanStatusManager.BulkDelete(context, ids); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan status bulk delete failed (/loan-status/bulk-delete) | error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete loan status records: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted loan statuses (/loan-status/bulk-delete)",
			Module:      "LoanStatus",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
