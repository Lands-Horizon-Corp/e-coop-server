package settings

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func CompanyController(service *horizon.HorizonService) {
	

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/company",
		Method:       "GET",
		Note:         "Returns all companies for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: types.CompanyResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		companies, err := core.CompanyCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No companies found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.CompanyManager(service).ToModels(companies))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/company/search",
		Method:       "GET",
		Note:         "Returns a paginated list of companies for the current user's organization and branch.",
		ResponseType: types.CompanyResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		companies, err := core.CompanyManager(service).NormalPagination(context, ctx, &types.Company{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch companies for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, companies)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/company/:company_id",
		Method:       "GET",
		Note:         "Returns a single company by its ID.",
		ResponseType: types.CompanyResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		companyID, err := helpers.EngineUUIDParam(ctx, "company_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid company ID"})
		}
		company, err := core.CompanyManager(service).GetByIDRaw(context, *companyID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Company not found"})
		}
		return ctx.JSON(http.StatusOK, company)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/company",
		Method:       "POST",
		Note:         "Creates a new company for the current user's organization and branch.",
		RequestType:  types.CompanyRequest{},
		ResponseType: types.CompanyResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.CompanyManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Company creation failed (/company), validation error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid company data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Company creation failed (/company), user org error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Company creation failed (/company), user not assigned to branch.",
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		company := &types.Company{
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

		if err := core.CompanyManager(service).Create(context, company); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Company creation failed (/company), db error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create company: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created company (/company): " + company.Name,
			Module:      "Company",
		})
		return ctx.JSON(http.StatusCreated, core.CompanyManager(service).ToModel(company))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/company/:company_id",
		Method:       "PUT",
		Note:         "Updates an existing company by its ID.",
		RequestType:  types.CompanyRequest{},
		ResponseType: types.CompanyResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		companyID, err := helpers.EngineUUIDParam(ctx, "company_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Company update failed (/company/:company_id), invalid company ID.",
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid company ID"})
		}

		req, err := core.CompanyManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Company update failed (/company/:company_id), validation error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid company data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Company update failed (/company/:company_id), user org error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		company, err := core.CompanyManager(service).GetByID(context, *companyID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.CompanyManager(service).UpdateByID(context, company.ID, company); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Company update failed (/company/:company_id), db error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update company: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated company (/company/:company_id): " + company.Name,
			Module:      "Company",
		})
		return ctx.JSON(http.StatusOK, core.CompanyManager(service).ToModel(company))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/company/:company_id",
		Method: "DELETE",
		Note:   "Deletes the specified company by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		companyID, err := helpers.EngineUUIDParam(ctx, "company_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Company delete failed (/company/:company_id), invalid company ID.",
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid company ID"})
		}
		company, err := core.CompanyManager(service).GetByID(context, *companyID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Company delete failed (/company/:company_id), not found.",
				Module:      "Company",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Company not found"})
		}
		if err := core.CompanyManager(service).Delete(context, *companyID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Company delete failed (/company/:company_id), db error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete company: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted company (/company/:company_id): " + company.Name,
			Module:      "Company",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/company/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple companies by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/company/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/company/bulk-delete) | no IDs provided",
				Module:      "Company",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No company IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.CompanyManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/company/bulk-delete) | error: " + err.Error(),
				Module:      "Company",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete companies: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted companies (/company/bulk-delete)",
			Module:      "Company",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
