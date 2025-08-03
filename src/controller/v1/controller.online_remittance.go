package controller_v1

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) OnlineRemittanceController() {
	req := c.provider.Service.Request

	// Retrieve batch online remittance (JWT) for the current transaction batch before ending.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/online-remittance",
		Method:       "GET",
		ResponseType: model.OnlineRemittanceResponse{},
		Note:         "Returns online remittance records for the current active transaction batch.",
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

		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.Filtered(context, ctx, onlineRemittance))
	})

	// Create a new online remittance for the current transaction batch before ending.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/online-remittance",
		Method:       "POST",
		ResponseType: model.OnlineRemittanceResponse{},
		RequestType:  model.OnlineRemittanceRequest{},
		Note:         "Creates a new online remittance record for the current transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := c.model.OnlineRemittanceManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create online remittance failed: validation error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create online remittance failed: user org error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create online remittance failed: unauthorized user type",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create online remittance failed: find batch error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create online remittance failed: no active batch",
				Module:      "OnlineRemittance",
			})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create online remittance failed: create error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create online remittance: " + err.Error()})
		}

		allOnlineRemittances, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create online remittance failed: find all error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create online remittance failed: update batch error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created online remittance for batch ID: " + transactionBatch.ID.String(),
			Module:      "OnlineRemittance",
		})

		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModel(onlineRemittance))
	})

	// Update an existing online remittance by ID for the current transaction batch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/online-remittance/:online_remittance_id",
		Method:       "PUT",
		ResponseType: model.OnlineRemittanceResponse{},
		RequestType:  model.OnlineRemittanceRequest{},
		Note:         "Updates an existing online remittance by its ID for the current active transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		onlineRemittanceId, err := handlers.EngineUUIDParam(ctx, "online_remittance_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: invalid online_remittance_id: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid online_remittance_id: " + err.Error()})
		}

		req, err := c.model.OnlineRemittanceManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: validation error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: user org error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: unauthorized user type",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		existingOnlineRemittance, err := c.model.OnlineRemittanceManager.GetByID(context, *onlineRemittanceId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: not found: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Online remittance not found: " + err.Error()})
		}

		if existingOnlineRemittance.OrganizationID != userOrg.OrganizationID ||
			existingOnlineRemittance.BranchID != *userOrg.BranchID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: not in org/branch",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Online remittance not found in your organization/branch"})
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: find batch error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: update error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update online remittance: " + err.Error()})
		}

		allOnlineRemittances, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: find all error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: update batch error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		updatedRemittance, err := c.model.OnlineRemittanceManager.GetByID(context, *onlineRemittanceId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update online remittance failed: get updated error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated online remittance: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated online remittance for batch ID: " + transactionBatch.ID.String(),
			Module:      "OnlineRemittance",
		})

		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModel(updatedRemittance))
	})

	// Delete an existing online remittance by ID for the current transaction batch.
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/online-remittance/:online_remittance_id",
		Method: "DELETE",
		Note:   "Deletes an online remittance by its ID for the current active transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		onlineRemittanceId, err := handlers.EngineUUIDParam(ctx, "online_remittance_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: invalid online_remittance_id: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid online_remittance_id: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: user org error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: unauthorized user type",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		existingOnlineRemittance, err := c.model.OnlineRemittanceManager.GetByID(context, *onlineRemittanceId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: not found: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Online remittance not found: " + err.Error()})
		}

		if existingOnlineRemittance.OrganizationID != userOrg.OrganizationID ||
			existingOnlineRemittance.BranchID != *userOrg.BranchID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: not in org/branch",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Online remittance not found in your organization/branch"})
		}

		transactionBatch, err := c.model.TransactionBatchManager.FindOneWithConditions(context, map[string]any{
			"organization_id": userOrg.OrganizationID,
			"branch_id":       *userOrg.BranchID,
			"is_closed":       false,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: find batch error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find active transaction batch: " + err.Error()})
		}
		if transactionBatch == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: no active batch",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No active transaction batch found"})
		}

		if existingOnlineRemittance.TransactionBatchID == nil ||
			*existingOnlineRemittance.TransactionBatchID != transactionBatch.ID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: not in current batch",
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Online remittance does not belong to current transaction batch"})
		}

		if err := c.model.OnlineRemittanceManager.DeleteByID(context, *onlineRemittanceId); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: delete error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete online remittance: " + err.Error()})
		}

		allOnlineRemittances, err := c.model.OnlineRemittanceManager.Find(context, &model.OnlineRemittance{
			TransactionBatchID: &transactionBatch.ID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: find all error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete online remittance failed: update batch error: " + err.Error(),
				Module:      "OnlineRemittance",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction batch: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted online remittance for batch ID: " + transactionBatch.ID.String(),
			Module:      "OnlineRemittance",
		})

		return ctx.JSON(http.StatusOK, c.model.OnlineRemittanceManager.ToModel(existingOnlineRemittance))
	})
}
