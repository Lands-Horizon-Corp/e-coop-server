package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

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
