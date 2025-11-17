package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// CompanyController registers routes for managing companies.
func (c *Controller) companyController() {
	req := c.provider.Service.Request

	// GET /company: List all companies for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/company",
		Method:       "GET",
		Note:         "Returns all companies for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.CompanyResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		companies, err := c.core.CompanyCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No companies found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.CompanyManager.ToModels(companies))
	})

	// GET /company/search: Paginated search of companies for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/company/search",
		Method:       "GET",
		Note:         "Returns a paginated list of companies for the current user's organization and branch.",
		ResponseType: core.CompanyResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		companies, err := c.core.CompanyManager.PaginationWithFields(context, ctx, &core.Company{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch companies for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, companies)
	})

	// GET /company/:company_id: Get specific company by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/company/:company_id",
		Method:       "GET",
		Note:         "Returns a single company by its ID.",
		ResponseType: core.CompanyResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		companyID, err := handlers.EngineUUIDParam(ctx, "company_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid company ID"})
		}
		company, err := c.core.CompanyManager.GetByIDRaw(context, *companyID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Company not found"})
		}
		return ctx.JSON(http.StatusOK, company)
	})

	// POST /company: Create a new company. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/company",
		Method:       "POST",
		Note:         "Creates a new company for the current user's organization and branch.",
		RequestType:  core.CompanyRequest{},
		ResponseType: core.CompanyResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.CompanyManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Company creation failed (/company), validation error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid company data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Company creation failed (/company), user org error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Company creation failed (/company), user not assigned to branch.",
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		company := &core.Company{
			MediaID:        req.MediaID,
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := c.core.CompanyManager.Create(context, company); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Company creation failed (/company), db error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create company: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created company (/company): " + company.Name,
			Module:      "Company",
		})
		return ctx.JSON(http.StatusCreated, c.core.CompanyManager.ToModel(company))
	})

	// PUT /company/:company_id: Update company by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/company/:company_id",
		Method:       "PUT",
		Note:         "Updates an existing company by its ID.",
		RequestType:  core.CompanyRequest{},
		ResponseType: core.CompanyResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		companyID, err := handlers.EngineUUIDParam(ctx, "company_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Company update failed (/company/:company_id), invalid company ID.",
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid company ID"})
		}

		req, err := c.core.CompanyManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Company update failed (/company/:company_id), validation error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid company data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Company update failed (/company/:company_id), user org error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		company, err := c.core.CompanyManager.GetByID(context, *companyID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Company update failed (/company/:company_id), company not found.",
				Module:      "Company",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Company not found"})
		}
		company.MediaID = req.MediaID
		company.Name = req.Name
		company.Description = req.Description
		company.UpdatedAt = time.Now().UTC()
		company.UpdatedByID = userOrg.UserID
		if err := c.core.CompanyManager.UpdateByID(context, company.ID, company); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Company update failed (/company/:company_id), db error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update company: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated company (/company/:company_id): " + company.Name,
			Module:      "Company",
		})
		return ctx.JSON(http.StatusOK, c.core.CompanyManager.ToModel(company))
	})

	// DELETE /company/:company_id: Delete a company by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/company/:company_id",
		Method: "DELETE",
		Note:   "Deletes the specified company by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		companyID, err := handlers.EngineUUIDParam(ctx, "company_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Company delete failed (/company/:company_id), invalid company ID.",
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid company ID"})
		}
		company, err := c.core.CompanyManager.GetByID(context, *companyID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Company delete failed (/company/:company_id), not found.",
				Module:      "Company",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Company not found"})
		}
		if err := c.core.CompanyManager.Delete(context, *companyID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Company delete failed (/company/:company_id), db error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete company: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted company (/company/:company_id): " + company.Name,
			Module:      "Company",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/company/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple companies by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/company/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/company/bulk-delete) | no IDs provided",
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No company IDs provided for bulk delete"})
		}

		if err := c.core.CompanyManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/company/bulk-delete) | error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete companies: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted companies (/company/bulk-delete)",
			Module:      "Company",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
