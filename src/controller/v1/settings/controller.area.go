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

func AreaController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/area",
		Method:       "GET",
		Note:         "Returns all areas for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: types.AreaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		areas, err := core.AreaCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No areas found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.AreaManager(service).ToModels(areas))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/area/search",
		Method:       "GET",
		Note:         "Returns a paginated list of areas for the current user's organization and branch.",
		ResponseType: types.AreaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		areas, err := core.AreaManager(service).NormalPagination(context, ctx, &types.Area{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch areas for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, areas)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/area/:area_id",
		Method:       "GET",
		Note:         "Returns a single area by its ID.",
		ResponseType: types.AreaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		areaID, err := helpers.EngineUUIDParam(ctx, "area_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid area ID"})
		}
		area, err := core.AreaManager(service).GetByIDRaw(context, *areaID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Area not found"})
		}
		return ctx.JSON(http.StatusOK, area)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/area",
		Method:       "POST",
		Note:         "Creates a new area for the current user's organization and branch.",
		RequestType:  types.AreaRequest{},
		ResponseType: types.AreaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.AreaManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Area creation failed (/area), validation error: " + err.Error(),
				Module:      "Area",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid area data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Area creation failed (/area), user org error: " + err.Error(),
				Module:      "Area",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Area creation failed (/area), user not assigned to branch.",
				Module:      "Area",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		area := &types.Area{
			MediaID:        req.MediaID,
			Name:           req.Name,
			Latitude:       req.Latitude,
			Longitude:      req.Longitude,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.AreaManager(service).Create(context, area); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Area creation failed (/area), db error: " + err.Error(),
				Module:      "Area",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create area: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created area (/area): " + area.Name,
			Module:      "Area",
		})
		return ctx.JSON(http.StatusCreated, core.AreaManager(service).ToModel(area))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/area/:area_id",
		Method:       "PUT",
		Note:         "Updates an existing area by its ID.",
		RequestType:  types.AreaRequest{},
		ResponseType: types.AreaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		areaID, err := helpers.EngineUUIDParam(ctx, "area_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Area update failed (/area/:area_id), invalid area ID.",
				Module:      "Area",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid area ID"})
		}

		req, err := core.AreaManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Area update failed (/area/:area_id), validation error: " + err.Error(),
				Module:      "Area",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid area data: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Area update failed (/area/:area_id), user org error: " + err.Error(),
				Module:      "Area",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		area, err := core.AreaManager(service).GetByID(context, *areaID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Area update failed (/area/:area_id), area not found.",
				Module:      "Area",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Area not found"})
		}

		area.MediaID = req.MediaID
		area.Name = req.Name
		area.Latitude = req.Latitude
		area.Longitude = req.Longitude
		area.UpdatedAt = time.Now().UTC()
		area.UpdatedByID = userOrg.UserID

		if err := core.AreaManager(service).UpdateByID(context, area.ID, area); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Area update failed (/area/:area_id), db error: " + err.Error(),
				Module:      "Area",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update area: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated area (/area/:area_id): " + area.Name,
			Module:      "Area",
		})
		return ctx.JSON(http.StatusOK, core.AreaManager(service).ToModel(area))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/area/:area_id",
		Method: "DELETE",
		Note:   "Deletes the specified area by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		areaID, err := helpers.EngineUUIDParam(ctx, "area_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Area delete failed (/area/:area_id), invalid area ID.",
				Module:      "Area",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid area ID"})
		}
		area, err := core.AreaManager(service).GetByID(context, *areaID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Area delete failed (/area/:area_id), not found.",
				Module:      "Area",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Area not found"})
		}
		if err := core.AreaManager(service).Delete(context, *areaID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Area delete failed (/area/:area_id), db error: " + err.Error(),
				Module:      "Area",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete area: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted area (/area/:area_id): " + area.Name,
			Module:      "Area",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/area/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple areas by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete areas (/area/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Area",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete areas (/area/bulk-delete) | no IDs provided",
				Module:      "Area",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No area IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}

		if err := core.AreaManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete areas (/area/bulk-delete) | error: " + err.Error(),
				Module:      "Area",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete areas: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted areas (/area/bulk-delete)",
			Module:      "Area",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
