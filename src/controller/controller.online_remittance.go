package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) OnlineRemittanceController() {
	req := c.provider.Service.Request

	// Retrieve batch online remittance (JWT) for the current transaction batch before ending.
	req.RegisterRoute(horizon.Route{
		Route:    "/online-remittance",
		Method:   "GET",
		Response: "IOnlineRemittance[]",
		Note:     "Returns online remittance records for the current active transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found"})
		}

		onlineRemittance, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve online remittance: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModels(onlineRemittance))
	})

	// Create a new online remittance for the current transaction batch before ending.
	req.RegisterRoute(horizon.Route{
		Route:    "/online-remittance",
		Method:   "POST",
		Response: "IOnlineRemittance",
		Request:  "IOnlineRemittance",
		Note:     "Creates a new online remittance record for the current transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := c.model.OnlineRemittanceManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found"})
		}

		onlineRemittance := &model.OnlineRemittance{
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
			CountryCode:        req.CountryCode,
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

		if err := c.model.OnlineRemittanceManager.Create(context, onlineRemittance); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create online remittance: " + err.Error()})
		}

		allOnlineRemittances, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve online remittances: " + err.Error()})
		}

		var totalOnlineRemittance float64
		for _, remittance := range allOnlineRemittances {
			totalOnlineRemittance += remittance.Amount
		}

		transactionBatch.TotalOnlineRemittance = totalOnlineRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModel(onlineRemittance))
	})

	// Update an existing online remittance by ID for the current transaction batch.
	req.RegisterRoute(horizon.Route{
		Route:    "/online-remittance/:online_remittance_id",
		Method:   "PUT",
		Response: "IOnlineRemittance",
		Request:  "IOnlineRemittance",
		Note:     "Updates an existing online remittance by its ID for the current active transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		onlineRemittanceId, err := horizon.EngineUUIDParam(ctx, "online_remittance_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid online_remittance_id: " + err.Error()})
		}

		req, err := c.model.OnlineRemittanceManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		existingOnlineRemittance, err := c.model.OnlineRemittanceManager.GetByID(context, *onlineRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Online remittance not found: " + err.Error()})
		}

		if existingOnlineRemittance.OrganizationID != userOrg.OrganizationID ||
			existingOnlineRemittance.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Online remittance not found in your organization/branch"})
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
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
		existingOnlineRemittance.CountryCode = req.CountryCode
		existingOnlineRemittance.ReferenceNumber = req.ReferenceNumber
		existingOnlineRemittance.AccountName = req.AccountName
		existingOnlineRemittance.Amount = req.Amount
		existingOnlineRemittance.DateEntry = req.DateEntry
		existingOnlineRemittance.Description = req.Description

		if existingOnlineRemittance.DateEntry == nil {
			now := time.Now().UTC()
			existingOnlineRemittance.DateEntry = &now
		}

		if err := c.model.OnlineRemittanceManager.UpdateFields(context, *onlineRemittanceId, existingOnlineRemittance); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update online remittance: " + err.Error()})
		}

		allOnlineRemittances, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve online remittances: " + err.Error()})
		}

		var totalOnlineRemittance float64
		for _, remittance := range allOnlineRemittances {
			totalOnlineRemittance += remittance.Amount
		}

		transactionBatch.TotalOnlineRemittance = totalOnlineRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		updatedRemittance, err := c.model.OnlineRemittanceManager.GetByID(context, *onlineRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated online remittance: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModel(updatedRemittance))
	})

	// Delete an existing online remittance by ID for the current transaction batch.
	req.RegisterRoute(horizon.Route{
		Route:    "/online-remittance/:online_remittance_id",
		Method:   "DELETE",
		Response: "IOnlineRemittance",
		Note:     "Deletes an online remittance by its ID for the current active transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		onlineRemittanceId, err := horizon.EngineUUIDParam(ctx, "online_remittance_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid online_remittance_id: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		existingOnlineRemittance, err := c.model.OnlineRemittanceManager.GetByID(context, *onlineRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Online remittance not found: " + err.Error()})
		}

		if existingOnlineRemittance.OrganizationID != userOrg.OrganizationID ||
			existingOnlineRemittance.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Online remittance not found in your organization/branch"})
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found"})
		}

		if existingOnlineRemittance.TransactionBatchID == nil ||
			*existingOnlineRemittance.TransactionBatchID != transactionBatch.ID {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Online remittance does not belong to current transaction batch"})
		}

		if err := c.model.OnlineRemittanceManager.DeleteByID(context, *onlineRemittanceId); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete online remittance: " + err.Error()})
		}

		allOnlineRemittances, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve online remittances: " + err.Error()})
		}

		var totalOnlineRemittance float64
		for _, remittance := range allOnlineRemittances {
			totalOnlineRemittance += remittance.Amount
		}

		transactionBatch.TotalOnlineRemittance = totalOnlineRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModel(existingOnlineRemittance))
	})
}
