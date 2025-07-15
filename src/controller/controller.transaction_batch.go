package controller

import (
	"net/http"
	"time"

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
		Route:    "/transaction-batch/search",
		Method:   "GET",
		Request:  "Filter<TTransactionBatch>",
		Response: "Paginated<TTransactionBatch>",
		Note:     "Get pagination for transaction batches",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		transactionBatch, err := c.model.TransactionBatchCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionBatchManager.Pagination(context, ctx, transactionBatch))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-batch/:transaction_batch_id/signature",
		Method:   "PUT",
		Request:  "Filter<TTransactionBatch>",
		Response: "Paginated<TTransactionBatch>",
		Note:     "Get pagination for transaction batches",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.TransactionBatchSignatureRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
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
			return err
		}
		if transactionBatch == nil {
			return c.NotFound(ctx, "Transaction batch not found")
		}
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

		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionBatchManager.ToModel(transactionBatch))
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

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
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
		transactionBatch, _ := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
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
		now := time.Now().UTC()
		transactionBatch.IsClosed = true
		transactionBatch.EmployeeUserID = &userOrg.UserID
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
		transactionBatch, err := c.model.TransactionBatchManager.FindWithConditions(context, map[string]any{
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
		Route:  "/transaction-batch/ended-batch",
		Method: "GET",
		Note:   "List all approvals (blotter) requests for transaction batches from the current day",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		now := time.Now().UTC()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay := startOfDay.Add(24 * time.Hour)

		// Create conditions map without the operators in keys
		conditions := map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       true,
		}

		// Use GORM's DB.Where to handle date range filtering
		db := c.provider.Service.Database.Client().Model(new(model.TransactionBatch))
		// Add necessary preloads using array
		preloads := []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
			"EmployeeUser",
			"EmployeeUser.Media",
			"ApprovedBySignatureMedia",
			"PreparedBySignatureMedia",
			"CertifiedBySignatureMedia",
			"VerifiedBySignatureMedia",
			"CheckBySignatureMedia",
			"AcknowledgeBySignatureMedia",
			"NotedBySignatureMedia",
			"PostedBySignatureMedia",
			"PaidBySignatureMedia",
		}

		// Apply preloads
		for _, preload := range preloads {
			db = db.Preload(preload)
		}

		// Apply conditions
		for field, value := range conditions {
			db = db.Where(field+" = ?", value)
		}
		db = db.Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay)

		var transactionBatch []*model.TransactionBatch
		if err := db.Order("updated_at DESC").Find(&transactionBatch).Error; err != nil {
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
