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

// TimeDepositTypeController registers routes for managing time deposit types.
func (c *Controller) timeDepositTypeController(
	req := c.provider.Service.Request

	// GET /time-deposit-type/search: Paginated search of time deposit types for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-type",
		Method:       "GET",
		Note:         "Returns a paginated list of time deposit types for the current user's organization and branch.",
		ResponseType: modelcore.TimeDepositTypeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		timeDepositTypes, err := c.modelcore.TimeDepositTypeCurrentbranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch time deposit types for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.TimeDepositTypeManager.ToModels(timeDepositTypes))
	})

	// GET /time-deposit-type/:time_deposit_type_id: Get specific time deposit type by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-type/:time_deposit_type_id",
		Method:       "GET",
		Note:         "Returns a single time deposit type by its ID.",
		ResponseType: modelcore.TimeDepositTypeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositTypeID, err := handlers.EngineUUIDParam(ctx, "time_deposit_type_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type ID"})
		}
		timeDepositType, err := c.modelcore.TimeDepositTypeManager.GetByIDRaw(context, *timeDepositTypeID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Time deposit type not found"})
		}
		return ctx.JSON(http.StatusOK, timeDepositType)
	})

	// POST /time-deposit-type: Create a new time deposit type. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-type",
		Method:       "POST",
		Note:         "Creates a new time deposit type for the current user's organization and branch.",
		RequestType:  modelcore.TimeDepositTypeRequest{},
		ResponseType: modelcore.TimeDepositTypeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelcore.TimeDepositTypeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit type creation failed (/time-deposit-type), validation error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit type creation failed (/time-deposit-type), user org error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit type creation failed (/time-deposit-type), user not assigned to branch.",
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		timeDepositType := &modelcore.TimeDepositType{
			Name:           req.Name,
			Description:    req.Description,
			PreMature:      req.PreMature,
			PreMatureRate:  req.PreMatureRate,
			Excess:         req.Excess,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
			CurrencyID:     req.CurrencyID,
		}

		if err := c.modelcore.TimeDepositTypeManager.Create(context, timeDepositType); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit type creation failed (/time-deposit-type), db error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create time deposit type: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created time deposit type (/time-deposit-type): " + timeDepositType.Name,
			Module:      "TimeDepositType",
		})
		return ctx.JSON(http.StatusCreated, c.modelcore.TimeDepositTypeManager.ToModel(timeDepositType))
	})

	// PUT /time-deposit-type/:time_deposit_type_id: Update time deposit type by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-type/:time_deposit_type_id",
		Method:       "PUT",
		Note:         "Updates an existing time deposit type by its ID.",
		RequestType:  modelcore.TimeDepositTypeRequest{},
		ResponseType: modelcore.TimeDepositTypeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositTypeID, err := handlers.EngineUUIDParam(ctx, "time_deposit_type_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit type update failed (/time-deposit-type/:time_deposit_type_id), invalid time deposit type ID.",
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type ID"})
		}

		req, err := c.modelcore.TimeDepositTypeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit type update failed (/time-deposit-type/:time_deposit_type_id), validation error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit type update failed (/time-deposit-type/:time_deposit_type_id), user org error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		timeDepositType, err := c.modelcore.TimeDepositTypeManager.GetByID(context, *timeDepositTypeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit type update failed (/time-deposit-type/:time_deposit_type_id), time deposit type not found.",
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Time deposit type not found"})
		}

		// Start database transaction
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + tx.Error.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		// Update main time deposit type fields
		timeDepositType.Name = req.Name
		timeDepositType.Description = req.Description
		timeDepositType.CurrencyID = req.CurrencyID
		timeDepositType.PreMature = req.PreMature
		timeDepositType.PreMatureRate = req.PreMatureRate
		timeDepositType.Excess = req.Excess
		timeDepositType.UpdatedAt = time.Now().UTC()
		timeDepositType.UpdatedByID = user.UserID

		timeDepositType.Header1 = req.Header1
		timeDepositType.Header2 = req.Header2
		timeDepositType.Header3 = req.Header3
		timeDepositType.Header4 = req.Header4
		timeDepositType.Header5 = req.Header5
		timeDepositType.Header6 = req.Header6
		timeDepositType.Header7 = req.Header7
		timeDepositType.Header8 = req.Header8
		timeDepositType.Header9 = req.Header9
		timeDepositType.Header10 = req.Header10
		timeDepositType.Header11 = req.Header11

		if err := c.modelcore.TimeDepositTypeManager.UpdateFieldsWithTx(context, tx, timeDepositType.ID, timeDepositType); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit type update failed (/time-deposit-type/:time_deposit_type_id), db error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update time deposit type: " + err.Error()})
		}

		// Handle deletions first
		if req.TimeDepositComputationsDeleted != nil {
			for _, id := range req.TimeDepositComputationsDeleted {
				if err := c.modelcore.TimeDepositComputationManager.DeleteByIDWithTx(context, tx, id); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Failed to delete time deposit computation: " + err.Error(),
						Module:      "TimeDepositType",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete time deposit computation: " + err.Error()})
				}
			}
		}

		if req.TimeDepositComputationPreMaturesDeleted != nil {
			for _, id := range req.TimeDepositComputationPreMaturesDeleted {
				if err := c.modelcore.TimeDepositComputationPreMatureManager.DeleteByIDWithTx(context, tx, id); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Failed to delete time deposit computation pre mature: " + err.Error(),
						Module:      "TimeDepositType",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete time deposit computation pre mature: " + err.Error()})
				}
			}
		}

		// Handle TimeDepositComputations creation/update
		if req.TimeDepositComputations != nil {
			for _, computationReq := range req.TimeDepositComputations {
				if computationReq.ID != nil {
					// Update existing record
					existingComputation, err := c.modelcore.TimeDepositComputationManager.GetByID(context, *computationReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get time deposit computation: " + err.Error()})
					}
					existingComputation.MinimumAmount = computationReq.MinimumAmount
					existingComputation.MaximumAmount = computationReq.MaximumAmount
					existingComputation.Header1 = computationReq.Header1
					existingComputation.Header2 = computationReq.Header2
					existingComputation.Header3 = computationReq.Header3
					existingComputation.Header4 = computationReq.Header4
					existingComputation.Header5 = computationReq.Header5
					existingComputation.Header6 = computationReq.Header6
					existingComputation.Header7 = computationReq.Header7
					existingComputation.Header8 = computationReq.Header8
					existingComputation.Header9 = computationReq.Header9
					existingComputation.Header10 = computationReq.Header10
					existingComputation.Header11 = computationReq.Header11
					existingComputation.UpdatedAt = time.Now().UTC()
					existingComputation.UpdatedByID = user.UserID
					if err := c.modelcore.TimeDepositComputationManager.UpdateFieldsWithTx(context, tx, existingComputation.ID, existingComputation); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update time deposit computation: " + err.Error()})
					}
				} else {
					// Create new record
					newComputation := &modelcore.TimeDepositComputation{
						TimeDepositTypeID: timeDepositType.ID,
						MinimumAmount:     computationReq.MinimumAmount,
						MaximumAmount:     computationReq.MaximumAmount,
						Header1:           computationReq.Header1,
						Header2:           computationReq.Header2,
						Header3:           computationReq.Header3,
						Header4:           computationReq.Header4,
						Header5:           computationReq.Header5,
						Header6:           computationReq.Header6,
						Header7:           computationReq.Header7,
						Header8:           computationReq.Header8,
						Header9:           computationReq.Header9,
						Header10:          computationReq.Header10,
						Header11:          computationReq.Header11,
						CreatedAt:         time.Now().UTC(),
						CreatedByID:       user.UserID,
						UpdatedAt:         time.Now().UTC(),
						UpdatedByID:       user.UserID,
						BranchID:          *user.BranchID,
						OrganizationID:    user.OrganizationID,
					}
					if err := c.modelcore.TimeDepositComputationManager.CreateWithTx(context, tx, newComputation); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create time deposit computation: " + err.Error()})
					}
				}
			}
		}

		// Handle TimeDepositComputationPreMatures creation/update
		if req.TimeDepositComputationPreMatures != nil {
			for _, preMatureReq := range req.TimeDepositComputationPreMatures {
				if preMatureReq.ID != nil {
					// Update existing record
					existingPreMature, err := c.modelcore.TimeDepositComputationPreMatureManager.GetByID(context, *preMatureReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get time deposit computation pre mature: " + err.Error()})
					}
					existingPreMature.Terms = preMatureReq.Terms
					existingPreMature.From = preMatureReq.From
					existingPreMature.To = preMatureReq.To
					existingPreMature.Rate = preMatureReq.Rate
					existingPreMature.UpdatedAt = time.Now().UTC()
					existingPreMature.UpdatedByID = user.UserID
					if err := c.modelcore.TimeDepositComputationPreMatureManager.UpdateFieldsWithTx(context, tx, existingPreMature.ID, existingPreMature); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update time deposit computation pre mature: " + err.Error()})
					}
				} else {
					// Create new record
					newPreMature := &modelcore.TimeDepositComputationPreMature{
						TimeDepositTypeID: timeDepositType.ID,
						Terms:             preMatureReq.Terms,
						From:              preMatureReq.From,
						To:                preMatureReq.To,
						Rate:              preMatureReq.Rate,
						CreatedAt:         time.Now().UTC(),
						CreatedByID:       user.UserID,
						UpdatedAt:         time.Now().UTC(),
						UpdatedByID:       user.UserID,
						BranchID:          *user.BranchID,
						OrganizationID:    user.OrganizationID,
					}
					if err := c.modelcore.TimeDepositComputationPreMatureManager.CreateWithTx(context, tx, newPreMature); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create time deposit computation pre mature: " + err.Error()})
					}
				}
			}
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to commit time deposit type update transaction: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit time deposit type update: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated time deposit type (/time-deposit-type/:time_deposit_type_id): " + timeDepositType.Name,
			Module:      "TimeDepositType",
		})
		newTimeDepositType, err := c.modelcore.TimeDepositTypeManager.GetByID(context, timeDepositType.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated time deposit type: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.TimeDepositTypeManager.ToModel(newTimeDepositType))
	})

	// DELETE /time-deposit-type/:time_deposit_type_id: Delete a time deposit type by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/time-deposit-type/:time_deposit_type_id",
		Method: "DELETE",
		Note:   "Deletes the specified time deposit type by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositTypeID, err := handlers.EngineUUIDParam(ctx, "time_deposit_type_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit type delete failed (/time-deposit-type/:time_deposit_type_id), invalid time deposit type ID.",
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type ID"})
		}
		timeDepositType, err := c.modelcore.TimeDepositTypeManager.GetByID(context, *timeDepositTypeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit type delete failed (/time-deposit-type/:time_deposit_type_id), not found.",
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Time deposit type not found"})
		}
		if err := c.modelcore.TimeDepositTypeManager.DeleteByID(context, *timeDepositTypeID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit type delete failed (/time-deposit-type/:time_deposit_type_id), db error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete time deposit type: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted time deposit type (/time-deposit-type/:time_deposit_type_id): " + timeDepositType.Name,
			Module:      "TimeDepositType",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /time-deposit-type/bulk-delete: Bulk delete time deposit types by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/time-deposit-type/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple time deposit types by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: modelcore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/time-deposit-type/bulk-delete), invalid request body.",
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/time-deposit-type/bulk-delete), no IDs provided.",
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No time deposit type IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/time-deposit-type/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			timeDepositTypeID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/time-deposit-type/bulk-delete), invalid UUID: " + rawID,
					Module:      "TimeDepositType",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			timeDepositType, err := c.modelcore.TimeDepositTypeManager.GetByID(context, timeDepositTypeID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/time-deposit-type/bulk-delete), not found: " + rawID,
					Module:      "TimeDepositType",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Time deposit type not found with ID: %s", rawID)})
			}
			names += timeDepositType.Name + ","
			if err := c.modelcore.TimeDepositTypeManager.DeleteByIDWithTx(context, tx, timeDepositTypeID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/time-deposit-type/bulk-delete), db error: " + err.Error(),
					Module:      "TimeDepositType",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete time deposit type: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/time-deposit-type/bulk-delete), commit error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted time deposit types (/time-deposit-type/bulk-delete): " + names,
			Module:      "TimeDepositType",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// GET /api/v1/time-deposit-type/currency/:currency_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-type/currency/:currency_id",
		Method:       "GET",
		Note:         "Fetch time deposit types by currency ID.",
		ResponseType: modelcore.TimeDepositTypeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), user org error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		timeDepositTypes, err := c.modelcore.TimeDepositTypeManager.Find(context, &modelcore.TimeDepositType{
			CurrencyID:     *currencyID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch time deposit types: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.TimeDepositTypeManager.ToModels(timeDepositTypes))
	})
}
