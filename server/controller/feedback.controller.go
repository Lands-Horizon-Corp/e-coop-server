package controller

import (
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
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

	return &FeedbackController{
		request:            request,
		feedbackRepository: feedbackRepository,
	}, nil
}

func (fc *FeedbackController) Create(c echo.Context) error {
	return nil
}

func (fc *FeedbackController) Update(c echo.Context) error {
	return nil

}

func (fc *FeedbackController) Delete(c echo.Context) error {
	return nil
}
