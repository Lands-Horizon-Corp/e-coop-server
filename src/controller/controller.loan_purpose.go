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

// LoanPurposeController manages endpoints for loan purpose records.
func (c *Controller) LoanPurposeController() {
	req := c.provider.Service.Request

	// GET /loan-purpose: List all loan purposes for the current user's branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/loan-purpose",
		Method:   "GET",
		Response: "TLoanPurpose[]",
		Note:     "Returns all loan purposes for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		purposes, err := c.model.LoanPurposeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No loan purpose records found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.model.LoanPurposeManager.ToModels(purposes))
	})

	// GET /loan-purpose/search: Paginated search of loan purposes for the current branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/loan-purpose/search",
		Method:   "GET",
		Request:  "Filter<ILoanPurpose>",
		Response: "Paginated<ILoanPurpose>",
		Note:     "Returns a paginated list of loan purposes for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value, err := c.model.LoanPurposeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch loan purpose records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.LoanPurposeManager.Pagination(context, ctx, value))
	})

	// GET /loan-purpose/:loan_purpose_id: Get a specific loan purpose record by ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/loan-purpose/:loan_purpose_id",
		Method:   "GET",
		Response: "TLoanPurpose",
		Note:     "Returns a loan purpose record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose ID"})
		}
		purpose, err := c.model.LoanPurposeManager.GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan purpose record not found"})
		}
		return ctx.JSON(http.StatusOK, purpose)
	})

	// POST /loan-purpose: Create a new loan purpose record.
	req.RegisterRoute(horizon.Route{
		Route:    "/loan-purpose",
		Method:   "POST",
		Request:  "TLoanPurpose",
		Response: "TLoanPurpose",
		Note:     "Creates a new loan purpose record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.LoanPurposeManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		purpose := &model.LoanPurpose{
			Description:    req.Description,
			Icon:           req.Icon,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}
		if err := c.model.LoanPurposeManager.Create(context, purpose); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan purpose record: " + err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.model.LoanPurposeManager.ToModel(purpose))
	})

	// PUT /loan-purpose/:loan_purpose_id: Update a loan purpose record by ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/loan-purpose/:loan_purpose_id",
		Method:   "PUT",
		Request:  "TLoanPurpose",
		Response: "TLoanPurpose",
		Note:     "Updates an existing loan purpose record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose ID"})
		}
		req, err := c.model.LoanPurposeManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		purpose, err := c.model.LoanPurposeManager.GetByID(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan purpose record not found"})
		}
		purpose.Description = req.Description
		purpose.Icon = req.Icon
		purpose.UpdatedAt = time.Now().UTC()
		purpose.UpdatedByID = user.UserID
		if err := c.model.LoanPurposeManager.UpdateFields(context, purpose.ID, purpose); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan purpose record: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.LoanPurposeManager.ToModel(purpose))
	})

	// DELETE /loan-purpose/:loan_purpose_id: Delete a loan purpose record by ID.
	req.RegisterRoute(horizon.Route{
		Route:  "/loan-purpose/:loan_purpose_id",
		Method: "DELETE",
		Note:   "Deletes the specified loan purpose record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose ID"})
		}
		if err := c.model.LoanPurposeManager.DeleteByID(context, *id); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan purpose record: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /loan-purpose/bulk-delete: Bulk delete loan purpose records by IDs.
	req.RegisterRoute(horizon.Route{
		Route:   "/loan-purpose/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple loan purpose records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
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
			if _, err := c.model.LoanPurposeManager.GetByID(context, id); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Loan purpose record not found with ID: %s", rawID)})
			}
			if err := c.model.LoanPurposeManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan purpose record: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
