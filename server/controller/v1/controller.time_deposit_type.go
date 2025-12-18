package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) timeDepositTypeController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-type",
		Method:       "GET",
		Note:         "Returns a paginated list of time deposit types for the current user's organization and branch.",
		ResponseType: core.TimeDepositTypeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		timeDepositTypes, err := c.core.TimeDepositTypeCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch time deposit types for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.TimeDepositTypeManager.ToModels(timeDepositTypes))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-type/:time_deposit_type_id",
		Method:       "GET",
		Note:         "Returns a single time deposit type by its ID.",
		ResponseType: core.TimeDepositTypeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositTypeID, err := handlers.EngineUUIDParam(ctx, "time_deposit_type_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type ID"})
		}
		timeDepositType, err := c.core.TimeDepositTypeManager.GetByIDRaw(context, *timeDepositTypeID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Time deposit type not found"})
		}
		return ctx.JSON(http.StatusOK, timeDepositType)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-type",
		Method:       "POST",
		Note:         "Creates a new time deposit type for the current user's organization and branch.",
		RequestType:  core.TimeDepositTypeRequest{},
		ResponseType: core.TimeDepositTypeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.TimeDepositTypeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit type creation failed (/time-deposit-type), validation error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit type creation failed (/time-deposit-type), user org error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit type creation failed (/time-deposit-type), user not assigned to branch.",
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		timeDepositType := &core.TimeDepositType{
			Name:           req.Name,
			Description:    req.Description,
			PreMature:      req.PreMature,
			PreMatureRate:  req.PreMatureRate,
			Excess:         req.Excess,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			CurrencyID:     req.CurrencyID,
		}

		if err := c.core.TimeDepositTypeManager.Create(context, timeDepositType); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time deposit type creation failed (/time-deposit-type), db error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create time deposit type: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created time deposit type (/time-deposit-type): " + timeDepositType.Name,
			Module:      "TimeDepositType",
		})
		return ctx.JSON(http.StatusCreated, c.core.TimeDepositTypeManager.ToModel(timeDepositType))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-type/:time_deposit_type_id",
		Method:       "PUT",
		Note:         "Updates an existing time deposit type by its ID.",
		RequestType:  core.TimeDepositTypeRequest{},
		ResponseType: core.TimeDepositTypeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositTypeID, err := handlers.EngineUUIDParam(ctx, "time_deposit_type_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit type update failed (/time-deposit-type/:time_deposit_type_id), invalid time deposit type ID.",
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type ID"})
		}

		req, err := c.core.TimeDepositTypeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit type update failed (/time-deposit-type/:time_deposit_type_id), validation error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit type update failed (/time-deposit-type/:time_deposit_type_id), user org error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		timeDepositType, err := c.core.TimeDepositTypeManager.GetByID(context, *timeDepositTypeID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit type update failed (/time-deposit-type/:time_deposit_type_id), time deposit type not found.",
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Time deposit type not found"})
		}

		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + tx.Error.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		timeDepositType.Name = req.Name
		timeDepositType.Description = req.Description
		timeDepositType.CurrencyID = req.CurrencyID
		timeDepositType.PreMature = req.PreMature
		timeDepositType.PreMatureRate = req.PreMatureRate
		timeDepositType.Excess = req.Excess
		timeDepositType.UpdatedAt = time.Now().UTC()
		timeDepositType.UpdatedByID = userOrg.UserID

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

		if err := c.core.TimeDepositTypeManager.UpdateByIDWithTx(context, tx, timeDepositType.ID, timeDepositType); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time deposit type update failed (/time-deposit-type/:time_deposit_type_id), db error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update time deposit type: " + endTx(err).Error()})
		}

		if req.TimeDepositComputationsDeleted != nil {
			for _, id := range req.TimeDepositComputationsDeleted {
				if err := c.core.TimeDepositComputationManager.DeleteWithTx(context, tx, id); err != nil {
					c.event.Footstep(ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Failed to delete time deposit computation: " + err.Error(),
						Module:      "TimeDepositType",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete time deposit computation: " + endTx(err).Error()})
				}
			}
		}

		if req.TimeDepositComputationPreMaturesDeleted != nil {
			for _, id := range req.TimeDepositComputationPreMaturesDeleted {
				if err := c.core.TimeDepositComputationPreMatureManager.DeleteWithTx(context, tx, id); err != nil {
					c.event.Footstep(ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Failed to delete time deposit computation pre mature: " + err.Error(),
						Module:      "TimeDepositType",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete time deposit computation pre mature: " + endTx(err).Error()})
				}
			}
		}

		if req.TimeDepositComputations != nil {
			for _, computationReq := range req.TimeDepositComputations {
				if computationReq.ID != nil {
					existingComputation, err := c.core.TimeDepositComputationManager.GetByID(context, *computationReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get time deposit computation: " + endTx(err).Error()})
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
					existingComputation.UpdatedByID = userOrg.UserID
					if err := c.core.TimeDepositComputationManager.UpdateByIDWithTx(context, tx, existingComputation.ID, existingComputation); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update time deposit computation: " + endTx(err).Error()})
					}
				} else {
					newComputation := &core.TimeDepositComputation{
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
						CreatedByID:       userOrg.UserID,
						UpdatedAt:         time.Now().UTC(),
						UpdatedByID:       userOrg.UserID,
						BranchID:          *userOrg.BranchID,
						OrganizationID:    userOrg.OrganizationID,
					}
					if err := c.core.TimeDepositComputationManager.CreateWithTx(context, tx, newComputation); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create time deposit computation: " + endTx(err).Error()})
					}
				}
			}
		}

		if req.TimeDepositComputationPreMatures != nil {
			for _, preMatureReq := range req.TimeDepositComputationPreMatures {
				if preMatureReq.ID != nil {
					existingPreMature, err := c.core.TimeDepositComputationPreMatureManager.GetByID(context, *preMatureReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get time deposit computation pre mature: " + endTx(err).Error()})
					}
					existingPreMature.Terms = preMatureReq.Terms
					existingPreMature.From = preMatureReq.From
					existingPreMature.To = preMatureReq.To
					existingPreMature.Rate = preMatureReq.Rate
					existingPreMature.UpdatedAt = time.Now().UTC()
					existingPreMature.UpdatedByID = userOrg.UserID
					if err := c.core.TimeDepositComputationPreMatureManager.UpdateByIDWithTx(context, tx, existingPreMature.ID, existingPreMature); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update time deposit computation pre mature: " + endTx(err).Error()})
					}
				} else {
					newPreMature := &core.TimeDepositComputationPreMature{
						TimeDepositTypeID: timeDepositType.ID,
						Terms:             preMatureReq.Terms,
						From:              preMatureReq.From,
						To:                preMatureReq.To,
						Rate:              preMatureReq.Rate,
						CreatedAt:         time.Now().UTC(),
						CreatedByID:       userOrg.UserID,
						UpdatedAt:         time.Now().UTC(),
						UpdatedByID:       userOrg.UserID,
						BranchID:          *userOrg.BranchID,
						OrganizationID:    userOrg.OrganizationID,
					}
					if err := c.core.TimeDepositComputationPreMatureManager.CreateWithTx(context, tx, newPreMature); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create time deposit computation pre mature: " + endTx(err).Error()})
					}
				}
			}
		}

		if err := endTx(nil); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to commit time deposit type update transaction: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit time deposit type update: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated time deposit type (/time-deposit-type/:time_deposit_type_id): " + timeDepositType.Name,
			Module:      "TimeDepositType",
		})
		newTimeDepositType, err := c.core.TimeDepositTypeManager.GetByID(context, timeDepositType.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated time deposit type: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.TimeDepositTypeManager.ToModel(newTimeDepositType))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/time-deposit-type/:time_deposit_type_id",
		Method: "DELETE",
		Note:   "Deletes the specified time deposit type by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timeDepositTypeID, err := handlers.EngineUUIDParam(ctx, "time_deposit_type_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit type delete failed (/time-deposit-type/:time_deposit_type_id), invalid time deposit type ID.",
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid time deposit type ID"})
		}
		timeDepositType, err := c.core.TimeDepositTypeManager.GetByID(context, *timeDepositTypeID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit type delete failed (/time-deposit-type/:time_deposit_type_id), not found.",
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Time deposit type not found"})
		}
		if err := c.core.TimeDepositTypeManager.Delete(context, *timeDepositTypeID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Time deposit type delete failed (/time-deposit-type/:time_deposit_type_id), db error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete time deposit type: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted time deposit type (/time-deposit-type/:time_deposit_type_id): " + timeDepositType.Name,
			Module:      "TimeDepositType",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/time-deposit-type/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple time deposit types by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Time deposit type bulk delete failed (/time-deposit-type/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Time deposit type bulk delete failed (/time-deposit-type/bulk-delete) | no IDs provided",
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := c.core.TimeDepositTypeManager.BulkDelete(context, ids); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Time deposit type bulk delete failed (/time-deposit-type/bulk-delete) | error: " + err.Error(),
				Module:      "TimeDepositType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete time deposit types: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted time deposit types (/time-deposit-type/bulk-delete)",
			Module:      "TimeDepositType",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/time-deposit-type/currency/:currency_id",
		Method:       "GET",
		Note:         "Fetch time deposit types by currency ID.",
		ResponseType: core.TimeDepositTypeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), user org error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		timeDepositTypes, err := c.core.TimeDepositTypeManager.Find(context, &core.TimeDepositType{
			CurrencyID:     *currencyID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch time deposit types: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.TimeDepositTypeManager.ToModels(timeDepositTypes))
	})
}
