package provider

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
	"horizon.com/server/server/repository"
)

type UserProvider struct {
	repo           *repository.Repository
	model          *model.Model
	authentication *horizon.HorizonAuthentication
}

func NewUserProvider(
	repo *repository.Repository,
	model *model.Model,
	authentication *horizon.HorizonAuthentication,

) (*UserProvider, error) {
	return &UserProvider{
		repo:           repo,
		model:          model,
		authentication: authentication,
	}, nil
}

func (up *UserProvider) CurrentUser(c echo.Context) (*model.User, error) {
	claim, err := up.authentication.GetUserFromToken(c)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	id, err := uuid.Parse(claim.ID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}
	user, err := up.repo.UserGetByID(id)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return user, nil
}
