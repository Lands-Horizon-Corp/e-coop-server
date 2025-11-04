package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func (r *Registry[TData, TResponse, TRequest]) PaginationWithFields(
	context context.Context,
	ctx echo.Context,

	fields *TData,
	preloads ...string,
) (*filter.PaginationResult[TData], error) {
	var entities []*TData
	filterRoot, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to parse query")
	}
	if preloads == nil {
		preloads = r.preloads
	}
	filterRoot.Preload = preloads
	data, err := r.filtering.Hybrid(
		r.Client(context).Where(fields).Find(&entities),
		r.threshold,
		filterRoot,
		pageIndex, pageSize,
	)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to find filtered entities")
	}
	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) Pagination(
	context context.Context,
	ctx echo.Context,
	preloads ...string,
) (*filter.PaginationResult[TData], error) {
	filterRoot, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to parse query")
	}
	if preloads == nil {
		preloads = r.preloads
	}
	filterRoot.Preload = preloads
	data, err := r.filtering.Hybrid(
		r.Client(context),
		r.threshold,
		filterRoot,
		pageIndex, pageSize,
	)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to find filtered entities")
	}
	return data, nil
}
