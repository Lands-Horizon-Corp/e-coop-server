package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// TimeDepositComputationController registers routes for managing time deposit computations.
func (c *Controller) timeDepositComputationController() {
	req := c.provider.Service.Request

	// POST /time-deposit-computation: Create a new time deposit computation. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-computation/time-deposit-type/:time_deposit_type_id",
		Method:       "POST",
		Note:         "Creates a new time deposit computation for the current user's organization and branch.",
		RequestType:  core.TimeDepositComputationRequest{},
		ResponseType: core.TimeDepositComputationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositTypeID, err := handlers.EngineUUIDParam(ctx, "time_deposit_type_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation creation failed (/time-deposit-computation/time-deposit-type/:time_deposit_type_id), invalid time deposit type ID.",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type ID"})
		}
		req, err := c.core.TimeDepositComputationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation creation failed (/time-deposit-computation), validation error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation creation failed (/time-deposit-computation), user org error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation creation failed (/time-deposit-computation), user not assigned to branch.",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		timeDepositComputation := &core.TimeDepositComputation{
			TimeDepositTypeID: *timeDepositTypeID,
			MinimumAmount:     req.MinimumAmount,
			MaximumAmount:     req.MaximumAmount,
			Header1:           req.Header1,
			Header2:           req.Header2,
			Header3:           req.Header3,
			Header4:           req.Header4,
			Header5:           req.Header5,
			Header6:           req.Header6,
			Header7:           req.Header7,
			Header8:           req.Header8,
			Header9:           req.Header9,
			Header10:          req.Header10,
			Header11:          req.Header11,
			CreatedAt:         time.Now().UTC(),
			CreatedByID:       userOrg.UserID,
			UpdatedAt:         time.Now().UTC(),
			UpdatedByID:       userOrg.UserID,
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrg.OrganizationID,
		}

		if err := c.core.TimeDepositComputationManager.Create(context, timeDepositComputation); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation creation failed (/time-deposit-computation), db error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create time deposit computation: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created time deposit computation (/time-deposit-computation): " + timeDepositComputation.ID.String(),
			Module:      "TimeDepositComputation",
		})
		return ctx.JSON(http.StatusCreated, c.core.TimeDepositComputationManager.ToModel(timeDepositComputation))
	})

	// PUT /time-deposit-computation/:time_deposit_computation_id: Update time deposit computation by ID. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-computation/:time_deposit_computation_id",
		Method:       "PUT",
		Note:         "Updates an existing time deposit computation by its ID.",
		RequestType:  core.TimeDepositComputationRequest{},
		ResponseType: core.TimeDepositComputationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositComputationID, err := handlers.EngineUUIDParam(ctx, "time_deposit_computation_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation update failed (/time-deposit-computation/:time_deposit_computation_id), invalid time deposit computation ID.",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation ID"})
		}

		req, err := c.core.TimeDepositComputationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation update failed (/time-deposit-computation/:time_deposit_computation_id), validation error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation update failed (/time-deposit-computation/:time_deposit_computation_id), user org error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		timeDepositComputation, err := c.core.TimeDepositComputationManager.GetByID(context, *timeDepositComputationID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation update failed (/time-deposit-computation/:time_deposit_computation_id), time deposit computation not found.",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Time deposit computation not found"})
		}
		timeDepositComputation.MinimumAmount = req.MinimumAmount
		timeDepositComputation.MaximumAmount = req.MaximumAmount
		timeDepositComputation.Header1 = req.Header1
		timeDepositComputation.Header2 = req.Header2
		timeDepositComputation.Header3 = req.Header3
		timeDepositComputation.Header4 = req.Header4
		timeDepositComputation.Header5 = req.Header5
		timeDepositComputation.Header6 = req.Header6
		timeDepositComputation.Header7 = req.Header7
		timeDepositComputation.Header8 = req.Header8
		timeDepositComputation.Header9 = req.Header9
		timeDepositComputation.Header10 = req.Header10
		timeDepositComputation.Header11 = req.Header11
		timeDepositComputation.UpdatedAt = time.Now().UTC()
		timeDepositComputation.UpdatedByID = userOrg.UserID
		if err := c.core.TimeDepositComputationManager.UpdateByID(context, timeDepositComputation.ID, timeDepositComputation); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation update failed (/time-deposit-computation/:time_deposit_computation_id), db error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update time deposit computation: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated time deposit computation (/time-deposit-computation/:time_deposit_computation_id): " + timeDepositComputation.ID.String(),
			Module:      "TimeDepositComputation",
		})
		return ctx.JSON(http.StatusOK, c.core.TimeDepositComputationManager.ToModel(timeDepositComputation))
	})

	// DELETE /time-deposit-computation/:time_deposit_computation_id: Delete a time deposit computation by ID. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/time-deposit-computation/:time_deposit_computation_id",
		Method: "DELETE",
		Note:   "Deletes the specified time deposit computation by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositComputationID, err := handlers.EngineUUIDParam(ctx, "time_deposit_computation_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit computation delete failed (/time-deposit-computation/:time_deposit_computation_id), invalid time deposit computation ID.",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation ID"})
		}
		timeDepositComputation, err := c.core.TimeDepositComputationManager.GetByID(context, *timeDepositComputationID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit computation delete failed (/time-deposit-computation/:time_deposit_computation_id), not found.",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Time deposit computation not found"})
		}
		if err := c.core.TimeDepositComputationManager.Delete(context, *timeDepositComputationID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit computation delete failed (/time-deposit-computation/:time_deposit_computation_id), db error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete time deposit computation: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted time deposit computation (/time-deposit-computation/:time_deposit_computation_id): " + timeDepositComputation.ID.String(),
			Module:      "TimeDepositComputation",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Simplified bulk-delete handler for time deposit computations (mirrors feedback/holiday pattern)
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/time-deposit-computation/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple time deposit computations by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Time deposit computation bulk delete failed (/time-deposit-computation/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Time deposit computation bulk delete failed (/time-deposit-computation/bulk-delete) | no IDs provided",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := c.core.TimeDepositComputationManager.BulkDelete(context, ids); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Time deposit computation bulk delete failed (/time-deposit-computation/bulk-delete) | error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete time deposit computations: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted time deposit computations (/time-deposit-computation/bulk-delete)",
			Module:      "TimeDepositComputation",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

}
