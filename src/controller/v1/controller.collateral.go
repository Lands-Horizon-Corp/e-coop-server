package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func collateralController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/collateral",
		Method:       "GET",
		Note:         "Returns all collateral records for the current user's organization and branch. Returns error if not authenticated.",
		ResponseType: core.CollateralResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		collaterals, err := core.CollateralCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No collateral records found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.CollateralManager(service).ToModels(collaterals))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/collateral/search",
		Method:       "GET",
		ResponseType: core.CollateralResponse{},
		Note:         "Returns a paginated list of collateral records for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		collaterals, err := core.CollateralManager(service).NormalPagination(context, ctx, &core.Collateral{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch collateral records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, collaterals)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/collateral/:collateral_id",
		Method:       "GET",
		Note:         "Returns a collateral record by its ID.",
		ResponseType: core.CollateralResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		collateralID, err := helpers.EngineUUIDParam(ctx, "collateral_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid collateral ID"})
		}
		collateral, err := core.CollateralManager(service).GetByIDRaw(context, *collateralID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Collateral record not found"})
		}
		return ctx.JSON(http.StatusOK, collateral)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/collateral",
		Method:       "POST",
		RequestType:  core.CollateralRequest{},
		ResponseType: core.CollateralResponse{},
		Note:         "Creates a new collateral record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.CollateralManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Collateral creation failed (/collateral), validation error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid collateral data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Collateral creation failed (/collateral), user org error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Collateral creation failed (/collateral), user not assigned to branch.",
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		collateral := &core.Collateral{
			Icon:           req.Icon,
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.CollateralManager(service).Create(context, collateral); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Collateral creation failed (/collateral), db error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create collateral record: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created collateral (/collateral): " + collateral.Name,
			Module:      "Collateral",
		})
		return ctx.JSON(http.StatusCreated, core.CollateralManager(service).ToModel(collateral))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/collateral/:collateral_id",
		Method:       "PUT",
		RequestType:  core.CollateralRequest{},
		ResponseType: core.CollateralResponse{},
		Note:         "Updates an existing collateral record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		collateralID, err := helpers.EngineUUIDParam(ctx, "collateral_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Collateral update failed (/collateral/:collateral_id), invalid collateral ID.",
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid collateral ID"})
		}

		req, err := core.CollateralManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Collateral update failed (/collateral/:collateral_id), validation error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid collateral data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Collateral update failed (/collateral/:collateral_id), user org error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		collateral, err := core.CollateralManager(service).GetByID(context, *collateralID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Collateral update failed (/collateral/:collateral_id), record not found.",
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Collateral record not found"})
		}
		collateral.Icon = req.Icon
		collateral.Name = req.Name
		collateral.Description = req.Description
		collateral.UpdatedAt = time.Now().UTC()
		collateral.UpdatedByID = userOrg.UserID
		if err := core.CollateralManager(service).UpdateByID(context, collateral.ID, collateral); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Collateral update failed (/collateral/:collateral_id), db error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update collateral record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated collateral (/collateral/:collateral_id): " + collateral.Name,
			Module:      "Collateral",
		})
		return ctx.JSON(http.StatusOK, core.CollateralManager(service).ToModel(collateral))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/collateral/:collateral_id",
		Method: "DELETE",
		Note:   "Deletes the specified collateral record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		collateralID, err := helpers.EngineUUIDParam(ctx, "collateral_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Collateral delete failed (/collateral/:collateral_id), invalid collateral ID.",
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid collateral ID"})
		}
		collateral, err := core.CollateralManager(service).GetByID(context, *collateralID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Collateral delete failed (/collateral/:collateral_id), record not found.",
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Collateral record not found"})
		}
		if err := core.CollateralManager(service).Delete(context, *collateralID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Collateral delete failed (/collateral/:collateral_id), db error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete collateral record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted collateral (/collateral/:collateral_id): " + collateral.Name,
			Module:      "Collateral",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/collateral/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple collateral records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Collateral bulk delete failed (/collateral/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Collateral bulk delete failed (/collateral/bulk-delete) | no IDs provided",
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.CollateralManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Collateral bulk delete failed (/collateral/bulk-delete) | error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete collateral records: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted collaterals (/collateral/bulk-delete)",
			Module:      "Collateral",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
