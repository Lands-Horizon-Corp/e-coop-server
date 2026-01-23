package voucher

import (
	"net/http"
	"sort"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/usecase"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func CashCheckVoucherController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher",
		Method:       "GET",
		Note:         "Returns all cash check vouchers for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCheckVouchers, err := core.CashCheckVoucherCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No cash check vouchers found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.CashCheckVoucherManager(service).ToModels(cashCheckVouchers))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/search",
		Method:       "GET",
		Note:         "Returns a paginated list of cash check vouchers for the current user's organization and branch.",
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCheckVouchers, err := core.CashCheckVoucherManager(service).NormalPagination(context, ctx, &types.CashCheckVoucher{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch cash check vouchers for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, cashCheckVouchers)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/draft",
		Method:       "GET",
		Note:         "Fetches draft cash check vouchers for the current user's organization and branch.",
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "draft-error",
				Description: "Cash check voucher draft failed, user org error.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCheckVouchers, err := core.CashCheckVoucherDraft(context, service, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch draft cash check vouchers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.CashCheckVoucherManager(service).ToModels(cashCheckVouchers))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/printed",
		Method:       "GET",
		Note:         "Fetches printed cash check vouchers for the current user's organization and branch.",
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "printed-error",
				Description: "Cash check voucher printed fetch failed, user org error.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCheckVouchers, err := core.CashCheckVoucherPrinted(context, service, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch printed cash check vouchers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.CashCheckVoucherManager(service).ToModels(cashCheckVouchers))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/approved",
		Method:       "GET",
		Note:         "Fetches approved cash check vouchers for the current user's organization and branch.",
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approved-error",
				Description: "Cash check voucher approved fetch failed, user org error.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCheckVouchers, err := core.CashCheckVoucherApproved(context, service, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch approved cash check vouchers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.CashCheckVoucherManager(service).ToModels(cashCheckVouchers))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/released",
		Method:       "GET",
		Note:         "Fetches released cash check vouchers for the current user's organization and branch.",
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "released-error",
				Description: "Cash check voucher released fetch failed, user org error.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cashCheckVouchers, err := core.CashCheckVoucherReleased(context, service, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch released cash check vouchers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.CashCheckVoucherManager(service).ToModels(cashCheckVouchers))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id",
		Method:       "GET",
		Note:         "Returns a single cash check voucher by its ID.",
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := helpers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}
		cashCheckVoucher, err := core.CashCheckVoucherManager(service).GetByIDRaw(context, *cashCheckVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}
		return ctx.JSON(http.StatusOK, cashCheckVoucher)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher",
		Method:       "POST",
		Note:         "Creates a new cash check voucher for the current user's organization and branch.",
		RequestType:  types.CashCheckVoucherRequest{},
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		request, err := core.CashCheckVoucherManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher creation failed (/cash-check-voucher), validation error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher creation failed (/cash-check-voucher), user org error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher creation failed (/cash-check-voucher), user not assigned to branch.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		tx, endTx := service.Database.StartTransaction(context)

		balance, err := usecase.CalculateBalance(usecase.Balance{
			CashCheckVoucherEntriesRequest: request.CashCheckVoucherEntries,
		})

		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher creation failed (/cash-check-voucher), balance calculation error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to calculate balance: " + endTx(err).Error()})
		}

		cashCheckVoucher := &types.CashCheckVoucher{
			PayTo:                         request.PayTo,
			Status:                        request.Status,
			Description:                   request.Description,
			CashVoucherNumber:             request.CashVoucherNumber,
			TotalDebit:                    balance.Debit,
			TotalCredit:                   balance.Credit,
			PrintCount:                    request.PrintCount,
			EmployeeUserID:                &userOrg.UserID,
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

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			Name:           request.Name,
			CurrencyID:     request.CurrencyID,
		}

		if err := core.CashCheckVoucherManager(service).CreateWithTx(context, tx, cashCheckVoucher); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher creation failed (/cash-check-voucher), save error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cash check voucher: " + endTx(err).Error()})
		}

		if request.CashCheckVoucherEntries != nil {
			for _, entryReq := range request.CashCheckVoucherEntries {
				entry := &types.CashCheckVoucherEntry{
					AccountID:          entryReq.AccountID,
					EmployeeUserID:     &userOrg.UserID,
					CashCheckVoucherID: cashCheckVoucher.ID,
					Debit:              entryReq.Debit,
					Credit:             entryReq.Credit,
					Description:        entryReq.Description,
					CreatedAt:          time.Now().UTC(),
					CreatedByID:        userOrg.UserID,
					UpdatedAt:          time.Now().UTC(),
					UpdatedByID:        userOrg.UserID,
					BranchID:           *userOrg.BranchID,
					OrganizationID:     userOrg.OrganizationID,
					LoanTransactionID:  entryReq.LoanTransactionID,
					MemberProfileID:    entryReq.MemberProfileID,
				}

				if err := core.CashCheckVoucherEntryManager(service).CreateWithTx(context, tx, entry); err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "create-error",
						Description: "Cash check voucher creation failed (/cash-check-voucher), entry save error: " + err.Error(),
						Module:      "CashCheckVoucher",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cash check voucher entry: " + endTx(err).Error()})
				}
			}
		}

		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher creation failed (/cash-check-voucher), commit error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		newCashCheckVoucher, err := core.CashCheckVoucherManager(service).GetByIDRaw(context, cashCheckVoucher.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated cash check voucher: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created cash check voucher (/cash-check-voucher): " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})
		return ctx.JSON(http.StatusCreated, newCashCheckVoucher)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id",
		Method:       "PUT",
		Note:         "Updates an existing cash check voucher by its ID.",
		RequestType:  types.CashCheckVoucherRequest{},
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := helpers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}

		request, err := core.CashCheckVoucherManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), validation error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher data: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), user org error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		cashCheckVoucher, err := core.CashCheckVoucherManager(service).GetByID(context, *cashCheckVoucherID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), voucher not found: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}

		balance, err := usecase.CalculateStrictBalance(usecase.Balance{
			CashCheckVoucherEntriesRequest: request.CashCheckVoucherEntries,
		})

		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), balance calculation error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to calculate balance: " + err.Error()})
		}

		tx, endTx := service.Database.StartTransaction(context)

		cashCheckVoucher.PayTo = request.PayTo
		cashCheckVoucher.Status = request.Status
		cashCheckVoucher.Description = request.Description
		cashCheckVoucher.CashVoucherNumber = request.CashVoucherNumber
		cashCheckVoucher.TotalDebit = balance.Debit
		cashCheckVoucher.TotalCredit = balance.Credit
		cashCheckVoucher.PrintCount = request.PrintCount
		cashCheckVoucher.EmployeeUserID = &userOrg.UserID
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
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = userOrg.UserID
		cashCheckVoucher.Name = request.Name

		if request.CashCheckVoucherEntriesDeleted != nil {
			for _, entryID := range request.CashCheckVoucherEntriesDeleted {
				entry, err := core.CashCheckVoucherEntryManager(service).GetByID(context, entryID)
				if err != nil {
					continue
				}
				if entry.CashCheckVoucherID != cashCheckVoucher.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete entry that doesn't belong to this cash check voucher: " + endTx(eris.New("invalid entry")).Error()})
				}
				entry.DeletedByID = &userOrg.UserID
				if err := core.CashCheckVoucherEntryManager(service).DeleteWithTx(context, tx, entry.ID); err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), delete entry error: " + err.Error(),
						Module:      "CashCheckVoucher",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash check voucher entry: " + endTx(err).Error()})
				}
			}
		}

		if request.CashCheckVoucherEntries != nil {
			for _, entryReq := range request.CashCheckVoucherEntries {
				if entryReq.ID != nil {
					entry, err := core.CashCheckVoucherEntryManager(service).GetByID(context, *entryReq.ID)
					if err != nil {
						event.Footstep(ctx, service, event.FootstepEvent{
							Activity:    "update-error",
							Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), get entry error: " + err.Error(),
							Module:      "CashCheckVoucher",
						})
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get cash check voucher entry: " + endTx(err).Error()})
					}
					entry.AccountID = entryReq.AccountID
					entry.EmployeeUserID = &userOrg.UserID
					entry.Debit = entryReq.Debit
					entry.Credit = entryReq.Credit
					entry.Description = entryReq.Description
					entry.UpdatedAt = time.Now().UTC()
					entry.UpdatedByID = userOrg.UserID
					entry.MemberProfileID = entryReq.MemberProfileID
					entry.LoanTransactionID = entryReq.LoanTransactionID
					if err := core.CashCheckVoucherEntryManager(service).UpdateByIDWithTx(context, tx, entry.ID, entry); err != nil {
						event.Footstep(ctx, service, event.FootstepEvent{
							Activity:    "update-error",
							Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), update entry error: " + err.Error(),
							Module:      "CashCheckVoucher",
						})
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update cash check voucher entry: " + endTx(err).Error()})
					}
				} else {
					entry := &types.CashCheckVoucherEntry{
						AccountID:          entryReq.AccountID,
						EmployeeUserID:     &userOrg.UserID,
						CashCheckVoucherID: cashCheckVoucher.ID,
						Debit:              entryReq.Debit,
						Credit:             entryReq.Credit,
						Description:        entryReq.Description,
						CreatedAt:          time.Now().UTC(),
						CreatedByID:        userOrg.UserID,
						UpdatedAt:          time.Now().UTC(),
						UpdatedByID:        userOrg.UserID,
						BranchID:           *userOrg.BranchID,
						OrganizationID:     userOrg.OrganizationID,
						LoanTransactionID:  entryReq.LoanTransactionID,
						MemberProfileID:    entryReq.MemberProfileID,
					}

					if err := core.CashCheckVoucherEntryManager(service).CreateWithTx(context, tx, entry); err != nil {
						event.Footstep(ctx, service, event.FootstepEvent{
							Activity:    "update-error",
							Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), entry save error: " + err.Error(),
							Module:      "CashCheckVoucher",
						})
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cash check voucher entry: " + endTx(err).Error()})
					}
				}
			}
		}

		if err := core.CashCheckVoucherManager(service).UpdateByIDWithTx(context, tx, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), save error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update cash check voucher: " + endTx(err).Error()})
		}

		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher update failed (/cash-check-voucher/:cash_check_voucher_id), commit error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		newCashCheckVoucher, err := core.CashCheckVoucherManager(service).GetByIDRaw(context, cashCheckVoucher.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated cash check voucher: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated cash check voucher (/cash-check-voucher/:cash_check_voucher_id): " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})
		return ctx.JSON(http.StatusOK, newCashCheckVoucher)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/cash-check-voucher/:cash_check_voucher_id",
		Method: "DELETE",
		Note:   "Deletes the specified cash check voucher by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := helpers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}
		cashCheckVoucher, err := core.CashCheckVoucherManager(service).GetByID(context, *cashCheckVoucherID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash check voucher deletion failed (/cash-check-voucher/:cash_check_voucher_id), voucher not found: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}
		if err := core.CashCheckVoucherManager(service).Delete(context, *cashCheckVoucherID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash check voucher deletion failed (/cash-check-voucher/:cash_check_voucher_id), delete error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash check voucher: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted cash check voucher (/cash-check-voucher/:cash_check_voucher_id): " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/cash-check-voucher/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple cash check vouchers by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Cash check voucher bulk deletion failed (/cash-check-voucher/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Cash check voucher bulk deletion failed (/cash-check-voucher/bulk-delete) | no IDs provided",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.CashCheckVoucherManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Cash check voucher bulk deletion failed (/cash-check-voucher/bulk-delete) | error: " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete cash check vouchers: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted cash check vouchers (/cash-check-voucher/bulk-delete)",
			Module:      "CashCheckVoucher",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id/print",
		Method:       "PUT",
		Note:         "Marks a cash check voucher as printed by ID and updates print count.",
		RequestType:  types.CashCheckVoucherPrintRequest{},
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := helpers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to print cash check voucher"})
		}

		var req types.CashCheckVoucherPrintRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "print-error",
				Description: "Cash check voucher print failed, invalid request body.",
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if err := service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		tx, endTx := service.Database.StartTransaction(context)
		if req.ORAutoGenerated {
			if err := event.IncrementOfficialReceipt(context, service, tx, req.CashVoucherNumber, types.GeneralLedgerSourceCheckVoucher, userOrg); err != nil {
				return endTx(err)
			}
		}
		cashCheckVoucher, err := core.CashCheckVoucherManager(service).GetByID(context, *cashCheckVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found: " + endTx(err).Error()})
		}
		if cashCheckVoucher.OrganizationID != userOrg.OrganizationID || cashCheckVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this cash check voucher: " + endTx(err).Error()})
		}

		timeNow := userOrg.TimeMachine()

		cashCheckVoucher.CashVoucherNumber = req.CashVoucherNumber
		cashCheckVoucher.EntryDate = &timeNow
		cashCheckVoucher.PrintCount++
		cashCheckVoucher.PrintedDate = &timeNow
		cashCheckVoucher.Status = types.CashCheckVoucherStatusPrinted
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = userOrg.UserID
		cashCheckVoucher.PrintedByID = &userOrg.UserID

		if err := core.CashCheckVoucherManager(service).UpdateByIDWithTx(context, tx, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update cash check voucher print status: " + endTx(err).Error()})
		}
		if endTx(nil) != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Transaction error" + err.Error()})

		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "print-success",
			Description: "Successfully printed cash check voucher: " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})

		return ctx.JSON(http.StatusOK, core.CashCheckVoucherManager(service).ToModel(cashCheckVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id/approve",
		Method:       "PUT",
		Note:         "Approves a cash check voucher by ID.",
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := helpers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to approve cash check voucher"})
		}

		cashCheckVoucher, err := core.CashCheckVoucherManager(service).GetByID(context, *cashCheckVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}

		if cashCheckVoucher.OrganizationID != userOrg.OrganizationID || cashCheckVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this cash check voucher"})
		}

		if cashCheckVoucher.ApprovedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash check voucher is already approved"})
		}

		timeNow := userOrg.TimeMachine()
		cashCheckVoucher.ApprovedDate = &timeNow
		cashCheckVoucher.Status = types.CashCheckVoucherStatusApproved
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = userOrg.UserID
		cashCheckVoucher.ApprovedByID = &userOrg.UserID

		if err := core.CashCheckVoucherManager(service).UpdateByID(context, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to approve cash check voucher: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "approve-success",
			Description: "Successfully approved cash check voucher: " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "approve-success",
			Description: "Successfully approved cash check voucher: " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})

		return ctx.JSON(http.StatusOK, core.CashCheckVoucherManager(service).ToModel(cashCheckVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id/release",
		Method:       "POST",
		Note:         "Releases a cash check voucher by ID. RELEASED SHOULD NOT BE UNAPPROVED.",
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := helpers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to release cash check voucher"})
		}

		cashCheckVoucher, err := core.CashCheckVoucherManager(service).GetByID(context, *cashCheckVoucherID)
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

		transactionBatch, err := core.TransactionBatchCurrent(context, service, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "batch-retrieval-failed",
				Description: "Unable to retrieve active transaction batch for user " + userOrg.UserID.String() + ": " + err.Error(),
				Module:      "CashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}
		timeNow := userOrg.TimeMachine()
		cashCheckVoucher.ReleasedDate = &timeNow
		cashCheckVoucher.Status = types.CashCheckVoucherStatusReleased
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = userOrg.UserID
		cashCheckVoucher.ReleasedByID = &userOrg.UserID
		cashCheckVoucher.TransactionBatchID = &transactionBatch.ID

		cashCheckVoucherEntries, err := core.CashCheckVoucherEntryManager(service).Find(context, &types.CashCheckVoucherEntry{
			CashCheckVoucherID: cashCheckVoucher.ID,
			BranchID:           *userOrg.BranchID,
			OrganizationID:     userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash check voucher entries: " + err.Error()})
		}

		if transactionBatch == nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batch. TB is Nil"})

		}
		for _, entry := range cashCheckVoucherEntries {
			transactionRequest := event.RecordTransactionRequest{
				Debit:  entry.Debit,
				Credit: entry.Credit,

				AccountID:       entry.AccountID,
				MemberProfileID: entry.MemberProfileID,

				ReferenceNumber:       cashCheckVoucher.CashVoucherNumber,
				Description:           entry.Description,
				EntryDate:             &timeNow,
				BankReferenceNumber:   "",
				BankID:                nil,
				ProofOfPaymentMediaID: nil,
				TransactionBatchID:    transactionBatch.ID,
				LoanTransactionID:     entry.LoanTransactionID,
			}

			if err := event.RecordTransaction(context, service, transactionRequest, types.GeneralLedgerSourceCheckVoucher, userOrg); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "cash-check-voucher-transaction-recording-failed",
					Description: "Failed to record cash check voucher entry transaction in general ledger for voucher " + cashCheckVoucher.CashVoucherNumber + ": " + err.Error(),
					Module:      "CashCheckVoucher",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Cash check voucher release initiated but failed to record transaction: " + err.Error(),
				})
			}
		}

		if err := core.CashCheckVoucherManager(service).UpdateByID(context, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to release cash check voucher: " + err.Error()})
		}

		if err := event.TransactionBatchBalancing(context, service, cashCheckVoucher.TransactionBatchID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch: " + err.Error()})
		}
		newCashCheckVoucher, err := core.CashCheckVoucherManager(service).GetByID(context, cashCheckVoucher.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch cash/check voucher: " + err.Error()})
		}
		sort.Slice(newCashCheckVoucher.CashCheckVoucherEntries, func(i, j int) bool {
			return newCashCheckVoucher.CashCheckVoucherEntries[i].CreatedAt.After(newCashCheckVoucher.CashCheckVoucherEntries[j].CreatedAt)
		})
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated cash/check voucher (/cash-check-voucher/:cash_check_voucher_id): " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})

		return ctx.JSON(http.StatusOK, core.CashCheckVoucherManager(service).ToModel(newCashCheckVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id/print-undo",
		Method:       "PUT",
		Note:         "Reverts the print status of a cash check voucher by ID.",
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := helpers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to undo print for cash check voucher"})
		}

		cashCheckVoucher, err := core.CashCheckVoucherManager(service).GetByID(context, *cashCheckVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}

		if cashCheckVoucher.OrganizationID != userOrg.OrganizationID || cashCheckVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this cash check voucher"})
		}

		if cashCheckVoucher.PrintedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash check voucher has not been printed yet"})
		}

		cashCheckVoucher.PrintCount = 0
		cashCheckVoucher.PrintedDate = nil
		cashCheckVoucher.Status = types.CashCheckVoucherStatusPending
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = userOrg.UserID
		cashCheckVoucher.PrintedByID = nil

		if err := core.CashCheckVoucherManager(service).UpdateByID(context, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to undo print for cash check voucher: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "print-undo-success",
			Description: "Successfully undid print for cash check voucher: " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})

		return ctx.JSON(http.StatusOK, core.CashCheckVoucherManager(service).ToModel(cashCheckVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id/print-only",
		Method:       "POST",
		Note:         "Marks a cash check voucher as printed without additional details by ID.",
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := helpers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to print cash check voucher"})
		}

		cashCheckVoucher, err := core.CashCheckVoucherManager(service).GetByID(context, *cashCheckVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher not found"})
		}

		if cashCheckVoucher.OrganizationID != userOrg.OrganizationID || cashCheckVoucher.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this cash check voucher"})
		}

		cashCheckVoucher.PrintCount++
		cashCheckVoucher.PrintedDate = helpers.Ptr(time.Now().UTC())
		cashCheckVoucher.Status = types.CashCheckVoucherStatusPrinted
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = userOrg.UserID
		cashCheckVoucher.PrintedByID = &userOrg.UserID

		if err := core.CashCheckVoucherManager(service).UpdateByID(context, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to print cash check voucher: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "print-only-success",
			Description: "Successfully printed cash check voucher (print-only): " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})

		return ctx.JSON(http.StatusOK, core.CashCheckVoucherManager(service).ToModel(cashCheckVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/:cash_check_voucher_id/approve-undo",
		Method:       "POST",
		Note:         "Reverts the approval status of a cash check voucher by ID.",
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := helpers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to undo approval for cash check voucher"})
		}

		cashCheckVoucher, err := core.CashCheckVoucherManager(service).GetByID(context, *cashCheckVoucherID)
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

		cashCheckVoucher.ApprovedDate = nil
		cashCheckVoucher.Status = types.CashCheckVoucherStatusPrinted // Or pending if not printed
		if cashCheckVoucher.PrintedDate == nil {
			cashCheckVoucher.Status = types.CashCheckVoucherStatusPending
		}
		cashCheckVoucher.UpdatedAt = time.Now().UTC()
		cashCheckVoucher.UpdatedByID = userOrg.UserID
		cashCheckVoucher.ApprovedBy = nil

		if err := core.CashCheckVoucherManager(service).UpdateByID(context, cashCheckVoucher.ID, cashCheckVoucher); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to undo approval for cash check voucher: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "approve-undo-success",
			Description: "Successfully undid approval for cash check voucher: " + cashCheckVoucher.CashVoucherNumber,
			Module:      "CashCheckVoucher",
		})

		return ctx.JSON(http.StatusOK, core.CashCheckVoucherManager(service).ToModel(cashCheckVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher/released/today",
		Method:       "GET",
		Note:         "Retrieves all cash check vouchers released today.",
		ResponseType: types.CashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		vouchers, err := core.CashCheckVoucherReleasedCurrentDay(context, service, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve today's released cash check vouchers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.CashCheckVoucherManager(service).ToModels(vouchers))
	})
}
