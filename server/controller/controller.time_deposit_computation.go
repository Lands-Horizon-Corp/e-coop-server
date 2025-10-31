package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelcore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelcore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// TimeDepositComputationController registers routes for managing time deposit computations.
func (c *Controller) TimeDepositComputationController() {
	req := c.provider.Service.Request

	// POST /time-deposit-computation: Create a new time deposit computation. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-computation/time-deposit-type/:time_deposit_type_id",
		Method:       "POST",
		Note:         "Creates a new time deposit computation for the current user's organization and branch.",
		RequestType:  modelcore.TimeDepositComputationRequest{},
		ResponseType: modelcore.TimeDepositComputationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositTypeID, err := handlers.EngineUUIDParam(ctx, "time_deposit_type_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation creation failed (/time-deposit-computation/time-deposit-type/:time_deposit_type_id), invalid time deposit type ID.",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type ID"})
		}
		req, err := c.modelcore.TimeDepositComputationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation creation failed (/time-deposit-computation), validation error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation creation failed (/time-deposit-computation), user org error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation creation failed (/time-deposit-computation), user not assigned to branch.",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		timeDepositComputation := &modelcore.TimeDepositComputation{
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
			CreatedByID:       user.UserID,
			UpdatedAt:         time.Now().UTC(),
			UpdatedByID:       user.UserID,
			BranchID:          *user.BranchID,
			OrganizationID:    user.OrganizationID,
		}

		if err := c.modelcore.TimeDepositComputationManager.Create(context, timeDepositComputation); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit computation creation failed (/time-deposit-computation), db error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create time deposit computation: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created time deposit computation (/time-deposit-computation): " + timeDepositComputation.ID.String(),
			Module:      "TimeDepositComputation",
		})
		return ctx.JSON(http.StatusCreated, c.modelcore.TimeDepositComputationManager.ToModel(timeDepositComputation))
	})

	// PUT /time-deposit-computation/:time_deposit_computation_id: Update time deposit computation by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-computation/:time_deposit_computation_id",
		Method:       "PUT",
		Note:         "Updates an existing time deposit computation by its ID.",
		RequestType:  modelcore.TimeDepositComputationRequest{},
		ResponseType: modelcore.TimeDepositComputationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositComputationID, err := handlers.EngineUUIDParam(ctx, "time_deposit_computation_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation update failed (/time-deposit-computation/:time_deposit_computation_id), invalid time deposit computation ID.",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation ID"})
		}

		req, err := c.modelcore.TimeDepositComputationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation update failed (/time-deposit-computation/:time_deposit_computation_id), validation error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation update failed (/time-deposit-computation/:time_deposit_computation_id), user org error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		timeDepositComputation, err := c.modelcore.TimeDepositComputationManager.GetByID(context, *timeDepositComputationID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		timeDepositComputation.UpdatedByID = user.UserID
		if err := c.modelcore.TimeDepositComputationManager.UpdateFields(context, timeDepositComputation.ID, timeDepositComputation); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit computation update failed (/time-deposit-computation/:time_deposit_computation_id), db error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update time deposit computation: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated time deposit computation (/time-deposit-computation/:time_deposit_computation_id): " + timeDepositComputation.ID.String(),
			Module:      "TimeDepositComputation",
		})
		return ctx.JSON(http.StatusOK, c.modelcore.TimeDepositComputationManager.ToModel(timeDepositComputation))
	})

	// DELETE /time-deposit-computation/:time_deposit_computation_id: Delete a time deposit computation by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/time-deposit-computation/:time_deposit_computation_id",
		Method: "DELETE",
		Note:   "Deletes the specified time deposit computation by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositComputationID, err := handlers.EngineUUIDParam(ctx, "time_deposit_computation_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit computation delete failed (/time-deposit-computation/:time_deposit_computation_id), invalid time deposit computation ID.",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit computation ID"})
		}
		timeDepositComputation, err := c.modelcore.TimeDepositComputationManager.GetByID(context, *timeDepositComputationID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit computation delete failed (/time-deposit-computation/:time_deposit_computation_id), not found.",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Time deposit computation not found"})
		}
		if err := c.modelcore.TimeDepositComputationManager.DeleteByID(context, *timeDepositComputationID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit computation delete failed (/time-deposit-computation/:time_deposit_computation_id), db error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete time deposit computation: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted time deposit computation (/time-deposit-computation/:time_deposit_computation_id): " + timeDepositComputation.ID.String(),
			Module:      "TimeDepositComputation",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /time-deposit-computation/bulk-delete: Bulk delete time deposit computations by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/time-deposit-computation/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple time deposit computations by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: modelcore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/time-deposit-computation/bulk-delete), invalid request body.",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/time-deposit-computation/bulk-delete), no IDs provided.",
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No time deposit computation IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/time-deposit-computation/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		ids := ""
		for _, rawID := range reqBody.IDs {
			timeDepositComputationID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/time-deposit-computation/bulk-delete), invalid UUID: " + rawID,
					Module:      "TimeDepositComputation",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			timeDepositComputation, err := c.modelcore.TimeDepositComputationManager.GetByID(context, timeDepositComputationID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/time-deposit-computation/bulk-delete), not found: " + rawID,
					Module:      "TimeDepositComputation",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Time deposit computation not found with ID: %s", rawID)})
			}
			ids += timeDepositComputation.ID.String() + ","
			if err := c.modelcore.TimeDepositComputationManager.DeleteByIDWithTx(context, tx, timeDepositComputationID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/time-deposit-computation/bulk-delete), db error: " + err.Error(),
					Module:      "TimeDepositComputation",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete time deposit computation: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/time-deposit-computation/bulk-delete), commit error: " + err.Error(),
				Module:      "TimeDepositComputation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted time deposit computations (/time-deposit-computation/bulk-delete): " + ids,
			Module:      "TimeDepositComputation",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

}
