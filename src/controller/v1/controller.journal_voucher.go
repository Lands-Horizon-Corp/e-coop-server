package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// JournalVoucherController registers routes for managing journal vouchers.
func (c *Controller) JournalVoucherController() {
	req := c.provider.Service.Request

	// GET /journal-voucher: List all journal vouchers for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/journal-voucher",
		Method:       "GET",
		Note:         "Returns all journal vouchers for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: model.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		journalVouchers, err := c.model.JournalVoucherCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No journal vouchers found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.model.JournalVoucherManager.Filtered(context, ctx, journalVouchers))
	})

	// GET /journal-voucher/search: Paginated search of journal vouchers for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/journal-voucher/search",
		Method:       "GET",
		Note:         "Returns a paginated list of journal vouchers for the current user's organization and branch.",
		ResponseType: model.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		journalVouchers, err := c.model.JournalVoucherCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch journal vouchers for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.JournalVoucherManager.Pagination(context, ctx, journalVouchers))
	})

	// GET /journal-voucher/:journal_voucher_id: Get specific journal voucher by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/journal-voucher/:journal_voucher_id",
		Method:       "GET",
		Note:         "Returns a single journal voucher by its ID.",
		ResponseType: model.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := handlers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}
		journalVoucher, err := c.model.JournalVoucherManager.GetByIDRaw(context, *journalVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher not found"})
		}
		return ctx.JSON(http.StatusOK, journalVoucher)
	})

	// POST /journal-voucher: Create a new journal voucher. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/journal-voucher",
		Method:       "POST",
		Note:         "Creates a new journal voucher for the current user's organization and branch.",
		RequestType:  model.JournalVoucherRequest{},
		ResponseType: model.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		request, err := c.model.JournalVoucherManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher creation failed (/journal-voucher), validation error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher creation failed (/journal-voucher), user org error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher creation failed (/journal-voucher), user not assigned to branch.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		// Start transaction
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher creation failed (/journal-voucher), transaction start error: " + tx.Error.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}

		journalVoucher := &model.JournalVoucher{
			VoucherNumber:       request.VoucherNumber,
			Date:                request.Date,
			Description:         request.Description,
			Reference:           request.Reference,
			Status:              request.Status,
			JournalVoucherTagID: request.JournalVoucherTagID,
			CreatedAt:           time.Now().UTC(),
			CreatedByID:         user.UserID,
			UpdatedAt:           time.Now().UTC(),
			UpdatedByID:         user.UserID,
			BranchID:            *user.BranchID,
			OrganizationID:      user.OrganizationID,
		}

		if err := c.model.JournalVoucherManager.CreateWithTx(context, tx, journalVoucher); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher creation failed (/journal-voucher), db error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create journal voucher: " + err.Error()})
		}

		// Handle journal voucher entries
		if request.JournalVoucherEntries != nil {
			for _, entryReq := range request.JournalVoucherEntries {
				entry := &model.JournalVoucherEntry{
					AccountID:        entryReq.AccountID,
					MemberProfileID:  entryReq.MemberProfileID,
					EmployeeUserID:   entryReq.EmployeeUserID,
					JournalVoucherID: journalVoucher.ID,
					Description:      entryReq.Description,
					Debit:            entryReq.Debit,
					Credit:           entryReq.Credit,
					CreatedAt:        time.Now().UTC(),
					CreatedByID:      user.UserID,
					UpdatedAt:        time.Now().UTC(),
					UpdatedByID:      user.UserID,
					BranchID:         *user.BranchID,
					OrganizationID:   user.OrganizationID,
				}

				if err := c.model.JournalVoucherEntryManager.CreateWithTx(context, tx, entry); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "create-error",
						Description: "Journal voucher entry creation failed (/journal-voucher), db error: " + err.Error(),
						Module:      "JournalVoucher",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create journal voucher entry: " + err.Error()})
				}
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher creation failed (/journal-voucher), commit error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created journal voucher (/journal-voucher): " + journalVoucher.VoucherNumber,
			Module:      "JournalVoucher",
		})
		return ctx.JSON(http.StatusCreated, c.model.JournalVoucherManager.ToModel(journalVoucher))
	})

	// PUT /journal-voucher/:journal_voucher_id: Update journal voucher by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/journal-voucher/:journal_voucher_id",
		Method:       "PUT",
		Note:         "Updates an existing journal voucher by its ID.",
		RequestType:  model.JournalVoucherRequest{},
		ResponseType: model.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := handlers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), invalid ID.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}

		request, err := c.model.JournalVoucherManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), validation error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher data: " + err.Error()})
		}

		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), user org error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		journalVoucher, err := c.model.JournalVoucherManager.GetByID(context, *journalVoucherID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), not found.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher not found"})
		}

		// Start transaction
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), transaction start error: " + tx.Error.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}

		// Update journal voucher fields
		journalVoucher.VoucherNumber = request.VoucherNumber
		journalVoucher.Date = request.Date
		journalVoucher.Description = request.Description
		journalVoucher.Reference = request.Reference
		journalVoucher.Status = request.Status
		journalVoucher.JournalVoucherTagID = request.JournalVoucherTagID
		journalVoucher.UpdatedAt = time.Now().UTC()
		journalVoucher.UpdatedByID = user.UserID

		if err := c.model.JournalVoucherManager.UpdateFieldsWithTx(context, tx, journalVoucher.ID, journalVoucher); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), db error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update journal voucher: " + err.Error()})
		}

		// Handle deleted entries
		if request.JournalVoucherEntriesDeleted != nil {
			for _, deletedID := range request.JournalVoucherEntriesDeleted {
				entry, err := c.model.JournalVoucherEntryManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find journal voucher entry for deletion: " + err.Error()})
				}
				if entry.JournalVoucherID != journalVoucher.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete entry that doesn't belong to this journal voucher"})
				}
				entry.DeletedByID = &user.UserID
				if err := c.model.JournalVoucherEntryManager.DeleteWithTx(context, tx, entry); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete journal voucher entry: " + err.Error()})
				}
			}
		}

		// Handle journal voucher entries (create new or update existing)
		if request.JournalVoucherEntries != nil {
			for _, entryReq := range request.JournalVoucherEntries {
				if entryReq.ID != nil {
					entry, err := c.model.JournalVoucherEntryManager.GetByID(context, *entryReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find journal voucher entry for update: " + err.Error()})
					}
					if entry.JournalVoucherID != journalVoucher.ID {
						tx.Rollback()
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update entry that doesn't belong to this journal voucher"})
					}
					entry.AccountID = entryReq.AccountID
					entry.MemberProfileID = entryReq.MemberProfileID
					entry.EmployeeUserID = entryReq.EmployeeUserID
					entry.Description = entryReq.Description
					entry.Debit = entryReq.Debit
					entry.Credit = entryReq.Credit
					entry.UpdatedAt = time.Now().UTC()
					entry.UpdatedByID = user.UserID
					if err := c.model.JournalVoucherEntryManager.UpdateFieldsWithTx(context, tx, entry.ID, entry); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update journal voucher entry: " + err.Error()})
					}
				} else {
					entry := &model.JournalVoucherEntry{
						AccountID:        entryReq.AccountID,
						MemberProfileID:  entryReq.MemberProfileID,
						EmployeeUserID:   entryReq.EmployeeUserID,
						JournalVoucherID: journalVoucher.ID,
						Description:      entryReq.Description,
						Debit:            entryReq.Debit,
						Credit:           entryReq.Credit,
						CreatedAt:        time.Now().UTC(),
						CreatedByID:      user.UserID,
						UpdatedAt:        time.Now().UTC(),
						UpdatedByID:      user.UserID,
						BranchID:         *user.BranchID,
						OrganizationID:   user.OrganizationID,
					}
					if err := c.model.JournalVoucherEntryManager.CreateWithTx(context, tx, entry); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create journal voucher entry: " + err.Error()})
					}
				}

			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), commit error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated journal voucher (/journal-voucher/:journal_voucher_id): " + journalVoucher.VoucherNumber,
			Module:      "JournalVoucher",
		})
		return ctx.JSON(http.StatusOK, c.model.JournalVoucherManager.ToModel(journalVoucher))
	})

	// DELETE /journal-voucher/:journal_voucher_id: Delete a journal voucher by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/journal-voucher/:journal_voucher_id",
		Method: "DELETE",
		Note:   "Deletes the specified journal voucher by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := handlers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Journal voucher delete failed (/journal-voucher/:journal_voucher_id), invalid ID.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}
		journalVoucher, err := c.model.JournalVoucherManager.GetByID(context, *journalVoucherID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Journal voucher delete failed (/journal-voucher/:journal_voucher_id), not found.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher not found"})
		}
		if err := c.model.JournalVoucherManager.DeleteByID(context, *journalVoucherID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Journal voucher delete failed (/journal-voucher/:journal_voucher_id), db error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete journal voucher: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted journal voucher (/journal-voucher/:journal_voucher_id): " + journalVoucher.VoucherNumber,
			Module:      "JournalVoucher",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /journal-voucher/bulk-delete: Bulk delete journal vouchers by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/journal-voucher/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple journal vouchers by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher/bulk-delete), invalid request body.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher/bulk-delete), no IDs provided.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No journal voucher IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		voucherNumbers := ""
		for _, rawID := range reqBody.IDs {
			journalVoucherID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/journal-voucher/bulk-delete), invalid UUID: " + rawID,
					Module:      "JournalVoucher",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			journalVoucher, err := c.model.JournalVoucherManager.GetByID(context, journalVoucherID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/journal-voucher/bulk-delete), not found: " + rawID,
					Module:      "JournalVoucher",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Journal voucher not found with ID: %s", rawID)})
			}
			voucherNumbers += journalVoucher.VoucherNumber + ","
			if err := c.model.JournalVoucherManager.DeleteByIDWithTx(context, tx, journalVoucherID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/journal-voucher/bulk-delete), db error: " + err.Error(),
					Module:      "JournalVoucher",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete journal voucher: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher/bulk-delete), commit error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted journal vouchers (/journal-voucher/bulk-delete): " + voucherNumbers,
			Module:      "JournalVoucher",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

}
