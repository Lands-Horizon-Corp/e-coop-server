package provider

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/repository"
)

type UserProvider struct {
	repo           *repository.UserRepository
	collector      *collection.UserCollection
	authentication *horizon.HorizonAuthentication
}

func NewUserProvider(
	repo *repository.UserRepository,
	collector *collection.UserCollection,
	authentication *horizon.HorizonAuthentication,

) (*UserProvider, error) {
	return &UserProvider{
		repo:           repo,
		collector:      collector,
		authentication: authentication,
	}, nil
}

func (up *UserProvider) CurrentUser(c echo.Context) (*collection.User, error) {
	claim, err := up.authentication.GetUserFromToken(c)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}
	user, err := up.repo.GetByID(id)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return user, nil
}
