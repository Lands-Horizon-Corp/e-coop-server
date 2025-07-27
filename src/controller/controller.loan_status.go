package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

// LoanStatusController manages endpoints for loan status records.
func (c *Controller) LoanStatusController() {
	req := c.provider.Service.Request

	// GET /loan-status: List all loan statuses for the current user's branch. (NO footstep)
	req.RegisterRoute(horizon.Route{
		Route:        "/loan-status",
		Method:       "GET",
		ResponseType: model.LoanStatusResponse{},
		Note:         "Returns all loan statuses for the current user's organization and branch.",
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
		return ctx.JSON(http.StatusOK, c.model.LoanStatusManager.Filtered(context, ctx, statuses))
	})

	// GET /loan-status/search: Paginated search of loan statuses for the current branch. (NO footstep)
	req.RegisterRoute(horizon.Route{
		Route:        "/loan-status/search",
		Method:       "GET",
		ResponseType: model.LoanStatusResponse{},
		Note:         "Returns a paginated list of loan statuses for the current user's organization and branch.",
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

	// GET /loan-status/:loan_status_id: Get a specific loan status record by ID. (NO footstep)
	req.RegisterRoute(horizon.Route{
		Route:        "/loan-status/:loan_status_id",
		Method:       "GET",
		ResponseType: model.LoanStatusResponse{},
		Note:         "Returns a loan status record by its ID.",
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

	// POST /loan-status: Create a new loan status record. (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:        "/loan-status",
		Method:       "POST",
		ResponseType: model.LoanStatusResponse{},
		RequestType:  model.LoanStatusRequest{},
		Note:         "Creates a new loan status record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.LoanStatusManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan status creation failed (/loan-status), validation error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan status creation failed (/loan-status), user org error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan status creation failed (/loan-status), user not assigned to branch.",
				Module:      "LoanStatus",
			})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan status creation failed (/loan-status), db error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan status record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created loan status (/loan-status): " + status.Name,
			Module:      "LoanStatus",
		})
		return ctx.JSON(http.StatusCreated, c.model.LoanStatusManager.ToModel(status))
	})

	// PUT /loan-status/:loan_status_id: Update a loan status record by ID. (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:        "/loan-status/:loan_status_id",
		Method:       "PUT",
		ResponseType: model.LoanStatusResponse{},
		RequestType:  model.LoanStatusRequest{},
		Note:         "Updates an existing loan status record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), invalid loan status ID.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status ID"})
		}
		req, err := c.model.LoanStatusManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), validation error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), user org error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), user not assigned to branch.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		status, err := c.model.LoanStatusManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		status.UpdatedByID = user.UserID
		if err := c.model.LoanStatusManager.UpdateFields(context, status.ID, status); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), db error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan status record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated loan status (/loan-status/:loan_status_id): " + status.Name,
			Module:      "LoanStatus",
		})
		return ctx.JSON(http.StatusOK, c.model.LoanStatusManager.ToModel(status))
	})

	// DELETE /loan-status/:loan_status_id: Delete a loan status record by ID. (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:  "/loan-status/:loan_status_id",
		Method: "DELETE",
		Note:   "Deletes the specified loan status record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan status delete failed (/loan-status/:loan_status_id), invalid loan status ID.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status ID"})
		}
		status, err := c.model.LoanStatusManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan status delete failed (/loan-status/:loan_status_id), not found.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan status record not found"})
		}
		if err := c.model.LoanStatusManager.DeleteByID(context, *id); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan status delete failed (/loan-status/:loan_status_id), db error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan status record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted loan status (/loan-status/:loan_status_id): " + status.Name,
			Module:      "LoanStatus",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /loan-status/bulk-delete: Bulk delete loan status records by IDs. (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:       "/loan-status/bulk-delete",
		Method:      "DELETE",
		RequestType: model.IDSRequest{},
		Note:        "Deletes multiple loan status records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan status bulk delete failed (/loan-status/bulk-delete), invalid request body.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan status bulk delete failed (/loan-status/bulk-delete), no IDs provided.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan status bulk delete failed (/loan-status/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Loan status bulk delete failed (/loan-status/bulk-delete), invalid UUID: " + rawID,
					Module:      "LoanStatus",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			status, err := c.model.LoanStatusManager.GetByID(context, id)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Loan status bulk delete failed (/loan-status/bulk-delete), not found: " + rawID,
					Module:      "LoanStatus",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Loan status record not found with ID: %s", rawID)})
			}
			names += status.Name + ","
			if err := c.model.LoanStatusManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Loan status bulk delete failed (/loan-status/bulk-delete), db error: " + err.Error(),
					Module:      "LoanStatus",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan status record: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan status bulk delete failed (/loan-status/bulk-delete), commit error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted loan statuses (/loan-status/bulk-delete): " + names,
			Module:      "LoanStatus",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
