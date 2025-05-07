package controller

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/repository"
)

type FeedbackController struct {
	repository *repository.FeedbackRepository
	collection *collection.FeedbackCollection
}

func NewFeedbackController(
	repository *repository.FeedbackRepository,
	collection *collection.FeedbackCollection,
) (*FeedbackController, error) {
	return &FeedbackController{
		repository: repository,
		collection: collection,
	}, nil
}

func (fc *FeedbackController) List(c echo.Context) error {
	feedbacks, err := fc.repository.List()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	responses := fc.collection.ToModels(feedbacks)
	if len(responses) <= 0 {
		c.JSON(http.StatusOK, make([]any, 0))
	}
	return c.JSON(http.StatusOK, responses)
}

func (fc *FeedbackController) Get(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid feedback ID"})
	}
	feedback, err := fc.repository.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	resp := fc.collection.ToModel(feedback)
	return c.JSON(http.StatusOK, resp)
}

func (fc *FeedbackController) Create(c echo.Context) error {
	req, err := fc.collection.ValidateCreate(c)
	if err != nil {
		return err
	}
	model := &collection.Feedback{
		Email:        req.Email,
		Description:  req.Description,
		FeedbackType: req.FeedbackType,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if err := fc.repository.Create(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, fc.collection.ToModel(model))
}

func (fc *FeedbackController) Update(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid feedback ID"})
	}
	req, err := fc.collection.ValidateCreate(c)
	if err != nil {
		return err
	}
	model := &collection.Feedback{
		ID:           id,
		Email:        req.Email,
		Description:  req.Description,
		FeedbackType: req.FeedbackType,
		UpdatedAt:    time.Now().UTC(),
	}
	if err := fc.repository.Update(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, fc.collection.ToModel(model))

}

func (fc *FeedbackController) Delete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid feedback ID"})
	}
	model := &collection.Feedback{ID: id}
	if err := fc.repository.Delete(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (fc *FeedbackController) APIRoutes(service *echo.Echo) {
	service.GET("/feedback", fc.List)
	service.GET("/feedback/:id", fc.Get)
	service.POST("/feedback", fc.Create)
	service.PUT("/feedback/:id", fc.Update)
	service.DELETE("/feedback/:id", fc.Delete)
}
