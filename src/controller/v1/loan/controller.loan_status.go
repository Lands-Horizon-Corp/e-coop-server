package loan

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

func LoanStatusController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-status",
		Method:       "GET",
		ResponseType: types.LoanStatusResponse{},
		Note:         "Returns all loan statuses for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		statuses, err := core.LoanStatusCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No loan status records found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.LoanStatusManager(service).ToModels(statuses))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-status/search",
		Method:       "GET",
		ResponseType: types.LoanStatusResponse{},
		Note:         "Returns a paginated list of loan statuses for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value, err := core.LoanStatusManager(service).NormalPagination(context, ctx, &types.LoanStatus{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch loan status records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-status/:loan_status_id",
		Method:       "GET",
		ResponseType: types.LoanStatusResponse{},
		Note:         "Returns a loan status record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status ID"})
		}
		status, err := core.LoanStatusManager(service).GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan status record not found"})
		}
		return ctx.JSON(http.StatusOK, status)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-status",
		Method:       "POST",
		ResponseType: types.LoanStatusResponse{},
		RequestType:  types.LoanStatusRequest{},
		Note:         "Creates a new loan status record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.LoanStatusManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan status creation failed (/loan-status), validation error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan status creation failed (/loan-status), user org error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan status creation failed (/loan-status), user not assigned to branch.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		status := &types.LoanStatus{
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
		if err := core.LoanStatusManager(service).Create(context, status); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan status creation failed (/loan-status), db error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan status record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created loan status (/loan-status): " + status.Name,
			Module:      "LoanStatus",
		})
		return ctx.JSON(http.StatusCreated, core.LoanStatusManager(service).ToModel(status))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-status/:loan_status_id",
		Method:       "PUT",
		ResponseType: types.LoanStatusResponse{},
		RequestType:  types.LoanStatusRequest{},
		Note:         "Updates an existing loan status record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), invalid loan status ID.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status ID"})
		}
		req, err := core.LoanStatusManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), validation error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), user org error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), user not assigned to branch.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		status, err := core.LoanStatusManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.LoanStatusManager(service).UpdateByID(context, status.ID, status); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan status update failed (/loan-status/:loan_status_id), db error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan status record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated loan status (/loan-status/:loan_status_id): " + status.Name,
			Module:      "LoanStatus",
		})
		return ctx.JSON(http.StatusOK, core.LoanStatusManager(service).ToModel(status))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/loan-status/:loan_status_id",
		Method: "DELETE",
		Note:   "Deletes the specified loan status record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan status delete failed (/loan-status/:loan_status_id), invalid loan status ID.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan status ID"})
		}
		status, err := core.LoanStatusManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan status delete failed (/loan-status/:loan_status_id), not found.",
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan status record not found"})
		}
		if err := core.LoanStatusManager(service).Delete(context, *id); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Loan status delete failed (/loan-status/:loan_status_id), db error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan status record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted loan status (/loan-status/:loan_status_id): " + status.Name,
			Module:      "LoanStatus",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/loan-status/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple loan status records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan status bulk delete failed (/loan-status/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.LoanStatusManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Loan status bulk delete failed (/loan-status/bulk-delete) | error: " + err.Error(),
				Module:      "LoanStatus",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete loan status records: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted loan statuses (/loan-status/bulk-delete)",
			Module:      "LoanStatus",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
