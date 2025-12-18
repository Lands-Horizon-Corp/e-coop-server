package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) checkRemittanceController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/check-remittance",
		Method:       "GET",
		Note:         "Returns all check remittances for the current active transaction batch of the authenticated user's branch. Only 'owner' or 'employee' roles are allowed.",
		ResponseType: core.CheckRemittanceResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view check remittances"})
		}

		transactionBatch, err := c.core.TransactionBatchCurrent(
			context,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		checkRemittance, err := c.core.CheckRemittanceManager.Find(context, &core.CheckRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve check remittances: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.core.CheckRemittanceManager.ToModels(checkRemittance))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/check-remittance",
		Method:       "POST",
		ResponseType: core.CheckRemittanceResponse{},
		RequestType:  core.CheckRemittanceRequest{},
		Note:         "Creates a new check remittance for the current active transaction batch. Only 'owner' or 'employee' roles are allowed.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.CheckRemittanceManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Check remittance creation failed (/check-remittance), validation error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check remittance data: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Check remittance creation failed (/check-remittance), user org error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for check remittance (/check-remittance)",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create check remittances"})
		}

		transactionBatch, err := c.core.TransactionBatchCurrent(
			context,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Check remittance creation failed (/check-remittance), transaction batch lookup error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Check remittance creation failed (/check-remittance), no open transaction batch.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		checkRemittance := &core.CheckRemittance{
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
			CurrencyID:         req.CurrencyID,
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

		if err := c.core.CheckRemittanceManager.Create(context, checkRemittance); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Check remittance creation failed (/check-remittance), db error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create check remittance: " + err.Error()})
		}

		if err := c.event.TransactionBatchBalancing(context, &transactionBatch.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch after saving: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created check remittance (/check-remittance): " + checkRemittance.AccountName,
			Module:      "CheckRemittance",
		})

		return ctx.JSON(http.StatusCreated, c.core.CheckRemittanceManager.ToModel(checkRemittance))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/check-remittance/:check_remittance_id",
		Method:       "PUT",
		Note:         "Updates an existing check remittance by ID for the current transaction batch. Only 'owner' or 'employee' roles are allowed.",
		ResponseType: core.CheckRemittanceResponse{},
		RequestType:  core.CheckRemittanceRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		checkRemittanceID, err := handlers.EngineUUIDParam(ctx, "check_remittance_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), invalid ID.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check remittance ID"})
		}

		req, err := c.core.CheckRemittanceManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), validation error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check remittance data: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), user org error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for check remittance (/check-remittance/:check_remittance_id)",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update check remittances"})
		}

		existingCheckRemittance, err := c.core.CheckRemittanceManager.GetByID(context, *checkRemittanceID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), not found.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Check remittance not found"})
		}

		if existingCheckRemittance.OrganizationID != userOrg.OrganizationID || existingCheckRemittance.BranchID != *userOrg.BranchID {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), wrong org/branch.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Check remittance does not belong to your organization/branch"})
		}

		transactionBatch, err := c.core.TransactionBatchCurrent(
			context,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), batch lookup error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), no open transaction batch.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		updatedCheckRemittance := &core.CheckRemittance{
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			CreatedByID:        existingCheckRemittance.CreatedByID,
			TransactionBatchID: &transactionBatch.ID,
			BankID:             req.BankID,
			MediaID:            req.MediaID,
			CurrencyID:         req.CurrencyID,
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

		if err := c.core.CheckRemittanceManager.UpdateByID(context, *checkRemittanceID, updatedCheckRemittance); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Check remittance update failed (/check-remittance/:check_remittance_id), db error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update check remittance: " + err.Error()})
		}

		if err := c.event.TransactionBatchBalancing(context, &transactionBatch.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch after saving: " + err.Error()})
		}
		updatedRemittance, err := c.core.CheckRemittanceManager.GetByID(context, *checkRemittanceID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated check remittance: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.core.CheckRemittanceManager.ToModel(updatedRemittance))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/check-remittance/:check_remittance_id",
		Method: "DELETE",
		Note:   "Deletes a check remittance by ID for the current transaction batch. Only 'owner' or 'employee' roles are allowed.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		checkRemittanceID, err := handlers.EngineUUIDParam(ctx, "check_remittance_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), invalid ID.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check remittance ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), user org error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete attempt for check remittance (/check-remittance/:check_remittance_id)",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete check remittance"})
		}

		existingCheckRemittance, err := c.core.CheckRemittanceManager.GetByID(context, *checkRemittanceID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), not found.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Check remittance not found"})
		}

		if existingCheckRemittance.OrganizationID != userOrg.OrganizationID || existingCheckRemittance.BranchID != *userOrg.BranchID {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), wrong org/branch.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Check remittance does not belong to your organization/branch"})
		}

		transactionBatch, err := c.core.TransactionBatchCurrent(
			context,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), batch lookup error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), no open transaction batch.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found for your branch"})
		}

		if existingCheckRemittance.TransactionBatchID == nil || *existingCheckRemittance.TransactionBatchID != transactionBatch.ID {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), wrong batch.",
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Check remittance does not belong to current transaction batch"})
		}

		if err := c.core.CheckRemittanceManager.Delete(context, *checkRemittanceID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Check remittance delete failed (/check-remittance/:check_remittance_id), db error: " + err.Error(),
				Module:      "CheckRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete check remittance: " + err.Error()})
		}

		if err := c.event.TransactionBatchBalancing(context, &transactionBatch.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch after saving: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted check remittance (/check-remittance/:check_remittance_id): " + existingCheckRemittance.AccountName,
			Module:      "CheckRemittance",
		})

		return ctx.JSON(http.StatusOK, c.core.CheckRemittanceManager.ToModel(existingCheckRemittance))
	})
}
