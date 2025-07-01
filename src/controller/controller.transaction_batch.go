package controller

import (
	"fmt"
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
		transactionBatch, _ := c.model.TransactionBatchManager.FindOne(context, &model.TransactionBatch{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			IsClosed:       false,
		})
		if transactionBatch == nil {
			return ctx.NoContent(http.StatusNoContent)
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
		Route:    "/transaction-batch",
		Method:   "POST",
		Response: "ITransactionBatch",
		Request:  "ITransactionBatch",
		Note:     "Create and start a new transaction batch; returns the created batch. (Will populate Cashcount)",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.BatchFundingManager.Validate(ctx)
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
		transactionBatch, _ := c.model.TransactionBatchManager.FindOne(context, &model.TransactionBatch{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			IsClosed:       false,
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
			TotalSupposedRemitance:        0,
			TotalCashOnHand:               0,
			TotalCheckRemittance:          0,
			TotalOnlineRemittance:         0,
			TotalDepositInBank:            0,
			TotalActualRemittance:         0,
			TotalActualSupposedComparison: 0,

			IsClosed:       false,
			CanView:        false,
			RequestView:    nil,
			EndedAt:        nil,
			TotalBatchTime: nil,
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
			Name:             req.Name,
			Description:      req.Description,
			Amount:           req.Amount,
			SignatureMediaID: req.SignatureMediaID,
		}
		fmt.Println(transBatch)
		fmt.Println(batchFunding)
		fmt.Println("-----")
		if err := c.model.BatchFundingManager.CreateWithTx(context, tx, batchFunding); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
		}
		result, err := c.model.TransactionBatchMinimal(context, transBatch.ID)
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, result)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-batch/end",
		Method:   "POST",
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
		transactionBatch, _ := c.model.TransactionBatchManager.FindOne(context, &model.TransactionBatch{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			IsClosed:       false,
		})
		if transactionBatch == nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "No current transaction batch"})
		}
		now := time.Now().UTC()
		totalTime := transactionBatch.CreatedAt.Sub(now)
		transactionBatch.IsClosed = true
		transactionBatch.EmployeeBySignatureMediaID = req.EmployeeBySignatureMediaID
		transactionBatch.EmployeeByName = req.EmployeeByName
		transactionBatch.EmployeeByPosition = req.EmployeeByPosition
		transactionBatch.EndedAt = &now
		transactionBatch.TotalBatchTime = &totalTime
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
		transactionBatch, err := c.model.TransactionBatchManager.Find(context, &model.TransactionBatch{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CanView:        false,
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
		return nil
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
		return nil
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

func (c *Controller) BatchFunding() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/batch-funding",
		Method:   "POST",
		Request:  "IBatchFunding",
		Response: "IBatchFunding",
		Note:     "Sart: create batch funding based on current transaction batch",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/batch-funding/transaction-batch/:transaction_batch_id",
		Method:   "GET",
		Response: "IBatchFunding[]",
		Note:     "Get all batch funding of transaction batch.",
	}, func(ctx echo.Context) error {
		return nil
	})
}

func (C *Controller) CheckRemittance() {

}
