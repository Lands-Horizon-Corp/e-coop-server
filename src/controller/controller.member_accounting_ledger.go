package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberAccountingLedgerController() {
	req := c.provider.Service.Request
	req.RegisterRoute(horizon.Route{
		Route:        "/member-accounting-ledger/member-profile/:member_profile_id/total",
		Method:       "GET",
		ResponseType: model.MemberAccountingLedgerSummary{},
		Note:         "Returns the total amount for a specific member profile's general ledger entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
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
		entries, err := c.model.MemberAccountingLedgerManager.Find(context, &model.MemberAccountingLedger{
			MemberProfileID: *memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		var totalAmount float64
		for _, entry := range entries {
			totalAmount += entry.Balance
		}
		summary := model.MemberAccountingLedgerSummary{
			TotalAmount:                  totalAmount,
			TotalShareCapitalPlusSavings: 0,
			TotalLoans:                   totalAmount,
		}
		return ctx.JSON(http.StatusOK, summary)
	})

	req.RegisterRoute(horizon.Route{
		Route:        "/member-accounting-ledger/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: model.MemberAccountingLedger{},
		Note:         "Returns paginated general ledger entries for a specific member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
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
		entries, err := c.model.MemberAccountingLedgerManager.Find(context, &model.MemberAccountingLedger{
			MemberProfileID: *memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberAccountingLedgerManager.Pagination(context, ctx, entries))
	})

}
