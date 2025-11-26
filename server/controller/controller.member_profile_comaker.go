package v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) memberProfileComaker() {
	req := c.provider.Service.Request

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-profile-comaker/member-profile/:member_profile_id",
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
		loanTransactions, err := c.core.LoanTransactionManager.Find(context, &core.LoanTransaction{
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
			MemberProfileID: memberProfileID,
		}, "Account")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}
		comakerResponse := []core.ComakerMemberProfileResponse{}
		for _, lt := range loanTransactions {
			comakers, err := c.core.ComakerMemberProfileManager.Find(context, &core.ComakerMemberProfile{
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				LoanTransactionID: lt.ID,
			}, "MemberProfile", "MemberProfile.Media")
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve comaker details: " + err.Error()})
			}
			for _, cm := range comakers {
				comakerResponse = append(comakerResponse, core.ComakerMemberProfileResponse{
					LoanTransaction: c.core.LoanTransactionManager.ToModel(lt),
					MemberProfile:   c.core.MemberProfileManager.ToModel(cm.MemberProfile),
				})
			}
		}
		return ctx.JSON(http.StatusOK, comakerResponse)
	})
}
