package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

// GeneralLedgerController manages endpoints for general ledger accounts, definitions, and member ledgers.
func (c *Controller) GeneralLedgerController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:        "/general-ledger/account/:account_id/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for an account with pagination.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := horizon.EngineUUIDParam(ctx, "account_id")
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

	req.RegisterRoute(horizon.Route{
		Route:        "/general-ledger/member-profile/:member_profile_id/account/:account_id/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns paginated general ledger entries for a specific member profile and account.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := horizon.EngineUUIDParam(ctx, "account_id")
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
	req.RegisterRoute(horizon.Route{
		Route:        "/general-ledger/member-profile/:member_profile_id/account/:account_id/total",
		Method:       "GET",
		ResponseType: model.MemberGeneralLedgerTotal{},
		Note:         "Returns the total amount for a specific member profile's general ledger entries for an account.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := horizon.EngineUUIDParam(ctx, "account_id")
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
			totalAmount += entry.Debit - entry.Credit
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
}
