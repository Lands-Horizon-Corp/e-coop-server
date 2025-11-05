package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// BatchFundingController handles creation and retrieval of batch funding records with proper error handling and authorization checks.
func (c *Controller) batchFundingController() {
	req := c.provider.Service.Request

	// POST /batch-funding: Create a new batch funding for the current open transaction batch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/batch-funding",
		Method:       "POST",
		Note:         "Creates a new batch funding for the currently active transaction batch of the user's organization and branch. Also updates the related transaction batch balances.",
		RequestType:  core.BatchFundingRequest{},
		ResponseType: core.BatchFundingResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		batchFundingReq, err := c.core.BatchFundingManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), validation error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid batch funding data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), user org error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unable to determine user organization. Please login again."})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for batch funding (/batch-funding)",
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You do not have permission to create batch funding."})
		}
		transactionBatch, err := c.core.CurrentOpenTransactionBatch(
			context,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), transaction batch lookup error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not find an active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), no open transaction batch.",
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch is open for this branch."})
		}

		cashCounts, err := c.core.CashCountManager.Find(context, &core.CashCount{
			TransactionBatchID: transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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

		if err := c.core.TransactionBatchManager.UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), transaction batch update error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusConflict, map[string]string{"error": "Could not update transaction batch balances: " + err.Error()})
		}
		batchFunding := &core.BatchFunding{
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
			CurrencyID:         batchFundingReq.CurrencyID,
		}

		if err := c.core.BatchFundingManager.Create(context, batchFunding); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), db error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusConflict, map[string]string{"error": "Unable to create batch funding record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created batch funding (/batch-funding): " + batchFunding.Name,
			Module:      "BatchFunding",
		})
		return ctx.JSON(http.StatusOK, c.core.BatchFundingManager.ToModel(batchFunding))
	})

	// GET /batch-funding/transaction-batch/:transaction_batch_id/search: Paginated batch funding for a transaction batch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/batch-funding/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		Note:         "Retrieves a paginated list of batch funding records for the specified transaction batch, if the user is authorized for the branch.",
		ResponseType: core.BatchFundingResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "The transaction batch ID provided is invalid."})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unable to determine user organization. Please login again."})
		}

		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You do not have permission to view batch funding records."})
		}

		transactionBatch, err := c.core.TransactionBatchManager.GetByID(context, *transactionBatchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction batch not found for this ID."})
		}

		if transactionBatch.OrganizationID != userOrg.OrganizationID || transactionBatch.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this transaction batch. The batch does not belong to your organization or branch."})
		}

		batchFunding, err := c.core.BatchFundingManager.PaginationWithFields(context, ctx, &core.BatchFunding{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TransactionBatchID: *transactionBatchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to retrieve batch funding records: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, batchFunding)
	})

	// GET /api/v1/batch-funding/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/batch-funding/search",
		Method:       "GET",
		ResponseType: core.BatchFundingResponse{},
		Note:         "Returns all batch funding records for the current user's organization and branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view batch funding records"})
		}

		batchFundings, err := c.core.BatchFundingManager.PaginationWithFields(context, ctx, &core.BatchFunding{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve batch funding records: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, batchFundings)
	})
}
