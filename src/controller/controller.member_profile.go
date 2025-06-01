package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) MemberProfileController() {}

// History and Maintenance
func (c *Controller) MemberGender() {
	req := c.provider.Service.Request

	// User history

	// Maintenance
	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender",
		Method:   "GET",
		Response: "TMemberGender[]",
		Note:     "Getting Member gender on current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGender, err := c.model.MemberGenderCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.ToModels(memberGender))
	})
}

func (c *Controller) MemberCenter()         {}
func (c *Controller) MemberType()           {}
func (c *Controller) MemberClassification() {}
func (c *Controller) MemberOccupation()     {}
func (c *Controller) MemberGroup()          {}
