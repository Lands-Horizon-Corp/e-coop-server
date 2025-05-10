package controller

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/repository"
)

type FootstepController struct {
	repository *repository.FootstepRepository
	collection *collection.FootstepCollection
}

func NewFootstepController(
	repository *repository.FootstepRepository,
	collection *collection.FootstepCollection,
) (*FootstepController, error) {
	return &FootstepController{
		repository: repository,
		collection: collection,
	}, nil
}

func (fc *FootstepController) List(c echo.Context) error {
	footsteps, err := fc.repository.List()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, fc.collection.ToModels(footsteps))
}

func (fc *FootstepController) Get(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid footstep ID"})
	}
	footstep, err := fc.repository.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	resp := fc.collection.ToModel(footstep)
	return c.JSON(http.StatusOK, resp)
}

func (fc *FootstepController) Delete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid footstep ID"})
	}
	model := &collection.Footstep{ID: id}
	if err := fc.repository.Delete(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (fc *FootstepController) APIRoutes(service *echo.Echo) {
	service.GET("/footstep", fc.List)
	service.GET("/footstep/:id", fc.Get)
	service.DELETE("/footstep/:id", fc.Delete)
}
