package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

func (h *Handler) FootstepList(c echo.Context) error {
	footstep, err := h.repository.FootstepList()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.FootstepModels(footstep))
}

func (h *Handler) FootstepGet(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid footstep ID"})
	}
	footstep, err := h.repository.FootstepGetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.FootstepModel(footstep))
}

func (h *Handler) FootstepDelete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid footstep ID"})
	}
	model := &model.Footstep{ID: id}
	if err := h.repository.FootstepDelete(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
