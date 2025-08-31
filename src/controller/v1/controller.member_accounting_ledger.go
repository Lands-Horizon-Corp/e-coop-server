package controller_v1

import (
	"net/http"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/labstack/echo/v4"
)

func (c *Controller) MemberAccountingLedgerController() {
	req := c.provider.Service.Request
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/member-profile/:member_profile_id/total",
		Method:       "GET",
		ResponseType: model.MemberAccountingLedgerSummary{},
		Note:         "Returns the total amount for a specific member profile's general ledger entries.",
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger totals"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.model.MemberAccountingLedgerManager.FindWithFilters(context, []horizon_services.Filter{
			{Field: "member_accounting_ledgers.member_profile_id", Op: horizon_services.OpEq, Value: memberProfileID},
			{Field: "member_accounting_ledgers.organization_id", Op: horizon_services.OpEq, Value: userOrg.OrganizationID},
			{Field: "member_accounting_ledgers.branch_id", Op: horizon_services.OpEq, Value: *userOrg.BranchID},
			{Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.CashOnHandAccountID},
			{Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID},
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		paidUpShareCapital, err := c.model.MemberAccountingLedgerManager.Find(context, &model.MemberAccountingLedger{
			MemberProfileID: *memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       *userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve paid-up share capital entries: " + err.Error()})
		}

		var totalShareCapitalPlusSavings float64
		for _, entry := range paidUpShareCapital {
			totalShareCapitalPlusSavings += entry.Balance
		}
		var totalDeposits float64
		for _, entry := range entries {
			totalDeposits += entry.Balance
		}

		summary := model.MemberAccountingLedgerSummary{
			TotalDeposits:                totalDeposits,
			TotalShareCapitalPlusSavings: totalShareCapitalPlusSavings,
			TotalLoans:                   0,
		}
		return ctx.JSON(http.StatusOK, summary)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: model.MemberAccountingLedger{},
		Note:         "Returns paginated general ledger entries for a specific member profile.",
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
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Branch ID is missing for user organization"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.model.MemberAccountingLedgerManager.FindWithFilters(context, []horizon_services.Filter{
			{Field: "member_accounting_ledgers.member_profile_id", Op: horizon_services.OpEq, Value: memberProfileID},
			{Field: "member_accounting_ledgers.organization_id", Op: horizon_services.OpEq, Value: userOrg.OrganizationID},
			{Field: "member_accounting_ledgers.branch_id", Op: horizon_services.OpEq, Value: *userOrg.BranchID},
			{Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.CashOnHandAccountID},
			{Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID},
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberAccountingLedgerManager.Pagination(context, ctx, entries))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/member-profile/:member_profile_id",
		Method:       "GET",
		ResponseType: model.MemberAccountingLedger{},
		Note:         "Returns paginated general ledger entries for a specific member profile.",
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
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Branch ID is missing for user organization"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.model.MemberAccountingLedgerManager.FindWithFilters(context, []horizon_services.Filter{
			{Field: "member_accounting_ledgers.member_profile_id", Op: horizon_services.OpEq, Value: memberProfileID},
			{Field: "member_accounting_ledgers.organization_id", Op: horizon_services.OpEq, Value: userOrg.OrganizationID},
			{Field: "member_accounting_ledgers.branch_id", Op: horizon_services.OpEq, Value: *userOrg.BranchID},
			{Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.CashOnHandAccountID},
			{Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID},
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberAccountingLedgerManager.Filtered(context, ctx, entries))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/branch/search",
		Method:       "GET",
		ResponseType: model.MemberAccountingLedger{},
		Note:         "Returns paginated general ledger entries for a specific member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
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
		entries, err := c.model.MemberAccountingLedgerManager.FindWithFilters(context, []horizon_services.Filter{
			{Field: "member_accounting_ledgers.organization_id", Op: horizon_services.OpEq, Value: userOrg.OrganizationID},
			{Field: "member_accounting_ledgers.branch_id", Op: horizon_services.OpEq, Value: *userOrg.BranchID},
			{Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.CashOnHandAccountID},
			{Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID},
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberAccountingLedgerManager.Pagination(context, ctx, entries))
	})

}
