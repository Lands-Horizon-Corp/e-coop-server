package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

func (h *Handler) ContactUsList(c echo.Context) error {
	contact_us, err := h.repository.ContactUsList()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.ContactUsModels(contact_us))
}

func (h *Handler) ContactUsGet(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid contact_us ID"})
	}
	contact_us, err := h.repository.ContactUsGetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.ContactUsModel(contact_us))
}

func (h *Handler) ContactUsCreate(c echo.Context) error {
	req, err := h.model.ContactUsValidate(c)
	if err != nil {
		return err
	}
	model := &model.ContactUs{
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		ContactNumber: req.ContactNumber,
		Description:   req.Description,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if err := h.repository.ContactUsCreate(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, h.model.ContactUsModel(model))
}

func (h *Handler) ContactUsUpdate(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid contact_us ID"})
	}
	req, err := h.model.ContactUsValidate(c)
	if err != nil {
		return err
	}
	model := &model.ContactUs{
		ID:            id,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		ContactNumber: req.ContactNumber,
		Description:   req.Description,
		UpdatedAt:     time.Now().UTC(),
	}
	if err := h.repository.ContactUsUpdate(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.ContactUsModel(model))

}

func (h *Handler) ContactUsDelete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid contact-us ID"})
	}
	model := &model.ContactUs{ID: id}
	if err := h.repository.ContactUsDelete(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
