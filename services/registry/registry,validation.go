package registry

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

func Validate[T any](ctx echo.Context, v *validator.Validate) (*T, error) {
	var req T
	if err := ctx.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}
	if err := v.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "validation failed: "+err.Error())
	}
	return &req, nil
}

// Validate binds and validates request data from echo context.
func (c *Registry[TData, TResponse, TRequest]) Validate(ctx echo.Context) (*TRequest, error) {
	var req TRequest
	if err := ctx.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}
	if err := c.service.Validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "validation failed: "+err.Error())
	}
	return &req, nil
}
