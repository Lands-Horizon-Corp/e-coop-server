package controller

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

// CashCountController provides endpoints for managing cash counts during the transaction batch workflow.
func (c *Controller) CashCountController() {
	req := c.provider.Service.Request

	// GET /cash-count: Retrieve all cash count bills for the current active transaction batch for the user's branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count",
		Method:   "GET",
		Response: "ICashCount[]",
		Note:     "Returns all cash count bills for the current active transaction batch of the authenticated user's branch. Only allowed for 'owner' or 'employee'.",
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

		return ctx.JSON(http.StatusOK, c.model.CashCountManager.ToModels(cashCounts))
	})

	// POST /cash-count: Add a cash count bill to the current transaction batch before ending.
	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count",
		Method:   "POST",
		Response: "ICashCount",
		Request:  "ICashCount",
		Note:     "Adds a cash count bill to the current active transaction batch for the user's branch. Only allowed for 'owner' or 'employee'.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var cashCountReq model.CashCountRequest
		if err := ctx.Bind(&cashCountReq); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to add cash counts"})
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

		// Validate and set required fields
		if err := c.provider.Service.Validator.Struct(cashCountReq); err != nil {
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cash count: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, c.model.CashCountManager.ToModel(newCashCount))
	})

	// PUT /cash-count: Update a list of cash count bills for the current transaction batch before ending.
	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count",
		Method:   "PUT",
		Response: "ICashCount[]",
		Request:  "ICashCount[]",
		Note:     "Updates cash count bills in the current active transaction batch for the user's branch. Only allowed for 'owner' or 'employee'.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
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
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data: " + err.Error()})
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

		if batchRequest.DeletedCashCounts != nil {
			for _, deletedID := range *batchRequest.DeletedCashCounts {
				if err := c.model.CashCountManager.DeleteByID(context, deletedID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash count: " + err.Error()})
				}
			}
		}

		var updatedCashCounts []*model.CashCount
		for _, cashCountReq := range batchRequest.CashCounts {
			if err := c.provider.Service.Validator.Struct(cashCountReq); err != nil {
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
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to update cash count: " + err.Error()})
				}
				updatedCashCount, err := c.model.CashCountManager.GetByID(context, *cashCountReq.ID)
				if err != nil {
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

		return ctx.JSON(http.StatusOK, response)
	})

	// DELETE /cash-count/:id: Delete a specific cash count by ID from the current transaction batch.
	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count/:id",
		Method:   "DELETE",
		Response: "ICashCount",
		Note:     "Deletes a specific cash count bill with the given ID from the current active transaction batch. Only allowed for 'owner' or 'employee'.",
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete cash counts"})
		}

		if err := c.model.CashCountManager.DeleteByID(context, *cashCountID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash count: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// GET /cash-count/:id: Retrieve a specific cash count by ID from the current transaction batch.
	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count/:id",
		Method:   "GET",
		Response: "ICashCount",
		Note:     "Retrieves a specific cash count bill by its ID from the current active transaction batch. Only allowed for 'owner' or 'employee'.",
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
