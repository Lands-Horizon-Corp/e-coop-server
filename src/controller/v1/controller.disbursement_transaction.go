package controller_v1

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) DisbursementTransactionController() {
	req := c.provider.Service.Request

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction",
		Method:       "POST",
		Note:         "Returns all disbursement transactions for a specific transaction batch.",
		ResponseType: model.DisbursementTransactionResponse{},
		RequestType:  model.DisbursementTransactionRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.DisbursementTransactionManager.Validate(ctx)
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
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		transactionBatch, err := c.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil || transactionBatch == nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		data := &model.DisbursementTransaction{
			CreatedAt:                  time.Now().UTC(),
			CreatedByID:                userOrg.UserID,
			UpdatedAt:                  time.Now().UTC(),
			UpdatedByID:                userOrg.UserID,
			OrganizationID:             userOrg.OrganizationID,
			BranchID:                   *userOrg.BranchID,
			TransactionBatchID:         transactionBatch.ID,
			DisbursementID:             *req.DisbursementID,
			EmployeeName:               userOrg.User.FullName,
			Description:                req.Description,
			TransactionReferenceNumber: req.TransactionReferenceNumber,
			ReferenceNumber:            req.ReferenceNumber,
			Amount:                     req.Amount,
		}
		if err := c.model.DisbursementTransactionManager.Create(context, data); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Disbursement transaction creation failed (/disbursement-transaction), db error: " + err.Error(),
				Module:      "DisbursementTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create disbursement transaction: " + err.Error()})
		}
		if req.IsReferenceNumberChecked {
			userOrg.UserSettingUsedOR += 1
			if err := c.model.UserOrganizationManager.UpdateByID(context, data.ID, userOrg); err != nil {
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
		return ctx.JSON(http.StatusCreated, c.model.DisbursementTransactionManager.ToModel(data))

	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions for a specific transaction batch.",
		ResponseType: model.DisbursementResponse{},
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
		disbursementTransactions, err := c.model.DisbursementTransactionManager.Find(context, &model.DisbursementTransaction{
			TransactionBatchID: *transactionBatchID,
			BranchID:           *user.BranchID,
			OrganizationID:     user.OrganizationID,
		})
		return ctx.JSON(http.StatusOK, c.model.DisbursementTransactionManager.Pagination(context, ctx, disbursementTransactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction/employee/:user_organization_id/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions handled by a specific employee.",
		ResponseType: model.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		useOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if useOrganization.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		disbursementTransactions, err := c.model.DisbursementTransactionManager.Find(context, &model.DisbursementTransaction{
			CreatedByID:    useOrganization.UserID,
			BranchID:       *useOrganization.BranchID,
			OrganizationID: useOrganization.OrganizationID,
		})
		return ctx.JSON(http.StatusOK, c.model.DisbursementTransactionManager.Pagination(context, ctx, disbursementTransactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction/current/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions for the currently authenticated user.",
		ResponseType: model.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		disbursementTransactions, err := c.model.DisbursementTransactionManager.Find(context, &model.DisbursementTransaction{
			CreatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})
		return ctx.JSON(http.StatusOK, c.model.DisbursementTransactionManager.Pagination(context, ctx, disbursementTransactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction/branch/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions for the current user's branch.",
		ResponseType: model.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		disbursementTransactions, err := c.model.DisbursementTransactionManager.Find(context, &model.DisbursementTransaction{
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})
		return ctx.JSON(http.StatusOK, c.model.DisbursementTransactionManager.Pagination(context, ctx, disbursementTransactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction/disbursement/:disbursement_id/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions for a specific disbursement ID.",
		ResponseType: model.DisbursementResponse{},
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
		disbursementTransactions, err := c.model.DisbursementTransactionManager.Find(context, &model.DisbursementTransaction{
			DisbursementID: *disbursementID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve disbursement transactions"})
		}
		return ctx.JSON(http.StatusOK, c.model.DisbursementTransactionManager.Pagination(context, ctx, disbursementTransactions))
	})
}
