package controller_v1

import (
	"net/http"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	modelCore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelCore"
	"github.com/labstack/echo/v4"
)

func (c *Controller) MemberAccountingLedgerController() {
	req := c.provider.Service.Request
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/member-profile/:member_profile_id/total",
		Method:       "GET",
		ResponseType: modelCore.MemberAccountingLedgerSummary{},
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
		if userOrg.UserType != modelCore.UserOrganizationTypeOwner && userOrg.UserType != modelCore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger totals"})
		}

		if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cash on hand account not set for branch"})
		}
		if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Paid-up shared capital account not set for branch"})
		}
		entries, err := c.modelCore.MemberAccountingLedgerManager.FindWithFilters(context, []horizon_services.Filter{
			{Field: "member_accounting_ledgers.member_profile_id", Op: horizon_services.OpEq, Value: memberProfileID},
			{Field: "member_accounting_ledgers.organization_id", Op: horizon_services.OpEq, Value: userOrg.OrganizationID},
			{Field: "member_accounting_ledgers.branch_id", Op: horizon_services.OpEq, Value: *userOrg.BranchID},
			{Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.CashOnHandAccountID},
			// {Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID},
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		paidUpShareCapital, err := c.modelCore.MemberAccountingLedgerManager.Find(context, &modelCore.MemberAccountingLedger{
			MemberProfileID: *memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       *userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve paid-up share capital entries: " + err.Error()})
		}

		var TotalShareCapitalPlusFixedSavings float64
		for _, entry := range paidUpShareCapital {
			TotalShareCapitalPlusFixedSavings += entry.Balance
		}
		var totalDeposits float64
		for _, entry := range entries {
			totalDeposits += entry.Balance
		}

		summary := modelCore.MemberAccountingLedgerSummary{
			TotalDeposits:                     totalDeposits,
			TotalShareCapitalPlusFixedSavings: TotalShareCapitalPlusFixedSavings,
			TotalLoans:                        0,
		}
		return ctx.JSON(http.StatusOK, summary)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/member-profile/:member_profile_id/account/:account_id/total",
		Method:       "GET",
		ResponseType: modelCore.MemberAccountingLedgerAccountSummary{},
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelCore.UserOrganizationTypeOwner && userOrg.UserType != modelCore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger totals"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Branch ID is missing for user organization"})
		}

		entries, err := c.modelCore.GeneralLedgerManager.Find(context, &modelCore.GeneralLedger{
			MemberProfileID: memberProfileID,
			AccountID:       accountID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		memberAccountingLedger, err := c.modelCore.MemberAccountingLedgerManager.FindOne(context, &modelCore.MemberAccountingLedger{
			MemberProfileID: *memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       *accountID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}

		var totalDebit float64
		var totalCredit float64

		for _, entry := range entries {
			totalCredit += entry.Credit
			totalDebit += entry.Debit
		}
		return ctx.JSON(http.StatusOK, modelCore.MemberAccountingLedgerAccountSummary{
			Balance:     memberAccountingLedger.Balance,
			TotalDebit:  totalDebit,
			TotalCredit: totalCredit,
		})
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: modelCore.MemberAccountingLedger{},
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
		if userOrg.UserType != modelCore.UserOrganizationTypeOwner && userOrg.UserType != modelCore.UserOrganizationTypeEmployee {
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
		entries, err := c.modelCore.MemberAccountingLedgerManager.FindWithFilters(context, []horizon_services.Filter{
			{Field: "member_accounting_ledgers.member_profile_id", Op: horizon_services.OpEq, Value: memberProfileID},
			{Field: "member_accounting_ledgers.organization_id", Op: horizon_services.OpEq, Value: userOrg.OrganizationID},
			{Field: "member_accounting_ledgers.branch_id", Op: horizon_services.OpEq, Value: *userOrg.BranchID},
			{Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.CashOnHandAccountID},
			// {Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID},
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.MemberAccountingLedgerManager.Pagination(context, ctx, entries))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/member-profile/:member_profile_id",
		Method:       "GET",
		ResponseType: modelCore.MemberAccountingLedger{},
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
		if userOrg.UserType != modelCore.UserOrganizationTypeOwner && userOrg.UserType != modelCore.UserOrganizationTypeEmployee {
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
		entries, err := c.modelCore.MemberAccountingLedgerManager.FindWithFilters(context, []horizon_services.Filter{
			{Field: "member_accounting_ledgers.member_profile_id", Op: horizon_services.OpEq, Value: memberProfileID},
			{Field: "member_accounting_ledgers.organization_id", Op: horizon_services.OpEq, Value: userOrg.OrganizationID},
			{Field: "member_accounting_ledgers.branch_id", Op: horizon_services.OpEq, Value: *userOrg.BranchID},
			{Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.CashOnHandAccountID},
			// {Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID},
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.MemberAccountingLedgerManager.Filtered(context, ctx, entries))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-accounting-ledger/branch/search",
		Method:       "GET",
		ResponseType: modelCore.MemberAccountingLedger{},
		Note:         "Returns paginated general ledger entries for a specific member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelCore.UserOrganizationTypeOwner && userOrg.UserType != modelCore.UserOrganizationTypeEmployee {
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
		entries, err := c.modelCore.MemberAccountingLedgerManager.FindWithFilters(context, []horizon_services.Filter{
			{Field: "member_accounting_ledgers.organization_id", Op: horizon_services.OpEq, Value: userOrg.OrganizationID},
			{Field: "member_accounting_ledgers.branch_id", Op: horizon_services.OpEq, Value: *userOrg.BranchID},
			{Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.CashOnHandAccountID},
			// {Field: "member_accounting_ledgers.account_id", Op: horizon_services.OpNe, Value: userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID},
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.MemberAccountingLedgerManager.Pagination(context, ctx, entries))
	})

}
