package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /contact-us
func (c *Controller) ContactUsList(ctx echo.Context) error {
	contact_us, err := c.contactUs.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.ContactUsModels(contact_us))
}

// GET /contact-us/:contact_us_id
func (c *Controller) ContactUsGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "contact_us_id")
	if err != nil {
		return err
	}
	contact_us, err := c.contactUs.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.ContactUsModel(contact_us))
}

// POST /contact_us
func (c *Controller) ContactUsCreate(ctx echo.Context) error {
	req, err := c.model.ContactUsValidate(ctx)
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
	if err := c.contactUs.Manager.Create(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.ContactUsModel(model))
}

// PUT /contact-us/contact_us_id
func (c *Controller) ContactUsUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "contact_us_id")
	if err != nil {
		return err
	}
	req, err := c.model.ContactUsValidate(ctx)
	if err != nil {
		return err
	}
	model := &model.ContactUs{
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		ContactNumber: req.ContactNumber,
		Description:   req.Description,
		UpdatedAt:     time.Now().UTC(),
	}
	if err := c.contactUs.Manager.UpdateByID(*id, model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.ContactUsModel(model))
}

// DELETE /contact-us/contact_us_id
func (c *Controller) ContactUsDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "contact_us_id")
	if err != nil {
		return err
	}
	model := &model.ContactUs{ID: *id}
	if err := c.contactUs.Manager.Delete(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}
