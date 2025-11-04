package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) disbursementTransactionController() {
	req := c.provider.Service.Request

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction",
		Method:       "POST",
		Note:         "Returns all disbursement transactions for a specific/current transaction batch.",
		ResponseType: core.DisbursementTransactionResponse{},
		RequestType:  core.DisbursementTransactionRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.DisbursementTransactionManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Disbursement transaction creation failed (/disbursement-transaction), validation error: " + err.Error(),
				Module:      "DisbursementTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid disbursement transaction data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)

		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		transactionBatch, err := c.core.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil || transactionBatch == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for the user"})
		}
		data := &core.DisbursementTransaction{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			EmployeeUserID:     userOrg.UserID,
			TransactionBatchID: transactionBatch.ID,
			DisbursementID:     *req.DisbursementID,
			EmployeeName:       userOrg.User.FullName,
			Description:        req.Description,
			ReferenceNumber:    req.ReferenceNumber,
			Amount:             req.Amount,
		}
		if err := c.core.DisbursementTransactionManager.Create(context, data); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Disbursement transaction creation failed (/disbursement-transaction), db error: " + err.Error(),
				Module:      "DisbursementTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create disbursement transaction: " + err.Error()})
		}
		if req.IsReferenceNumberChecked {
			userOrg.UserSettingUsedOR++
			if err := c.core.UserOrganizationManager.UpdateByID(context, userOrg.ID, userOrg); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Disbursement transaction reference number update failed (/disbursement-transaction), db error: " + err.Error(),
					Module:      "DisbursementTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update reference number: " + err.Error()})
			}
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created disbursement transaction (/disbursement-transaction): " + data.ID.String(),
			Module:      "DisbursementTransaction",
		})
		return ctx.JSON(http.StatusCreated, c.core.DisbursementTransactionManager.ToModel(data))

	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions for a specific transaction batch.",
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		disbursementTransactions, err := c.core.DisbursementTransactionManager.PaginationWithFields(context, ctx, &core.DisbursementTransaction{
			TransactionBatchID: *transactionBatchID,
			BranchID:           *user.BranchID,
			OrganizationID:     user.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve disbursement transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, disbursementTransactions)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction/employee/:user_organization_id/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions handled by a specific employee.",
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		useOrganization, err := c.core.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if useOrganization.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		disbursementTransactions, err := c.core.DisbursementTransactionManager.PaginationWithFields(context, ctx, &core.DisbursementTransaction{
			CreatedByID:    useOrganization.UserID,
			BranchID:       *useOrganization.BranchID,
			OrganizationID: useOrganization.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve disbursement transactions: " + err.Error()})
		}
		// Return paginated response
		return ctx.JSON(http.StatusOK, disbursementTransactions)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction/current/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions for the currently authenticated user.",
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		disbursementTransactions, err := c.core.DisbursementTransactionManager.PaginationWithFields(context, ctx, &core.DisbursementTransaction{
			CreatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve disbursement transactions: " + err.Error()})
		}
		// Return paginated response
		return ctx.JSON(http.StatusOK, disbursementTransactions)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction/current",
		Method:       "GET",
		Note:         "Returns all disbursement transactions for the currently authenticated user.",
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		disbursementTransactions, err := c.core.DisbursementTransactionManager.FindRaw(context, &core.DisbursementTransaction{
			CreatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve disbursement transactions: " + err.Error()})
		}
		// Return paginated response
		return ctx.JSON(http.StatusOK, disbursementTransactions)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction/branch/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions for the current user's branch.",
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		disbursementTransactions, err := c.core.DisbursementTransactionManager.PaginationWithFields(context, ctx, &core.DisbursementTransaction{
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve disbursement transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, disbursementTransactions)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction/disbursement/:disbursement_id/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions for a specific disbursement ID.",
		ResponseType: core.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		disbursementID, err := handlers.EngineUUIDParam(ctx, "disbursement_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid disbursement ID"})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		disbursementTransactions, err := c.core.DisbursementTransactionManager.PaginationWithFields(context, ctx, &core.DisbursementTransaction{
			DisbursementID: *disbursementID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve disbursement transactions"})
		}
		return ctx.JSON(http.StatusOK, disbursementTransactions)
	})
}
