package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) transactionBatchController() {
	req := c.provider.Service.Request

	// List all transaction batches for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction-batch",
		Method:       "GET",
		ResponseType: core.TransactionBatchResponse{},
		Note:         "Returns all transaction batches for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		transactionBatch, err := c.core.TransactionBatchManager.Find(context, &core.TransactionBatch{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batches: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.TransactionBatchManager.ToModels(transactionBatch))
	})

	// Paginate transaction batches for current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction-batch/search",
		Method:       "GET",
		ResponseType: core.TransactionBatchResponse{},
		Note:         "Returns paginated transaction batches for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactionBatch, err := c.core.TransactionBatchCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve paginated transaction batches: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.TransactionBatchManager.Pagination(context, ctx, transactionBatch))
	})

	// Update batch signatures for a transaction batch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction-batch/:transaction_batch_id/signature",
		Method:       "PUT",
		ResponseType: core.TransactionBatchResponse{},
		RequestType:  core.TransactionBatchSignatureRequest{},
		Note:         "Updates signature and position fields for a transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.TransactionBatchSignatureRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: invalid request body: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: validation error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: invalid transaction_batch_id: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction_batch_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: user org error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: user not authorized",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		transactionBatch, err := c.core.TransactionBatchManager.GetByID(context, *transactionBatchID)
		if err != nil || transactionBatch == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: Transaction batch not found",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction batch not found"})
		}

		// Update all signature fields
		transactionBatch.EmployeeBySignatureMediaID = req.EmployeeBySignatureMediaID
		transactionBatch.EmployeeByName = req.EmployeeByName
		transactionBatch.EmployeeByPosition = req.EmployeeByPosition
		transactionBatch.ApprovedBySignatureMediaID = req.ApprovedBySignatureMediaID
		transactionBatch.ApprovedByName = req.ApprovedByName
		transactionBatch.ApprovedByPosition = req.ApprovedByPosition
		transactionBatch.PreparedBySignatureMediaID = req.PreparedBySignatureMediaID
		transactionBatch.PreparedByName = req.PreparedByName
		transactionBatch.PreparedByPosition = req.PreparedByPosition
		transactionBatch.CertifiedBySignatureMediaID = req.CertifiedBySignatureMediaID
		transactionBatch.CertifiedByName = req.CertifiedByName
		transactionBatch.CertifiedByPosition = req.CertifiedByPosition
		transactionBatch.VerifiedBySignatureMediaID = req.VerifiedBySignatureMediaID
		transactionBatch.VerifiedByName = req.VerifiedByName
		transactionBatch.VerifiedByPosition = req.VerifiedByPosition
		transactionBatch.CheckBySignatureMediaID = req.CheckBySignatureMediaID
		transactionBatch.CheckByName = req.CheckByName
		transactionBatch.CheckByPosition = req.CheckByPosition
		transactionBatch.AcknowledgeBySignatureMediaID = req.AcknowledgeBySignatureMediaID
		transactionBatch.AcknowledgeByName = req.AcknowledgeByName
		transactionBatch.AcknowledgeByPosition = req.AcknowledgeByPosition
		transactionBatch.NotedBySignatureMediaID = req.NotedBySignatureMediaID
		transactionBatch.NotedByName = req.NotedByName
		transactionBatch.NotedByPosition = req.NotedByPosition
		transactionBatch.PostedBySignatureMediaID = req.PostedBySignatureMediaID
		transactionBatch.PostedByName = req.PostedByName
		transactionBatch.PostedByPosition = req.PostedByPosition
		transactionBatch.PaidBySignatureMediaID = req.PaidBySignatureMediaID
		transactionBatch.PaidByName = req.PaidByName
		transactionBatch.PaidByPosition = req.PaidByPosition

		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		if err := c.core.TransactionBatchManager.UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: update error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated transaction batch signatures for batch " + transactionBatch.ID.String(),
			Module:      "TransactionBatch",
		})
		return ctx.JSON(http.StatusOK, c.core.TransactionBatchManager.ToModel(transactionBatch))
	})

	// Get the current active transaction batch for the user
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction-batch/current",
		Method:       "GET",
		ResponseType: core.TransactionBatchResponse{},
		Note:         "Returns the current active transaction batch for the current user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		transactionBatch, err := c.core.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil || transactionBatch == nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		if !transactionBatch.CanView {
			result, err := c.core.TransactionBatchMinimal(context, transactionBatch.ID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get minimal transaction batch: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, c.core.TransactionBatchManager.ToModel(transactionBatch))
	})

	// Update deposit in bank amount for a specific transaction batch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction-batch/:transaction_batch_id/deposit-in-bank",
		Method:       "PUT",
		ResponseType: core.TransactionBatchResponse{},
		RequestType:  core.BatchFundingRequest{},
		Note:         "Updates the deposit in bank amount for a specific transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: invalid transaction_batch_id: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction_batch_id: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: user org error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: user not authorized",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		type DepositInBankRequest struct {
			DepositInBank float64 `json:"deposit_in_bank" validate:"min=0"`
		}
		var req DepositInBankRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: invalid request body: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: validation error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		transactionBatch, err := c.core.TransactionBatchManager.GetByID(context, *transactionBatchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: transaction batch not found: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction batch not found: " + err.Error()})
		}

		if transactionBatch.OrganizationID != userOrg.OrganizationID || transactionBatch.BranchID != *userOrg.BranchID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: batch not in org/branch",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Transaction batch not found in your organization/branch"})
		}

		if transactionBatch.IsClosed {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: batch is closed",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update deposit for a closed transaction batch"})
		}

		cashCounts, err := c.core.CashCountManager.Find(context, &core.CashCount{
			TransactionBatchID: transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: get cash counts error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash counts: " + err.Error()})
		}

		var totalCashCount float64
		for _, cashCount := range cashCounts {
			totalCashCount += cashCount.Amount
		}

		transactionBatch.DepositInBank = req.DepositInBank
		transactionBatch.GrandTotal = totalCashCount + req.DepositInBank
		transactionBatch.TotalCashHandled = transactionBatch.BeginningBalance + req.DepositInBank + totalCashCount
		transactionBatch.TotalDepositInBank = req.DepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		if err := c.core.TransactionBatchManager.UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: update error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated deposit in bank for batch " + transactionBatch.ID.String(),
			Module:      "TransactionBatch",
		})

		if !transactionBatch.CanView {
			result, err := c.core.TransactionBatchMinimal(context, transactionBatch.ID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get minimal transaction batch: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, c.core.TransactionBatchManager.ToModel(transactionBatch))
	})

	// Create a new transaction batch and batch funding
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction-batch",
		Method:       "POST",
		ResponseType: core.TransactionBatchResponse{},
		RequestType:  core.TransactionBatchRequest{},
		Note:         "Creates and starts a new transaction batch for the current branch (will also populate cash count).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		batchFundingReq, err := c.core.BatchFundingManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: validation error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: user org error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: user not authorized",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		transactionBatch, _ := c.core.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if transactionBatch != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: ongoing batch",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusConflict, map[string]string{"error": "There is an ongoing transaction batch"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: begin tx error: " + tx.Error.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}
		transBatch := &core.TransactionBatch{
			CreatedAt:                     time.Now().UTC(),
			CreatedByID:                   userOrg.UserID,
			UpdatedAt:                     time.Now().UTC(),
			UpdatedByID:                   userOrg.UserID,
			OrganizationID:                userOrg.OrganizationID,
			BranchID:                      *userOrg.BranchID,
			EmployeeUserID:                &userOrg.UserID,
			CurrencyID:                    batchFundingReq.CurrencyID,
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
			BatchName:                     batchFundingReq.Name,
			IsClosed:                      false,
			CanView:                       false,
			RequestView:                   false,
		}
		if err := c.core.TransactionBatchManager.CreateWithTx(context, tx, transBatch); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: create error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create transaction batch: " + err.Error()})
		}
		batchFunding := &core.BatchFunding{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TransactionBatchID: transBatch.ID,
			ProvidedByUserID:   userOrg.UserID,
			Name:               batchFundingReq.Name,
			Description:        batchFundingReq.Description,
			Amount:             batchFundingReq.Amount,
			SignatureMediaID:   batchFundingReq.SignatureMediaID,
			CurrencyID:         batchFundingReq.CurrencyID,
		}
		if err := c.core.BatchFundingManager.CreateWithTx(context, tx, batchFunding); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: create batch funding error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create batch funding: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: commit tx error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created transaction batch and batch funding for branch " + userOrg.BranchID.String(),
			Module:      "TransactionBatch",
		})
		result, err := c.core.TransactionBatchMinimal(context, transBatch.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve minimal transaction batch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, result)
	})

	// End the current transaction batch for the authenticated user
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction-batch/end",
		Method:       "PUT",
		RequestType:  core.TransactionBatchEndRequest{},
		ResponseType: core.TransactionBatchResponse{},
		Note:         "Ends the current transaction batch for the authenticated user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.TransactionBatchEndRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "End transaction batch failed: invalid request body: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "End transaction batch failed: validation error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "End transaction batch failed: user org error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "End transaction batch failed: user not authorized",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		transactionBatch, err := c.core.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "End transaction batch failed: retrieve error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "End transaction batch failed: no active batch",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No active transaction batch found"})
		}
		now := time.Now().UTC()
		transactionBatch.IsClosed = true
		transactionBatch.EmployeeUserID = &userOrg.UserID
		transactionBatch.EmployeeBySignatureMediaID = req.EmployeeBySignatureMediaID
		transactionBatch.EmployeeByName = req.EmployeeByName
		transactionBatch.EmployeeByPosition = req.EmployeeByPosition
		transactionBatch.EndedAt = &now
		if err := c.core.TransactionBatchManager.UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "End transaction batch failed: update error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Ended transaction batch for branch " + userOrg.BranchID.String(),
			Module:      "TransactionBatch",
		})

		if !transactionBatch.CanView {
			result, err := c.core.TransactionBatchMinimal(context, transactionBatch.ID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve minimal transaction batch: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, c.core.TransactionBatchManager.ToModel(transactionBatch))
	})

	// Retrieve a transaction batch by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction-batch/:transaction_batch_id",
		Method:       "GET",
		Note:         "Returns a transaction batch by its ID.",
		ResponseType: core.TransactionBatchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction_batch_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		transactionBatch, err := c.core.TransactionBatchManager.GetByID(context, *transactionBatchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction batch not found: " + err.Error()})
		}
		if !transactionBatch.CanView {
			result, err := c.core.TransactionBatchMinimal(context, transactionBatch.ID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve minimal transaction batch: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, c.core.TransactionBatchManager.ToModel(transactionBatch))
	})

	// Submit a request to view (blotter) a specific transaction batch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction-batch/:transaction_batch_id/view-request",
		Method:       "PUT",
		RequestType:  core.TransactionBatchEndRequest{},
		ResponseType: core.TransactionBatchResponse{},
		Note:         "Submits a request to view (blotter) a specific transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.TransactionBatchEndRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View request failed: invalid request body: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change request failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View request failed: user org error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View request failed: user not authorized",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View request failed: invalid transaction_batch_id: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction_batch_id: " + err.Error()})
		}
		transactionBatch, err := c.core.TransactionBatchManager.GetByID(context, *transactionBatchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View request failed: batch not found: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction batch not found: " + err.Error()})
		}
		transactionBatch.RequestView = true
		transactionBatch.CanView = false
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID
		if err := c.core.TransactionBatchManager.UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View request failed: update error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Requested view for transaction batch " + transactionBatch.ID.String(),
			Module:      "TransactionBatch",
		})
		if !transactionBatch.CanView {
			result, err := c.core.TransactionBatchMinimal(context, transactionBatch.ID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve minimal transaction batch: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, c.core.TransactionBatchManager.ToModel(transactionBatch))
	})

	// List all pending view (blotter) requests for transaction batches
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction-batch/view-request",
		Method:       "GET",
		Note:         "Returns all pending view (blotter) requests for transaction batches on the current branch.",
		ResponseType: core.TransactionBatchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		transactionBatch, err := c.core.TransactionBatchViewRequests(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve pending view requests: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.TransactionBatchManager.ToModels(transactionBatch))
	})

	// List all ended (closed) batches for the current day
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction-batch/ended-batch",
		Method:       "GET",
		Note:         "Returns all ended (closed) transaction batches for the current day.",
		ResponseType: core.TransactionBatchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		batches, err := c.core.TransactionBatchCurrentDay(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ended transaction batches: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.TransactionBatchManager.ToModels(batches))
	})

	// Accept a view (blotter) request for a transaction batch by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction-batch/:transaction_batch_id/view-accept",
		Method:       "PUT",
		Note:         "Accepts a view (blotter) request for a transaction batch by its ID.",
		ResponseType: core.TransactionBatchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Accept view request failed: invalid transaction_batch_id: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction_batch_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Accept view request failed: user org error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Accept view request failed: user not authorized",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		transactionBatch, err := c.core.TransactionBatchManager.GetByID(context, *transactionBatchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Accept view request failed: batch not found: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction batch not found: " + err.Error()})
		}

		if transactionBatch.OrganizationID != userOrg.OrganizationID || transactionBatch.BranchID != *userOrg.BranchID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Accept view request failed: batch not in org/branch",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Transaction batch not found in your organization/branch"})
		}

		transactionBatch.CanView = true

		if err := c.core.TransactionBatchManager.UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Accept view request failed: update error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Accepted view request for transaction batch " + transactionBatch.ID.String(),
			Module:      "TransactionBatch",
		})

		return ctx.JSON(http.StatusOK, c.core.TransactionBatchManager.ToModel(transactionBatch))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction-batch/employee/:user_organization_id/search",
		Method:       "GET",
		ResponseType: core.TransactionBatchResponse{},
		Note:         "Returns transaction batches for a specific employee (user_id) in the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}
		userOrganization, err := c.core.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}

		batches, err := c.core.TransactionBatchManager.Find(context, &core.TransactionBatch{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			EmployeeUserID: &userOrganization.UserID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batches: " + err.Error()})
		}
		paginated := c.core.TransactionBatchManager.Pagination(context, ctx, batches)

		// ✅ Fix: Use index to properly update the slice
		for i, batch := range paginated.Data {
			if !batch.CanView {
				minimalBatch, err := c.core.TransactionBatchMinimal(context, batch.ID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve minimal transaction batch: " + err.Error()})
				}
				paginated.Data[i] = minimalBatch
			}
		}
		return ctx.JSON(http.StatusOK, paginated)
	})

}
