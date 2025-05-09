package controller

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/repository"
)

type ContactUsController struct {
	repository *repository.ContactUsRepository
	collection *collection.ContactUsCollection
}

func NewContactUsController(
	repository *repository.ContactUsRepository,
	collection *collection.ContactUsCollection,
) (*ContactUsController, error) {
	return &ContactUsController{
		repository: repository,
		collection: collection,
	}, nil
}

func (fc *ContactUsController) List(c echo.Context) error {
	contact_uss, err := fc.repository.List()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, fc.collection.ToModels(contact_uss))
}

func (fc *ContactUsController) Get(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid contact_us ID"})
	}
	contact_us, err := fc.repository.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	resp := fc.collection.ToModel(contact_us)
	return c.JSON(http.StatusOK, resp)
}

func (fc *ContactUsController) Create(c echo.Context) error {
	req, err := fc.collection.ValidateCreate(c)
	if err != nil {
		return err
	}
	model := &collection.ContactUs{
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		ContactNumber: req.ContactNumber,
		Description:   req.Description,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if err := fc.repository.Create(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, fc.collection.ToModel(model))
}

func (fc *ContactUsController) Update(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid contact_us ID"})
	}
	req, err := fc.collection.ValidateCreate(c)
	if err != nil {
		return err
	}
	model := &collection.ContactUs{
		ID:            id,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		ContactNumber: req.ContactNumber,
		Description:   req.Description,
		UpdatedAt:     time.Now().UTC(),
	}
	if err := fc.repository.Update(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, fc.collection.ToModel(model))

}

func (fc *ContactUsController) Delete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid contact-us ID"})
	}
	model := &collection.ContactUs{ID: id}
	if err := fc.repository.Delete(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (fc *ContactUsController) APIRoutes(service *echo.Echo) {
	service.GET("/contact-us", fc.List)
	service.GET("/contact-us/:id", fc.Get)
	service.POST("/contact-us", fc.Create)
	service.PUT("/contact-us/:id", fc.Update)
	service.DELETE("/contact-us/:id", fc.Delete)
}
