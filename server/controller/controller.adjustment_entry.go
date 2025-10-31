package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelcore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelcore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AdjustmentEntryController registers routes for managing adjustment entries.
func (c *Controller) AdjustmentEntryController() {
	req := c.provider.Service.Request

	// GET /adjustment-entry: List all adjustment entries for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry",
		Method:       "GET",
		Note:         "Returns all adjustment entries for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: modelcore.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		adjustmentEntries, err := c.modelcore.AdjustmentEntryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No adjustment entries found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.AdjustmentEntryManager.Filtered(context, ctx, adjustmentEntries))
	})

	// GET /adjustment-entry/search: Paginated search of adjustment entries for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/search",
		Method:       "GET",
		Note:         "Returns a paginated list of adjustment entries for the current user's organization and branch.",
		ResponseType: modelcore.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		adjustmentEntries, err := c.modelcore.AdjustmentEntryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entries for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.AdjustmentEntryManager.Pagination(context, ctx, adjustmentEntries))
	})

	// GET /adjustment-entry/:adjustment_entry_id: Get specific adjustment entry by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/:adjustment_entry_id",
		Method:       "GET",
		Note:         "Returns a single adjustment entry by its ID.",
		ResponseType: modelcore.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		adjustmentEntryID, err := handlers.EngineUUIDParam(ctx, "adjustment_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry ID"})
		}
		adjustmentEntry, err := c.modelcore.AdjustmentEntryManager.GetByIDRaw(context, *adjustmentEntryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Adjustment entry not found"})
		}
		return ctx.JSON(http.StatusOK, adjustmentEntry)
	})

	// POST /adjustment-entry: Create a new adjustment entry. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry",
		Method:       "POST",
		Note:         "Creates a new adjustment entry for the current user's organization and branch.",
		RequestType:  modelcore.AdjustmentEntryRequest{},
		ResponseType: modelcore.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelcore.AdjustmentEntryManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), validation error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), user org error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), user not assigned to branch.",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		adjustmentEntry := &modelcore.AdjustmentEntry{
			SignatureMediaID:  req.SignatureMediaID,
			AccountID:         req.AccountID,
			MemberProfileID:   req.MemberProfileID,
			EmployeeUserID:    req.EmployeeUserID,
			PaymentTypeID:     req.PaymentTypeID,
			TypeOfPaymentType: req.TypeOfPaymentType,
			Description:       req.Description,
			ReferenceNumber:   req.ReferenceNumber,
			EntryDate:         req.EntryDate,
			Debit:             req.Debit,
			Credit:            req.Credit,
			CreatedAt:         time.Now().UTC(),
			CreatedByID:       user.UserID,
			UpdatedAt:         time.Now().UTC(),
			UpdatedByID:       user.UserID,
			BranchID:          *user.BranchID,
			OrganizationID:    user.OrganizationID,
		}

		if err := c.modelcore.AdjustmentEntryManager.Create(context, adjustmentEntry); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry creation failed (/adjustment-entry), db error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create adjustment entry: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created adjustment entry (/adjustment-entry): " + adjustmentEntry.ReferenceNumber,
			Module:      "AdjustmentEntry",
		})
		return ctx.JSON(http.StatusCreated, c.modelcore.AdjustmentEntryManager.ToModel(adjustmentEntry))
	})

	// PUT /adjustment-entry/:adjustment_entry_id: Update adjustment entry by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/:adjustment_entry_id",
		Method:       "PUT",
		Note:         "Updates an existing adjustment entry by its ID.",
		RequestType:  modelcore.AdjustmentEntryRequest{},
		ResponseType: modelcore.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		adjustmentEntryID, err := handlers.EngineUUIDParam(ctx, "adjustment_entry_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Adjustment entry update failed (/adjustment-entry/:adjustment_entry_id), invalid adjustment entry ID.",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry ID"})
		}

		req, err := c.modelcore.AdjustmentEntryManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Adjustment entry update failed (/adjustment-entry/:adjustment_entry_id), validation error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Adjustment entry update failed (/adjustment-entry/:adjustment_entry_id), user org error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		adjustmentEntry, err := c.modelcore.AdjustmentEntryManager.GetByID(context, *adjustmentEntryID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Adjustment entry update failed (/adjustment-entry/:adjustment_entry_id), adjustment entry not found.",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Adjustment entry not found"})
		}
		adjustmentEntry.SignatureMediaID = req.SignatureMediaID
		adjustmentEntry.AccountID = req.AccountID
		adjustmentEntry.MemberProfileID = req.MemberProfileID
		adjustmentEntry.EmployeeUserID = req.EmployeeUserID
		adjustmentEntry.PaymentTypeID = req.PaymentTypeID
		adjustmentEntry.TypeOfPaymentType = req.TypeOfPaymentType
		adjustmentEntry.Description = req.Description
		adjustmentEntry.ReferenceNumber = req.ReferenceNumber
		adjustmentEntry.EntryDate = req.EntryDate
		adjustmentEntry.Debit = req.Debit
		adjustmentEntry.Credit = req.Credit
		adjustmentEntry.UpdatedAt = time.Now().UTC()
		adjustmentEntry.UpdatedByID = user.UserID
		if err := c.modelcore.AdjustmentEntryManager.UpdateFields(context, adjustmentEntry.ID, adjustmentEntry); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Adjustment entry update failed (/adjustment-entry/:adjustment_entry_id), db error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update adjustment entry: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated adjustment entry (/adjustment-entry/:adjustment_entry_id): " + adjustmentEntry.ReferenceNumber,
			Module:      "AdjustmentEntry",
		})
		return ctx.JSON(http.StatusOK, c.modelcore.AdjustmentEntryManager.ToModel(adjustmentEntry))
	})

	// DELETE /adjustment-entry/:adjustment_entry_id: Delete an adjustment entry by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/adjustment-entry/:adjustment_entry_id",
		Method: "DELETE",
		Note:   "Deletes the specified adjustment entry by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		adjustmentEntryID, err := handlers.EngineUUIDParam(ctx, "adjustment_entry_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Adjustment entry delete failed (/adjustment-entry/:adjustment_entry_id), invalid adjustment entry ID.",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry ID"})
		}
		adjustmentEntry, err := c.modelcore.AdjustmentEntryManager.GetByID(context, *adjustmentEntryID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Adjustment entry delete failed (/adjustment-entry/:adjustment_entry_id), not found.",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Adjustment entry not found"})
		}
		if err := c.modelcore.AdjustmentEntryManager.DeleteByID(context, *adjustmentEntryID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Adjustment entry delete failed (/adjustment-entry/:adjustment_entry_id), db error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete adjustment entry: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted adjustment entry (/adjustment-entry/:adjustment_entry_id): " + adjustmentEntry.ReferenceNumber,
			Module:      "AdjustmentEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /adjustment-entry/bulk-delete: Bulk delete adjustment entries by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/adjustment-entry/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple adjustment entries by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: modelcore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/adjustment-entry/bulk-delete), invalid request body.",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/adjustment-entry/bulk-delete), no IDs provided.",
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No adjustment entry IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/adjustment-entry/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			adjustmentEntryID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/adjustment-entry/bulk-delete), invalid UUID: " + rawID,
					Module:      "AdjustmentEntry",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			adjustmentEntry, err := c.modelcore.AdjustmentEntryManager.GetByID(context, adjustmentEntryID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/adjustment-entry/bulk-delete), not found: " + rawID,
					Module:      "AdjustmentEntry",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Adjustment entry not found with ID: %s", rawID)})
			}
			names += adjustmentEntry.ReferenceNumber + ","
			if err := c.modelcore.AdjustmentEntryManager.DeleteByIDWithTx(context, tx, adjustmentEntryID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/adjustment-entry/bulk-delete), db error: " + err.Error(),
					Module:      "AdjustmentEntry",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete adjustment entry: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/adjustment-entry/bulk-delete), commit error: " + err.Error(),
				Module:      "AdjustmentEntry",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted adjustment entries (/adjustment-entry/bulk-delete): " + names,
			Module:      "AdjustmentEntry",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// GET /api/v1/adjustment-entry/total
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/total",
		Method:       "GET",
		Note:         "Returns the total debit and credit of all adjustment entries for the current user's organization and branch.",
		ResponseType: modelcore.AdjustmentEntryTotalResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		adjustmentEntries, err := c.modelcore.AdjustmentEntryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No adjustment entries found for the current branch"})
		}
		totalDebit := 0.0
		totalCredit := 0.0
		for _, entry := range adjustmentEntries {
			totalDebit += entry.Debit
			totalCredit += entry.Credit
		}
		return ctx.JSON(http.StatusOK, modelcore.AdjustmentEntryTotalResponse{
			TotalDebit:  totalDebit,
			TotalCredit: totalCredit,
		})
	})

	// GET api/v1/adjustment-entry/currency/:currency_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/currency/:currency_id/search",
		Method:       "GET",
		Note:         "Returns a paginated list of adjustment entries filtered by currency and optionally by user organization.",
		ResponseType: modelcore.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		adjustmentEntries, err := c.modelcore.AdjustmentEntryManager.Find(context, &modelcore.AdjustmentEntry{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entries for pagination: " + err.Error()})
		}
		result := []*modelcore.AdjustmentEntry{}
		for _, entry := range adjustmentEntries {
			if handlers.UuidPtrEqual(entry.Account.CurrencyID, currencyID) {
				result = append(result, entry)
			}
		}
		return ctx.JSON(http.StatusOK, c.modelcore.AdjustmentEntryManager.Pagination(context, ctx, result))
	})

	// GET api/v1/adjustment-entry/currency/:currency_id/total
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/currency/:currency_id/total",
		Method:       "GET",
		Note:         "Returns the total amount of adjustment entries filtered by currency and optionally by user organization.",
		ResponseType: modelcore.AdjustmentEntryTotalResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		adjustmentEntries, err := c.modelcore.AdjustmentEntryManager.Find(context, &modelcore.AdjustmentEntry{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entries for pagination: " + err.Error()})
		}
		totalDebit := 0.0
		totalCredit := 0.0
		for _, entry := range adjustmentEntries {
			if handlers.UuidPtrEqual(entry.Account.CurrencyID, currencyID) {
				totalDebit += entry.Debit
				totalCredit += entry.Credit
			}
		}
		return ctx.JSON(http.StatusOK, modelcore.AdjustmentEntryTotalResponse{
			TotalDebit:  totalDebit,
			TotalCredit: totalCredit,
		})
	})

	// GET api/v1/adjustment-entry/currency/:currency_id/employee/:user_organization_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/currency/:currency_id/employee/:user_organization_id/search",
		Method:       "GET",
		Note:         "Returns a paginated list of adjustment entries filtered by currency and user organization.",
		ResponseType: modelcore.AdjustmentEntryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrganization, err := c.modelcore.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User organization not found"})
		}
		if userOrganization.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User organization is not assigned to a branch"})
		}
		adjustmentEntries, err := c.modelcore.AdjustmentEntryManager.Find(context, &modelcore.AdjustmentEntry{
			OrganizationID: userOrganization.OrganizationID,
			BranchID:       *userOrganization.BranchID,
			EmployeeUserID: &userOrganization.UserID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entries for pagination: " + err.Error()})
		}
		result := []*modelcore.AdjustmentEntry{}
		for _, entry := range adjustmentEntries {
			if handlers.UuidPtrEqual(entry.Account.CurrencyID, currencyID) {
				result = append(result, entry)
			}
		}
		return ctx.JSON(http.StatusOK, c.modelcore.AdjustmentEntryManager.Pagination(context, ctx, result))
	})

	// GET api/v1/adjustment-entry/currency/:currency_id/employee/:user_organization_id/total
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry/currency/:currency_id/employee/:user_organization_id/total",
		Method:       "GET",
		Note:         "Returns the total amount of adjustment entries filtered by currency and user organization.",
		ResponseType: modelcore.AdjustmentEntryTotalResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrganization, err := c.modelcore.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User organization not found"})
		}
		if userOrganization.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User organization is not assigned to a branch"})
		}
		adjustmentEntries, err := c.modelcore.AdjustmentEntryManager.Find(context, &modelcore.AdjustmentEntry{
			OrganizationID: userOrganization.OrganizationID,
			BranchID:       *userOrganization.BranchID,
			EmployeeUserID: &userOrganization.UserID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entries for pagination: " + err.Error()})
		}
		totalDebit := 0.0
		totalCredit := 0.0
		for _, entry := range adjustmentEntries {
			if handlers.UuidPtrEqual(entry.Account.CurrencyID, currencyID) {
				totalDebit += entry.Debit
				totalCredit += entry.Credit
			}
		}
		return ctx.JSON(http.StatusOK, modelcore.AdjustmentEntryTotalResponse{
			TotalDebit:  totalDebit,
			TotalCredit: totalCredit,
		})
	})

}
