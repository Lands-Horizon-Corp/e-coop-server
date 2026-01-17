package time_deposit

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

func TimeDepositComputationPreMatureController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/time-deposit-computation-pre-mature/time-deposit-type/:time_deposit_type_id",
		Method:       "POST",
		Note:         "Creates a new time deposit computation pre mature for the current user's organization and branch.",
		RequestType:  core.TimeDepositComputationPreMatureRequest{},
		ResponseType: core.TimeDepositComputationPreMatureResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositTypeID, err := helpers.EngineUUIDParam(ctx, "time_deposit_type_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation pre mature creation failed (/time-deposit-computation-pre-mature), invalid time deposit type ID.",
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type ID"})
		}
		req, err := core.TimeDepositComputationPreMatureManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation pre mature creation failed (/time-deposit-computation-pre-mature), validation error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation pre mature data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation pre mature creation failed (/time-deposit-computation-pre-mature), user org error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation pre mature creation failed (/time-deposit-computation-pre-mature), user not assigned to branch.",
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		timeDepositComputationPreMature := &types.TimeDepositComputationPreMature{
			TimeDepositTypeID: *timeDepositTypeID,
			Terms:             req.Terms,
			From:              req.From,
			To:                req.To,
			Rate:              req.Rate,
			CreatedAt:         time.Now().UTC(),
			CreatedByID:       userOrg.UserID,
			UpdatedAt:         time.Now().UTC(),
			UpdatedByID:       userOrg.UserID,
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrg.OrganizationID,
		}

		if err := core.TimeDepositComputationPreMatureManager(service).Create(context, timeDepositComputationPreMature); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation pre mature creation failed (/time-deposit-computation-pre-mature), db error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create time deposit computation pre mature: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created time deposit computation pre mature (/time-deposit-computation-pre-mature): " + timeDepositComputationPreMature.ID.String(),
			Module:      "TimeDepositComputationPreMature",
		})
		return ctx.JSON(http.StatusCreated, core.TimeDepositComputationPreMatureManager(service).ToModel(timeDepositComputationPreMature))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id",
		Method:       "PUT",
		Note:         "Updates an existing time deposit computation pre mature by its ID.",
		RequestType:  core.TimeDepositComputationPreMatureRequest{},
		ResponseType: core.TimeDepositComputationPreMatureResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositComputationPreMatureID, err := helpers.EngineUUIDParam(ctx, "time_deposit_computation_pre_mature_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation pre mature update failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), invalid time deposit computation pre mature ID.",
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation pre mature ID"})
		}

		req, err := core.TimeDepositComputationPreMatureManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation pre mature update failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), validation error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation pre mature data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation pre mature update failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), user org error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		timeDepositComputationPreMature, err := core.TimeDepositComputationPreMatureManager(service).GetByID(context, *timeDepositComputationPreMatureID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation pre mature update failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), time deposit computation pre mature not found.",
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Time deposit computation pre mature not found"})
		}
		timeDepositComputationPreMature.TimeDepositTypeID = req.TimeDepositTypeID
		timeDepositComputationPreMature.Terms = req.Terms
		timeDepositComputationPreMature.From = req.From
		timeDepositComputationPreMature.To = req.To
		timeDepositComputationPreMature.Rate = req.Rate
		timeDepositComputationPreMature.UpdatedAt = time.Now().UTC()
		timeDepositComputationPreMature.UpdatedByID = userOrg.UserID
		if err := core.TimeDepositComputationPreMatureManager(service).UpdateByID(context, timeDepositComputationPreMature.ID, timeDepositComputationPreMature); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation pre mature update failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), db error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update time deposit computation pre mature: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated time deposit computation pre mature (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id): " + timeDepositComputationPreMature.ID.String(),
			Module:      "TimeDepositComputationPreMature",
		})
		return ctx.JSON(http.StatusOK, core.TimeDepositComputationPreMatureManager(service).ToModel(timeDepositComputationPreMature))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id",
		Method: "DELETE",
		Note:   "Deletes the specified time deposit computation pre mature by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositComputationPreMatureID, err := helpers.EngineUUIDParam(ctx, "time_deposit_computation_pre_mature_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit computation pre mature delete failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), invalid time deposit computation pre mature ID.",
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation pre mature ID"})
		}
		timeDepositComputationPreMature, err := core.TimeDepositComputationPreMatureManager(service).GetByID(context, *timeDepositComputationPreMatureID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit computation pre mature delete failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), not found.",
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Time deposit computation pre mature not found"})
		}
		if err := core.TimeDepositComputationPreMatureManager(service).Delete(context, *timeDepositComputationPreMatureID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit computation pre mature delete failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), db error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete time deposit computation pre mature: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted time deposit computation pre mature (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id): " + timeDepositComputationPreMature.ID.String(),
			Module:      "TimeDepositComputationPreMature",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/time-deposit-computation-pre-mature/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple time deposit computation pre mature by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Time deposit computation pre-mature bulk delete failed (/time-deposit-computation-pre-mature/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Time deposit computation pre-mature bulk delete failed (/time-deposit-computation-pre-mature/bulk-delete) | no IDs provided",
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.TimeDepositComputationPreMatureManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Time deposit computation pre-mature bulk delete failed (/time-deposit-computation-pre-mature/bulk-delete) | error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete time deposit computation pre-mature records: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted time deposit computation pre-mature records (/time-deposit-computation-pre-mature/bulk-delete)",
			Module:      "TimeDepositComputationPreMature",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

}
