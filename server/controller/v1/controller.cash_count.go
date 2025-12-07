package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CashCountController provides endpoints for managing cash counts during the transaction batch workflow.
func (c *Controller) cashCountController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/cash-count/search",
		Method:       "GET",
		Note:         "Returns all cash counts of the current branch",
		ResponseType: core.CashCountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCount, err := c.core.CashCountManager.PaginationWithFields(context, ctx, &core.CashCount{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No cash counts found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, cashCount)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/cash-count/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		Note:         "Returns all cash counts for a specific transaction batch",
		ResponseType: core.CashCountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		cashCount, err := c.core.CashCountManager.PaginationWithFields(context, ctx, &core.CashCount{
			TransactionBatchID: *transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No cash counts found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, cashCount)
	})

	// GET /cash-count: Retrieve all cash count bills for the current active transaction batch for the user's branch. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/cash-count",
		Method:       "GET",
		Note:         "Returns all cash count bills for the current active transaction batch of the authenticated user's branch. Only allowed for 'owner' or 'employee'.",
		ResponseType: core.CashCountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view cash counts"})
		}

		transactionBatch, err := c.core.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		cashCounts, err := c.core.CashCountManager.Find(context, &core.CashCount{
			TransactionBatchID: transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash counts: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.core.CashCountManager.ToModels(cashCounts))
	})

	// POST /cash-count: Add a cash count bill to the current transaction batch before ending. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/cash-count",
		Method:       "POST",
		ResponseType: core.CashCountResponse{},
		RequestType:  core.CashCountRequest{},
		Note:         "Adds a cash count bill to the current active transaction batch for the user's branch. Only allowed for 'owner' or 'employee'.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var cashCountReq core.CashCountRequest
		if err := ctx.Bind(&cashCountReq); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash count creation failed (/cash-count), invalid data: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash count creation failed (/cash-count), user org error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for cash count (/cash-count)",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to add cash counts"})
		}

		transactionBatch, err := c.core.TransactionBatchCurrent(
			context,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash count creation failed (/cash-count), transaction batch lookup error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash count creation failed (/cash-count), no open transaction batch.",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		// Validate and set required fields
		if err := c.provider.Service.Validator.Struct(cashCountReq); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash count creation failed (/cash-count), validation error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash count validation failed: " + err.Error()})
		}
		cashCountReq.TransactionBatchID = transactionBatch.ID
		cashCountReq.EmployeeUserID = userOrg.UserID
		cashCountReq.Amount = c.provider.Service.Decimal.Multiply(cashCountReq.BillAmount, float64(cashCountReq.Quantity))

		newCashCount := &core.CashCount{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			CurrencyID:         cashCountReq.CurrencyID,
			TransactionBatchID: transactionBatch.ID,
			EmployeeUserID:     userOrg.UserID,
			BillAmount:         cashCountReq.BillAmount,
			Quantity:           cashCountReq.Quantity,
			Amount:             cashCountReq.Amount,
			Name:               cashCountReq.Name,
		}

		if err := c.core.CashCountManager.Create(context, newCashCount); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash count creation failed (/cash-count), db error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cash count: " + err.Error()})
		}

		if err := c.event.TransactionBatchBalancing(context, &transactionBatch.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch after saving: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created cash count (/cash-count): " + newCashCount.Name,
			Module:      "CashCount",
		})
		return ctx.JSON(http.StatusCreated, c.core.CashCountManager.ToModel(newCashCount))
	})

	// PUT /cash-count: Update a list of cash count bills for the current transaction batch before ending. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/cash-count",
		Method:       "PUT",
		ResponseType: core.CashCountResponse{},
		RequestType:  core.CashCountRequest{},
		Note:         "Updates cash count bills in the current active transaction batch for the user's branch. Only allowed for 'owner' or 'employee'.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash counts update failed (/cash-count), user org error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for cash counts (/cash-count)",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update cash counts"})
		}
		type CashCountBatchRequest struct {
			CashCounts        []core.CashCountRequest `json:"cash_counts" validate:"required"`
			DeletedCashCounts *uuid.UUIDs             `json:"deleted_cash_counts,omitempty"`
			DepositInBank     *float64                `json:"deposit_in_bank,omitempty"`
			CashCountTotal    *float64                `json:"cash_count_total,omitempty"`
			GrandTotal        *float64                `json:"grand_total,omitempty"`
		}
		var batchRequest CashCountBatchRequest
		if err := ctx.Bind(&batchRequest); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash counts update failed (/cash-count), invalid data: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data: " + err.Error()})
		}

		transactionBatch, err := c.core.TransactionBatchCurrent(
			context,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash counts update failed (/cash-count), transaction batch lookup error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash counts update failed (/cash-count), no open transaction batch.",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		if batchRequest.DeletedCashCounts != nil {
			for _, deletedID := range *batchRequest.DeletedCashCounts {
				if err := c.core.CashCountManager.Delete(context, deletedID); err != nil {
					c.event.Footstep(ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Cash count delete failed during update (/cash-count), db error: " + err.Error(),
						Module:      "CashCount",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash count: " + err.Error()})
				}
			}
		}

		var updatedCashCounts []*core.CashCount
		for _, cashCountReq := range batchRequest.CashCounts {
			if err := c.provider.Service.Validator.Struct(cashCountReq); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Cash count validation failed during update (/cash-count): " + err.Error(),
					Module:      "CashCount",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash count validation failed: " + err.Error()})
			}
			cashCountReq.TransactionBatchID = transactionBatch.ID
			cashCountReq.EmployeeUserID = userOrg.UserID
			cashCountReq.Amount = c.provider.Service.Decimal.Multiply(cashCountReq.BillAmount, float64(cashCountReq.Quantity))
			if cashCountReq.ID != nil {
				updatedCashCount, err := c.core.CashCountManager.GetByID(context, *cashCountReq.ID)
				if err != nil {
					c.event.Footstep(ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Cash count fetch failed after update (/cash-count): " + err.Error(),
						Module:      "CashCount",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated cash count: " + err.Error()})
				}
				updatedCashCount.CurrencyID = cashCountReq.CurrencyID
				updatedCashCount.TransactionBatchID = transactionBatch.ID
				updatedCashCount.EmployeeUserID = userOrg.UserID
				updatedCashCount.BillAmount = cashCountReq.BillAmount
				updatedCashCount.Quantity = cashCountReq.Quantity
				updatedCashCount.Amount = cashCountReq.Amount
				updatedCashCount.Name = cashCountReq.Name
				updatedCashCount.CreatedAt = time.Now().UTC()
				updatedCashCount.CreatedByID = userOrg.UserID
				updatedCashCount.UpdatedAt = time.Now().UTC()
				updatedCashCount.UpdatedByID = userOrg.UserID
				updatedCashCount.OrganizationID = userOrg.OrganizationID
				updatedCashCount.BranchID = *userOrg.BranchID
				if err := c.core.CashCountManager.UpdateByID(context, *cashCountReq.ID, updatedCashCount); err != nil {
					c.event.Footstep(ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Cash count update failed during update (/cash-count), db error: " + err.Error(),
						Module:      "CashCount",
					})
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to update cash count: " + err.Error()})
				}
				updatedCashCounts = append(updatedCashCounts, updatedCashCount)
			} else {
				newCashCount := &core.CashCount{
					CreatedAt:          time.Now().UTC(),
					CreatedByID:        userOrg.UserID,
					UpdatedAt:          time.Now().UTC(),
					UpdatedByID:        userOrg.UserID,
					OrganizationID:     userOrg.OrganizationID,
					BranchID:           *userOrg.BranchID,
					CurrencyID:         cashCountReq.CurrencyID,
					TransactionBatchID: transactionBatch.ID,
					EmployeeUserID:     userOrg.UserID,
					BillAmount:         cashCountReq.BillAmount,
					Quantity:           cashCountReq.Quantity,
					Amount:             cashCountReq.Amount,
					Name:               cashCountReq.Name,
				}
				if err := c.core.CashCountManager.Create(context, newCashCount); err != nil {
					c.event.Footstep(ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Cash count creation failed during update (/cash-count), db error: " + err.Error(),
						Module:      "CashCount",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cash count: " + err.Error()})
				}
				updatedCashCounts = append(updatedCashCounts, newCashCount)
			}
		}
		allCashCounts, err := c.core.CashCountManager.Find(context, &core.CashCount{
			TransactionBatchID: transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash count find failed after update (/cash-count): " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated cash counts: " + err.Error()})
		}

		var totalCashCount float64
		for _, cashCount := range allCashCounts {
			totalCashCount = c.provider.Service.Decimal.Add(totalCashCount, cashCount.Amount)
		}

		depositInBank := transactionBatch.DepositInBank
		if batchRequest.DepositInBank != nil {
			depositInBank = *batchRequest.DepositInBank
		}

		grandTotal := c.provider.Service.Decimal.Add(totalCashCount, depositInBank)
		var responseRequests []core.CashCountRequest
		for _, cashCount := range updatedCashCounts {
			responseRequests = append(responseRequests, core.CashCountRequest{
				ID:                 &cashCount.ID,
				TransactionBatchID: cashCount.TransactionBatchID,
				EmployeeUserID:     cashCount.EmployeeUserID,
				CurrencyID:         cashCount.CurrencyID,
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

		if err := c.event.TransactionBatchBalancing(context, &transactionBatch.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch after saving: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated cash counts (/cash-count) for transaction batch",
			Module:      "CashCount",
		})

		return ctx.JSON(http.StatusOK, response)
	})

	// DELETE /cash-count/:id: Delete a specific cash count by ID from the current transaction batch. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/cash-count/:id",
		Method: "DELETE",
		Note:   "Deletes a specific cash count bill with the given ID from the current active transaction batch. Only allowed for 'owner' or 'employee'.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCountID, err := handlers.EngineUUIDParam(ctx, "id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash count delete failed (/cash-count/:id), invalid ID.",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash count ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash count delete failed (/cash-count/:id), user org error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete attempt for cash count (/cash-count/:id)",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete cash counts"})
		}

		cashCount, err := c.core.CashCountManager.GetByID(context, *cashCountID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash count delete failed (/cash-count/:id), record not found.",
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash count not found for the given ID"})
		}

		if err := c.core.CashCountManager.Delete(context, *cashCountID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash count delete failed (/cash-count/:id), db error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash count: " + err.Error()})
		}
		transactionBatch, err := c.core.TransactionBatchCurrent(
			context,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash counts delete failed (/cash-count), transaction batch lookup error: " + err.Error(),
				Module:      "CashCount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if err := c.event.TransactionBatchBalancing(context, &transactionBatch.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch after saving: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted cash count (/cash-count/:id): " + cashCount.Name,
			Module:      "CashCount",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// GET /cash-count/:id: Retrieve a specific cash count by ID from the current transaction batch. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/cash-count/:id",
		Method:       "GET",
		Note:         "Retrieves a specific cash count bill by its ID from the current active transaction batch. Only allowed for 'owner' or 'employee'.",
		ResponseType: core.CashCountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCountID, err := handlers.EngineUUIDParam(ctx, "id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash count ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view this cash count"})
		}
		cashCount, err := c.core.CashCountManager.GetByID(context, *cashCountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash count not found for the given ID"})
		}
		return ctx.JSON(http.StatusOK, c.core.CashCountManager.ToModel(cashCount))
	})
}
