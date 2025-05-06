package controller

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/repository"
)

type FeedbackController struct {
	request            *horizon.HorizonRequest
	feedbackRepository *repository.FeedbackRepository
}

func NewFeedbackController(
	request *horizon.HorizonRequest,
	feedbackRepository *repository.FeedbackRepository,
) (*FeedbackController, error) {
	controller := &FeedbackController{
		request:            request,
		feedbackRepository: feedbackRepository,
	}

	// Register routes
	e := request.Service()
	e.POST("/feedback", controller.Create)
	e.PATCH("/feedback/:id", controller.Update)
	e.DELETE("/feedback/:id", controller.Delete)

	return controller, nil
}

func (fc *FeedbackController) Create(c echo.Context) error {
	var feedback collection.Feedback
	if err := c.Bind(&feedback); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	feedback.ID = uuid.New() // Ensure ID is set
	if err := fc.feedbackRepository.Create(&feedback); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, feedback)
}

func (fc *FeedbackController) Update(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid UUID"})
	}

	var feedback collection.Feedback
	if err := c.Bind(&feedback); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	feedback.ID = id

	if err := fc.feedbackRepository.Update(&feedback); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, feedback)
}

func (fc *FeedbackController) Delete(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid UUID"})
	}

	feedback := &collection.Feedback{ID: id}
	if err := fc.feedbackRepository.Delete(feedback); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}
