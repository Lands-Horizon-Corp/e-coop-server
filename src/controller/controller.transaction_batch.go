package controller

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) TransactionBatchController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-batch",
		Method:   "GET",
		Response: "ITransactionBatch[]",
		Note:     "List all transaction batches for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		transactionBatch, err := c.model.TransactionBatchManager.Find(context, &model.TransactionBatch{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionBatchManager.ToModels(transactionBatch))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-batch/current",
		Method:   "GET",
		Response: "ITransactionBatch",
		Note:     "Get the current active transaction batch for the user",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]interface{}{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if transactionBatch == nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "No current transaction batch"})
		}
		if !transactionBatch.CanView {
			result, err := c.model.TransactionBatchMinimal(context, transactionBatch.ID)
			if err != nil {
				return c.InternalServerError(ctx, err)
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionBatchManager.ToModel(transactionBatch))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-batch/:transaction_batch_id/deposit-in-bank",
		Method:   "PUT",
		Response: "ITransactionBatch",
		Request:  "IDepositInBankRequest",
		Note:     "Update the deposit in bank amount for a specific transaction batch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get transaction batch ID from URL parameter
		transactionBatchId, err := horizon.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return err
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Bind the deposit in bank request
		type DepositInBankRequest struct {
			DepositInBank float64 `json:"deposit_in_bank" validate:"min=0"`
		}
		var req DepositInBankRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Validate the request
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Get the transaction batch by ID
		transactionBatch, err := c.model.TransactionBatchManager.GetByID(context, *transactionBatchId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction batch not found"})
		}

		// Verify the transaction batch belongs to the user's organization and branch
		if transactionBatch.OrganizationID != userOrg.OrganizationID ||
			transactionBatch.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "Transaction batch not found in your organization/branch")
		}

		// Check if the transaction batch is still open
		if transactionBatch.IsClosed {
			return c.BadRequest(ctx, "Cannot update deposit for a closed transaction batch")
		}

		// Get all cash counts for recalculating totals
		cashCounts, err := c.model.CashCountManager.Find(context, &model.CashCount{
			TransactionBatchID: transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total cash count value
		var totalCashCount float64
		for _, cashCount := range cashCounts {
			totalCashCount += cashCount.Amount
		}

		// Update transaction batch with new deposit amount and recalculated totals
		transactionBatch.DepositInBank = req.DepositInBank
		transactionBatch.GrandTotal = totalCashCount + req.DepositInBank
		transactionBatch.TotalCashHandled = transactionBatch.BeginningBalance + req.DepositInBank + totalCashCount
		transactionBatch.TotalDepositInBank = req.DepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		// Save the updated transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the updated transaction batch
		if !transactionBatch.CanView {
			result, err := c.model.TransactionBatchMinimal(context, transactionBatch.ID)
			if err != nil {
				return c.InternalServerError(ctx, err)
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionBatchManager.ToModel(transactionBatch))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-batch",
		Method:   "POST",
		Response: "ITransactionBatch",
		Request:  "ITransactionBatch",
		Note:     "Create and start a new transaction batch; returns the created batch. (Will populate Cashcount)",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		batchFundingReq, err := c.model.BatchFundingManager.Validate(ctx)
		if err != nil {
			return err
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		transactionBatch, _ := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]interface{}{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if transactionBatch != nil {
			return c.BadRequest(ctx, "There is ongoing transaction batch")
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		transBatch := &model.TransactionBatch{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			EmployeeUserID: &userOrg.UserID,

			BeginningBalance:              batchFundingReq.Amount,
			DepositInBank:                 0,
			CashCountTotal:                0,
			GrandTotal:                    0,
			TotalCashCollection:           0,
			TotalDepositEntry:             0,
			PettyCash:                     0,
			LoanReleases:                  0,
			TimeDepositWithdrawal:         0,
			SavingsWithdrawal:             0,
			TotalCashHandled:              0,
			TotalSupposedRemitance:        0,
			TotalCashOnHand:               0,
			TotalCheckRemittance:          0,
			TotalOnlineRemittance:         0,
			TotalDepositInBank:            0,
			TotalActualRemittance:         0,
			TotalActualSupposedComparison: 0,

			IsClosed:    false,
			CanView:     false,
			RequestView: nil,
		}
		if err := c.model.TransactionBatchManager.CreateWithTx(context, tx, transBatch); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
		}
		batchFunding := &model.BatchFunding{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TransactionBatchID: transBatch.ID,

			ProvidedByUserID: userOrg.UserID,
			Name:             batchFundingReq.Name,
			Description:      batchFundingReq.Description,
			Amount:           batchFundingReq.Amount,
			SignatureMediaID: batchFundingReq.SignatureMediaID,
		}
		if err := c.model.BatchFundingManager.CreateWithTx(context, tx, batchFunding); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
		}
		result, err := c.model.TransactionBatchMinimal(context, transBatch.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, result)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-batch/end",
		Method:   "PUT",
		Response: "ITransactionBatch",
		Request:  "ITransactionBatch",
		Note:     "End the current transaction batch for the authenticated user",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.TransactionBatchEndRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		transactionBatch, err := c.model.TransactionBatchManager.FindOne(context, &model.TransactionBatch{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			IsClosed:       false,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if transactionBatch == nil {
			return c.BadRequest(ctx, "No active transaction batch found")
		}
		now := time.Now().UTC()
		transactionBatch.IsClosed = true
		transactionBatch.EmployeeBySignatureMediaID = req.EmployeeBySignatureMediaID
		transactionBatch.EmployeeByName = req.EmployeeByName
		transactionBatch.EmployeeByPosition = req.EmployeeByPosition
		transactionBatch.EndedAt = &now
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update transaction batch: "+err.Error())
		}

		if !transactionBatch.CanView {
			result, err := c.model.TransactionBatchMinimal(context, transactionBatch.ID)
			if err != nil {
				return c.InternalServerError(ctx, err)
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionBatchManager.ToModel(transactionBatch))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/transaction-batch/:transaction_batch_id",
		Method: "GET",
		Note:   "Retrieve a transaction batch by its ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchId, err := horizon.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return err
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		transactionBatch, err := c.model.TransactionBatchManager.GetByID(context, *transactionBatchId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "No current transaction batch"})
		}
		if !transactionBatch.CanView {
			result, err := c.model.TransactionBatchMinimal(context, transactionBatch.ID)
			if err != nil {
				return c.InternalServerError(ctx, err)
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionBatchManager.ToModel(transactionBatch))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-batch/:transaction_batch_id/view-request",
		Method:   "PUT",
		Response: "ITransactionBatch",
		Note:     "Submit a request to view (blotter) a specific transaction batch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.TransactionBatchEndRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		transactionBatchId, err := horizon.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return err
		}
		transactionBatch, err := c.model.TransactionBatchManager.GetByID(context, *transactionBatchId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "No current transaction batch"})
		}
		now := time.Now().UTC()
		transactionBatch.RequestView = &now
		transactionBatch.CanView = false
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update transaction batch: "+err.Error())
		}
		if !transactionBatch.CanView {
			result, err := c.model.TransactionBatchMinimal(context, transactionBatch.ID)
			if err != nil {
				return c.InternalServerError(ctx, err)
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionBatchManager.ToModel(transactionBatch))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/transaction-batch/view-request",
		Method: "GET",
		Note:   "List all pending view (blotter) requests for transaction batches",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		transactionBatch, err := c.model.TransactionBatchManager.FindWithConditions(context, map[string]interface{}{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"can_view":        false,
			"is_closed":       false,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionBatchManager.ToModels(transactionBatch))
	})
	req.RegisterRoute(horizon.Route{
		Route:  "/transaction-batch/:transaction_batch_id/view-accept",
		Method: "PUT",
		Note:   "Accept a view (blotter) request for a transaction batch by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get transaction batch ID from URL parameter
		transactionBatchId, err := horizon.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return err
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}

		// Check authorization
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Get the transaction batch by ID
		transactionBatch, err := c.model.TransactionBatchManager.GetByID(context, *transactionBatchId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction batch not found"})
		}

		// Verify the transaction batch belongs to the user's organization and branch
		if transactionBatch.OrganizationID != userOrg.OrganizationID ||
			transactionBatch.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "Transaction batch not found in your organization/branch")
		}

		// Update CanView to true
		transactionBatch.CanView = true

		// Update the transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the updated transaction batch
		return ctx.JSON(http.StatusOK, c.model.TransactionBatchManager.ToModel(transactionBatch))
	})
}

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
		transactionBatch, err := c.model.TransactionBatchManager.FindOne(context, &model.TransactionBatch{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			IsClosed:       false,
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
		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]interface{}{
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

func (c *Controller) BatchFundingController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/batch-funding",
		Method:   "POST",
		Request:  "IBatchFunding",
		Response: "IBatchFunding",
		Note:     "Sart: create batch funding based on current transaction batch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		batchFundingReq, err := c.model.BatchFundingManager.Validate(ctx)
		if err != nil {
			return err
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		transactionBatch, err := c.model.TransactionBatchManager.FindOne(context, &model.TransactionBatch{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			IsClosed:       false,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if transactionBatch == nil {
			return c.BadRequest(ctx, "No active transaction batch found")
		}

		cashCounts, err := c.model.CashCountManager.Find(context, &model.CashCount{
			TransactionBatchID: transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total cash count value
		var totalCashCount float64
		for _, cashCount := range cashCounts {
			totalCashCount += cashCount.Amount * float64(cashCount.Quantity)
		}

		transactionBatch.BeginningBalance += batchFundingReq.Amount
		transactionBatch.TotalCashHandled = batchFundingReq.Amount + transactionBatch.DepositInBank + totalCashCount
		transactionBatch.CashCountTotal = totalCashCount
		transactionBatch.GrandTotal = totalCashCount + transactionBatch.DepositInBank

		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
		}
		batchFunding := &model.BatchFunding{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TransactionBatchID: transactionBatch.ID,

			ProvidedByUserID: userOrg.UserID,
			Name:             batchFundingReq.Name,
			Description:      batchFundingReq.Description,
			Amount:           batchFundingReq.Amount,
			SignatureMediaID: batchFundingReq.SignatureMediaID,
		}

		if err := c.model.BatchFundingManager.Create(context, batchFunding); err != nil {
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.BatchFundingManager.ToModel(batchFunding))

	})

	req.RegisterRoute(horizon.Route{
		Route:    "/batch-funding/transaction-batch/:transaction_batch_id/search",
		Method:   "GET",
		Request:  "Filter<IBatchFunding>",
		Response: "Paginated<IBatchFunding>",
		Note:     "Get all batch funding of transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get transaction batch ID from URL parameter
		transactionBatchId, err := horizon.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid transaction batch ID")
		}

		// Get current user organization for authorization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		// Check if user is authorized
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Verify the transaction batch exists and belongs to the user's organization/branch
		transactionBatch, err := c.model.TransactionBatchManager.GetByID(context, *transactionBatchId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction batch not found"})
		}

		// Verify the transaction batch belongs to the current user's organization and branch
		if transactionBatch.OrganizationID != userOrg.OrganizationID || transactionBatch.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this transaction batch"})
		}

		// Find all batch funding records for the transaction batch
		batchFunding, err := c.model.BatchFundingManager.Find(context, &model.BatchFunding{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TransactionBatchID: *transactionBatchId,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.BatchFundingManager.Pagination(context, ctx, batchFunding))
	})
}

func (c *Controller) CheckRemittanceController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/check-remittance",
		Method:   "GET",
		Response: "ICheckRemittance[]",
		Note:     "Retrieve batch check remittance (JWT) for the current transaction batch before ending.",
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
		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]interface{}{
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

		// Retrieve check remittance for the current transaction batch
		checkRemittance, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.ToModels(checkRemittance))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/check-remittance",
		Method:   "POST",
		Response: "ICheckRemittance",
		Request:  "ICheckRemittance",
		Note:     "Create a new check remittance for the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Validate the check remittance request
		req, err := c.model.CheckRemittanceManager.Validate(ctx)
		if err != nil {
			return err
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Find the current active transaction batch
		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]interface{}{
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

		// Set required fields for the check remittance
		checkRemittance := &model.CheckRemittance{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,

			BankID:             req.BankID,
			MediaID:            req.MediaID,
			EmployeeUserID:     &userOrg.UserID,
			TransactionBatchID: &transactionBatch.ID,
			CountryCode:        req.CountryCode,
			ReferenceNumber:    req.ReferenceNumber,
			AccountName:        req.AccountName,
			Amount:             req.Amount,
			DateEntry:          req.DateEntry,
			Description:        req.Description,
		}

		// Set default date entry if not provided
		if checkRemittance.DateEntry == nil {
			now := time.Now().UTC()
			checkRemittance.DateEntry = &now
		}

		// Create the check remittance
		if err := c.model.CheckRemittanceManager.Create(context, checkRemittance); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create check remittance: " + err.Error()})
		}

		// Get all check remittances for recalculating transaction batch totals
		allCheckRemittances, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total check remittance amount
		var totalCheckRemittance float64
		for _, remittance := range allCheckRemittances {
			totalCheckRemittance += remittance.Amount
		}

		// Update transaction batch totals
		transactionBatch.TotalCheckRemittance = totalCheckRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		// Save the updated transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the created check remittance
		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.ToModel(checkRemittance))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/check-remittance/:check_remittance_id",
		Method:   "PUT",
		Response: "ICheckRemittance",
		Request:  "ICheckRemittance",
		Note:     "Update an existing check remittance by ID for the current transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get check remittance ID from URL parameter
		checkRemittanceId, err := horizon.EngineUUIDParam(ctx, "check_remittance_id")
		if err != nil {
			return err
		}

		// Validate the check remittance request
		req, err := c.model.CheckRemittanceManager.Validate(ctx)
		if err != nil {
			return err
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Get the existing check remittance
		existingCheckRemittance, err := c.model.CheckRemittanceManager.GetByID(context, *checkRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Check remittance not found"})
		}

		// Verify ownership
		if existingCheckRemittance.OrganizationID != userOrg.OrganizationID ||
			existingCheckRemittance.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "Check remittance not found in your organization/branch")
		}

		// Find the current active transaction batch
		transactionBatch, err := c.model.TransactionBatchManager.FindOne(context, &model.TransactionBatch{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			IsClosed:       false,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if transactionBatch == nil {
			return c.BadRequest(ctx, "No active transaction batch found")
		}

		// Update the check remittance fields
		updatedCheckRemittance := &model.CheckRemittance{
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,

			BankID:          req.BankID,
			MediaID:         req.MediaID,
			CountryCode:     req.CountryCode,
			ReferenceNumber: req.ReferenceNumber,
			AccountName:     req.AccountName,
			Amount:          req.Amount,
			DateEntry:       req.DateEntry,
			Description:     req.Description,
		}

		// Set default date entry if not provided
		if updatedCheckRemittance.DateEntry == nil {
			now := time.Now().UTC()
			updatedCheckRemittance.DateEntry = &now
		}

		// Update the check remittance
		if err := c.model.CheckRemittanceManager.UpdateFields(context, *checkRemittanceId, updatedCheckRemittance); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update check remittance: " + err.Error()})
		}

		// Get all check remittances for recalculating transaction batch totals
		allCheckRemittances, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total check remittance amount
		var totalCheckRemittance float64
		for _, remittance := range allCheckRemittances {
			totalCheckRemittance += remittance.Amount
		}

		// Update transaction batch totals
		transactionBatch.TotalCheckRemittance = totalCheckRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		// Save the updated transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the updated check remittance
		updatedRemittance, err := c.model.CheckRemittanceManager.GetByID(context, *checkRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated check remittance: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.ToModel(updatedRemittance))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/check-remittance/:check_remittance_id",
		Method:   "DELETE",
		Response: "ICheckRemittance",
		Note:     "Delete an existing check remittance by ID for the current transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get check remittance ID from URL parameter
		checkRemittanceId, err := horizon.EngineUUIDParam(ctx, "check_remittance_id")
		if err != nil {
			return err
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Get the existing check remittance
		existingCheckRemittance, err := c.model.CheckRemittanceManager.GetByID(context, *checkRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Check remittance not found"})
		}

		// Verify ownership
		if existingCheckRemittance.OrganizationID != userOrg.OrganizationID ||
			existingCheckRemittance.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "Check remittance not found in your organization/branch")
		}

		// Find the current active transaction batch
		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]interface{}{
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

		// Verify the remittance belongs to the current transaction batch
		if existingCheckRemittance.TransactionBatchID == nil ||
			*existingCheckRemittance.TransactionBatchID != transactionBatch.ID {
			return c.BadRequest(ctx, "Check remittance does not belong to current transaction batch")
		}

		// Delete the check remittance
		if err := c.model.CheckRemittanceManager.DeleteByID(context, *checkRemittanceId); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete check remittance: " + err.Error()})
		}

		// Get all remaining check remittances for recalculating transaction batch totals
		allCheckRemittances, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total check remittance amount
		var totalCheckRemittance float64
		for _, remittance := range allCheckRemittances {
			totalCheckRemittance += remittance.Amount
		}

		// Update transaction batch totals
		transactionBatch.TotalCheckRemittance = totalCheckRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		// Save the updated transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the deleted check remittance
		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.ToModel(existingCheckRemittance))
	})
}

func (c *Controller) OnlineRemittanceController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/online-remittance",
		Method:   "GET",
		Response: "IOnlineRemittance[]",
		Note:     "Retrieve batch online remittance (JWT) for the current transaction batch before ending.",
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
		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]interface{}{
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

		// Retrieve online remittance for the current transaction batch
		onlineRemittance, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModels(onlineRemittance))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/online-remittance",
		Method:   "POST",
		Response: "IOnlineRemittance",
		Request:  "IOnlineRemittance",
		Note:     "Create a new online remittance for the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Validate the online remittance request
		req, err := c.model.OnlineRemittanceManager.Validate(ctx)
		if err != nil {
			return err
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Find the current active transaction batch
		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]interface{}{
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

		// Set required fields for the online remittance
		onlineRemittance := &model.OnlineRemittance{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,

			BankID:             req.BankID,
			MediaID:            req.MediaID,
			EmployeeUserID:     &userOrg.UserID,
			TransactionBatchID: &transactionBatch.ID,
			CountryCode:        req.CountryCode,
			ReferenceNumber:    req.ReferenceNumber,
			AccountName:        req.AccountName,
			Amount:             req.Amount,
			DateEntry:          req.DateEntry,
			Description:        req.Description,
		}

		// Set default date entry if not provided
		if onlineRemittance.DateEntry == nil {
			now := time.Now().UTC()
			onlineRemittance.DateEntry = &now
		}

		// Create the online remittance
		if err := c.model.OnlineRemittanceManager.Create(context, onlineRemittance); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create online remittance: " + err.Error()})
		}

		// Get all online remittances for recalculating transaction batch totals
		allOnlineRemittances, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total online remittance amount
		var totalOnlineRemittance float64
		for _, remittance := range allOnlineRemittances {
			totalOnlineRemittance += remittance.Amount
		}

		// Update transaction batch totals
		transactionBatch.TotalOnlineRemittance = totalOnlineRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		// Save the updated transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the created online remittance
		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModel(onlineRemittance))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/online-remittance/:online_remittance_id",
		Method:   "PUT",
		Response: "IOnlineRemittance",
		Request:  "IOnlineRemittance",
		Note:     "Update an existing online remittance by ID for the current transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get online remittance ID from URL parameter
		onlineRemittanceId, err := horizon.EngineUUIDParam(ctx, "online_remittance_id")
		if err != nil {
			return err
		}

		// Validate the online remittance request
		req, err := c.model.OnlineRemittanceManager.Validate(ctx)
		if err != nil {
			return err
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Get the existing online remittance
		existingOnlineRemittance, err := c.model.OnlineRemittanceManager.GetByID(context, *onlineRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Online remittance not found"})
		}

		// Verify ownership
		if existingOnlineRemittance.OrganizationID != userOrg.OrganizationID ||
			existingOnlineRemittance.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "Online remittance not found in your organization/branch")
		}

		// Find the current active transaction batch
		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]interface{}{
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

		// Update the online remittance fields
		updatedOnlineRemittance := &model.OnlineRemittance{
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,

			BankID:          req.BankID,
			MediaID:         req.MediaID,
			CountryCode:     req.CountryCode,
			ReferenceNumber: req.ReferenceNumber,
			AccountName:     req.AccountName,
			Amount:          req.Amount,
			DateEntry:       req.DateEntry,
			Description:     req.Description,
		}

		// Set default date entry if not provided
		if updatedOnlineRemittance.DateEntry == nil {
			now := time.Now().UTC()
			updatedOnlineRemittance.DateEntry = &now
		}

		// Update the online remittance
		if err := c.model.OnlineRemittanceManager.UpdateFields(context, *onlineRemittanceId, updatedOnlineRemittance); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update online remittance: " + err.Error()})
		}

		// Get all online remittances for recalculating transaction batch totals
		allOnlineRemittances, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total online remittance amount
		var totalOnlineRemittance float64
		for _, remittance := range allOnlineRemittances {
			totalOnlineRemittance += remittance.Amount
		}

		// Update transaction batch totals
		transactionBatch.TotalOnlineRemittance = totalOnlineRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		// Save the updated transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the updated online remittance
		updatedRemittance, err := c.model.OnlineRemittanceManager.GetByID(context, *onlineRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated online remittance: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModel(updatedRemittance))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/online-remittance/:online_remittance_id",
		Method:   "DELETE",
		Response: "IOnlineRemittance",
		Note:     "Delete an existing online remittance by ID for the current transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get online remittance ID from URL parameter
		onlineRemittanceId, err := horizon.EngineUUIDParam(ctx, "online_remittance_id")
		if err != nil {
			return err
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Get the existing online remittance
		existingOnlineRemittance, err := c.model.OnlineRemittanceManager.GetByID(context, *onlineRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Online remittance not found"})
		}

		// Verify ownership
		if existingOnlineRemittance.OrganizationID != userOrg.OrganizationID ||
			existingOnlineRemittance.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "Online remittance not found in your organization/branch")
		}

		// Find the current active transaction batch
		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]interface{}{
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

		// Verify the remittance belongs to the current transaction batch
		if existingOnlineRemittance.TransactionBatchID == nil ||
			*existingOnlineRemittance.TransactionBatchID != transactionBatch.ID {
			return c.BadRequest(ctx, "Online remittance does not belong to current transaction batch")
		}

		// Delete the online remittance
		if err := c.model.OnlineRemittanceManager.DeleteByID(context, *onlineRemittanceId); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete online remittance: " + err.Error()})
		}

		// Get all remaining online remittances for recalculating transaction batch totals
		allOnlineRemittances, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total online remittance amount
		var totalOnlineRemittance float64
		for _, remittance := range allOnlineRemittances {
			totalOnlineRemittance += remittance.Amount
		}

		// Update transaction batch totals
		transactionBatch.TotalOnlineRemittance = totalOnlineRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		// Save the updated transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the deleted online remittance
		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModel(existingOnlineRemittance))
	})
}
