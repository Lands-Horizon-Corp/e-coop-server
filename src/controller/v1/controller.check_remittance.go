package controller_v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/labstack/echo/v4"
)

// CheckRemittanceController manages endpoints for check remittance operations within the current transaction batch.
func (c *Controller) CheckRemittanceController() {
	req := c.provider.Service.Request

	// GET /check-remittance: List all check remittances for the active transaction batch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/check-remittance",
		Method:       "GET",
		Note:         "Returns all check remittances for the current active transaction batch of the authenticated user's branch. Only 'owner' or 'employee' roles are allowed.",
		ResponseType: model.CheckRemittanceResponse{},
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

		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.Filtered(context, ctx, checkRemittance))
	})

	// POST /check-remittance: Create a new check remittance for the current transaction batch. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/check-remittance",
		Method:       "POST",
		ResponseType: model.CheckRemittanceResponse{},
		RequestType:  model.CheckRemittanceRequest{},
		Note:         "Creates a new check remittance for the current active transaction batch. Only 'owner' or 'employee' roles are allowed.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.CheckRemittanceManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Check remittance creation failed (/check-remittance), validation error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check remittance data: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Check remittance creation failed (/check-remittance), user org error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for check remittance (/check-remittance)",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create check remittances"})
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Check remittance creation failed (/check-remittance), transaction batch lookup error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Check remittance creation failed (/check-remittance), no open transaction batch.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		checkRemittance := &model.CheckRemittance{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Check remittance creation failed (/check-remittance), db error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create check remittance: " + err.Error()})
		}

		allCheckRemittances, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Check remittance creation failed (/check-remittance), recalc error: " + err.Error(),
				Module:      "CheckRemittance",
			})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Check remittance creation failed (/check-remittance), batch totals update error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch totals: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created check remittance (/check-remittance): " + checkRemittance.AccountName,
			Module:      "CheckRemittance",
		})

		return ctx.JSON(http.StatusCreated, c.model.CheckRemittanceManager.ToModel(checkRemittance))
	})

	// PUT /check-remittance/:check_remittance_id: Update a check remittance by ID for the current transaction batch. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/check-remittance/:check_remittance_id",
		Method:       "PUT",
		Note:         "Updates an existing check remittance by ID for the current transaction batch. Only 'owner' or 'employee' roles are allowed.",
		ResponseType: model.CheckRemittanceResponse{},
		RequestType:  model.CheckRemittanceRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		checkRemittanceId, err := handlers.EngineUUIDParam(ctx, "check_remittance_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), invalid ID.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check remittance ID"})
		}

		req, err := c.model.CheckRemittanceManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), validation error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check remittance data: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), user org error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for check remittance (/check-remittance/:check_remittance_id)",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update check remittances"})
		}

		existingCheckRemittance, err := c.model.CheckRemittanceManager.GetByID(context, *checkRemittanceId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), not found.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Check remittance not found"})
		}

		if existingCheckRemittance.OrganizationID != userOrg.OrganizationID || existingCheckRemittance.BranchID != *userOrg.BranchID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), wrong org/branch.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Check remittance does not belong to your organization/branch"})
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), batch lookup error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), no open transaction batch.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		updatedCheckRemittance := &model.CheckRemittance{
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			CreatedByID:        existingCheckRemittance.CreatedByID,
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), db error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update check remittance: " + err.Error()})
		}

		allCheckRemittances, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), recalc error: " + err.Error(),
				Module:      "CheckRemittance",
			})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), batch totals update error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch totals: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated check remittance (/check-remittance/:check_remittance_id): " + updatedCheckRemittance.AccountName,
			Module:      "CheckRemittance",
		})

		updatedRemittance, err := c.model.CheckRemittanceManager.GetByID(context, *checkRemittanceId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated check remittance: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.ToModel(updatedRemittance))
	})

	// DELETE /check-remittance/:check_remittance_id: Delete a check remittance by ID for the current transaction batch. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/check-remittance/:check_remittance_id",
		Method: "DELETE",
		Note:   "Deletes a check remittance by ID for the current transaction batch. Only 'owner' or 'employee' roles are allowed.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		checkRemittanceId, err := handlers.EngineUUIDParam(ctx, "check_remittance_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), invalid ID.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check remittance ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), user org error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete attempt for check remittance (/check-remittance/:check_remittance_id)",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete check remittance"})
		}

		existingCheckRemittance, err := c.model.CheckRemittanceManager.GetByID(context, *checkRemittanceId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), not found.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Check remittance not found"})
		}

		if existingCheckRemittance.OrganizationID != userOrg.OrganizationID || existingCheckRemittance.BranchID != *userOrg.BranchID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), wrong org/branch.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Check remittance does not belong to your organization/branch"})
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), batch lookup error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), no open transaction batch.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		if existingCheckRemittance.TransactionBatchID == nil || *existingCheckRemittance.TransactionBatchID != transactionBatch.ID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), wrong batch.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Check remittance does not belong to current transaction batch"})
		}

		if err := c.model.CheckRemittanceManager.DeleteByID(context, *checkRemittanceId); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), db error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete check remittance: " + err.Error()})
		}

		allCheckRemittances, err := c.model.CheckRemittanceManager.Find(context, &model.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), recalc error: " + err.Error(),
				Module:      "CheckRemittance",
			})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), batch totals update error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch totals: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted check remittance (/check-remittance/:check_remittance_id): " + existingCheckRemittance.AccountName,
			Module:      "CheckRemittance",
		})

		return ctx.JSON(http.StatusOK, c.model.CheckRemittanceManager.ToModel(existingCheckRemittance))
	})
}
