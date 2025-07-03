package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) GeneralLedgerController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-definition",
		Method:   "GET",
		Response: "GeneralLedgerDefinition[]",
		Note:     "List all general ledger definitions for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		gl, err := c.model.GeneralLedgerDefinitionManager.FindRaw(context, &model.GeneralLedgerDefinition{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, gl)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-accounts-grouping",
		Method:   "GET",
		Response: "GeneralLedgerAccountsGrouping[]",
		Note:     "List all general ledger accounts grouping for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		gl, err := c.model.GeneralLedgerAccountsGroupingManager.FindRaw(context, &model.GeneralLedgerAccountsGrouping{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, gl)
	})
}
