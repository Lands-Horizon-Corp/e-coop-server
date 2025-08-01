package controller_v1

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

// BatchFundingController handles creation and retrieval of batch funding records with proper error handling and authorization checks.
func (c *Controller) BatchFundingController() {
	req := c.provider.Service.Request

	// POST /batch-funding: Create a new batch funding for the current open transaction batch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/batch-funding",
		Method:       "POST",
		Note:         "Creates a new batch funding for the currently active transaction batch of the user's organization and branch. Also updates the related transaction batch balances.",
		RequestType:  model.BatchFundingRequest{},
		ResponseType: model.BatchFundingResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		batchFundingReq, err := c.model.BatchFundingManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), validation error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid batch funding data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), user org error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unable to determine user organization. Please login again."})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for batch funding (/batch-funding)",
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You do not have permission to create batch funding."})
		}
		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), transaction batch lookup error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not find an active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), no open transaction batch.",
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch is open for this branch."})
		}

		cashCounts, err := c.model.CashCountManager.Find(context, &model.CashCount{
			TransactionBatchID: transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), cash count lookup error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to retrieve cash counts: " + err.Error()})
		}

		var totalCashCount float64
		for _, cashCount := range cashCounts {
			totalCashCount += cashCount.Amount * float64(cashCount.Quantity)
		}

		transactionBatch.BeginningBalance += batchFundingReq.Amount
		transactionBatch.TotalCashHandled = batchFundingReq.Amount + transactionBatch.DepositInBank + totalCashCount
		transactionBatch.CashCountTotal = totalCashCount
		transactionBatch.GrandTotal = totalCashCount + transactionBatch.DepositInBank

		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), transaction batch update error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusConflict, map[string]string{"error": "Could not update transaction batch balances: " + err.Error()})
		}
		batchFunding := &model.BatchFunding{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TransactionBatchID: transactionBatch.ID,
			ProvidedByUserID:   userOrg.UserID,
			Name:               batchFundingReq.Name,
			Description:        batchFundingReq.Description,
			Amount:             batchFundingReq.Amount,
			SignatureMediaID:   batchFundingReq.SignatureMediaID,
		}

		if err := c.model.BatchFundingManager.Create(context, batchFunding); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), db error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusConflict, map[string]string{"error": "Unable to create batch funding record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created batch funding (/batch-funding): " + batchFunding.Name,
			Module:      "BatchFunding",
		})
		return ctx.JSON(http.StatusOK, c.model.BatchFundingManager.ToModel(batchFunding))
	})

	// GET /batch-funding/transaction-batch/:transaction_batch_id/search: Paginated batch funding for a transaction batch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/batch-funding/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		Note:         "Retrieves a paginated list of batch funding records for the specified transaction batch, if the user is authorized for the branch.",
		ResponseType: model.BatchFundingResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		transactionBatchId, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "The transaction batch ID provided is invalid."})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unable to determine user organization. Please login again."})
		}

		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You do not have permission to view batch funding records."})
		}

		transactionBatch, err := c.model.TransactionBatchManager.GetByID(context, *transactionBatchId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction batch not found for this ID."})
		}

		if transactionBatch.OrganizationID != userOrg.OrganizationID || transactionBatch.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this transaction batch. The batch does not belong to your organization or branch."})
		}

		batchFunding, err := c.model.BatchFundingManager.Find(context, &model.BatchFunding{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TransactionBatchID: *transactionBatchId,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to retrieve batch funding records: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.BatchFundingManager.Pagination(context, ctx, batchFunding))
	})
}
