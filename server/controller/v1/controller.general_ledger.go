package v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) generalLedgerController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/total",
		Method:       "GET",
		ResponseType: core.MemberGeneralLedgerTotal{},
		Note:         "Returns the total amount for a specific member profile's general ledger entries for an account.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger totals"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.core.GeneralLedgerMemberAccountTotal(context,
			*memberProfileID,
			*accountID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		balance, err := usecase.CalculateBalance(usecase.Balance{
			GeneralLedgers: entries,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to compute total balance: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberGeneralLedgerTotal{
			Balance:     balance.Balance,
			TotalDebit:  balance.Debit,
			TotalCredit: balance.Credit,
		})
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/:general_ledger_id",
		Method:       "GET",
		ResponseType: core.GeneralLedger{},
		Note:         "Returns a specific general ledger entry by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generalLedgerID, err := handlers.EngineUUIDParam(ctx, "general_ledger_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		entry, err := c.core.GeneralLedgerManager().GetByIDRaw(context, *generalLedgerID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger entry not found"})
		}
		if entry.OrganizationID != userOrg.OrganizationID || entry.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view this general ledger entry"})
		}
		return ctx.JSON(http.StatusOK, entry)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/check-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrganization.OrganizationID,
			TypeOfPaymentType: core.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/online-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrganization.OrganizationID,
			TypeOfPaymentType: core.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/cash-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrganization.OrganizationID,
			TypeOfPaymentType: core.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/payment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         core.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/withdraw-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         core.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/deposit-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         core.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/journal-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         core.GeneralLedgerSourceJournal,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/adjustment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         core.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/journal-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         core.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/check-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         core.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case core.UserOrganizationTypeOwner, core.UserOrganizationTypeEmployee:
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case core.UserOrganizationTypeMember:
			member, err := c.core.MemberProfileManager().FindOne(context, &core.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/check-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case core.UserOrganizationTypeOwner, core.UserOrganizationTypeEmployee:
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				EmployeeUserID:    &userOrganization.UserID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: core.PaymentTypeCheck,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case core.UserOrganizationTypeMember:
			member, err := c.core.MemberProfileManager().FindOne(context, &core.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				MemberProfileID:   &member.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: core.PaymentTypeCheck,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/online-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case core.UserOrganizationTypeOwner, core.UserOrganizationTypeEmployee:
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				EmployeeUserID:    &userOrganization.UserID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: core.PaymentTypeOnline,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case core.UserOrganizationTypeMember:
			member, err := c.core.MemberProfileManager().FindOne(context, &core.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				MemberProfileID:   &member.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: core.PaymentTypeOnline,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/cash-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case core.UserOrganizationTypeOwner, core.UserOrganizationTypeEmployee:
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				EmployeeUserID:    &userOrganization.UserID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: core.PaymentTypeCash,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case core.UserOrganizationTypeMember:
			member, err := c.core.MemberProfileManager().FindOne(context, &core.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				MemberProfileID:   &member.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: core.PaymentTypeCash,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/payment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case core.UserOrganizationTypeOwner, core.UserOrganizationTypeEmployee:
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         core.GeneralLedgerSourcePayment,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case core.UserOrganizationTypeMember:
			member, err := c.core.MemberProfileManager().FindOne(context, &core.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          core.GeneralLedgerSourcePayment,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/withdraw-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "User authentication failed or organization not found",
			})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "User branch is not assigned",
			})
		}
		switch userOrg.UserType {
		case core.UserOrganizationTypeOwner, core.UserOrganizationTypeEmployee:
			entries, err := c.core.GeneralLedgerManager().NormalPagination(
				context,
				ctx,
				&core.GeneralLedger{
					EmployeeUserID: &userOrg.UserID,
					OrganizationID: userOrg.OrganizationID,
					BranchID:       *userOrg.BranchID,
					Source:         core.GeneralLedgerSourceWithdraw,
				},
			)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to retrieve ledger entries: " + err.Error(),
				})
			}

			return ctx.JSON(http.StatusOK, entries)

		case core.UserOrganizationTypeMember:
			member, err := c.core.MemberProfileManager().FindOne(
				context,
				&core.MemberProfile{
					UserID:         &userOrg.UserID,
					BranchID:       *userOrg.BranchID,
					OrganizationID: userOrg.OrganizationID,
				},
			)
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{
					"error": "Member profile not found",
				})
			}
			entries, err := c.core.GeneralLedgerManager().NormalPagination(
				context,
				ctx,
				&core.GeneralLedger{
					MemberProfileID: &member.ID,
					OrganizationID:  userOrg.OrganizationID,
					BranchID:        *userOrg.BranchID,
					Source:          core.GeneralLedgerSourceWithdraw,
				},
			)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to retrieve ledger entries: " + err.Error(),
				})
			}
			return ctx.JSON(http.StatusOK, entries)

		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{
				"error": "User is not authorized to view employee general ledger entries",
			})
		}
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/deposit-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case core.UserOrganizationTypeOwner, core.UserOrganizationTypeEmployee:
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         core.GeneralLedgerSourceDeposit,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case core.UserOrganizationTypeMember:
			member, err := c.core.MemberProfileManager().FindOne(context, &core.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          core.GeneralLedgerSourceDeposit,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/journal-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case core.UserOrganizationTypeOwner, core.UserOrganizationTypeEmployee:
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         core.GeneralLedgerSourceJournal,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case core.UserOrganizationTypeMember:
			member, err := c.core.MemberProfileManager().FindOne(context, &core.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          core.GeneralLedgerSourceJournal,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/adjustment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case core.UserOrganizationTypeOwner, core.UserOrganizationTypeEmployee:
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         core.GeneralLedgerSourceAdjustment,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case core.UserOrganizationTypeMember:
			member, err := c.core.MemberProfileManager().FindOne(context, &core.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          core.GeneralLedgerSourceAdjustment,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/journal-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case core.UserOrganizationTypeOwner, core.UserOrganizationTypeEmployee:
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         core.GeneralLedgerSourceJournalVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case core.UserOrganizationTypeMember:
			member, err := c.core.MemberProfileManager().FindOne(context, &core.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          core.GeneralLedgerSourceJournalVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/check-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case core.UserOrganizationTypeOwner, core.UserOrganizationTypeEmployee:
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         core.GeneralLedgerSourceCheckVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case core.UserOrganizationTypeMember:
			member, err := c.core.MemberProfileManager().FindOne(context, &core.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          core.GeneralLedgerSourceCheckVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified employee (by user organization ID) with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/check-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			EmployeeUserID:    &userOrganization.UserID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: core.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/online-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			EmployeeUserID:    &userOrganization.UserID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: core.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/cash-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			EmployeeUserID:    &userOrganization.UserID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: core.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/payment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/journal-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourceJournal,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/check-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		filters := []registry.FilterSQL{
			{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
			{Field: "organization_id", Op: query.ModeEqual, Value: userOrg.OrganizationID},
			{Field: "branch_id", Op: query.ModeEqual, Value: userOrg.BranchID},
		}
		cashOnHandID := userOrg.Branch.BranchSetting.CashOnHandAccountID
		if cashOnHandID != nil {
			filters = append(filters, registry.FilterSQL{
				Field: "account_id", Op: query.ModeNotEqual, Value: *userOrg.Branch.BranchSetting.CashOnHandAccountID,
			})
		}
		sorts := []query.ArrFilterSortSQL{
			{Field: "updated_at", Order: query.SortOrderDesc},
		}
		entries, err := c.core.GeneralLedgerManager().ArrPagination(context, ctx, filters, sorts)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/check-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.core.GeneralLedgerMemberProfileEntriesByPaymentType(context,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			core.PaymentTypeCheck,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/online-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.core.GeneralLedgerMemberProfileEntriesByPaymentType(context,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			core.PaymentTypeOnline,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/cash-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.core.GeneralLedgerMemberProfileEntriesByPaymentType(context,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			core.PaymentTypeCash,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/payment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.core.GeneralLedgerMemberProfileEntriesBySource(context,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			core.GeneralLedgerSourcePayment,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.core.GeneralLedgerMemberProfileEntriesBySource(context,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			core.GeneralLedgerSourceWithdraw,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.core.GeneralLedgerMemberProfileEntriesBySource(context,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			core.GeneralLedgerSourceDeposit,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/journal-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.core.GeneralLedgerMemberProfileEntriesBySource(context,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			core.GeneralLedgerSourceJournal,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.core.GeneralLedgerMemberProfileEntriesBySource(context,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			core.GeneralLedgerSourceAdjustment,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.core.GeneralLedgerMemberProfileEntriesBySource(context,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			core.GeneralLedgerSourceJournalVoucher,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/check-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.core.GeneralLedgerMemberProfileEntriesBySource(context,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			core.GeneralLedgerSourceCheckVoucher,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			MemberProfileID:   memberProfileID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			AccountID:         accountID,
			TypeOfPaymentType: core.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/online-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			MemberProfileID:   memberProfileID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			AccountID:         accountID,
			TypeOfPaymentType: core.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/cash-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			MemberProfileID:   memberProfileID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			AccountID:         accountID,
			TypeOfPaymentType: core.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/payment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          core.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          core.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          core.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          core.GeneralLedgerSourceJournal,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          core.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          core.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          core.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TypeOfPaymentType:  core.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/online-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TypeOfPaymentType:  core.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/cash-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TypeOfPaymentType:  core.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/payment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             core.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             core.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             core.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             core.GeneralLedgerSourceJournal,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             core.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             core.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             core.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerExcludeCashonHand(context, *transactionID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerManager().ToModels(entries))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/check-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionID:     transactionID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: core.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/online-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionID:     transactionID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: core.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/cash-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionID:     transactionID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: core.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/payment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionID:  transactionID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			TransactionID:  transactionID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		source := core.GeneralLedgerSourceDeposit
		entries, err := c.core.GeneralLedgerExcludeCashonHandWithSource(context, *transactionID, userOrg.OrganizationID, *userOrg.BranchID, &source)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/journal-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		source := core.GeneralLedgerSourceJournal
		entries, err := c.core.GeneralLedgerExcludeCashonHandWithSource(context, *transactionID, userOrg.OrganizationID, *userOrg.BranchID, &source)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		source := core.GeneralLedgerSourceAdjustment
		entries, err := c.core.GeneralLedgerExcludeCashonHandWithSource(context, *transactionID, userOrg.OrganizationID, *userOrg.BranchID, &source)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		source := core.GeneralLedgerSourceJournalVoucher
		entries, err := c.core.GeneralLedgerExcludeCashonHandWithSource(context, *transactionID, userOrg.OrganizationID, *userOrg.BranchID, &source)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/check-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		source := core.GeneralLedgerSourceCheckVoucher
		entries, err := c.core.GeneralLedgerExcludeCashonHandWithSource(context, *transactionID, userOrg.OrganizationID, *userOrg.BranchID, &source)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/check-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			AccountID:         accountID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: core.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/online-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			AccountID:         accountID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: core.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/cash-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			AccountID:         accountID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: core.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/payment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/journal-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourceJournal,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/check-voucher/search",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.core.GeneralLedgerManager().NormalPagination(context, ctx, &core.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         core.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/loan-transaction/:loan_transaction_id",
		Method:       "GET",
		ResponseType: core.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified loan transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan general ledger entries"})
		}
		entries, err := c.core.GeneralLedgerByLoanTransaction(
			context,
			*loanTransactionID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerManager().ToModels(entries))
	})
}
