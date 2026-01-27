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
)

func OnlineRemittanceController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/online-remittance",
		Method:       "GET",
		ResponseType: types.OnlineRemittanceResponse{},
		Note:         "Returns online remittance records for the current active transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		transactionBatch, err := core.TransactionBatchCurrent(
			context, service,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found"})
		}

		onlineRemittance, err := core.OnlineRemittanceManager(service).Find(context, &types.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve online remittance: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, core.OnlineRemittanceManager(service).ToModels(onlineRemittance))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/online-remittance",
		Method:       "POST",
		ResponseType: types.OnlineRemittanceResponse{},
		RequestType:  types.OnlineRemittanceRequest{},
		Note:         "Creates a new online remittance record for the current transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := core.OnlineRemittanceManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create online remittance failed: validation error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create online remittance failed: user org error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create online remittance failed: unauthorized user type",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
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
				Description: "Create online remittance failed: find batch error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create online remittance failed: no active batch",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found"})
		}

		onlineRemittance := &types.OnlineRemittance{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TransactionBatchID: &transactionBatch.ID,
			BankID:             req.BankID,
			MediaID:            req.MediaID,
			EmployeeUserID:     &userOrg.UserID,
			CurrencyID:         req.CurrencyID,
			ReferenceNumber:    req.ReferenceNumber,
			AccountName:        req.AccountName,
			Amount:             req.Amount,
			DateEntry:          req.DateEntry,
			Description:        req.Description,
		}

		if onlineRemittance.DateEntry == nil {
			now := time.Now().UTC()
			onlineRemittance.DateEntry = &now
		}

		if err := core.OnlineRemittanceManager(service).Create(context, onlineRemittance); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create online remittance failed: create error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create online remittance: " + err.Error()})
		}

		if err := event.TransactionBatchBalancing(context, service, &transactionBatch.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch after saving: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created online remittance for batch ID: " + transactionBatch.ID.String(),
			Module:      "OnlineRemittance",
		})

		return ctx.JSON(http.StatusOK, core.OnlineRemittanceManager(service).ToModel(onlineRemittance))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/online-remittance/:online_remittance_id",
		Method:       "PUT",
		ResponseType: types.OnlineRemittanceResponse{},
		RequestType:  types.OnlineRemittanceRequest{},
		Note:         "Updates an existing online remittance by its ID for the current active transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		onlineRemittanceID, err := helpers.EngineUUIDParam(ctx, "online_remittance_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: invalid online_remittance_id: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid online_remittance_id: " + err.Error()})
		}

		req, err := core.OnlineRemittanceManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: validation error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: user org error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: unauthorized user type",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		existingOnlineRemittance, err := core.OnlineRemittanceManager(service).GetByID(context, *onlineRemittanceID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: not found: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Online remittance not found: " + err.Error()})
		}

		if existingOnlineRemittance.OrganizationID != userOrg.OrganizationID ||
			existingOnlineRemittance.BranchID != *userOrg.BranchID {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: not in org/branch",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Online remittance not found in your organization/branch"})
		}

		transactionBatch, err := core.TransactionBatchCurrent(
			context, service,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: find batch error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: no active batch",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found"})
		}

		existingOnlineRemittance.UpdatedAt = time.Now().UTC()
		existingOnlineRemittance.UpdatedByID = userOrg.UserID
		existingOnlineRemittance.OrganizationID = userOrg.OrganizationID
		existingOnlineRemittance.BranchID = *userOrg.BranchID
		existingOnlineRemittance.CreatedByID = userOrg.UserID
		existingOnlineRemittance.TransactionBatchID = &transactionBatch.ID
		existingOnlineRemittance.BankID = req.BankID
		existingOnlineRemittance.MediaID = req.MediaID
		existingOnlineRemittance.CurrencyID = req.CurrencyID
		existingOnlineRemittance.ReferenceNumber = req.ReferenceNumber
		existingOnlineRemittance.AccountName = req.AccountName
		existingOnlineRemittance.Amount = req.Amount
		existingOnlineRemittance.DateEntry = req.DateEntry
		existingOnlineRemittance.Description = req.Description

		if existingOnlineRemittance.DateEntry == nil {
			now := time.Now().UTC()
			existingOnlineRemittance.DateEntry = &now
		}

		if err := core.OnlineRemittanceManager(service).UpdateByID(context, *onlineRemittanceID, existingOnlineRemittance); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: update error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update online remittance: " + err.Error()})
		}

		updatedRemittance, err := core.OnlineRemittanceManager(service).GetByID(context, *onlineRemittanceID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: get updated error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated online remittance: " + err.Error()})
		}

		if err := event.TransactionBatchBalancing(context, service, &transactionBatch.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch after saving: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated online remittance for batch ID: " + transactionBatch.ID.String(),
			Module:      "OnlineRemittance",
		})

		return ctx.JSON(http.StatusOK, core.OnlineRemittanceManager(service).ToModel(updatedRemittance))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/online-remittance/:online_remittance_id",
		Method: "DELETE",
		Note:   "Deletes an online remittance by its ID for the current active transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		onlineRemittanceID, err := helpers.EngineUUIDParam(ctx, "online_remittance_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: invalid online_remittance_id: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid online_remittance_id: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: user org error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: unauthorized user type",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		existingOnlineRemittance, err := core.OnlineRemittanceManager(service).GetByID(context, *onlineRemittanceID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: not found: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Online remittance not found: " + err.Error()})
		}

		if existingOnlineRemittance.OrganizationID != userOrg.OrganizationID ||
			existingOnlineRemittance.BranchID != *userOrg.BranchID {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: not in org/branch",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Online remittance not found in your organization/branch"})
		}

		transactionBatch, err := core.TransactionBatchCurrent(
			context, service,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: find batch error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: no active batch",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found"})
		}

		if existingOnlineRemittance.TransactionBatchID == nil ||
			*existingOnlineRemittance.TransactionBatchID != transactionBatch.ID {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: not in current batch",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Online remittance does not belong to current transaction batch"})
		}

		if err := core.OnlineRemittanceManager(service).Delete(context, *onlineRemittanceID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: delete error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete online remittance: " + err.Error()})
		}

		if err := event.TransactionBatchBalancing(context, service, &transactionBatch.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch after saving: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted online remittance for batch ID: " + transactionBatch.ID.String(),
			Module:      "OnlineRemittance",
		})

		return ctx.JSON(http.StatusOK, core.OnlineRemittanceManager(service).ToModel(existingOnlineRemittance))
	})
}
