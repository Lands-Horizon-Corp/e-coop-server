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

	req.RegisterRoute(horizon.Route{
		Route:    "/online-remittance",
		Method:   "GET",
		Response: "IOnlineRemittance[]",
		Note:     "Retrieve batch online remittance (JWT) for the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Find the current active transaction batch
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

		// Retrieve online remittance for the current transaction batch
		onlineRemittance, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModels(onlineRemittance))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/online-remittance",
		Method:   "POST",
		Response: "IOnlineRemittance",
		Request:  "IOnlineRemittance",
		Note:     "Create a new online remittance for the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Validate the online remittance request
		req, err := c.model.OnlineRemittanceManager.Validate(ctx)
		if err != nil {
			return err
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Find the current active transaction batch
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

		// Set required fields for the online remittance
		onlineRemittance := &model.OnlineRemittance{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TransactionBatchID: &transactionBatch.ID,

			BankID:          req.BankID,
			MediaID:         req.MediaID,
			EmployeeUserID:  &userOrg.UserID,
			CountryCode:     req.CountryCode,
			ReferenceNumber: req.ReferenceNumber,
			AccountName:     req.AccountName,
			Amount:          req.Amount,
			DateEntry:       req.DateEntry,
			Description:     req.Description,
		}

		// Set default date entry if not provided
		if onlineRemittance.DateEntry == nil {
			now := time.Now().UTC()
			onlineRemittance.DateEntry = &now
		}

		// Create the online remittance
		if err := c.model.OnlineRemittanceManager.Create(context, onlineRemittance); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create online remittance: " + err.Error()})
		}

		// Get all online remittances for recalculating transaction batch totals
		allOnlineRemittances, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total online remittance amount
		var totalOnlineRemittance float64
		for _, remittance := range allOnlineRemittances {
			totalOnlineRemittance += remittance.Amount
		}

		// Update transaction batch totals
		transactionBatch.TotalOnlineRemittance = totalOnlineRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		// Save the updated transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the created online remittance
		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModel(onlineRemittance))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/online-remittance/:online_remittance_id",
		Method:   "PUT",
		Response: "IOnlineRemittance",
		Request:  "IOnlineRemittance",
		Note:     "Update an existing online remittance by ID for the current transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get online remittance ID from URL parameter
		onlineRemittanceId, err := horizon.EngineUUIDParam(ctx, "online_remittance_id")
		if err != nil {
			return err
		}

		// Validate the online remittance request
		req, err := c.model.OnlineRemittanceManager.Validate(ctx)
		if err != nil {
			return err
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Get the existing online remittance
		existingOnlineRemittance, err := c.model.OnlineRemittanceManager.GetByID(context, *onlineRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Online remittance not found"})
		}

		// Verify ownership
		if existingOnlineRemittance.OrganizationID != userOrg.OrganizationID ||
			existingOnlineRemittance.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "Online remittance not found in your organization/branch")
		}

		// Find the current active transaction batch
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

		// Update the online remittance fields
		updatedOnlineRemittance := &model.OnlineRemittance{
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			CreatedByID:        userOrg.UserID,
			TransactionBatchID: &transactionBatch.ID,
			BankID:             req.BankID,
			MediaID:            req.MediaID,
			CountryCode:        req.CountryCode,
			ReferenceNumber:    req.ReferenceNumber,
			AccountName:        req.AccountName,
			Amount:             req.Amount,
			DateEntry:          req.DateEntry,
			Description:        req.Description,
		}

		// Set default date entry if not provided
		if updatedOnlineRemittance.DateEntry == nil {
			now := time.Now().UTC()
			updatedOnlineRemittance.DateEntry = &now
		}

		// Update the online remittance
		if err := c.model.OnlineRemittanceManager.UpdateFields(context, *onlineRemittanceId, updatedOnlineRemittance); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update online remittance: " + err.Error()})
		}

		// Get all online remittances for recalculating transaction batch totals
		allOnlineRemittances, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total online remittance amount
		var totalOnlineRemittance float64
		for _, remittance := range allOnlineRemittances {
			totalOnlineRemittance += remittance.Amount
		}

		// Update transaction batch totals
		transactionBatch.TotalOnlineRemittance = totalOnlineRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		// Save the updated transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the updated online remittance
		updatedRemittance, err := c.model.OnlineRemittanceManager.GetByID(context, *onlineRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated online remittance: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModel(updatedRemittance))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/online-remittance/:online_remittance_id",
		Method:   "DELETE",
		Response: "IOnlineRemittance",
		Note:     "Delete an existing online remittance by ID for the current transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get online remittance ID from URL parameter
		onlineRemittanceId, err := horizon.EngineUUIDParam(ctx, "online_remittance_id")
		if err != nil {
			return err
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Get the existing online remittance
		existingOnlineRemittance, err := c.model.OnlineRemittanceManager.GetByID(context, *onlineRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Online remittance not found"})
		}

		// Verify ownership
		if existingOnlineRemittance.OrganizationID != userOrg.OrganizationID ||
			existingOnlineRemittance.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "Online remittance not found in your organization/branch")
		}

		// Find the current active transaction batch
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

		// Verify the remittance belongs to the current transaction batch
		if existingOnlineRemittance.TransactionBatchID == nil ||
			*existingOnlineRemittance.TransactionBatchID != transactionBatch.ID {
			return c.BadRequest(ctx, "Online remittance does not belong to current transaction batch")
		}

		// Delete the online remittance
		if err := c.model.OnlineRemittanceManager.DeleteByID(context, *onlineRemittanceId); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete online remittance: " + err.Error()})
		}

		// Get all remaining online remittances for recalculating transaction batch totals
		allOnlineRemittances, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total online remittance amount
		var totalOnlineRemittance float64
		for _, remittance := range allOnlineRemittances {
			totalOnlineRemittance += remittance.Amount
		}

		// Update transaction batch totals
		transactionBatch.TotalOnlineRemittance = totalOnlineRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		// Save the updated transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the deleted online remittance
		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModel(existingOnlineRemittance))
	})
}
