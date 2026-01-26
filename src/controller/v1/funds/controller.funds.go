package funds

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

func FundsController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/funds",
		Method:       "GET",
		ResponseType: types.FundsResponse{},
		Note:         "Returns all funds for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		funds, err := core.FundsCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get funds: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.FundsManager(service).ToModels(funds))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/funds/search",
		Method:       "GET",
		ResponseType: types.FundsResponse{},
		Note:         "Returns paginated funds for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		funds, err := core.FundsManager(service).NormalPagination(context, ctx, &types.Funds{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get funds for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, funds)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/funds",
		Method:       "POST",
		ResponseType: types.FundsResponse{},
		RequestType:  types.FundsRequest{},
		Note:         "Creates a new funds record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.FundsManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create funds failed (/funds), validation error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create funds failed (/funds), user org error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		funds := &types.Funds{
			AccountID:      req.AccountID,
			Type:           req.Type,
			Description:    req.Description,
			Icon:           req.Icon,
			GLBooks:        req.GLBooks,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.FundsManager(service).Create(context, funds); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create funds failed (/funds), db error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create funds: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created funds (/funds): " + funds.Type,
			Module:      "Funds",
		})

		return ctx.JSON(http.StatusOK, core.FundsManager(service).ToModel(funds))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/funds/:funds_id",
		Method:       "PUT",
		ResponseType: types.FundsResponse{},
		RequestType:  types.FundsRequest{},
		Note:         "Updates an existing funds record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fundsID, err := helpers.EngineUUIDParam(ctx, "funds_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), invalid funds_id: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid funds_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), user org error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := core.FundsManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), validation error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		funds, err := core.FundsManager(service).GetByID(context, *fundsID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), not found: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Funds not found: " + err.Error()})
		}
		funds.UpdatedAt = time.Now().UTC()
		funds.UpdatedByID = userOrg.UserID
		funds.OrganizationID = userOrg.OrganizationID
		funds.BranchID = *userOrg.BranchID
		funds.AccountID = req.AccountID
		funds.Type = req.Type
		funds.Description = req.Description
		funds.Icon = req.Icon
		funds.GLBooks = req.GLBooks
		if err := core.FundsManager(service).UpdateByID(context, funds.ID, funds); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), db error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update funds: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated funds (/funds/:funds_id): " + funds.Type,
			Module:      "Funds",
		})
		return ctx.JSON(http.StatusOK, core.FundsManager(service).ToModel(funds))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/funds/:funds_id",
		Method: "DELETE",
		Note:   "Deletes a funds record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fundsID, err := helpers.EngineUUIDParam(ctx, "funds_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete funds failed (/funds/:funds_id), invalid funds_id: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid funds_id: " + err.Error()})
		}
		value, err := core.FundsManager(service).GetByID(context, *fundsID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete funds failed (/funds/:funds_id), record not found: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Funds not found: " + err.Error()})
		}
		if err := core.FundsManager(service).Delete(context, *fundsID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete funds failed (/funds/:funds_id), db error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete funds: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted funds (/funds/:funds_id): " + value.Type,
			Module:      "Funds",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/funds/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple funds records by their IDs.",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete funds failed (/funds/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete funds failed (/funds/bulk-delete) | no IDs provided",
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.FundsManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete funds failed (/funds/bulk-delete) | error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete funds: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted funds (/funds/bulk-delete)",
			Module:      "Funds",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
