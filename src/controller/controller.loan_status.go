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

// LoanStatusController manages endpoints for loan status records.
func (c *Controller) LoanStatusController() {
	req := c.provider.Service.Request

	// GET /loan-status: List all loan statuses for the current user's branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/loan-status",
		Method:   "GET",
		Response: "TLoanStatus[]",
		Note:     "Returns all loan statuses for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		statuses, err := c.model.LoanStatusCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No loan status records found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.model.LoanStatusManager.ToModels(statuses))
	})

	// GET /loan-status/search: Paginated search of loan statuses for the current branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/loan-status/search",
		Method:   "GET",
		Request:  "Filter<ILoanStatus>",
		Response: "Paginated<ILoanStatus>",
		Note:     "Returns a paginated list of loan statuses for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value, err := c.model.LoanStatusCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch loan status records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.LoanStatusManager.Pagination(context, ctx, value))
	})

	// GET /loan-status/:loan_status_id: Get a specific loan status record by ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/loan-status/:loan_status_id",
		Method:   "GET",
		Response: "TLoanStatus",
		Note:     "Returns a loan status record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status ID"})
		}
		status, err := c.model.LoanStatusManager.GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan status record not found"})
		}
		return ctx.JSON(http.StatusOK, status)
	})

	// POST /loan-status: Create a new loan status record.
	req.RegisterRoute(horizon.Route{
		Route:    "/loan-status",
		Method:   "POST",
		Request:  "TLoanStatus",
		Response: "TLoanStatus",
		Note:     "Creates a new loan status record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.LoanStatusManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		status := &model.LoanStatus{
			Name:           req.Name,
			Icon:           req.Icon,
			Color:          req.Color,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}
		if err := c.model.LoanStatusManager.Create(context, status); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan status record: " + err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.model.LoanStatusManager.ToModel(status))
	})

	// PUT /loan-status/:loan_status_id: Update a loan status record by ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/loan-status/:loan_status_id",
		Method:   "PUT",
		Request:  "TLoanStatus",
		Response: "TLoanStatus",
		Note:     "Updates an existing loan status record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status ID"})
		}
		req, err := c.model.LoanStatusManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		status, err := c.model.LoanStatusManager.GetByID(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan status record not found"})
		}
		status.Name = req.Name
		status.Icon = req.Icon
		status.Color = req.Color
		status.Description = req.Description
		status.UpdatedAt = time.Now().UTC()
		status.UpdatedByID = user.UserID
		if err := c.model.LoanStatusManager.UpdateFields(context, status.ID, status); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan status record: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.LoanStatusManager.ToModel(status))
	})

	// DELETE /loan-status/:loan_status_id: Delete a loan status record by ID.
	req.RegisterRoute(horizon.Route{
		Route:  "/loan-status/:loan_status_id",
		Method: "DELETE",
		Note:   "Deletes the specified loan status record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status ID"})
		}
		if err := c.model.LoanStatusManager.DeleteByID(context, *id); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan status record: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /loan-status/bulk-delete: Bulk delete loan status records by IDs.
	req.RegisterRoute(horizon.Route{
		Route:   "/loan-status/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple loan status records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
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
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			if _, err := c.model.LoanStatusManager.GetByID(context, id); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Loan status record not found with ID: %s", rawID)})
			}
			if err := c.model.LoanStatusManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan status record: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
