package reports

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/usecase"
	"github.com/labstack/echo/v4"
)

func GeneralLedgerController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-accounting-ledger/:member_accounting_ledger_id/total",
		Method:       "GET",
		ResponseType: types.MemberGeneralLedgerTotal{},
		Note:         "Returns the total amount for a specific member profile's general ledger entries for an account.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAccountingLedgerID, err := helpers.EngineUUIDParam(ctx, "member_accounting_ledger_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member accounting ledger id"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		memberAccountingLedger, err := core.MemberAccountingLedgerManager(service).GetByID(context, memberAccountingLedgerID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "member accounting ledger " + err.Error()})
		}
		entries, err := core.GeneralLedgerMemberAccountTotal(context, service,
			memberAccountingLedger.MemberProfileID,
			memberAccountingLedger.AccountID,
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
		return ctx.JSON(http.StatusOK, types.MemberGeneralLedgerTotal{
			Balance:     balance.Balance,
			TotalDebit:  balance.Debit,
			TotalCredit: balance.Credit,
		})
	})
	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-accounting-ledger/:member_accounting_ledger_id",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAccountingLedgerID, err := helpers.EngineUUIDParam(ctx, "member_accounting_ledger_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		memberAccountingLedgerm, err := core.MemberAccountingLedgerManager(service).GetByID(context, memberAccountingLedgerID)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Member accounting ledger not found " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		entries, err := core.GeneralLedgerManager(service).ArrFind(context, []query.ArrFilterSQL{
			{Field: "member_profile_id", Op: query.ModeEqual, Value: memberAccountingLedgerm.MemberProfileID},
			{Field: "organization_id", Op: query.ModeEqual, Value: userOrg.OrganizationID},
			{Field: "branch_id", Op: query.ModeEqual, Value: userOrg.BranchID},
			{Field: "account_id", Op: query.ModeEqual, Value: memberAccountingLedgerm.AccountID},
		}, []query.ArrFilterSortSQL{
			{Field: "entry_date", Order: query.SortOrderAsc},
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.GeneralLedgerManager(service).ToModels(entries))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/:general_ledger_id",
		Method:       "GET",
		ResponseType: types.GeneralLedger{},
		Note:         "Returns a specific general ledger entry by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generalLedgerID, err := helpers.EngineUUIDParam(ctx, "general_ledger_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		entry, err := core.GeneralLedgerManager(service).GetByIDRaw(context, *generalLedgerID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger entry not found"})
		}
		if entry.OrganizationID != userOrg.OrganizationID || entry.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view this general ledger entry"})
		}
		return ctx.JSON(http.StatusOK, entry)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/branch/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/branch/check-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrganization.OrganizationID,
			TypeOfPaymentType: types.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/branch/online-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrganization.OrganizationID,
			TypeOfPaymentType: types.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/branch/cash-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrganization.OrganizationID,
			TypeOfPaymentType: types.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/branch/payment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         types.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/branch/withdraw-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         types.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/branch/deposit-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         types.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/branch/journal-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         types.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/branch/adjustment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         types.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/branch/journal-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         types.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/branch/check-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         types.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/current/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case types.UserOrganizationTypeOwner, types.UserOrganizationTypeEmployee:
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case types.UserOrganizationTypeMember:
			member, err := core.MemberProfileManager(service).FindOne(context, &types.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/current/check-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case types.UserOrganizationTypeOwner, types.UserOrganizationTypeEmployee:
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				EmployeeUserID:    &userOrganization.UserID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: types.PaymentTypeCheck,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case types.UserOrganizationTypeMember:
			member, err := core.MemberProfileManager(service).FindOne(context, &types.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				MemberProfileID:   &member.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: types.PaymentTypeCheck,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/current/online-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case types.UserOrganizationTypeOwner, types.UserOrganizationTypeEmployee:
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				EmployeeUserID:    &userOrganization.UserID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: types.PaymentTypeOnline,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case types.UserOrganizationTypeMember:
			member, err := core.MemberProfileManager(service).FindOne(context, &types.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				MemberProfileID:   &member.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: types.PaymentTypeOnline,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/current/cash-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case types.UserOrganizationTypeOwner, types.UserOrganizationTypeEmployee:
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				EmployeeUserID:    &userOrganization.UserID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: types.PaymentTypeCash,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case types.UserOrganizationTypeMember:
			member, err := core.MemberProfileManager(service).FindOne(context, &types.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				MemberProfileID:   &member.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: types.PaymentTypeCash,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/current/payment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case types.UserOrganizationTypeOwner, types.UserOrganizationTypeEmployee:
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         types.GeneralLedgerSourcePayment,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case types.UserOrganizationTypeMember:
			member, err := core.MemberProfileManager(service).FindOne(context, &types.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          types.GeneralLedgerSourcePayment,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/current/withdraw-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
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
		case types.UserOrganizationTypeOwner, types.UserOrganizationTypeEmployee:
			entries, err := core.GeneralLedgerManager(service).NormalPagination(
				context,
				ctx,
				&types.GeneralLedger{
					EmployeeUserID: &userOrg.UserID,
					OrganizationID: userOrg.OrganizationID,
					BranchID:       *userOrg.BranchID,
					Source:         types.GeneralLedgerSourceWithdraw,
				},
			)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to retrieve ledger entries: " + err.Error(),
				})
			}

			return ctx.JSON(http.StatusOK, entries)

		case types.UserOrganizationTypeMember:
			member, err := core.MemberProfileManager(service).FindOne(
				context,
				&types.MemberProfile{
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
			entries, err := core.GeneralLedgerManager(service).NormalPagination(
				context,
				ctx,
				&types.GeneralLedger{
					MemberProfileID: &member.ID,
					OrganizationID:  userOrg.OrganizationID,
					BranchID:        *userOrg.BranchID,
					Source:          types.GeneralLedgerSourceWithdraw,
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/current/deposit-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case types.UserOrganizationTypeOwner, types.UserOrganizationTypeEmployee:
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         types.GeneralLedgerSourceDeposit,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case types.UserOrganizationTypeMember:
			member, err := core.MemberProfileManager(service).FindOne(context, &types.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          types.GeneralLedgerSourceDeposit,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/current/journal-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case types.UserOrganizationTypeOwner, types.UserOrganizationTypeEmployee:
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         types.GeneralLedgerSourceJournalVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case types.UserOrganizationTypeMember:
			member, err := core.MemberProfileManager(service).FindOne(context, &types.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          types.GeneralLedgerSourceJournalVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/current/adjustment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case types.UserOrganizationTypeOwner, types.UserOrganizationTypeEmployee:
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         types.GeneralLedgerSourceAdjustment,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case types.UserOrganizationTypeMember:
			member, err := core.MemberProfileManager(service).FindOne(context, &types.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          types.GeneralLedgerSourceAdjustment,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/current/journal-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case types.UserOrganizationTypeOwner, types.UserOrganizationTypeEmployee:
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         types.GeneralLedgerSourceJournalVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case types.UserOrganizationTypeMember:
			member, err := core.MemberProfileManager(service).FindOne(context, &types.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          types.GeneralLedgerSourceJournalVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/current/check-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case types.UserOrganizationTypeOwner, types.UserOrganizationTypeEmployee:
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         types.GeneralLedgerSourceCheckVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)

		case types.UserOrganizationTypeMember:
			member, err := core.MemberProfileManager(service).FindOne(context, &types.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          types.GeneralLedgerSourceCheckVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, entries)
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified employee (by user organization ID) with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/check-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			EmployeeUserID:    &userOrganization.UserID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: types.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/online-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			EmployeeUserID:    &userOrganization.UserID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: types.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/cash-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			EmployeeUserID:    &userOrganization.UserID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: types.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/payment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/journal-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/check-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		filters := []query.ArrFilterSQL{
			{Field: "member_profile_id", Op: query.ModeEqual, Value: memberProfileID},
			{Field: "organization_id", Op: query.ModeEqual, Value: userOrg.OrganizationID},
			{Field: "branch_id", Op: query.ModeEqual, Value: userOrg.BranchID},
		}
		cashOnHandID := userOrg.Branch.BranchSetting.CashOnHandAccountID
		if cashOnHandID != nil {
			filters = append(filters, query.ArrFilterSQL{
				Field: "account_id", Op: query.ModeNotEqual, Value: *userOrg.Branch.BranchSetting.CashOnHandAccountID,
			})
		}
		sorts := []query.ArrFilterSortSQL{
			{Field: "updated_at", Order: query.SortOrderDesc},
		}
		entries, err := core.GeneralLedgerManager(service).ArrPagination(context, ctx, filters, sorts)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/check-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := core.GeneralLedgerMemberProfileEntriesByPaymentType(context, service,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			types.PaymentTypeCheck,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/online-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := core.GeneralLedgerMemberProfileEntriesByPaymentType(context, service,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			types.PaymentTypeOnline,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/cash-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := core.GeneralLedgerMemberProfileEntriesByPaymentType(context, service,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			types.PaymentTypeCash,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/payment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := core.GeneralLedgerMemberProfileEntriesBySource(context, service,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			types.GeneralLedgerSourcePayment,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := core.GeneralLedgerMemberProfileEntriesBySource(context, service,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			types.GeneralLedgerSourceWithdraw,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := core.GeneralLedgerMemberProfileEntriesBySource(context, service,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			types.GeneralLedgerSourceDeposit,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/journal-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := core.GeneralLedgerMemberProfileEntriesBySource(context, service,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			types.GeneralLedgerSourceJournalVoucher,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := core.GeneralLedgerMemberProfileEntriesBySource(context, service,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			types.GeneralLedgerSourceAdjustment,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := core.GeneralLedgerMemberProfileEntriesBySource(context, service,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			types.GeneralLedgerSourceJournalVoucher,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/check-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := core.GeneralLedgerMemberProfileEntriesBySource(context, service,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
			types.GeneralLedgerSourceCheckVoucher,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			MemberProfileID:   memberProfileID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			AccountID:         accountID,
			TypeOfPaymentType: types.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/online-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			MemberProfileID:   memberProfileID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			AccountID:         accountID,
			TypeOfPaymentType: types.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/cash-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			MemberProfileID:   memberProfileID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			AccountID:         accountID,
			TypeOfPaymentType: types.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/payment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          types.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          types.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          types.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          types.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          types.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          types.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          types.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TypeOfPaymentType:  types.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/online-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TypeOfPaymentType:  types.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/cash-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TypeOfPaymentType:  types.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/payment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             types.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             types.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             types.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             types.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             types.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             types.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             types.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerExcludeCashonHand(context, service, *transactionID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, core.GeneralLedgerManager(service).ToModels(entries))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/check-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionID:     transactionID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: types.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/online-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionID:     transactionID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: types.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/cash-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionID:     transactionID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: types.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/payment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionID:  transactionID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			TransactionID:  transactionID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		source := types.GeneralLedgerSourceDeposit
		entries, err := core.GeneralLedgerExcludeCashonHandWithSource(context, service, *transactionID, userOrg.OrganizationID, *userOrg.BranchID, &source)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/journal-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		source := types.GeneralLedgerSourceJournalVoucher
		entries, err := core.GeneralLedgerExcludeCashonHandWithSource(context, service, *transactionID, userOrg.OrganizationID, *userOrg.BranchID, &source)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		source := types.GeneralLedgerSourceAdjustment
		entries, err := core.GeneralLedgerExcludeCashonHandWithSource(context, service, *transactionID, userOrg.OrganizationID, *userOrg.BranchID, &source)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		source := types.GeneralLedgerSourceJournalVoucher
		entries, err := core.GeneralLedgerExcludeCashonHandWithSource(
			context, service, *transactionID, userOrg.OrganizationID, *userOrg.BranchID, &source)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/check-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		source := types.GeneralLedgerSourceCheckVoucher
		entries, err := core.GeneralLedgerExcludeCashonHandWithSource(context, service, *transactionID, userOrg.OrganizationID, *userOrg.BranchID, &source)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/check-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			AccountID:         accountID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: types.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/online-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			AccountID:         accountID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: types.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/cash-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			AccountID:         accountID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: types.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/payment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/journal-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/check-voucher/search",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := core.GeneralLedgerManager(service).NormalPagination(context, ctx, &types.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         types.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, entries)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/general-ledger/loan-transaction/:loan_transaction_id",
		Method:       "GET",
		ResponseType: types.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified loan transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan general ledger entries"})
		}
		entries, err := core.GeneralLedgerByLoanTransaction(
			context, service,
			*loanTransactionID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.GeneralLedgerManager(service).ToModels(entries))
	})
}
