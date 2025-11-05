package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// TimeDepositComputationPreMatureController registers routes for managing time deposit computation pre mature.
func (c *Controller) timeDepositComputationPreMatureController() {
	req := c.provider.Service.Request

	// POST /time-deposit-computation-pre-mature: Create a new time deposit computation pre mature. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-computation-pre-mature/time-deposit-type/:time_deposit_type_id",
		Method:       "POST",
		Note:         "Creates a new time deposit computation pre mature for the current user's organization and branch.",
		RequestType:  core.TimeDepositComputationPreMatureRequest{},
		ResponseType: core.TimeDepositComputationPreMatureResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositTypeID, err := handlers.EngineUUIDParam(ctx, "time_deposit_type_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation pre mature creation failed (/time-deposit-computation-pre-mature), invalid time deposit type ID.",
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type ID"})
		}
		req, err := c.core.TimeDepositComputationPreMatureManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation pre mature creation failed (/time-deposit-computation-pre-mature), validation error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation pre mature data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation pre mature creation failed (/time-deposit-computation-pre-mature), user org error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation pre mature creation failed (/time-deposit-computation-pre-mature), user not assigned to branch.",
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		timeDepositComputationPreMature := &core.TimeDepositComputationPreMature{
			TimeDepositTypeID: *timeDepositTypeID,
			Terms:             req.Terms,
			From:              req.From,
			To:                req.To,
			Rate:              req.Rate,
			CreatedAt:         time.Now().UTC(),
			CreatedByID:       user.UserID,
			UpdatedAt:         time.Now().UTC(),
			UpdatedByID:       user.UserID,
			BranchID:          *user.BranchID,
			OrganizationID:    user.OrganizationID,
		}

		if err := c.core.TimeDepositComputationPreMatureManager.Create(context, timeDepositComputationPreMature); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation pre mature creation failed (/time-deposit-computation-pre-mature), db error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create time deposit computation pre mature: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created time deposit computation pre mature (/time-deposit-computation-pre-mature): " + timeDepositComputationPreMature.ID.String(),
			Module:      "TimeDepositComputationPreMature",
		})
		return ctx.JSON(http.StatusCreated, c.core.TimeDepositComputationPreMatureManager.ToModel(timeDepositComputationPreMature))
	})

	// PUT /time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id: Update time deposit computation pre mature by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id",
		Method:       "PUT",
		Note:         "Updates an existing time deposit computation pre mature by its ID.",
		RequestType:  core.TimeDepositComputationPreMatureRequest{},
		ResponseType: core.TimeDepositComputationPreMatureResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositComputationPreMatureID, err := handlers.EngineUUIDParam(ctx, "time_deposit_computation_pre_mature_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation pre mature update failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), invalid time deposit computation pre mature ID.",
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation pre mature ID"})
		}

		req, err := c.core.TimeDepositComputationPreMatureManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation pre mature update failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), validation error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation pre mature data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation pre mature update failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), user org error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		timeDepositComputationPreMature, err := c.core.TimeDepositComputationPreMatureManager.GetByID(context, *timeDepositComputationPreMatureID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		timeDepositComputationPreMature.UpdatedByID = user.UserID
		if err := c.core.TimeDepositComputationPreMatureManager.UpdateByID(context, timeDepositComputationPreMature.ID, timeDepositComputationPreMature); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation pre mature update failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), db error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update time deposit computation pre mature: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated time deposit computation pre mature (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id): " + timeDepositComputationPreMature.ID.String(),
			Module:      "TimeDepositComputationPreMature",
		})
		return ctx.JSON(http.StatusOK, c.core.TimeDepositComputationPreMatureManager.ToModel(timeDepositComputationPreMature))
	})

	// DELETE /time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id: Delete a time deposit computation pre mature by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id",
		Method: "DELETE",
		Note:   "Deletes the specified time deposit computation pre mature by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositComputationPreMatureID, err := handlers.EngineUUIDParam(ctx, "time_deposit_computation_pre_mature_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit computation pre mature delete failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), invalid time deposit computation pre mature ID.",
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation pre mature ID"})
		}
		timeDepositComputationPreMature, err := c.core.TimeDepositComputationPreMatureManager.GetByID(context, *timeDepositComputationPreMatureID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit computation pre mature delete failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), not found.",
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Time deposit computation pre mature not found"})
		}
		if err := c.core.TimeDepositComputationPreMatureManager.Delete(context, *timeDepositComputationPreMatureID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit computation pre mature delete failed (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id), db error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete time deposit computation pre mature: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted time deposit computation pre mature (/time-deposit-computation-pre-mature/:time_deposit_computation_pre_mature_id): " + timeDepositComputationPreMature.ID.String(),
			Module:      "TimeDepositComputationPreMature",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Simplified bulk-delete handler for time deposit computation pre-mature (mirrors feedback/holiday pattern)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/time-deposit-computation-pre-mature/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple time deposit computation pre mature by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Time deposit computation pre-mature bulk delete failed (/time-deposit-computation-pre-mature/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Time deposit computation pre-mature bulk delete failed (/time-deposit-computation-pre-mature/bulk-delete) | no IDs provided",
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		// Delegate deletion to the manager. Manager should handle transactionality, per-record validation and DeletedBy bookkeeping.
		if err := c.core.TimeDepositComputationPreMatureManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Time deposit computation pre-mature bulk delete failed (/time-deposit-computation-pre-mature/bulk-delete) | error: " + err.Error(),
				Module:      "TimeDepositComputationPreMature",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete time deposit computation pre-mature records: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted time deposit computation pre-mature records (/time-deposit-computation-pre-mature/bulk-delete)",
			Module:      "TimeDepositComputationPreMature",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

}
