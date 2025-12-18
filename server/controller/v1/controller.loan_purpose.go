package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) loanPurposeController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-purpose",
		Method:       "GET",
		ResponseType: core.LoanPurposeResponse{},
		Note:         "Returns all loan purposes for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		purposes, err := c.core.LoanPurposeCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No loan purpose records found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanPurposeManager.ToModels(purposes))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-purpose/search",
		Method:       "GET",
		ResponseType: core.LoanPurposeResponse{},
		Note:         "Returns a paginated list of loan purposes for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value, err := c.core.LoanPurposeManager.NormalPagination(context, ctx, &core.LoanPurpose{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch loan purpose records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-purpose/:loan_purpose_id",
		Method:       "GET",
		Note:         "Returns a loan purpose record by its ID.",
		ResponseType: core.LoanPurposeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose ID"})
		}
		purpose, err := c.core.LoanPurposeManager.GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan purpose record not found"})
		}
		return ctx.JSON(http.StatusOK, purpose)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-purpose",
		Method:       "POST",
		RequestType:  core.LoanPurposeRequest{},
		ResponseType: core.LoanPurposeResponse{},
		Note:         "Creates a new loan purpose record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.LoanPurposeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan purpose creation failed (/loan-purpose), validation error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan purpose creation failed (/loan-purpose), user org error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan purpose creation failed (/loan-purpose), user not assigned to branch.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		purpose := &core.LoanPurpose{
			Description:    req.Description,
			Icon:           req.Icon,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}
		if err := c.core.LoanPurposeManager.Create(context, purpose); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan purpose creation failed (/loan-purpose), db error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan purpose record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created loan purpose (/loan-purpose): " + purpose.Description,
			Module:      "LoanPurpose",
		})
		return ctx.JSON(http.StatusCreated, c.core.LoanPurposeManager.ToModel(purpose))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/loan-purpose/:loan_purpose_id",
		Method:       "PUT",
		RequestType:  core.LoanPurposeRequest{},
		ResponseType: core.LoanPurposeResponse{},
		Note:         "Updates an existing loan purpose record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), invalid loan purpose ID.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose ID"})
		}
		req, err := c.core.LoanPurposeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), validation error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), user org error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), user not assigned to branch.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		purpose, err := c.core.LoanPurposeManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), not found.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan purpose record not found"})
		}
		purpose.Description = req.Description
		purpose.Icon = req.Icon
		purpose.UpdatedAt = time.Now().UTC()
		purpose.UpdatedByID = userOrg.UserID
		if err := c.core.LoanPurposeManager.UpdateByID(context, purpose.ID, purpose); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), db error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan purpose record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated loan purpose (/loan-purpose/:loan_purpose_id): " + purpose.Description,
			Module:      "LoanPurpose",
		})
		return ctx.JSON(http.StatusOK, c.core.LoanPurposeManager.ToModel(purpose))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/loan-purpose/:loan_purpose_id",
		Method: "DELETE",
		Note:   "Deletes the specified loan purpose record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan purpose delete failed (/loan-purpose/:loan_purpose_id), invalid loan purpose ID.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose ID"})
		}
		purpose, err := c.core.LoanPurposeManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan purpose delete failed (/loan-purpose/:loan_purpose_id), not found.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan purpose record not found"})
		}
		if err := c.core.LoanPurposeManager.Delete(context, *id); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan purpose delete failed (/loan-purpose/:loan_purpose_id), db error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan purpose record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted loan purpose (/loan-purpose/:loan_purpose_id): " + purpose.Description,
			Module:      "LoanPurpose",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/loan-purpose/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple loan purpose records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan purpose bulk delete failed (/loan-purpose/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan purpose bulk delete failed (/loan-purpose/bulk-delete) | no IDs provided",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := c.core.LoanPurposeManager.BulkDelete(context, ids); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan purpose bulk delete failed (/loan-purpose/bulk-delete) | error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete loan purpose records: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted loan purposes (/loan-purpose/bulk-delete)",
			Module:      "LoanPurpose",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
