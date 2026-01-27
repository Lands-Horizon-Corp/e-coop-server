package member_profile

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func MemberProfileComakerController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile-comaker/member-profile/:member_profile_id",
		Method:       "GET",
		Note:         "Retrieves comaker details for a specific member profile ID.",
		ResponseType: types.ComakerMemberProfileResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		loanTransactions, err := core.LoanTransactionManager(service).Find(context, &types.LoanTransaction{
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
			MemberProfileID: memberProfileID,
		}, "Account")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}
		comakerResponse := []types.ComakerMemberProfileResponse{}
		for _, lt := range loanTransactions {
			comakers, err := core.ComakerMemberProfileManager(service).Find(context, &types.ComakerMemberProfile{
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				LoanTransactionID: lt.ID,
			}, "MemberProfile", "MemberProfile.Media", "LoanTransaction")
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve comaker details: " + err.Error()})
			}
			for _, cm := range comakers {
				comakerResponse = append(comakerResponse, *core.ComakerMemberProfileManager(service).ToModel(cm))
			}
		}
		return ctx.JSON(http.StatusOK, comakerResponse)
	})
}
