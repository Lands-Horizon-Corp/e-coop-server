package providers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

func (p *Providers) SetUser(c echo.Context, user *model.User) error {
	p.authentication.CleanToken(c)
	if err := p.authentication.SetToken(c, horizon.Claim{
		ID:            user.ID.String(),
		Email:         user.Email,
		ContactNumber: user.ContactNumber,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
	}
	return nil
}
func (p *Providers) CurrentUser(c echo.Context) (*model.User, error) {
	claim, err := p.authentication.GetUserFromToken(c)
	if err != nil {
		p.authentication.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		p.authentication.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}
	user, err := p.user.Manager.GetByID(id)
	if err != nil {
		p.authentication.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if user.Email != claim.Email {
		p.authentication.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if user.ContactNumber != claim.ContactNumber {
		p.authentication.CleanToken(c)
		return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return user, nil
}
