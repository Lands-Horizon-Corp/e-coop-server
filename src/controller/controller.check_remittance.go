package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

// CheckRemittanceController manages endpoints for check remittance operations within the current transaction batch.
func (c *Controller) CheckRemittanceController() {
	req := c.provider.Service.Request

	// GET /check-remittance: List all check remittances for the active transaction batch.
	req.RegisterRoute(horizon.Route{
		Route:    "/check-remittance",
		Method:   "GET",
		Response: "ICheckRemittance[]",
		Note:     "Returns all check remittances for the current active transaction batch of the authenticated user's branch. Only 'owner' or 'employee' roles are allowed.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view check remittances"})
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
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		checkRemittance, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve check remittances: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.ToModels(checkRemittance))
	})

	// POST /check-remittance: Create a new check remittance for the current transaction batch.
	req.RegisterRoute(horizon.Route{
		Route:    "/check-remittance",
		Method:   "POST",
		Response: "ICheckRemittance",
		Request:  "ICheckRemittance",
		Note:     "Creates a new check remittance for the current active transaction batch. Only 'owner' or 'employee' roles are allowed.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.CheckRemittanceManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check remittance data: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create check remittances"})
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
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

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

		if checkRemittance.DateEntry == nil {
			now := time.Now().UTC()
			checkRemittance.DateEntry = &now
		}

		if err := c.model.CheckRemittanceManager.Create(context, checkRemittance); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create check remittance: " + err.Error()})
		}

		allCheckRemittances, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to recalculate check remittances: " + err.Error()})
		}

		// Recalculate totals
		var totalCheckRemittance float64
		for _, remittance := range allCheckRemittances {
			totalCheckRemittance += remittance.Amount
		}
		transactionBatch.TotalCheckRemittance = totalCheckRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch totals: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, c.model.CheckRemittanceManager.ToModel(checkRemittance))
	})

	// PUT /check-remittance/:check_remittance_id: Update a check remittance by ID for the current transaction batch.
	req.RegisterRoute(horizon.Route{
		Route:    "/check-remittance/:check_remittance_id",
		Method:   "PUT",
		Response: "ICheckRemittance",
		Request:  "ICheckRemittance",
		Note:     "Updates an existing check remittance by ID for the current transaction batch. Only 'owner' or 'employee' roles are allowed.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		checkRemittanceId, err := horizon.EngineUUIDParam(ctx, "check_remittance_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check remittance ID"})
		}

		req, err := c.model.CheckRemittanceManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check remittance data: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update check remittances"})
		}

		existingCheckRemittance, err := c.model.CheckRemittanceManager.GetByID(context, *checkRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Check remittance not found"})
		}

		if existingCheckRemittance.OrganizationID != userOrg.OrganizationID || existingCheckRemittance.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Check remittance does not belong to your organization/branch"})
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
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		updatedCheckRemittance := &model.CheckRemittance{
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CreatedByID:    existingCheckRemittance.CreatedByID,
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

		if updatedCheckRemittance.DateEntry == nil {
			now := time.Now().UTC()
			updatedCheckRemittance.DateEntry = &now
		}

		if err := c.model.CheckRemittanceManager.UpdateFields(context, *checkRemittanceId, updatedCheckRemittance); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update check remittance: " + err.Error()})
		}

		allCheckRemittances, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to recalculate check remittances: " + err.Error()})
		}

		var totalCheckRemittance float64
		for _, remittance := range allCheckRemittances {
			totalCheckRemittance += remittance.Amount
		}
		transactionBatch.TotalCheckRemittance = totalCheckRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch totals: " + err.Error()})
		}

		updatedRemittance, err := c.model.CheckRemittanceManager.GetByID(context, *checkRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated check remittance: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.ToModel(updatedRemittance))
	})

	// DELETE /check-remittance/:check_remittance_id: Delete a check remittance by ID for the current transaction batch.
	req.RegisterRoute(horizon.Route{
		Route:    "/check-remittance/:check_remittance_id",
		Method:   "DELETE",
		Response: "ICheckRemittance",
		Note:     "Deletes a check remittance by ID for the current transaction batch. Only 'owner' or 'employee' roles are allowed.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		checkRemittanceId, err := horizon.EngineUUIDParam(ctx, "check_remittance_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check remittance ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete check remittance"})
		}

		existingCheckRemittance, err := c.model.CheckRemittanceManager.GetByID(context, *checkRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Check remittance not found"})
		}

		if existingCheckRemittance.OrganizationID != userOrg.OrganizationID || existingCheckRemittance.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Check remittance does not belong to your organization/branch"})
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
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		if existingCheckRemittance.TransactionBatchID == nil || *existingCheckRemittance.TransactionBatchID != transactionBatch.ID {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Check remittance does not belong to current transaction batch"})
		}

		if err := c.model.CheckRemittanceManager.DeleteByID(context, *checkRemittanceId); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete check remittance: " + err.Error()})
		}

		allCheckRemittances, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to recalculate check remittances: " + err.Error()})
		}

		var totalCheckRemittance float64
		for _, remittance := range allCheckRemittances {
			totalCheckRemittance += remittance.Amount
		}
		transactionBatch.TotalCheckRemittance = totalCheckRemittance
		transactionBatch.TotalActualRemittance = transactionBatch.TotalCheckRemittance + transactionBatch.TotalOnlineRemittance + transactionBatch.TotalDepositInBank
		transactionBatch.UpdatedAt = time.Now().UTC()
		transactionBatch.UpdatedByID = userOrg.UserID

		if err := c.model.TransactionBatchManager.UpdateFields(context, transactionBatch.ID, transactionBatch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch totals: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.ToModel(existingCheckRemittance))
	})
}