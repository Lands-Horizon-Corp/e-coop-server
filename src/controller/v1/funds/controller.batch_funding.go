package funds

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func BatchFundingController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/batch-funding",
		Method:       "POST",
		Note:         "Creates a new batch funding for the currently active transaction batch of the user's organization and branch. Also updates the related transaction batch balances.",
		RequestType:  types.BatchFundingRequest{},
		ResponseType: types.BatchFundingResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		batchFundingReq, err := core.BatchFundingManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), validation error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid batch funding data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), user org error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unable to determine user organization. Please login again."})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for batch funding (/batch-funding)",
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You do not have permission to create batch funding."})
		}
		transactionBatch, err := core.TransactionBatchCurrent(
			context, service,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), transaction batch lookup error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not find an active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), no open transaction batch.",
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch is open for this branch."})
		}

		batchFunding := &types.BatchFunding{
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

		if err := core.BatchFundingManager(service).Create(context, batchFunding); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Batch funding creation failed (/batch-funding), db error: " + err.Error(),
				Module:      "BatchFunding",
			})
			return ctx.JSON(http.StatusConflict, map[string]string{"error": "Unable to create batch funding record: " + err.Error()})
		}
		if err := event.TransactionBatchBalancing(context, service, &transactionBatch.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch after saving: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created batch funding (/batch-funding): " + batchFunding.Name,
			Module:      "BatchFunding",
		})
		return ctx.JSON(http.StatusOK, core.BatchFundingManager(service).ToModel(batchFunding))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/batch-funding/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		Note:         "Retrieves a paginated list of batch funding records for the specified transaction batch, if the user is authorized for the branch.",
		ResponseType: types.BatchFundingResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "The transaction batch ID provided is invalid."})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unable to determine user organization. Please login again."})
		}

		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You do not have permission to view batch funding records."})
		}

		transactionBatch, err := core.TransactionBatchManager(service).GetByID(context, *transactionBatchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction batch not found for this ID."})
		}

		if transactionBatch.OrganizationID != userOrg.OrganizationID || transactionBatch.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this transaction batch. The batch does not belong to your organization or branch."})
		}

		batchFunding, err := core.BatchFundingManager(service).NormalPagination(context, ctx, &types.BatchFunding{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TransactionBatchID: *transactionBatchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to retrieve batch funding records: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, batchFunding)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/batch-funding/search",
		Method:       "GET",
		ResponseType: types.BatchFundingResponse{},
		Note:         "Returns all batch funding records for the current user's organization and branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view batch funding records"})
		}

		batchFundings, err := core.BatchFundingManager(service).NormalPagination(context, ctx, &types.BatchFunding{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve batch funding records: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, batchFundings)
	})
}
