package controller

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

// CashCountController provides endpoints for managing cash counts during the transaction batch workflow.
func (c *Controller) CashCountController() {
	req := c.provider.Service.Request

	req.RegisterRoute(handlers.Route{
		Route:        "/cash-count/search",
		Method:       "GET",
		Note:         "Returns all cash counts of the current branch",
		ResponseType: model.CashCountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCount, err := c.model.CashCountCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No cash counts found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.model.CashCountManager.Pagination(context, ctx, cashCount))
	})

	// GET /cash-count: Retrieve all cash count bills for the current active transaction batch for the user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/cash-count",
		Method:       "GET",
		Note:         "Returns all cash count bills for the current active transaction batch of the authenticated user's branch. Only allowed for 'owner' or 'employee'.",
		ResponseType: model.CashCountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view cash counts"})
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		cashCounts, err := c.model.CashCountManager.Find(context, &model.CashCount{
			TransactionBatchID: transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash counts: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.CashCountManager.Filtered(context, ctx, cashCounts))
	})

	// POST /cash-count: Add a cash count bill to the current transaction batch before ending. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/cash-count",
		Method:       "POST",
		ResponseType: model.CashCountResponse{},
		RequestType:  model.CashCountRequest{},
		Note:         "Adds a cash count bill to the current active transaction batch for the user's branch. Only allowed for 'owner' or 'employee'.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var cashCountReq model.CashCountRequest
		if err := ctx.Bind(&cashCountReq); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash count creation failed (/cash-count), invalid data: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash count creation failed (/cash-count), user org error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for cash count (/cash-count)",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to add cash counts"})
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash count creation failed (/cash-count), transaction batch lookup error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash count creation failed (/cash-count), no open transaction batch.",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		// Validate and set required fields
		if err := c.provider.Service.Validator.Struct(cashCountReq); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash count creation failed (/cash-count), validation error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash count validation failed: " + err.Error()})
		}
		cashCountReq.TransactionBatchID = transactionBatch.ID
		cashCountReq.EmployeeUserID = userOrg.UserID
		cashCountReq.Amount = cashCountReq.BillAmount * float64(cashCountReq.Quantity)

		newCashCount := &model.CashCount{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			CountryCode:        cashCountReq.CountryCode,
			TransactionBatchID: transactionBatch.ID,
			EmployeeUserID:     userOrg.UserID,
			BillAmount:         cashCountReq.BillAmount,
			Quantity:           cashCountReq.Quantity,
			Amount:             cashCountReq.Amount,
			Name:               cashCountReq.Name,
		}

		if err := c.model.CashCountManager.Create(context, newCashCount); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash count creation failed (/cash-count), db error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cash count: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created cash count (/cash-count): " + newCashCount.Name,
			Module:      "CashCount",
		})
		return ctx.JSON(http.StatusCreated, c.model.CashCountManager.ToModel(newCashCount))
	})

	// PUT /cash-count: Update a list of cash count bills for the current transaction batch before ending. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/cash-count",
		Method:       "PUT",
		ResponseType: model.CashCountResponse{},
		RequestType:  model.CashCountRequest{},
		Note:         "Updates cash count bills in the current active transaction batch for the user's branch. Only allowed for 'owner' or 'employee'.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash counts update failed (/cash-count), user org error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for cash counts (/cash-count)",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update cash counts"})
		}
		type CashCountBatchRequest struct {
			CashCounts        []model.CashCountRequest `json:"cash_counts" validate:"required"`
			DeletedCashCounts *[]uuid.UUID             `json:"deleted_cash_counts,omitempty"`
			DepositInBank     *float64                 `json:"deposit_in_bank,omitempty"`
			CashCountTotal    *float64                 `json:"cash_count_total,omitempty"`
			GrandTotal        *float64                 `json:"grand_total,omitempty"`
		}
		var batchRequest CashCountBatchRequest
		if err := ctx.Bind(&batchRequest); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash counts update failed (/cash-count), invalid data: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data: " + err.Error()})
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash counts update failed (/cash-count), transaction batch lookup error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash counts update failed (/cash-count), no open transaction batch.",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		if batchRequest.DeletedCashCounts != nil {
			for _, deletedID := range *batchRequest.DeletedCashCounts {
				if err := c.model.CashCountManager.DeleteByID(context, deletedID); err != nil {
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Cash count delete failed during update (/cash-count), db error: " + err.Error(),
						Module:      "CashCount",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash count: " + err.Error()})
				}
			}
		}

		var updatedCashCounts []*model.CashCount
		for _, cashCountReq := range batchRequest.CashCounts {
			if err := c.provider.Service.Validator.Struct(cashCountReq); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Cash count validation failed during update (/cash-count): " + err.Error(),
					Module:      "CashCount",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash count validation failed: " + err.Error()})
			}
			cashCountReq.TransactionBatchID = transactionBatch.ID
			cashCountReq.EmployeeUserID = userOrg.UserID
			cashCountReq.Amount = cashCountReq.BillAmount * float64(cashCountReq.Quantity)

			if cashCountReq.ID != nil {
				data := &model.CashCount{
					ID:                 *cashCountReq.ID,
					CountryCode:        cashCountReq.CountryCode,
					TransactionBatchID: transactionBatch.ID,
					EmployeeUserID:     userOrg.UserID,
					BillAmount:         cashCountReq.BillAmount,
					Quantity:           cashCountReq.Quantity,
					Amount:             cashCountReq.Amount,
					Name:               cashCountReq.Name,
					CreatedAt:          time.Now().UTC(),
					CreatedByID:        userOrg.UserID,
					UpdatedAt:          time.Now().UTC(),
					UpdatedByID:        userOrg.UserID,
					OrganizationID:     userOrg.OrganizationID,
					BranchID:           *userOrg.BranchID,
				}
				if err := c.model.CashCountManager.UpdateFields(context, *cashCountReq.ID, data); err != nil {
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Cash count update failed during update (/cash-count), db error: " + err.Error(),
						Module:      "CashCount",
					})
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to update cash count: " + err.Error()})
				}
				updatedCashCount, err := c.model.CashCountManager.GetByID(context, *cashCountReq.ID)
				if err != nil {
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Cash count fetch failed after update (/cash-count): " + err.Error(),
						Module:      "CashCount",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated cash count: " + err.Error()})
				}
				updatedCashCounts = append(updatedCashCounts, updatedCashCount)
			} else {
				newCashCount := &model.CashCount{
					CreatedAt:          time.Now().UTC(),
					CreatedByID:        userOrg.UserID,
					UpdatedAt:          time.Now().UTC(),
					UpdatedByID:        userOrg.UserID,
					OrganizationID:     userOrg.OrganizationID,
					BranchID:           *userOrg.BranchID,
					CountryCode:        cashCountReq.CountryCode,
					TransactionBatchID: transactionBatch.ID,
					EmployeeUserID:     userOrg.UserID,
					BillAmount:         cashCountReq.BillAmount,
					Quantity:           cashCountReq.Quantity,
					Amount:             cashCountReq.Amount,
					Name:               cashCountReq.Name,
				}
				if err := c.model.CashCountManager.Create(context, newCashCount); err != nil {
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Cash count creation failed during update (/cash-count), db error: " + err.Error(),
						Module:      "CashCount",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cash count: " + err.Error()})
				}
				updatedCashCounts = append(updatedCashCounts, newCashCount)
			}
		}

		allCashCounts, err := c.model.CashCountManager.Find(context, &model.CashCount{
			TransactionBatchID: transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash count find failed after update (/cash-count): " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated cash counts: " + err.Error()})
		}

		var totalCashCount float64
		for _, cashCount := range allCashCounts {
			totalCashCount += cashCount.Amount
		}

		depositInBank := transactionBatch.DepositInBank
		if batchRequest.DepositInBank != nil {
			depositInBank = *batchRequest.DepositInBank
		}

		grandTotal := totalCashCount + depositInBank

		var responseRequests []model.CashCountRequest
		for _, cashCount := range updatedCashCounts {
			responseRequests = append(responseRequests, model.CashCountRequest{
				ID:                 &cashCount.ID,
				TransactionBatchID: cashCount.TransactionBatchID,
				EmployeeUserID:     cashCount.EmployeeUserID,
				CountryCode:        cashCount.CountryCode,
				BillAmount:         cashCount.BillAmount,
				Quantity:           cashCount.Quantity,
				Amount:             cashCount.Amount,
				Name:               cashCount.Name,
			})
		}

		response := CashCountBatchRequest{
			CashCounts:     responseRequests,
			DepositInBank:  &depositInBank,
			CashCountTotal: &totalCashCount,
			GrandTotal:     &grandTotal,
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated cash counts (/cash-count) for transaction batch",
			Module:      "CashCount",
		})

		return ctx.JSON(http.StatusOK, response)
	})

	// DELETE /cash-count/:id: Delete a specific cash count by ID from the current transaction batch. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/cash-count/:id",
		Method: "DELETE",
		Note:   "Deletes a specific cash count bill with the given ID from the current active transaction batch. Only allowed for 'owner' or 'employee'.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCountID, err := horizon.EngineUUIDParam(ctx, "id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash count delete failed (/cash-count/:id), invalid ID.",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash count ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash count delete failed (/cash-count/:id), user org error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete attempt for cash count (/cash-count/:id)",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete cash counts"})
		}

		cashCount, err := c.model.CashCountManager.GetByID(context, *cashCountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash count delete failed (/cash-count/:id), record not found.",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash count not found for the given ID"})
		}

		if err := c.model.CashCountManager.DeleteByID(context, *cashCountID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash count delete failed (/cash-count/:id), db error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash count: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted cash count (/cash-count/:id): " + cashCount.Name,
			Module:      "CashCount",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// GET /cash-count/:id: Retrieve a specific cash count by ID from the current transaction batch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/cash-count/:id",
		Method:       "GET",
		Note:         "Retrieves a specific cash count bill by its ID from the current active transaction batch. Only allowed for 'owner' or 'employee'.",
		ResponseType: model.CashCountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCountID, err := horizon.EngineUUIDParam(ctx, "id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash count ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view this cash count"})
		}
		cashCount, err := c.model.CashCountManager.GetByID(context, *cashCountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash count not found for the given ID"})
		}
		return ctx.JSON(http.StatusOK, c.model.CashCountManager.ToModel(cashCount))
	})
}
