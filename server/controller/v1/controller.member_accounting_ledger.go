package v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) memberAccountingLedgerController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/member-profile/:member_profile_id/total",
		Method:       "GET",
		ResponseType: event.MemberAccountingLedgerSummary{},
		Note:         "Returns the total amount for a specific member profile's general ledger entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}

		summary, err := c.event.MemberAccountingLedgerSummary(context, ctx, memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, summary)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/member-profile/:member_profile_id/account/:account_id/total",
		Method:       "GET",
		ResponseType: core.MemberAccountingLedgerAccountSummary{},
		Note:         "Returns the total amount for a specific member profile and account ID.",
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
		userOrg, err := c.token.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger totals"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Branch ID is missing for user organization"})
		}

		entries, err := c.core.GeneralLedgerManager().Find(context, &core.GeneralLedger{
			MemberProfileID: memberProfileID,
			AccountID:       accountID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		balance, err := usecase.CalculateBalance(usecase.Balance{
			GeneralLedgers: entries,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to compute balance: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberAccountingLedgerAccountSummary{
			Balance:     balance.Balance,
			TotalDebit:  balance.Debit,
			TotalCredit: balance.Credit,
		})
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: core.MemberAccountingLedger{},
		Note:         "Returns paginated general ledger entries for a specific member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.token.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Branch ID is missing for user organization"})
		}
		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}

		paginatedResult, err := c.core.MemberAccountingLedgerManager().NormalPagination(context, ctx, &core.MemberAccountingLedger{
			MemberProfileID: *memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to paginate entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, paginatedResult)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/member-profile/:member_profile_id",
		Method:       "GET",
		ResponseType: core.MemberAccountingLedger{},
		Note:         "Returns paginated general ledger entries for a specific member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.token.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Branch ID is missing for user organization"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.core.MemberAccountingLedgerMemberProfileEntries(context,
			*memberProfileID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			*userOrg.Branch.BranchSetting.CashOnHandAccountID,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MemberAccountingLedgerManager().ToModels(entries))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/branch/search",
		Method:       "GET",
		ResponseType: core.MemberAccountingLedger{},
		Note:         "Returns paginated general ledger entries for a specific member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.token.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Branch ID is missing for user organization"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}

		paginatedResult, err := c.core.MemberAccountingLedgerManager().ArrPagination(context, ctx, []registry.FilterSQL{
			{Field: "organization_id", Op: query.ModeEqual, Value: userOrg.OrganizationID},
			{Field: "branch_id", Op: query.ModeEqual, Value: userOrg.BranchID},
			{Field: "account_id", Op: query.ModeNotEqual, Value: userOrg.Branch.BranchSetting.CashOnHandAccountID},
		}, nil)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to paginate entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, paginatedResult)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/member-profile/:member_profile_id/compassion-fund-account",
		Method:       "GET",
		ResponseType: core.MemberAccountingLedger{},
		Note:         "Returns single account for member accounting ledger compassion fund.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.token.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Branch ID is missing for user organization"})
		}
		if userOrg.Branch.BranchSetting.CompassionFundAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Compassion fund account not set for branch"})
		}
		ledger, err := c.core.MemberAccountingLedgerManager().FindOne(context, &core.MemberAccountingLedger{
			MemberProfileID: *memberProfileID,
			AccountID:       *userOrg.Branch.BranchSetting.CompassionFundAccountID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entry: " + err.Error()})
		}
		if ledger == nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member accounting ledger entry not found"})
		}
		return ctx.JSON(http.StatusOK, c.core.MemberAccountingLedgerManager().ToModel(ledger))
	})

}
