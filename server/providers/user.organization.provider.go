package providers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

func (p *Providers) UserOrganization(c echo.Context, organizationID, branchID string) (*model.UserOrganization, error) {
	user, err := p.CurrentUser(c)
	if err != nil {
		return nil, err
	}
	org, err := uuid.Parse(organizationID)
	if err != nil {
		p.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid user Organization")
	}
	branch, err := uuid.Parse(branchID)
	if err != nil {
		p.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid user Branch")
	}
	userOrg, err := p.userOrganization.ByUserOrganizationBranch(user.ID, org, branch)
	if err != nil {
		p.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid user organization")
	}
	return userOrg, nil
}

func (p *Providers) UserOwner(c echo.Context, organizationID, branchID string) (*model.UserOrganization, error) {
	user, err := p.UserOrganization(c, organizationID, branchID)
	if err != nil {
		p.CleanToken(c)
		return nil, err
	}
	if user.UserType != "owner" {
		p.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusForbidden, "only owners may manage this page")
	}
	return user, nil
}

func (p *Providers) UserEmployee(c echo.Context, organizationID, branchID string) (*model.UserOrganization, error) {
	user, err := p.UserOrganization(c, organizationID, branchID)
	if err != nil {
		p.CleanToken(c)
		return nil, err
	}
	if user.UserType != "employee" {
		p.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusForbidden, "only employees may manage this page")
	}
	return user, nil
}

func (p *Providers) UserOwnerEmployee(c echo.Context, organizationID, branchID string) (*model.UserOrganization, error) {
	user, err := p.UserOrganization(c, organizationID, branchID)
	if err != nil {
		p.CleanToken(c)
		return nil, err
	}
	if user.UserType != "employee" && user.UserType != "owner" {
		p.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusForbidden, "only employees &  owner may manage this page")
	}
	return user, nil
}

func (p *Providers) UserMember(c echo.Context, organizationID, branchID string) (*model.UserOrganization, error) {
	user, err := p.UserOrganization(c, organizationID, branchID)
	if err != nil {
		p.CleanToken(c)
		return nil, err
	}
	if user.UserType != "member" {
		p.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusForbidden, "only members may manage this page")
	}
	return user, nil
}
