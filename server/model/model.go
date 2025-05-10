package models

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

type Model struct {
	validator *validator.Validate
	storage   *horizon.HorizonStorage
	qr        *horizon.HorizonQR
}

func NewModel(
	storage *horizon.HorizonStorage,
	qr *horizon.HorizonQR,
) (*Model, error) {
	return &Model{
		validator: validator.New(),
		storage:   storage,
		qr:        qr,
	}, nil
}

func ValidateCreate[T any](ctx echo.Context) (*T, error) {
	validator := validator.New()
	var req *T
	if err := ctx.Bind(&req); err != nil {
		return req, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := validator.Struct(req); err != nil {
		return req, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return req, nil
}
