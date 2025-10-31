package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelCore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// LoanPurposeController manages endpoints for loan purpose records.
func (c *Controller) LoanPurposeController() {
	req := c.provider.Service.Request

	// GET /loan-purpose: List all loan purposes for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-purpose",
		Method:       "GET",
		ResponseType: modelCore.LoanPurposeResponse{},
		Note:         "Returns all loan purposes for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		purposes, err := c.modelCore.LoanPurposeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No loan purpose records found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.LoanPurposeManager.Filtered(context, ctx, purposes))
	})

	// GET /loan-purpose/search: Paginated search of loan purposes for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-purpose/search",
		Method:       "GET",
		ResponseType: modelCore.LoanPurposeResponse{},
		Note:         "Returns a paginated list of loan purposes for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value, err := c.modelCore.LoanPurposeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch loan purpose records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.LoanPurposeManager.Pagination(context, ctx, value))
	})

	// GET /loan-purpose/:loan_purpose_id: Get a specific loan purpose record by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-purpose/:loan_purpose_id",
		Method:       "GET",
		Note:         "Returns a loan purpose record by its ID.",
		ResponseType: modelCore.LoanPurposeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose ID"})
		}
		purpose, err := c.modelCore.LoanPurposeManager.GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan purpose record not found"})
		}
		return ctx.JSON(http.StatusOK, purpose)
	})

	// POST /loan-purpose: Create a new loan purpose record. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-purpose",
		Method:       "POST",
		RequestType:  modelCore.LoanPurposeRequest{},
		ResponseType: modelCore.LoanPurposeResponse{},
		Note:         "Creates a new loan purpose record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelCore.LoanPurposeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan purpose creation failed (/loan-purpose), validation error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan purpose creation failed (/loan-purpose), user org error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan purpose creation failed (/loan-purpose), user not assigned to branch.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		purpose := &modelCore.LoanPurpose{
			Description:    req.Description,
			Icon:           req.Icon,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}
		if err := c.modelCore.LoanPurposeManager.Create(context, purpose); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan purpose creation failed (/loan-purpose), db error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan purpose record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created loan purpose (/loan-purpose): " + purpose.Description,
			Module:      "LoanPurpose",
		})
		return ctx.JSON(http.StatusCreated, c.modelCore.LoanPurposeManager.ToModel(purpose))
	})

	// PUT /loan-purpose/:loan_purpose_id: Update a loan purpose record by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-purpose/:loan_purpose_id",
		Method:       "PUT",
		RequestType:  modelCore.LoanPurposeRequest{},
		ResponseType: modelCore.LoanPurposeResponse{},
		Note:         "Updates an existing loan purpose record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), invalid loan purpose ID.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose ID"})
		}
		req, err := c.modelCore.LoanPurposeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), validation error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), user org error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), user not assigned to branch.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		purpose, err := c.modelCore.LoanPurposeManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), not found.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan purpose record not found"})
		}
		purpose.Description = req.Description
		purpose.Icon = req.Icon
		purpose.UpdatedAt = time.Now().UTC()
		purpose.UpdatedByID = user.UserID
		if err := c.modelCore.LoanPurposeManager.UpdateFields(context, purpose.ID, purpose); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), db error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan purpose record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated loan purpose (/loan-purpose/:loan_purpose_id): " + purpose.Description,
			Module:      "LoanPurpose",
		})
		return ctx.JSON(http.StatusOK, c.modelCore.LoanPurposeManager.ToModel(purpose))
	})

	// DELETE /loan-purpose/:loan_purpose_id: Delete a loan purpose record by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/loan-purpose/:loan_purpose_id",
		Method: "DELETE",
		Note:   "Deletes the specified loan purpose record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan purpose delete failed (/loan-purpose/:loan_purpose_id), invalid loan purpose ID.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose ID"})
		}
		purpose, err := c.modelCore.LoanPurposeManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan purpose delete failed (/loan-purpose/:loan_purpose_id), not found.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan purpose record not found"})
		}
		if err := c.modelCore.LoanPurposeManager.DeleteByID(context, *id); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan purpose delete failed (/loan-purpose/:loan_purpose_id), db error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan purpose record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted loan purpose (/loan-purpose/:loan_purpose_id): " + purpose.Description,
			Module:      "LoanPurpose",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /loan-purpose/bulk-delete: Bulk delete loan purpose records by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/loan-purpose/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple loan purpose records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: modelCore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelCore.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan purpose bulk delete failed (/loan-purpose/bulk-delete), invalid request body.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan purpose bulk delete failed (/loan-purpose/bulk-delete), no IDs provided.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan purpose bulk delete failed (/loan-purpose/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		descriptions := ""
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Loan purpose bulk delete failed (/loan-purpose/bulk-delete), invalid UUID: " + rawID,
					Module:      "LoanPurpose",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			purpose, err := c.modelCore.LoanPurposeManager.GetByID(context, id)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Loan purpose bulk delete failed (/loan-purpose/bulk-delete), not found: " + rawID,
					Module:      "LoanPurpose",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Loan purpose record not found with ID: %s", rawID)})
			}
			descriptions += purpose.Description + ","
			if err := c.modelCore.LoanPurposeManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Loan purpose bulk delete failed (/loan-purpose/bulk-delete), db error: " + err.Error(),
					Module:      "LoanPurpose",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan purpose record: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan purpose bulk delete failed (/loan-purpose/bulk-delete), commit error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted loan purposes (/loan-purpose/bulk-delete): " + descriptions,
			Module:      "LoanPurpose",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
