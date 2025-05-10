package provider

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

func (p *Provider) CurrentUser(c echo.Context) (*model.User, error) {
	claim, err := p.authentication.GetUserFromToken(c)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}
	user, err := p.repository.UserGetByID(id)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return user, nil
}

func (h *Provider) EnsureEmployeeOrOwner(
	c echo.Context,
	orgID, branchID uuid.UUID,
) (*model.UserOrganization, error) {
	user, err := h.CurrentUser(c)
	if err != nil {
		return nil, err
	}
	uo, err := h.repository.UserOrganizationGetByUserOrgBranch(user.ID, orgID, branchID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusForbidden, "not a member of this branch")
	}
	if uo.UserType != "employee" && uo.UserType != "owner" {
		return nil, echo.NewHTTPError(http.StatusForbidden, "only owners or employees may manage this branch")
	}
	return uo, nil
}
