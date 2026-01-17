package loan

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func LoanPurposeController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-purpose",
		Method:       "GET",
		ResponseType: types.LoanPurposeResponse{},
		Note:         "Returns all loan purposes for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		purposes, err := core.LoanPurposeCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No loan purpose records found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.LoanPurposeManager(service).ToModels(purposes))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-purpose/search",
		Method:       "GET",
		ResponseType: types.LoanPurposeResponse{},
		Note:         "Returns a paginated list of loan purposes for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value, err := core.LoanPurposeManager(service).NormalPagination(context, ctx, &types.LoanPurpose{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch loan purpose records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-purpose/:loan_purpose_id",
		Method:       "GET",
		Note:         "Returns a loan purpose record by its ID.",
		ResponseType: types.LoanPurposeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose ID"})
		}
		purpose, err := core.LoanPurposeManager(service).GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan purpose record not found"})
		}
		return ctx.JSON(http.StatusOK, purpose)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-purpose",
		Method:       "POST",
		RequestType: types.LoanPurposeRequest{},
		ResponseType: types.LoanPurposeResponse{},
		Note:         "Creates a new loan purpose record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.LoanPurposeManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan purpose creation failed (/loan-purpose), validation error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan purpose creation failed (/loan-purpose), user org error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan purpose creation failed (/loan-purpose), user not assigned to branch.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		purpose := &types.LoanPurpose{
			Description:    req.Description,
			Icon:           req.Icon,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}
		if err := core.LoanPurposeManager(service).Create(context, purpose); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan purpose creation failed (/loan-purpose), db error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan purpose record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created loan purpose (/loan-purpose): " + purpose.Description,
			Module:      "LoanPurpose",
		})
		return ctx.JSON(http.StatusCreated, core.LoanPurposeManager(service).ToModel(purpose))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-purpose/:loan_purpose_id",
		Method:       "PUT",
		RequestType: types.LoanPurposeRequest{},
		ResponseType: types.LoanPurposeResponse{},
		Note:         "Updates an existing loan purpose record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), invalid loan purpose ID.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose ID"})
		}
		req, err := core.LoanPurposeManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), validation error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), user org error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), user not assigned to branch.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		purpose, err := core.LoanPurposeManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.LoanPurposeManager(service).UpdateByID(context, purpose.ID, purpose); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan purpose update failed (/loan-purpose/:loan_purpose_id), db error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan purpose record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated loan purpose (/loan-purpose/:loan_purpose_id): " + purpose.Description,
			Module:      "LoanPurpose",
		})
		return ctx.JSON(http.StatusOK, core.LoanPurposeManager(service).ToModel(purpose))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/loan-purpose/:loan_purpose_id",
		Method: "DELETE",
		Note:   "Deletes the specified loan purpose record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan purpose delete failed (/loan-purpose/:loan_purpose_id), invalid loan purpose ID.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan purpose ID"})
		}
		purpose, err := core.LoanPurposeManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan purpose delete failed (/loan-purpose/:loan_purpose_id), not found.",
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan purpose record not found"})
		}
		if err := core.LoanPurposeManager(service).Delete(context, *id); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan purpose delete failed (/loan-purpose/:loan_purpose_id), db error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan purpose record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted loan purpose (/loan-purpose/:loan_purpose_id): " + purpose.Description,
			Module:      "LoanPurpose",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/loan-purpose/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple loan purpose records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan purpose bulk delete failed (/loan-purpose/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.LoanPurposeManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan purpose bulk delete failed (/loan-purpose/bulk-delete) | error: " + err.Error(),
				Module:      "LoanPurpose",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete loan purpose records: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted loan purposes (/loan-purpose/bulk-delete)",
			Module:      "LoanPurpose",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
