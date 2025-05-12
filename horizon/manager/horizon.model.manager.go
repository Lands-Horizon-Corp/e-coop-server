package horizon_manager

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

func Validate[T any](ctx echo.Context, v *validator.Validate) (*T, error) {
	var req T
	if err := ctx.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := v.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return &req, nil
}
func ToModel[T any, G any](data *T, mapFunc func(*T) *G) *G {
	if data == nil {
		return nil
	}
	return mapFunc(data)
}
func ToModels[T any, G any](data []*T, mapFunc func(*T) *G) []*G {
	if data == nil {
		return []*G{}
	}
	out := make([]*G, 0, len(data))
	for _, item := range data {
		if m := mapFunc(item); m != nil {
			out = append(out, m)
		}
	}
	return out
}
