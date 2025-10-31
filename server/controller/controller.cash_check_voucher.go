package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CashCheckVoucherController registers routes for managing cash check vouchers.
func (c *Controller) cashCheckVoucherController() {
	req := c.provider.Service.Request

	// GET /cash-check-voucher: List all cash check vouchers for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher",
		Method:       "GET",
		Note:         "Returns all cash check vouchers for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCheckVouchers, err := c.modelcore.CashCheckVoucherCurrentbranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No cash check vouchers found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.CashCheckVoucherManager.Filtered(context, ctx, cashCheckVouchers))
	})

	// GET /cash-check-voucher/search: Paginated search of cash check vouchers for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/search",
		Method:       "GET",
		Note:         "Returns a paginated list of cash check vouchers for the current user's organization and branch.",
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCheckVouchers, err := c.modelcore.CashCheckVoucherCurrentbranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch cash check vouchers for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.CashCheckVoucherManager.Pagination(context, ctx, cashCheckVouchers))
	})

	// GET /api/v1/cash-check-voucher/draft
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/draft",
		Method:       "GET",
		Note:         "Fetches draft cash check vouchers for the current user's organization and branch.",
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "draft-error",
				Description: "Cash check voucher draft failed, user org error.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCheckVouchers, err := c.modelcore.CashCheckVoucherDraft(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch draft cash check vouchers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.CashCheckVoucherManager.ToModels(cashCheckVouchers))
	})

	// GET /api/v1/cash-check-voucher/printed
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/printed",
		Method:       "GET",
		Note:         "Fetches printed cash check vouchers for the current user's organization and branch.",
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "printed-error",
				Description: "Cash check voucher printed fetch failed, user org error.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCheckVouchers, err := c.modelcore.CashCheckVoucherPrinted(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch printed cash check vouchers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.CashCheckVoucherManager.ToModels(cashCheckVouchers))
	})

	// GET /api/v1/cash-check-voucher/approved
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/approved",
		Method:       "GET",
		Note:         "Fetches approved cash check vouchers for the current user's organization and branch.",
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "approved-error",
				Description: "Cash check voucher approved fetch failed, user org error.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCheckVouchers, err := c.modelcore.CashCheckVoucherApproved(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch approved cash check vouchers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.CashCheckVoucherManager.ToModels(cashCheckVouchers))
	})

	// GET /api/v1/cash-check-voucher/released
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/released",
		Method:       "GET",
		Note:         "Fetches released cash check vouchers for the current user's organization and branch.",
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "released-error",
				Description: "Cash check voucher released fetch failed, user org error.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCheckVouchers, err := c.modelcore.CashCheckVoucherReleased(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch released cash check vouchers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.CashCheckVoucherManager.ToModels(cashCheckVouchers))
	})

	// GET /cash-check-voucher/:cash_check_voucher_id: Get specific cash check voucher by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id",
		Method:       "GET",
		Note:         "Returns a single cash check voucher by its ID.",
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := handlers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}
		cashCheckVoucher, err := c.modelcore.CashCheckVoucherManager.GetByIDRaw(context, *cashCheckVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}
		return ctx.JSON(http.StatusOK, cashCheckVoucher)
	})

	// POST /cash-check-voucher: Create a new cash check voucher. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher",
		Method:       "POST",
		Note:         "Creates a new cash check voucher for the current user's organization and branch.",
		RequestType:  modelcore.CashCheckVoucherRequest{},
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		request, err := c.modelcore.CashCheckVoucherManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher creation failed (/cash-check-voucher), validation error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher creation failed (/cash-check-voucher), user org error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher creation failed (/cash-check-voucher), user not assigned to branch.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		// Start transaction
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher creation failed (/cash-check-voucher), transaction error: " + tx.Error.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}

		// Calculate totals from entries
		totalDebit, totalCredit := 0.0, 0.0
		if request.CashCheckVoucherEntries != nil {
			for _, entry := range request.CashCheckVoucherEntries {
				totalDebit += entry.Debit
				totalCredit += entry.Credit
			}
		}

		// Validate balance (optional - some vouchers might not require balanced entries)
		if totalDebit != totalCredit && totalDebit > 0 && totalCredit > 0 {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher creation failed (/cash-check-voucher), unbalanced entries.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Cash check voucher is not balanced: debit %.2f != credit %.2f", totalDebit, totalCredit)})
		}

		cashCheckVoucher := &modelcore.CashCheckVoucher{
			PayTo:                         request.PayTo,
			Status:                        request.Status,
			Description:                   request.Description,
			CashVoucherNumber:             request.CashVoucherNumber,
			TotalDebit:                    totalDebit,
			TotalCredit:                   totalCredit,
			PrintCount:                    request.PrintCount,
			PrintedDate:                   request.PrintedDate,
			ApprovedDate:                  request.ApprovedDate,
			ReleasedDate:                  request.ReleasedDate,
			EmployeeUserID:                request.EmployeeUserID,
			TransactionBatchID:            request.TransactionBatchID,
			ApprovedBySignatureMediaID:    request.ApprovedBySignatureMediaID,
			ApprovedByName:                request.ApprovedByName,
			ApprovedByPosition:            request.ApprovedByPosition,
			PreparedBySignatureMediaID:    request.PreparedBySignatureMediaID,
			PreparedByName:                request.PreparedByName,
			PreparedByPosition:            request.PreparedByPosition,
			CertifiedBySignatureMediaID:   request.CertifiedBySignatureMediaID,
			CertifiedByName:               request.CertifiedByName,
			CertifiedByPosition:           request.CertifiedByPosition,
			VerifiedBySignatureMediaID:    request.VerifiedBySignatureMediaID,
			VerifiedByName:                request.VerifiedByName,
			VerifiedByPosition:            request.VerifiedByPosition,
			CheckBySignatureMediaID:       request.CheckBySignatureMediaID,
			CheckByName:                   request.CheckByName,
			CheckByPosition:               request.CheckByPosition,
			AcknowledgeBySignatureMediaID: request.AcknowledgeBySignatureMediaID,
			AcknowledgeByName:             request.AcknowledgeByName,
			AcknowledgeByPosition:         request.AcknowledgeByPosition,
			NotedBySignatureMediaID:       request.NotedBySignatureMediaID,
			NotedByName:                   request.NotedByName,
			NotedByPosition:               request.NotedByPosition,
			PostedBySignatureMediaID:      request.PostedBySignatureMediaID,
			PostedByName:                  request.PostedByName,
			PostedByPosition:              request.PostedByPosition,
			PaidBySignatureMediaID:        request.PaidBySignatureMediaID,
			PaidByName:                    request.PaidByName,
			PaidByPosition:                request.PaidByPosition,
			CheckEntryAmount:              request.CheckEntryAmount,
			CheckEntryCheckNumber:         request.CheckEntryCheckNumber,
			CheckEntryCheckDate:           request.CheckEntryCheckDate,
			CheckEntryAccountID:           request.CheckEntryAccountID,
			CreatedAt:                     time.Now().UTC(),
			CreatedByID:                   user.UserID,
			UpdatedAt:                     time.Now().UTC(),
			UpdatedByID:                   user.UserID,
			BranchID:                      *user.BranchID,
			OrganizationID:                user.OrganizationID,
			Name:                          request.Name,
			CurrencyID:                    request.CurrencyID,
		}

		// Save cash check voucher first
		if err := c.modelcore.CashCheckVoucherManager.CreateWithTx(context, tx, cashCheckVoucher); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher creation failed (/cash-check-voucher), save error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cash check voucher: " + err.Error()})
		}

		transactionBatch, err := c.modelcore.TransactionBatchCurrent(context, user.UserID, user.OrganizationID, *user.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "End transaction batch failed: retrieve error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}
		if request.CashCheckVoucherEntries != nil {
			for _, entryReq := range request.CashCheckVoucherEntries {
				entry := &modelcore.CashCheckVoucherEntry{
					AccountID:              entryReq.AccountID,
					EmployeeUserID:         &user.UserID,
					TransactionBatchID:     &transactionBatch.ID,
					CashCheckVoucherID:     cashCheckVoucher.ID,
					Debit:                  entryReq.Debit,
					Credit:                 entryReq.Credit,
					Description:            entryReq.Description,
					CreatedAt:              time.Now().UTC(),
					CreatedByID:            user.UserID,
					UpdatedAt:              time.Now().UTC(),
					UpdatedByID:            user.UserID,
					BranchID:               *user.BranchID,
					OrganizationID:         user.OrganizationID,
					CashCheckVoucherNumber: entryReq.CashCheckVoucherNumber,
					MemberProfileID:        entryReq.MemberProfileID,
				}

				if err := c.modelcore.CashCheckVoucherEntryManager.CreateWithTx(context, tx, entry); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "create-error",
						Description: "Cash check voucher creation failed (/cash-check-voucher), entry save error: " + err.Error(),
						Module:      "CashCheckVoucher",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cash check voucher entry: " + err.Error()})
				}
			}
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher creation failed (/cash-check-voucher), commit error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		newCashCheckVoucher, err := c.modelcore.CashCheckVoucherManager.GetByIDRaw(context, cashCheckVoucher.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated cash check voucher: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created cash check voucher (/cash-check-voucher): " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})
		return ctx.JSON(http.StatusCreated, newCashCheckVoucher)
	})

	// PUT /cash-check-voucher/:cash_check_voucher_id: Update cash check voucher by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id",
		Method:       "PUT",
		Note:         "Updates an existing cash check voucher by its ID.",
		RequestType:  modelcore.CashCheckVoucherRequest{},
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := handlers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}

		request, err := c.modelcore.CashCheckVoucherManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), validation error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher data: " + err.Error()})
		}

		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), user org error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		cashCheckVoucher, err := c.modelcore.CashCheckVoucherManager.GetByID(context, *cashCheckVoucherID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), voucher not found: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}

		// Calculate totals from entries
		totalDebit, totalCredit := 0.0, 0.0
		if request.CashCheckVoucherEntries != nil {
			for _, entry := range request.CashCheckVoucherEntries {
				totalDebit += entry.Debit
				totalCredit += entry.Credit
			}
		}

		// Validate balance (optional)
		if totalDebit != totalCredit && totalDebit > 0 && totalCredit > 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), unbalanced entries.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Cash check voucher is not balanced: debit %.2f != credit %.2f", totalDebit, totalCredit)})
		}

		// Start transaction
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), transaction error: " + tx.Error.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}

		// Update cash check voucher fields
		cashCheckVoucher.PayTo = request.PayTo
		cashCheckVoucher.Status = request.Status
		cashCheckVoucher.Description = request.Description
		cashCheckVoucher.CashVoucherNumber = request.CashVoucherNumber
		cashCheckVoucher.TotalDebit = totalDebit
		cashCheckVoucher.TotalCredit = totalCredit
		cashCheckVoucher.PrintCount = request.PrintCount
		cashCheckVoucher.PrintedDate = request.PrintedDate
		cashCheckVoucher.ApprovedDate = request.ApprovedDate
		cashCheckVoucher.ReleasedDate = request.ReleasedDate
		cashCheckVoucher.EmployeeUserID = request.EmployeeUserID
		cashCheckVoucher.TransactionBatchID = request.TransactionBatchID
		cashCheckVoucher.ApprovedBySignatureMediaID = request.ApprovedBySignatureMediaID
		cashCheckVoucher.ApprovedByName = request.ApprovedByName
		cashCheckVoucher.ApprovedByPosition = request.ApprovedByPosition
		cashCheckVoucher.PreparedBySignatureMediaID = request.PreparedBySignatureMediaID
		cashCheckVoucher.PreparedByName = request.PreparedByName
		cashCheckVoucher.PreparedByPosition = request.PreparedByPosition
		cashCheckVoucher.CertifiedBySignatureMediaID = request.CertifiedBySignatureMediaID
		cashCheckVoucher.CertifiedByName = request.CertifiedByName
		cashCheckVoucher.CertifiedByPosition = request.CertifiedByPosition
		cashCheckVoucher.VerifiedBySignatureMediaID = request.VerifiedBySignatureMediaID
		cashCheckVoucher.VerifiedByName = request.VerifiedByName
		cashCheckVoucher.VerifiedByPosition = request.VerifiedByPosition
		cashCheckVoucher.CheckBySignatureMediaID = request.CheckBySignatureMediaID
		cashCheckVoucher.CheckByName = request.CheckByName
		cashCheckVoucher.CheckByPosition = request.CheckByPosition
		cashCheckVoucher.AcknowledgeBySignatureMediaID = request.AcknowledgeBySignatureMediaID
		cashCheckVoucher.AcknowledgeByName = request.AcknowledgeByName
		cashCheckVoucher.AcknowledgeByPosition = request.AcknowledgeByPosition
		cashCheckVoucher.NotedBySignatureMediaID = request.NotedBySignatureMediaID
		cashCheckVoucher.NotedByName = request.NotedByName
		cashCheckVoucher.NotedByPosition = request.NotedByPosition
		cashCheckVoucher.PostedBySignatureMediaID = request.PostedBySignatureMediaID
		cashCheckVoucher.PostedByName = request.PostedByName
		cashCheckVoucher.PostedByPosition = request.PostedByPosition
		cashCheckVoucher.PaidBySignatureMediaID = request.PaidBySignatureMediaID
		cashCheckVoucher.PaidByName = request.PaidByName
		cashCheckVoucher.PaidByPosition = request.PaidByPosition
		cashCheckVoucher.CheckEntryAmount = request.CheckEntryAmount
		cashCheckVoucher.CheckEntryCheckNumber = request.CheckEntryCheckNumber
		cashCheckVoucher.CheckEntryCheckDate = request.CheckEntryCheckDate
		cashCheckVoucher.CheckEntryAccountID = request.CheckEntryAccountID
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = user.UserID
		cashCheckVoucher.Name = request.Name

		// Handle deleted entries
		if request.CashCheckVoucherEntriesDeleted != nil {
			for _, entryID := range request.CashCheckVoucherEntriesDeleted {
				entry, err := c.modelcore.CashCheckVoucherEntryManager.GetByID(context, entryID)
				if err != nil {
					tx.Rollback()
					continue
				}
				if entry.CashCheckVoucherID != cashCheckVoucher.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete entry that doesn't belong to this cash check voucher"})
				}
				entry.DeletedByID = &user.UserID
				if err := c.modelcore.CashCheckVoucherEntryManager.DeleteWithTx(context, tx, entry); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), delete entry error: " + err.Error(),
						Module:      "CashCheckVoucher",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash check voucher entry: " + err.Error()})
				}
			}
		}
		transactionBatch, err := c.modelcore.TransactionBatchCurrent(context, user.UserID, user.OrganizationID, *user.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "End transaction batch failed: retrieve error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}
		// Handle cash check voucher entries (create new or update existing)
		if request.CashCheckVoucherEntries != nil {
			for _, entryReq := range request.CashCheckVoucherEntries {
				if entryReq.ID != nil {
					// Update existing entry
					entry, err := c.modelcore.CashCheckVoucherEntryManager.GetByID(context, *entryReq.ID)
					if err != nil {
						tx.Rollback()
						c.event.Footstep(context, ctx, event.FootstepEvent{
							Activity:    "update-error",
							Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), get entry error: " + err.Error(),
							Module:      "CashCheckVoucher",
						})
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get cash check voucher entry: " + err.Error()})
					}
					entry.AccountID = entryReq.AccountID
					entry.EmployeeUserID = &user.UserID
					entry.TransactionBatchID = &transactionBatch.ID
					entry.Debit = entryReq.Debit
					entry.Credit = entryReq.Credit
					entry.Description = entryReq.Description
					entry.UpdatedAt = time.Now().UTC()
					entry.UpdatedByID = user.UserID
					entry.MemberProfileID = entryReq.MemberProfileID
					entry.CashCheckVoucherNumber = entryReq.CashCheckVoucherNumber
					if err := c.modelcore.CashCheckVoucherEntryManager.UpdateFieldsWithTx(context, tx, entry.ID, entry); err != nil {
						tx.Rollback()
						c.event.Footstep(context, ctx, event.FootstepEvent{
							Activity:    "update-error",
							Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), update entry error: " + err.Error(),
							Module:      "CashCheckVoucher",
						})
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update cash check voucher entry: " + err.Error()})
					}
				} else {
					entry := &modelcore.CashCheckVoucherEntry{
						AccountID:              entryReq.AccountID,
						EmployeeUserID:         &user.UserID,
						TransactionBatchID:     &transactionBatch.ID,
						CashCheckVoucherID:     cashCheckVoucher.ID,
						Debit:                  entryReq.Debit,
						Credit:                 entryReq.Credit,
						Description:            entryReq.Description,
						CreatedAt:              time.Now().UTC(),
						CreatedByID:            user.UserID,
						UpdatedAt:              time.Now().UTC(),
						UpdatedByID:            user.UserID,
						BranchID:               *user.BranchID,
						OrganizationID:         user.OrganizationID,
						CashCheckVoucherNumber: entryReq.CashCheckVoucherNumber,
						MemberProfileID:        entryReq.MemberProfileID,
					}

					if err := c.modelcore.CashCheckVoucherEntryManager.CreateWithTx(context, tx, entry); err != nil {
						tx.Rollback()
						c.event.Footstep(context, ctx, event.FootstepEvent{
							Activity:    "update-error",
							Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), entry save error: " + err.Error(),
							Module:      "CashCheckVoucher",
						})
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cash check voucher entry: " + err.Error()})
					}
				}
			}
		}

		// Save updated cash check voucher
		if err := c.modelcore.CashCheckVoucherManager.UpdateFieldsWithTx(context, tx, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), save error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update cash check voucher: " + err.Error()})
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), commit error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		newCashCheckVoucher, err := c.modelcore.CashCheckVoucherManager.GetByIDRaw(context, cashCheckVoucher.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated cash check voucher: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated cash check voucher (/cash-check-voucher/:cash_check_voucher_id): " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})
		return ctx.JSON(http.StatusOK, newCashCheckVoucher)
	})

	// DELETE /cash-check-voucher/:cash_check_voucher_id: Delete a cash check voucher by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/cash-check-voucher/:cash_check_voucher_id",
		Method: "DELETE",
		Note:   "Deletes the specified cash check voucher by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := handlers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}
		cashCheckVoucher, err := c.modelcore.CashCheckVoucherManager.GetByID(context, *cashCheckVoucherID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash check voucher deletion failed (/cash-check-voucher/:cash_check_voucher_id), voucher not found: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}
		if err := c.modelcore.CashCheckVoucherManager.DeleteByID(context, *cashCheckVoucherID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash check voucher deletion failed (/cash-check-voucher/:cash_check_voucher_id), delete error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash check voucher: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted cash check voucher (/cash-check-voucher/:cash_check_voucher_id): " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /cash-check-voucher/bulk-delete: Bulk delete cash check vouchers by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/cash-check-voucher/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple cash check vouchers by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: modelcore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided"})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}

		voucherNumbers := ""
		for i, rawID := range reqBody.IDs {
			voucherID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid UUID format: " + rawID})
			}

			cashCheckVoucher, err := c.modelcore.CashCheckVoucherManager.GetByID(context, voucherID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Cash check voucher bulk deletion failed (/cash-check-voucher/bulk-delete), voucher not found: " + err.Error(),
					Module:      "CashCheckVoucher",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found: " + rawID})
			}

			if err := tx.Delete(&modelcore.CashCheckVoucher{}, voucherID).Error; err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Cash check voucher bulk deletion failed (/cash-check-voucher/bulk-delete), delete error: " + err.Error(),
					Module:      "CashCheckVoucher",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash check voucher: " + err.Error()})
			}

			if i > 0 {
				voucherNumbers += ", "
			}
			voucherNumbers += cashCheckVoucher.CashVoucherNumber
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Cash check voucher bulk deletion failed (/cash-check-voucher/bulk-delete), commit error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted cash check vouchers (/cash-check-voucher/bulk-delete): " + voucherNumbers,
			Module:      "CashCheckVoucher",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// PUT /api/v1/cash-check-voucher/:cash_check_voucher_id/print
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id/print",
		Method:       "PUT",
		Note:         "Marks a cash check voucher as printed by ID and updates print count.",
		RequestType:  modelcore.CashCheckVoucherPrintRequest{},
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := handlers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to print cash check voucher"})
		}

		var req modelcore.CashCheckVoucherPrintRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "print-error",
				Description: "Cash check voucher print failed, invalid request body.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		cashCheckVoucher, err := c.modelcore.CashCheckVoucherManager.GetByID(context, *cashCheckVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}
		if cashCheckVoucher.OrganizationID != userOrg.OrganizationID || cashCheckVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this cash check voucher"})
		}

		// Update print details
		cashCheckVoucher.CashVoucherNumber = req.CashVoucherNumber
		cashCheckVoucher.PrintCount = cashCheckVoucher.PrintCount + 1
		cashCheckVoucher.PrintedDate = handlers.Ptr(time.Now().UTC())
		cashCheckVoucher.Status = modelcore.CashCheckVoucherStatusPrinted
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = userOrg.UserID
		cashCheckVoucher.PrintedByID = &userOrg.UserID

		if err := c.modelcore.CashCheckVoucherManager.UpdateFields(context, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update cash check voucher print status: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "print-success",
			Description: "Successfully printed cash check voucher: " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.CashCheckVoucherManager.ToModel(cashCheckVoucher))
	})

	// PUT /api/v1/cash-check-voucher/:cash_check_voucher_id/approve
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id/approve",
		Method:       "PUT",
		Note:         "Approves a cash check voucher by ID.",
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := handlers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to approve cash check voucher"})
		}

		cashCheckVoucher, err := c.modelcore.CashCheckVoucherManager.GetByID(context, *cashCheckVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}

		if cashCheckVoucher.OrganizationID != userOrg.OrganizationID || cashCheckVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this cash check voucher"})
		}

		if cashCheckVoucher.ApprovedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash check voucher is already approved"})
		}

		// Update approval details
		cashCheckVoucher.ApprovedDate = handlers.Ptr(time.Now().UTC())
		cashCheckVoucher.Status = modelcore.CashCheckVoucherStatusApproved
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = userOrg.UserID
		cashCheckVoucher.ApprovedByID = &userOrg.UserID

		if err := c.modelcore.CashCheckVoucherManager.UpdateFields(context, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to approve cash check voucher: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "approve-success",
			Description: "Successfully approved cash check voucher: " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.CashCheckVoucherManager.ToModel(cashCheckVoucher))
	})

	// POST /api/v1/cash-check-voucher/:cash_check_voucher_id/release
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id/release",
		Method:       "POST",
		Note:         "Releases a cash check voucher by ID. RELEASED SHOULD NOT BE UNAPPROVED.",
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := handlers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to release cash check voucher"})
		}

		cashCheckVoucher, err := c.modelcore.CashCheckVoucherManager.GetByID(context, *cashCheckVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}

		if cashCheckVoucher.OrganizationID != userOrg.OrganizationID || cashCheckVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this cash check voucher"})
		}

		if cashCheckVoucher.ApprovedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash check voucher must be approved before it can be released"})
		}

		if cashCheckVoucher.ReleasedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash check voucher is already released"})
		}

		// Update release details
		cashCheckVoucher.ReleasedDate = handlers.Ptr(time.Now().UTC())
		cashCheckVoucher.Status = modelcore.CashCheckVoucherStatusReleased
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = userOrg.UserID
		cashCheckVoucher.ReleasedByID = &userOrg.UserID

		if err := c.modelcore.CashCheckVoucherManager.UpdateFields(context, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to release cash check voucher: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "release-success",
			Description: "Successfully released cash check voucher: " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.CashCheckVoucherManager.ToModel(cashCheckVoucher))
	})

	// PUT /api/v1/cash-check-voucher/:cash_check_voucher_id/print-undo
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id/print-undo",
		Method:       "PUT",
		Note:         "Reverts the print status of a cash check voucher by ID.",
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := handlers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to undo print for cash check voucher"})
		}

		cashCheckVoucher, err := c.modelcore.CashCheckVoucherManager.GetByID(context, *cashCheckVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}

		if cashCheckVoucher.OrganizationID != userOrg.OrganizationID || cashCheckVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this cash check voucher"})
		}

		if cashCheckVoucher.PrintedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash check voucher has not been printed yet"})
		}

		// Revert print details
		cashCheckVoucher.PrintCount = 0
		cashCheckVoucher.PrintedDate = nil
		cashCheckVoucher.Status = modelcore.CashCheckVoucherStatusPending
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = userOrg.UserID
		cashCheckVoucher.PrintedByID = nil

		if err := c.modelcore.CashCheckVoucherManager.UpdateFields(context, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to undo print for cash check voucher: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "print-undo-success",
			Description: "Successfully undid print for cash check voucher: " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.CashCheckVoucherManager.ToModel(cashCheckVoucher))
	})

	// POST /api/v1/cash-check-voucher/:cash_check_voucher_id/print-only
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id/print-only",
		Method:       "POST",
		Note:         "Marks a cash check voucher as printed without additional details by ID.",
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := handlers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to print cash check voucher"})
		}

		cashCheckVoucher, err := c.modelcore.CashCheckVoucherManager.GetByID(context, *cashCheckVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}

		if cashCheckVoucher.OrganizationID != userOrg.OrganizationID || cashCheckVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this cash check voucher"})
		}

		// Update print details without changing voucher number
		cashCheckVoucher.PrintCount = cashCheckVoucher.PrintCount + 1
		cashCheckVoucher.PrintedDate = handlers.Ptr(time.Now().UTC())
		cashCheckVoucher.Status = modelcore.CashCheckVoucherStatusPrinted
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = userOrg.UserID
		cashCheckVoucher.PrintedByID = &userOrg.UserID

		if err := c.modelcore.CashCheckVoucherManager.UpdateFields(context, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to print cash check voucher: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "print-only-success",
			Description: "Successfully printed cash check voucher (print-only): " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.CashCheckVoucherManager.ToModel(cashCheckVoucher))
	})

	// POST /api/v1/cash-check-voucher/:cash_check_voucher_id/approve-undo
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id/approve-undo",
		Method:       "POST",
		Note:         "Reverts the approval status of a cash check voucher by ID.",
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := handlers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to undo approval for cash check voucher"})
		}

		cashCheckVoucher, err := c.modelcore.CashCheckVoucherManager.GetByID(context, *cashCheckVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}

		if cashCheckVoucher.OrganizationID != userOrg.OrganizationID || cashCheckVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this cash check voucher"})
		}

		if cashCheckVoucher.ApprovedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash check voucher has not been approved yet"})
		}

		if cashCheckVoucher.ReleasedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot unapprove a released cash check voucher"})
		}

		// Revert approval details
		cashCheckVoucher.ApprovedDate = nil
		cashCheckVoucher.Status = modelcore.CashCheckVoucherStatusPrinted // Or pending if not printed
		if cashCheckVoucher.PrintedDate == nil {
			cashCheckVoucher.Status = modelcore.CashCheckVoucherStatusPending
		}
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = userOrg.UserID
		cashCheckVoucher.ApprovedBy = nil

		if err := c.modelcore.CashCheckVoucherManager.UpdateFields(context, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to undo approval for cash check voucher: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "approve-undo-success",
			Description: "Successfully undid approval for cash check voucher: " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.CashCheckVoucherManager.ToModel(cashCheckVoucher))
	})

	// GET api/v1/cash-check-voucher/released/today
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher/released/today",
		Method:       "GET",
		Note:         "Retrieves all cash check vouchers released today.",
		ResponseType: modelcore.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		vouchers, err := c.modelcore.CashCheckVoucherReleasedCurrentDay(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve today's released cash check vouchers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.CashCheckVoucherManager.ToModels(vouchers))
	})
}
