package transactions

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func TransactionBatchController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction-batch",
		Method:       "GET",
		ResponseType: types.TransactionBatchResponse{},
		Note:         "Returns all transaction batches for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		transactionBatch, err := core.TransactionBatchManager(service).Find(context, &types.TransactionBatch{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batches: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.TransactionBatchManager(service).ToModels(transactionBatch))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction-batch/search",
		Method:       "GET",
		ResponseType: types.TransactionBatchResponse{},
		Note:         "Returns paginated transaction batches for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactionBatch, err := core.TransactionBatchManager(service).NormalPagination(context, ctx, &types.TransactionBatch{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve paginated transaction batches: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, transactionBatch)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction-batch/:transaction_batch_id/signature",
		Method:       "PUT",
		ResponseType: types.TransactionBatchResponse{},
		RequestType:  types.TransactionBatchSignatureRequest{},
		Note:         "Updates signature and position fields for a transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.TransactionBatchSignatureRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: invalid request body: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: validation error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: invalid transaction_batch_id: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction_batch_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: user org error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: user not authorized",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		transactionBatch, err := core.TransactionBatchManager(service).GetByID(context, *transactionBatchID)
		if err != nil || transactionBatch == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: Transaction batch not found",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction batch not found"})
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

		if err := core.TransactionBatchManager(service).UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update signature failed: update error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated transaction batch signatures for batch " + transactionBatch.ID.String(),
			Module:      "TransactionBatch",
		})
		return ctx.JSON(http.StatusOK, core.TransactionBatchManager(service).ToModel(transactionBatch))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction-batch/current",
		Method:       "GET",
		ResponseType: types.TransactionBatchResponse{},
		Note:         "Returns the current active transaction batch for the current user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		transactionBatch, err := core.TransactionBatchCurrent(context, service, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil || transactionBatch == nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		if !transactionBatch.CanView {
			result, err := core.TransactionBatchMinimal(context, service, transactionBatch.ID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get minimal transaction batch: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, core.TransactionBatchManager(service).ToModel(transactionBatch))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction-batch/:transaction_batch_id/deposit-in-bank",
		Method:       "PUT",
		ResponseType: types.TransactionBatchResponse{},
		RequestType:  types.BatchFundingRequest{},
		Note:         "Updates the deposit in bank amount for a specific transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: invalid transaction_batch_id: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction_batch_id: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: user org error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: user not authorized",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		var req types.TransactionBatchDepositInBankRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: invalid request body: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: validation error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		transactionBatch, err := core.TransactionBatchManager(service).GetByID(context, *transactionBatchID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: transaction batch not found: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction batch not found: " + err.Error()})
		}

		if transactionBatch.OrganizationID != userOrg.OrganizationID || transactionBatch.BranchID != *userOrg.BranchID {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: batch not in org/branch",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Transaction batch not found in your organization/branch"})
		}

		if transactionBatch.IsClosed {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: batch is closed",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update deposit for a closed transaction batch"})
		}
		transactionBatch.DepositInBank = req.DepositInBank

		if err := core.TransactionBatchManager(service).UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update deposit in bank failed: update error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		if err := event.TransactionBatchBalancing(context, service, &transactionBatch.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch after saving: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated deposit in bank for batch " + transactionBatch.ID.String(),
			Module:      "TransactionBatch",
		})

		if !transactionBatch.CanView {
			result, err := core.TransactionBatchMinimal(context, service, transactionBatch.ID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get minimal transaction batch: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, core.TransactionBatchManager(service).ToModel(transactionBatch))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction-batch",
		Method:       "POST",
		ResponseType: types.TransactionBatchResponse{},
		RequestType:  types.TransactionBatchRequest{},
		Note:         "Creates and starts a new transaction batch for the current branch (will also populate cash count).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.BatchFundingManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: validation error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: user org error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		branchSetting, err := core.BranchSettingManager(service).FindOne(context, &types.BranchSetting{BranchID: *userOrg.BranchID})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: branch setting not found",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch settings not found for this branch"})
		}

		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: user not authorized",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		transactionBatch, _ := core.TransactionBatchCurrent(context, service, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if transactionBatch != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: ongoing batch",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusConflict, map[string]string{"error": "There is an ongoing transaction batch"})
		}
		unbalanced, err := core.UnbalancedAccountManager(service).FindOne(context, &types.UnbalancedAccount{
			CurrencyID:       req.CurrencyID,
			BranchSettingsID: branchSetting.ID,
		})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: unbalanced account not configured",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "Unbalanced account is not configured for this branch and currency"})
		}

		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: begin tx error: " + tx.Error.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + endTx(tx.Error).Error()})
		}

		transBatch := &types.TransactionBatch{
			CreatedAt:                     time.Now().UTC(),
			CreatedByID:                   userOrg.UserID,
			UpdatedAt:                     time.Now().UTC(),
			UpdatedByID:                   userOrg.UserID,
			OrganizationID:                userOrg.OrganizationID,
			BranchID:                      *userOrg.BranchID,
			EmployeeUserID:                &userOrg.UserID,
			CurrencyID:                    req.CurrencyID,
			BeginningBalance:              req.Amount,
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
			TotalSupposedRemmitance:       0,
			TotalCashOnHand:               0,
			TotalCheckRemittance:          0,
			TotalOnlineRemittance:         0,
			TotalDepositInBank:            0,
			TotalActualRemittance:         0,
			TotalActualSupposedComparison: 0,
			BatchName:                     req.Name,
			IsClosed:                      false,
			CanView:                       false,
			RequestView:                   false,
			UnbalancedAccountID:           unbalanced.ID,
		}
		if err := core.TransactionBatchManager(service).CreateWithTx(context, tx, transBatch); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: create error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create transaction batch: " + endTx(err).Error()})
		}

		batchFunding := &types.BatchFunding{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TransactionBatchID: transBatch.ID,
			ProvidedByUserID:   userOrg.UserID,
			Name:               req.Name,
			Description:        req.Description,
			Amount:             req.Amount,
			SignatureMediaID:   req.SignatureMediaID,
			CurrencyID:         req.CurrencyID,
		}

		if err := core.BatchFundingManager(service).CreateWithTx(context, tx, batchFunding); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: create batch funding error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create batch funding: " + endTx(err).Error()})
		}

		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: commit tx error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created transaction batch and batch funding for branch " + userOrg.BranchID.String(),
			Module:      "TransactionBatch",
		})

		result, err := core.TransactionBatchMinimal(context, service, transBatch.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve minimal transaction batch: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, result)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction-batch/end",
		Method:       "PUT",
		RequestType:  types.TransactionBatchEndRequest{},
		ResponseType: types.TransactionBatchResponse{},
		Note:         "Ends the current transaction batch for the authenticated user.",
	}, func(ctx echo.Context) error {
		c := ctx.Request().Context()
		var req types.TransactionBatchEndRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: eris.Wrap(err, "invalid request body").Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: eris.Wrap(err, "request validation failed").Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed"})
		}

		userOrg, err := event.CurrentUserOrganization(c, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: eris.Wrap(err, "failed to retrieve user organization").Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "user not authorized to end transaction batch",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		transactionBatch, err := event.TransactionBatchEnd(c, service, userOrg, &req)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to end transaction batch"})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Ended transaction batch for branch " + userOrg.BranchID.String(),
			Module:      "TransactionBatch",
		})
		if !transactionBatch.CanView {
			result, err := core.TransactionBatchMinimal(c, service, transactionBatch.ID)
			if err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{Activity: "update-error", Description: eris.Wrap(err, "failed to retrieve minimal transaction batch").Error(), Module: "TransactionBatch"})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batch"})
			}
			return ctx.JSON(http.StatusOK, result)
		}

		return ctx.JSON(http.StatusOK, core.TransactionBatchManager(service).ToModel(transactionBatch))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction-batch/:transaction_batch_id",
		Method:       "GET",
		Note:         "Returns a transaction batch by its ID.",
		ResponseType: types.TransactionBatchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction_batch_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		transactionBatch, err := core.TransactionBatchManager(service).GetByID(context, *transactionBatchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction batch not found: " + err.Error()})
		}
		if !transactionBatch.CanView {
			result, err := core.TransactionBatchMinimal(context, service, transactionBatch.ID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve minimal transaction batch: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, core.TransactionBatchManager(service).ToModel(transactionBatch))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction-batch/:transaction_batch_id/view-request",
		Method:       "PUT",
		RequestType:  types.TransactionBatchEndRequest{},
		ResponseType: types.TransactionBatchResponse{},
		Note:         "Submits a request to view (blotter) a specific transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.TransactionBatchEndRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View request failed: invalid request body: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change request failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View request failed: user org error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View request failed: user not authorized",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View request failed: invalid transaction_batch_id: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction_batch_id: " + err.Error()})
		}
		transactionBatch, err := core.TransactionBatchManager(service).GetByID(context, *transactionBatchID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.TransactionBatchManager(service).UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "View request failed: update error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Requested view for transaction batch " + transactionBatch.ID.String(),
			Module:      "TransactionBatch",
		})
		if !transactionBatch.CanView {
			result, err := core.TransactionBatchMinimal(context, service, transactionBatch.ID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve minimal transaction batch: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, result)
		}
		return ctx.JSON(http.StatusOK, core.TransactionBatchManager(service).ToModel(transactionBatch))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction-batch/view-request",
		Method:       "GET",
		Note:         "Returns all pending view (blotter) requests for transaction batches on the current branch.",
		ResponseType: types.TransactionBatchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		transactionBatch, err := core.TransactionBatchViewRequests(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve pending view requests: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.TransactionBatchManager(service).ToModels(transactionBatch))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction-batch/ended-batch",
		Method:       "GET",
		Note:         "Returns all ended (closed) transaction batches for the current day.",
		ResponseType: types.TransactionBatchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		batches, err := core.TransactionBatchCurrentDay(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ended transaction batches: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.TransactionBatchManager(service).ToModels(batches))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction-batch/:transaction_batch_id/view-accept",
		Method:       "PUT",
		Note:         "Accepts a view (blotter) request for a transaction batch by its ID.",
		ResponseType: types.TransactionBatchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Accept view request failed: invalid transaction_batch_id: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction_batch_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Accept view request failed: user org error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Accept view request failed: user not authorized",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		transactionBatch, err := core.TransactionBatchManager(service).GetByID(context, *transactionBatchID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Accept view request failed: batch not found: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction batch not found: " + err.Error()})
		}

		if transactionBatch.OrganizationID != userOrg.OrganizationID || transactionBatch.BranchID != *userOrg.BranchID {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Accept view request failed: batch not in org/branch",
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Transaction batch not found in your organization/branch"})
		}

		transactionBatch.CanView = true

		if err := core.TransactionBatchManager(service).UpdateByID(context, transactionBatch.ID, transactionBatch); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Accept view request failed: update error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Accepted view request for transaction batch " + transactionBatch.ID.String(),
			Module:      "TransactionBatch",
		})

		return ctx.JSON(http.StatusOK, core.TransactionBatchManager(service).ToModel(transactionBatch))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction-batch/employee/:user_organization_id/search",
		Method:       "GET",
		ResponseType: types.TransactionBatchResponse{},
		Note:         "Returns transaction batches for a specific employee (user_id) in the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}

		paginated, err := core.TransactionBatchManager(service).NormalPagination(context, ctx, &types.TransactionBatch{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			EmployeeUserID: &userOrganization.UserID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to paginate transaction batches: " + err.Error()})
		}

		for i, batch := range paginated.Data {
			if !batch.CanView {
				minimalBatch, err := core.TransactionBatchMinimal(context, service, batch.ID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve minimal transaction batch: " + err.Error()})
				}
				paginated.Data[i] = minimalBatch
			}
		}
		return ctx.JSON(http.StatusOK, paginated)
	})

	// /transaction-batch/:transaction-batch/history/total
	// GET
}
