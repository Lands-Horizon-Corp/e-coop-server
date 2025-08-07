package controller_v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/model"
)

// GeneralLedgerController manages endpoints for general ledger accounts, definitions, and member ledgers.
func (c *Controller) GeneralLedgerController() {
	req := c.provider.Service.Request

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/total",
		Method:       "GET",
		ResponseType: model.MemberGeneralLedgerTotal{},
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger totals"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		var totalAmount float64
		var debit float64
		var credit float64
		for _, entry := range entries {
			switch entry.Account.Type {
			case model.AccountTypeDeposit:
				totalAmount += entry.Debit - entry.Credit
			case model.AccountTypeLoan:
				totalAmount += entry.Credit - entry.Debit
			case model.AccountTypeARLedger:
				totalAmount += entry.Debit - entry.Credit
			case model.AccountTypeARAging:
				totalAmount += entry.Debit - entry.Credit
			case model.AccountTypeFines:
				totalAmount += entry.Credit - entry.Debit
			case model.AccountTypeInterest:
				totalAmount += entry.Credit - entry.Debit
			case model.AccountTypeSVFLedger:
				totalAmount += entry.Debit - entry.Credit
			case model.AccountTypeWOff:
				totalAmount += entry.Debit - entry.Credit
			case model.AccountTypeAPLedger:
				totalAmount += entry.Credit - entry.Debit
			case model.AccountTypeOther:
				totalAmount += 0
			case model.AccountTypeTimeDeposit:
				totalAmount += entry.Credit - entry.Debit
			}
			debit += entry.Debit
			credit += entry.Credit
		}
		result := model.MemberGeneralLedgerTotal{
			Balance:     totalAmount,
			TotalDebit:  debit,
			TotalCredit: credit,
		}
		return ctx.JSON(http.StatusOK, result)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/:general_ledger_id",
		Method:       "GET",
		ResponseType: model.GeneralLedger{},
		Note:         "Returns a specific general ledger entry by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generalLedgerID, err := handlers.EngineUUIDParam(ctx, "general_ledger_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		entry, err := c.model.GeneralLedgerManager.GetByIDRaw(context, *generalLedgerID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger entry not found"})
		}
		if entry.OrganizationID != userOrg.OrganizationID || entry.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view this general ledger entry"})
		}
		return ctx.JSON(http.StatusOK, entry)
	})
	// BRANCH GENERAL LEDGER ROUTES

	// GET z
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/branch/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrganization.OrganizationID,
			TypeOfPaymentType: model.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/branch/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrganization.OrganizationID,
			TypeOfPaymentType: model.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/branch/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrganization.OrganizationID,
			TypeOfPaymentType: model.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/branch/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         model.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/branch/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         model.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/branch/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         model.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/branch/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         model.GeneralLedgerSourceJournal,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/branch/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         model.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/branch/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         model.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/branch/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view branch general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrganization.OrganizationID,
			Source:         model.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})
	// ME
	// GET /api/v1/general-ledger/current/search
	// GET /api/v1/general-ledger/current/check-entry/search
	// GET /api/v1/general-ledger/current/online-entry/search
	// GET /api/v1/general-ledger/current/cash-entry/search
	// GET /api/v1/general-ledger/current/payment-entry/search
	// GET /api/v1/general-ledger/current/withdraw-entry/search
	// GET /api/v1/general-ledger/current/deposit-entry/search
	// GET /api/v1/general-ledger/current/journal-entry/search
	// GET /api/v1/general-ledger/current/adjustment-entry/search
	// GET /api/v1/general-ledger/current/journal-voucher/search
	// GET /api/v1/general-ledger/current/check-voucher/search
	// ME GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/current/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case "owner", "employee":
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))

		case "member":
			member, err := c.model.MemberProfileManager.FindOne(context, &model.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	// GET /api/v1/general-ledger/current/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case "owner", "employee":
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				EmployeeUserID:    &userOrganization.UserID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: model.PaymentTypeCheck,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))

		case "member":
			member, err := c.model.MemberProfileManager.FindOne(context, &model.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				MemberProfileID:   &member.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: model.PaymentTypeCheck,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	// GET /api/v1/general-ledger/current/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case "owner", "employee":
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				EmployeeUserID:    &userOrganization.UserID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: model.PaymentTypeOnline,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))

		case "member":
			member, err := c.model.MemberProfileManager.FindOne(context, &model.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				MemberProfileID:   &member.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: model.PaymentTypeOnline,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	// GET /api/v1/general-ledger/current/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case "owner", "employee":
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				EmployeeUserID:    &userOrganization.UserID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: model.PaymentTypeCash,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))

		case "member":
			member, err := c.model.MemberProfileManager.FindOne(context, &model.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				MemberProfileID:   &member.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				TypeOfPaymentType: model.PaymentTypeCash,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	// GET /api/v1/general-ledger/current/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case "owner", "employee":
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         model.GeneralLedgerSourcePayment,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))

		case "member":
			member, err := c.model.MemberProfileManager.FindOne(context, &model.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          model.GeneralLedgerSourcePayment,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	// GET /api/v1/general-ledger/current/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case "owner", "employee":
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         model.GeneralLedgerSourceWithdraw,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))

		case "member":
			member, err := c.model.MemberProfileManager.FindOne(context, &model.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          model.GeneralLedgerSourceWithdraw,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	// GET /api/v1/general-ledger/current/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case "owner", "employee":
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         model.GeneralLedgerSourceDeposit,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))

		case "member":
			member, err := c.model.MemberProfileManager.FindOne(context, &model.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          model.GeneralLedgerSourceDeposit,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	// GET /api/v1/general-ledger/current/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case "owner", "employee":
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         model.GeneralLedgerSourceJournal,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))

		case "member":
			member, err := c.model.MemberProfileManager.FindOne(context, &model.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          model.GeneralLedgerSourceJournal,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	// GET /api/v1/general-ledger/current/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case "owner", "employee":
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         model.GeneralLedgerSourceAdjustment,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))

		case "member":
			member, err := c.model.MemberProfileManager.FindOne(context, &model.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          model.GeneralLedgerSourceAdjustment,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	// GET /api/v1/general-ledger/current/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case "owner", "employee":
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         model.GeneralLedgerSourceJournalVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))

		case "member":
			member, err := c.model.MemberProfileManager.FindOne(context, &model.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          model.GeneralLedgerSourceJournalVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	// GET /api/v1/general-ledger/current/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, userOrg.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		switch userOrg.UserType {
		case "owner", "employee":
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				EmployeeUserID: &userOrganization.UserID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Source:         model.GeneralLedgerSourceCheckVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))

		case "member":
			member, err := c.model.MemberProfileManager.FindOne(context, &model.MemberProfile{
				UserID:         &userOrganization.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrganization.OrganizationID,
			})
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
			}
			entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
				MemberProfileID: &member.ID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          model.GeneralLedgerSourceCheckVoucher,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})

		}
	})

	// EMPLOYEE
	// GET /api/v1/general-ledger/employee/:user_organization_id/search
	// GET /api/v1/general-ledger/employee/:user_organization_id/check-entry/search
	// GET /api/v1/general-ledger/employee/:user_organization_id/online-entry/search
	// GET /api/v1/general-ledger/employee/:user_organization_id/cash-entry/search
	// GET /api/v1/general-ledger/employee/:user_organization_id/payment-entry/search
	// GET /api/v1/general-ledger/employee/:user_organization_id/withdraw-entry/search
	// GET /api/v1/general-ledger/employee/:user_organization_id/deposit-entry/search
	// GET /api/v1/general-ledger/employee/:user_organization_id/journal-entry/search
	// GET /api/v1/general-ledger/employee/:user_organization_id/adjustment-entry/search
	// GET /api/v1/general-ledger/employee/:user_organization_id/journal-voucher
	// GET /api/v1/general-ledger/employee/:user_organization_id/check-voucher
	// EMPLOYEE GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/employee/:user_organization_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified employee (by user organization ID) with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/employee/:user_organization_id/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			EmployeeUserID:    &userOrganization.UserID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: model.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/employee/:user_organization_id/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			EmployeeUserID:    &userOrganization.UserID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: model.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/employee/:user_organization_id/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			EmployeeUserID:    &userOrganization.UserID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: model.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/employee/:user_organization_id/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/employee/:user_organization_id/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/employee/:user_organization_id/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/employee/:user_organization_id/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceJournal,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/employee/:user_organization_id/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/employee/:user_organization_id/journal-voucher
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/employee/:user_organization_id/check-voucher
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/employee/:user_organization_id/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view employee general ledger entries"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// MEMBER
	// MEMBER GENERAL LEDGER ROUTES
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/check-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/online-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/cash-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/payment-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/withdraw-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/deposit-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/journal-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/adjustment-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/journal-voucher/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/check-voucher/search

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID:   memberProfileID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: model.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID:   memberProfileID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: model.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID:   memberProfileID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: model.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			Source:          model.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			Source:          model.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			Source:          model.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			Source:          model.GeneralLedgerSourceJournal,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			Source:          model.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			Source:          model.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}

		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			Source:          model.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// MEMBER ACCOUNT
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/online-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/cash-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/payment-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/withdraw-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/deposit-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/adjustment-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-voucher/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-voucher/search
	// MEMBER ACCOUNT GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID:   memberProfileID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			AccountID:         accountID,
			TypeOfPaymentType: model.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID:   memberProfileID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			AccountID:         accountID,
			TypeOfPaymentType: model.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID:   memberProfileID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			AccountID:         accountID,
			TypeOfPaymentType: model.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          model.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          model.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          model.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          model.GeneralLedgerSourceJournal,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          model.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          model.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
			Source:          model.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// TRANSACTION BATCH
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/online-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/cash-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/payment-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/withdraw-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/deposit-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/adjustment-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-voucher/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-voucher/search
	// TRANSACTION BATCH GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TypeOfPaymentType:  model.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TypeOfPaymentType:  model.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			TypeOfPaymentType:  model.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             model.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             model.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             model.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             model.GeneralLedgerSourceJournal,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             model.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             model.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			Source:             model.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// TRANSACTION
	// GET /api/v1/general-ledger/transaction/:transaction_id/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/check-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/online-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/cash-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/payment-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/withdraw-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/deposit-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/journal-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/adjustment-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/journal-voucher/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/check-voucher/search
	// TRANSACTION GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/transaction/:transaction_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionID:  transactionId,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionID:     transactionId,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: model.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionID:     transactionId,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: model.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionID:     transactionId,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: model.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionID:  transactionId,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionID:  transactionId,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionID:  transactionId,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionID:  transactionId,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceJournal,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionID:  transactionId,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionID:  transactionId,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			TransactionID:  transactionId,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// ACCOUNTS
	// GET /api/v1/general-ledger/account/:account_id/search
	// GET /api/v1/general-ledger/account/:account_id/check-entry/search
	// GET /api/v1/general-ledger/account/:account_id/online-entry/search
	// GET /api/v1/general-ledger/account/:account_id/cash-entry/search
	// GET /api/v1/general-ledger/account/:account_id/payment-entry/search
	// GET /api/v1/general-ledger/account/:account_id/withdraw-entry/search
	// GET /api/v1/general-ledger/account/:account_id/deposit-entry/search
	// GET /api/v1/general-ledger/account/:account_id/journal-entry/search
	// GET /api/v1/general-ledger/account/:account_id/adjustment-entry/search
	// GET /api/v1/general-ledger/account/:account_id/journal-voucher/search
	// GET /api/v1/general-ledger/account/:account_id/check-voucher/search
	// ACCOUNTS GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/account/:account_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/account/:account_id/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:         accountID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: model.PaymentTypeCheck,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/account/:account_id/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:         accountID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: model.PaymentTypeOnline,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/account/:account_id/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:         accountID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			TypeOfPaymentType: model.PaymentTypeCash,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/account/:account_id/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourcePayment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/account/:account_id/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceWithdraw,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/account/:account_id/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/account/:account_id/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceJournal,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/account/:account_id/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceAdjustment,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/account/:account_id/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceJournalVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// GET /api/v1/general-ledger/account/:account_id/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Source:         model.GeneralLedgerSourceCheckVoucher,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})
}
