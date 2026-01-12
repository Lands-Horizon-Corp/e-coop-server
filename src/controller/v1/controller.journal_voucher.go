package v1

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func journalVoucherController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher",
		Method:       "GET",
		Note:         "Returns all journal vouchers for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		journalVouchers, err := core.JournalVoucherCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No journal vouchers found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.JournalVoucherManager(service).ToModels(journalVouchers))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/search",
		Method:       "GET",
		Note:         "Returns a paginated list of journal vouchers for the current user's organization and branch.",
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		journalVouchers, err := core.JournalVoucherManager(service).NormalPagination(context, ctx, &core.JournalVoucher{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch journal vouchers for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, journalVouchers)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/:journal_voucher_id",
		Method:       "GET",
		Note:         "Returns a single journal voucher by its ID.",
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := helpers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}
		journalVoucher, err := core.JournalVoucherManager(service).GetByIDRaw(context, *journalVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher not found"})
		}
		return ctx.JSON(http.StatusOK, journalVoucher)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher",
		Method:       "POST",
		Note:         "Creates a new journal voucher for the current user's organization and branch.",
		RequestType:  core.JournalVoucherRequest{},
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		request, err := core.JournalVoucherManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher creation failed (/journal-voucher), validation error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher creation failed (/journal-voucher), user org error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher creation failed (/journal-voucher), user not assigned to branch.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		tx, endTx := c.provider.Service.Database.StartTransaction(context)

		balance, err := usecase.CalculateBalance(usecase.Balance{
			JournalVoucherEntriesRequest: request.JournalVoucherEntries,
		})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher creation failed (/journal-voucher), entries not balanced: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Journal voucher entries are not balanced: " + endTx(err).Error()})
		}

		journalVoucher := &core.JournalVoucher{
			Date:              request.Date,
			Description:       request.Description,
			Reference:         request.Reference,
			Status:            request.Status,
			CreatedAt:         time.Now().UTC(),
			CreatedByID:       userOrg.UserID,
			UpdatedAt:         time.Now().UTC(),
			UpdatedByID:       userOrg.UserID,
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrg.OrganizationID,
			TotalDebit:        balance.Debit,
			TotalCredit:       balance.Credit,
			CashVoucherNumber: request.CashVoucherNumber,
			Name:              request.Name,
			CurrencyID:        request.CurrencyID,
			EmployeeUserID:    &userOrg.UserID,
		}

		if err := core.JournalVoucherManager(service).CreateWithTx(context, tx, journalVoucher); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher creation failed (/journal-voucher), db error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create journal voucher: " + endTx(err).Error()})
		}

		if request.JournalVoucherEntries != nil {
			for _, entryReq := range request.JournalVoucherEntries {
				entry := &core.JournalVoucherEntry{
					AccountID:         entryReq.AccountID,
					MemberProfileID:   entryReq.MemberProfileID,
					EmployeeUserID:    entryReq.EmployeeUserID,
					JournalVoucherID:  journalVoucher.ID,
					Description:       entryReq.Description,
					Debit:             entryReq.Debit,
					Credit:            entryReq.Credit,
					CreatedAt:         time.Now().UTC(),
					CreatedByID:       userOrg.UserID,
					UpdatedAt:         time.Now().UTC(),
					UpdatedByID:       userOrg.UserID,
					BranchID:          *userOrg.BranchID,
					OrganizationID:    userOrg.OrganizationID,
					LoanTransactionID: entryReq.LoanTransactionID,
				}

				if err := core.JournalVoucherEntryManager(service).CreateWithTx(context, tx, entry); err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "create-error",
						Description: "Journal voucher entry creation failed (/journal-voucher), db error: " + err.Error(),
						Module:      "JournalVoucher",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create journal voucher entry: " + endTx(err).Error()})
				}
			}
		}

		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher creation failed (/journal-voucher), commit error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created journal voucher (/journal-voucher): " + journalVoucher.CashVoucherNumber,
			Module:      "JournalVoucher",
		})
		return ctx.JSON(http.StatusCreated, core.JournalVoucherManager(service).ToModel(journalVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/:journal_voucher_id",
		Method:       "PUT",
		Note:         "Updates an existing journal voucher by its ID.",
		RequestType:  core.JournalVoucherRequest{},
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := helpers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), invalid ID.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}

		request, err := core.JournalVoucherManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), validation error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher data: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), user org error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		journalVoucher, err := core.JournalVoucherManager(service).GetByID(context, *journalVoucherID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), not found.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher not found"})
		}

		balance, err := usecase.CalculateStrictBalance(usecase.Balance{
			JournalVoucherEntriesRequest: request.JournalVoucherEntries,
		})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), entries not balanced: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Journal voucher entries are not balanced: " + err.Error()})
		}
		tx, endTx := c.provider.Service.Database.StartTransaction(context)

		journalVoucher.Date = request.Date
		journalVoucher.Description = request.Description
		journalVoucher.Reference = request.Reference
		journalVoucher.Status = request.Status
		journalVoucher.UpdatedAt = time.Now().UTC()
		journalVoucher.UpdatedByID = userOrg.UserID
		journalVoucher.CashVoucherNumber = request.CashVoucherNumber
		journalVoucher.Name = request.Name
		journalVoucher.EmployeeUserID = &userOrg.UserID
		if request.JournalVoucherEntriesDeleted != nil {
			for _, deletedID := range request.JournalVoucherEntriesDeleted {
				entry, err := core.JournalVoucherEntryManager(service).GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find journal voucher entry for deletion: " + endTx(err).Error()})
				}
				if entry.JournalVoucherID != journalVoucher.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete entry that doesn't belong to this journal voucher: " + endTx(eris.New("invalid journal voucher entry")).Error()})
				}
				entry.DeletedByID = &userOrg.UserID
				if err := core.JournalVoucherEntryManager(service).DeleteWithTx(context, tx, entry.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete journal voucher entry: " + endTx(err).Error()})
				}
			}
		}

		if request.JournalVoucherEntries != nil {
			for _, entryReq := range request.JournalVoucherEntries {
				if entryReq.ID != nil {
					entry, err := core.JournalVoucherEntryManager(service).GetByID(context, *entryReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find journal voucher entry for update: " + endTx(err).Error()})
					}
					if entry.JournalVoucherID != journalVoucher.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update entry that doesn't belong to this journal voucher: " + endTx(eris.New("invalid journal voucher entry")).Error()})
					}
					entry.AccountID = entryReq.AccountID
					entry.MemberProfileID = entryReq.MemberProfileID
					entry.EmployeeUserID = entryReq.EmployeeUserID
					entry.Description = entryReq.Description
					entry.Debit = entryReq.Debit
					entry.Credit = entryReq.Credit
					entry.UpdatedAt = time.Now().UTC()
					entry.UpdatedByID = userOrg.UserID
					entry.LoanTransactionID = entryReq.LoanTransactionID
					if err := core.JournalVoucherEntryManager(service).UpdateByIDWithTx(context, tx, entry.ID, entry); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update journal voucher entry: " + endTx(err).Error()})
					}
				} else {
					entry := &core.JournalVoucherEntry{
						AccountID:         entryReq.AccountID,
						MemberProfileID:   entryReq.MemberProfileID,
						EmployeeUserID:    entryReq.EmployeeUserID,
						JournalVoucherID:  journalVoucher.ID,
						Description:       entryReq.Description,
						Debit:             entryReq.Debit,
						Credit:            entryReq.Credit,
						CreatedAt:         time.Now().UTC(),
						CreatedByID:       userOrg.UserID,
						UpdatedAt:         time.Now().UTC(),
						UpdatedByID:       userOrg.UserID,
						BranchID:          *userOrg.BranchID,
						OrganizationID:    userOrg.OrganizationID,
						LoanTransactionID: entryReq.LoanTransactionID,
					}
					if err := core.JournalVoucherEntryManager(service).CreateWithTx(context, tx, entry); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create journal voucher entry: " + endTx(err).Error()})
					}
				}

			}
		}
		journalVoucher.TotalCredit = balance.Credit
		journalVoucher.TotalDebit = balance.Debit
		if err := core.JournalVoucherManager(service).UpdateByIDWithTx(context, tx, journalVoucher.ID, journalVoucher); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), db error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update journal voucher: " + endTx(err).Error()})
		}

		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher update failed (/journal-voucher/:journal_voucher_id), commit error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		newJournalVoucher, err := core.JournalVoucherManager(service).GetByID(context, journalVoucher.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update journal voucher: " + err.Error()})

		}
		sort.Slice(newJournalVoucher.JournalVoucherEntries, func(i, j int) bool {
			return newJournalVoucher.JournalVoucherEntries[i].CreatedAt.After(newJournalVoucher.JournalVoucherEntries[j].CreatedAt)
		})
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated journal voucher (/journal-voucher/:journal_voucher_id): " + journalVoucher.CashVoucherNumber,
			Module:      "JournalVoucher",
		})
		return ctx.JSON(http.StatusOK, core.JournalVoucherManager(service).ToModel(newJournalVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/journal-voucher/:journal_voucher_id",
		Method: "DELETE",
		Note:   "Deletes the specified journal voucher by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := helpers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Journal voucher delete failed (/journal-voucher/:journal_voucher_id), invalid ID.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}
		journalVoucher, err := core.JournalVoucherManager(service).GetByID(context, *journalVoucherID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Journal voucher delete failed (/journal-voucher/:journal_voucher_id), not found.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher not found"})
		}
		if err := core.JournalVoucherManager(service).Delete(context, *journalVoucherID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Journal voucher delete failed (/journal-voucher/:journal_voucher_id), db error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete journal voucher: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted journal voucher (/journal-voucher/:journal_voucher_id): " + journalVoucher.CashVoucherNumber,
			Module:      "JournalVoucher",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/journal-voucher/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple journal vouchers by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher/bulk-delete) | no IDs provided",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.JournalVoucherManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher/bulk-delete) | error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete journal vouchers: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted journal vouchers (/journal-voucher/bulk-delete)",
			Module:      "JournalVoucher",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/:journal_voucher_id/print",
		Method:       "PUT",
		Note:         "Marks a journal voucher as printed by ID.",
		RequestType:  core.JournalVoucherPrintRequest{},
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := helpers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-error",
				Description: "Journal voucher print failed, invalid ID.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}

		var req core.JournalVoucherPrintRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-error",
				Description: "Journal voucher print failed, invalid request body.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-error",
				Description: "Journal voucher print failed, user org error.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}

		journalVoucher, err := core.JournalVoucherManager(service).GetByID(context, *journalVoucherID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-error",
				Description: "Journal voucher print failed, not found.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher not found"})
		}

		if journalVoucher.OrganizationID != userOrg.OrganizationID || journalVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}

		if journalVoucher.PrintedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Journal voucher has already been printed"})
		}

		journalVoucher.PrintNumber++
		journalVoucher.PrintedDate = helpers.Ptr(time.Now().UTC())
		journalVoucher.PrintedByID = &userOrg.UserID
		journalVoucher.CashVoucherNumber = req.CashVoucherNumber
		journalVoucher.UpdatedAt = time.Now().UTC()
		journalVoucher.UpdatedByID = userOrg.UserID

		if err := core.JournalVoucherManager(service).UpdateByID(context, journalVoucher.ID, journalVoucher); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-error",
				Description: "Journal voucher print failed, db error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to print journal voucher: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "print-success",
			Description: "Successfully printed journal voucher: " + journalVoucher.CashVoucherNumber,
			Module:      "JournalVoucher",
		})

		return ctx.JSON(http.StatusOK, core.JournalVoucherManager(service).ToModel(journalVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/:journal_voucher_id/print-undo",
		Method:       "PUT",
		Note:         "Reverts the print status of a journal voucher by ID.",
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := helpers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-undo-error",
				Description: "Journal voucher print undo failed, invalid ID.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-undo-error",
				Description: "Journal voucher print undo failed, user org error.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}

		journalVoucher, err := core.JournalVoucherManager(service).GetByID(context, *journalVoucherID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-undo-error",
				Description: "Journal voucher print undo failed, not found.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher not found"})
		}

		if journalVoucher.OrganizationID != userOrg.OrganizationID || journalVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}

		if journalVoucher.PrintedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Journal voucher has not been printed yet"})
		}

		journalVoucher.PrintNumber = 0
		journalVoucher.PrintedDate = nil
		journalVoucher.PrintedByID = nil
		journalVoucher.UpdatedAt = time.Now().UTC()
		journalVoucher.UpdatedByID = userOrg.UserID

		if err := core.JournalVoucherManager(service).UpdateByID(context, journalVoucher.ID, journalVoucher); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-undo-error",
				Description: "Journal voucher print undo failed, db error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to undo print journal voucher: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "print-undo-success",
			Description: "Successfully undid print for journal voucher: " + journalVoucher.CashVoucherNumber,
			Module:      "JournalVoucher",
		})

		return ctx.JSON(http.StatusOK, core.JournalVoucherManager(service).ToModel(journalVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/:journal_voucher_id/approve",
		Method:       "PUT",
		Note:         "Approves a journal voucher by ID.",
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := helpers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Journal voucher approve failed, invalid ID.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Journal voucher approve failed, user org error.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}

		journalVoucher, err := core.JournalVoucherManager(service).GetByID(context, *journalVoucherID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Journal voucher approve failed, not found.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher not found"})
		}

		if journalVoucher.OrganizationID != userOrg.OrganizationID || journalVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}

		if journalVoucher.ApprovedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Journal voucher has already been approved"})
		}

		timeNow := userOrg.UserOrgTime()
		journalVoucher.ApprovedDate = helpers.Ptr(timeNow)
		journalVoucher.ApprovedByID = &userOrg.UserID
		journalVoucher.UpdatedAt = time.Now().UTC()
		journalVoucher.UpdatedByID = userOrg.UserID

		if err := core.JournalVoucherManager(service).UpdateByID(context, journalVoucher.ID, journalVoucher); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Journal voucher approve failed, db error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to approve journal voucher: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "approve-success",
			Description: "Successfully approved journal voucher: " + journalVoucher.CashVoucherNumber,
			Module:      "JournalVoucher",
		})

		return ctx.JSON(http.StatusOK, core.JournalVoucherManager(service).ToModel(journalVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/:journal_voucher_id/print-only",
		Method:       "POST",
		Note:         "Marks a journal voucher as printed without additional details by ID.",
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := helpers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-only-error",
				Description: "Journal voucher print-only failed, invalid ID.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-only-error",
				Description: "Journal voucher print-only failed, user org error.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}

		journalVoucher, err := core.JournalVoucherManager(service).GetByID(context, *journalVoucherID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-only-error",
				Description: "Journal voucher print-only failed, not found.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher not found"})
		}

		if journalVoucher.OrganizationID != userOrg.OrganizationID || journalVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}

		timeNow := userOrg.UserOrgTime()
		journalVoucher.PrintNumber++
		journalVoucher.PrintedDate = &timeNow
		journalVoucher.PrintedByID = &userOrg.UserID
		journalVoucher.UpdatedAt = time.Now().UTC()
		journalVoucher.UpdatedByID = userOrg.UserID

		if err := core.JournalVoucherManager(service).UpdateByID(context, journalVoucher.ID, journalVoucher); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-only-error",
				Description: "Journal voucher print-only failed, db error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to print journal voucher: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "print-only-success",
			Description: "Successfully printed journal voucher (print-only): " + journalVoucher.CashVoucherNumber,
			Module:      "JournalVoucher",
		})

		return ctx.JSON(http.StatusOK, core.JournalVoucherManager(service).ToModel(journalVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/:journal_voucher_id/approve-undo",
		Method:       "POST",
		Note:         "Reverts the approval status of a journal voucher by ID.",
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := helpers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-undo-error",
				Description: "Journal voucher approve undo failed, invalid ID.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-undo-error",
				Description: "Journal voucher approve undo failed, user org error.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}

		journalVoucher, err := core.JournalVoucherManager(service).GetByID(context, *journalVoucherID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-undo-error",
				Description: "Journal voucher approve undo failed, not found.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher not found"})
		}

		if journalVoucher.OrganizationID != userOrg.OrganizationID || journalVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}

		if journalVoucher.ApprovedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Journal voucher has not been approved yet"})
		}

		if journalVoucher.ReleasedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot unapprove a released journal voucher"})
		}

		journalVoucher.ApprovedDate = nil
		journalVoucher.ApprovedByID = nil
		journalVoucher.UpdatedAt = time.Now().UTC()
		journalVoucher.UpdatedByID = userOrg.UserID

		if err := core.JournalVoucherManager(service).UpdateByID(context, journalVoucher.ID, journalVoucher); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-undo-error",
				Description: "Journal voucher approve undo failed, db error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to undo approval for journal voucher: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "approve-undo-success",
			Description: "Successfully undid approval for journal voucher: " + journalVoucher.CashVoucherNumber,
			Module:      "JournalVoucher",
		})

		return ctx.JSON(http.StatusOK, core.JournalVoucherManager(service).ToModel(journalVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/:journal_voucher_id/release",
		Method:       "POST",
		Note:         "Releases a journal voucher by ID. RELEASED SHOULD NOT BE UNAPPROVED.",
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := helpers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "release-error",
				Description: "Journal voucher release failed, invalid ID.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "release-error",
				Description: "Journal voucher release failed, user org error.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}

		journalVoucher, err := core.JournalVoucherManager(service).GetByID(context, *journalVoucherID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "release-error",
				Description: "Journal voucher release failed, not found.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher not found"})
		}

		if journalVoucher.OrganizationID != userOrg.OrganizationID || journalVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}

		if journalVoucher.ApprovedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Journal voucher must be approved before release"})
		}

		if journalVoucher.ReleasedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Journal voucher has already been released"})
		}

		transactionBatch, err := core.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "batch-retrieval-failed",
				Description: "Unable to retrieve active transaction batch for user " + userOrg.UserID.String() + ": " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}
		timeNow := userOrg.UserOrgTime()
		journalVoucher.ReleasedDate = &timeNow
		journalVoucher.ReleasedByID = &userOrg.UserID
		journalVoucher.UpdatedAt = time.Now().UTC()
		journalVoucher.UpdatedByID = userOrg.UserID
		journalVoucher.TransactionBatchID = &transactionBatch.ID

		journalVoucherEntries, err := core.JournalVoucherEntryManager(service).Find(context, &core.JournalVoucherEntry{
			JournalVoucherID: journalVoucher.ID,
			OrganizationID:   userOrg.OrganizationID,
			BranchID:         *userOrg.BranchID,
		})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "journal-voucher-entries-retrieval-failed",
				Description: "Failed to retrieve journal voucher entries for release: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve journal voucher entries: " + err.Error()})
		}

		for _, entry := range journalVoucherEntries {
			transactionRequest := event.RecordTransactionRequest{
				Debit:  entry.Debit,
				Credit: entry.Credit,

				AccountID:          entry.AccountID,
				MemberProfileID:    entry.MemberProfileID,
				TransactionBatchID: transactionBatch.ID,

				ReferenceNumber:       journalVoucher.CashVoucherNumber,
				Description:           entry.Description,
				EntryDate:             &timeNow,
				BankReferenceNumber:   "",  // Not applicable for journal voucher entries
				BankID:                nil, // Not applicable for journal voucher entries
				ProofOfPaymentMediaID: nil, // Not applicable for journal voucher entries
				LoanTransactionID:     entry.LoanTransactionID,
			}

			if err := c.event.RecordTransaction(context, ctx, transactionRequest, core.GeneralLedgerSourceJournalVoucher); err != nil {

				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "journal-voucher-transaction-recording-failed",
					Description: "Failed to record journal voucher entry transaction in general ledger for voucher " + journalVoucher.CashVoucherNumber + ": " + err.Error(),
					Module:      "JournalVoucher",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Journal voucher release initiated but failed to record transaction: " + err.Error(),
				})
			}
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "journal-voucher-transactions-recorded",
			Description: "Successfully recorded all journal voucher entry transactions in general ledger for voucher: " + journalVoucher.CashVoucherNumber,
			Module:      "JournalVoucher",
		})

		if err := core.JournalVoucherManager(service).UpdateByID(context, journalVoucher.ID, journalVoucher); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "release-error",
				Description: "Journal voucher release failed, db error: " + err.Error(),
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to release journal voucher: " + err.Error()})
		}

		if err := c.event.TransactionBatchBalancing(context, journalVoucher.TransactionBatchID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "release-success",
			Description: "Successfully released journal voucher: " + journalVoucher.CashVoucherNumber,
			Module:      "JournalVoucher",
		})

		c.event.OrganizationAdminsNotification(ctx, event.NotificationEvent{
			Description:      fmt.Sprintf("Journal vouchers approved list has been accessed by %s", *userOrg.User.FirstName),
			Title:            "Journal Vouchers - Approved List Accessed",
			NotificationType: core.NotificationSystem,
		})

		return ctx.JSON(http.StatusOK, core.JournalVoucherManager(service).ToModel(journalVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/draft",
		Method:       "GET",
		Note:         "Fetches draft journal vouchers for the current user's organization and branch.",
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "release-error",
				Description: "Journal voucher release failed, user org error.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		journalVouchers, err := core.JournalVoucherDraft(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch draft journal vouchers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.JournalVoucherManager(service).ToModels(journalVouchers))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/printed",
		Method:       "GET",
		Note:         "Fetches printed journal vouchers for the current user's organization and branch.",
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "printed-error",
				Description: "Journal voucher printed fetch failed, user org error.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		journalVouchers, err := core.JournalVoucherPrinted(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch printed journal vouchers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.JournalVoucherManager(service).ToModels(journalVouchers))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/approved",
		Method:       "GET",
		Note:         "Fetches approved journal vouchers for the current user's organization and branch.",
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approved-error",
				Description: "Journal voucher approved fetch failed, user org error.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		journalVouchers, err := core.JournalVoucherApproved(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch approved journal vouchers: " + err.Error()})
		}

		c.event.OrganizationAdminsNotification(ctx, event.NotificationEvent{
			Description:      fmt.Sprintf("Journal vouchers approved list has been accessed by %s", *userOrg.User.FirstName),
			Title:            "Journal Vouchers - Approved List Accessed",
			NotificationType: core.NotificationInfo,
		})
		return ctx.JSON(http.StatusOK, core.JournalVoucherManager(service).ToModels(journalVouchers))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/released",
		Method:       "GET",
		Note:         "Fetches released journal vouchers for the current user's organization and branch.",
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "released-error",
				Description: "Journal voucher released fetch failed, user org error.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		journalVouchers, err := core.JournalVoucherReleased(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch released journal vouchers: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, core.JournalVoucherManager(service).ToModels(journalVouchers))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher/released/today",
		Method:       "GET",
		Note:         "Fetches journal vouchers released today for the current user's organization and branch.",
		ResponseType: core.JournalVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "released-today-error",
				Description: "Journal voucher released today fetch failed, user org error.",
				Module:      "JournalVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		journalVouchers, err := core.JournalVoucherReleasedCurrentDay(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch journal vouchers released today: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.JournalVoucherManager(service).ToModels(journalVouchers))
	})

}
