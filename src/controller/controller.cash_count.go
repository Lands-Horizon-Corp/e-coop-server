package controller

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) CashCountController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count",
		Method:   "GET",
		Response: "ICashCount[]",
		Note:     "Retrieve batch cash count bills (JWT) for the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Find the current active transaction batch
		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if transactionBatch == nil {
			return c.BadRequest(ctx, "No active transaction batch found")
		}

		// Retrieve cash counts for the current transaction batch
		cashCounts, err := c.model.CashCountManager.Find(context, &model.CashCount{
			TransactionBatchID: transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.CashCountManager.ToModels(cashCounts))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count",
		Method:   "POST",
		Response: "ICashCount",
		Request:  "ICashCount",
		Note:     "Add a cash count bill to the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count",
		Method:   "PUT",
		Response: "ICashCount[]",
		Request:  "ICashCount[]",
		Note:     "Update a cash count bill in the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
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
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Find the current active transaction batch
		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if transactionBatch == nil {
			return c.BadRequest(ctx, "No active transaction batch found")
		}

		// Handle deleted cash counts first
		if batchRequest.DeletedCashCounts != nil {
			for _, deletedID := range *batchRequest.DeletedCashCounts {
				if err := c.model.CashCountManager.DeleteByID(context, deletedID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash count: " + err.Error()})
				}
			}
		}

		// Validate and update each cash count
		var updatedCashCounts []*model.CashCount
		for _, cashCountReq := range batchRequest.CashCounts {
			// Validate each cash count request
			if err := c.provider.Service.Validator.Struct(cashCountReq); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}

			// Set required fields
			cashCountReq.TransactionBatchID = transactionBatch.ID
			cashCountReq.EmployeeUserID = userOrg.UserID

			// Calculate amount
			cashCountReq.Amount = cashCountReq.BillAmount * float64(cashCountReq.Quantity)

			// Handle update or create based on ID presence
			if cashCountReq.ID != nil {
				// Update existing cash count
				data := &model.CashCount{
					ID:                 *cashCountReq.ID,
					CountryCode:        cashCountReq.CountryCode,
					TransactionBatchID: transactionBatch.ID,
					EmployeeUserID:     userOrg.UserID,
					BillAmount:         cashCountReq.BillAmount,
					Quantity:           cashCountReq.Quantity,
					Amount:             cashCountReq.Amount,

					CreatedAt:      time.Now().UTC(),
					CreatedByID:    userOrg.UserID,
					UpdatedAt:      time.Now().UTC(),
					UpdatedByID:    userOrg.UserID,
					OrganizationID: userOrg.OrganizationID,
					BranchID:       *userOrg.BranchID,
					Name:           cashCountReq.Name,
				}
				if err := c.model.CashCountManager.UpdateFields(context, *cashCountReq.ID, data); err != nil {
					return echo.NewHTTPError(http.StatusForbidden, "failed to update user: "+err.Error())
				}

				updatedCashCount, err := c.model.CashCountManager.GetByID(context, *cashCountReq.ID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated cash count: " + err.Error()})
				}
				updatedCashCounts = append(updatedCashCounts, updatedCashCount)
			} else {
				// Create new cash count
				newCashCount := &model.CashCount{
					CreatedAt:      time.Now().UTC(),
					CreatedByID:    userOrg.UserID,
					UpdatedAt:      time.Now().UTC(),
					UpdatedByID:    userOrg.UserID,
					OrganizationID: userOrg.OrganizationID,
					BranchID:       *userOrg.BranchID,

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

		// Recalculate totals for response (don't update transaction batch)
		allCashCounts, err := c.model.CashCountManager.Find(context, &model.CashCount{
			TransactionBatchID: transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate new total cash count value
		var totalCashCount float64
		for _, cashCount := range allCashCounts {
			totalCashCount += cashCount.Amount
		}

		// Calculate deposit in bank (use provided value or existing)
		depositInBank := transactionBatch.DepositInBank
		if batchRequest.DepositInBank != nil {
			depositInBank = *batchRequest.DepositInBank
		}

		// Calculate grand total
		grandTotal := totalCashCount + depositInBank

		// Convert cash counts to request format for response
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
			})
		}

		// Return the batch response with calculated totals
		response := CashCountBatchRequest{
			CashCounts:     responseRequests,
			DepositInBank:  &depositInBank,
			CashCountTotal: &totalCashCount,
			GrandTotal:     &grandTotal,
		}

		return ctx.JSON(http.StatusOK, response)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count/:id",
		Method:   "DELETE",
		Response: "ICashCount",
		Note:     "Delete cash count (JWT) with the specified ID from the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count/:id",
		Method:   "GET",
		Response: "ICashCount",
		Note:     "Retrieve specific cash count information based on ID from the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		return nil
	})
}
