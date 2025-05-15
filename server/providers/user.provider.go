package providers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

func (p *Providers) SetUser(c echo.Context, user *model.User) error {
	if err := p.authentication.SetToken(c, horizon.Claim{
		ID:            user.ID.String(),
		Email:         user.Email,
		ContactNumber: user.ContactNumber,
		Password:      user.Password,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
	}
	return nil
}

func (p *Providers) SetCustom(c echo.Context, user *model.UserOrganization) error {
	if err := p.customAuth.SetToken(c, horizon.CustomClaim{
		UserOrganizationID: user.ID.String(),
		UserID:             user.UserID.String(),
		OrganizationID:     user.OrganizationID.String(),
		BranchID:           user.BranchID.String(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
	}
	return nil
}

func (p *Providers) CurrentUser(c echo.Context) (*model.User, error) {
	claim, err := p.authentication.GetUserFromToken(c)
	if err != nil {
		p.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		p.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}
	user, err := p.user.Manager.GetByID(id)
	if err != nil {
		p.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if user.Email != claim.Email {
		p.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusNotFound, "user changes email")
	}
	if user.ContactNumber != claim.ContactNumber {
		p.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusNotFound, "user changes contact number")
	}
	if !p.authentication.VerifyPassword(user.Password, claim.Password) {
		p.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusNotFound, "user changes password")
	}
	return user, nil
}

func (p *Providers) CurrentUserOrganization(c echo.Context) (*model.UserOrganization, error) {
	claim, err := p.customAuth.GetCustomFromToken(c)
	if err != nil {
		return nil, err
	}
	userOrgId, err := uuid.Parse(claim.UserOrganizationID)
	if err != nil {
		return nil, err
	}
	userOrganization, err := p.userOrganization.Manager.GetByID(userOrgId)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "user organization not found")
	}
	if userOrganization.ID != userOrgId {
		return nil, echo.NewHTTPError(http.StatusNotFound, "user changes organization")
	}
	return userOrganization, nil
}
