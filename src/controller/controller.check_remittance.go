package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) CheckRemittanceController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/check-remittance",
		Method:   "GET",
		Response: "ICheckRemittance[]",
		Note:     "Retrieve batch check remittance (JWT) for the current transaction batch before ending.",
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

		// Retrieve check remittance for the current transaction batch
		checkRemittance, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.ToModels(checkRemittance))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/check-remittance",
		Method:   "POST",
		Response: "ICheckRemittance",
		Request:  "ICheckRemittance",
		Note:     "Create a new check remittance for the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Validate the check remittance request
		req, err := c.model.CheckRemittanceManager.Validate(ctx)
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

		// Set required fields for the check remittance
		checkRemittance := &model.CheckRemittance{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,

			BankID:             req.BankID,
			MediaID:            req.MediaID,
			EmployeeUserID:     &userOrg.UserID,
			TransactionBatchID: &transactionBatch.ID,
			CountryCode:        req.CountryCode,
			ReferenceNumber:    req.ReferenceNumber,
			AccountName:        req.AccountName,
			Amount:             req.Amount,
			DateEntry:          req.DateEntry,
			Description:        req.Description,
		}

		// Set default date entry if not provided
		if checkRemittance.DateEntry == nil {
			now := time.Now().UTC()
			checkRemittance.DateEntry = &now
		}

		// Create the check remittance
		if err := c.model.CheckRemittanceManager.Create(context, checkRemittance); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create check remittance: " + err.Error()})
		}

		// Get all check remittances for recalculating transaction batch totals
		allCheckRemittances, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total check remittance amount
		var totalCheckRemittance float64
		for _, remittance := range allCheckRemittances {
			totalCheckRemittance += remittance.Amount
		}

		// Update transaction batch totals
		transactionBatch.TotalCheckRemittance = totalCheckRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		// Save the updated transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the created check remittance
		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.ToModel(checkRemittance))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/check-remittance/:check_remittance_id",
		Method:   "PUT",
		Response: "ICheckRemittance",
		Request:  "ICheckRemittance",
		Note:     "Update an existing check remittance by ID for the current transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get check remittance ID from URL parameter
		checkRemittanceId, err := horizon.EngineUUIDParam(ctx, "check_remittance_id")
		if err != nil {
			return err
		}

		// Validate the check remittance request
		req, err := c.model.CheckRemittanceManager.Validate(ctx)
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

		// Get the existing check remittance
		existingCheckRemittance, err := c.model.CheckRemittanceManager.GetByID(context, *checkRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Check remittance not found"})
		}

		// Verify ownership
		if existingCheckRemittance.OrganizationID != userOrg.OrganizationID ||
			existingCheckRemittance.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "Check remittance not found in your organization/branch")
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

		// Update the check remittance fields
		updatedCheckRemittance := &model.CheckRemittance{
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CreatedByID:    userOrg.UserID,

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
		if updatedCheckRemittance.DateEntry == nil {
			now := time.Now().UTC()
			updatedCheckRemittance.DateEntry = &now
		}

		// Update the check remittance
		if err := c.model.CheckRemittanceManager.UpdateFields(context, *checkRemittanceId, updatedCheckRemittance); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update check remittance: " + err.Error()})
		}

		// Get all check remittances for recalculating transaction batch totals
		allCheckRemittances, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total check remittance amount
		var totalCheckRemittance float64
		for _, remittance := range allCheckRemittances {
			totalCheckRemittance += remittance.Amount
		}

		// Update transaction batch totals
		transactionBatch.TotalCheckRemittance = totalCheckRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		// Save the updated transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the updated check remittance
		updatedRemittance, err := c.model.CheckRemittanceManager.GetByID(context, *checkRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated check remittance: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.ToModel(updatedRemittance))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/check-remittance/:check_remittance_id",
		Method:   "DELETE",
		Response: "ICheckRemittance",
		Note:     "Delete an existing check remittance by ID for the current transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get check remittance ID from URL parameter
		checkRemittanceId, err := horizon.EngineUUIDParam(ctx, "check_remittance_id")
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

		// Get the existing check remittance
		existingCheckRemittance, err := c.model.CheckRemittanceManager.GetByID(context, *checkRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Check remittance not found"})
		}

		// Verify ownership
		if existingCheckRemittance.OrganizationID != userOrg.OrganizationID ||
			existingCheckRemittance.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "Check remittance not found in your organization/branch")
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
		if existingCheckRemittance.TransactionBatchID == nil ||
			*existingCheckRemittance.TransactionBatchID != transactionBatch.ID {
			return c.BadRequest(ctx, "Check remittance does not belong to current transaction batch")
		}

		// Delete the check remittance
		if err := c.model.CheckRemittanceManager.DeleteByID(context, *checkRemittanceId); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete check remittance: " + err.Error()})
		}

		// Get all remaining check remittances for recalculating transaction batch totals
		allCheckRemittances, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate total check remittance amount
		var totalCheckRemittance float64
		for _, remittance := range allCheckRemittances {
			totalCheckRemittance += remittance.Amount
		}

		// Update transaction batch totals
		transactionBatch.TotalCheckRemittance = totalCheckRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		// Save the updated transaction batch
		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		// Return the deleted check remittance
		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.ToModel(existingCheckRemittance))
	})
}
