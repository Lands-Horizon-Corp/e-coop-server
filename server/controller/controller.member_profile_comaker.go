package v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) memberProfileComaker() {
	req := c.provider.Service.Request

	// GET /api/v1/loan-transaction/member-profile/:member_profile_id/comaker
	req.RegisterRoute(handlers.Route{
		Route:        "	",
		Method:       "GET",
		Note:         "Retrieves comaker details for a specific member profile ID.",
		ResponseType: core.ComakerMemberProfileResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		loanTransactions, err := c.core.ComakerMemberProfileManager.FindRaw(context, &core.ComakerMemberProfile{
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			MemberProfileID: *memberProfileID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve comaker details: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, loanTransactions)
	})
}
