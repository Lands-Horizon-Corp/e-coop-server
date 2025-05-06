package controller

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/repository"
)

type FeedbackController struct {
	request *horizon.HorizonRequest

	repository *repository.FeedbackRepository
	collection *collection.FeedbackCollection
}

func NewFeedbackController(
	request *horizon.HorizonRequest,
	repository *repository.FeedbackRepository,
	collection *collection.FeedbackCollection,
) (*FeedbackController, error) {
	return &FeedbackController{
		request:    request,
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
	}
	if err := fc.repository.Create(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	resp := fc.collection.ToModel(model)
	return c.JSON(http.StatusCreated, resp)
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

func (fc *FeedbackController) APIRoutes() {
	fc.request.Service().GET("/feedback", fc.List)
	fc.request.Service().GET("/feedback/:id", fc.Get)
	fc.request.Service().POST("/feedback", fc.Create)
	fc.request.Service().PUT("/feedback/:id", fc.Update)
	fc.request.Service().DELETE("/feedback/:id", fc.Delete)
}
