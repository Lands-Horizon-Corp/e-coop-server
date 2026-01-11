package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/labstack/echo/v4"
)

func (r *Registry[TData, TResponse, TRequest]) NoPagination(
	context context.Context,
	echocontext echo.Context,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPagination(r.Client(context), echocontext, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}

	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) NoPaginationNormal(
	context context.Context,
	echocontext echo.Context,
	filter *TData,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationNormal(r.Client(context), context, echocontext, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}

	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) NoPaginationArray(
	context context.Context,
	echocontext echo.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationArray(r.Client(context), context, echocontext, filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}

	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) NoPaginationStructured(
	context context.Context,
	echocontext echo.Context,
	filter query.StructuredFilter,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationStructured(r.Client(context), context, echocontext, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}

	return r.ToModels(data), nil
}
